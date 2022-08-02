package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	sentry "groundcover.com/pkg/custom_sentry"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CHART_NAME                    = "groundcover"
	DEFAULT_GROUNDCOVER_RELEASE   = "groundcover"
	DEFAULT_GROUNDCOVER_NAMESPACE = "groundcover"
	GROUNDCOVER_URL               = "https://app.groundcover.com"
	HELM_REPO_URL                 = "https://helm.groundcover.com"
)

func init() {
	RootCmd.AddCommand(DeployCmd)
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		namespace := viper.GetString(NAMESPACE_FLAG)
		kubeconfig := viper.GetString(KUBECONFIG_FLAG)
		kubecontext := viper.GetString(KUBECONTEXT_FLAG)
		releaseName := viper.GetString(HELM_RELEASE_FLAG)

		var apiKey *auth.ApiKey
		if apiKey, err = auth.LoadApiKey(); err != nil {
			return fmt.Errorf("failed to load api key. error: %s", err.Error())
		}

		var token *auth.Auth0Token
		if token, err = auth.MustLoadDefaultCredentials(); err != nil {
			return err
		}

		var customClaims *auth.CustomClaims
		if customClaims = viper.Get(USER_CUSTOM_CLAIMS_KEY).(*auth.CustomClaims); customClaims == nil {
			return fmt.Errorf("deployment failed to get user custom claims")
		}

		var kubeClient *k8s.Client
		if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
			return err
		}

		var helmClient *helm.Client
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}

		var clusterName string
		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		var chart *helm.Chart
		if chart, err = helmClient.GetLatestChart(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}

		var nodesSummeries []k8s.NodeSummary
		if nodesSummeries, err = kubeClient.GetNodesSummeries(cmd.Context()); err != nil {
			return err
		}
		nodesCount := len(nodesSummeries)

		var isUpgrade bool
		if isUpgrade, err = helmClient.IsReleaseInstalled(releaseName); err != nil {
			return err
		}

		var promptMessage string
		var expectedAlligatorsCount int
		switch {
		case !isUpgrade:
			nodeRequirements := k8s.NewNodeMinimumRequirements()
			adequateNodesReports, _ := nodeRequirements.GenerateNodeReports(nodesSummeries)
			expectedAlligatorsCount = len(adequateNodesReports)

			promptMessage = fmt.Sprintf(
				"Deploying groundcover (cluster: %s, namespace: %s, compatible nodes: %d/%d, version: %s).\nDo you want to deploy?",
				clusterName, namespace, expectedAlligatorsCount, nodesCount, chart.Version(),
			)
		case isUpgrade:
			var release *helm.Release
			if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
				return err
			}

			var podList *v1.PodList
			listOptions := metav1.ListOptions{LabelSelector: "app=alligator", FieldSelector: "status.phase=Running"}
			if podList, err = kubeClient.CoreV1().Pods(release.Namespace).List(cmd.Context(), listOptions); err != nil {
				return err
			}
			expectedAlligatorsCount = len(podList.Items)

			if chart.Version().GT(release.Version()) {
				promptMessage = fmt.Sprintf(
					"Current groundcover (cluster: %s, namespace: %s, compatible nodes: %d/%d, version: %s) is out of date!, The latest version is %s.\nDo you want to upgrade?",
					clusterName, namespace, expectedAlligatorsCount, nodesCount, release.Version(), chart.Version(),
				)
			} else {
				promptMessage = fmt.Sprintf(
					"Current groundcover (cluster: %s, namespace: %s, compatible nodes: %d/%d, version: %s) is latest version.\nDo you want to redeploy?",
					clusterName, namespace, expectedAlligatorsCount, nodesCount, chart.Version(),
				)
			}
		}

		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}
		sentry.CaptureDeploymentEvent(customClaims, isUpgrade, chart.Version().String(), nodesCount)

		chartValues := defaultChartValues(clusterName, apiKey.ApiKey)
		if err = helmClient.Upgrade(cmd.Context(), releaseName, chart, chartValues); err != nil {
			return err
		}

		var release *helm.Release
		if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
			return err
		}

		if err = waitForAlligators(cmd.Context(), kubeClient, release, expectedAlligatorsCount); err != nil {
			return err
		}

		if err = api.WaitUntilClusterConnectedToSaas(token, clusterName); err != nil {
			return err
		}
		fmt.Printf("Cluster %q is connected to SaaS!\n", clusterName)

		utils.TryOpenBrowser(fmt.Sprintf("%s/clusterId=%s", GROUNDCOVER_URL, clusterName))
		return nil
	},
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
	chartValues["global"] = map[string]interface{}{"groundcover_token": apikey}

	return chartValues
}
