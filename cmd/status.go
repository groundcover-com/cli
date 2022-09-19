package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PORTAL_POLLING_TIMEOUT     = time.Minute * 3
	ALLIGATORS_POLLING_TIMEOUT = time.Minute * 3
	PODS_POLLING_INTERVAL      = time.Second * 10
	WAIT_FOR_PORTAL_FORMAT     = "Waiting until cluster establish connectivity"
	WAIT_FOR_ALLIGATORS_FORMAT = "Waiting until all nodes are monitored (%d/%d Nodes)"
	ALLIGATOR_LABEL_SELECTOR   = "app=alligator"
	PORTAL_LABEL_SELECTOR      = "app=portal"
	RUNNING_FIELD_SELECTOR     = "status.phase=Running"
)

func init() {
	RootCmd.AddCommand(StatusCmd)
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get groundcover deployment status",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := cmd.Context()
		namespace := viper.GetString(NAMESPACE_FLAG)
		kubeconfig := viper.GetString(KUBECONFIG_FLAG)
		kubecontext := viper.GetString(KUBECONTEXT_FLAG)
		releaseName := viper.GetString(HELM_RELEASE_FLAG)

		sentryKubeContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)
		sentryKubeContext.SetOnCurrentScope()

		var kubeClient *k8s.Client
		if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
			return err
		}

		var clusterSummary *k8s.ClusterSummary
		if clusterSummary, err = kubeClient.GetClusterSummary(namespace); err != nil {
			sentryKubeContext.ClusterReport = &k8s.ClusterReport{
				ClusterSummary: clusterSummary,
			}
			sentryKubeContext.SetOnCurrentScope()
			return err
		}

		clusterReport := k8s.DefaultClusterRequirements.Validate(ctx, kubeClient, clusterSummary)

		sentryKubeContext.ClusterReport = clusterReport
		sentryKubeContext.SetOnCurrentScope()

		sentryHelmContext := sentry_utils.NewHelmContext(releaseName, CHART_NAME, HELM_REPO_URL)
		sentryHelmContext.SetOnCurrentScope()

		var helmClient *helm.Client
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}

		var release *helm.Release
		if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
			return err
		}

		sentryHelmContext.ChartVersion = release.Version().String()
		sentryHelmContext.SetOnCurrentScope()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

		var chart *helm.Chart
		if chart, err = getLatestChart(helmClient, sentryHelmContext); err != nil {
			return err
		}

		if chart.Version().GT(release.Version()) {
			fmt.Printf("Current groundcover installation in your cluster version: %s is out of date!, The latest version is %s.", release.Version(), chart.Version())
		}

		var nodeList *v1.NodeList
		if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
			return err
		}
		nodesCount := len(nodeList.Items)

		if err = waitForAlligators(ctx, kubeClient, release, nodesCount, sentryHelmContext); err != nil {
			return err
		}

		sentry.CaptureMessage("status executed successfully")
		return nil
	},
}

func waitForPortal(ctx context.Context, kubeClient *k8s.Client, helmRelease *helm.Release, sentryHelmContext *sentry_utils.HelmContext) error {
	spinner := ui.NewSpinner(WAIT_FOR_PORTAL_FORMAT)
	spinner.StopMessage("Cluster established connectivity")
	spinner.StopFailMessage("Cluster failed to establish connectivity")

	spinner.Start()
	defer spinner.Stop()

	appVersion := helmRelease.Chart.AppVersion()

	isPortalRunningFunc := func() (bool, error) {
		podClient := kubeClient.CoreV1().Pods(helmRelease.Namespace)
		listOptions := metav1.ListOptions{
			LabelSelector: PORTAL_LABEL_SELECTOR,
			FieldSelector: RUNNING_FIELD_SELECTOR,
		}

		podList, err := podClient.List(ctx, listOptions)
		if err != nil {
			return false, err
		}

		for _, pod := range podList.Items {
			if pod.Annotations["groundcover_version"] == appVersion {
				return true, nil
			}
		}

		return false, nil
	}

	err := spinner.Poll(isPortalRunningFunc, PODS_POLLING_INTERVAL, PORTAL_POLLING_TIMEOUT)

	if err == nil {
		return nil
	}

	spinner.StopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return fmt.Errorf("timeout waiting for cluster to establish connectivity")
	}

	return err
}

func waitForAlligators(ctx context.Context, kubeClient *k8s.Client, helmRelease *helm.Release, expectedAlligatorsCount int, sentryHelmContext *sentry_utils.HelmContext) error {
	spinner := ui.NewSpinner(fmt.Sprintf(WAIT_FOR_ALLIGATORS_FORMAT, 0, expectedAlligatorsCount))
	spinner.StopMessage(fmt.Sprintf("All nodes are monitored (%d/%d Nodes)", expectedAlligatorsCount, expectedAlligatorsCount))

	spinner.Start()
	defer spinner.Stop()

	runningAlligators := 0
	err := spinner.Poll(
		func() (bool, error) {
			var err error
			runningAlligators, err = getRunningAlligators(ctx, kubeClient, helmRelease.Chart.AppVersion(), helmRelease.Namespace)
			if err != nil {
				return false, err
			}

			spinner.Message(fmt.Sprintf(WAIT_FOR_ALLIGATORS_FORMAT, runningAlligators, expectedAlligatorsCount))
			return runningAlligators == expectedAlligatorsCount, nil
		},
		PODS_POLLING_INTERVAL,
		ALLIGATORS_POLLING_TIMEOUT,
	)

	sentryHelmContext.RunningAlligators = fmt.Sprintf("%d/%d", runningAlligators, expectedAlligatorsCount)
	sentryHelmContext.SetOnCurrentScope()

	if err == nil {
		return nil
	}

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		sentry_utils.SetLevelOnCurrentScope(sentry.LevelWarning)
		spinner.SetWarningSign()
		spinner.StopFailMessage(fmt.Sprintf("Timeout waiting for all nodes to be monitored (%d/%d Nodes)", runningAlligators, expectedAlligatorsCount))
		spinner.StopFail()
		return nil
	}

	spinner.StopFailMessage(fmt.Sprintf("Not all nodes are monitored (%d/%d Nodes)", runningAlligators, expectedAlligatorsCount))
	spinner.StopFail()

	return err
}

func getRunningAlligators(ctx context.Context, kubeClient *k8s.Client, helmVersion string, namespace string) (int, error) {
	podClient := kubeClient.CoreV1().Pods(namespace)
	listOptions := metav1.ListOptions{
		LabelSelector: ALLIGATOR_LABEL_SELECTOR,
		FieldSelector: RUNNING_FIELD_SELECTOR,
	}

	runningAlligators := 0

	podList, err := podClient.List(ctx, listOptions)
	if err != nil {
		return runningAlligators, err
	}

	for _, pod := range podList.Items {
		if pod.Annotations["groundcover_version"] == helmVersion {
			runningAlligators++
		}
	}

	return runningAlligators, nil
}

func reportPodsStatus(ctx context.Context, kubeClient *k8s.Client, helmVersion string, namespace string, sentryHelmContext *sentry_utils.HelmContext) {
	podClient := kubeClient.CoreV1().Pods(namespace)

	podList, err := podClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}

	podsStatus := make(map[string]k8s.PodStatus)
	for _, pod := range podList.Items {
		podsStatus[pod.Name] = k8s.BuildPodStatus(pod)
	}

	sentryHelmContext.PodsStatus = podsStatus
	sentryHelmContext.SetOnCurrentScope()
}
