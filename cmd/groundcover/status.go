package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sentry "groundcover.com/pkg/custom_sentry"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SPINNER_TYPE                = 8 // .oO@*
	ALLIGATORS_POLLING_TIMEOUT  = time.Minute * 2
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
		var kubeClient *k8s.KubeClient
		var helmChart *helm.HelmCharter
		var helmRelease *helm.HelmReleaser

		if kubeClient, err = k8s.NewKubeClient(viper.GetString(KUBECONFIG_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmRelease, err = helm.NewHelmReleaser(viper.GetString(HELM_RELEASE_FLAG), viper.GetString(NAMESPACE_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmChart, err = helm.NewHelmCharter(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}

		currentVersion, isLatestNewer := checkCurrentDeployedVersion(helmRelease, helmChart)
		if isLatestNewer {
			fmt.Printf("Current groundcover %s is out of date!, The latest version is %s.", currentVersion, helmChart.Version)
		}

		if err = waitForAlligators(cmd.Context(), kubeClient, helmRelease); err != nil {
			return err
		}

		return nil
	},
}

func checkCurrentDeployedVersion(helmRelease *helm.HelmReleaser, helmChart *helm.HelmCharter) (string, bool) {
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

func waitForAlligators(ctx context.Context, kubeClient *k8s.KubeClient, helmRelease *helm.HelmReleaser) error {
	var err error
	var numberOfNodes int
	var runningAlligators int
	var podList *v1.PodList
	var nodeList *v1.NodeList

	version := helmRelease.Version.String()
	nodeClient := kubeClient.CoreV1().Nodes()
	podClient := kubeClient.CoreV1().Pods(helmRelease.Namespace)
	listOptions := metav1.ListOptions{LabelSelector: "app=alligator", FieldSelector: "status.phase=Running"}

	if nodeList, err = nodeClient.List(ctx, metav1.ListOptions{}); err != nil {
		return err
	}
	numberOfNodes = len(nodeList.Items)

	spinner := utils.NewSpinner(SPINNER_TYPE, "Waiting until all nodes are monitored ")
	spinner.Suffix = fmt.Sprintf(" (%d/%d)", runningAlligators, numberOfNodes)

	areAlligatorsRunning := func() (bool, error) {
		runningAlligators = 0
		if podList, err = podClient.List(ctx, listOptions); err != nil {
			return false, err
		}
		for _, pod := range podList.Items {
			if pod.Annotations["groundcover_version"] == version {
				runningAlligators++
			}
		}
		spinner.Suffix = fmt.Sprintf(" (%d/%d)", runningAlligators, numberOfNodes)
		if numberOfNodes > runningAlligators {
			return false, nil
		}
		spinner.FinalMSG = fmt.Sprintf("All nodes are monitored (%d/%d) !\n", runningAlligators, numberOfNodes)
		return true, nil
	}

	if err = spinner.Poll(areAlligatorsRunning, ALLIGATORS_POLLING_INTERVAL, ALLIGATORS_POLLING_TIMEOUT); err != nil {
		return err
	}

	return nil
}
