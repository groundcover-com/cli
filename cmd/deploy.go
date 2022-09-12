package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
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

	var chartValues map[string]interface{}
	if chartValues, err = getChartValues(clusterName, nodesReport.CompatibleNodes, sentryHelmContext); err != nil {
		return err
	}

	var shouldInstall bool
	if shouldInstall, err = promptInstallSummary(helmClient, releaseName, clusterName, namespace, chart, nodesReport, sentryHelmContext); err != nil {
		return err
	}

	if !shouldInstall {
		sentry.CaptureMessage("deploy execution aborted")
		return nil
	}

	var release *helm.Release
	if release, err = installHelmRelease(ctx, helmClient, releaseName, chart, chartValues); err != nil {
		return err
	}

	if err = validateInstall(ctx, kubeClient, release, clusterName, len(nodesReport.CompatibleNodes), sentryHelmContext); err != nil {
		return errors.Wrap(err, "Helm installation validation failed")
	}

	fmt.Println("\nThat was easy. groundcover installed!")
	utils.TryOpenBrowser("Check out:", fmt.Sprintf("%s/?clusterId=%s&viewType=Overview", GROUNDCOVER_URL, clusterName))
	fmt.Println(JOIN_SLACK_MESSAGE)

	sentry.CaptureMessage("deploy executed successfully")
	return nil
}

func validateCluster(ctx context.Context, kubeClient *k8s.Client, namespace string, sentryKubeContext *sentry_utils.KubeContext) error {
	var err error

	fmt.Println("Validating cluster compatibility:")

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

	if !clusterReport.IsCompatible {
		return fmt.Errorf("can't continue with installation, cluster is not compatible for installation")
	}

	return nil
}

func validateNodes(ctx context.Context, kubeClient *k8s.Client, sentryKubeContext *sentry_utils.KubeContext) (*k8s.NodesReport, error) {
	var err error

	fmt.Println("\nValidating cluster nodes compatibility:")

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

	if len(nodesReport.CompatibleNodes) == 0 {
		return nil, fmt.Errorf("can't continue with installation, no compatible nodes for installation")
	}

	return nodesReport, nil
}

func promptInstallSummary(helmClient *helm.Client, releaseName string, clusterName string, namespace string, chart *helm.Chart, nodesReport *k8s.NodesReport, sentryHelmContext *sentry_utils.HelmContext) (bool, error) {
	var err error

	fmt.Println("\nInstalling groundcover:")

	var isUpgrade bool
	var release *helm.Release
	if release, isUpgrade, err = helmClient.IsReleaseInstalled(releaseName); err != nil {
		return false, err
	}

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
			clusterName, namespace, len(nodesReport.CompatibleNodes), len(nodesReport.IncompatibleNodes)+len(nodesReport.CompatibleNodes), chart.Version(),
		)
	}

	return ui.YesNoPrompt(promptMessage, !isUpgrade), nil
}

func installHelmRelease(ctx context.Context, helmClient *helm.Client, releaseName string, chart *helm.Chart, chartValues map[string]interface{}) (*helm.Release, error) {
	var err error

	spinner := ui.NewSpinner("Installing groundcover helm release")
	spinner.Start()
	spinner.StopMessage("groundcover helm release is installed")
	defer spinner.Stop()

	var release *helm.Release
	if release, err = helmClient.Upgrade(ctx, releaseName, chart, chartValues); err != nil {
		spinner.StopFailMessage("groundcover helm release installation failed")
		spinner.StopFail()
		return nil, err
	}

	return release, nil
}

func validateInstall(ctx context.Context, kubeClient *k8s.Client, release *helm.Release, clusterName string, compatibleNodes int, sentryHelmContext *sentry_utils.HelmContext) error {
	var err error

	fmt.Println("\nValidating groundcover installation:")

	var auth0Token auth.Auth0Token
	if err = auth0Token.Load(); err != nil {
		return err
	}

	if err = waitForAlligators(ctx, kubeClient, release, compatibleNodes, sentryHelmContext); err != nil {
		return err
	}

	apiClient := api.NewClient(&auth0Token)
	if err = apiClient.PollIsClusterExist(clusterName); err != nil {
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
	var chart *helm.Chart

	if chart, err = helmClient.GetLatestChart(CHART_NAME, HELM_REPO_URL); err != nil {
		return nil, err
	}

	sentryHelmContext.ChartVersion = chart.Version().String()
	sentryHelmContext.SetOnCurrentScope()
	sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

	return chart, nil
}

func getChartValues(clusterName string, compatibleNodes []*k8s.NodeSummary, sentryHelmContext *sentry_utils.HelmContext) (map[string]interface{}, error) {
	var err error

	var apiKey api.ApiKey
	if err = apiKey.Load(); err != nil {
		return nil, err
	}

	chartValues := make(map[string]interface{})
	chartValues["clusterId"] = clusterName
	chartValues["origin"] = map[string]interface{}{"tag": ""}
	chartValues["global"] = map[string]interface{}{"groundcover_token": apiKey.ApiKey}
	chartValues["commitHashKeyName"] = viper.GetString(COMMIT_HASH_KEY_NAME_FLAG)
	chartValues["repositoryUrlKeyName"] = viper.GetString(REPOSITORY_URL_KEY_NAME_FLAG)

	userValuesOverridePaths := viper.GetStringSlice(VALUES_FLAG)

	var resourcesTunerPresetPaths []string
	if resourcesTunerPresetPaths, err = helm.GetResourcesTunerPresetPaths(compatibleNodes); err != nil {
		return nil, err
	}

	if len(resourcesTunerPresetPaths) > 0 {
		sentryHelmContext.ResourcesPresets = resourcesTunerPresetPaths
		sentryHelmContext.SetOnCurrentScope()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.DEFAULT_RESOURCES_PRESET_TAG, "false")
	} else {
		sentry_utils.SetTagOnCurrentScope(sentry_utils.DEFAULT_RESOURCES_PRESET_TAG, "true")
	}

	var valuesOverride map[string]interface{}
	if valuesOverride, err = helm.SetChartValuesOverrides(&chartValues, append(resourcesTunerPresetPaths, userValuesOverridePaths...)); err != nil {
		return nil, err
	}

	sentryHelmContext.ValuesOverride = valuesOverride
	sentryHelmContext.SetOnCurrentScope()

	return chartValues, nil
}
