package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	cs "groundcover.com/pkg/custom_sentry"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
)

const (
	CLUSTER_NAME_FLAG             = "cluster-name"
	KUBECONFIG_PATH_FLAG          = "kubeconfig-path"
	GROUNDCOVER_URL               = "https://app.groundcover.com"
	GROUNDCOVER_HELM_REPO_ADDR    = "https://helm.groundcover.com"
	GROUNDCOVER_HELM_REPO_NAME    = "groundcover"
	GROUNDCOVER_CHART_NAME        = "groundcover/groundcover"
	HELM_BINARY_NAME              = "helm"
	MANUAL_DEPLOYMENT_TAG         = "manual"
	AUTO_DEPLOYMENT_TAG           = "auto"
	GROUNDCOVER_NAMESPACE_FLAG    = "groundcover-namespace"
	DEFAULT_GROUNDCOVER_NAMESPACE = "groundcover"
)

var (
	waitForClusterConnectedToSaasTimeout = time.Minute * 5
)

func init() {
	RootCmd.AddCommand(DeployCmd)

	DeployCmd.PersistentFlags().Bool(MANUAL_FLAG, true, "install groundcover helm chart manually")
	viper.BindPFlag(MANUAL_FLAG, DeployCmd.PersistentFlags().Lookup(MANUAL_FLAG))

	DeployCmd.PersistentFlags().String(CLUSTER_NAME_FLAG, "", "cluster name")
	viper.BindPFlag(CLUSTER_NAME_FLAG, DeployCmd.PersistentFlags().Lookup(CLUSTER_NAME_FLAG))
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Deploying groundcover")

		customClaims, ok := cmd.Context().Value(USER_CUSTOM_CLAIMS_KEY).(*auth.CustomClaims)
		if !ok {
			return fmt.Errorf("deployment failed to get user custom claims")
		}

		apiKey, err := auth.LoadApiKey()
		if err != nil {
			return fmt.Errorf("failed to load api key. error: %s", err.Error())
		}

		clusterName, err := getClusterName(cmd)
		if err != nil {
			return err
		}
		formattedClusterName := k8s.FormatClusterName(clusterName)

		metadataFetcher, err := k8s.NewMetadataFetcher(viper.GetString(KUBECONFIG_PATH_FLAG))
		if err != nil {
			return err
		}

		numberOfNodes, err := metadataFetcher.GetNumberOfNodes(cmd.Context())
		if err != nil {
			return err
		}

		helmCmd, err := helm.NewHelmCmd()
		if err != nil {
			return err
		}

		version, err := helmCmd.GetLatestChartVersion(cmd.Context())
		if err != nil {
			return err
		}

		fmt.Printf("Installing groundcover on cluster: %q with cluster name: %q\n", clusterName, formattedClusterName)
		fmt.Printf("Available Nodes: %d\n", numberOfNodes)
		fmt.Printf("Installing groundcover version: %s\n", version)

		groundcoverNamespace := viper.GetString(GROUNDCOVER_NAMESPACE_FLAG)
		automatedInstallation := utils.YesNoPrompt("Do you want to run automated installation", true)
		if !automatedInstallation {
			cs.CaptureDeploymentEvent(customClaims, MANUAL_DEPLOYMENT_TAG, version, numberOfNodes)
			return manualInstallation(helmCmd, apiKey.ApiKey, formattedClusterName, groundcoverNamespace)
		}

		cs.CaptureDeploymentEvent(customClaims, AUTO_DEPLOYMENT_TAG, version, numberOfNodes)
		return autoInstallation(cmd.Context(), helmCmd, metadataFetcher, apiKey.ApiKey, formattedClusterName, groundcoverNamespace)
	},
}

func getClusterName(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Lookup(CLUSTER_NAME_FLAG).Changed {
		return viper.GetString(CLUSTER_NAME_FLAG), nil
	}

	clusterName, err := k8s.GetClusterName(viper.GetString(KUBECONFIG_PATH_FLAG))
	if err != nil {
		return "", fmt.Errorf("failed to get cluster name. error: %s", err.Error())
	}

	return clusterName, nil
}

func autoInstallation(ctx context.Context, helmCmd *helm.HelmCmd, metadataFetcher *k8s.MetadataFetcher, apiKey, clusterName, groundcoverNamespace string) error {
	fmt.Println("Installing groundcover...")

	err := helmCmd.RepoAdd(ctx)
	if err != nil {
		return err
	}

	err = helmCmd.RepoUpdate(ctx)
	if err != nil {
		return err
	}

	err = helmCmd.Upgrade(ctx, apiKey, clusterName, groundcoverNamespace)
	if err != nil {
		return err
	}

	token, err := auth.MustLoadDefaultCredentials()
	if err != nil {
		return err
	}

	err = api.WaitUntilClusterConnectedToSaas(ctx, token, clusterName)
	if err != nil {
		return fmt.Errorf("failed while waiting for groundcover installation to connect: %s", err.Error())
	}

	fmt.Printf("Cluster %q is connected to SaaS!\n", clusterName)

	url := fmt.Sprintf("%s/?clusterId=%s", GROUNDCOVER_URL, clusterName)
	cmd := exec.Command("xdg-open", url)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to open groundcover url in browser. You can browse to: %s", url)
	}

	return nil
}

func manualInstallation(helmCmd *helm.HelmCmd, apiKey, clusterName, groundcoverNamespace string) error {
	fmt.Print("To deploy groundcover please run the following command:\n")
	fmt.Print("\n\n")
	fmt.Printf(helmCmd.BuildInstallCommand(apiKey, clusterName, groundcoverNamespace))
	return nil
}
