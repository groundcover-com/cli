package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
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
}

type ClusterReport struct {
	IsCompatible         bool
	UserAuthorized       Requirement
	ServerVersionAllowed Requirement
}

func (clusterRequirements ClusterRequirements) Validate(ctx context.Context, client *Client, namespace string) *ClusterReport {
	clusterReport := &ClusterReport{
		ServerVersionAllowed: clusterRequirements.validateServerVersion(client),
		UserAuthorized:       clusterRequirements.validateAuthorization(ctx, client, namespace),
	}

	clusterReport.IsCompatible = clusterReport.ServerVersionAllowed.IsCompatible &&
		clusterReport.UserAuthorized.IsCompatible

	return clusterReport
}

func (clusterRequirements ClusterRequirements) validateServerVersion(client *Client) Requirement {
	var err error

	var versionInfo *version.Info
	if versionInfo, err = client.Discovery().ServerVersion(); err != nil {
		return Requirement{
			IsCompatible: false,
			Message:      err.Error(),
		}
	}

	var serverVersion semver.Version
	if serverVersion, err = semver.ParseTolerant(versionInfo.GitVersion); err != nil {
		return Requirement{
			IsCompatible: false,
			Message:      fmt.Sprintf("unknown server version: %s", versionInfo),
		}
	}

	if serverVersion.LT(clusterRequirements.ServerVersion) {
		return Requirement{
			IsCompatible: false,
			Message:      fmt.Sprintf("%s is unsupported cluster version - minimal: %s", serverVersion, clusterRequirements.ServerVersion),
		}
	}

	return Requirement{
		IsCompatible: true,
		Message:      fmt.Sprintf("Server version >= %s", clusterRequirements.ServerVersion),
	}
}

func (clusterRequirements ClusterRequirements) validateAuthorization(ctx context.Context, client *Client, namespace string) Requirement {
	var err error
	var permitted bool
	var deniedResources []string

	for _, action := range clusterRequirements.Actions {
		action.Namespace = namespace
		if permitted, err = client.isActionPermitted(ctx, action); err != nil {
			return Requirement{
				IsCompatible: false,
				Message:      err.Error(),
			}
		}

		if !permitted {
			deniedResources = append(deniedResources, action.Resource)
		}
	}

	if len(deniedResources) > 0 {
		return Requirement{
			IsCompatible: false,
			Message:      fmt.Sprintf("denied permissions on resources: %s", strings.Join(deniedResources, ", ")),
		}
	}

	return Requirement{
		IsCompatible: true,
		Message:      "User authorized",
	}
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
