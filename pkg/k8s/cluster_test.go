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

func (suite *KubeClusterTestSuite) SetupTest() {
	suite.KubeClient = k8s.Client{
		Interface: fake.NewSimpleClientset(),
	}
}

func (suite *KubeClusterTestSuite) TearDownSuite() {}

func TestKubeClusterTestSuite(t *testing.T) {
	suite.Run(t, &KubeClusterTestSuite{})
}

func (suite *KubeClusterTestSuite) TestClusterReportSuccess() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	suite.KubeClient.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "24",
		GitVersion: "v1.24.1-test",
	}

	//act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert
	expected := &k8s.ClusterReport{
		IsCompatible: true,
		ServerVersionAllowed: k8s.Requirement{
			IsCompatible: true,
			Message:      "K8s version >= 1.24.0",
		},
		UserAuthorized: k8s.Requirement{
			IsCompatible: true,
			Message:      "User authorized",
		},
	}

	suite.Equal(expected, clusterReport)
}

func (suite *KubeClusterTestSuite) TestClusterReportUserAuthorizedDenied() {
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
		},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert
	expected := k8s.Requirement{
		IsCompatible: false,
		Message:      "denied permissions on resources: pods",
	}

	suite.Equal(expected, clusterReport.UserAuthorized)
}

func (suite *KubeClusterTestSuite) TestClusterReportUserAuthorizedAPIError() {
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

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert
	expected := k8s.Requirement{
		IsCompatible: false,
		Message:      "api error on resource: services: selfsubjectaccessreviews.authorization.k8s.io \"\" already exists",
	}

	suite.Equal(expected, clusterReport.UserAuthorized)
}

func (suite *KubeClusterTestSuite) TestClusterReportServerVersionFail() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	suite.KubeClient.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "23",
		GitVersion: "v1.23.0-test",
	}

	//act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert
	expected := k8s.Requirement{
		IsCompatible: false,
		Message:      "1.23.0-test is unsupported cluster version - minimal: 1.24.0",
	}

	suite.Equal(expected, clusterReport.ServerVersionAllowed)
}

func (suite *KubeClusterTestSuite) TestClusterReportServerVersionUnknown() {
	//prepare
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	suite.KubeClient.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "23",
		GitVersion: "v1.23+.4",
	}

	//act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 23},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, "default")

	// assert
	expected := k8s.Requirement{
		IsCompatible: false,
		Message:      "unknown server version: v1.23+.4",
	}

	suite.Equal(expected, clusterReport.ServerVersionAllowed)
}
