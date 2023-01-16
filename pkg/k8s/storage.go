package k8s

import (
	"context"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageProvision struct {
	UseEmptyDir bool
	Reason      string
}

func GenerateStorageProvision(ctx context.Context, client *Client, clusterSummary *ClusterSummary) StorageProvision {
	if !eksClusterRegex.MatchString(clusterSummary.ClusterName) {
		return StorageProvision{
			UseEmptyDir: false,
			Reason:      "Not an EKS cluster",
		}
	}

	if clusterSummary.ServerVersion.LT(semver.MustParse("1.23.0")) {
		return StorageProvision{
			UseEmptyDir: false,
			Reason:      "Kubernetes version is less than 1.23.0",
		}
	}

	list, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=aws-ebs-csi-driver",
	})

	if err != nil {
		return StorageProvision{
			UseEmptyDir: true,
			Reason:      "Error listing aws csi driver pods",
		}
	}

	if len(list.Items) == 0 {
		return StorageProvision{
			UseEmptyDir: true,
			Reason:      "No aws csi driver pods found",
		}
	}

	hasStorageClass := HasDefaultStorageClass(ctx, client, clusterSummary)
	if hasStorageClass {
		return StorageProvision{
			UseEmptyDir: false,
			Reason:      "Has default storage class",
		}
	}

	return StorageProvision{
		UseEmptyDir: true,
		Reason:      "Has aws ebs dirver without default storage class",
	}
}

func HasDefaultStorageClass(ctx context.Context, client *Client, clusterSummary *ClusterSummary) bool {
	storageClasses, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false
	}

	for _, storageClass := range storageClasses.Items {
		if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return true
		}
	}

	return false
}
