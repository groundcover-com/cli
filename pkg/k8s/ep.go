package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kuber *Kuber) DeleteEp(ctx context.Context, resource v1.Endpoints) error {
	client := kuber.client.CoreV1().Endpoints(resource.Namespace)
	return client.Delete(ctx, resource.Name, metav1.DeleteOptions{})
}

func (kuber *Kuber) ListEps(ctx context.Context, namespace string, options metav1.ListOptions) ([]v1.Endpoints, error) {
	var err error
	var list *v1.EndpointsList

	client := kuber.client.CoreV1().Endpoints(namespace)

	if list, err = client.List(ctx, options); err != nil {
		return nil, err
	}

	return list.Items, nil
}
