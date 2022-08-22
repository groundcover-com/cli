package helm

import (
	"embed"

	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	AGENT_CPU_THRESHOLD            = 20
	AGENT_MEMORY_THRESHOLD         = 20
	BACKEND_TOTAL_CPU_THRESHOLD    = 80
	BACKEND_TOTAL_MEMORY_THRESHOLD = 80
	AGENT_LOW_RESOURCES_PATH       = "presets/agent/low-resources.yaml"
	BACKEND_LOW_RESOURCES_PATH     = "presets/backend/low-resources.yaml"
)

//go:embed presets/*
var presets embed.FS

type AllocatableResources struct {
	MinCpu      int64
	MinMemory   int64
	TotalCpu    int64
	TotalMemory int64
}

func TuneResourcesValues(chartValues *map[string]interface{}, nodeReports []*k8s.NodeReport) error {
	var err error

	allocatableResources := calcAllocatableResources(nodeReports)

	if allocatableResources.MinCpu <= AGENT_CPU_THRESHOLD || allocatableResources.MinMemory <= AGENT_MEMORY_THRESHOLD {
		var data []byte
		if data, err = presets.ReadFile(AGENT_LOW_RESOURCES_PATH); err != nil {
			return err
		}

		if err = yaml.Unmarshal(data, chartValues); err != nil {
			return err
		}
	}

	if allocatableResources.TotalCpu <= BACKEND_TOTAL_CPU_THRESHOLD || allocatableResources.TotalMemory <= BACKEND_TOTAL_MEMORY_THRESHOLD {
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
		MinCpu:    nodeReports[0].CPU,
		MinMemory: nodeReports[0].Memory,
	}

	for _, nodeReport := range nodeReports {
		allocatableResources.TotalCpu += nodeReport.CPU
		allocatableResources.TotalMemory += nodeReport.Memory

		if nodeReport.CPU < allocatableResources.MinCpu {
			allocatableResources.MinCpu = nodeReport.CPU
		}

		if nodeReport.Memory < allocatableResources.MinMemory {
			allocatableResources.MinMemory = nodeReport.Memory
		}
	}

	return allocatableResources
}
