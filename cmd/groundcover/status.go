package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sentry "groundcover.com/pkg/custom_sentry"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SPINNER_TYPE                = 8 // .oO@*
	WAIT_FOR_ALLIGATORS_TIMEOUT = time.Minute * 2
	ALLIGATORS_POLLING_INTERVAL = time.Second * 10
)

func init() {
	RootCmd.AddCommand(StatusCmd)
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get groundcover current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var kuber *k8s.Kuber
		var helmChart *helm.HelmCharter
		var helmRelease *helm.HelmReleaser

		if kuber, err = k8s.NewKuber(viper.GetString(KUBECONFIG_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmRelease, err = helm.NewHelmReleaser(viper.GetString(HELM_RELEASE_FLAG)); err != nil {
			return err
		}
		if helmChart, err = helm.NewHelmCharter(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}

		currentVersion, isLatestNewer := checkCurrentDeploy(helmRelease, helmChart)
		if isLatestNewer {
			fmt.Printf("Current groundcover %s is out of date!, The latest version is %s.", currentVersion, helmChart.Version)
		}

		if err = waitForAlligators(cmd.Context(), kuber, helmRelease); err != nil {
			return err
		}

		return nil
	},
}

func checkCurrentDeploy(helmRelease *helm.HelmReleaser, helmChart *helm.HelmCharter) (string, bool) {
	var err error
	var release *release.Release
	var currentVersion semver.Version

	if release, err = helmRelease.Get(); err != nil {
		return "", false
	}

	if currentVersion, err = semver.ParseTolerant(release.Chart.Metadata.Version); err != nil {
		sentry.CaptureException(err)
	}

	return currentVersion.String(), helmChart.IsLatestNewer(currentVersion)
}

func waitForAlligators(ctx context.Context, kuber *k8s.Kuber, helmRelease *helm.HelmReleaser) error {
	var err error
	var podList []v1.Pod
	var numberOfNodes int
	var runningAlligators int

	if numberOfNodes, err = kuber.NodesCount(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(ALLIGATORS_POLLING_INTERVAL)
	ctx, cancel := context.WithTimeout(ctx, WAIT_FOR_ALLIGATORS_TIMEOUT)
	defer cancel()

	spinner := spinner.New(spinner.CharSets[SPINNER_TYPE], 100*time.Millisecond)
	spinner.Prefix = "Waiting until all nodes are monitored "
	spinner.Suffix = fmt.Sprintf(" %d/%d", 0, numberOfNodes)
	spinner.Color("green")
	spinner.Start()
	defer spinner.Stop()

	version := helmRelease.Version.String()
	listOptions := metav1.ListOptions{LabelSelector: "app=alligator", FieldSelector: "status.phase=Running"}

	for {
		select {
		case <-ticker.C:
			runningAlligators := 0
			if podList, err = kuber.ListPods(ctx, helmRelease.Namespace, listOptions); err != nil {
				return err
			}
			for _, pod := range podList {
				if pod.Annotations["groundcover_version"] == version {
					runningAlligators++
				}
			}
			spinner.Suffix = fmt.Sprintf(" %d/%d", runningAlligators, numberOfNodes)
			if runningAlligators == numberOfNodes {
				spinner.FinalMSG = fmt.Sprintf("All nodes are monitored %d/%d !\n", runningAlligators, numberOfNodes)
				spinner.Stop()
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("timed out while waiting for all nodes to be monitored, got only: %d/%d", runningAlligators, numberOfNodes)
		}
	}
}
