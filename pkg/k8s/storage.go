package k8s

import (
	"context"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageProvision struct {
	PersistentStorage bool
	Reason            string
}

func GenerateStorageProvision(ctx context.Context, client *Client, clusterSummary *ClusterSummary) StorageProvision {
	if IsEksCluster(clusterSummary.ClusterName) {
		return generateEksStorageProvision(ctx, client, clusterSummary)
	}

	return generateDefaultStorageProvision(ctx, client, clusterSummary)
}

func hasDefaultStorageClass(ctx context.Context, client *Client, clusterSummary *ClusterSummary) bool {
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

func generateEksStorageProvision(ctx context.Context, client *Client, clusterSummary *ClusterSummary) StorageProvision {
	if clusterSummary.ServerVersion.LT(semver.MustParse("1.23.0")) {
		return StorageProvision{
			PersistentStorage: true,
			Reason:            "Kubernetes version is less than 1.23.0",
		}
	}

	list, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=aws-ebs-csi-driver",
	})

	if err != nil {
		return StorageProvision{
			PersistentStorage: false,
			Reason:            "Error listing aws csi driver pods",
		}
	}

	if len(list.Items) == 0 {
		return StorageProvision{
			PersistentStorage: false,
			Reason:            "No aws csi driver pods found",
		}
	}

	hasStorageClass := hasDefaultStorageClass(ctx, client, clusterSummary)
	if hasStorageClass {
		return StorageProvision{
			PersistentStorage: true,
			Reason:            "Has default storage class",
		}
	}

	return StorageProvision{
		PersistentStorage: false,
		Reason:            "Has aws ebs dirver without default storage class",
	}

}

func generateDefaultStorageProvision(ctx context.Context, client *Client, clusterSummary *ClusterSummary) StorageProvision {
	if hasDefaultStorageClass(ctx, client, clusterSummary) {
		return StorageProvision{
			PersistentStorage: true,
			Reason:            "Has default storage class",
		}
	}

	return StorageProvision{
		PersistentStorage: false,
		Reason:            "Cluster has no default storage class",
	}
}
