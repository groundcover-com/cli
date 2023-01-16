package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"groundcover.com/pkg/utils"
	"k8s.io/utils/strings/slices"
)

const (
	HELM_DEPLOY_POLLING_RETIRES         = 2
	HELM_DEPLOY_POLLING_INTERVAL        = time.Second * 1
	HELM_DEPLOY_POLLING_TIMEOUT         = time.Minute * 5
	VALUES_FLAG                         = "values"
	EXPERIMENTAL_FLAG                   = "experimental"
	LOW_RESOURCES_FLAG                  = "low-resources"
	STORE_ALL_LOG_FLAG                  = "store-all-logs"
	STORE_ALL_LOGS_KEY                  = "storeAllLogs"
	CHART_NAME                          = "groundcover/groundcover"
	HELM_REPO_NAME                      = "groundcover"
	DEFAULT_GROUNDCOVER_RELEASE         = "groundcover"
	DEFAULT_GROUNDCOVER_NAMESPACE       = "groundcover"
	COMMIT_HASH_KEY_NAME_FLAG           = "git-commit-hash-key-name"
	REPOSITORY_URL_KEY_NAME_FLAG        = "git-repository-url-key-name"
	GROUNDCOVER_URL                     = "https://app.groundcover.com"
	HELM_REPO_URL                       = "https://helm.groundcover.com"
	EXPERIMENTAL_PRESET_PATH            = "presets/agent/experimental.yaml"
	LOW_RESOURCES_NOTICE_MESSAGE_FORMAT = "We get it, you like things light 🪁\n   But since you’re deploying on a %s we’ll have to limit some of our features to make sure it’s smooth sailing.\n   For the full groundcover experience, try deploying on a different cluster\n"
	BELOW_RESOURCES_MESSAGE             = "🚨 We get it, you like things light 🪁 - but since you’re deploying on a cluster with extremely low resources we cannot deploy groundcover and provide a smooth sailing.\n     To check out groundcover, please try deploying on a different cluster.\n     Minimum CPU %s, Memory %s"
)

func init() {
	RootCmd.AddCommand(DeployCmd)

	DeployCmd.PersistentFlags().StringSliceP(VALUES_FLAG, "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	viper.BindPFlag(VALUES_FLAG, DeployCmd.PersistentFlags().Lookup(VALUES_FLAG))

	DeployCmd.PersistentFlags().Bool(EXPERIMENTAL_FLAG, false, "enable groundcover experimental features")
	viper.BindPFlag(EXPERIMENTAL_FLAG, DeployCmd.PersistentFlags().Lookup(EXPERIMENTAL_FLAG))

	DeployCmd.PersistentFlags().Bool(LOW_RESOURCES_FLAG, false, "set low resources limits")
	viper.BindPFlag(LOW_RESOURCES_FLAG, DeployCmd.PersistentFlags().Lookup(LOW_RESOURCES_FLAG))

	DeployCmd.PersistentFlags().Bool(STORE_ALL_LOG_FLAG, false, "store all logs")
	viper.BindPFlag(STORE_ALL_LOG_FLAG, DeployCmd.PersistentFlags().Lookup(STORE_ALL_LOG_FLAG))

	DeployCmd.PersistentFlags().String(COMMIT_HASH_KEY_NAME_FLAG, "", "the annotation/label key name that contains the app git commit hash")
	viper.BindPFlag(COMMIT_HASH_KEY_NAME_FLAG, DeployCmd.PersistentFlags().Lookup(COMMIT_HASH_KEY_NAME_FLAG))

	DeployCmd.PersistentFlags().String(REPOSITORY_URL_KEY_NAME_FLAG, "", "the annotation key name that contains the app git repository url")
	viper.BindPFlag(REPOSITORY_URL_KEY_NAME_FLAG, DeployCmd.PersistentFlags().Lookup(REPOSITORY_URL_KEY_NAME_FLAG))
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy groundcover",
	RunE:  runDeployCmd,
}

func runDeployCmd(cmd *cobra.Command, args []string) error {
	var err error

	ctx := cmd.Context()
	namespace := viper.GetString(NAMESPACE_FLAG)
	kubeconfig := viper.GetString(KUBECONFIG_FLAG)
	kubecontext := viper.GetString(KUBECONTEXT_FLAG)
	releaseName := viper.GetString(HELM_RELEASE_FLAG)

	var auth0Token auth.Auth0Token
	if err = auth0Token.Load(); err != nil {
		return err
	}

	sentryKubeContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)
	sentryKubeContext.SetOnCurrentScope()

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

	var helmClient *helm.Client
	if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
		return err
	}

	var chart *helm.Chart
	if chart, err = getLatestChart(helmClient, sentryHelmContext); err != nil {
		return err
	}

	deployableNodes, tolerations, err := getDeployableNodesAndTolerations(nodesReport, sentryKubeContext)
	if err != nil {
		return err
	}

	var isUpgrade bool
	var release *helm.Release
	if release, isUpgrade, err = helmClient.IsReleaseInstalled(releaseName); err != nil {
		return err
	}

	var chartValues map[string]interface{}
	if isUpgrade {
		chartValues = release.Config
	}

	if chartValues, err = generateChartValues(chartValues, clusterName, deployableNodes, tolerations, sentryHelmContext); err != nil {
		return err
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

	err = validateInstall(ctx, kubeClient, namespace, chart.AppVersion(), &auth0Token, clusterName, len(deployableNodes), sentryHelmContext)
	reportPodsStatus(ctx, kubeClient, namespace, sentryHelmContext)

	if err != nil {
		ui.GlobalWriter.PrintflnWithPrefixln("Installation takes longer then expected, you can check the status using \"kubectl get pods -n %s\"", namespace)
		ui.GlobalWriter.Printf("If pods in %q namespce are running, Check out: %s\n", namespace, ui.GlobalWriter.UrlLink(fmt.Sprintf("%s/?clusterId=%s&viewType=Overview\n", GROUNDCOVER_URL, clusterName)))
		ui.GlobalWriter.Printf("%s\n", SUPPORT_SLACK_MESSAGE)
		return err
	}

	ui.GlobalWriter.PrintlnWithPrefixln("That was easy. groundcover installed!")
	utils.TryOpenBrowser(*ui.GlobalWriter, "Check out:", fmt.Sprintf("%s/?clusterId=%s&viewType=Overview", GROUNDCOVER_URL, clusterName))
	ui.GlobalWriter.PrintlnWithPrefixln(JOIN_SLACK_MESSAGE)

	return nil
}

func validateCluster(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryKubeContext *sentry_utils.KubeContext) error {
	var err error

	ui.GlobalWriter.PrintlnWithPrefixln("Validating cluster compatibility:")

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

	clusterReport.PrintStatus()

	if clusterReport.IsLocalCluster() {
		viper.Set(LOW_RESOURCES_FLAG, true)
	}

	if !clusterReport.IsCompatible {
		return errors.New("can't continue with installation, cluster is not compatible for installation. Check solutions suggested by the CLI")
	}

	return nil
}

func validateNodes(ctx context.Context, kubeClient *k8s.Client, sentryKubeContext *sentry_utils.KubeContext) (*k8s.NodesReport, error) {
	var err error

	ui.GlobalWriter.PrintlnWithPrefixln("Validating cluster nodes compatibility:")

	var nodesSummeries []*k8s.NodeSummary
	if nodesSummeries, err = kubeClient.GetNodesSummeries(ctx); err != nil {
		return nil, err
	}

	sentryKubeContext.NodesCount = len(nodesSummeries)
	sentryKubeContext.SetOnCurrentScope()

	nodesReport := k8s.DefaultNodeRequirements.Validate(nodesSummeries)

	sentryKubeContext.SetNodesSamples(nodesReport)
	sentryKubeContext.SetOnCurrentScope()

	nodesReport.PrintStatus()

	if len(nodesReport.CompatibleNodes) == 0 || nodesReport.Schedulable.IsNonCompatible {
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

	spinner := ui.GlobalWriter.NewSpinner("Installing groundcover helm release")
	spinner.Start()
	spinner.StopMessage("groundcover helm release is installed")
	spinner.StopFailMessage("groundcover helm release installation failed")
	defer spinner.Stop()

	helmUpgradeFunc := func() error {
		if _, err = helmClient.Upgrade(ctx, releaseName, chart, chartValues); err != nil {
			return ui.RetryableError(err)
		}

		return nil
	}

	err = spinner.Poll(ctx, helmUpgradeFunc, HELM_DEPLOY_POLLING_INTERVAL, HELM_DEPLOY_POLLING_TIMEOUT, HELM_DEPLOY_POLLING_RETIRES)

	if err == nil {
		return nil
	}

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		sentry_utils.SetLevelOnCurrentScope(sentry.LevelWarning)
		spinner.SetWarningSign()
		spinner.StopFailMessage("Timeout waiting for helm release installation")
		spinner.StopFail()
		return nil
	}

	spinner.StopFail()

	return err
}

func validateInstall(ctx context.Context, kubeClient *k8s.Client, namespace, appVersion string, auth0Token *auth.Auth0Token, clusterName string, deployableNodesCount int, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	ui.GlobalWriter.PrintlnWithPrefixln("Validating groundcover installation:")

	if err = waitForPvcs(ctx, kubeClient, namespace, sentryHelmContext); err != nil {
		return err
	}

	if err = waitForPortal(ctx, kubeClient, namespace, appVersion, sentryHelmContext); err != nil {
		return err
	}

	if err = waitForAlligators(ctx, kubeClient, namespace, appVersion, deployableNodesCount, sentryHelmContext); err != nil {
		return err
	}

	apiClient := api.NewClient(auth0Token)
	if err = apiClient.PollIsClusterExist(ctx, clusterName); err != nil {
		return err
	}

	return nil
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

func getLatestChart(helmClient *helm.Client, sentryHelmContext *sentry_utils.HelmContext) (*helm.Chart, error) {
	var err error

	if err = helmClient.AddRepo(HELM_REPO_NAME, HELM_REPO_URL); err != nil {
		return nil, err
	}

	var chart *helm.Chart
	if chart, err = helmClient.GetLatestChart(CHART_NAME); err != nil {
		return nil, err
	}

	sentryHelmContext.ChartVersion = chart.Version().String()
	sentryHelmContext.SetOnCurrentScope()
	sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

	return chart, nil
}

func generateChartValues(chartValues map[string]interface{}, clusterName string, deployableNodes []*k8s.NodeSummary, tolerations []map[string]interface{}, sentryHelmContext *sentry_utils.HelmContext) (map[string]interface{}, error) {
	var err error

	var apiKey api.ApiKey
	if err = apiKey.Load(); err != nil {
		return nil, err
	}

	defaultChartValues := map[string]interface{}{
		"clusterId":            clusterName,
		"commitHashKeyName":    viper.GetString(COMMIT_HASH_KEY_NAME_FLAG),
		"repositoryUrlKeyName": viper.GetString(REPOSITORY_URL_KEY_NAME_FLAG),
		"global":               map[string]interface{}{"groundcover_token": apiKey.ApiKey},
	}

	// we always want to override tolerations
	agent, ok := chartValues["agent"]
	if ok {
		agentMap, ok := agent.(map[string]interface{})
		if ok {
			agentMap["tolerations"] = tolerations
		}
	} else {
		defaultChartValues["agent"] = map[string]interface{}{"tolerations": tolerations}
	}

	if err = mergo.Merge(&chartValues, defaultChartValues, mergo.WithSliceDeepCopy); err != nil {
		return nil, err
	}

	var overridePaths []string
	allocatableResources := helm.CalcAllocatableResources(deployableNodes)
	sentryHelmContext.AllocatableResources = allocatableResources

	if !helm.CanRunGroundcover(allocatableResources) {
		return nil, errors.Errorf(BELOW_RESOURCES_MESSAGE, helm.GROUNDCOVER_MINIUM_CPU, helm.GROUNDCOVER_MINIUM_MEMORY)
	}

	if viper.GetBool(LOW_RESOURCES_FLAG) {
		overridePaths = []string{
			helm.AGENT_LOW_RESOURCES_PATH,
			helm.BACKEND_LOW_RESOURCES_PATH,
		}
	} else {
		agentPresetPath := helm.GetAgentResourcePresetPath(allocatableResources)
		if agentPresetPath != helm.NO_PRESET {
			overridePaths = append(overridePaths, agentPresetPath)
		}

		backendPresetPath := helm.GetBackendResourcePresetPath(allocatableResources)
		if backendPresetPath != helm.NO_PRESET {
			overridePaths = append(overridePaths, backendPresetPath)
		}
	}

	useExperimental := viper.GetBool(EXPERIMENTAL_FLAG)
	if useExperimental {
		overridePaths = append(overridePaths, EXPERIMENTAL_PRESET_PATH)
	}

	if slices.Contains(overridePaths, helm.AGENT_LOW_RESOURCES_PATH) {
		clusterType := "low resources"

		for _, localClusterType := range k8s.LocalClusterTypes {
			if strings.HasPrefix(clusterName, localClusterType) {
				clusterType = localClusterType
				break
			}
		}

		ui.GlobalWriter.Println("")
		ui.GlobalWriter.PrintNoticeMessage(fmt.Sprintf(LOW_RESOURCES_NOTICE_MESSAGE_FORMAT, color.New().Add(color.Bold).Sprintf("%s cluster", clusterType)))
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

	var valuesOverride map[string]interface{}
	if valuesOverride, err = helm.GetChartValuesOverrides(overridePaths); err != nil {
		return nil, err
	}

	if viper.GetBool(STORE_ALL_LOG_FLAG) {
		valuesOverride[STORE_ALL_LOGS_KEY] = true
	}

	if err = mergo.Merge(&chartValues, valuesOverride, mergo.WithSliceDeepCopy); err != nil {
		return nil, err
	}

	sentryHelmContext.ValuesOverride = valuesOverride
	sentryHelmContext.SetOnCurrentScope()

	return chartValues, nil
}
