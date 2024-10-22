package sentry_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	v1 "k8s.io/api/core/v1"
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
	// prepare
	nodesCount := 2
	kubeconfig := uuid.New().String()
	kubecontext := uuid.New().String()
	tolerationsAndTaintsRatio := "1/1"

	sentryContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)
	sentryContext.NodesCount = nodesCount
	sentryContext.TolerationsAndTaintsRatio = tolerationsAndTaintsRatio

	// act
	sentryContext.SetOnCurrentScope()
	sentry.CaptureMessage("kube context")

	// assert
	expect := map[string]interface{}{
		"kubernetes": &sentry_utils.KubeContext{
			NodesCount:                nodesCount,
			Kubeconfig:                kubeconfig,
			Kubecontext:               kubecontext,
			TolerationsAndTaintsRatio: tolerationsAndTaintsRatio,
			CompatibleNodeSamples:     nil,
			IncompatibleNodeSamples:   nil,
			ClusterReport:             nil,
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.KUBE_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}

func (suite *SentryContextTestSuite) TestKubeContextSetNodeReportSamplesDoesNotExceedMaxLenght() {
	// prepare
	kubeconfig := uuid.New().String()
	kubecontext := uuid.New().String()
	nodesCount := sentry_utils.MAX_NODE_REPORT_SAMPLES + 2

	nodesReport := &k8s.NodesReport{
		CompatibleNodes:   make([]*k8s.NodeSummary, nodesCount),
		IncompatibleNodes: make([]*k8s.IncompatibleNode, nodesCount),
	}

	sentryContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)

	// act
	sentryContext.SetNodesSamples(nodesReport)

	// assert
	expectCompatibleNodes := nodesReport.CompatibleNodes[:sentry_utils.MAX_NODE_REPORT_SAMPLES]
	expectInCompatibleNodes := nodesReport.IncompatibleNodes[:sentry_utils.MAX_NODE_REPORT_SAMPLES]

	suite.Equal(expectCompatibleNodes, sentryContext.CompatibleNodeSamples)
	suite.Equal(expectInCompatibleNodes, sentryContext.IncompatibleNodeSamples)
}

func (suite *SentryContextTestSuite) TestKubeContextSetNodeReportSamplesWithTaints() {
	// prepare
	kubeconfig := uuid.New().String()
	kubecontext := uuid.New().String()

	nodesReport := &k8s.NodesReport{
		TaintedNodes: []*k8s.IncompatibleNode{
			{
				NodeSummary: &k8s.NodeSummary{
					Name: "node",
					Taints: []v1.Taint{
						{
							Key:    "key",
							Value:  "value",
							Effect: "effect",
						},
					},
				},
			},
		},
	}

	sentryContext := sentry_utils.NewKubeContext(kubeconfig, kubecontext)

	// act
	sentryContext.SetNodesSamples(nodesReport)

	// assert
	suite.Equal(nodesReport.TaintedNodes, sentryContext.TaintedNodeSamples)
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
	runningSensors := "1/1"
	previousChartVersion := "0.9.0"
	repoUrl := uuid.New().String()
	chartName := uuid.New().String()
	releaseName := uuid.New().String()
	resourcesPresets := []string{uuid.New().String()}
	valuesOverride := map[string]interface{}{"override": uuid.New().String()}

	sentryContext := sentry_utils.NewHelmContext(releaseName, chartName, repoUrl)
	sentryContext.Upgrade = true
	sentryContext.ChartVersion = chartVersion
	sentryContext.PreviousChartVersion = previousChartVersion
	sentryContext.ValuesOverride = valuesOverride
	sentryContext.ResourcesPresets = resourcesPresets
	sentryContext.RunningSensors = runningSensors

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
			ValuesOverride:       valuesOverride,
			ResourcesPresets:     resourcesPresets,
			RunningSensors:       runningSensors,
			PreviousChartVersion: previousChartVersion,
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.HELM_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}

func (suite *SentryContextTestSuite) TestSelfUpdateContextSetOnCurrentScopeSuccess() {
	//prepare
	currentVersion := semver.MustParse("0.1.0")
	lastestVersion := semver.MustParse("1.0.0")

	sentryContext := sentry_utils.NewSelfUpdateContext(currentVersion, lastestVersion)

	//act
	sentryContext.SetOnCurrentScope()
	sentry.CaptureMessage("cli update context")

	// assert
	expect := map[string]interface{}{
		"cli-update": &sentry_utils.SelfUpdateContext{
			CurrentVersion: currentVersion,
			LatestVersion:  lastestVersion,
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.SELF_UPDATE_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}

func (suite *SentryContextTestSuite) TestCommandContextSetOnCurrentScopeSuccess() {
	//prepare
	start := time.Now()
	sentry.CurrentHub().Scope().SetTransaction("test")

	//act
	sentryContext := sentry_utils.NewCommandContext(start)
	sentryContext.SetOnCurrentScope()
	sentry.CaptureMessage("command context")

	// assert
	expect := map[string]interface{}{
		"command": &sentry_utils.CommandContext{
			Name: "test",
			Took: "0s",
			Log:  ui.NewWriter(),
		},
	}

	event := suite.Transport.lastEvent
	sentry.CurrentHub().Scope().RemoveContext(sentry_utils.COMMAND_CONTEXT_NAME)

	suite.Equal(expect, event.Contexts)
}
