package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	SENSORS_POLLING_INTERVAL = time.Second * 15
	SENSORS_POLLING_RETRIES  = 28
	SENSORS_POLLING_TIMEOUT  = time.Minute * 7

	SENSOR_LABEL_SELECTOR  = "app=sensor"
	BACKEND_LABEL_SELECTOR = "app!=sensor"
	PORTAL_LABEL_SELECTOR  = "app=portal"
	RUNNING_FIELD_SELECTOR = "status.phase=Running"

	WAIT_FOR_PORTAL_FORMAT      = "Waiting until cluster establish connectivity"
	WAIT_FOR_PVCS_FORMAT        = "Waiting until all PVCs are bound (%d/%d PVCs)"
	WAIT_FOR_SENSORS_FORMAT     = "Waiting until all nodes are monitored (%d/%d Nodes)"
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

		ui.GlobalWriter.Println("Checking cluster requirements...")
		clusterReport.PrintStatus()

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

		ui.GlobalWriter.Println("Checking groundcover version...")

		if chart.Version().GT(release.Version()) {
			msg := fmt.Sprintf("Current groundcover installation in your cluster version: %s is out of date!, The latest version is %s.", release.Version(), chart.Version())
			ui.GlobalWriter.PrintWarningMessageln(msg)
		} else {
			msg := fmt.Sprintf("Current groundcover installation in your cluster version: %s is up to date.", release.Version())
			ui.GlobalWriter.PrintSuccessMessageln(msg)
		}
		ui.GlobalWriter.Println("Checking groundcover pods...")

		runningPods, err := listPodsStatuses(ctx, kubeClient, namespace, metav1.ListOptions{})
		if err != nil {
			ui.GlobalWriter.PrintErrorMessageln("Failed to list sensor pods")
			return err
		}

		if podStatusCheck(runningPods) {
			ui.GlobalWriter.PrintSuccessMessageln("All pods are running")
		}

		ui.GlobalWriter.Println("Checking connectivity to cloud...")
		success, err := checkClusterConnectivity(ctx, kubeClient, namespace)
		if err != nil {
			msg := fmt.Sprintf("Failed to check cluster connectivity: %s", err)
			ui.GlobalWriter.PrintErrorMessageln(msg)
		} else if success {
			ui.GlobalWriter.PrintSuccessMessageln("Cluster established connectivity")
		} else {
			ui.GlobalWriter.PrintErrorMessageln("Cluster failed to establish connectivity")
		}
		/*

			var nodeList *v1.NodeList
			if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
				return err
			}
			nodesCount := len(nodeList.Items)

			if err = waitForSensors(ctx, kubeClient, namespace, chart.AppVersion(), nodesCount, sentryHelmContext); err != nil {
				return err
			}
		*/

		return nil
	},
}

func checkClusterConnectivity(ctx context.Context, kubeClient *k8s.Client, namespace string) (bool, error) {
	apiKeySecret, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, "api-key", metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if apiKeySecret == nil {
		return false, errors.New("api-key secret not found")
	}

	apiKeyValue, ok := apiKeySecret.Data["API_KEY"]
	if !ok {
		return false, errors.New("api-key secret not found")
	}

	//decode the data using base64
	b64Decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(apiKeyValue))
	decodedApiKey, err := io.ReadAll(b64Decoder)
	if err != nil {
		return false, err
	}

	decodedApiKeyString := string(decodedApiKey)
	if decodedApiKeyString == "" {
		return false, errors.New("api-key secret is empty")
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodGet, "https://client.groundcover.com/client/status", nil); err != nil {
		return false, err
	}

	request.Header.Add("apikey", decodedApiKeyString)

	var client = &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return false, errors.New("invalid api-key")
	}

	return true, nil
}

func podStatusCheck(pods map[string]k8s.PodStatus) bool {
	hasErrors := false
	for podName, podStatus := range pods {
		podErrors := podStatus.GetContainerErrors()
		if len(podErrors) > 0 {
			hasErrors = true
			ui.GlobalWriter.PrintErrorMessageln(fmt.Sprintf("Pod %s has errors", podName))
			for _, podError := range podErrors {
				ui.GlobalWriter.Printf("  - %s\n", podError)
			}
		}
	}

	return !hasErrors
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

func waitForSensors(ctx context.Context, kubeClient *k8s.Client, namespace, appVersion string, expectedSensorsCount int, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	event := segment.NewEvent(AGENTS_VALIDATION_EVENT_NAME)
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	spinner := ui.GlobalWriter.NewSpinner(fmt.Sprintf(WAIT_FOR_SENSORS_FORMAT, 0, expectedSensorsCount))
	spinner.SetStopMessage(fmt.Sprintf("All nodes are monitored (%d/%d Nodes)", expectedSensorsCount, expectedSensorsCount))
	spinner.SetStopFailMessage(fmt.Sprintf(TIMEOUT_INSTALLATION_FORMAT, namespace))

	spinner.Start()
	defer spinner.WriteStop()

	runningSensors := 0

	isSensorRunningFunc := func() error {
		var err error

		if runningSensors, err = getRunningSensors(ctx, kubeClient, appVersion, namespace); err != nil {
			return err
		}

		spinner.WriteMessage(fmt.Sprintf(WAIT_FOR_SENSORS_FORMAT, runningSensors, expectedSensorsCount))

		if runningSensors >= expectedSensorsCount {
			return nil
		}

		err = errors.New("not all expected sensors are running")
		return ui.RetryableError(err)
	}

	err = spinner.Poll(ctx, isSensorRunningFunc, SENSORS_POLLING_INTERVAL, SENSORS_POLLING_TIMEOUT, SENSORS_POLLING_RETRIES)

	runningSensorsStr := fmt.Sprintf("%d/%d", runningSensors, expectedSensorsCount)
	sentryHelmContext.RunningSensors = runningSensorsStr
	sentry_utils.SetTagOnCurrentScope(sentry_utils.EXPECTED_NODES_COUNT_TAG, fmt.Sprintf("%d", expectedSensorsCount))
	sentry_utils.SetTagOnCurrentScope(sentry_utils.RUNNING_SENSORS_TAG, runningSensorsStr)

	sentryHelmContext.SetOnCurrentScope()
	event.
		Set("sensorsCount", expectedSensorsCount).
		Set("runningSensorsCount", runningSensors)

	if err == nil {
		return nil
	}

	defer spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		if runningSensors > 0 {
			spinner.SetWarningSign()
			spinner.SetStopFailMessage(fmt.Sprintf("groundcover managed to provision %d/%d nodes", runningSensors, expectedSensorsCount))
		}

		return ErrExecutionPartialSuccess
	}

	return err
}

func getRunningSensors(ctx context.Context, kubeClient *k8s.Client, appVersion string, namespace string) (int, error) {
	podClient := kubeClient.CoreV1().Pods(namespace)
	listOptions := metav1.ListOptions{
		LabelSelector: SENSOR_LABEL_SELECTOR,
		FieldSelector: RUNNING_FIELD_SELECTOR,
	}

	runningSensors := 0

	podList, err := podClient.List(ctx, listOptions)
	if err != nil {
		return runningSensors, err
	}

	for _, pod := range podList.Items {
		if pod.Annotations["groundcover_version"] == appVersion {
			runningSensors++
		}
	}

	return runningSensors, nil
}

func reportPodsStatus(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryHelmContext *sentry_utils.HelmContext) {
	backendPodStatus, err := listPodsStatuses(ctx, kubeClient, namespace, metav1.ListOptions{LabelSelector: BACKEND_LABEL_SELECTOR})
	if err != nil {
		return
	}

	agentPodsStatus, err := listPodsStatuses(ctx, kubeClient, namespace, metav1.ListOptions{LabelSelector: SENSOR_LABEL_SELECTOR})
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
