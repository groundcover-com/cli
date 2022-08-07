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
	CPU             int64  `json:",omitempty"`
	Memory          int64  `json:",omitempty"`
	Name            string `json:"-"`
	Kernel          string `json:",omitempty"`
	Provider        string `json:",omitempty"`
	OSImage         string `json:",omitempty"`
	Architecture    string `json:",omitempty"`
	OperatingSystem string `json:",omitempty"`
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
			CPU:             node.Status.Allocatable.Cpu().Value(),
			Memory:          node.Status.Allocatable.Memory().ScaledValue(resource.Giga),
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
	CPUAmount               int64
	MemoryAmount            int64
	KernelVersion           semver.Version
	BlockedProviders        []string
	AllowedArchitectures    []string
	AllowedOperatingSystems []string
}

type NodeReport struct {
	*NodeSummary
	IsAdequate bool    `json:",omitempty"`
	Errors     []error `json:",omitempty"`
}

func NewNodeMinimumRequirements() *NodeMinimumRequirements {
	return &NodeMinimumRequirements{
		CPUAmount:               1,
		MemoryAmount:            2,
		AllowedOperatingSystems: []string{"linux"},
		AllowedArchitectures:    []string{"amd64"},
		BlockedProviders:        []string{"fargate"},
		KernelVersion:           semver.Version{Major: 4, Minor: 14},
	}
}

func (nodeRequirements *NodeMinimumRequirements) GenerateNodeReports(nodesSummeries []NodeSummary) ([]*NodeReport, []*NodeReport) {
	var adequates []*NodeReport
	var inadequates []*NodeReport

	for _, node := range nodesSummeries {
		report := nodeRequirements.GetReport(node)
		if report.IsAdequate {
			adequates = append(adequates, report)
		} else {
			inadequates = append(inadequates, report)
		}
	}

	return adequates, inadequates
}

func (nodeRequirements *NodeMinimumRequirements) GetReport(node NodeSummary) *NodeReport {
	var err error
	var errors []error

	if err = nodeRequirements.isCpuSufficient(node.CPU); err != nil {
		errors = append(errors, err)
	}

	if err = nodeRequirements.isMemorySufficient(node.Memory); err != nil {
		errors = append(errors, err)
	}

	if err = nodeRequirements.isProviderAllowed(node.Provider); err != nil {
		errors = append(errors, err)
	}

	if err = nodeRequirements.isKernelVersionAllowed(node.Kernel); err != nil {
		errors = append(errors, err)
	}

	if err = nodeRequirements.isArchitectureAllowed(node.Architecture); err != nil {
		errors = append(errors, err)
	}

	if err = nodeRequirements.isOperatingSystemAllowed(node.OperatingSystem); err != nil {
		errors = append(errors, err)
	}

	return &NodeReport{
		NodeSummary: &node,
		Errors:      errors,
		IsAdequate:  len(errors) == 0,
	}
}

func (nodeRequirements *NodeMinimumRequirements) isCpuSufficient(cpus int64) error {
	if nodeRequirements.CPUAmount > cpus {
		return NewNodeRequirementError(fmt.Errorf(CPU_REPORT_MESSAGE_FORMAT, cpus, nodeRequirements.CPUAmount))
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) isMemorySufficient(memory int64) error {
	if nodeRequirements.MemoryAmount > memory {
		return NewNodeRequirementError(fmt.Errorf(MEMORY_REPORT_MESSAGE_FORMAT, memory, nodeRequirements.MemoryAmount))
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) isProviderAllowed(provider string) error {
	for _, blockedProvider := range nodeRequirements.BlockedProviders {
		if strings.Contains(provider, blockedProvider) {
			return NewNodeRequirementError(fmt.Errorf(PROVIDER_REPORT_MESSAGE_FORMAT, provider))
		}
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) isKernelVersionAllowed(kernel string) error {
	var err error
	var kernelVersion semver.Version

	if kernelVersion, err = semver.Parse(KERNEL_VERSION_REGEX.FindString(kernel)); err != nil {
		return NewNodeRequirementError(fmt.Errorf(KERNEL_REPORT_MESSAGE_FORMAT, kernel, nodeRequirements.KernelVersion))
	}

	if nodeRequirements.KernelVersion.GT(kernelVersion) {
		return NewNodeRequirementError(fmt.Errorf(KERNEL_REPORT_MESSAGE_FORMAT, kernel, nodeRequirements.KernelVersion))
	}

	return nil
}

func (nodeRequirements *NodeMinimumRequirements) isArchitectureAllowed(architecture string) error {
	for _, allowedArchitecture := range nodeRequirements.AllowedArchitectures {
		if allowedArchitecture == architecture {
			return nil
		}
	}

	return NewNodeRequirementError(fmt.Errorf(ARCHITECTURE_REPORT_MESSAGE_FORMAT, architecture, strings.Join(nodeRequirements.AllowedArchitectures, ", ")))
}

func (nodeRequirements *NodeMinimumRequirements) isOperatingSystemAllowed(operatingSystem string) error {
	for _, allowedOperatingSystem := range nodeRequirements.AllowedOperatingSystems {
		if allowedOperatingSystem == operatingSystem {
			return nil
		}
	}

	return NewNodeRequirementError(fmt.Errorf(OPERATING_SYSTEM_REPORT_MESSAGE_FORMAT, operatingSystem, strings.Join(nodeRequirements.AllowedOperatingSystems, ", ")))
}
