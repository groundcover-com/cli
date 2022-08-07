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
					Name: "adequate",
				},
				Spec: v1.NodeSpec{
					ProviderID: "aws://eu-west-3/i-53df4efedd",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("4G"),
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
					Name: "inadequate",
				},
				Spec: v1.NodeSpec{
					ProviderID: "aws://eu-west-3/fargate-i-53df4efedd",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("0"),
						v1.ResourceMemory: resource.MustParse("1G"),
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
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	//act
	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
	suite.NoError(err)

	// assert

	expected := []k8s.NodeSummary{
		{
			CPU:             2,
			Memory:          4,
			Name:            "adequate",
			Architecture:    "amd64",
			OperatingSystem: "linux",
			Kernel:          "4.14.0",
			OSImage:         "amazon linux",
			Provider:        "aws://eu-west-3/i-53df4efedd",
		},
		{
			CPU:             0,
			Memory:          1,
			Name:            "inadequate",
			Architecture:    "arm64",
			OperatingSystem: "windows",
			Kernel:          "4.13.0",
			OSImage:         "amazon linux",
			Provider:        "aws://eu-west-3/fargate-i-53df4efedd",
		},
	}

	suite.Equal(expected, nodesSummeries)
}

func (suite *KubeNodeTestSuite) TestGenerateNodeReportsSuccess() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
	suite.NoError(err)

	nodeRequirements := k8s.NewNodeMinimumRequirements()

	//act
	adequateNodesReports, inadequateNodesReports := nodeRequirements.GenerateNodeReports(nodesSummeries)

	// assert
	adequateExpected := make([]*k8s.NodeReport, 1)
	adequateExpected[0] = &k8s.NodeReport{
		IsAdequate:  true,
		NodeSummary: &nodesSummeries[0],
	}

	suite.Equal(adequateExpected, adequateNodesReports)

	inadequateExpected := make([]*k8s.NodeReport, 1)
	inadequateExpected[0] = &k8s.NodeReport{
		IsAdequate:  false,
		NodeSummary: &nodesSummeries[1],
		Errors: []error{
			k8s.NewNodeRequirementError(fmt.Errorf("insufficient cpu - acutal: 0 / minimal: 1")),
			k8s.NewNodeRequirementError(fmt.Errorf("insufficient memory - acutal: 1G / minimal: 2G")),
			k8s.NewNodeRequirementError(fmt.Errorf("aws://eu-west-3/fargate-i-53df4efedd is unsupported node provider")),
			k8s.NewNodeRequirementError(fmt.Errorf("4.13.0 is unsupported kernel - minimal: 4.14.0")),
			k8s.NewNodeRequirementError(fmt.Errorf("arm64 is unsupported architecture - only amd64 supported")),
			k8s.NewNodeRequirementError(fmt.Errorf("windows is unsupported os - only linux supported")),
		},
	}

	suite.Equal(inadequateExpected, inadequateNodesReports)
}

func (suite *KubeNodeTestSuite) TestNodeRequirementErrorMarshalJSONSuccess() {
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
