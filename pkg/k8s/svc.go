package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kuber *Kuber) DeleteSvc(ctx context.Context, resource v1.Service) error {
	client := kuber.client.CoreV1().Services(resource.Namespace)
	return client.Delete(ctx, resource.Name, metav1.DeleteOptions{})
}

func (kuber *Kuber) ListSvcs(ctx context.Context, namespace string, options metav1.ListOptions) ([]v1.Service, error) {
	var err error
	var list *v1.ServiceList

	client := kuber.client.CoreV1().Services(namespace)

	if list, err = client.List(ctx, options); err != nil {
		return nil, err
	}

	return list.Items, nil
}
