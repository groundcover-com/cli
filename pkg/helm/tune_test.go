package helm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTuneResourcesValuesAgentLow(t *testing.T) {
	// arrange
	agentLowCpu := resource.MustParse(helm.AGENT_MEDIUM_CPU_THRESHOLD)
	agentLowCpu.Sub(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentLowMemory := resource.MustParse(helm.AGENT_MEDIUM_MEMORY_THRESHOLD)
	agentLowMemory.Sub(*resource.NewQuantity(1, resource.BinarySI))

	lowNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentLowCpu,
			Memory: &agentLowMemory,
		},
	}

	resources := helm.CalcAllocatableResources(lowNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.AGENT_LOW_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesAgentMedium(t *testing.T) {
	// arrange
	agentMediumCpu := resource.MustParse(helm.AGENT_MEDIUM_CPU_THRESHOLD)
	agentMediumCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentMediumMemory := resource.MustParse(helm.AGENT_MEDIUM_MEMORY_THRESHOLD)
	agentMediumMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	mediumNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentMediumCpu,
			Memory: &agentMediumMemory,
		},
	}

	resources := helm.CalcAllocatableResources(mediumNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.AGENT_MEDIUM_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesAgentHigh(t *testing.T) {
	// arrange
	agentHighCpu := resource.MustParse(helm.AGENT_HIGH_CPU_THRESHOLD)
	agentHighCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentHighMemory := resource.MustParse(helm.AGENT_HIGH_MEMORY_THRESHOLD)
	agentHighMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	highNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentHighCpu,
			Memory: &agentHighMemory,
		},
	}

	resources := helm.CalcAllocatableResources(highNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.NO_PRESET, cpu)
}

func TestTuneResourcesValuesBackendLow(t *testing.T) {
	// arrange
	backendLowCpu := resource.MustParse(helm.BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	backendLowCpu.Sub(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendLowMemory := resource.MustParse(helm.BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)
	backendLowMemory.Sub(*resource.NewQuantity(1, resource.BinarySI))

	lowNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &backendLowCpu,
			Memory: &backendLowMemory,
		},
	}

	resources := helm.CalcAllocatableResources(lowNodeReport)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.BACKEND_LOW_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesBackendMedium(t *testing.T) {
	// arrange
	backendMediumCpu := resource.MustParse(helm.BACKEND_MEDIUM_TOTAL_CPU_THRESHOLD)
	backendMediumCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendMediumMemory := resource.MustParse(helm.BACKEND_MEDIUM_TOTAL_MEMORY_THRESHOLD)
	backendMediumMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	mediumNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &backendMediumCpu,
			Memory: &backendMediumMemory,
		},
	}

	resources := helm.CalcAllocatableResources(mediumNodeReport)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.BACKEND_MEDIUM_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesBackendHigh(t *testing.T) {
	// arrange
	backendHighCpu := resource.MustParse(helm.BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	backendHighCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendHighMemory := resource.MustParse(helm.BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)
	backendHighMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	highNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &backendHighCpu,
			Memory: &backendHighMemory,
		},
	}

	resources := helm.CalcAllocatableResources(highNodeReport)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.NO_PRESET, cpu)
}

func TestTuneResourcesValuesBackendHuge(t *testing.T) {
	// arrange
	backendHighCpu := resource.MustParse(helm.BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	backendHighCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendHighMemory := resource.MustParse(helm.BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)
	backendHighMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	nodes := []*k8s.NodeSummary{}

	for i := 0; i < 101; i++ {
		nodes = append(nodes, &k8s.NodeSummary{
			CPU:    &backendHighCpu,
			Memory: &backendHighMemory,
		})
	}

	resources := helm.CalcAllocatableResources(nodes)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.BACKEND_HUGE_RESOURCES_PATH, cpu)
}

func TestCalcAllocatableResourcesSingleNode(t *testing.T) {
	// arrange
	nodes := []*k8s.NodeSummary{
		{
			CPU:    resource.NewMilliQuantity(1000, resource.DecimalSI),
			Memory: resource.NewQuantity(1000, resource.BinarySI),
		},
	}

	// act
	resources := helm.CalcAllocatableResources(nodes)

	// assert
	assert.Equal(t, resource.NewMilliQuantity(1000, resource.DecimalSI), resources.MinCpu)
	assert.Equal(t, resource.NewQuantity(1000, resource.BinarySI), resources.MinMemory)
	assert.Equal(t, resource.NewMilliQuantity(1000, resource.DecimalSI), resources.TotalCpu)
	assert.Equal(t, resource.NewQuantity(1000, resource.BinarySI), resources.TotalMemory)
}

func TestCalcAllocatableResourcesMultiNode(t *testing.T) {
	// arrange
	nodes := []*k8s.NodeSummary{
		{
			CPU:    resource.NewMilliQuantity(2000, resource.DecimalSI),
			Memory: resource.NewQuantity(2000, resource.BinarySI),
		},
		{
			CPU:    resource.NewMilliQuantity(1000, resource.DecimalSI),
			Memory: resource.NewQuantity(1000, resource.BinarySI),
		},
	}

	// act
	resources := helm.CalcAllocatableResources(nodes)

	// assert
	assert.Equal(t, resource.NewMilliQuantity(1000, resource.DecimalSI), resources.MinCpu)
	assert.Equal(t, resource.NewQuantity(1000, resource.BinarySI), resources.MinMemory)
	assert.Equal(t, resource.NewMilliQuantity(3000, resource.DecimalSI), resources.TotalCpu)
	assert.Equal(t, resource.NewQuantity(3000, resource.BinarySI), resources.TotalMemory)
}

func TestCalcAllocatableResourcesMultiNodeWithTaints(t *testing.T) {
	// arrange
	nodes := []*k8s.NodeSummary{
		{
			CPU:    resource.NewMilliQuantity(2000, resource.DecimalSI),
			Memory: resource.NewQuantity(2000, resource.BinarySI),
		},
		{
			CPU:    resource.NewMilliQuantity(1000, resource.DecimalSI),
			Memory: resource.NewQuantity(1000, resource.BinarySI),
			Taints: []v1.Taint{
				{
					Key: "key",
				},
			},
		},
	}

	// act
	resources := helm.CalcAllocatableResources(nodes)

	// assert
	assert.Equal(t, resource.NewMilliQuantity(2000, resource.DecimalSI), resources.MinCpu)
	assert.Equal(t, resource.NewQuantity(2000, resource.BinarySI), resources.MinMemory)
	assert.Equal(t, resource.NewMilliQuantity(2000, resource.DecimalSI), resources.TotalCpu)
	assert.Equal(t, resource.NewQuantity(2000, resource.BinarySI), resources.TotalMemory)
}

func TestCalcAllocatableResourcesMultiNodeWithArmArch(t *testing.T) {
	// arrange
	nodes := []*k8s.NodeSummary{
		{
			CPU:    resource.NewMilliQuantity(2000, resource.DecimalSI),
			Memory: resource.NewQuantity(2000, resource.BinarySI),
		},
		{
			Architecture: "arm64",
			CPU:          resource.NewMilliQuantity(1000, resource.DecimalSI),
			Memory:       resource.NewQuantity(1000, resource.BinarySI),
		},
	}

	// act
	resources := helm.CalcAllocatableResources(nodes)

	// assert
	assert.Equal(t, resource.NewMilliQuantity(2000, resource.DecimalSI), resources.MinCpu)
	assert.Equal(t, resource.NewQuantity(2000, resource.BinarySI), resources.MinMemory)
	assert.Equal(t, resource.NewMilliQuantity(2000, resource.DecimalSI), resources.TotalCpu)
	assert.Equal(t, resource.NewQuantity(2000, resource.BinarySI), resources.TotalMemory)
}
