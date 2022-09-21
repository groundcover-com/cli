package k8s

import (
	"context"

	"github.com/pkg/errors"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
