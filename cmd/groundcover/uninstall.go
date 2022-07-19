package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	"helm.sh/helm/v3/pkg/release"
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
		var release *release.Release
		var kubeClient *k8s.KubeClient
		var helmRelease *helm.HelmReleaser

		if kubeClient, err = k8s.NewKubeClient(viper.GetString(KUBECONFIG_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmRelease, err = helm.NewHelmReleaser(viper.GetString(HELM_RELEASE_FLAG), viper.GetString(NAMESPACE_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if release, err = helmRelease.Get(); err != nil {
			return err
		}

		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		promptMessage := fmt.Sprintf(
			"Current groundcover (cluster: %s, namespace: %s, version: %s)\nAre you sure you want to uninstall?",
			clusterName, helmRelease.Namespace, release.Chart.Metadata.Version,
		)
		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}

		if err = helmRelease.Uninstall(); err != nil {
			return err
		}
		if err = deleteReleaseLeftovers(cmd.Context(), kubeClient, helmRelease); err != nil {
			return err
		}

		if !utils.YesNoPrompt("Do you want to delete groundcover's Persistent Volume Claims? This will remove all of groundcover data", false) {
			return nil
		}

		if err = deletePvcs(cmd.Context(), kubeClient, helmRelease); err != nil {
			return err
		}

		return nil
	},
}

func deleteReleaseLeftovers(ctx context.Context, kubeClient *k8s.KubeClient, helmRelease *helm.HelmReleaser) error {
	var err error
	var svcList *v1.ServiceList
	var epList *v1.EndpointsList

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
	if epList, err = epClient.List(ctx, listOptions); err != nil {
		return err
	}
	for _, ep := range epList.Items {
		if err = epClient.Delete(ctx, ep.Name, deleteOptions); err != nil {
			return err
		}
	}

	return nil
}

func deletePvcs(ctx context.Context, kubeClient *k8s.KubeClient, helmRelease *helm.HelmReleaser) error {
	var err error
	var pvcList *v1.PersistentVolumeClaimList

	deleteOptions := metav1.DeleteOptions{}
	allPvcs := make([]v1.PersistentVolumeClaim, 0)
	labelNames := []string{"release", "app.kubernetes.io/instance"}

	pvcClient := kubeClient.CoreV1().PersistentVolumeClaims(helmRelease.Namespace)
	for _, labelName := range labelNames {
		listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, helmRelease.Name)}
		if pvcList, err = pvcClient.List(ctx, listOptions); err != nil {
			return err
		}
		allPvcs = append(allPvcs, pvcList.Items...)
	}

	for _, pvc := range allPvcs {
		if err = pvcClient.Delete(ctx, pvc.Name, deleteOptions); err != nil {
			return err
		}
	}

	return nil
}
