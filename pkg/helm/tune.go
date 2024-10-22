package helm

import (
	"embed"

	"github.com/blang/semver/v4"
	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	DEFAULT_PRESET = ""

	HIGH_RESOURCES_CLUSTER_NODE_COUNT = 30
	HUGE_RESOURCES_CLUSTER_NODE_COUNT = 100

	AGENT_DEFAULT_CPU_THRESHOLD    = "1000m"
	AGENT_DEFAULT_MEMORY_THRESHOLD = "1024Mi"
	AGENT_LOW_RESOURCES_PATH       = "presets/agent/low-resources.yaml"

	// Starting from Linux kernel version 5.11, eBPF maps are accounted for in the memory cgroup
	// of the process that created them. For this reason we need to increase the memory limit for
	// the agent.
	// https://github.com/cilium/ebpf/blob/v0.16.0/docs/ebpf/concepts/rlimit.md#resource-limits
	AGENT_KERNEL_5_11_PRESET_PATH = "presets/agent/kernel-5-11.yaml"
	KERNEL_5_11_SEMVER_EXPRESSION = ">=5.11.0"

	EMPTYDIR_STORAGE_PATH = "presets/backend/emptydir-storage.yaml"

	BACKEND_DEFAULT_TOTAL_CPU_THRESHOLD    = "12000m"
	BACKEND_DEFAULT_TOTAL_MEMORY_THRESHOLD = "20000Mi"
	BACKEND_HIGH_TOTAL_CPU_THRESHOLD       = "30000m"
	BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD    = "60000Mi"
	BACKEND_LOW_RESOURCES_PATH             = "presets/backend/low-resources.yaml"
	BACKEND_HIGH_RESOURCES_PATH            = "presets/backend/high-resources.yaml"
	BACKEND_HUGE_RESOURCES_PATH            = "presets/backend/huge-resources.yaml"
)

//go:embed presets/*
var presetsFS embed.FS

type AllocatableResources struct {
	MinCpu      *resource.Quantity
	MinMemory   *resource.Quantity
	TotalCpu    *resource.Quantity
	TotalMemory *resource.Quantity
	NodeCount   int
}

func GetAgentResourcePresetPath(allocatableResources *AllocatableResources, maxKernelVersion semver.Version) string {
	defaultCpuThreshold := resource.MustParse(AGENT_DEFAULT_CPU_THRESHOLD)
	defaultMemoryThreshold := resource.MustParse(AGENT_DEFAULT_MEMORY_THRESHOLD)

	minAllocatableCpu := allocatableResources.MinCpu.AsApproximateFloat64()
	minAllocatableMemory := allocatableResources.MinMemory.AsApproximateFloat64()

	if minAllocatableCpu <= defaultCpuThreshold.AsApproximateFloat64() || minAllocatableMemory <= defaultMemoryThreshold.AsApproximateFloat64() {
		return AGENT_LOW_RESOURCES_PATH
	}

	if semver.MustParseRange(KERNEL_5_11_SEMVER_EXPRESSION)(maxKernelVersion) {
		return AGENT_KERNEL_5_11_PRESET_PATH
	}

	return DEFAULT_PRESET
}

func GetBackendResourcePresetPath(allocatableResources *AllocatableResources) string {
	defaultCpuThreshold := resource.MustParse(BACKEND_DEFAULT_TOTAL_CPU_THRESHOLD)
	defaultMemoryThreshold := resource.MustParse(BACKEND_DEFAULT_TOTAL_MEMORY_THRESHOLD)

	highCpuThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)

	totalAllocatableCpu := allocatableResources.TotalCpu.AsApproximateFloat64()
	totalAllocatableMemory := allocatableResources.TotalMemory.AsApproximateFloat64()

	var presetPath string
	switch {
	case totalAllocatableCpu <= defaultCpuThreshold.AsApproximateFloat64(), totalAllocatableMemory <= defaultMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_LOW_RESOURCES_PATH
	case totalAllocatableCpu <= highCpuThreshold.AsApproximateFloat64(), totalAllocatableMemory <= highMemoryThreshold.AsApproximateFloat64():
		presetPath = DEFAULT_PRESET
	case allocatableResources.NodeCount < HUGE_RESOURCES_CLUSTER_NODE_COUNT:
		presetPath = BACKEND_HIGH_RESOURCES_PATH
	default:
		return BACKEND_HUGE_RESOURCES_PATH
	}

	return presetPath
}

func CalcAllocatableResources(nodesSummaries []*k8s.NodeSummary) *AllocatableResources {
	allocatableResources := &AllocatableResources{
		MinCpu:      nodesSummaries[0].CPU,
		MinMemory:   nodesSummaries[0].Memory,
		TotalCpu:    &resource.Quantity{},
		TotalMemory: &resource.Quantity{},
		NodeCount:   len(nodesSummaries),
	}

	for _, nodeSummary := range nodesSummaries {
		if len(nodeSummary.Taints) > 0 || nodeSummary.IsArm64() {
			continue
		}

		allocatableResources.TotalCpu.Add(*nodeSummary.CPU)
		allocatableResources.TotalMemory.Add(*nodeSummary.Memory)

		if allocatableResources.MinCpu.Cmp(*nodeSummary.CPU) > 0 {
			allocatableResources.MinCpu = nodeSummary.CPU
		}

		if allocatableResources.MinMemory.Cmp(*nodeSummary.Memory) > 0 {
			allocatableResources.MinMemory = nodeSummary.Memory
		}
	}

	return allocatableResources
}
