package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ALLIGATOR_LABEL_SELECTOR       = "app=alligator"
	RUNNING_ONLY_LABEL_SELECTOR    = "status.phase=Running"
	GROUNDCOVER_VERSION_ANNOTATION = "groundcover_version"
)

type MetadataFetcher struct {
	kubeconfigPath string
	clientSet      *kubernetes.Clientset
}

func NewMetadataFetcher(kubeconfig string) (*MetadataFetcher, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &MetadataFetcher{
		kubeconfigPath: kubeconfig,
		clientSet:      clientset,
	}, nil
}

func (m *MetadataFetcher) GetNumberOfNodes(ctx context.Context) (int, error) {
	nodes, err := m.clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return len(nodes.Items), nil
}

func (m *MetadataFetcher) GetNumberOfAlligators(ctx context.Context, namespace string, version string) (int, error) {
	alligators, err := m.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: ALLIGATOR_LABEL_SELECTOR,
		FieldSelector: RUNNING_ONLY_LABEL_SELECTOR,
	})
	if err != nil {
		return 0, err
	}

	correctVersionCounter := 0
	for _, alligator := range alligators.Items {
		if alligator.Annotations[GROUNDCOVER_VERSION_ANNOTATION] == version {
			correctVersionCounter++
		}
	}

	return correctVersionCounter, nil
}
