package helm

import (
	"os"

	"github.com/blang/semver/v4"
	"github.com/containerd/containerd/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmCharter struct {
	Name     string
	Version  semver.Version
	repoUrl  string
	chart    *chart.Chart
	settings *cli.EnvSettings
	config   *action.Configuration
}

func NewHelmCharter(name, repoUrl string) (*HelmCharter, error) {
	var err error

	helmCharter := new(HelmCharter)
	helmCharter.Name = name
	helmCharter.repoUrl = repoUrl
	helmCharter.settings = cli.New()
	helmCharter.config = new(action.Configuration)

	if err = helmCharter.config.Init(helmCharter.settings.RESTClientGetter(), helmCharter.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	if err = helmCharter.load(); err != nil {
		return nil, err
	}

	if helmCharter.Version, err = semver.Parse(helmCharter.chart.Metadata.Version); err != nil {
		return nil, err
	}

	return helmCharter, nil
}

func (helmCharter *HelmCharter) load() error {
	var err error
	var chartPath string
	var chart *chart.Chart

	client := action.NewShowWithConfig(action.ShowChart, helmCharter.config)
	client.RepoURL = helmCharter.repoUrl

	if chartPath, err = client.ChartPathOptions.LocateChart(helmCharter.Name, helmCharter.settings); err != nil {
		return err
	}

	if chart, err = loader.Load(chartPath); err != nil {
		return err
	}

	helmCharter.chart = chart
	return nil
}

func (helmCharter *HelmCharter) Get() *chart.Chart {
	return helmCharter.chart
}

func (helmCharter *HelmCharter) IsLatestNewer(currentVersion semver.Version) bool {
	return helmCharter.Version.GT(currentVersion)
}
