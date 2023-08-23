package k8s

import (
	"context"
	"errors"

	"github.com/blang/semver/v4"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWS_EBS_CSI_DRIVER_NAME           = "ebs.csi.aws.com"
	AWS_EBS_STORAGE_CLASS_NOT_DEFAULT = "found default storage class without aws-ebs provisioner"

	HINT_INSTALL_AWS_EBS_CSI_DRIVER = `Hint: 
  * Install Amazon EBS CSI driver: https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html`
	HINT_DEFINE_DEFAULT_STORAGE_CLASS = `Hint:
  * Define default StorageClass: https://kubernetes.io/docs/concepts/storage/storage-classes/#the-storageclass-resource`
)

var (
	ErrNoDefaultStorageClass               = errors.New("cluster has no default storage class")
	DEFAULT_STORAGE_CLASS_ANNOTATION_NAMES = []string{"storageclass.kubernetes.io/is-default-class", "storageclass.beta.kubernetes.io/is-default-class"}
)

func (clusterRequirements ClusterRequirements) validateStorage(ctx context.Context, client *Client, clusterSummary *ClusterSummary) Requirement {
	var err error

	var requirement Requirement
	requirement.Message = CLUSTER_STORAGE_SUPPORTED

	if clusterSummary.StorageClass == nil {
		requirement.IsCompatible = false
		requirement.IsNonCompatible = true
		requirement.ErrorMessages = append(requirement.ErrorMessages, ErrNoDefaultStorageClass.Error(), HINT_DEFINE_DEFAULT_STORAGE_CLASS)
		return requirement
	}

	if IsEksCluster(clusterSummary.ClusterName) {
		if semver.MustParseRange("<1.23.0")(clusterSummary.ServerVersion) {
			requirement.IsCompatible = len(requirement.ErrorMessages) == 0
			return requirement
		}

		if err = hasEbsCsiDriver(ctx, client); err != nil {
			requirement.IsCompatible = false
			requirement.IsNonCompatible = true
			requirement.ErrorMessages = append(requirement.ErrorMessages, err.Error(), HINT_INSTALL_AWS_EBS_CSI_DRIVER)
			return requirement
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0

	return requirement
}

func (kubeClient *Client) GetDefaultStorageClass(ctx context.Context) (*v1.StorageClass, error) {
	storageClassList, err := kubeClient.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, storageClass := range storageClassList.Items {
		for _, annotation := range DEFAULT_STORAGE_CLASS_ANNOTATION_NAMES {
			if value, ok := storageClass.Annotations[annotation]; ok && value == "true" {
				return &storageClass, nil
			}
		}
	}

	return nil, ErrNoDefaultStorageClass
}

func hasEbsCsiDriver(ctx context.Context, client *Client) error {
	_, err := client.StorageV1().CSIDrivers().Get(ctx, AWS_EBS_CSI_DRIVER_NAME, metav1.GetOptions{})
	return err
}
