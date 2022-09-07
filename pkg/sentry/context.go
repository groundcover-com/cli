package sentry

import (
	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"groundcover.com/pkg/k8s"
)

const (
	MAX_NODE_REPORT_SAMPLES  = 10
	HELM_CONTEXT_NAME        = "helm"
	KUBE_CONTEXT_NAME        = "kubernetes"
	SELF_UPDATE_CONTEXT_NAME = "cli-update"
)

type SentryContext interface {
	SetOnCurrentScope()
}

type KubeContext struct {
	NodesCount              int                `json:",omitempty"`
	Kubeconfig              string             `json:",omitempty"`
	Kubecontext             string             `json:",omitempty"`
	ClusterReport           *k8s.ClusterReport `json:",omitempty"`
	IncompatibleNodeReports []*k8s.NodeReport  `json:",omitempty"`
	NodeReportSamples       []*k8s.NodeReport  `json:",omitempty"`
}

func NewKubeContext(kubeconfig, kubecontext string) *KubeContext {
	return &KubeContext{
		Kubeconfig:  kubeconfig,
		Kubecontext: kubecontext,
	}
}

func (context *KubeContext) SetNodeReportsSamples(nodeReports []*k8s.NodeReport) {
	samplesSize := len(nodeReports)
	if samplesSize > MAX_NODE_REPORT_SAMPLES {
		samplesSize = MAX_NODE_REPORT_SAMPLES
	}

	context.NodeReportSamples = make([]*k8s.NodeReport, samplesSize)
	copy(context.NodeReportSamples, nodeReports)
}

func (context KubeContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(KUBE_CONTEXT_NAME, &context)
}

type HelmContext struct {
	Upgrade              bool                   `json:",omitempty"`
	RepoUrl              string                 `json:",omitempty"`
	ChartName            string                 `json:",omitempty"`
	ReleaseName          string                 `json:",omitempty"`
	ChartVersion         string                 `json:",omitempty"`
	PreviousChartVersion string                 `json:",omitempty"`
	ResourcesPresets     []string               `json:",omitempty"`
	ValuesOverride       map[string]interface{} `json:",omitempty"`
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
