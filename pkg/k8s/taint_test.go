package k8s_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	v1 "k8s.io/api/core/v1"
)

type KubeTaintTestSuite struct {
	suite.Suite
	TaintedNodes []*k8s.IncompatibleNode
}

func (suite *KubeTaintTestSuite) SetupSuite() {
	suite.TaintedNodes = []*k8s.IncompatibleNode{
		{
			NodeSummary: &k8s.NodeSummary{
				Taints: []v1.Taint{
					{
						Key:    "test",
						Value:  "test",
						Effect: "NoSchedule",
					},
				},
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				Taints: []v1.Taint{
					{
						Key:    "bad",
						Value:  "bad",
						Effect: "NoSchedule",
					},
				},
			},
		},
		{
			NodeSummary: &k8s.NodeSummary{
				Taints: []v1.Taint{
					{
						Key:    "bad",
						Value:  "bad",
						Effect: "NoSchedule",
					},
				},
			},
		},
	}
}

func (suite *KubeTaintTestSuite) TearDownSuite() {}

func TestKubeTaintTestSuite(t *testing.T) {
	suite.Run(t, &KubeTaintTestSuite{})
}

func (suite *KubeTaintTestSuite) TestGetTaintsSuccess() {
	// prepare
	tolerationManager := &k8s.TolerationManager{
		TaintedNodes: suite.TaintedNodes,
	}

	// act
	taints, err := tolerationManager.GetTaints()
	suite.NoError(err)

	// assert

	expected := []string{
		"{\"key\":\"test\",\"value\":\"test\",\"effect\":\"NoSchedule\"}",
		"{\"key\":\"bad\",\"value\":\"bad\",\"effect\":\"NoSchedule\"}",
	}

	suite.Equal(expected, taints)
}

func (suite *KubeTaintTestSuite) TestGetTolerationsSuccess() {
	// prepare
	tolerationManager := &k8s.TolerationManager{
		TaintedNodes: suite.TaintedNodes,
	}

	// act
	tolerations, err := tolerationManager.GetTolerations([]string{"{\"key\":\"test\",\"value\":\"test\",\"effect\":\"NoSchedule\"}"})
	suite.NoError(err)

	// assert

	expected := []v1.Toleration{
		{
			Key:      "test",
			Value:    "test",
			Operator: "Equal",
			Effect:   "NoSchedule",
		},
	}

	suite.Equal(expected, tolerations)
}

func (suite *KubeTaintTestSuite) TestGetTolerableNodesSuccess() {
	// prepare
	tolerationManager := &k8s.TolerationManager{
		TaintedNodes: suite.TaintedNodes,
	}

	// act
	nodes, err := tolerationManager.GetTolerableNodes([]string{"{\"key\":\"test\",\"value\":\"test\",\"effect\":\"NoSchedule\"}"})
	suite.NoError(err)

	// assert

	expected := []*k8s.NodeSummary{
		suite.TaintedNodes[0].NodeSummary,
	}

	suite.Equal(expected, nodes)
}

// func (suite *KubeTaintTestSuite) TestApproveTaintedNodesSuccess() {
// 	// prepare
// 	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
// 	defer cancel()

// 	nodesSummeries, err := suite.KubeClient.GetNodesSummeries(ctx)
// 	suite.NoError(err)

// 	// act
// 	nodesReport := k8s.DefaultNodeRequirements.Validate(nodesSummeries[2:])
// 	taints, err := nodesReport.GetTaints()
// 	suite.NoError(err)
// 	nodesReport.IdentifyTolerableNodes(taints)

// 	// assert

// 	expected := &k8s.NodesReport{
// 		CompatibleNodes: nodesSummeries[2:],
// 		TaintedNodes: []*k8s.IncompatibleNode{
// 			{
// 				NodeSummary: nodesSummeries[2],
// 				RequirementErrors: []string{
// 					"taints are set",
// 				},
// 			},
// 		},
// 		KernelVersionAllowed: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Kernel version >= 4.14.0 (1/1 Nodes)",
// 		},
// 		CpuSufficient: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Sufficient node CPU (1/1 Nodes)",
// 		},
// 		MemorySufficient: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Sufficient node memory (1/1 Nodes)",
// 		},
// 		ProviderAllowed: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Cloud provider supported (1/1 Nodes)",
// 		},
// 		ArchitectureAllowed: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Node architecture supported (1/1 Nodes)",
// 		},
// 		OperatingSystemAllowed: k8s.Requirement{
// 			IsCompatible: true,
// 			Message:      "Node operating system supported (1/1 Nodes)",
// 		},
// 		Schedulable: k8s.Requirement{
// 			IsCompatible:    false,
// 			IsNonCompatible: true,
// 			Message:         "Node is schedulable (0/1 Nodes)",
// 			ErrorMessages:   []string{"node: pending - taints are set"},
// 		},
// 	}

// 	suite.Equal(expected, nodesReport)
// }
