package sentry

import (
	"time"

	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/ui"
)

const (
	MAX_NODE_REPORT_SAMPLES  = 10
	COMMAND_CONTEXT_NAME     = "command"
	HELM_CONTEXT_NAME        = "helm"
	KUBE_CONTEXT_NAME        = "kubernetes"
	SELF_UPDATE_CONTEXT_NAME = "cli-update"
)

type SentryContext interface {
	SetOnCurrentScope()
}

type CommandContext struct {
	Name string     `json:",omitempty"`
	Took string     `json:",omitempty"`
	Log  *ui.Writer `json:",omitempty"`
}

func NewCommandContext(start time.Time) *CommandContext {
	return &CommandContext{
		Name: sentry.CurrentHub().Scope().Transaction(),
		Took: time.Since(start).Round(time.Second).String(),
		Log:  ui.SingletonWriter,
	}
}

func (context CommandContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(COMMAND_CONTEXT_NAME, &context)
}

type KubeContext struct {
	NodesCount                int                     `json:",omitempty"`
	Kubeconfig                string                  `json:",omitempty"`
	Kubecontext               string                  `json:",omitempty"`
	TolerationsAndTaintsRatio string                  `json:",omitempty"`
	ClusterReport             *k8s.ClusterReport      `json:",omitempty"`
	CompatibleNodeSamples     []*k8s.NodeSummary      `json:",omitempty"`
	IncompatibleNodeSamples   []*k8s.IncompatibleNode `json:",omitempty"`
	TaintedNodeSamples        []*k8s.IncompatibleNode `json:",omitempty"`
}

func NewKubeContext(kubeconfig, kubecontext string) *KubeContext {
	return &KubeContext{
		Kubeconfig:  kubeconfig,
		Kubecontext: kubecontext,
	}
}

func (context *KubeContext) SetNodesSamples(nodesReport *k8s.NodesReport) {
	compatibleSamplesSize := len(nodesReport.CompatibleNodes)
	if compatibleSamplesSize > MAX_NODE_REPORT_SAMPLES {
		compatibleSamplesSize = MAX_NODE_REPORT_SAMPLES
	}

	context.CompatibleNodeSamples = make([]*k8s.NodeSummary, compatibleSamplesSize)
	copy(context.CompatibleNodeSamples, nodesReport.CompatibleNodes)

	incompatibleSamplesSize := len(nodesReport.IncompatibleNodes)
	if incompatibleSamplesSize > MAX_NODE_REPORT_SAMPLES {
		incompatibleSamplesSize = MAX_NODE_REPORT_SAMPLES
	}

	context.IncompatibleNodeSamples = make([]*k8s.IncompatibleNode, incompatibleSamplesSize)
	copy(context.IncompatibleNodeSamples, nodesReport.IncompatibleNodes)

	incompatibleTaintsSize := len(nodesReport.TaintedNodes)
	if incompatibleTaintsSize > MAX_NODE_REPORT_SAMPLES {
		incompatibleTaintsSize = MAX_NODE_REPORT_SAMPLES
	}

	context.TaintedNodeSamples = make([]*k8s.IncompatibleNode, incompatibleTaintsSize)
	copy(context.TaintedNodeSamples, nodesReport.TaintedNodes)
}

func (context KubeContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(KUBE_CONTEXT_NAME, &context)
}

type HelmContext struct {
	Upgrade              bool                       `json:",omitempty"`
	RepoUrl              string                     `json:",omitempty"`
	ChartName            string                     `json:",omitempty"`
	ReleaseName          string                     `json:",omitempty"`
	ChartVersion         string                     `json:",omitempty"`
	RunningAlligators    string                     `json:",omitempty"`
	PreviousChartVersion string                     `json:",omitempty"`
	ResourcesPresets     []string                   `json:",omitempty"`
	ValuesOverride       map[string]interface{}     `json:",omitempty"`
	PodsStatus           map[string]k8s.PodStatus   `json:",omitempty"`
	BoundPvcs            []string                   `json:",omitempty"`
	AllocatableResources *helm.AllocatableResources `json:",omitempty"`
}

func NewHelmContext(releaseName, chartName, repoUrl string) *HelmContext {
	return &HelmContext{
		RepoUrl:     repoUrl,
		ChartName:   chartName,
		ReleaseName: releaseName,
	}
}

func (context HelmContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(HELM_CONTEXT_NAME, &context)
}

type SelfUpdateContext struct {
	CurrentVersion semver.Version `json:",omitempty"`
	LatestVersion  semver.Version `json:",omitempty"`
}

func NewSelfUpdateContext(currentVersion, latestVersion semver.Version) *SelfUpdateContext {
	return &SelfUpdateContext{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
	}
}

func (context SelfUpdateContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(SELF_UPDATE_CONTEXT_NAME, &context)
}
