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

func (suite *HelmTuneTestSuite) TestTuneResourcesValuesSuccess() {
	//prepare
	var err error

	cpuChartValues := make(map[string]interface{})
	memoryChartValues := make(map[string]interface{})

	lowerThenThresholdCpu := resource.MustParse("750m")
	higherThenThresholdCpu := resource.MustParse("1250m")
	lowerThenThresholdMemory := resource.MustParse("1500Mi")
	higherThenThresholdMemory := resource.MustParse("3000Mi")

	lowerCpuNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &lowerThenThresholdCpu,
				Memory: &higherThenThresholdMemory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &lowerThenThresholdCpu,
				Memory: &higherThenThresholdMemory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &lowerThenThresholdCpu,
				Memory: &higherThenThresholdMemory,
			},
		},
	}

	lowerMemoryNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &higherThenThresholdCpu,
				Memory: &lowerThenThresholdMemory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &higherThenThresholdCpu,
				Memory: &lowerThenThresholdMemory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &higherThenThresholdCpu,
				Memory: &lowerThenThresholdMemory,
			},
		},
	}

	//act
	err = helm.TuneResourcesValues(&cpuChartValues, lowerCpuNodeReports)
	suite.NoError(err)

	err = helm.TuneResourcesValues(&memoryChartValues, lowerMemoryNodeReports)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})

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
	suite.Equal(expected, memoryChartValues)
}

func (suite *HelmTuneTestSuite) TestTuneResourcesValuesEmpty() {
	//prepare
	var err error

	cpu := resource.MustParse("1250m")
	memory := resource.MustParse("3000Mi")
	chartValues := make(map[string]interface{})

	nodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &cpu,
				Memory: &memory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &cpu,
				Memory: &memory,
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    &cpu,
				Memory: &memory,
			},
		},
	}

	//act

	err = helm.TuneResourcesValues(&chartValues, nodeReports)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})

	suite.Equal(expected, chartValues)
}
