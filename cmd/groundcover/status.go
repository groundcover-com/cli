package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SPINNER_TYPE                = 8 // .oO@*
	ALLIGATORS_POLLING_TIMEOUT  = time.Minute * 2
	ALLIGATORS_POLLING_INTERVAL = time.Second * 10
)

func init() {
	RootCmd.AddCommand(StatusCmd)
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get groundcover current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		namespace := viper.GetString(NAMESPACE_FLAG)
		kubeconfig := viper.GetString(KUBECONFIG_FLAG)
		kubecontext := viper.GetString(KUBECONTEXT_FLAG)
		releaseName := viper.GetString(HELM_RELEASE_FLAG)

		var kubeClient *k8s.Client
		if kubeClient, err = k8s.NewKubeClient(kubeconfig, kubecontext); err != nil {
			return err
		}

		var helmClient *helm.Client
		if helmClient, err = helm.NewHelmClient(namespace, kubecontext); err != nil {
			return err
		}

		var release *helm.Release
		if release, err = helmClient.GetCurrentRelease(releaseName); err != nil {
			return err
		}

		var chart *helm.Chart
		if chart, err = helmClient.GetLatestChart(CHART_NAME, HELM_REPO_URL); err != nil {
			return err
		}

		if chart.Version().GT(release.Version()) {
			fmt.Printf("Current groundcover %s is out of date!, The latest version is %s.", release.Version(), chart.Version())
		}

		var nodeList *v1.NodeList
		if nodeList, err = kubeClient.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{}); err != nil {
			return err
		}
		numberOfNodes := len(nodeList.Items)

		if err = waitForAlligators(cmd.Context(), kubeClient, release, numberOfNodes); err != nil {
			return err
		}

		return nil
	},
}

func waitForAlligators(ctx context.Context, kubeClient *k8s.Client, helmRelease *helm.Release, numberOfNodes int) error {
	var err error
	var podList *v1.PodList

	version := helmRelease.Version().String()
	podClient := kubeClient.CoreV1().Pods(helmRelease.Namespace)
	listOptions := metav1.ListOptions{LabelSelector: "app=alligator", FieldSelector: "status.phase=Running"}
	spinner := utils.NewSpinner(SPINNER_TYPE, "Waiting until all nodes are monitored ")
	spinner.Suffix = fmt.Sprintf(" (%d/%d)", 0, numberOfNodes)

	areAlligatorsRunning := func() (bool, error) {
		runningAlligators := 0
		if podList, err = podClient.List(ctx, listOptions); err != nil {
			return false, err
		}
		for _, pod := range podList.Items {
			if pod.Annotations["groundcover_version"] == version {
				runningAlligators++
			}
		}
		spinner.Suffix = fmt.Sprintf(" (%d/%d)", runningAlligators, numberOfNodes)
		if numberOfNodes > runningAlligators {
			return false, nil
		}
		spinner.FinalMSG = fmt.Sprintf("All nodes are monitored (%d/%d) !\n", runningAlligators, numberOfNodes)
		return true, nil
	}

	if err = spinner.Poll(areAlligatorsRunning, ALLIGATORS_POLLING_INTERVAL, ALLIGATORS_POLLING_TIMEOUT); err != nil {
		return err
	}

	return nil
}
