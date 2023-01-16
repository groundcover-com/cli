package k8s

import (
	"context"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ShouldUseEmptydir(ctx context.Context, client *Client, clusterSummary *ClusterSummary) bool {
	if !eksClusterRegex.MatchString(clusterSummary.ClusterName) {
		return false
	}

	if clusterSummary.ServerVersion.LT(semver.MustParse("1.23.0")) {
		return false
	}

	list, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=aws-ebs-csi-driver",
	})

	if err != nil {
		return true
	}

	if len(list.Items) == 0 {
		return true
	}

	storageClasses, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return true
	}

	var hasDefaultStorageClass bool
	for _, storageClass := range storageClasses.Items {
		if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			hasDefaultStorageClass = true
		}
	}

	return hasDefaultStorageClass
}
