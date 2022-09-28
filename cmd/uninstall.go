package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	helm_driver "helm.sh/helm/v3/pkg/storage/driver"
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

		var clusterName string
		if clusterName, err = getClusterName(kubeClient); err != nil {
			return err
		}

		var helmClient *helm.Client
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}

		var sentryHelmContext sentry_utils.HelmContext
		sentryHelmContext.ReleaseName = viper.GetString(HELM_RELEASE_FLAG)
		sentryHelmContext.SetOnCurrentScope()

		var release *helm.Release
		if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
			if errors.Is(err, helm_driver.ErrReleaseNotFound) {
				ui.PrintWarningMessage(fmt.Sprintf("could not find release %s in namespace %s, maybe groundcover is installed elsewhere?\n", releaseName, namespace))
				return nil
			}

			return err
		}

		sentryHelmContext.RepoUrl = HELM_REPO_URL
		sentryHelmContext.ChartName = release.Chart.Name()
		sentryHelmContext.ChartVersion = release.Chart.Metadata.Version
		sentryHelmContext.SetOnCurrentScope()
		sentry_utils.SetTagOnCurrentScope(sentry_utils.CHART_VERSION_TAG, sentryHelmContext.ChartVersion)

		promptMessage := fmt.Sprintf(
			"Current groundcover installation in your cluster: (cluster: %s, namespace: %s, version: %s).\nAre you sure you want to uninstall?",
			clusterName, namespace, release.Version(),
		)
		if !ui.YesNoPrompt(promptMessage, false) {
			return ErrExecutionAborted
		}

		if err = helmClient.Uninstall(release.Name); err != nil {
			return err
		}
		if err = deleteReleaseLeftovers(ctx, kubeClient, release); err != nil {
			return err
		}
		fmt.Println("uninstall executed successfully")

		if !ui.YesNoPrompt("Do you want to delete groundcover's Persistent Volume Claims? This will remove all of groundcover data", false) {
			sentry.CaptureMessage("delete pvcs execution aborted")
			return nil
		}

		if err = deletePvcs(ctx, kubeClient, release); err != nil {
			return err
		}
		fmt.Println("delete pvcs executed successfully")
		sentry.CaptureMessage("delete pvcs executed successfully")

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
