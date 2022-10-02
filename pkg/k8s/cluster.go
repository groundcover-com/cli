package k8s

import (
	"context"
	"fmt"
	"os/exec"
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

	INSTALL_AWS_CLI_HINT = `Hint: 
  * Install aws cli by following https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html`
)

var (
	gkeClusterRegex    = regexp.MustCompile("^gke_(?P<project>.+)_(?P<zone>.+)_(?P<name>.+)$")
	eksClusterRegex    = regexp.MustCompile("^arn:aws:eks:(?P<region>.+):(?P<account>.+):cluster/(?P<name>.+)$")
	awsCliVersionRegex = regexp.MustCompile(`aws-cli/(\d+\.\d+\.\d+)`)

	MinSupportedAwsCliV1Version = semver.Version{Major: 1, Minor: 23, Patch: 9}
	MinSupportedAwsCliV2Version = semver.Version{Major: 2, Minor: 7, Patch: 0}

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
	CliAuthSupported     Requirement
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
	if !clusterReport.UserAuthorized.IsCompatible {
		return
	}

	clusterReport.CliAuthSupported.PrintStatus()
}

func (clusterRequirements ClusterRequirements) Validate(ctx context.Context, client *Client, clusterSummary *ClusterSummary) *ClusterReport {
	clusterReport := &ClusterReport{
		ClusterSummary:       clusterSummary,
		UserAuthorized:       clusterRequirements.validateAuthorization(ctx, client, clusterSummary.Namespace),
		CliAuthSupported:     clusterRequirements.validateCliAuthSupported(ctx, clusterSummary.ClusterName),
		ServerVersionAllowed: clusterRequirements.validateServerVersion(clusterSummary.ServerVersion),
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

	for _, blockedType := range clusterRequirements.BlockedTypes {
		if strings.HasPrefix(clusterName, blockedType) {
			requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("%s is unsupported cluster type", blockedType))
		}
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.IsNonCompatible = requirement.IsCompatible
	requirement.Message = CLUSTER_TYPE_REPORT_MESSAGE_FORMAT

	return requirement
}

func (clusterRequirements ClusterRequirements) validateServerVersion(serverVersion semver.Version) Requirement {
	var requirement Requirement

	if serverVersion.LT(clusterRequirements.ServerVersion) {
		requirement.ErrorMessages = append(requirement.ErrorMessages, fmt.Sprintf("%s is unsupported K8s version", serverVersion))
	}

	requirement.IsCompatible = len(requirement.ErrorMessages) == 0
	requirement.IsNonCompatible = requirement.IsCompatible
	requirement.Message = fmt.Sprintf(CLUSTER_VERSION_REPORT_MESSAGE_FORMAT, clusterRequirements.ServerVersion)

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
	requirement.IsNonCompatible = requirement.IsCompatible
	requirement.Message = CLUSTER_AUTHORIZATION_REPORT_MESSAGE_FORMAT

	return requirement
}

func (clusterRequirements ClusterRequirements) validateCliAuthSupported(ctx context.Context, clusterName string) Requirement {
	if !eksClusterRegex.MatchString(clusterName) {
		return Requirement{
			Message:      CLUSTER_CLI_AUTH_SUPPORTED,
			IsCompatible: true,
		}
	}

	awsCliCmd := exec.CommandContext(ctx, "aws", "--version")
	awsVersion, err := awsCliCmd.Output()
	if err != nil {
		return Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         CLUSTER_CLI_AUTH_SUPPORTED,
			ErrorMessages:   []string{"Failed getting aws cli version, make sure aws cli is installed", INSTALL_AWS_CLI_HINT},
		}
	}

	strippedAwsVersion := strings.TrimSuffix(string(awsVersion), "\n")
	supported := ValidateAwsCliVersionSupported(strippedAwsVersion)
	if !supported {
		return Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         CLUSTER_CLI_AUTH_SUPPORTED,
			ErrorMessages:   []string{fmt.Sprintf("Unsupported aws-cli version: %q", strippedAwsVersion), HINT_EKS_AUTH_PLUGIN_UPGRADE},
		}
	}

	return Requirement{
		Message:      CLUSTER_CLI_AUTH_SUPPORTED,
		IsCompatible: true,
	}
}

func ValidateAwsCliVersionSupported(version string) bool {
	matches := awsCliVersionRegex.FindStringSubmatch(version)
	if len(matches) != 2 {
		return false
	}

	awsCliVersion, err := semver.Parse(matches[1])
	if err != nil {
		return false
	}

	switch awsCliVersion.Major {
	case 1:
		return awsCliVersion.GTE(MinSupportedAwsCliV1Version)
	case 2:
		return awsCliVersion.GTE(MinSupportedAwsCliV2Version)
	default:
		return false
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
	case eksClusterRegex.MatchString(clusterName):
		return extractRegexClusterName(eksClusterRegex, clusterName)
	case gkeClusterRegex.MatchString(clusterName):
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
