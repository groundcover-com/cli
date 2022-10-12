package k8s_test

import (
	"context"
	"testing"
	"time"

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
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pending",
				},
				Spec: v1.NodeSpec{
					ProviderID: "aws://eu-west-3/i-53df4efedg",
					Taints: []v1.Taint{
						{
							Key:    "test",
							Value:  "test",
							Effect: "NoSchedule",
						},
					},
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
	expected := []*k8s.NodeSummary{
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
		{
			CPU:             resource.NewScaledQuantity(2000, resource.Milli),
			Memory:          resource.NewScaledQuantity(4000, resource.Mega),
			Name:            "pending",
			Architecture:    "amd64",
			OperatingSystem: "linux",
			Kernel:          "4.14.0",
			OSImage:         "amazon linux",
			Provider:        "aws://eu-west-3/i-53df4efedg",
			Taints: []v1.Taint{
				{
					Key:    "test",
					Value:  "test",
					Effect: "NoSchedule",
				},
			},
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
	nodesReport := k8s.DefaultNodeRequirements.Validate(nodesSummeries)

	// assert

	expected := &k8s.NodesReport{
		CompatibleNodes: nodesSummeries[:1],
		TaintedNodes: []*k8s.IncompatibleNode{
			{
				NodeSummary: nodesSummeries[2],
				RequirementErrors: []string{
					"taints are set",
				},
			},
		},
		IncompatibleNodes: []*k8s.IncompatibleNode{
			{
				NodeSummary: nodesSummeries[1],
				RequirementErrors: []string{
					"insufficient cpu 500m < 1500m",
					"insufficient memory 1G < 1500Mi",
					"fargate is unsupported provider",
					"4.13.0 is unsupported kernel version",
					"arm64 is unspported architecture",
					"windows is unspported operating system",
				},
			},
		},
		KernelVersionAllowed: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Kernel version >= 4.14.0 (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - 4.13.0 is unsupported kernel version"},
		},
		CpuSufficient: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Sufficient node CPU (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - insufficient cpu 500m < 1500m"},
		},
		MemorySufficient: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Sufficient node memory (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - insufficient memory 1G < 1500Mi"},
		},
		ProviderAllowed: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Cloud provider supported (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - fargate is unsupported provider"},
		},
		ArchitectureAllowed: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Node architecture supported (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - arm64 is unspported architecture"},
		},
		OperatingSystemAllowed: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Node operating system supported (2/3 Nodes)",
			ErrorMessages: []string{"node: incompatible - windows is unspported operating system"},
		},
		Schedulable: k8s.Requirement{
			IsCompatible:  false,
			Message:       "Node is schedulable (2/3 Nodes)",
			ErrorMessages: []string{"node: pending - taints are set"},
		},
	}

	suite.Equal(expected, nodesReport)
}

func (suite *KubeNodeTestSuite) TestNonCompatibleSuccess() {
	// prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
	suite.NoError(err)

	// act
	nodesReport := k8s.DefaultNodeRequirements.Validate(nodesSummeries[1:2])

	// assert

	expected := &k8s.NodesReport{
		IncompatibleNodes: []*k8s.IncompatibleNode{
			{
				NodeSummary: nodesSummeries[1],
				RequirementErrors: []string{
					"insufficient cpu 500m < 1500m",
					"insufficient memory 1G < 1500Mi",
					"fargate is unsupported provider",
					"4.13.0 is unsupported kernel version",
					"arm64 is unspported architecture",
					"windows is unspported operating system",
				},
			},
		},
		KernelVersionAllowed: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Kernel version >= 4.14.0 (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - 4.13.0 is unsupported kernel version"},
		},
		CpuSufficient: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Sufficient node CPU (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - insufficient cpu 500m < 1500m"},
		},
		MemorySufficient: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Sufficient node memory (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - insufficient memory 1G < 1500Mi"},
		},
		ProviderAllowed: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Cloud provider supported (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - fargate is unsupported provider"},
		},
		ArchitectureAllowed: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Node architecture supported (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - arm64 is unspported architecture"},
		},
		OperatingSystemAllowed: k8s.Requirement{
			IsCompatible:    false,
			IsNonCompatible: true,
			Message:         "Node operating system supported (0/1 Nodes)",
			ErrorMessages:   []string{"node: incompatible - windows is unspported operating system"},
		},
		Schedulable: k8s.Requirement{
			IsCompatible: true,
			Message:      "Node is schedulable (1/1 Nodes)",
		},
	}

	suite.Equal(expected, nodesReport)
}
