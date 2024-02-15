package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/segment"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PORTAL_POLLING_INTERVAL = time.Second * 15
	PORTAL_POLLING_RETRIES  = 28
	PORTAL_POLLING_TIMEOUT  = time.Minute * 7

	PVC_POLLING_INTERVAL = time.Second * 15
	PVC_POLLING_RETRIES  = 40
	PVC_POLLING_TIMEOUT  = time.Minute * 10

	ALLIGATORS_POLLING_INTERVAL = time.Second * 15
	ALLIGATORS_POLLING_RETRIES  = 28
	ALLIGATORS_POLLING_TIMEOUT  = time.Minute * 7

	ALLIGATOR_LABEL_SELECTOR = "app=alligator"
	BACKEND_LABEL_SELECTOR   = "app!=alligator"
	PORTAL_LABEL_SELECTOR    = "app=portal"
	RUNNING_FIELD_SELECTOR   = "status.phase=Running"

	WAIT_FOR_PORTAL_FORMAT      = "Waiting until cluster establish connectivity"
	WAIT_FOR_PVCS_FORMAT        = "Waiting until all PVCs are bound (%d/%d PVCs)"
	WAIT_FOR_ALLIGATORS_FORMAT  = "Waiting until all nodes are monitored (%d/%d Nodes)"
	TIMEOUT_INSTALLATION_FORMAT = "Installation takes longer than expected, you can check the status using \"kubectl get pods -n %s\""

	PVCS_VALIDATION_EVENT_NAME   = "pvcs_validation"
	AGENTS_VALIDATION_EVENT_NAME = "agents_validation"
	PORTAL_VALIDATION_EVENT_NAME = "portal_validation"
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
		if clusterSummary, err = kubeClient.GetClusterSummary(ctx, namespace, viper.GetString(STORAGE_CLASS_FLAG)); err != nil {
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
		if chart, err = pollGetChart(ctx, helmClient, sentryHelmContext); err != nil {
			return err
		}

		if chart.Version().GT(release.Version()) {
			ui.GlobalWriter.Printf("Current groundcover installation in your cluster version: %s is out of date!, The latest version is %s.", release.Version(), chart.Version())
		}

		var nodeList *v1.NodeList
		if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
			return err
		}
		nodesCount := len(nodeList.Items)

		if err = waitForAlligators(ctx, kubeClient, namespace, chart.AppVersion(), nodesCount, sentryHelmContext); err != nil {
			return err
		}

		return nil
	},
}

func waitForPortal(ctx context.Context, kubeClient *k8s.Client, namespace, appVersion string, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	event := segment.NewEvent(PORTAL_VALIDATION_EVENT_NAME)
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	spinner := ui.GlobalWriter.NewSpinner(WAIT_FOR_PORTAL_FORMAT)
	spinner.SetStopMessage("Cluster established connectivity")
	spinner.SetStopFailMessage(fmt.Sprintf(TIMEOUT_INSTALLATION_FORMAT, namespace))

	spinner.Start()
	defer spinner.WriteStop()

	isPortalRunningFunc := func() error {
		podClient := kubeClient.CoreV1().Pods(namespace)
		listOptions := metav1.ListOptions{
			LabelSelector: PORTAL_LABEL_SELECTOR,
			FieldSelector: RUNNING_FIELD_SELECTOR,
		}

		podList, err := podClient.List(ctx, listOptions)
		if err != nil {
			return err
		}

		for _, pod := range podList.Items {
			if pod.Annotations["groundcover_version"] == appVersion {
				return nil
			}
		}

		err = errors.New("portal pod is not running")
		return ui.RetryableError(err)
	}

	err = spinner.Poll(ctx, isPortalRunningFunc, PORTAL_POLLING_INTERVAL, PORTAL_POLLING_TIMEOUT, PORTAL_POLLING_RETRIES)

	if err == nil {
		return nil
	}

	defer spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return ErrExecutionPartialSuccess
	}

	return err
}

func waitForAlligators(ctx context.Context, kubeClient *k8s.Client, namespace, appVersion string, expectedAlligatorsCount int, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	event := segment.NewEvent(AGENTS_VALIDATION_EVENT_NAME)
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	spinner := ui.GlobalWriter.NewSpinner(fmt.Sprintf(WAIT_FOR_ALLIGATORS_FORMAT, 0, expectedAlligatorsCount))
	spinner.SetStopMessage(fmt.Sprintf("All nodes are monitored (%d/%d Nodes)", expectedAlligatorsCount, expectedAlligatorsCount))
	spinner.SetStopFailMessage(fmt.Sprintf(TIMEOUT_INSTALLATION_FORMAT, namespace))

	spinner.Start()
	defer spinner.WriteStop()

	runningAlligators := 0

	isAlligatorRunningFunc := func() error {
		var err error

		if runningAlligators, err = getRunningAlligators(ctx, kubeClient, appVersion, namespace); err != nil {
			return err
		}

		spinner.WriteMessage(fmt.Sprintf(WAIT_FOR_ALLIGATORS_FORMAT, runningAlligators, expectedAlligatorsCount))

		if runningAlligators >= expectedAlligatorsCount {
			return nil
		}

		err = errors.New("not all expected alligators are running")
		return ui.RetryableError(err)
	}

	err = spinner.Poll(ctx, isAlligatorRunningFunc, ALLIGATORS_POLLING_INTERVAL, ALLIGATORS_POLLING_TIMEOUT, ALLIGATORS_POLLING_RETRIES)

	runningAlligatorsStr := fmt.Sprintf("%d/%d", runningAlligators, expectedAlligatorsCount)
	sentryHelmContext.RunningAlligators = runningAlligatorsStr
	sentry_utils.SetTagOnCurrentScope(sentry_utils.EXPECTED_NODES_COUNT_TAG, fmt.Sprintf("%d", expectedAlligatorsCount))
	sentry_utils.SetTagOnCurrentScope(sentry_utils.RUNNING_ALLIGATORS_TAG, runningAlligatorsStr)

	sentryHelmContext.SetOnCurrentScope()
	event.
		Set("alligatorsCount", expectedAlligatorsCount).
		Set("runningAlligatorsCount", runningAlligators)

	if err == nil {
		return nil
	}

	defer spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		if runningAlligators > 0 {
			spinner.SetWarningSign()
			spinner.SetStopFailMessage(fmt.Sprintf("groundcover managed to provision %d/%d nodes", runningAlligators, expectedAlligatorsCount))
		}

		return ErrExecutionPartialSuccess
	}

	return err
}

func getRunningAlligators(ctx context.Context, kubeClient *k8s.Client, appVersion string, namespace string) (int, error) {
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
		if pod.Annotations["groundcover_version"] == appVersion {
			runningAlligators++
		}
	}

	return runningAlligators, nil
}

func reportPodsStatus(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryHelmContext *sentry_utils.HelmContext) {
	backendPodStatus, err := listPodsStatuses(ctx, kubeClient, namespace, metav1.ListOptions{LabelSelector: BACKEND_LABEL_SELECTOR})
	if err != nil {
		return
	}

	agentPodsStatus, err := listPodsStatuses(ctx, kubeClient, namespace, metav1.ListOptions{LabelSelector: ALLIGATOR_LABEL_SELECTOR})
	if err != nil {
		return
	}

	sentryHelmContext.AgentStatus = agentPodsStatus
	sentryHelmContext.BackendStatus = backendPodStatus
	sentryHelmContext.SetOnCurrentScope()
}

func listPodsStatuses(ctx context.Context, kubeClient *k8s.Client, namespace string, options metav1.ListOptions) (map[string]k8s.PodStatus, error) {
	podList, err := kubeClient.CoreV1().Pods(namespace).List(ctx, options)
	if err != nil {
		return nil, err
	}

	podsStatuses := make(map[string]k8s.PodStatus)
	for _, pod := range podList.Items {
		podsStatuses[pod.Name] = k8s.BuildPodStatus(pod)
	}

	return podsStatuses, nil
}

func waitForPvcs(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace string, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	event := segment.NewEvent(PVCS_VALIDATION_EVENT_NAME)
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
	}

	var stsList *appsv1.StatefulSetList
	if stsList, err = kubeClient.AppsV1().StatefulSets(namespace).List(ctx, listOptions); err != nil {
		return err
	}

	expectedBoundPvcsCount := 0
	for _, sts := range stsList.Items {
		expectedBoundPvcsCount = expectedBoundPvcsCount + len(sts.Spec.VolumeClaimTemplates)
	}

	if expectedBoundPvcsCount == 0 {
		ui.GlobalWriter.PrintWarningMessageln("No Presistent volumes")
		return nil
	}

	spinner := ui.GlobalWriter.NewSpinner(fmt.Sprintf(WAIT_FOR_PVCS_FORMAT, 0, expectedBoundPvcsCount))
	spinner.SetStopMessage("Persistent Volumes are ready")
	spinner.SetStopFailMessage("Not all Persistent Volumes are bound, timeout waiting for them to be ready")

	spinner.Start()
	defer spinner.WriteStop()

	boundPvcs := make(map[string]bool, 0)

	var pvcList *v1.PersistentVolumeClaimList
	isPvcsReadyFunc := func() error {
		if pvcList, err = kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOptions); err != nil {
			return err
		}

		for _, pvc := range pvcList.Items {
			if pvc.Status.Phase == v1.ClaimBound {
				boundPvcs[pvc.Name] = true
			}
		}

		spinner.WriteMessage(fmt.Sprintf(WAIT_FOR_PVCS_FORMAT, len(boundPvcs), expectedBoundPvcsCount))

		if len(boundPvcs) >= expectedBoundPvcsCount {
			return nil
		}

		err = errors.New("not all expected pvcs are bound")
		return ui.RetryableError(err)
	}

	err = spinner.Poll(ctx, isPvcsReadyFunc, PVC_POLLING_INTERVAL, PVC_POLLING_TIMEOUT, PVC_POLLING_RETRIES)

	sentryHelmContext.BoundPvcs = maps.Keys(boundPvcs)
	sentryHelmContext.SetOnCurrentScope()
	event.
		Set("boundPvcsCount", len(boundPvcs)).
		Set("pvcsCount", expectedBoundPvcsCount)

	if err == nil {
		return nil
	}

	spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		err = errors.New("timeout waiting for persistent volume claims to be ready")
		return err
	}

	return err
}
