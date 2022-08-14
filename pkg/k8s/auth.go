package k8s

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	GKE_GCLOUD_AUTH_PLUGIN_MISSING      = "no Auth Provider found for name \"gcp\""
	GKE_GCLOUD_AUTH_PLUGIN_INSTALL_HINT = `
  * Install gke-gcloud-auth-plugin by following https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
`

	EKS_AUTH_PLUGIN_OUTDATED     = "exec plugin: invalid apiVersion \"client.authentication.k8s.io/v1alpha1\""
	EKS_AUTH_PLUGIN_UPGRADE_HINT = `
  * Upgrade AWS CLI by following https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
  * Update your kubeconfig by following https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
`
)

func (kubeClient *Client) printHintIfAuthError(err error) error {
	switch err.Error() {
	case EKS_AUTH_PLUGIN_OUTDATED:
		logrus.Warn(err)
		fmt.Printf("Hint:%s", EKS_AUTH_PLUGIN_UPGRADE_HINT)
	case GKE_GCLOUD_AUTH_PLUGIN_MISSING:
		logrus.Warn(err)
		fmt.Printf("Hint:%s", GKE_GCLOUD_AUTH_PLUGIN_INSTALL_HINT)
	default:
		return err
	}

	var clusterName string
	if clusterName, err = kubeClient.GetClusterName(); err != nil {
		clusterName = "cluster"
	}

	return fmt.Errorf("authentication failure to %s", clusterName)
}
