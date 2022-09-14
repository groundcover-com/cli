package helm

import (
	"github.com/blang/semver/v4"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type Chart struct {
	*chart.Chart
}

func (chart *Chart) Version() semver.Version {
	version, _ := semver.Parse(chart.Metadata.Version)
	return version
}

func (helmClient *Client) GetLatestChart(name string) (*Chart, error) {
	var err error
	var chartPath string
	var chart *chart.Chart

	client := action.NewShowWithConfig(action.ShowChart, helmClient.cfg)

	if chartPath, err = client.ChartPathOptions.LocateChart(name, helmClient.settings); err != nil {
		return nil, err
	}

	if chart, err = loader.Load(chartPath); err != nil {
		return nil, err
	}

	return &Chart{Chart: chart}, nil
}
