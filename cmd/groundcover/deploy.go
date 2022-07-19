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
		var numberOfNodes int
		var clusterName string
		var promptMessage string
		var nodeList *v1.NodeList
		var apiKey *auth.ApiKey
		var token *auth.Auth0Token
		var kubeClient *k8s.KubeClient
		var helmChart *helm.HelmCharter
		var helmRelease *helm.HelmReleaser
		var customClaims *auth.CustomClaims

		if apiKey, err = auth.LoadApiKey(); err != nil {
			return fmt.Errorf("failed to load api key. error: %s", err.Error())
		}
		if token, err = auth.MustLoadDefaultCredentials(); err != nil {
			return err
		}
		if customClaims = viper.Get(USER_CUSTOM_CLAIMS_KEY).(*auth.CustomClaims); customClaims == nil {
			return fmt.Errorf("deployment failed to get user custom claims")
		}

		if kubeClient, err = k8s.NewKubeClient(viper.GetString(KUBECONFIG_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmRelease, err = helm.NewHelmReleaser(viper.GetString(HELM_RELEASE_FLAG), viper.GetString(NAMESPACE_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmChart, err = helm.NewHelmCharter(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}

		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
			return err
		}
		numberOfNodes = len(nodeList.Items)

		currentVersion, isLatestNewer := checkCurrentDeployedVersion(helmRelease, helmChart)
		isUpgrade := currentVersion != ""
		switch {
		case !isUpgrade:
			promptMessage = fmt.Sprintf(
				"Deploying groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s).\nDo you want to deploy?",
				clusterName, helmRelease.Namespace, numberOfNodes, helmChart.Version,
			)
		case isLatestNewer:
			promptMessage = fmt.Sprintf(
				"Current groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s) is out of date!, The latest version is %s.\nDo you want to upgrade?",
				clusterName, helmRelease.Namespace, numberOfNodes, currentVersion, helmChart.Version,
			)
		case !isLatestNewer:
			promptMessage = fmt.Sprintf(
				"Current groundcover (cluster: %s, namespace: %s, nodes: %d, version: %s) is latest version.\nDo you want to redeploy?",
				clusterName, helmRelease.Namespace, numberOfNodes, helmChart.Version,
			)
		}

		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}

		chartValues := make(map[string]interface{})
		chartValues["clusterId"] = clusterName
		chartValues["global"] = map[string]interface{}{"groundcover_token": apiKey.ApiKey}
		sentry.CaptureDeploymentEvent(customClaims, isUpgrade, helmChart.Version.String(), numberOfNodes)
		if err = helmRelease.Upgrade(cmd.Context(), helmChart.Get(), chartValues); err != nil {
			return err
		}

		if err = waitForAlligators(cmd.Context(), kubeClient, helmRelease); err != nil {
			return err
		}

		if err = api.WaitUntilClusterConnectedToSaas(token, clusterName); err != nil {
			return err
		}
		fmt.Printf("Cluster %q is connected to SaaS!\n", clusterName)

		utils.OpenBrowser(fmt.Sprintf("%s/clusterId=%s", GROUNDCOVER_URL, clusterName))
		return nil
	},
}

func getClusterName(kubeClient *k8s.KubeClient) (string, error) {
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
