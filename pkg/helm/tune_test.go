package helm_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type HelmTuneTestSuite struct {
	suite.Suite
}

func (suite *HelmTuneTestSuite) SetupSuite() {}

func (suite *HelmTuneTestSuite) TearDownSuite() {}

func TestHelmTuneTestSuite(t *testing.T) {
	suite.Run(t, &HelmTuneTestSuite{})
}

func (suite *HelmTuneTestSuite) TestTuneResourcesValuesLowSuccess() {
	//prepare
	var err error
	var cpuChartValues map[string]interface{}
	var memoryChartValues map[string]interface{}

	lowerThenThresholdCpu := resource.MustParse("2000m")
	higherThenThresholdCpu := resource.MustParse("6000m")
	lowerThenThresholdMemory := resource.MustParse("4000Mi")
	higherThenThresholdMemory := resource.MustParse("3000Mi")

	lowerCpuNodeReports := []*k8s.NodeSummary{
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
	}

	lowerMemoryNodeReports := []*k8s.NodeSummary{
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
	}

	//act
	var cpuPresetPaths []string
	cpuPresetPaths, err = helm.GetResourcesTunerPresetPaths(lowerCpuNodeReports)
	suite.NoError(err)

	cpuChartValues, err = helm.GetChartValuesOverrides(cpuPresetPaths)
	suite.NoError(err)

	var memoryPresetPaths []string
	memoryPresetPaths, err = helm.GetResourcesTunerPresetPaths(lowerMemoryNodeReports)
	suite.NoError(err)

	memoryChartValues, err = helm.GetChartValuesOverrides(memoryPresetPaths)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})
	expectedPresetPaths := []string{helm.AGENT_LOW_RESOURCES_PATH, helm.BACKEND_LOW_RESOURCES_PATH}

	var data []byte
	data, err = os.ReadFile(helm.AGENT_LOW_RESOURCES_PATH)
	suite.NoError(err)

	err = yaml.Unmarshal(data, &expected)
	suite.NoError(err)

	data, err = os.ReadFile(helm.BACKEND_LOW_RESOURCES_PATH)
	suite.NoError(err)

	err = yaml.Unmarshal(data, &expected)
	suite.NoError(err)

	suite.Equal(expected, cpuChartValues)
	suite.Equal(expectedPresetPaths, cpuPresetPaths)
	suite.Equal(expected, memoryChartValues)
	suite.Equal(expectedPresetPaths, memoryPresetPaths)
}

func (suite *HelmTuneTestSuite) TestTuneResourcesValuesMediumSuccess() {
	//prepare
	var err error
	var cpuChartValues map[string]interface{}
	var memoryChartValues map[string]interface{}

	lowerThenThresholdCpu := resource.MustParse("8000m")
	higherThenThresholdCpu := resource.MustParse("6000m")
	lowerThenThresholdMemory := resource.MustParse("18000Mi")
	higherThenThresholdMemory := resource.MustParse("3000Mi")

	lowerCpuNodeReports := []*k8s.NodeSummary{
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
		{
			CPU:    &lowerThenThresholdCpu,
			Memory: &higherThenThresholdMemory,
		},
	}

	lowerMemoryNodeReports := []*k8s.NodeSummary{
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
		{
			CPU:    &higherThenThresholdCpu,
			Memory: &lowerThenThresholdMemory,
		},
	}

	//act
	var cpuPresetPaths []string
	cpuPresetPaths, err = helm.GetResourcesTunerPresetPaths(lowerCpuNodeReports)
	suite.NoError(err)

	cpuChartValues, err = helm.GetChartValuesOverrides(cpuPresetPaths)
	suite.NoError(err)

	var memoryPresetPaths []string
	memoryPresetPaths, err = helm.GetResourcesTunerPresetPaths(lowerMemoryNodeReports)
	suite.NoError(err)

	memoryChartValues, err = helm.GetChartValuesOverrides(memoryPresetPaths)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})
	expectedPresetPaths := []string{helm.AGENT_MEDIUM_RESOURCES_PATH, helm.BACKEND_MEDIUM_RESOURCES_PATH}

	var data []byte
	data, err = os.ReadFile(helm.AGENT_MEDIUM_RESOURCES_PATH)
	suite.NoError(err)

	err = yaml.Unmarshal(data, &expected)
	suite.NoError(err)

	data, err = os.ReadFile(helm.BACKEND_MEDIUM_RESOURCES_PATH)
	suite.NoError(err)

	err = yaml.Unmarshal(data, &expected)
	suite.NoError(err)

	suite.Equal(expected, cpuChartValues)
	suite.Equal(expectedPresetPaths, cpuPresetPaths)
	suite.Equal(expected, memoryChartValues)
	suite.Equal(expectedPresetPaths, memoryPresetPaths)
}

func (suite *HelmTuneTestSuite) TestTuneResourcesValuesHighSuccess() {
	//prepare
	var err error
	var chartValues map[string]interface{}

	cpu := resource.MustParse("12000m")
	memory := resource.MustParse("36000Mi")

	nodeReports := []*k8s.NodeSummary{
		{
			CPU:    &cpu,
			Memory: &memory,
		},
		{
			CPU:    &cpu,
			Memory: &memory,
		},
		{
			CPU:    &cpu,
			Memory: &memory,
		},
	}

	//act

	var presetPaths []string
	presetPaths, err = helm.GetResourcesTunerPresetPaths(nodeReports)
	suite.NoError(err)

	chartValues, err = helm.GetChartValuesOverrides(presetPaths)
	suite.NoError(err)
	// assert

	expected := make(map[string]interface{})
	var expectedPresetPaths []string

	suite.Equal(expected, chartValues)
	suite.Equal(expectedPresetPaths, presetPaths)
}
