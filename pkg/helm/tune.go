package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	NO_PRESET = ""

	HIGH_RESOURCES_CLUSTER_NODE_COUNT = 30
	HUGE_RESOURCES_CLUSTER_NODE_COUNT = 100

	AGENT_MEDIUM_CPU_THRESHOLD    = "1000m"
	AGENT_MEDIUM_MEMORY_THRESHOLD = "1024Mi"
	AGENT_HIGH_CPU_THRESHOLD      = "3000m"
	AGENT_HIGH_MEMORY_THRESHOLD   = "3072Mi"
	AGENT_LOW_RESOURCES_PATH      = "presets/agent/low-resources.yaml"
	AGENT_MEDIUM_RESOURCES_PATH   = "presets/agent/medium-resources.yaml"

	EMPTYDIR_STORAGE_PATH = "presets/backend/emptydir-storage.yaml"

	BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD    = "12000m"
	BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD = "20000Mi"
	BACKEND_HIGH_TOTAL_CPU_THRESHOLD      = "30000m"
	BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD   = "60000Mi"
	BACKEND_LOW_RESOURCES_PATH            = "presets/backend/low-resources.yaml"
	BACKEND_MEDIUM_RESOURCES_PATH         = "presets/backend/medium-resources.yaml"
	BACKEND_HIGH_RESOURCES_PATH           = "presets/backend/high-resources.yaml"
	BACKEND_HUGE_RESOURCES_PATH           = "presets/backend/huge-resources.yaml"
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

func GetAgentResourcePresetPath(allocatableResources *AllocatableResources) string {
	mediumCpuThreshold := resource.MustParse(AGENT_MEDIUM_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(AGENT_MEDIUM_MEMORY_THRESHOLD)
	highCpuThreshold := resource.MustParse(AGENT_HIGH_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(AGENT_HIGH_MEMORY_THRESHOLD)

	minAllocatableCpu := allocatableResources.MinCpu.AsApproximateFloat64()
	minAllocatableMemory := allocatableResources.MinMemory.AsApproximateFloat64()

	var presetPath string
	switch {
	case minAllocatableCpu <= mediumCpuThreshold.AsApproximateFloat64(), minAllocatableMemory <= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_LOW_RESOURCES_PATH
	case minAllocatableCpu <= highCpuThreshold.AsApproximateFloat64(), minAllocatableMemory <= highMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_MEDIUM_RESOURCES_PATH
	default:
		return NO_PRESET
	}

	return presetPath
}

func GetBackendResourcePresetPath(allocatableResources *AllocatableResources) string {
	mediumCpuThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)

	highCpuThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)

	totalAllocatableCpu := allocatableResources.TotalCpu.AsApproximateFloat64()
	totalAllocatableMemory := allocatableResources.TotalMemory.AsApproximateFloat64()

	var presetPath string
	switch {
	case totalAllocatableCpu <= mediumCpuThreshold.AsApproximateFloat64(), totalAllocatableMemory <= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_LOW_RESOURCES_PATH
	case totalAllocatableCpu <= highCpuThreshold.AsApproximateFloat64(), totalAllocatableMemory <= highMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_MEDIUM_RESOURCES_PATH
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
