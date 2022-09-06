package k8s_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const DEFAULT_CONTEXT_TIMEOUT = time.Duration(time.Minute * 1)

type KubeNodeTestSuite struct {
	suite.Suite
	KubeClient k8s.Client
}

func (suite *KubeNodeTestSuite) SetupSuite() {
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "compatible",
				},
				Spec: v1.NodeSpec{
					ProviderID: "aws://eu-west-3/i-53df4efedd",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewScaledQuantity(2000, resource.Milli),
						v1.ResourceMemory: *resource.NewScaledQuantity(4000, resource.Mega),
					},
					NodeInfo: v1.NodeSystemInfo{
						Architecture:    "amd64",
						OperatingSystem: "linux",
						KernelVersion:   "4.14.0",
						OSImage:         "amazon linux",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "incompatible",
				},
				Spec: v1.NodeSpec{
					ProviderID: "aws://eu-west-3/fargate-i-53df4efedd",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewScaledQuantity(500, resource.Milli),
						v1.ResourceMemory: *resource.NewScaledQuantity(1000, resource.Mega),
					},
					NodeInfo: v1.NodeSystemInfo{
						Architecture:    "arm64",
						OperatingSystem: "windows",
						KernelVersion:   "4.13.0",
						OSImage:         "amazon linux",
					},
				},
			},
		},
	}

	suite.KubeClient = k8s.Client{
		Interface: fake.NewSimpleClientset(nodeList),
	}
}

func (suite *KubeNodeTestSuite) TearDownSuite() {}

func TestKubeNodeTestSuite(t *testing.T) {
	suite.Run(t, &KubeNodeTestSuite{})
}

func (suite *KubeNodeTestSuite) TestGetNodesSummeriesSuccess() {
	// prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	// act
	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
	suite.NoError(err)

	// assert
	expected := []k8s.NodeSummary{
		{
			CPU:             resource.NewScaledQuantity(2000, resource.Milli),
			Memory:          resource.NewScaledQuantity(4000, resource.Mega),
			Name:            "compatible",
			Architecture:    "amd64",
			OperatingSystem: "linux",
			Kernel:          "4.14.0",
			OSImage:         "amazon linux",
			Provider:        "aws://eu-west-3/i-53df4efedd",
		},
		{
			CPU:             resource.NewScaledQuantity(500, resource.Milli),
			Memory:          resource.NewScaledQuantity(1000, resource.Mega),
			Name:            "incompatible",
			Architecture:    "arm64",
			OperatingSystem: "windows",
			Kernel:          "4.13.0",
			OSImage:         "amazon linux",
			Provider:        "aws://eu-west-3/fargate-i-53df4efedd",
		},
	}

	suite.Equal(expected, nodesSummeries)
}

func (suite *KubeNodeTestSuite) TestGenerateNodeReportSuccess() {
	// prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
	suite.NoError(err)

	// act
	compatibleNodesReports, incompatibleNodesReports := k8s.DefaultNodeRequirements.GenerateNodeReports(nodesSummeries)

	// assert
	suite.Len(compatibleNodesReports, 1)
	suite.Len(incompatibleNodesReports, 1)

	incompatibleExpected := &k8s.NodeReport{
		NodeSummary:            &nodesSummeries[1],
		KernelVersionAllowed:   k8s.Requirement{IsCompatible: false, Message: "4.13.0 is unsupported kernel - minimal: 4.14.0"},
		CpuSufficient:          k8s.Requirement{IsCompatible: false, Message: "insufficient cpu - acutal: 500m / minimal: 1750m"},
		MemorySufficient:       k8s.Requirement{IsCompatible: false, Message: "insufficient memory - acutal: 1000Mi / minimal: 1750Mi"},
		ProviderAllowed:        k8s.Requirement{IsCompatible: false, Message: "aws://eu-west-3/fargate-i-53df4efedd is unsupported node provider"},
		ArchitectureAllowed:    k8s.Requirement{IsCompatible: false, Message: "arm64 is unsupported architecture - only amd64 supported"},
		OperatingSystemAllowed: k8s.Requirement{IsCompatible: false, Message: "windows is unsupported os - only linux supported"},
		IsCompatible:           false,
	}

	suite.Equal(incompatibleExpected, incompatibleNodesReports[0])
}

func (suite *KubeNodeTestSuite) TestRequirementErrorMarshalJSONSuccess() {
	//prepare
	err := fmt.Errorf(uuid.New().String())

	//act
	emptyJson, _ := json.Marshal(err)
	nodeRequirementError := k8s.NewNodeRequirementError(err)
	json, _ := json.Marshal(nodeRequirementError)

	// assert
	expectEmpty := []byte("{}")
	expect := []byte(strconv.Quote(err.Error()))

	suite.Equal(expect, json)
	suite.Equal(expectEmpty, emptyJson)
}
