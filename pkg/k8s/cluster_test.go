package k8s_test

import (
	"context"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	defaultStorageClass := &v1.StorageClass{
		Provisioner: "local-path",
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}

	storageClass, err := suite.KubeClient.StorageV1().StorageClasses().Create(ctx, defaultStorageClass, metav1.CreateOptions{})
	suite.NoError(err)

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		StorageClass:  storageClass,
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
		ClusterSummary: clusterSummary,
		IsCompatible:   true,
		CliAuthSupported: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s CLI auth supported",
		},
		ServerVersionAllowed: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s server version >= 1.24.0",
		},
		UserAuthorized: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s user authorized for groundcover installation",
		},
		ClusterTypeAllowed: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s cluster type supported",
		},
		StroageProvisional: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s storage provision supported",
		},
	}

	suite.Equal(expected, clusterReport)
}

func (suite *KubeClusterTestSuite) TestClusterReportBetaStorageClassSuccess() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	defaultStorageClass := &v1.StorageClass{
		Provisioner: "local-path",
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"storageclass.beta.kubernetes.io/is-default-class": "true",
			},
		},
	}

	storageClass, err := suite.KubeClient.StorageV1().StorageClasses().Create(ctx, defaultStorageClass, metav1.CreateOptions{})
	suite.NoError(err)

	clusterSummary := &k8s.ClusterSummary{
		ClusterName:   "test",
		Namespace:     "default",
		StorageClass:  storageClass,
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
		ClusterSummary: clusterSummary,
		IsCompatible:   true,
		CliAuthSupported: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s CLI auth supported",
		},
		ServerVersionAllowed: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s server version >= 1.24.0",
		},
		UserAuthorized: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s user authorized for groundcover installation",
		},
		ClusterTypeAllowed: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s cluster type supported",
		},
		StroageProvisional: k8s.Requirement{
			IsCompatible:    true,
			IsNonCompatible: false,
			Message:         "K8s storage provision supported",
		},
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
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s user authorized for groundcover installation",
		ErrorMessages:   []string{"denied permissions on resource: pods"},
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
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s user authorized for groundcover installation",
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
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s server version >= 1.24.0",
		ErrorMessages:   []string{"1.23.0 is unsupported K8s version"},
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
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s cluster type supported",
		ErrorMessages:   []string{"minikube is unsupported cluster type"},
	}

	suite.Equal(expected, clusterReport.ClusterTypeAllowed)
}

func (suite *KubeClusterTestSuite) TestClusterReportStorageDefaultClassFail() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	clusterSummary := &k8s.ClusterSummary{
		Namespace:     "default",
		ClusterName:   "minikube",
		ServerVersion: semver.Version{Major: 1, Minor: 22},
	}

	// act
	clusterReport := k8s.DefaultClusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s storage provision supported",
		ErrorMessages: []string{
			"cluster has no default storage class",
			"Hint:\n  * Define default StorageClass: https://kubernetes.io/docs/concepts/storage/storage-classes/#the-storageclass-resource",
		},
	}

	suite.Equal(expected, clusterReport.StroageProvisional)
}

func (suite *KubeClusterTestSuite) TestClusterReportEbsCsiDriverFail() {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_CONTEXT_TIMEOUT)
	defer cancel()

	defaultStorageClass := &v1.StorageClass{
		Provisioner: "kubernetes.io/aws-ebs",
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}

	storageClass, err := suite.KubeClient.StorageV1().StorageClasses().Create(ctx, defaultStorageClass, metav1.CreateOptions{})
	suite.NoError(err)

	clusterSummary := &k8s.ClusterSummary{
		Namespace:     "default",
		ClusterName:   "arn:aws:eks:us-east-2:0123456789:cluster/test",
		StorageClass:  storageClass,
		ServerVersion: semver.Version{Major: 1, Minor: 24},
	}

	// act
	clusterReport := k8s.DefaultClusterRequirements.Validate(ctx, &suite.KubeClient, clusterSummary)

	// assert
	expected := k8s.Requirement{
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "K8s storage provision supported",
		ErrorMessages: []string{
			"csidrivers.storage.k8s.io \"ebs.csi.aws.com\" not found",
			"Hint: \n  * Install Amazon EBS CSI driver: https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html",
		},
	}

	suite.Equal(expected, clusterReport.StroageProvisional)
}
