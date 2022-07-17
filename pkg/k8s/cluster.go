package k8s

import (
	"fmt"
	"regexp"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	GKE_CLUSTER_REGEX = regexp.MustCompile("^gke_(?P<project>.+)_(?P<zone>.+)_(?P<name>.+)$")
	EKS_CLUSTER_REGEX = regexp.MustCompile("^arn:aws:eks:(?P<region>.+):(?P<account>.+):cluster/(?P<name>.+)$")
)

func (kuber *Kuber) GetClusterName() (string, error) {
	var err error
	var rawConfig clientcmdapi.Config

	if rawConfig, err = kuber.config.RawConfig(); err != nil {
		return "", err
	}

	return rawConfig.Contexts[kuber.kubecontext].Cluster, nil
}

func (kuber *Kuber) GetClusterShortName() (string, error) {
	var err error
	var clusterName string

	if clusterName, err = kuber.GetClusterName(); err != nil {
		return "", err
	}

	switch {
	case EKS_CLUSTER_REGEX.MatchString(clusterName):
		return extractRegexClusterName(EKS_CLUSTER_REGEX, clusterName)
	case GKE_CLUSTER_REGEX.MatchString(clusterName):
		return extractRegexClusterName(GKE_CLUSTER_REGEX, clusterName)
	default:
		return clusterName, nil
	}
}

func extractRegexClusterName(regex *regexp.Regexp, clusterName string) (string, error) {
	var subIndex int

	subMatch := regex.FindStringSubmatch(clusterName)

	if subIndex = regex.SubexpIndex("name"); subIndex == -1 {
		return "", fmt.Errorf("failed to extract cluster name from: %s", clusterName)
	}

	return subMatch[subIndex], nil
}
