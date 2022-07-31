package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PROVIDER_REPORT_MESSAGE_FORMAT         = "%s is unsupported node provider"
	KERNEL_REPORT_MESSAGE_FORMAT           = "%s is unsupported kernel - minimal: %s"
	OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT = "%s is unsupported os - only %s supported"
	CPU_REPORT_MESSAGE_FORMAT              = "insufficient cpu - acutal: %d / minimal: %d"
	MEMORY_REPORT_MESSAGE_FORMAT           = "insufficient memory - acutal: %dG / minimal: %dG"
	ARCHITECTURE_REPORT_MESSAGE_FORMAT     = "%s is unsupported architecture - only %s supported"
)

var (
	KERNEL_VERSION_REGEX = regexp.MustCompile("^(?P<major>[0-9]).(?P<minor>[0-9]+).(?P<patch>[0-9]+)")
)

type NodeSummary struct {
	CPU             int64
	Memory          int64
	Name            string
	Kernel          string
	Provider        string
	Architecture    string
	OperatingSystem string
}

func (kubeClient *Client) GetNodesSummeries(ctx context.Context) ([]NodeSummary, error) {
	var err error

	var nodeList *v1.NodeList
	if nodeList, err = kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	var nodeSummeries []NodeSummary
	for _, node := range nodeList.Items {
		nodeSummary := &NodeSummary{
			Name:            node.ObjectMeta.Name,
			Provider:        node.Spec.ProviderID,
			Architecture:    node.Status.NodeInfo.Architecture,
			Kernel:          node.Status.NodeInfo.KernelVersion,
			OperatingSystem: node.Status.NodeInfo.OperatingSystem,
			CPU:             node.Status.Allocatable.Cpu().Value(),
			Memory:          node.Status.Allocatable.Memory().Value() / (1 << 30),
		}
		nodeSummeries = append(nodeSummeries, *nodeSummary)
	}

	return nodeSummeries, nil
}

type NodeMinimumRequirements struct {
	CPUAmount               int64
	MemoryAmount            int64
	KernelVersion           semver.Version
	BlockedProviders        []string
	AllowedArchitectures    []string
	AllowedOperatingSystems []string
	report                  map[string][]string
}

func NewNodeMinimumRequirements() *NodeMinimumRequirements {
	return &NodeMinimumRequirements{
		CPUAmount:               2,
		MemoryAmount:            4,
		AllowedOperatingSystems: []string{"linux"},
		AllowedArchitectures:    []string{"amd64"},
		BlockedProviders:        []string{"fargate"},
		KernelVersion:           semver.Version{Major: 4, Minor: 14},
		report:                  make(map[string][]string),
	}
}

func (nodeRequirements *NodeMinimumRequirements) CheckAndAppendReport(node NodeSummary) bool {
	var report []string

	if !nodeRequirements.isCpuSufficient(node.CPU) {
		report = append(report, fmt.Sprintf(CPU_REPORT_MESSAGE_FORMAT, node.CPU, nodeRequirements.CPUAmount))
	}

	if !nodeRequirements.isMemorySufficient(node.Memory) {
		report = append(report, fmt.Sprintf(MEMORY_REPORT_MESSAGE_FORMAT, node.Memory, nodeRequirements.MemoryAmount))
	}

	if !nodeRequirements.isProviderAllowed(node.Provider) {
		report = append(report, fmt.Sprintf(PROVIDER_REPORT_MESSAGE_FORMAT, node.Provider))
	}

	if !nodeRequirements.isKernelVersionAllowed(node.Kernel) {
		report = append(report, fmt.Sprintf(KERNEL_REPORT_MESSAGE_FORMAT, node.Kernel, nodeRequirements.KernelVersion))
	}

	if !nodeRequirements.isArchitectureAllowed(node.Architecture) {
		report = append(report, fmt.Sprintf(ARCHITECTURE_REPORT_MESSAGE_FORMAT, node.Architecture, strings.Join(nodeRequirements.AllowedArchitectures, ", ")))
	}

	if !nodeRequirements.isOperatingSystemAllowed(node.OperatingSystem) {
		report = append(report, fmt.Sprintf(OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT, node.OperatingSystem, strings.Join(nodeRequirements.AllowedOperatingSystems, ", ")))
	}

	if len(report) > 0 {
		nodeRequirements.report[node.Name] = report
		return false
	}

	return true
}

func (nodeRequirements *NodeMinimumRequirements) isCpuSufficient(cpus int64) bool {
	return cpus >= nodeRequirements.CPUAmount
}

func (nodeRequirements *NodeMinimumRequirements) isMemorySufficient(memory int64) bool {
	return memory >= nodeRequirements.MemoryAmount
}

func (nodeRequirements *NodeMinimumRequirements) isProviderAllowed(provider string) bool {
	for _, blockedProvider := range nodeRequirements.BlockedProviders {
		if strings.Contains(provider, blockedProvider) {
			return false
		}
	}

	return true
}

func (nodeRequirements *NodeMinimumRequirements) isKernelVersionAllowed(kernel string) bool {
	var err error
	var kernelVersion semver.Version

	if kernelVersion, err = semver.Parse(KERNEL_VERSION_REGEX.FindString(kernel)); err != nil {
		return false
	}

	return kernelVersion.GTE(nodeRequirements.KernelVersion)
}

func (nodeRequirements *NodeMinimumRequirements) isArchitectureAllowed(architecture string) bool {
	for _, allowedArchitecture := range nodeRequirements.AllowedArchitectures {
		if allowedArchitecture == architecture {
			return true
		}
	}

	return false
}

func (nodeRequirements *NodeMinimumRequirements) isOperatingSystemAllowed(operatingSystem string) bool {
	for _, allowedOperatingSystem := range nodeRequirements.AllowedOperatingSystems {
		if allowedOperatingSystem == operatingSystem {
			return true
		}
	}

	return false
}
