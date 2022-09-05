package k8s_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type KubeClusterTestSuite struct {
	suite.Suite
	KubeClient k8s.Client
}

func (suite *KubeClusterTestSuite) SetupSuite() {
	suite.KubeClient = k8s.Client{
		Interface: fake.NewSimpleClientset(),
	}
}

func (suite *KubeClusterTestSuite) TearDownSuite() {}

func TestKubeClusterTestSuite(t *testing.T) {
	suite.Run(t, &KubeClusterTestSuite{})
}

func (suite *KubeClusterTestSuite) TestClusterAuthErrors() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	//act
	clusterRequirements := k8s.ClusterRequirements{
		Actions: []*authv1.ResourceAttributes{
			{
				Verb:     "*",
				Resource: "pods",
			},
			{
				Verb:     "*",
				Resource: "services",
			},
		},
	}

	errSlice := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert

	suite.Len(errSlice, 2)
	suite.Errorf(errSlice[0], "permission error on pods")
	suite.ErrorContains(errSlice[1], "api error on services")
}
