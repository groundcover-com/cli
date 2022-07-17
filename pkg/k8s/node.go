package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kuber *Kuber) ListNodes(ctx context.Context) ([]v1.Node, error) {
	var err error
	var noteList *v1.NodeList

	if noteList, err = kuber.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	return noteList.Items, nil
}

func (kuber *Kuber) NodesCount(ctx context.Context) (int, error) {
	var err error
	var nodes []v1.Node

	if nodes, err = kuber.ListNodes(ctx); err != nil {
		return 0, err
	}

	return len(nodes), nil
}
