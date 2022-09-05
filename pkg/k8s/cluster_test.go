package k8s_test

import (
	"context"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

type KubeClusterTestSuite struct {
	suite.Suite
	KubeClient k8s.Client
}

func (suite *KubeClusterTestSuite) SetupSuite() {

	client := fake.NewSimpleClientset()
	client.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "24",
	}

	suite.KubeClient = k8s.Client{
		Interface: client,
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
	suite.EqualError(errSlice[0], "permission error on resource: pods")
	suite.ErrorContains(errSlice[1], "api error on resource: services")
}

func (suite *KubeClusterTestSuite) TestClusterVersionError() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	//act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 25},
	}

	errSlice := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert

	suite.Len(errSlice, 1)
	suite.EqualError(errSlice[0], "1.24.0 is unsupported cluster version - minimal: 1.25.0")
}
