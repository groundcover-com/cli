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

const (
	CLUSTER_TYPE_REPORT_MESSAGE_FORMAT          = "K8s cluster type supported"
	CLUSTER_VERSION_REPORT_MESSAGE_FORMAT       = "K8s server version >= %s"
	CLUSTER_AUTHORIZATION_REPORT_MESSAGE_FORMAT = "K8s user authorized for groundcover installation"
	CLUSTER_CLI_AUTH_SUPPORTED                  = "K8s CLI auth supported"
)

var (
	gkeClusterRegex = regexp.MustCompile("^gke_(?P<project>.+)_(?P<zone>.+)_(?P<name>.+)$")
	eksClusterRegex = regexp.MustCompile("^arn:aws:eks:(?P<region>.+):(?P<account>.+):cluster/(?P<name>.+)$")

	LocalClusterTypes = []string{
		"k3d",
		"kind",
		"minikube",
		"docker-desktop",
	}

	MinimumServerVersionSupport = semver.Version{Major: 1, Minor: 12}
	DefaultClusterRequirements  = &ClusterRequirements{
		ServerVersion: MinimumServerVersionSupport,
		BlockedTypes: []string{
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
	CliAuthSupported     Requirement
	ServerVersionAllowed Requirement
	ClusterTypeAllowed   Requirement
}

func (clusterReport *ClusterReport) IsLocalCluster() bool {
	for _, localCluster := range LocalClusterTypes {
		if strings.HasPrefix(clusterReport.ClusterName, localCluster) {
			return true
		}
	}

	return false
}

func (clusterReport *ClusterReport) PrintStatus() {
	clusterReport.ClusterTypeAllowed.PrintStatus()
	if clusterReport.ClusterTypeAllowed.IsNonCompatible {
		return
	}

	clusterReport.CliAuthSupported.PrintStatus()
	if clusterReport.CliAuthSupported.IsNonCompatible {
		return
	}

	clusterReport.ServerVersionAllowed.PrintStatus()
	if clusterReport.ServerVersionAllowed.IsNonCompatible {
		return
	}

	clusterReport.UserAuthorized.PrintStatus()
	if !clusterReport.UserAuthorized.IsNonCompatible {
		return
	}
}

func (clusterRequirements ClusterRequirements) Validate(ctx context.Context, client *Client, clusterSummary *ClusterSummary) *ClusterReport {
	clusterReport := &ClusterReport{
		ClusterSummary:       clusterSummary,
		UserAuthorized:       clusterRequirements.validateAuthorization(ctx, client, clusterSummary.Namespace),
		CliAuthSupported:     clusterRequirements.validateCliAuthSupported(ctx, clusterSummary.ClusterName),
		ServerVersionAllowed: clusterRequirements.validateServerVersion(client, clusterSummary),
		ClusterTypeAllowed:   clusterRequirements.validateClusterType(clusterSummary.ClusterName),
	}

	clusterReport.IsCompatible = clusterReport.ServerVersionAllowed.IsCompatible &&
		clusterReport.UserAuthorized.IsCompatible &&
		clusterReport.ClusterTypeAllowed.IsCompatible &&
		clusterReport.CliAuthSupported.IsCompatible

	return clusterReport
}

func (clusterRequirements ClusterRequirements) validateClusterType(clusterName string) Requirement {
	var requirement Requirement
	requirement.Message = CLUSTER_TYPE_REPORT_MESSAGE_FORMAT

	for _, blockedType := range clusterRequirements.BlockedTypes {
		if strings.HasPrefix(clusterName, blockedType) {
			requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("%s is unsupported cluster type", blockedType))
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.IsNonCompatible = len(requirement.ErrorMessages) > 0

	return requirement
}

func (clusterRequirements ClusterRequirements) validateServerVersion(client *Client, clusterSummary *ClusterSummary) Requirement {
	var err error

	var requirement Requirement
	requirement.Message = fmt.Sprintf(CLUSTER_VERSION_REPORT_MESSAGE_FORMAT, clusterRequirements.ServerVersion)

	if clusterSummary.ServerVersion, err = client.GetServerVersion(); err != nil {
		requirement.IsNonCompatible = true
		requirement.ErrorMessages = append(requirement.ErrorMessages, err.Error())
		return requirement
	}

	if clusterSummary.ServerVersion.LT(clusterRequirements.ServerVersion) {
		requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("%s is unsupported K8s version", clusterSummary.ServerVersion))
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.IsNonCompatible = len(requirement.ErrorMessages) > 0

	return requirement
}

func (clusterRequirements ClusterRequirements) validateAuthorization(ctx context.Context, client *Client, namespace string) Requirement {
	var err error
	var permitted bool

	var requirement Requirement
	requirement.Message = CLUSTER_AUTHORIZATION_REPORT_MESSAGE_FORMAT

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
	requirement.IsNonCompatible = len(requirement.ErrorMessages) > 0

	return requirement
}

func (clusterRequirements ClusterRequirements) validateCliAuthSupported(ctx context.Context, clusterName string) Requirement {
	var err error

	var requirement Requirement
	requirement.Message = CLUSTER_CLI_AUTH_SUPPORTED

	if !IsEksCluster(clusterName) {
		requirement.IsCompatible = true
		return requirement
	}

	var awsCliVersion semver.Version
	if awsCliVersion, err = DefaultAwsCliVersionValidator.Fetch(ctx); err != nil {
		requirement.IsCompatible = true
		requirement.IsNonCompatible = false
		requirement.ErrorMessages = []string{
			err.Error(),
			HINT_INSTALL_AWS_CLI,
		}
		return requirement
	}

	if err = DefaultAwsCliVersionValidator.Validate(awsCliVersion); err != nil {
		requirement.ErrorMessages = []string{
			err.Error(),
			HINT_EKS_AUTH_PLUGIN_UPGRADE,
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.IsNonCompatible = len(requirement.ErrorMessages) > 0

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
	case IsEksCluster(clusterName):
		return extractRegexClusterName(eksClusterRegex, clusterName)
	case IsGkeCluster(clusterName):
		return extractRegexClusterName(gkeClusterRegex, clusterName)
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

func IsEksCluster(clusterName string) bool {
	return eksClusterRegex.MatchString(clusterName)
}

func IsGkeCluster(clusterName string) bool {
	return gkeClusterRegex.MatchString(clusterName)
}
