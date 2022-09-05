package k8s

import (
	"context"
	"fmt"
	"regexp"

	authv1 "k8s.io/api/authorization/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	GKE_CLUSTER_REGEX = regexp.MustCompile("^gke_(?P<project>.+)_(?P<zone>.+)_(?P<name>.+)$")
	EKS_CLUSTER_REGEX = regexp.MustCompile("^arn:aws:eks:(?P<region>.+):(?P<account>.+):cluster/(?P<name>.+)$")
)

type ClusterRequirements struct {
	Actions []*authv1.ResourceAttributes
}

func NewClusterRequirements() *ClusterRequirements {

	actions := []*authv1.ResourceAttributes{
		{
			Verb:     "*",
			Resource: "clusterroles",
		},
		{
			Verb:     "*",
			Resource: "configmaps",
		},
		{
			Verb:     "*",
			Resource: "daemonsets",
		},
		{
			Verb:     "*",
			Resource: "deployments",
		},
		{
			Verb:     "*",
			Resource: "ingresses",
		},
		{
			Verb:     "*",
			Resource: "namespaces",
		},
		{
			Verb:     "*",
			Resource: "nodes",
		},
		{
			Verb:     "*",
			Resource: "pods",
		},
		{
			Verb:     "*",
			Resource: "secrets",
		},
		{
			Verb:     "*",
			Resource: "services",
		},
		{
			Verb:     "*",
			Resource: "statefulsets",
		},
		{
			Verb:     "*",
			Resource: "persistentvolumeclaims",
		},
		{
			Verb:     "*",
			Resource: "persistentvolumes",
		},
	}

	return &ClusterRequirements{
		Actions: actions,
	}
}

func (clusterRequirements ClusterRequirements) Validate(ctx context.Context, client *Client, namespace string) []error {
	if authErrors := clusterRequirements.validateAuthorization(ctx, client, namespace); len(authErrors) > 0 {
		return authErrors
	}

	return []error{}
}

func (clusterRequirements ClusterRequirements) validateAuthorization(ctx context.Context, client *Client, namespace string) []error {
	var err error
	var errSlice []error

	for _, action := range clusterRequirements.Actions {
		action.Namespace = namespace
		if err = client.isActionPermitted(ctx, action); err != nil {
			errSlice = append(errSlice, err)
		}
	}

	return errSlice
}

func (kubeClient *Client) GetClusterName() (string, error) {
	var err error
	var rawConfig clientcmdapi.Config

	if rawConfig, err = kubeClient.RawConfig(); err != nil {
		return "", err
	}

	return rawConfig.Contexts[kubeClient.kubecontext].Cluster, nil
}

func (kubeClient *Client) GetClusterShortName() (string, error) {
	var err error
	var clusterName string

	if clusterName, err = kubeClient.GetClusterName(); err != nil {
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
