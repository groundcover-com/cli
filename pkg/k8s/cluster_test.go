package k8s_test

import (
	"context"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
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

func (suite *KubeClusterTestSuite) TestGetServerVersionSuccess() {
	// arrange
	suite.KubeClient.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "24",
		GitVersion: "v1.24.1",
	}

	// act

	serverVersion, err := suite.KubeClient.GetServerVersion()
	suite.NoError(err)

	// assert
	expected := semver.Version{Major: 1, Minor: 24, Patch: 1}

	suite.Equal(expected, serverVersion)
}

func (suite *KubeClusterTestSuite) TestServerVersionUnknown() {
	// arrange
	suite.KubeClient.Discovery().(*discoveryfake.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "23",
		GitVersion: "v1.23+.4",
	}

	// act
	_, err := suite.KubeClient.GetServerVersion()

	// assert
	suite.ErrorContains(err, "unknown server version v1.23+.4")
}

func (suite *KubeClusterTestSuite) TestClusterReportSuccess() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	// act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := &k8s.ClusterReport{
		ClusterSummary:       clusterSummary,
		IsCompatible:         true,
		UserAuthorized:       k8s.Requirement{IsCompatible: true, Message: "K8s user authorized for groundcover installation"},
		CliAuthSupported:     k8s.Requirement{IsCompatible: true, Message: "K8s CLI auth supported"},
		ServerVersionAllowed: k8s.Requirement{IsCompatible: true, Message: "K8s server version >= 1.24.0"},
		ClusterTypeAllowed:   k8s.Requirement{IsCompatible: true, Message: "K8s cluster type supported"},
	}

	suite.Equal(expected, clusterReport)
}

func (suite *KubeClusterTestSuite) TestClusterReportUserAuthorizedDenied() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	// act
	clusterRequirements := k8s.ClusterRequirements{
		Actions: []*authv1.ResourceAttributes{
			{
				Verb:     "*",
				Resource: "pods",
			},
		},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible:  false,
		Message:       "K8s user authorized for groundcover installation",
		ErrorMessages: []string{"denied permissions on resource: pods"},
	}

	suite.Equal(expected, clusterReport.UserAuthorized)
}

func (suite *KubeClusterTestSuite) TestClusterReportUserAuthorizedAPIError() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		ServerVersion: semver.Version{Major: 1, Minor: 23},
	}

	// act
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

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible: false,
		Message:      "K8s user authorized for groundcover installation",
		ErrorMessages: []string{
			"denied permissions on resource: pods",
			"api error on resource: services: selfsubjectaccessreviews.authorization.k8s.io \"\" already exists",
		},
	}

	suite.Equal(expected, clusterReport.UserAuthorized)
}

func (suite *KubeClusterTestSuite) TestClusterReportServerVersionFail() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		ServerVersion: semver.Version{Major: 1, Minor: 23},
	}

	// act
	clusterRequirements := k8s.ClusterRequirements{
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible:  false,
		Message:       "K8s server version >= 1.24.0",
		ErrorMessages: []string{"1.23.0 is unsupported K8s version"},
	}

	suite.Equal(expected, clusterReport.ServerVersionAllowed)
}

func (suite *KubeClusterTestSuite) TestClusterReportClusterTypeFail() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		Namespace:     "default",
		ClusterName:   "minikube",
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	// act
	clusterRequirements := k8s.ClusterRequirements{
		BlockedTypes:  []string{"minikube"},
		Actions:       []*authv1.ResourceAttributes{},
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	clusterReport := clusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible:  false,
		Message:       "K8s cluster type supported",
		ErrorMessages: []string{"minikube is unsupported cluster type"},
	}

	suite.Equal(expected, clusterReport.ClusterTypeAllowed)
}

func TestValidateAwsCliVersionSupported(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "bad format version, no spaces",
			version:  "bad",
			expected: false,
		},
		{
			name:     "bad format version, first path no slash",
			version:  "bad version",
			expected: false,
		},
		{
			name:     "aws cli version 1.18.0",
			version:  "aws-cli/1.18.0 Python/3.7.4 Darwin/19.4.0 botocore/1.17.0",
			expected: false,
		},
		{
			name:     "aws cli version 1.23.9",
			version:  "aws-cli/1.23.9 Python/3.7.4 Darwin/19.4.0 botocore/1.23.9",
			expected: true,
		},
		{
			name:     "aws cli version 1.23.10",
			version:  "aws-cli/1.23.10 Python/3.7.4 Darwin/19.4.0 botocore/1.23.10",
			expected: true,
		},
		{
			name:     "aws cli version 2.0.0",
			version:  "aws-cli/2.0.0 Python/3.7.4 Darwin/19.4.0 botocore/2.0.0dev0",
			expected: false,
		},
		{
			name:     "aws cli version 2.7.0",
			version:  "aws-cli/2.7.0 Python/3.7.4 Darwin/19.4.0 botocore/2.7.0",
			expected: true,
		},
		{
			name:     "aws cli version 2.7.1",
			version:  "aws-cli/2.7.1 Python/3.7.4 Darwin/19.4.0 botocore/2.7.1",
			expected: true,
		},
		{
			name:     "aws cli version 3.0.0",
			version:  "aws-cli/3.0.0 Python/3.7.4 Darwin/19.4.0 botocore/3.0.0",
			expected: false,
		},
		{
			name:     "aws cli version 0.9.0",
			version:  "aws-cli/0.9.0 Python/3.7.4 Darwin/19.4.0 botocore/0.9.0",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			actual := k8s.ValidateAwsCliVersionSupported(tc.version)

			// assert
			assert.Equal(t, tc.expected, actual)
		})
	}
}
