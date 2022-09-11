package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/version"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	GKE_CLUSTER_REGEX = regexp.MustCompile("^gke_(?P<project>.+)_(?P<zone>.+)_(?P<name>.+)$")
	EKS_CLUSTER_REGEX = regexp.MustCompile("^arn:aws:eks:(?P<region>.+):(?P<account>.+):cluster/(?P<name>.+)$")

	MinimumServerVersionSupport = semver.Version{Major: 1, Minor: 12}
	DefaultClusterRequirements  = &ClusterRequirements{
		ServerVersion: MinimumServerVersionSupport,
		BlockedTypes: []string{
			"kind",
			"minikube",
			"docker-desktop",
		},
		Actions: []*authv1.ResourceAttributes{
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
		},
	}
)

type ClusterRequirements struct {
	Actions       []*authv1.ResourceAttributes
	ServerVersion semver.Version
	BlockedTypes  []string
}

type ClusterSummary struct {
	Namespace     string
	ClusterName   string
	ServerVersion semver.Version
}

func (kubeClient *Client) GetClusterSummary(namespace string) (*ClusterSummary, error) {
	var err error

	clusterSummary := &ClusterSummary{
		Namespace: namespace,
	}

	if clusterSummary.ClusterName, err = kubeClient.GetClusterName(); err != nil {
		return clusterSummary, err
	}

	if clusterSummary.ServerVersion, err = kubeClient.GetServerVersion(); err != nil {
		return clusterSummary, err
	}

	return clusterSummary, nil
}

type ClusterReport struct {
	*ClusterSummary
	IsCompatible         bool
	UserAuthorized       Requirement
	ServerVersionAllowed Requirement
	ClusterTypeAllowed   Requirement
}

func (clusterReport *ClusterReport) PrintStatus() {
	clusterReport.ClusterTypeAllowed.PrintStatus()
	if !clusterReport.ClusterTypeAllowed.IsCompatible {
		return
	}

	clusterReport.ServerVersionAllowed.PrintStatus()
	if !clusterReport.ServerVersionAllowed.IsCompatible {
		return
	}

	clusterReport.UserAuthorized.PrintStatus()
}

func (clusterRequirements ClusterRequirements) Validate(ctx context.Context, client *Client, clusterSummary *ClusterSummary) *ClusterReport {
	clusterReport := &ClusterReport{
		ClusterSummary:       clusterSummary,
		ClusterTypeAllowed:   clusterRequirements.validateClusterType(clusterSummary.ClusterName),
		ServerVersionAllowed: clusterRequirements.validateServerVersion(clusterSummary.ServerVersion),
		UserAuthorized:       clusterRequirements.validateAuthorization(ctx, client, clusterSummary.Namespace),
	}

	clusterReport.IsCompatible = clusterReport.ServerVersionAllowed.IsCompatible &&
		clusterReport.UserAuthorized.IsCompatible &&
		clusterReport.ClusterTypeAllowed.IsCompatible

	return clusterReport
}

func (clusterRequirements ClusterRequirements) validateClusterType(clusterName string) Requirement {
	var requirement Requirement

	for _, blockedType := range clusterRequirements.BlockedTypes {
		if strings.HasPrefix(clusterName, blockedType) {
			requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("%s is unsupported cluster type", blockedType))
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.Message = "K8s cluster type supported"

	return requirement
}

func (clusterRequirements ClusterRequirements) validateServerVersion(serverVersion semver.Version) Requirement {
	var requirement Requirement

	if serverVersion.LT(clusterRequirements.ServerVersion) {
		requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("unsupported kernel version %s", serverVersion))
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.Message = fmt.Sprintf("K8s server version >= %s", clusterRequirements.ServerVersion)

	return requirement
}

func (clusterRequirements ClusterRequirements) validateAuthorization(ctx context.Context, client *Client, namespace string) Requirement {
	var err error
	var permitted bool
	var requirement Requirement

	for _, action := range clusterRequirements.Actions {
		action.Namespace = namespace
		if permitted, err = client.isActionPermitted(ctx, action); err != nil {
			requirement.ErrorMessages = append(requirement.ErrorMessages, err.Error())
			continue
		}

		if !permitted {
			requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("denied permissions on resource: %s", action.Resource))
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.Message = "K8s user authorized for groundcover installation"

	return requirement
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

func (kubeClient Client) GetServerVersion() (semver.Version, error) {
	var err error
	var serverVersion semver.Version

	var versionInfo *version.Info
	if versionInfo, err = kubeClient.Discovery().ServerVersion(); err != nil {
		return serverVersion, err
	}

	if serverVersion, err = semver.ParseTolerant(versionInfo.GitVersion); err != nil {
		return serverVersion, errors.Wrapf(err, "unknown server version %s", versionInfo.GitVersion)
	}

	return serverVersion, nil
}
