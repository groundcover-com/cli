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
		var clusterName string
		var promptMessage string
		var chart *helm.Chart
		var release *helm.Release
		var nodeList *v1.NodeList
		var apiKey *auth.ApiKey
		var token *auth.Auth0Token
		var kubeClient *k8s.Client
		var helmClient *helm.Client
		var customClaims *auth.CustomClaims

		namespace := viper.GetString(NAMESPACE_FLAG)
		kubeconfig := viper.GetString(KUBECONFIG_FLAG)
		kubecontext := viper.GetString(KUBECONTEXT_FLAG)
		releaseName := viper.GetString(HELM_RELEASE_FLAG)

		if apiKey, err = auth.LoadApiKey(); err != nil {
			return fmt.Errorf("failed to load api key. error: %s", err.Error())
		}
		if token, err = auth.MustLoadDefaultCredentials(); err != nil {
			return err
		}
		if customClaims = viper.Get(USER_CUSTOM_CLAIMS_KEY).(*auth.CustomClaims); customClaims == nil {
			return fmt.Errorf("deployment failed to get user custom claims")
		}

		if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
			return err
		}
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}

		isUpgrade := false
		isLatestNewer := false
		if chart, err = helmClient.Show(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}
		if release, _ = helmClient.Status(releaseName); release != nil {
			isUpgrade = true
			isLatestNewer = chart.Version().GT(release.Version())
		}

		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
			return err
		}
		numberOfNodes := len(nodeList.Items)

		switch {
		case !isUpgrade:
			promptMessage = fmt.Sprintf(
				"Deploying groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s).\nDo you want to deploy?",
				clusterName, namespace, numberOfNodes, chart.Version(),
			)
		case !isLatestNewer:
			promptMessage = fmt.Sprintf(
				"Current groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s) is latest version.\nDo you want to redeploy?",
				clusterName, namespace, numberOfNodes, chart.Version(),
			)
		case isLatestNewer:
			promptMessage = fmt.Sprintf(
				"Current groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s) is out of date!, The latest version is %s.\nDo you want to upgrade?",
				clusterName, namespace, numberOfNodes, release.Version(), chart.Version(),
			)
		}

		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}
		sentry.CaptureDeploymentEvent(customClaims, isUpgrade, chart.Version().String(), numberOfNodes)

		chartValues := make(map[string]interface{})
		chartValues["clusterId"] = clusterName
		chartValues["global"] = map[string]interface{}{"groundcover_token": apiKey.ApiKey}
		if err = helmClient.Upgrade(cmd.Context(), releaseName, chart, chartValues); err != nil {
			return err
		}
		if release, err = helmClient.Status(releaseName); err != nil {
			return err
		}

		if err = waitForAlligators(cmd.Context(), kubeClient, release); err != nil {
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
