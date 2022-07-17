package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kuber *Kuber) DeletePvc(ctx context.Context, resource v1.PersistentVolumeClaim) error {
	client := kuber.client.CoreV1().PersistentVolumeClaims(resource.Namespace)
	return client.Delete(ctx, resource.Name, metav1.DeleteOptions{})
}

func (kuber *Kuber) ListPvcs(ctx context.Context, namespace string, options metav1.ListOptions) ([]v1.PersistentVolumeClaim, error) {
	var err error
	var list *v1.PersistentVolumeClaimList

	client := kuber.client.CoreV1().PersistentVolumeClaims(namespace)

	if list, err = client.List(ctx, options); err != nil {
		return nil, err
	}

	return list.Items, nil
}
