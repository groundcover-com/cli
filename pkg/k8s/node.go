package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NODE_MINIUM_REQUIREMENTS_CPU           = "1750m"
	NODE_MINIUM_REQUIREMENTS_MEMORY        = "1750Mi"
	CPU_REPORT_MESSAGE_FORMAT              = "Sufficient node CPU (%d/%d Nodes)"
	KERNEL_REPORT_MESSAGE_FORMAT           = "Kernel version >= %s (%d/%d Nodes)"
	MEMORY_REPORT_MESSAGE_FORMAT           = "Sufficient node memory (%d/%d Nodes)"
	PROVIDER_REPORT_MESSAGE_FORMAT         = "Cloud provider supported (%d/%d Nodes)"
	ARCHITECTURE_REPORT_MESSAGE_FORMAT     = "Node architecture supported (%d/%d Nodes)"
	OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT = "Node operating system supported (%d/%d Nodes)"
)

var (
	KERNEL_VERSION_REGEX = regexp.MustCompile("^(?P<major>[0-9]).(?P<minor>[0-9]+).(?P<patch>[0-9]+)")

	NodeMinimumCpuRequired      = resource.MustParse(NODE_MINIUM_REQUIREMENTS_CPU)
	NodeMinimumMemoryRequired   = resource.MustParse(NODE_MINIUM_REQUIREMENTS_MEMORY)
	MinimumKernelVersionSupport = semver.Version{Major: 4, Minor: 14}

	DefaultNodeRequirements = &NodeMinimumRequirements{
		CPUAmount:               &NodeMinimumCpuRequired,
		MemoryAmount:            &NodeMinimumMemoryRequired,
		AllowedOperatingSystems: []string{"linux"},
		AllowedArchitectures:    []string{"amd64"},
		BlockedProviders:        []string{"fargate"},
		KernelVersion:           MinimumKernelVersionSupport,
	}
)

type NodeSummary struct {
	CPU             *resource.Quantity `json:",omitempty"`
	Memory          *resource.Quantity `json:",omitempty"`
	Name            string             `json:"-"`
	Kernel          string             `json:",omitempty"`
	Provider        string             `json:",omitempty"`
	OSImage         string             `json:",omitempty"`
	Architecture    string             `json:",omitempty"`
	OperatingSystem string             `json:",omitempty"`
}

func (kubeClient *Client) GetNodesSummeries(ctx context.Context) ([]*NodeSummary, error) {
	var err error

	var nodeList *v1.NodeList
	if nodeList, err = kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	var nodeSummeries []*NodeSummary
	for _, node := range nodeList.Items {
		nodeSummary := &NodeSummary{
			Name:            node.ObjectMeta.Name,
			Provider:        node.Spec.ProviderID,
			OSImage:         node.Status.NodeInfo.OSImage,
			Architecture:    node.Status.NodeInfo.Architecture,
			Kernel:          node.Status.NodeInfo.KernelVersion,
			OperatingSystem: node.Status.NodeInfo.OperatingSystem,
			CPU:             node.Status.Allocatable.Cpu(),
			Memory:          node.Status.Allocatable.Memory(),
		}
		nodeSummeries = append(nodeSummeries, nodeSummary)
	}

	return nodeSummeries, nil
}

type NodeMinimumRequirements struct {
	CPUAmount               *resource.Quantity
	MemoryAmount            *resource.Quantity
	KernelVersion           semver.Version
	BlockedProviders        []string
	AllowedArchitectures    []string
	AllowedOperatingSystems []string
}

type NodesReport struct {
	KernelVersionAllowed   Requirement
	CpuSufficient          Requirement
	MemorySufficient       Requirement
	ProviderAllowed        Requirement
	ArchitectureAllowed    Requirement
	OperatingSystemAllowed Requirement
	CompatibleNodes        []*NodeSummary      `json:"-"`
	IncompatibleNodes      []*IncompatibleNode `json:",omitempty"`
}

func (nodesReport *NodesReport) PrintStatus() {
	nodesReport.CpuSufficient.PrintStatus()
	nodesReport.MemorySufficient.PrintStatus()
	nodesReport.KernelVersionAllowed.PrintStatus()
	nodesReport.ArchitectureAllowed.PrintStatus()
	nodesReport.OperatingSystemAllowed.PrintStatus()
	nodesReport.ProviderAllowed.PrintStatus()
}

type IncompatibleNode struct {
	*NodeSummary
	RequirementErrors []string
}

func (nodeRequirements *NodeMinimumRequirements) Validate(nodesSummeries []*NodeSummary) *NodesReport {
	var err error
	var nodesReport NodesReport

	for _, nodeSummary := range nodesSummeries {
		var requirementErrors []string

		if err = nodeRequirements.validateNodeCPU(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.CpuSufficient.ErrorMessages = append(
				nodesReport.CpuSufficient.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if err = nodeRequirements.validateNodeMemory(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.MemorySufficient.ErrorMessages = append(
				nodesReport.MemorySufficient.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if err = nodeRequirements.validateNodeProvider(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.ProviderAllowed.ErrorMessages = append(
				nodesReport.ProviderAllowed.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if err = nodeRequirements.validateNodeKernelVersion(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.KernelVersionAllowed.ErrorMessages = append(
				nodesReport.KernelVersionAllowed.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if err = nodeRequirements.validateNodeArchitecture(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.ArchitectureAllowed.ErrorMessages = append(
				nodesReport.ArchitectureAllowed.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if err = nodeRequirements.validateNodeOperatingSystem(nodeSummary); err != nil {
			requirementErrors = append(requirementErrors, err.Error())
			nodesReport.OperatingSystemAllowed.ErrorMessages = append(
				nodesReport.OperatingSystemAllowed.ErrorMessages,
				fmt.Sprintf("node: %s - %s", nodeSummary.Name, err.Error()),
			)
		}

		if len(requirementErrors) > 0 {
			nodesReport.IncompatibleNodes = append(
				nodesReport.IncompatibleNodes,
				&IncompatibleNode{
					NodeSummary:       nodeSummary,
					RequirementErrors: requirementErrors,
				},
			)
			continue
		}

		nodesReport.CompatibleNodes = append(nodesReport.CompatibleNodes, nodeSummary)
	}

	nodesReport.CpuSufficient.IsCompatible = len(nodesReport.CpuSufficient.ErrorMessages) == 0
	nodesReport.CpuSufficient.Message = fmt.Sprintf(
		CPU_REPORT_MESSAGE_FORMAT,
		len(nodesSummeries)-len(nodesReport.CpuSufficient.ErrorMessages),
		len(nodesSummeries),
	)

	nodesReport.MemorySufficient.IsCompatible = len(nodesReport.MemorySufficient.ErrorMessages) == 0
	nodesReport.MemorySufficient.Message = fmt.Sprintf(
		MEMORY_REPORT_MESSAGE_FORMAT,
		len(nodesSummeries)-len(nodesReport.MemorySufficient.ErrorMessages),
		len(nodesSummeries),
	)

	nodesReport.ProviderAllowed.IsCompatible = len(nodesReport.ProviderAllowed.ErrorMessages) == 0
	nodesReport.ProviderAllowed.Message = fmt.Sprintf(
		PROVIDER_REPORT_MESSAGE_FORMAT,
		len(nodesSummeries)-len(nodesReport.ProviderAllowed.ErrorMessages),
		len(nodesSummeries),
	)

	nodesReport.KernelVersionAllowed.IsCompatible = len(nodesReport.KernelVersionAllowed.ErrorMessages) == 0
	nodesReport.KernelVersionAllowed.Message = fmt.Sprintf(
		KERNEL_REPORT_MESSAGE_FORMAT,
		MinimumKernelVersionSupport,
		len(nodesSummeries)-len(nodesReport.KernelVersionAllowed.ErrorMessages),
		len(nodesSummeries),
	)

	nodesReport.ArchitectureAllowed.IsCompatible = len(nodesReport.ArchitectureAllowed.ErrorMessages) == 0
	nodesReport.ArchitectureAllowed.Message = fmt.Sprintf(
		ARCHITECTURE_REPORT_MESSAGE_FORMAT,
		len(nodesSummeries)-len(nodesReport.ArchitectureAllowed.ErrorMessages),
		len(nodesSummeries),
	)

	nodesReport.OperatingSystemAllowed.IsCompatible = len(nodesReport.OperatingSystemAllowed.ErrorMessages) == 0
	nodesReport.OperatingSystemAllowed.Message = fmt.Sprintf(
		OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT,
		len(nodesSummeries)-len(nodesReport.OperatingSystemAllowed.ErrorMessages),
		len(nodesSummeries),
	)

	return &nodesReport
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeCPU(nodeSummary *NodeSummary) error {
	if nodeRequirements.CPUAmount.Cmp(*nodeSummary.CPU) > 0 {
		return fmt.Errorf("insufficient cpu %s < %s", nodeSummary.CPU, nodeRequirements.CPUAmount)
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeMemory(nodeSummary *NodeSummary) error {
	if nodeRequirements.MemoryAmount.Cmp(*nodeSummary.Memory) > 0 {
		return fmt.Errorf("insufficient memory %s < %s", nodeSummary.Memory, nodeRequirements.MemoryAmount)
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeProvider(nodeSummary *NodeSummary) error {
	for _, blockedProvider := range nodeRequirements.BlockedProviders {
		if strings.Contains(nodeSummary.Provider, blockedProvider) {
			return fmt.Errorf("%s is unsupported provider", blockedProvider)
		}
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeKernelVersion(nodeSummary *NodeSummary) error {
	var err error

	var kernelVersion semver.Version
	if kernelVersion, err = semver.Parse(KERNEL_VERSION_REGEX.FindString(nodeSummary.Kernel)); err != nil {
		return fmt.Errorf("%s is unknown kernel version", nodeSummary.Kernel)
	}

	if nodeRequirements.KernelVersion.GT(kernelVersion) {
		return fmt.Errorf("%s is unsupported kernel version", nodeSummary.Kernel)
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeArchitecture(nodeSummary *NodeSummary) error {
	for _, allowedArchitecture := range nodeRequirements.AllowedArchitectures {
		if allowedArchitecture == nodeSummary.Architecture {
			return nil
		}
	}

	return fmt.Errorf("%s is unspported architecture", nodeSummary.Architecture)
}

func (nodeRequirements *NodeMinimumRequirements) validateNodeOperatingSystem(nodeSummary *NodeSummary) error {
	for _, allowedOperatingSystem := range nodeRequirements.AllowedOperatingSystems {
		if allowedOperatingSystem == nodeSummary.OperatingSystem {
			return nil
		}
	}

	return fmt.Errorf("%s is unspported operating system", nodeSummary.OperatingSystem)
}
