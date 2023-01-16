package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

const (
	DELETE_NAMESPACE_FLAG = "delete-namespace"
)

var (
	pvcLabelNames = []string{"release", "app.kubernetes.io/instance"}
)

func init() {
	RootCmd.AddCommand(UninstallCmd)

	UninstallCmd.PersistentFlags().Bool(DELETE_NAMESPACE_FLAG, false, "force delete groundcover namespace")
	viper.BindPFlag(DELETE_NAMESPACE_FLAG, UninstallCmd.PersistentFlags().Lookup(DELETE_NAMESPACE_FLAG))
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
		sentryHelmContext.ReleaseName = releaseName
		sentryHelmContext.SetOnCurrentScope()

		if err = namespaceExists(ctx, kubeClient, namespace); err != nil {
			return err
		}

		var shouldUninstall bool
		var shouldEraseData bool
		var shouldDeleteNamespace bool
		if shouldUninstall, shouldEraseData, shouldDeleteNamespace, err = promptUninstall(ctx, kubeClient, helmClient, clusterName, releaseName, namespace, &sentryHelmContext); err != nil {
			return err
		}

		if shouldUninstall {
			if err = uninstallHelmRelease(ctx, kubeClient, helmClient, releaseName, namespace); err != nil {
				return err
			}
		}

		if shouldEraseData {
			if err = deletePvcs(ctx, kubeClient, releaseName, namespace); err != nil {
				return err
			}
		}

		if shouldDeleteNamespace {
			if err = deleteNamespace(ctx, kubeClient, namespace); err != nil {
				return err
			}
		}

		return nil
	},
}

func promptUninstall(ctx context.Context, kubeClient *k8s.Client, helmClient *helm.Client, clusterName, releaseName, namespace string, sentryHelmContext *sentry_utils.HelmContext) (bool, bool, bool, error) {
	var err error

	ui.GlobalWriter.PrintlnWithPrefixln("Uninstalling groundcover:")

	var shouldUninstall bool
	if shouldUninstall, err = promptUninstallRelease(ctx, kubeClient, helmClient, clusterName, releaseName, namespace, sentryHelmContext); err != nil {
		return false, false, false, err
	}

	var shouldEraseData bool
	if shouldEraseData, err = promptEraseData(ctx, kubeClient, releaseName, namespace); err != nil {
		return false, false, false, err
	}

	var shouldDeleteNamespace bool
	if viper.GetBool(DELETE_NAMESPACE_FLAG) {
		shouldDeleteNamespace = ui.GlobalWriter.YesNoPrompt(fmt.Sprintf("Are you sure you want to delete %s namespace?", namespace), true)
	}

	if !shouldUninstall && !shouldEraseData && !shouldDeleteNamespace {
		ui.GlobalWriter.PrintWarningMessageln(fmt.Sprintf(
			"could not find release %s in namespace %s, maybe groundcover is installed elsewhere? (use --%s, --%s flags)",
			releaseName, namespace, HELM_RELEASE_FLAG, NAMESPACE_FLAG),
		)
		return false, false, false, ErrSilentExecutionAbort
	}

	sentry_utils.SetTagOnCurrentScope(sentry_utils.ERASE_DATA_TAG, strconv.FormatBool(shouldEraseData))

	return shouldUninstall, shouldEraseData, shouldDeleteNamespace, nil
}

func namespaceExists(ctx context.Context, kubeClient *k8s.Client, namespace string) error {
	var err error

	namespaceListOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("kubernetes.io/metadata.name=%s", namespace),
	}

	var namespaceList *v1.NamespaceList
	if namespaceList, err = kubeClient.CoreV1().Namespaces().List(ctx, namespaceListOptions); err != nil {
		return err
	}

	if len(namespaceList.Items) == 0 {
		ui.GlobalWriter.PrintWarningMessageln(fmt.Sprintf("could not find namespace %s, maybe groundcover is installed elsewhere? (use --%s flag)", namespace, NAMESPACE_FLAG))
		return ErrSilentExecutionAbort
	}

	return nil
}

func promptUninstallRelease(ctx context.Context, kubeClient *k8s.Client, helmClient *helm.Client, clusterName, releaseName, namespace string, sentryHelmContext *sentry_utils.HelmContext) (bool, error) {
	var err error

	var release *helm.Release
	if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
		if errors.Is(err, helm_driver.ErrReleaseNotFound) {
			return releaseLeftoversExists(ctx, kubeClient, releaseName, namespace)
		}

		return false, err
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

	if !ui.GlobalWriter.YesNoPrompt(promptMessage, true) {
		return false, ErrExecutionAborted
	}

	return true, nil
}

func promptEraseData(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace string) (bool, error) {
	var err error
	var foundReleasePvcs bool

	pvcClient := kubeClient.CoreV1().PersistentVolumeClaims(namespace)
	for _, labelName := range pvcLabelNames {
		listOptions := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", labelName, releaseName),
		}

		var pvcList *v1.PersistentVolumeClaimList
		if pvcList, err = pvcClient.List(ctx, listOptions); err != nil {
			return false, err
		}
		if len(pvcList.Items) > 0 {
			foundReleasePvcs = true
			break
		}
	}

	if !foundReleasePvcs {
		return false, nil
	}

	// we found PVCs, and we are uninstalling
	return true, nil
}

func uninstallHelmRelease(ctx context.Context, kubeClient *k8s.Client, helmClient *helm.Client, releaseName, namespace string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Uninstalling groundcover helm release")
	spinner.Start()
	spinner.SetStopMessage("groundcover helm release is uninstalled")
	defer spinner.WriteStop()

	if err = helmClient.Uninstall(releaseName); err != nil {
		if !errors.Is(err, helm_driver.ErrReleaseNotFound) {
			spinner.WriteStopFail()
			return err
		}
	}

	if err = deleteReleaseLeftovers(ctx, kubeClient, releaseName, namespace); err != nil {
		spinner.WriteStopFail()
		return err
	}

	return nil
}

func releaseLeftoversExists(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace string) (bool, error) {
	var err error

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", releaseName),
	}

	svcClient := kubeClient.CoreV1().Services(namespace)

	var svcList *v1.ServiceList
	if svcList, err = svcClient.List(ctx, listOptions); err != nil {
		return false, err
	}

	if len(svcList.Items) > 0 {
		return true, nil
	}

	epClient := kubeClient.CoreV1().Endpoints(namespace)

	var epList *v1.EndpointsList
	if epList, err = epClient.List(ctx, listOptions); err != nil {
		return false, err
	}

	if len(epList.Items) > 0 {
		return true, nil
	}

	return false, nil
}

func deleteReleaseLeftovers(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace string) error {
	var err error

	deleteOptions := metav1.DeleteOptions{}
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", releaseName),
	}

	svcClient := kubeClient.CoreV1().Services(namespace)

	var svcList *v1.ServiceList
	if svcList, err = svcClient.List(ctx, listOptions); err != nil {
		return err
	}
	for _, svc := range svcList.Items {
		if err = svcClient.Delete(ctx, svc.Name, deleteOptions); err != nil {
			return err
		}
	}

	epClient := kubeClient.CoreV1().Endpoints(namespace)
	if err = epClient.DeleteCollection(ctx, deleteOptions, listOptions); err != nil {
		return err
	}

	return nil
}

func deletePvcs(ctx context.Context, kubeClient *k8s.Client, releaseName, namespace string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Deleting groundcover pvcs")
	spinner.Start()
	spinner.SetStopMessage("groundcover pvcs are deleted")
	spinner.SetStopFailMessage("failed to delete groundcover pvcs")
	defer spinner.WriteStop()

	deleteOptions := metav1.DeleteOptions{}

	pvcClient := kubeClient.CoreV1().PersistentVolumeClaims(namespace)
	for _, labelName := range pvcLabelNames {
		listOptions := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", labelName, releaseName),
		}

		if err = pvcClient.DeleteCollection(ctx, deleteOptions, listOptions); err != nil {
			spinner.WriteStopFail()
			return err
		}
	}

	return nil
}

func deleteNamespace(ctx context.Context, kubeClient *k8s.Client, namespace string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Deleting groundcover namespace")
	spinner.Start()
	spinner.SetStopMessage(fmt.Sprintf("%s namespace is deleted", namespace))
	spinner.SetStopFailMessage(fmt.Sprintf("failed to delete %s namespace", namespace))
	defer spinner.WriteStop()

	if err = kubeClient.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{}); err != nil {
		spinner.WriteStopFail()
		return err
	}

	return nil
}
