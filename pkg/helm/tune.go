package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	MAX_USAGE_RATIO = 15.0

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
var presets embed.FS

type AllocatableResources struct {
	MinCpu      *resource.Quantity
	MinMemory   *resource.Quantity
	TotalCpu    *resource.Quantity
	TotalMemory *resource.Quantity
}

func TuneResourcesValues(chartValues *map[string]interface{}, nodeReports []*k8s.NodeReport) ([]string, error) {
	var err error

	presetPaths := make([]string, 2)
	allocatableResources := calcAllocatableResources(nodeReports)

	if presetPaths[0], err = tuneAgentResourcesValues(chartValues, allocatableResources); err != nil {
		return nil, err
	}

	if presetPaths[1], err = tuneBackendResourcesValues(chartValues, allocatableResources); err != nil {
		return nil, err
	}

	return presetPaths, nil
}

func tuneAgentResourcesValues(chartValues *map[string]interface{}, allocatableResources *AllocatableResources) (string, error) {
	var err error

	mediumCpuThreshold := resource.MustParse(AGENT_MEDIUM_CPU_THRESHOLD)
	highCpuThreshold := resource.MustParse(AGENT_HIGH_CPU_THRESHOLD)

	mediumMemoryThreshold := resource.MustParse(AGENT_MEDIUM_MEMORY_THRESHOLD)
	highMemoryThreshold := resource.MustParse(AGENT_HIGH_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.MinCpu.AsApproximateFloat64() * MAX_USAGE_RATIO / 100
	maxMemoryUsage := allocatableResources.MinMemory.AsApproximateFloat64() * MAX_USAGE_RATIO / 100

	var presetPath string
	switch {
	case maxCpuUsage >= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= highMemoryThreshold.AsApproximateFloat64():
		return "", nil
	case maxCpuUsage >= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = AGENT_MEDIUM_RESOURCES_PATH
	default:
		presetPath = AGENT_LOW_RESOURCES_PATH
	}

	var data []byte
	if data, err = presets.ReadFile(presetPath); err != nil {
		return "", err
	}

	if err = yaml.Unmarshal(data, chartValues); err != nil {
		return "", err
	}

	return presetPath, nil
}

func tuneBackendResourcesValues(chartValues *map[string]interface{}, allocatableResources *AllocatableResources) (string, error) {
	var err error

	mediumCpuThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	highCpuThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_CPU_THRESHOLD)

	mediumMemoryThreshold := resource.MustParse(BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)
	highMemoryThreshold := resource.MustParse(BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)

	maxCpuUsage := allocatableResources.TotalCpu.AsApproximateFloat64() * MAX_USAGE_RATIO / 100
	maxMemoryUsage := allocatableResources.TotalMemory.AsApproximateFloat64() * MAX_USAGE_RATIO / 100

	var presetPath string
	switch {
	case maxCpuUsage >= highCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= highMemoryThreshold.AsApproximateFloat64():
		return "", nil
	case maxCpuUsage >= mediumCpuThreshold.AsApproximateFloat64(), maxMemoryUsage >= mediumMemoryThreshold.AsApproximateFloat64():
		presetPath = BACKEND_MEDIUM_RESOURCES_PATH
	default:
		presetPath = BACKEND_LOW_RESOURCES_PATH
	}

	var data []byte
	if data, err = presets.ReadFile(presetPath); err != nil {
		return "", err
	}

	if err = yaml.Unmarshal(data, chartValues); err != nil {
		return "", err
	}

	return presetPath, nil
}

func calcAllocatableResources(nodeReports []*k8s.NodeReport) *AllocatableResources {
	allocatableResources := &AllocatableResources{
		MinCpu:      nodeReports[0].CPU,
		MinMemory:   nodeReports[0].Memory,
		TotalCpu:    &resource.Quantity{},
		TotalMemory: &resource.Quantity{},
	}

	for _, nodeReport := range nodeReports {
		allocatableResources.TotalCpu.Add(*nodeReport.CPU)
		allocatableResources.TotalMemory.Add(*nodeReport.Memory)

		if allocatableResources.MinCpu.Cmp(*nodeReport.CPU) > 0 {
			allocatableResources.MinCpu = nodeReport.CPU
		}

		if allocatableResources.MinMemory.Cmp(*nodeReport.Memory) > 0 {
			allocatableResources.MinMemory = nodeReport.Memory
		}
	}

	return allocatableResources
}