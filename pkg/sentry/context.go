package sentry

import (
	"github.com/getsentry/sentry-go"
	"groundcover.com/pkg/k8s"
)

const (
	HELM_CONTEXT_NAME = "helm"
	KUBE_CONTEXT_NAME = "kubernetes"
)

type SentryContext interface {
	SetOnCurrentScope()
}

type KubeContext struct {
	NodesCount            int               `json:",omitempty"`
	Cluster               string            `json:",omitempty"`
	Namespace             string            `json:",omitempty"`
	Kubeconfig            string            `json:",omitempty"`
	Kubecontext           string            `json:",omitempty"`
	InadequateNodeReports []*k8s.NodeReport `json:",omitempty"`
}

func NewKubeContext(kubeconfig, kubecontext, namespace string) *KubeContext {
	return &KubeContext{
		Namespace:   namespace,
		Kubeconfig:  kubeconfig,
		Kubecontext: kubecontext,
	}
}

func (context KubeContext) SetOnCurrentScope() {
	sentry.CurrentHub().Scope().SetContext(KUBE_CONTEXT_NAME, &context)
}

type HelmContext struct {
	Upgrade              bool   `json:",omitempty"`
	RepoUrl              string `json:",omitempty"`
	ChartName            string `json:",omitempty"`
	ReleaseName          string `json:",omitempty"`
	ChartVersion         string `json:",omitempty"`
	PreviousChartVersion string `json:",omitempty"`
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
