package k8s

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/blang/semver/v4"
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
  * Update your kubeconfig by following https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html`
	HINT_INSTALL_AWS_CLI = `Hint: 
  * Install aws cli by following https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html`
)

var (
	DefaultAwsCliVersionValidator = &AwsCliVersionValidator{
		Regexp:                    regexp.MustCompile(`^aws-cli/(\d+\.\d+\.\d+)`),
		MinimumSupportedV1Version: semver.Version{Major: 1, Minor: 23, Patch: 9},
		MinimumSupportedV2Version: semver.Version{Major: 2, Minor: 7, Patch: 0},
	}
)

type AwsCliVersionValidator struct {
	Regexp                    *regexp.Regexp
	MinimumSupportedV1Version semver.Version
	MinimumSupportedV2Version semver.Version
}

func (validator *AwsCliVersionValidator) wrapError(err error) error {
	return errors.Wrapf(err, "failed getting aws cli version (required v%s+/v%s+), got", validator.MinimumSupportedV1Version, validator.MinimumSupportedV2Version)
}

func (validator *AwsCliVersionValidator) Fetch(ctx context.Context) (semver.Version, error) {
	var err error
	var version semver.Version

	var versionByte []byte
	if versionByte, err = exec.CommandContext(ctx, "aws", "--version").Output(); err != nil {
		return version, validator.wrapError(err)
	}

	return validator.Parse(string(versionByte))
}

func (validator *AwsCliVersionValidator) Parse(versionString string) (semver.Version, error) {
	var version semver.Version

	matches := validator.Regexp.FindStringSubmatch(versionString)
	if len(matches) != 2 {
		return version, validator.wrapError(fmt.Errorf("unknown aws cli version: %q", versionString))
	}

	return semver.Parse(matches[1])
}

func (validator *AwsCliVersionValidator) Validate(version semver.Version) error {
	switch version.Major {
	case 1:
		if version.LT(validator.MinimumSupportedV1Version) {
			return fmt.Errorf("aws-cli version is unsupported (%s < %s)", version, validator.MinimumSupportedV1Version)
		}
	case 2:
		if version.LT(validator.MinimumSupportedV2Version) {
			return fmt.Errorf("aws-cli version is unsupported (%s < %s)", version, validator.MinimumSupportedV2Version)
		}
	default:
		return fmt.Errorf("aws-cli version %s is unsupported", version)
	}

	return nil
}

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
		ui.SingletonWriter.PrintWarningMessage(fmt.Sprintf("%s\n%s", err, HINT_EKS_AUTH_PLUGIN_UPGRADE))
	case GKE_GCLOUD_AUTH_PLUGIN_MISSING:
		ui.SingletonWriter.PrintWarningMessage(fmt.Sprintf("%s\n%s", err, HINT_GKE_GCLOUD_AUTH_PLUGIN_INSTALL))
	default:
		return err
	}

	var clusterName string
	if clusterName, err = kubeClient.GetClusterName(); err != nil {
		clusterName = "cluster"
	}

	return fmt.Errorf("authentication failure to %s", clusterName)
}
