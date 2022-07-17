package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kuber *Kuber) ListPods(ctx context.Context, namespace string, options metav1.ListOptions) ([]v1.Pod, error) {
	var err error
	var list *v1.PodList

	client := kuber.client.CoreV1().Pods(namespace)

	if list, err = client.List(ctx, options); err != nil {
		return nil, err
	}

	return list.Items, nil
}
