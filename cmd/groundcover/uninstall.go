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
		var kuber *k8s.Kuber
		var release *release.Release
		var helmRelease *helm.HelmReleaser

		if kuber, err = k8s.NewKuber(viper.GetString(KUBECONFIG_FLAG), viper.GetString(KUBECONTEXT_FLAG)); err != nil {
			return err
		}
		if helmRelease, err = helm.NewHelmReleaser(viper.GetString(HELM_RELEASE_FLAG)); err != nil {
			return err
		}
		if release, err = helmRelease.Get(); err != nil {
			return err
		}

		if clusterName, err = getClusterName(kuber); err != nil {
			return err
		}

		promptMessage := fmt.Sprintf(
			"Current groundcover (cluster: %s, namespace: %s, version: %s), are you sure you want to uninstall?",
			clusterName, helmRelease.Namespace, release.Chart.Metadata.Version,
		)
		if !utils.YesNoPrompt(promptMessage, false) {
			return nil
		}

		if err = helmRelease.Uninstall(); err != nil {
			return err
		}
		if err = deleteReleaseLeftovers(cmd.Context(), kuber, helmRelease); err != nil {
			return err
		}

		if !utils.YesNoPrompt("Do you want to delete groundcover's Persistent Volume Claims? This will remove all of groundcover data", false) {
			return nil
		}

		if err = deletePvcs(cmd.Context(), kuber, helmRelease); err != nil {
			return err
		}

		return nil
	},
}

func deleteReleaseLeftovers(ctx context.Context, kuber *k8s.Kuber, helmRelease *helm.HelmReleaser) error {
	var err error
	var svcList []v1.Service
	var epList []v1.Endpoints

	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("release=%s", helmRelease.Name)}

	if svcList, err = kuber.ListSvcs(ctx, helmRelease.Namespace, listOptions); err != nil {
		return err
	}
	for _, svc := range svcList {
		if err = kuber.DeleteSvc(ctx, svc); err != nil {
			return err
		}
	}

	if epList, err = kuber.ListEps(ctx, helmRelease.Namespace, listOptions); err != nil {
		return err
	}
	for _, ep := range epList {
		if err = kuber.DeleteEp(ctx, ep); err != nil {
			return err
		}
	}

	return nil
}

func deletePvcs(ctx context.Context, kuber *k8s.Kuber, helmRelease *helm.HelmReleaser) error {
	var err error
	var _pvcList []v1.PersistentVolumeClaim

	pvcList := make([]v1.PersistentVolumeClaim, 0)
	labelNames := []string{"release", "app.kubernetes.io/instance"}

	for _, labelName := range labelNames {
		listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelName, helmRelease.Name)}
		if _pvcList, err = kuber.ListPvcs(ctx, helmRelease.Namespace, listOptions); err != nil {
			return err
		}
		pvcList = append(pvcList, _pvcList...)
	}

	for _, pvc := range pvcList {
		if err = kuber.DeletePvc(ctx, pvc); err != nil {
			return err
		}
	}

	return nil
}
