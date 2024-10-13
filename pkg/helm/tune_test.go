package helm_test

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	OldKernelSemver = semver.MustParse("5.10.0")
	NewKernelSemver = semver.MustParse("5.11.0")
)

func TestTuneResourcesValuesAgentLow(t *testing.T) {
	// arrange
	agentLowCpu := resource.MustParse(helm.AGENT_DEFAULT_CPU_THRESHOLD)
	agentLowCpu.Sub(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentLowMemory := resource.MustParse(helm.AGENT_DEFAULT_MEMORY_THRESHOLD)
	agentLowMemory.Sub(*resource.NewQuantity(1, resource.BinarySI))

	lowNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentLowCpu,
			Memory: &agentLowMemory,
		},
	}

	resources := helm.CalcAllocatableResources(lowNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources, OldKernelSemver)

	// assert
	assert.Equal(t, helm.AGENT_LOW_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesAgentDefault(t *testing.T) {
	// arrange
	agentDefaultCpu := resource.MustParse(helm.AGENT_DEFAULT_CPU_THRESHOLD)
	agentDefaultCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentDefaultMemory := resource.MustParse(helm.AGENT_DEFAULT_MEMORY_THRESHOLD)
	agentDefaultMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	defaultNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentDefaultCpu,
			Memory: &agentDefaultMemory,
		},
	}

	resources := helm.CalcAllocatableResources(defaultNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources, OldKernelSemver)

	// assert
	assert.Equal(t, helm.DEFAULT_PRESET, cpu)
}

func TestTuneResourcesValuesAgentNewKernel(t *testing.T) {
	// arrange
	agentDefaultCpu := resource.MustParse(helm.AGENT_DEFAULT_CPU_THRESHOLD)
	agentDefaultCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	agentDefaultMemory := resource.MustParse(helm.AGENT_DEFAULT_MEMORY_THRESHOLD)
	agentDefaultMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	defaultNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &agentDefaultCpu,
			Memory: &agentDefaultMemory,
		},
	}

	resources := helm.CalcAllocatableResources(defaultNodeReport)

	// act
	cpu := helm.GetAgentResourcePresetPath(resources, NewKernelSemver)

	// assert
	assert.Equal(t, helm.AGENT_KERNEL_5_11_PRESET_PATH, cpu)
}

func TestTuneResourcesValuesBackendLow(t *testing.T) {
	// arrange
	backendLowCpu := resource.MustParse(helm.BACKEND_DEFAULT_TOTAL_CPU_THRESHOLD)
	backendLowCpu.Sub(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendLowMemory := resource.MustParse(helm.BACKEND_DEFAULT_TOTAL_MEMORY_THRESHOLD)
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

func TestTuneResourcesValuesBackendDefault(t *testing.T) {
	// arrange
	backendDefaultCpu := resource.MustParse(helm.BACKEND_DEFAULT_TOTAL_CPU_THRESHOLD)
	backendDefaultCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendDefaultMemory := resource.MustParse(helm.BACKEND_DEFAULT_TOTAL_MEMORY_THRESHOLD)
	backendDefaultMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	defaultNodeReport := []*k8s.NodeSummary{
		{
			CPU:    &backendDefaultCpu,
			Memory: &backendDefaultMemory,
		},
	}

	resources := helm.CalcAllocatableResources(defaultNodeReport)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.DEFAULT_PRESET, cpu)
}

func TestTuneResourcesValuesBackendHigh(t *testing.T) {
	// arrange
	backendHighCpu := resource.MustParse(helm.BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	backendHighCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendHighMemory := resource.MustParse(helm.BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)
	backendHighMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	nodes := []*k8s.NodeSummary{}

	for i := 0; i < helm.HUGE_RESOURCES_CLUSTER_NODE_COUNT-1; i++ {
		nodes = append(nodes, &k8s.NodeSummary{
			CPU:    &backendHighCpu,
			Memory: &backendHighMemory,
		})
	}

	resources := helm.CalcAllocatableResources(nodes)

	// act
	cpu := helm.GetBackendResourcePresetPath(resources)

	// assert
	assert.Equal(t, helm.BACKEND_HIGH_RESOURCES_PATH, cpu)
}

func TestTuneResourcesValuesBackendHuge(t *testing.T) {
	// arrange
	backendHighCpu := resource.MustParse(helm.BACKEND_HIGH_TOTAL_CPU_THRESHOLD)
	backendHighCpu.Add(*resource.NewMilliQuantity(1, resource.DecimalSI))

	backendHighMemory := resource.MustParse(helm.BACKEND_HIGH_TOTAL_MEMORY_THRESHOLD)
	backendHighMemory.Add(*resource.NewQuantity(1, resource.BinarySI))

	nodes := []*k8s.NodeSummary{}

	for i := 0; i <= helm.HUGE_RESOURCES_CLUSTER_NODE_COUNT; i++ {
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
