package k8s

import (
	"context"
	"encoding/json"
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

type NodeReport struct {
	NodeSummary            *NodeSummary
	KernelVersionAllowed   bool
	CpuSufficient          bool
	MemorySufficient       bool
	ProviderAllowed        bool
	ArchitectureAllowed    bool
	OperatingSystemAllowed bool
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
		KernelVersionAllowed:   true,
		CpuSufficient:          true,
		MemorySufficient:       true,
		ProviderAllowed:        true,
		ArchitectureAllowed:    true,
		OperatingSystemAllowed: true,
		IsCompatible:           true,
	}

	if !nodeRequirements.isCpuSufficient(node.CPU) {
		nodeReport.CpuSufficient = false
		nodeReport.IsCompatible = false
	}

	if !nodeRequirements.isMemorySufficient(node.Memory) {
		nodeReport.MemorySufficient = false
		nodeReport.IsCompatible = false
	}

	if !nodeRequirements.isProviderAllowed(node.Provider) {
		nodeReport.ProviderAllowed = false
		nodeReport.IsCompatible = false
	}

	if !nodeRequirements.isKernelVersionAllowed(node.Kernel) {
		nodeReport.KernelVersionAllowed = false
		nodeReport.IsCompatible = false
	}

	if !nodeRequirements.isArchitectureAllowed(node.Architecture) {
		nodeReport.ArchitectureAllowed = false
		nodeReport.IsCompatible = false
	}

	if !nodeRequirements.isOperatingSystemAllowed(node.OperatingSystem) {
		nodeReport.OperatingSystemAllowed = false
		nodeReport.IsCompatible = false
	}

	return nodeReport
}

func (nodeRequirements *NodeMinimumRequirements) isCpuSufficient(cpus *resource.Quantity) bool {
	return nodeRequirements.CPUAmount.Cmp(*cpus) < 0
}

func (nodeRequirements *NodeMinimumRequirements) isMemorySufficient(memory *resource.Quantity) bool {
	return nodeRequirements.MemoryAmount.Cmp(*memory) < 0
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

	if nodeRequirements.KernelVersion.GT(kernelVersion) {
		return false
	}

	return true
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
