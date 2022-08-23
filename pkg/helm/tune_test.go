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

	cpuNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(500, resource.Milli),
				Memory: resource.NewScaledQuantity(1750, resource.Mega),
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(500, resource.Milli),
				Memory: resource.NewScaledQuantity(1750, resource.Mega),
			},
		},
	}

	memoryNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(1250, resource.Mega),
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(5000, resource.Mega),
			},
		},
	}

	//act
	err = helm.TuneResourcesValues(&cpuChartValues, cpuNodeReports)
	suite.NoError(err)

	err = helm.TuneResourcesValues(&memoryChartValues, memoryNodeReports)
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

	cpuChartValues := make(map[string]interface{})
	memoryChartValues := make(map[string]interface{})

	cpuNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(1750, resource.Mega),
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(6000, resource.Mega),
			},
		},
	}

	memoryNodeReports := []*k8s.NodeReport{
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(1750, resource.Mega),
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				CPU:    resource.NewScaledQuantity(1250, resource.Milli),
				Memory: resource.NewScaledQuantity(6000, resource.Mega),
			},
		},
	}

	//act
	err = helm.TuneResourcesValues(&cpuChartValues, cpuNodeReports)
	suite.NoError(err)

	err = helm.TuneResourcesValues(&memoryChartValues, memoryNodeReports)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})

	suite.Equal(expected, cpuChartValues)
	suite.Equal(expected, memoryChartValues)
}
