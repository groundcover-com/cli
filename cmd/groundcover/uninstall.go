package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	RootCmd.AddCommand(UninstallCmd)
}

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var clusterName string
		var release *helm.Release
		var kubeClient *k8s.Client
		var helmClient *helm.Client

		namespace := viper.GetString(NAMESPACE_FLAG)
		kubeconfig := viper.GetString(KUBECONFIG_FLAG)
		kubecontext := viper.GetString(KUBECONTEXT_FLAG)
		releaseName := viper.GetString(HELM_RELEASE_FLAG)

		if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
			return err
		}
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}
		if release, err = helmClient.Status(releaseName); err != nil {
			return err
		}

		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		promptMessage := fmt.Sprintf(
			"Current groundcover (cluster: %s, namespace: %s, version: %s)\nAre you sure you want to uninstall?",
			clusterName, namespace, release.Chart.Metadata.Version,
		)
		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}

		if err = helmClient.Uninstall(releaseName); err != nil {
			return err
		}
		if err = deleteReleaseLeftovers(cmd.Context(), kubeClient, release); err != nil {
			return err
		}

		if !utils.YesNoPrompt("Do you want to delete groundcover's Persistent Volume Claims? This will remove all of groundcover data", false) {
			return nil
		}

		if err = deletePvcs(cmd.Context(), kubeClient, release); err != nil {
			return err
		}

		return nil
	},
}

func deleteReleaseLeftovers(ctx context.Context, kubeClient *k8s.Client, helmRelease *helm.Release) error {
	var err error
	var svcList *v1.ServiceList

	deleteOptions := metav1.DeleteOptions{}
	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("release=%s", helmRelease.Name)}

	svcClient := kubeClient.CoreV1().Services(helmRelease.Namespace)
	if svcList, err = svcClient.List(ctx, listOptions); err != nil {
		return err
	}
	for _, svc := range svcList.Items {
		if err = svcClient.Delete(ctx, svc.Name, deleteOptions); err != nil {
			return err
		}
	}

	epClient := kubeClient.CoreV1().Endpoints(helmRelease.Namespace)
	if err = epClient.DeleteCollection(ctx, deleteOptions, listOptions); err != nil {
		return err
	}

	return nil
}

func deletePvcs(ctx context.Context, kubeClient *k8s.Client, helmRelease *helm.Release) error {
	var err error

	deleteOptions := metav1.DeleteOptions{}
	labelNames := []string{"release", "app.kubernetes.io/instance"}

	pvcClient := kubeClient.CoreV1().PersistentVolumeClaims(helmRelease.Namespace)
	for _, labelName := range labelNames {
		listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, helmRelease.Name)}
		if err = pvcClient.DeleteCollection(ctx, deleteOptions, listOptions); err != nil {
			return err
		}
	}

	return nil
}
