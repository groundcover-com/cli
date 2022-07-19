package helm

import (
	"context"
	"errors"
	"os"

	"github.com/blang/semver/v4"
	"github.com/containerd/containerd/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type HelmReleaser struct {
	Name      string
	Namespace string
	Version   semver.Version
	settings  *cli.EnvSettings
	config    *action.Configuration
}

func NewHelmReleaser(name, namespace, kubecontext string) (*HelmReleaser, error) {
	var err error

	helmReleaser := &HelmReleaser{
		Name:     name,
		settings: cli.New(),
		config:   new(action.Configuration),
	}

	helmReleaser.Namespace = namespace
	helmReleaser.settings.KubeContext = kubecontext

	if err = helmReleaser.config.Init(helmReleaser.settings.RESTClientGetter(), helmReleaser.Namespace, os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	return helmReleaser, nil
}

func (helmReleaser *HelmReleaser) Get() (*release.Release, error) {
	var err error
	var version semver.Version
	var release *release.Release

	client := action.NewStatus(helmReleaser.config)

	if release, err = client.Run(helmReleaser.Name); err != nil {
		return nil, err
	}

	if version, err = semver.Parse(release.Chart.Metadata.Version); err != nil {
		return nil, err
	}

	helmReleaser.Version = version
	return release, nil
}

func (helmReleaser *HelmReleaser) Install(ctx context.Context, chart *chart.Chart, values map[string]interface{}) error {
	var err error
	var version semver.Version

	client := action.NewInstall(helmReleaser.config)
	client.Wait = false
	client.CreateNamespace = true
	client.ReleaseName = helmReleaser.Name
	client.Namespace = helmReleaser.Namespace

	if version, err = semver.Parse(chart.Metadata.Version); err != nil {
		return err
	}

	if _, err = client.RunWithContext(ctx, chart, values); err != nil {
		return err
	}

	helmReleaser.Version = version
	return nil
}

func (helmReleaser *HelmReleaser) Upgrade(ctx context.Context, chart *chart.Chart, values map[string]interface{}) error {
	var err error
	var version semver.Version

	client := action.NewUpgrade(helmReleaser.config)
	client.Wait = false
	client.ReuseValues = true
	client.Namespace = helmReleaser.Namespace

	if version, err = semver.Parse(chart.Metadata.Version); err != nil {
		return err
	}

	_, err = client.RunWithContext(ctx, helmReleaser.Name, chart, values)

	switch {
	case err == nil:
		helmReleaser.Version = version
		return nil
	case errors.Is(err, driver.ErrNoDeployedReleases), errors.Is(err, driver.ErrReleaseNotFound):
		return helmReleaser.Install(ctx, chart, values)
	default:
		return err
	}
}

func (helmReleaser *HelmReleaser) Uninstall() error {
	var err error

	client := action.NewUninstall(helmReleaser.config)
	client.Wait = false

	if _, err = client.Run(helmReleaser.Name); err != nil {
		return err
	}

	helmReleaser.Version = semver.Version{}
	return nil
}
