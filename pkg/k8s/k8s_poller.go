package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ALLIGATOR_LABEL_SELECTOR = "app=alligator"
	ALLIGATOR_FIELD_SELECTOR = "status.phase=Running"
)

func GetRunningAlligators(ctx context.Context, kubeClient *Client, helmVersion string, namespace string) (int, error) {
	podClient := kubeClient.CoreV1().Pods(namespace)
	listOptions := metav1.ListOptions{
		LabelSelector: ALLIGATOR_LABEL_SELECTOR,
		FieldSelector: ALLIGATOR_FIELD_SELECTOR,
	}

	runningAlligators := 0

	podList, err := podClient.List(ctx, listOptions)
	if err != nil {
		return runningAlligators, err
	}

	for _, pod := range podList.Items {
		if pod.Annotations["groundcover_version"] == helmVersion {
			runningAlligators++
		}
	}

	return runningAlligators, nil
}
