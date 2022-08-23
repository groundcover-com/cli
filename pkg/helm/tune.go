package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	AGENT_CPU_THRESHOLD            = "1000m"
	AGENT_MEMORY_THRESHOLD         = "2700Mi"
	BACKEND_TOTAL_CPU_THRESHOLD    = "3000m"
	BACKEND_TOTAL_MEMORY_THRESHOLD = "9000Mi"
	AGENT_LOW_RESOURCES_PATH       = "presets/agent/low-resources.yaml"
	BACKEND_LOW_RESOURCES_PATH     = "presets/backend/low-resources.yaml"
)

//go:embed presets/*
var presets embed.FS

type AllocatableResources struct {
	MinCpu      *resource.Quantity
	MinMemory   *resource.Quantity
	TotalCpu    *resource.Quantity
	TotalMemory *resource.Quantity
}

func TuneResourcesValues(chartValues *map[string]interface{}, nodeReports []*k8s.NodeReport) error {
	var err error

	allocatableResources := calcAllocatableResources(nodeReports)

	agentCpuThreshold := resource.MustParse(AGENT_CPU_THRESHOLD)
	agentMemoryThreshold := resource.MustParse(AGENT_MEMORY_THRESHOLD)
	if allocatableResources.MinCpu.Cmp(agentCpuThreshold) < 0 || allocatableResources.MinMemory.Cmp(agentMemoryThreshold) < 0 {
		var data []byte
		if data, err = presets.ReadFile(AGENT_LOW_RESOURCES_PATH); err != nil {
			return err
		}

		if err = yaml.Unmarshal(data, chartValues); err != nil {
			return err
		}
	}

	backendTotalCpuThreshold := resource.MustParse(BACKEND_TOTAL_CPU_THRESHOLD)
	backendTotalMemoryThreshold := resource.MustParse(BACKEND_TOTAL_MEMORY_THRESHOLD)
	if allocatableResources.TotalCpu.Cmp(backendTotalCpuThreshold) < 0 || allocatableResources.TotalMemory.Cmp(backendTotalMemoryThreshold) < 0 {
		var data []byte
		if data, err = presets.ReadFile(BACKEND_LOW_RESOURCES_PATH); err != nil {
			return err
		}

		if err = yaml.Unmarshal(data, chartValues); err != nil {
			return err
		}
	}

	return nil
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
