package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"groundcover.com/pkg/ui"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

const (
	GKE_GCLOUD_AUTH_PLUGIN_MISSING      = "no Auth Provider found for name \"gcp\""
	HINT_GKE_GCLOUD_AUTH_PLUGIN_INSTALL = `Hint:
  * Install gke-gcloud-auth-plugin by following https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
`

	EKS_AUTH_PLUGIN_OUTDATED     = "exec plugin: invalid apiVersion \"client.authentication.k8s.io/v1alpha1\""
	HINT_EKS_AUTH_PLUGIN_UPGRADE = `Hint:
  * Upgrade AWS CLI by following https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
  * Update your kubeconfig by following https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
`
)

func OverrideDepartedAuthenticationApiVersion(restConfig *restclient.Config) {
	if restConfig.ExecProvider == nil {
		return
	}

	if restConfig.ExecProvider.APIVersion == "client.authentication.k8s.io/v1alpha1" {
		restConfig.ExecProvider.APIVersion = "client.authentication.k8s.io/v1beta1"
	}

}

func (kubeClient *Client) isActionPermitted(ctx context.Context, action *authv1.ResourceAttributes) (bool, error) {
	var err error

	accessReview := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{ResourceAttributes: action},
	}

	if accessReview, err = kubeClient.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, accessReview, metav1.CreateOptions{}); err != nil {
		return false, errors.Wrapf(err, "api error on resource: %s", action.Resource)
	}

	if accessReview.Status.Denied || !accessReview.Status.Allowed {
		return false, nil
	}

	return true, nil
}

func (kubeClient *Client) printHintIfAuthError(err error) error {
	switch err.Error() {
	case EKS_AUTH_PLUGIN_OUTDATED:
		ui.PrintWarningMessage(fmt.Sprintf("%s\n%s", err, HINT_EKS_AUTH_PLUGIN_UPGRADE))
	case GKE_GCLOUD_AUTH_PLUGIN_MISSING:
		ui.PrintWarningMessage(fmt.Sprintf("%s\n%s", err, HINT_GKE_GCLOUD_AUTH_PLUGIN_INSTALL))
	default:
		return err
	}

	var clusterName string
	if clusterName, err = kubeClient.GetClusterName(); err != nil {
		clusterName = "cluster"
	}

	return fmt.Errorf("authentication failure to %s", clusterName)
}
