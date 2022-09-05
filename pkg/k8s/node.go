package k8s

import (
	"context"
	"encoding/json"
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
	PROVIDER_REPORT_MESSAGE_FORMAT         = "%s is unsupported node provider"
	KERNEL_REPORT_MESSAGE_FORMAT           = "%s is unsupported kernel - minimal: %s"
	OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT = "%s is unsupported os - only %s supported"
	CPU_REPORT_MESSAGE_FORMAT              = "insufficient cpu - acutal: %dm / minimal: %s"
	MEMORY_REPORT_MESSAGE_FORMAT           = "insufficient memory - acutal: %dMi / minimal: %s"
	ARCHITECTURE_REPORT_MESSAGE_FORMAT     = "%s is unsupported architecture - only %s supported"
)

var (
	KERNEL_VERSION_REGEX = regexp.MustCompile("^(?P<major>[0-9]).(?P<minor>[0-9]+).(?P<patch>[0-9]+)")

	NodeMinimumCpuRequired      = resource.MustParse(NODE_MINIUM_REQUIREMENTS_CPU)
	NodeMinimumMemoryRequired   = resource.MustParse(NODE_MINIUM_REQUIREMENTS_MEMORY)
	MinimumKernelVersionSupport = semver.Version{Major: 4, Minor: 14}

	NodeRequirements = &NodeMinimumRequirements{
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
			OSImage:         node.Status.NodeInfo.OSImage,
			Architecture:    node.Status.NodeInfo.Architecture,
			Kernel:          node.Status.NodeInfo.KernelVersion,
			OperatingSystem: node.Status.NodeInfo.OperatingSystem,
			CPU:             node.Status.Allocatable.Cpu(),
			Memory:          node.Status.Allocatable.Memory(),
		}
		nodeSummeries = append(nodeSummeries, *nodeSummary)
	}

	return nodeSummeries, nil
}

type NodeRequirementError struct {
	error
}

func NewNodeRequirementError(err error) error {
	return NodeRequirementError{err}
}

func (err NodeRequirementError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Error())
}

type NodeMinimumRequirements struct {
	CPUAmount               *resource.Quantity
	MemoryAmount            *resource.Quantity
	KernelVersion           semver.Version
	BlockedProviders        []string
	AllowedArchitectures    []string
	AllowedOperatingSystems []string
}

type NodeRequirement struct {
	IsCompatible bool
	Message      string
}

type NodeReport struct {
	NodeSummary            *NodeSummary
	KernelVersionAllowed   NodeRequirement
	CpuSufficient          NodeRequirement
	MemorySufficient       NodeRequirement
	ProviderAllowed        NodeRequirement
	ArchitectureAllowed    NodeRequirement
	OperatingSystemAllowed NodeRequirement
	IsCompatible           bool
}

func (nodeRequirements *NodeMinimumRequirements) GenerateNodeReports(nodesSummeries []NodeSummary) ([]*NodeReport, []*NodeReport) {
	var compatible []*NodeReport
	var incompatible []*NodeReport

	for _, node := range nodesSummeries {
		report := nodeRequirements.GetReport(node)
		if report.IsCompatible {
			compatible = append(compatible, report)
		} else {
			incompatible = append(incompatible, report)
		}
	}

	return compatible, incompatible
}

func (nodeRequirements *NodeMinimumRequirements) GetReport(node NodeSummary) *NodeReport {
	nodeReport := &NodeReport{
		NodeSummary:            &node,
		KernelVersionAllowed:   nodeRequirements.isKernelVersionAllowed(node.Kernel),
		CpuSufficient:          nodeRequirements.isCpuSufficient(node.CPU),
		MemorySufficient:       nodeRequirements.isMemorySufficient(node.Memory),
		ProviderAllowed:        nodeRequirements.isProviderAllowed(node.Provider),
		ArchitectureAllowed:    nodeRequirements.isArchitectureAllowed(node.Architecture),
		OperatingSystemAllowed: nodeRequirements.isOperatingSystemAllowed(node.OperatingSystem),
	}

	nodeReport.IsCompatible = nodeReport.KernelVersionAllowed.IsCompatible &&
		nodeReport.CpuSufficient.IsCompatible &&
		nodeReport.MemorySufficient.IsCompatible &&
		nodeReport.ProviderAllowed.IsCompatible &&
		nodeReport.ArchitectureAllowed.IsCompatible &&
		nodeReport.OperatingSystemAllowed.IsCompatible

	return nodeReport
}

func (nodeRequirements *NodeMinimumRequirements) isCpuSufficient(cpus *resource.Quantity) NodeRequirement {
	if nodeRequirements.CPUAmount.Cmp(*cpus) > 0 {
		return NodeRequirement{
			IsCompatible: false,
			Message:      fmt.Sprintf(CPU_REPORT_MESSAGE_FORMAT, cpus.ScaledValue(resource.Milli), nodeRequirements.CPUAmount.String()),
		}
	}

	return NodeRequirement{IsCompatible: true}
}

func (nodeRequirements *NodeMinimumRequirements) isMemorySufficient(memory *resource.Quantity) NodeRequirement {
	if nodeRequirements.MemoryAmount.Cmp(*memory) > 0 {
		return NodeRequirement{
			IsCompatible: false,
			Message:      fmt.Sprintf(MEMORY_REPORT_MESSAGE_FORMAT, memory.ScaledValue(resource.Mega), nodeRequirements.MemoryAmount.String()),
		}
	}

	return NodeRequirement{IsCompatible: true}
}

func (nodeRequirements *NodeMinimumRequirements) isProviderAllowed(provider string) NodeRequirement {
	for _, blockedProvider := range nodeRequirements.BlockedProviders {
		if strings.Contains(provider, blockedProvider) {
			return NodeRequirement{
				IsCompatible: false,
				Message:      fmt.Sprintf(PROVIDER_REPORT_MESSAGE_FORMAT, provider),
			}
		}
	}

	return NodeRequirement{IsCompatible: true}
}

func (nodeRequirements *NodeMinimumRequirements) isKernelVersionAllowed(kernel string) NodeRequirement {
	var err error
	var kernelVersion semver.Version

	if kernelVersion, err = semver.Parse(KERNEL_VERSION_REGEX.FindString(kernel)); err != nil {
		return NodeRequirement{
			IsCompatible: false,
			Message:      fmt.Sprintf(KERNEL_REPORT_MESSAGE_FORMAT, kernel, nodeRequirements.KernelVersion.String()),
		}
	}

	if nodeRequirements.KernelVersion.GT(kernelVersion) {
		return NodeRequirement{
			IsCompatible: false,
			Message:      fmt.Sprintf(KERNEL_REPORT_MESSAGE_FORMAT, kernel, nodeRequirements.KernelVersion.String()),
		}
	}

	return NodeRequirement{IsCompatible: true}
}

func (nodeRequirements *NodeMinimumRequirements) isArchitectureAllowed(architecture string) NodeRequirement {
	for _, allowedArchitecture := range nodeRequirements.AllowedArchitectures {
		if allowedArchitecture == architecture {
			return NodeRequirement{IsCompatible: true}
		}
	}

	return NodeRequirement{
		IsCompatible: false,
		Message:      fmt.Sprintf(ARCHITECTURE_REPORT_MESSAGE_FORMAT, architecture, strings.Join(nodeRequirements.AllowedArchitectures, ", ")),
	}
}

func (nodeRequirements *NodeMinimumRequirements) isOperatingSystemAllowed(operatingSystem string) NodeRequirement {
	for _, allowedOperatingSystem := range nodeRequirements.AllowedOperatingSystems {
		if allowedOperatingSystem == operatingSystem {
			return NodeRequirement{IsCompatible: true}
		}
	}

	return NodeRequirement{
		IsCompatible: false,
		Message:      fmt.Sprintf(OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT, operatingSystem, strings.Join(nodeRequirements.AllowedOperatingSystems, ", ")),
	}
}
