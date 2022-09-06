package cmd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/utils"
)

const (
	VALUES_FLAG                   = "values"
	CHART_NAME                    = "groundcover"
	DEFAULT_GROUNDCOVER_RELEASE   = "groundcover"
	DEFAULT_GROUNDCOVER_NAMESPACE = "groundcover"
	COMMIT_HASH_KEY_NAME_FLAG     = "git-commit-hash-key-name"
	REPOSITORY_URL_KEY_NAME_FLAG  = "git-repository-url-key-name"
	GROUNDCOVER_URL               = "https://app.groundcover.com"
	HELM_REPO_URL                 = "https://helm.groundcover.com"
)

func init() {
	RootCmd.AddCommand(DeployCmd)

	DeployCmd.PersistentFlags().StringSliceP(VALUES_FLAG, "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	viper.BindPFlag(VALUES_FLAG, DeployCmd.PersistentFlags().Lookup(VALUES_FLAG))

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

	namespace := viper.GetString(NAMESPACE_FLAG)
	kubeconfig := viper.GetString(KUBECONFIG_FLAG)
	kubecontext := viper.GetString(KUBECONTEXT_FLAG)
	releaseName := viper.GetString(HELM_RELEASE_FLAG)

	var auth0Token auth.Auth0Token
	if err = auth0Token.Load(); err != nil {
		return err
	}

	var apiKey api.ApiKey
	if err = apiKey.Load(); err != nil {
		return err
	}

	sentryKubeContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext, namespace)
	sentryKubeContext.SetOnCurrentScope()

	var kubeClient *k8s.Client
	if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
		return err
	}

	if err = validateCluster(cmd.Context(), kubeClient, namespace, sentryKubeContext); err != nil {
		return err
	}

	var clusterName string
	if clusterName, err = getClusterName(kubeClient); err != nil {
		return err
	}

	sentryKubeContext.Cluster = clusterName
	sentryKubeContext.SetOnCurrentScope()

	sentryHelmContext := sentry_utils.NewHelmContext(releaseName, CHART_NAME, HELM_REPO_URL)
	sentryHelmContext.SetOnCurrentScope()

	var helmClient *helm.Client
	if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
		return err
	}

	shouldRedeploy, err := checkIfRedeployWanted(helmClient, releaseName, sentryHelmContext, clusterName, namespace)
	if err != nil {
		return err
	}

	if !shouldRedeploy {
		return nil
	}

	var nodesSummeries []k8s.NodeSummary
	if nodesSummeries, err = kubeClient.GetNodesSummeries(cmd.Context()); err != nil {
		return err
	}

	nodesCount := len(nodesSummeries)
	compatible, err := checkClusterNodes(sentryKubeContext, nodesCount, nodesSummeries)
	if err != nil {
		return err
	}

	expectedAlligatorsCount, err := helmInstallation(cmd.Context(), helmClient, sentryHelmContext, clusterName, apiKey, compatible, releaseName, namespace, nodesCount, kubeClient)
	if err != nil {
		return err
	}

	_, err = watchAlligators(helmClient, releaseName, cmd, kubeClient, expectedAlligatorsCount)
	if err != nil {
		return err
	}

	apiClient := api.NewClient(&auth0Token)
	if err = apiClient.PollIsClusterExist(clusterName); err != nil {
		return err
	}

	utils.TryOpenBrowser(fmt.Sprintf("%s/?clusterId=%s&viewType=Overview", GROUNDCOVER_URL, clusterName))
	sentry.CaptureMessage("deploy executed successfully")
	return nil
}

func watchAlligators(helmClient *helm.Client, releaseName string, cmd *cobra.Command, kubeClient *k8s.Client, expectedAlligatorsCount int) (bool, error) {
	release, err := helmClient.GetCurrentRelease(releaseName)
	if err != nil {
		return false, err
	}

	err = waitForAlligators(cmd.Context(), kubeClient, release, expectedAlligatorsCount)
	if err != nil {
		return false, err
	}

	return true, nil
}

func validateCluster(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryKubeContext *sentry_utils.KubeContext) error {
	var err error

	fmt.Println("Validating cluster is compatible with groundcover installation:")
	clusterReport := k8s.DefaultClusterRequirements.Validate(ctx, kubeClient, namespace)

	sentryKubeContext.ClusterReport = clusterReport
	sentryKubeContext.SetOnCurrentScope()

	if !clusterReport.ServerVersionAllowed.IsCompatible {
		return fmt.Errorf(clusterReport.ServerVersionAllowed.Message)
	}
	fmt.Println(clusterReport.ServerVersionAllowed.Message)

	if sentryKubeContext.ServerVersion, err = kubeClient.Discovery().ServerVersion(); err != nil {
		return err
	}

	if !clusterReport.UserAuthorized.IsCompatible {
		return fmt.Errorf(clusterReport.UserAuthorized.Message)
	}
	fmt.Println(clusterReport.UserAuthorized.Message)

	return nil
}

func checkIfRedeployWanted(helmClient *helm.Client, releaseName string, sentryHelmContext *sentry_utils.HelmContext, clusterName string, namespace string) (bool, error) {
	installed, _ := checkIfGroundcoverAlreadyInstalled(helmClient, releaseName)
	if !installed {
		return true, nil
	}

	release, err := helmClient.GetCurrentRelease(releaseName)
	if err != nil {
		return false, err
	}

	sentryHelmContext.PreviousChartVersion = release.Version().String()
	sentryHelmContext.SetOnCurrentScope()

	chart, err := helmClient.GetLatestChart(CHART_NAME, HELM_REPO_URL)
	if err != nil {
		return false, err
	}

	var promptMessage string
	if chart.Version().GT(release.Version()) {
		promptMessage = fmt.Sprintf(
			"Current groundcover installation in your cluster is out of date! (cluster: %s, namespace: %s, version: %s), The latest version is %s.\nDo you want to upgrade?",
			clusterName, namespace, release.Version(), chart.Version(),
		)
	} else {
		promptMessage = fmt.Sprintf(
			"Current groundcover installation in your cluster is latest (cluster: %s, namespace: %s, version: %s) .\nDo you want to redeploy?",
			clusterName, namespace, chart.Version(),
		)
	}

	if !utils.YesNoPrompt(promptMessage, false) {
		sentry.CaptureMessage("deploy execution aborted")
		return false, fmt.Errorf("deploy execution aborted")
	}

	return true, nil
}

func checkIfGroundcoverAlreadyInstalled(helmClient *helm.Client, releaseName string) (bool, error) {
	return helmClient.IsReleaseInstalled(releaseName)
}

func helmInstallation(ctx context.Context,
	helmClient *helm.Client,
	sentryHelmContext *sentry_utils.HelmContext,
	clusterName string,
	apiKey api.ApiKey,
	compatible []*k8s.NodeReport,
	releaseName string,
	namespace string,
	nodesCount int,
	kubeClient *k8s.Client) (int, error) {
	var chart *helm.Chart
	var err error
	if chart, err = helmClient.GetLatestChart(CHART_NAME, HELM_REPO_URL); err != nil {
		return 0, err
	}

	sentryHelmContext.ChartVersion = chart.Version().String()
	sentryHelmContext.SetOnCurrentScope()
	sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

	chartValues := defaultChartValues(clusterName, apiKey.ApiKey)
	userValuesOverridePaths := viper.GetStringSlice(VALUES_FLAG)

	var resourcesTunerPresetPaths []string
	if resourcesTunerPresetPaths, err = helm.GetResourcesTunerPresetPaths(compatible); err != nil {
		return 0, err
	}

	sentryHelmContext.ResourcesPresets = resourcesTunerPresetPaths
	sentryHelmContext.SetOnCurrentScope()

	var valuesOverride map[string]interface{}
	if valuesOverride, err = helm.SetChartValuesOverrides(&chartValues, append(resourcesTunerPresetPaths, userValuesOverridePaths...)); err != nil {
		return 0, err
	}

	sentryHelmContext.ValuesOverride = valuesOverride
	sentryHelmContext.SetOnCurrentScope()

	var isUpgrade bool
	if isUpgrade, err = helmClient.IsReleaseInstalled(releaseName); err != nil {
		return 0, err
	}

	sentryHelmContext.Upgrade = isUpgrade
	sentryHelmContext.SetOnCurrentScope()

	expectedAlligatorsCount := len(compatible)
	promptMessage := fmt.Sprintf(
		"Deploying groundcover (cluster: %s, namespace: %s, compatible nodes: %d/%d, version: %s).\nDo you want to deploy?",
		clusterName, namespace, expectedAlligatorsCount, nodesCount, chart.Version(),
	)

	if !utils.YesNoPrompt(promptMessage, false) {
		sentry.CaptureMessage("deploy execution aborted")
		return 0, fmt.Errorf("deploy execution aborted")
	}

	if err = helmClient.Upgrade(ctx, releaseName, chart, chartValues); err != nil {
		return 0, err
	}

	return expectedAlligatorsCount, nil
}

func checkClusterNodes(sentryKubeContext *sentry_utils.KubeContext, nodesCount int, nodesSummeries []k8s.NodeSummary) ([]*k8s.NodeReport, error) {
	sentryKubeContext.NodesCount = nodesCount
	sentryKubeContext.SetOnCurrentScope()

	compatible, incompatible := k8s.DefaultNodeRequirements.GenerateNodeReports(nodesSummeries)
	nodes := append(compatible, incompatible...)

	sentryKubeContext.SetNodeReportsSamples(compatible)
	sentryKubeContext.SetOnCurrentScope()

	if len(incompatible) > 0 {
		sentry_utils.SetLevelOnCurrentScope(sentry.LevelWarning)
		sentryKubeContext.IncompatibleNodeReports = incompatible
		sentryKubeContext.SetOnCurrentScope()
	}

	if !validateCompatibleNodes(nodes) {
		return nil, fmt.Errorf("can't continue with installation, no compatible nodes for installation")
	}

	return compatible, nil
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

func defaultChartValues(clusterName, apikey string) map[string]interface{} {
	chartValues := make(map[string]interface{})
	chartValues["clusterId"] = clusterName
	chartValues["origin"] = map[string]interface{}{"tag": ""}
	chartValues["global"] = map[string]interface{}{"groundcover_token": apikey}
	chartValues["commitHashKeyName"] = viper.GetString(COMMIT_HASH_KEY_NAME_FLAG)
	chartValues["repositoryUrlKeyName"] = viper.GetString(REPOSITORY_URL_KEY_NAME_FLAG)

	return chartValues
}

func validateCompatibleNodes(nodes []*k8s.NodeReport) bool {
	fmt.Println("Validating cluster nodes are compatible with groundcover installation:")

	return hasAllowedKernelVersions(nodes) &&
		hasCpuSufficient(nodes) &&
		hasMemorySufficient(nodes) &&
		hasProviderAllowed(nodes) &&
		hasArchitectureAllowed(nodes) &&
		hasOperatingSystemAllowed(nodes)
}

func hasAllowedKernelVersions(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "KernelVersionAllowed")
	fmt.Printf("Kernel version > %s (%d/%d)\n", k8s.MinimumKernelVersionSupport.String(), allowedCount, len(nodes))
	return allowedCount > 0
}

func hasCpuSufficient(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "CpuSufficient")
	fmt.Printf("Sufficient CPU > %s (%d/%d)\n", k8s.NodeMinimumCpuRequired.String(), allowedCount, len(nodes))
	return allowedCount > 0
}

func hasMemorySufficient(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "MemorySufficient")
	fmt.Printf("Sufficient Memory > %s (%d/%d)\n", k8s.NodeMinimumMemoryRequired.String(), allowedCount, len(nodes))
	return allowedCount > 0
}

func hasProviderAllowed(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "ProviderAllowed")
	fmt.Printf("Provider Allowed (%d/%d)\n", allowedCount, len(nodes))
	return allowedCount > 0
}

func hasArchitectureAllowed(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "ArchitectureAllowed")
	fmt.Printf("Architecture Allowed (%d/%d)\n", allowedCount, len(nodes))
	return allowedCount > 0
}

func hasOperatingSystemAllowed(nodes []*k8s.NodeReport) bool {
	allowedCount := isNodePropertySupported(nodes, "OperatingSystemAllowed")
	fmt.Printf("Operating System Allowed (%d/%d)\n", allowedCount, len(nodes))
	return allowedCount > 0
}

func isNodePropertySupported(nodes []*k8s.NodeReport, propertyName string) int {
	var allowedCount int
	for _, node := range nodes {
		obj := reflect.ValueOf(node)
		fieldValue := reflect.Indirect(obj).FieldByName(propertyName)
		if fieldValue.FieldByName("IsCompatible").Bool() {
			allowedCount++
		}
	}

	return allowedCount
}
