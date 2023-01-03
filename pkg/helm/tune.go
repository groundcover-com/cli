package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	AGENT_MEDIUM_CPU_THRESHOLD    = "1500m"
	AGENT_HIGH_CPU_THRESHOLD      = "3000m"
	AGENT_MEDIUM_MEMORY_THRESHOLD = "1024Mi"
	AGENT_HIGH_MEMORY_THRESHOLD   = "3072Mi"
	AGENT_LOW_RESOURCES_PATH      = "presets/agent/low-resources.yaml"
	AGENT_MEDIUM_RESOURCES_PATH   = "presets/agent/medium-resources.yaml"

	BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD    = "4000m"
	BACKEND_HIGH_TOTAL_CPU_THRESHOLD      = "8000m"
	BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD = "3072Mi"
	BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD   = "6144Mi"
	BACKEND_LOW_RESOURCES_PATH            = "presets/backend/low-resources.yaml"
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

	allocatableResources := calcAllocatableResources(nodesSummeries)

	tuneFuncs := []func(*AllocatableResources) (string, error){
		tuneAgentResourcesValues,
		tuneBackendResourcesValues,
	}

	var presetPath string
	var presetPaths []string
	for _, tuneFunc := range tuneFuncs {
		if presetPath, err = tuneFunc(allocatableResources); err != nil {
			return nil, err
		}

		if presetPath == "" {
			continue
		}

		presetPaths = append(presetPaths, presetPath)
	}

	return presetPaths, nil
}

func tuneAgentResourcesValues(allocatableResources *AllocatableResources) (string, error) {
	mediumCpuThreshold := resource.MustParse(AGENT_MEDIUM_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(AGENT_MEDIUM_MEMORY_THRESHOLD)
	highCpuThreshold := resource.MustParse(AGENT_HIGH_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(AGENT_HIGH_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.MinCpu.AsApproximateFloat64()
	maxMemoryUsage := allocatableResources.MinMemory.AsApproximateFloat64()

	var presetPath string
	switch {
	case maxCpuUsage <= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage <= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_LOW_RESOURCES_PATH
	case maxCpuUsage <= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage <= highMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_MEDIUM_RESOURCES_PATH
	default:
		return "", nil
	}

	return presetPath, nil
}

func tuneBackendResourcesValues(allocatableResources *AllocatableResources) (string, error) {
	mediumCpuThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	mediumMemoryThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)
	highCpuThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	highMemoryThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.TotalCpu.AsApproximateFloat64()
	maxMemoryUsage := allocatableResources.TotalMemory.AsApproximateFloat64()

	var presetPath string
	switch {
	case maxCpuUsage <= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage <= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_LOW_RESOURCES_PATH
	case maxCpuUsage <= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage <= highMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_MEDIUM_RESOURCES_PATH
	default:
		return "", nil
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
