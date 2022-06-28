package k8s

import (
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	GKE_CLUSTER_NAME_PREFIX = "gke_"
	EKS_CLUSTER_NAME_PREFIX = "arn:aws:eks"
	AKS_CLUSTER_NAME_PREFIX = "aks-"

	GKE_CLUSTER_NAME_INDEX = 3
	EKS_CLUSTER_NAME_INDEX = 1
	AKS_CLUSTER_NAME_INDEX = 1
)

func GetClusterName(kubeconfigPath string) (string, error) {
	path, err := GetKubeConfigPath(kubeconfigPath)
	if err != nil {
		return "", err
	}

	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil
}

func GetKubeConfigPath(kubeconfigPath string) (string, error) {
	if kubeconfigPath == "" {
		home := homedir.HomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	return kubeconfigPath, nil
}

func FormatClusterName(clusterName string) string {
	if strings.HasPrefix(clusterName, EKS_CLUSTER_NAME_PREFIX) {
		return formatEKSClusterName(clusterName)
	} else if strings.HasPrefix(clusterName, GKE_CLUSTER_NAME_PREFIX) {
		return formatGKEClusterName(clusterName)
	} else if strings.HasPrefix(clusterName, AKS_CLUSTER_NAME_PREFIX) {
		return formatAKSClusterName(clusterName)
	}

	// if we can't identify cloud provider, just return the cluster name
	return clusterName
}

// EKS cluster name format: arn:aws:eks:<region>:<account-id>:cluster/<cluster-name>
func formatEKSClusterName(clusterName string) string {
	splitClusterName := strings.Split(clusterName, "/")
	if len(splitClusterName) != 2 {
		return clusterName
	}

	return splitClusterName[EKS_CLUSTER_NAME_INDEX]
}

// GKE cluster name format: gke_<project_id>_<zone>_<cluster_name>
func formatGKEClusterName(clusterName string) string {
	splitClusterName := strings.Split(clusterName, "_")
	if len(splitClusterName) != 4 {
		return clusterName
	}

	return splitClusterName[GKE_CLUSTER_NAME_INDEX]
}

// azure AKS cluster name format:
func formatAKSClusterName(clusterName string) string {
	splitClusterName := strings.Split(clusterName, "-")
	if len(splitClusterName) != 2 {
		return clusterName
	}

	return splitClusterName[AKS_CLUSTER_NAME_INDEX]
}
