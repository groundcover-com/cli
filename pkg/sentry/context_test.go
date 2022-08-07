package sentry_test

import (
	"encoding/json"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
)

type SentryContextTestSuite struct {
	suite.Suite
	Transport *TransportMock
}

func (suite *SentryContextTestSuite) SetupSuite() {
	suite.Transport = &TransportMock{}

	clientOptions := sentry.ClientOptions{
		Dsn:          "http://whatever@really.com/1337",
		Transport:    suite.Transport,
		Integrations: func(i []sentry.Integration) []sentry.Integration { return []sentry.Integration{} },
	}

	client, _ := sentry.NewClient(clientOptions)
	sentry.CurrentHub().BindClient(client)
}

func (suite *SentryContextTestSuite) TearDownSuite() {}

func TestSentryContextSuite(t *testing.T) {
	suite.Run(t, &SentryContextTestSuite{})
}

func (suite *SentryContextTestSuite) TestKubeContexJsonOmitEmpty() {
	//prepare
	sentryContext := &sentry_utils.KubeContext{}

	//act
	json, err := json.Marshal(sentryContext)
	suite.NoError(err)

	// assert
	expect := []byte("{}")
	suite.Equal(expect, json)
}

func (suite *SentryContextTestSuite) TestKubeContextSetOnCurrentScopeSuccess() {
	//prepare
	nodesCount := 2
	cluster := uuid.New().String()
	namespace := uuid.New().String()
	kubeconfig := uuid.New().String()
	kubecontext := uuid.New().String()

	sentryContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext, namespace)
	sentryContext.Cluster = cluster
	sentryContext.NodesCount = nodesCount

	//act
	sentryContext.SetOnCurrentScope()
	sentry.CaptureMessage("kube context")

	// assert
	expect := map[string]interface{}{
		"kubernetes": &sentry_utils.KubeContext{
			Cluster:               cluster,
			Namespace:             namespace,
			NodesCount:            nodesCount,
			Kubeconfig:            kubeconfig,
			Kubecontext:           kubecontext,
			NodeReportSamples:     make([]*k8s.NodeReport, sentry_utils.MAX_NODE_REPORT_SAMPLES),
			ServerVersion:         nil,
			InadequateNodeReports: nil,
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.KUBE_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}

func (suite *SentryContextTestSuite) TestHelmContexJsonOmitEmpty() {
	//prepare
	sentryContext := &sentry_utils.HelmContext{}

	//act
	json, err := json.Marshal(sentryContext)
	suite.NoError(err)

	// assert
	expect := []byte("{}")
	suite.Equal(expect, json)
}

func (suite *SentryContextTestSuite) TestHelmContextSetOnCurrentScopeSuccess() {
	//prepare
	chartVersion := "1.0.0"
	previousChartVersion := "0.9.0"
	repoUrl := uuid.New().String()
	chartName := uuid.New().String()
	releaseName := uuid.New().String()

	sentryContext := sentry_utils.NewHelmContext(releaseName, chartName, repoUrl)
	sentryContext.Upgrade = true
	sentryContext.ChartVersion = chartVersion
	sentryContext.PreviousChartVersion = previousChartVersion

	//act
	sentryContext.SetOnCurrentScope()
	sentry.CaptureMessage("helm context")

	// assert
	expect := map[string]interface{}{
		"helm": &sentry_utils.HelmContext{
			Upgrade:              true,
			RepoUrl:              repoUrl,
			ChartName:            chartName,
			ReleaseName:          releaseName,
			ChartVersion:         chartVersion,
			PreviousChartVersion: previousChartVersion,
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.HELM_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}
