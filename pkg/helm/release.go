package helm

import (
	"context"
	"errors"

	"github.com/blang/semver/v4"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Release struct {
	*release.Release
}

func (release *Release) Version() semver.Version {
	version, _ := semver.Parse(release.Chart.Metadata.Version)
	return version
}

func (helmClient *Client) IsReleaseInstalled(name string) (*Release, bool, error) {
	var err error

	client := action.NewStatus(helmClient.cfg)

	var release Release
	release.Release, err = client.Run(name)

	switch {
	case errors.Is(err, driver.ErrReleaseNotFound):
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &release, true, nil
	}
}

func (helmClient *Client) GetCurrentRelease(name string) (*Release, error) {
	var err error

	client := action.NewStatus(helmClient.cfg)

	var release Release
	if release.Release, err = client.Run(name); err != nil {
		return nil, err
	}

	return &release, nil
}

func (helmClient *Client) Install(ctx context.Context, name string, chart *Chart, values map[string]interface{}) (*Release, error) {
	var err error

	client := action.NewInstall(helmClient.cfg)
	client.Wait = false
	client.ReleaseName = name
	client.CreateNamespace = true
	client.Namespace = helmClient.settings.Namespace()

	var release Release
	if release.Release, err = client.RunWithContext(ctx, chart.Chart, values); err != nil {
		return nil, err
	}

	return &release, nil
}

func (helmClient *Client) Upgrade(ctx context.Context, name string, chart *Chart, values map[string]interface{}) (*Release, error) {
	var err error

	client := action.NewUpgrade(helmClient.cfg)
	client.Wait = false
	client.ReuseValues = false
	client.Namespace = helmClient.settings.Namespace()

	var release Release
	release.Release, err = client.RunWithContext(ctx, name, chart.Chart, values)

	switch {
	case err == nil:
		return &release, nil
	case errors.Is(err, driver.ErrNoDeployedReleases), errors.Is(err, driver.ErrReleaseNotFound):
		return helmClient.Install(ctx, name, chart, values)
	default:
		return nil, err
	}
}

func (helmClient *Client) Uninstall(name string) error {
	var err error

	client := action.NewUninstall(helmClient.cfg)
	client.Wait = false

	if _, err = client.Run(name); err != nil {
		return err
	}

	return nil
}
