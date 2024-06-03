package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/segment"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"groundcover.com/pkg/utils"
)

const (
	HELM_DEPLOY_POLLING_RETRIES       = 2
	HELM_DEPLOY_POLLING_INTERVAL      = time.Second * 1
	HELM_DEPLOY_POLLING_TIMEOUT       = time.Minute * 5
	VALUES_FLAG                       = "values"
	MODE_FLAG                         = "mode"
	VERSION_FLAG                      = "version"
	REGISTRY_FLAG                     = "registry"
	STORAGE_CLASS_FLAG                = "storage-class"
	LOW_RESOURCES_FLAG                = "low-resources"
	ENABLE_CUSTOM_METRICS_FLAG        = "custom-metrics"
	ENABLE_KUBE_STATE_METRICS_FLAG    = "kube-state-metrics"
	STORE_ISSUES_LOGS_ONLY_FLAG       = "store-issues-logs-only"
	STORE_ISSUES_LOGS_ONLY_KEY        = "storeIssuesLogsOnly"
	CHART_NAME                        = "groundcover/groundcover"
	HELM_REPO_NAME                    = "groundcover"
	DEFAULT_GROUNDCOVER_RELEASE       = "groundcover"
	DEFAULT_GROUNDCOVER_NAMESPACE     = "groundcover"
	COMMIT_HASH_KEY_NAME_FLAG         = "git-commit-hash-key-name"
	REPOSITORY_URL_KEY_NAME_FLAG      = "git-repository-url-key-name"
	GROUNDCOVER_URL                   = "https://app.groundcover.com"
	HELM_REPO_URL                     = "https://helm.groundcover.com"
	CLUSTER_URL_FORMAT                = "%s/?clusterId=%s&viewType=Overview"
	QUAY_REGISTRY_PRESET_PATH         = "presets/quay.yaml"
	AGENT_KERNEL_5_11_PRESET_PATH     = "presets/agent/kernel-5-11.yaml"
	CUSTOM_METRICS_PRESET_PATH        = "presets/backend/custom-metrics.yaml"
	KUBE_STATE_METRICS_PRESET_PATH    = "presets/backend/kube-state-metrics.yaml"
	STORAGE_CLASS_TEMPLATE_PATH       = "templates/backend/storage-class.yaml"
	WAIT_FOR_GET_CHART_FORMAT         = "Waiting for downloading chart to complete"
	WAIT_FOR_GET_CHART_SUCCESS        = "Downloading chart completed successfully"
	WAIT_FOR_GET_CHART_FAILURE        = "Chart download failed:"
	WAIT_FOR_GET_CHART_TIMEOUT        = "Chart download timeout"
	LEGACY_KERNEL_MODE_MESSAGE_FORMAT = "Kernel is outdated, agent deployment in legacy mode.\n   Additional protocol support and a reduced footprint can be achieved on %s kernel"
	GET_CHART_POLLING_RETRIES         = 10
	GET_CHART_POLLING_INTERVAL        = time.Second * 1
	GET_CHART_POLLING_TIMEOUT         = time.Second * 10
	LEGACY_MODE                       = "legacy"
	STABLE_MODE                       = "stable"

	NODES_VALIDATION_EVENT_NAME     = "nodes_validation"
	HELM_INSTALLATION_EVENT_NAME    = "helm_installation"
	CLUSTER_VALIDATION_EVENT_NAME   = "cluster_validation"
	CLUSTER_REGISTRATION_EVENT_NAME = "cluster_registration"
)

func init() {
	RootCmd.AddCommand(DeployCmd)

	DeployCmd.PersistentFlags().StringSliceP(VALUES_FLAG, "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	viper.BindPFlag(VALUES_FLAG, DeployCmd.PersistentFlags().Lookup(VALUES_FLAG))

	DeployCmd.PersistentFlags().String(MODE_FLAG, "", "deployment mode [options: stable, legacy, experimental]")
	viper.BindPFlag(MODE_FLAG, DeployCmd.PersistentFlags().Lookup(MODE_FLAG))

	DeployCmd.PersistentFlags().String(REGISTRY_FLAG, "ecr", "image registry [options: ecr, quay]")
	viper.BindPFlag(REGISTRY_FLAG, DeployCmd.PersistentFlags().Lookup(REGISTRY_FLAG))

	DeployCmd.PersistentFlags().String(STORAGE_CLASS_FLAG, "", "override storage class")
	viper.BindPFlag(STORAGE_CLASS_FLAG, DeployCmd.PersistentFlags().Lookup(STORAGE_CLASS_FLAG))

	DeployCmd.PersistentFlags().Bool(LOW_RESOURCES_FLAG, false, "set low resources limits")
	viper.BindPFlag(LOW_RESOURCES_FLAG, DeployCmd.PersistentFlags().Lookup(LOW_RESOURCES_FLAG))

	DeployCmd.PersistentFlags().Bool(STORE_ISSUES_LOGS_ONLY_FLAG, false, "store issues logs only")
	viper.BindPFlag(STORE_ISSUES_LOGS_ONLY_FLAG, DeployCmd.PersistentFlags().Lookup(STORE_ISSUES_LOGS_ONLY_FLAG))

	DeployCmd.PersistentFlags().Bool(ENABLE_CUSTOM_METRICS_FLAG, false, "enable custom metrics scraping")
	viper.BindPFlag(ENABLE_CUSTOM_METRICS_FLAG, DeployCmd.PersistentFlags().Lookup(ENABLE_CUSTOM_METRICS_FLAG))

	DeployCmd.PersistentFlags().Bool(ENABLE_KUBE_STATE_METRICS_FLAG, false, "enable kube state metrics deployment")
	viper.BindPFlag(ENABLE_KUBE_STATE_METRICS_FLAG, DeployCmd.PersistentFlags().Lookup(ENABLE_KUBE_STATE_METRICS_FLAG))

	DeployCmd.PersistentFlags().String(COMMIT_HASH_KEY_NAME_FLAG, "", "the annotation/label key name that contains the app git commit hash")
	viper.BindPFlag(COMMIT_HASH_KEY_NAME_FLAG, DeployCmd.PersistentFlags().Lookup(COMMIT_HASH_KEY_NAME_FLAG))

	DeployCmd.PersistentFlags().String(REPOSITORY_URL_KEY_NAME_FLAG, "", "the annotation key name that contains the app git repository url")
	viper.BindPFlag(REPOSITORY_URL_KEY_NAME_FLAG, DeployCmd.PersistentFlags().Lookup(REPOSITORY_URL_KEY_NAME_FLAG))

	DeployCmd.PersistentFlags().String(VERSION_FLAG, "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is not specified, the latest version is used")
	viper.BindPFlag(VERSION_FLAG, DeployCmd.PersistentFlags().Lookup(VERSION_FLAG))
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy groundcover",
	RunE:  runDeployCmd,
}

func runDeployCmd(cmd *cobra.Command, args []string) error {
	var err error

	ctx := cmd.Context()
	isAuthenticated := !viper.IsSet(TOKEN_FLAG)
	namespace := viper.GetString(NAMESPACE_FLAG)
	kubeconfig := viper.GetString(KUBECONFIG_FLAG)
	kubecontext := viper.GetString(KUBECONTEXT_FLAG)
	releaseName := viper.GetString(HELM_RELEASE_FLAG)
	installationId := viper.GetString(INSTALLATION_ID_FLAG)

	sentryKubeContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)
	sentryKubeContext.SetOnCurrentScope()

	var tenantUUID string
	if tenantUUID = viper.GetString(TENANT_UUID_FLAG); isAuthenticated && tenantUUID == "" {
		var tenant *api.TenantInfo
		if tenant, err = fetchTenant(); err != nil {
			return err
		}

		if tenant == nil {
			return errors.New("tenant not found")
		}

		tenantUUID = tenant.UUID
	}

	var apiKey string
	if apiKey = viper.GetString(API_KEY_FLAG); apiKey == "" {
		var authApiKey *auth.ApiKey
		if authApiKey, err = fetchApiKey(tenantUUID); err != nil {
			return err
		}

		apiKey = authApiKey.ApiKey
	}

	var kubeClient *k8s.Client
	if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
		return err
	}

	if err = validateCluster(ctx, kubeClient, namespace, sentryKubeContext); err != nil {
		return err
	}

	var nodesReport *k8s.NodesReport
	if nodesReport, err = validateNodes(ctx, kubeClient, sentryKubeContext); err != nil {
		return err
	}

	var clusterName string
	if clusterName, err = getClusterName(kubeClient); err != nil {
		return err
	}

	sentryHelmContext := sentry_utils.NewHelmContext(releaseName, CHART_NAME, HELM_REPO_URL)
	sentryHelmContext.SetOnCurrentScope()
	sentry_utils.SetTagOnCurrentScope(sentry_utils.CLUSTER_NAME_TAG, clusterName)
	sentry_utils.SetTagOnCurrentScope(sentry_utils.NODES_COUNT_TAG, fmt.Sprintf("%d", nodesReport.NodesCount()))

	var helmClient *helm.Client
	if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
		return err
	}

	var chart *helm.Chart
	if chart, err = pollGetChart(ctx, helmClient, sentryHelmContext); err != nil {
		return err
	}

	deployableNodes, tolerations, err := getDeployableNodesAndTolerations(nodesReport, sentryKubeContext)
	if err != nil {
		return err
	}
	sentry_utils.SetTagOnCurrentScope(sentry_utils.EXPECTED_NODES_COUNT_TAG, fmt.Sprintf("%d", len(deployableNodes)))

	var isUpgrade bool
	var release *helm.Release
	if release, isUpgrade, err = helmClient.IsReleaseInstalled(releaseName); err != nil {
		return err
	}

	var chartValues map[string]interface{}
	if isUpgrade {
		chartValues = release.Config
	}

	if chartValues, err = generateChartValues(chartValues, apiKey, installationId, clusterName, deployableNodes, tolerations, nodesReport, sentryHelmContext); err != nil {
		return err
	}

	backendEnabled := true
	if backendValues, ok := chartValues["backend"]; ok {
		if isEnabled, ok := backendValues.(map[string]interface{})["enabled"]; ok {
			if enabled, ok := isEnabled.(bool); ok && !enabled {
				backendEnabled = false
			}
		}
	}

	agentEnabled := true
	if agentValues, ok := chartValues["agent"]; ok {
		if isEnabled, ok := agentValues.(map[string]interface{})["enabled"]; ok {
			if enabled, ok := isEnabled.(bool); ok && !enabled {
				agentEnabled = false
			}
		}
	}

	var shouldInstall bool
	if shouldInstall, err = promptInstallSummary(isUpgrade, releaseName, clusterName, namespace, release, chart, len(deployableNodes), nodesReport.NodesCount(), sentryHelmContext); err != nil {
		return err
	}

	if !shouldInstall {
		return ErrExecutionAborted
	}

	if err = installHelmRelease(ctx, helmClient, releaseName, chart, chartValues); err != nil {
		return err
	}

	if err = validateInstall(ctx, kubeClient, releaseName, namespace, chart.AppVersion(), tenantUUID, clusterName, len(deployableNodes), isAuthenticated, agentEnabled, backendEnabled, sentryHelmContext); err != nil {
		return err
	}

	printOrOpenClusterUrl(clusterName, namespace, isAuthenticated)

	ui.GlobalWriter.PrintlnWithPrefixln(JOIN_SLACK_MESSAGE)

	return nil
}

func validateCluster(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryKubeContext *sentry_utils.KubeContext) error {
	var err error

	event := segment.NewEvent(CLUSTER_VALIDATION_EVENT_NAME)
	defer func() {
		event.StatusByError(err)
	}()

	ui.GlobalWriter.PrintlnWithPrefixln("Validating cluster compatibility:")

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

	clusterReport.PrintStatus()

	if clusterReport.IsLocalCluster() {
		viper.Set(LOW_RESOURCES_FLAG, true)
	}

	if !clusterReport.IsCompatible {
		err = errors.New("can't continue with installation, cluster is not compatible for installation. Check solutions suggested by the CLI")
		return err
	}

	return nil
}

func validateNodes(ctx context.Context, kubeClient *k8s.Client, sentryKubeContext *sentry_utils.KubeContext) (*k8s.NodesReport, error) {
	var err error

	event := segment.NewEvent(NODES_VALIDATION_EVENT_NAME)
	defer func() {
		event.StatusByError(err)
	}()

	ui.GlobalWriter.PrintlnWithPrefixln("Validating cluster nodes compatibility:")

	var nodesSummaries []*k8s.NodeSummary
	if nodesSummaries, err = kubeClient.GetNodesSummaries(ctx); err != nil {
		return nil, err
	}

	sentryKubeContext.NodesCount = len(nodesSummaries)
	sentryKubeContext.SetOnCurrentScope()

	nodesReport := k8s.DefaultNodeRequirements.GenerateNodeReport(nodesSummaries)

	event.
		Set("nodesCount", len(nodesSummaries)).
		Set("taintedNodesCount", len(nodesReport.TaintedNodes)).
		Set("compatibleNodesCount", len(nodesReport.CompatibleNodes)).
		Set("incompatibleNodesCount", len(nodesReport.IncompatibleNodes))

	sentryKubeContext.SetNodesSamples(nodesReport)
	sentryKubeContext.SetOnCurrentScope()

	nodesReport.PrintStatus()

	if len(nodesReport.CompatibleNodes) == 0 || nodesReport.Schedulable.IsNonCompatible || nodesReport.ArchitectureAllowed.IsNonCompatible {
		return nil, errors.New("can't continue with installation, no compatible nodes for installation")
	}

	return nodesReport, nil
}

func getDeployableNodesAndTolerations(nodesReport *k8s.NodesReport, sentryKubeContext *sentry_utils.KubeContext) ([]*k8s.NodeSummary, []map[string]interface{}, error) {
	var err error

	tolerations := make([]map[string]interface{}, 0)
	deployableNodes := nodesReport.CompatibleNodes

	if len(nodesReport.TaintedNodes) > 0 {
		tolerationManager := &k8s.TolerationManager{
			TaintedNodes: nodesReport.TaintedNodes,
		}

		var allowedTaints []string
		if allowedTaints, err = promptTaints(tolerationManager, sentryKubeContext); err != nil {
			return nil, nil, err
		}

		if tolerations, err = tolerationManager.GetTolerationsMap(allowedTaints); err != nil {
			return nil, nil, err
		}

		var tolerableNodes []*k8s.NodeSummary
		if tolerableNodes, err = tolerationManager.GetTolerableNodes(allowedTaints); err != nil {
			return nil, nil, err
		}

		deployableNodes = append(deployableNodes, tolerableNodes...)
	}

	return deployableNodes, tolerations, nil
}

func promptTaints(tolerationManager *k8s.TolerationManager, sentryKubeContext *sentry_utils.KubeContext) ([]string, error) {
	var err error

	var taints []string
	if taints, err = tolerationManager.GetTaints(); err != nil {
		return nil, err
	}

	allowedTaints := ui.GlobalWriter.MultiSelectPrompt("Do you want set tolerations to allow scheduling groundcover on following taints:", taints, taints)

	sentryKubeContext.TolerationsAndTaintsRatio = fmt.Sprintf("%d/%d", len(allowedTaints), len(taints))
	sentryKubeContext.SetOnCurrentScope()
	sentry_utils.SetTagOnCurrentScope(sentry_utils.TAINTED_TAG, "true")

	return allowedTaints, nil
}

func promptInstallSummary(isUpgrade bool, releaseName string, clusterName string, namespace string, release *helm.Release, chart *helm.Chart, deployableNodesCount, nodesCount int, sentryHelmContext *sentry_utils.HelmContext) (bool, error) {
	ui.GlobalWriter.PrintlnWithPrefixln("Installing groundcover:")

	var promptMessage string
	if isUpgrade {
		sentryHelmContext.Upgrade = isUpgrade
		sentryHelmContext.PreviousChartVersion = release.Version().String()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.UPGRADE_TAG, strconv.FormatBool(isUpgrade))
		sentryHelmContext.SetOnCurrentScope()

		if chart.Version().GT(release.Version()) {
			promptMessage = fmt.Sprintf(
				"Your groundcover version is out of date (cluster: %s, namespace: %s, version: %s), The latest version is %s.\nDo you want to upgrade?",
				clusterName, namespace, release.Version(), chart.Version(),
			)
		} else {
			promptMessage = fmt.Sprintf(
				"Latest version of groundcover is already installed in your cluster! (cluster: %s, namespace: %s, version: %s).\nDo you want to redeploy?",
				clusterName, namespace, chart.Version(),
			)
		}
	} else {
		promptMessage = fmt.Sprintf(
			"Deploy groundcover (cluster: %s, namespace: %s, compatible nodes: %d/%d, version: %s)",
			clusterName, namespace, deployableNodesCount, nodesCount, chart.Version(),
		)
	}

	return ui.GlobalWriter.YesNoPrompt(promptMessage, !isUpgrade), nil
}

func installHelmRelease(ctx context.Context, helmClient *helm.Client, releaseName string, chart *helm.Chart, chartValues map[string]interface{}) error {
	var err error

	event := segment.NewEvent(HELM_INSTALLATION_EVENT_NAME)
	event.Set("chartVersion", chart.Version())
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	spinner := ui.GlobalWriter.NewSpinner("Installing groundcover helm release")
	spinner.Start()
	spinner.SetStopMessage("groundcover helm release is installed")
	spinner.SetStopFailMessage("groundcover helm release installation failed")
	defer spinner.WriteStop()

	helmUpgradeFunc := func() error {
		if _, err = helmClient.Upgrade(ctx, releaseName, chart, chartValues); err != nil {
			return ui.RetryableError(err)
		}

		return nil
	}

	err = spinner.Poll(ctx, helmUpgradeFunc, HELM_DEPLOY_POLLING_INTERVAL, HELM_DEPLOY_POLLING_TIMEOUT, HELM_DEPLOY_POLLING_RETRIES)

	if err == nil {
		return nil
	}

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		sentry_utils.SetLevelOnCurrentScope(sentry.LevelWarning)
		spinner.SetWarningSign()
		spinner.SetStopFailMessage("Timeout waiting for helm release installation")
		spinner.WriteStopFail()
		return nil
	}

	spinner.WriteStopFail()
	return err
}

func validateInstall(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace, appVersion, tenantUUID, clusterName string, deployableNodesCount int, isAuthenticated, agentEnabled, backendEnabled bool, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	defer reportPodsStatus(ctx, kubeClient, namespace, sentryHelmContext)

	ui.GlobalWriter.PrintlnWithPrefixln("Validating groundcover installation:")

	if err = waitForPvcs(ctx, kubeClient, releaseName, namespace, sentryHelmContext); err != nil {
		return err
	}

	if backendEnabled {
		if err = waitForPortal(ctx, kubeClient, namespace, appVersion, sentryHelmContext); err != nil {
			return err
		}
	}

	if isAuthenticated {
		if err = validateClusterRegistered(ctx, tenantUUID, clusterName); err != nil {
			return err
		}
	}

	if agentEnabled {
		if err = waitForAlligators(ctx, kubeClient, namespace, appVersion, deployableNodesCount, sentryHelmContext); err != nil {
			return err
		}
	}

	ui.GlobalWriter.PrintlnWithPrefixln("That was easy. groundcover installed!")

	return nil
}

func validateClusterRegistered(ctx context.Context, tenantUUID, clusterName string) error {
	var err error

	event := segment.NewEvent(CLUSTER_REGISTRATION_EVENT_NAME)
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return err
	}

	apiClient := api.NewClient(auth0Token)

	if err = apiClient.PollIsClusterExist(ctx, tenantUUID, clusterName); err != nil {
		return err
	}

	return nil
}

func printOrOpenClusterUrl(clusterName string, namespace string, isAuthenticated bool) {
	clusterUrl := fmt.Sprintf(CLUSTER_URL_FORMAT, GROUNDCOVER_URL, clusterName)
	clusterUrlLink := ui.GlobalWriter.UrlLink(clusterUrl)

	if isAuthenticated {
		utils.TryOpenBrowser(ui.GlobalWriter, "Check out: ", clusterUrl)
	} else {
		ui.GlobalWriter.Printf("Return to browser tab or visit %s if you closed tab\n", clusterUrlLink)
	}
}

func getClusterName(kubeClient *k8s.Client) (string, error) {
	var err error
	var clusterName string

	if clusterName = viper.GetString(CLUSTER_NAME_FLAG); clusterName != "" {
		return clusterName, nil
	}

	if clusterName, err = kubeClient.GetClusterShortName(); err != nil {
		return "", err
	}

	return clusterName, nil
}

func pollGetChart(ctx context.Context, helmClient *helm.Client, sentryHelmContext *sentry_utils.HelmContext) (*helm.Chart, error) {
	spinner := ui.GlobalWriter.NewSpinner(WAIT_FOR_GET_CHART_FORMAT)
	spinner.SetStopMessage(WAIT_FOR_GET_CHART_SUCCESS)
	spinner.SetStopFailMessage(WAIT_FOR_GET_CHART_FAILURE)

	spinner.Start()
	defer spinner.WriteStop()

	chartVersion := viper.GetString(VERSION_FLAG)

	var chart *helm.Chart
	var err error
	getChartFunc := func() error {
		if err := helmClient.AddRepo(HELM_REPO_NAME, HELM_REPO_URL); err != nil {
			return ui.RetryableError(err)
		}

		if chart, err = helmClient.GetChart(CHART_NAME, chartVersion); err != nil {
			return err
		}

		return nil
	}

	err = spinner.Poll(ctx, getChartFunc, GET_CHART_POLLING_INTERVAL, GET_CHART_POLLING_TIMEOUT, GET_CHART_POLLING_RETRIES)

	if err == nil {
		sentryHelmContext.ChartVersion = chart.Version().String()
		sentryHelmContext.SetOnCurrentScope()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

		return chart, nil
	}

	spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return nil, errors.New(WAIT_FOR_GET_CHART_TIMEOUT)
	}

	return nil, err
}

func generateChartValues(chartValues map[string]interface{}, apiKey, installationId, clusterName string, deployableNodes []*k8s.NodeSummary, tolerations []map[string]interface{}, nodesReport *k8s.NodesReport, sentryHelmContext *sentry_utils.HelmContext) (map[string]interface{}, error) {
	var err error

	defaultChartValues := map[string]interface{}{
		"clusterId":            clusterName,
		"installationId":       installationId,
		"commitHashKeyName":    viper.GetString(COMMIT_HASH_KEY_NAME_FLAG),
		"repositoryUrlKeyName": viper.GetString(REPOSITORY_URL_KEY_NAME_FLAG),
		"global":               map[string]interface{}{"groundcover_token": apiKey},
	}

	if nodesReport.IsLegacyKernel() {
		ui.GlobalWriter.PrintWarningMessageln(fmt.Sprintf(LEGACY_KERNEL_MODE_MESSAGE_FORMAT, k8s.StableKernelVersionRange))
		viper.Set(MODE_FLAG, LEGACY_MODE)
	}

	if mode := viper.GetString(MODE_FLAG); mode != "" {
		defaultChartValues["mode"] = mode
		sentry_utils.SetTagOnCurrentScope(sentry_utils.MODE_TAG, mode)
	}

	if err = mergo.Merge(&chartValues, defaultChartValues, mergo.WithSliceDeepCopy); err != nil {
		return nil, err
	}

	var overridePaths []string
	allocatableResources := helm.CalcAllocatableResources(deployableNodes)
	sentryHelmContext.AllocatableResources = allocatableResources
	if viper.GetBool(LOW_RESOURCES_FLAG) {
		overridePaths = []string{
			helm.AGENT_LOW_RESOURCES_PATH,
			helm.BACKEND_LOW_RESOURCES_PATH,
		}
	} else {
		agentPresetPath := helm.GetAgentResourcePresetPath(allocatableResources)
		if agentPresetPath != helm.DEFAULT_PRESET {
			overridePaths = append(overridePaths, agentPresetPath)
		}

		backendPresetPath := helm.GetBackendResourcePresetPath(allocatableResources)
		if backendPresetPath != helm.DEFAULT_PRESET {
			overridePaths = append(overridePaths, backendPresetPath)
		}
	}

	if viper.GetString(REGISTRY_FLAG) == "quay" {
		overridePaths = append(overridePaths, QUAY_REGISTRY_PRESET_PATH)
	}

	enableCustomMetrics := viper.GetBool(ENABLE_CUSTOM_METRICS_FLAG)
	if enableCustomMetrics {
		overridePaths = append(overridePaths, CUSTOM_METRICS_PRESET_PATH)
	}

	enableKubeStateMetrics := viper.GetBool(ENABLE_KUBE_STATE_METRICS_FLAG)
	if enableKubeStateMetrics {
		overridePaths = append(overridePaths, KUBE_STATE_METRICS_PRESET_PATH)
	}

	if semver.MustParseRange(">=5.11.0")(nodesReport.MaximalKernelVersion()) {
		overridePaths = append(overridePaths, AGENT_KERNEL_5_11_PRESET_PATH)
	}

	if len(overridePaths) > 0 {
		sentryHelmContext.ResourcesPresets = overridePaths
		sentryHelmContext.SetOnCurrentScope()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.DEFAULT_RESOURCES_PRESET_TAG, "false")
	} else {
		sentry_utils.SetTagOnCurrentScope(sentry_utils.DEFAULT_RESOURCES_PRESET_TAG, "true")
	}

	userValuesOverridePaths := viper.GetStringSlice(VALUES_FLAG)
	overridePaths = append(overridePaths, userValuesOverridePaths...)

	templateValues := helm.TemplateValues{}
	if storageClassName := viper.GetString(STORAGE_CLASS_FLAG); storageClassName != "" {
		templateValues.StorageClassName = storageClassName
		overridePaths = append(overridePaths, STORAGE_CLASS_TEMPLATE_PATH)
	}

	var valuesOverride map[string]interface{}
	if valuesOverride, err = helm.GetChartValuesOverrides(overridePaths, &templateValues); err != nil {
		return nil, err
	}

	valuesOverride[STORE_ISSUES_LOGS_ONLY_KEY] = viper.GetBool(STORE_ISSUES_LOGS_ONLY_FLAG)

	// we always want to override tolerations
	if agentValues, exist := valuesOverride["agent"].(map[string]interface{}); exist {
		if _, exist := agentValues["tolerations"]; exist {
			tolerations = []map[string]interface{}{}
		}
	}

	if agentValues, ok := chartValues["agent"].(map[string]interface{}); ok {
		agentValues["tolerations"] = tolerations
	} else {
		chartValues["agent"] = map[string]interface{}{"tolerations": tolerations}
	}

	if err = mergo.Merge(&chartValues, valuesOverride, mergo.WithOverride); err != nil {
		return nil, err
	}

	sentryHelmContext.ValuesOverride = valuesOverride
	sentryHelmContext.SetOnCurrentScope()

	return chartValues, nil
}
