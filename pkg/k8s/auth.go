package k8s

import "fmt"

const (
	GKE_GCLOUD_AUTH_PLUGIN_MISSING      = "no Auth Provider found for name \"gcp\""
	GKE_GCLOUD_AUTH_PLUGIN_INSTALL_HINT = `
Hint:
  * Install gke-gcloud-auth-plugin by following https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
`

	EKS_AUTH_PLUGIN_OUTDATED     = "exec plugin: invalid apiVersion \"client.authentication.k8s.io/v1alpha1\""
	EKS_AUTH_PLUGIN_UPGRADE_HINT = `
Hint:
  * Upgrade AWS CLI by following https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
  * Update your kubeconfig by following https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
`
)

func wrapHintIfAuthError(err error) error {
	switch err.Error() {
	case EKS_AUTH_PLUGIN_OUTDATED:
		return fmt.Errorf("%v\n%s", err, EKS_AUTH_PLUGIN_UPGRADE_HINT)
	case GKE_GCLOUD_AUTH_PLUGIN_MISSING:
		return fmt.Errorf("%v\n%s", err, GKE_GCLOUD_AUTH_PLUGIN_INSTALL_HINT)
	}

	return err
}
