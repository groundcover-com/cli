package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	MAX_USAGE_RATIO = 15.0 / 100

	AGENT_MEDIUM_CPU_THRESHOLD    = "1000m"
	AGENT_HIGH_CPU_THRESHOLD      = "1500m"
	AGENT_MEDIUM_MEMORY_THRESHOLD = "2500Mi"
	AGENT_HIGH_MEMORY_THRESHOLD   = "3000Mi"
	AGENT_LOW_RESOURCES_PATH      = "presets/agent/medium-resources.yaml"
	AGENT_MEDIUM_RESOURCES_PATH   = "presets/agent/medium-resources.yaml"

	BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD    = "3000m"
	BACKEND_HIGH_TOTAL_CPU_THRESHOLD      = "4000m"
	BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD = "6000Mi"
	BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD   = "9000Mi"
	BACKEND_LOW_RESOURCES_PATH            = "presets/backend/medium-resources.yaml"
	BACKEND_MEDIUM_RESOURCES_PATH         = "presets/backend/medium-resources.yaml"
)

//go:embed presets/*
var presetsFS embed.FS

type AllocatableResources struct {
	MinCpu      *resource.Quantity
	MinMemory   *resource.Quantity
	TotalCpu    *resource.Quantity
	TotalMemory *resource.Quantity
}

func GetResourcesTunerPresetPaths(nodesSummeries []*k8s.NodeSummary) ([]string, error) {
	var err error

	presetPaths := make([]string, 2)
	allocatableResources := calcAllocatableResources(nodesSummeries)

	if presetPaths[0], err = tuneAgentResourcesValues(allocatableResources); err != nil {
		return nil, err
	}

	if presetPaths[1], err = tuneBackendResourcesValues(allocatableResources); err != nil {
		return nil, err
	}

	return presetPaths, nil
}

func tuneAgentResourcesValues(allocatableResources *AllocatableResources) (string, error) {
	mediumCpuThreshold := resource.MustParse(AGENT_MEDIUM_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(AGENT_MEDIUM_MEMORY_THRESHOLD)
	highCpuThreshold := resource.MustParse(AGENT_HIGH_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(AGENT_HIGH_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.MinCpu.AsApproximateFloat64() * MAX_USAGE_RATIO
	maxMemoryUsage := allocatableResources.MinMemory.AsApproximateFloat64() * MAX_USAGE_RATIO

	var presetPath string
	switch {
	case maxCpuUsage >= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= highMemoryThreshold.AsApproximateFloat64():
		return "", nil
	case maxCpuUsage >= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_MEDIUM_RESOURCES_PATH
	default:
		presetPath = AGENT_LOW_RESOURCES_PATH
	}

	return presetPath, nil
}

func tuneBackendResourcesValues(allocatableResources *AllocatableResources) (string, error) {
	mediumCpuThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)
	highCpuThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.TotalCpu.AsApproximateFloat64() * MAX_USAGE_RATIO
	maxMemoryUsage := allocatableResources.TotalMemory.AsApproximateFloat64() * MAX_USAGE_RATIO

	var presetPath string
	switch {
	case maxCpuUsage >= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= highMemoryThreshold.AsApproximateFloat64():
		return "", nil
	case maxCpuUsage >= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_MEDIUM_RESOURCES_PATH
	default:
		presetPath = BACKEND_LOW_RESOURCES_PATH
	}

	return presetPath, nil
}

func calcAllocatableResources(nodesSummeries []*k8s.NodeSummary) *AllocatableResources {
	allocatableResources := &AllocatableResources{
		MinCpu:      nodesSummeries[0].CPU,
		MinMemory:   nodesSummeries[0].Memory,
		TotalCpu:    &resource.Quantity{},
		TotalMemory: &resource.Quantity{},
	}

	for _, nodeSummary := range nodesSummeries {
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
