package helm

import (
	"os"
	"path/filepath"

	"github.com/containerd/containerd/log"
	"groundcover.com/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type Client struct {
	settings *cli.EnvSettings
	cfg      *action.Configuration
}

func NewHelmClient(namespace, kubecontext string) (*Client, error) {
	var err error

	helmPath := filepath.Join(utils.PresistentStorage.BasePath, "/helm")

	os.Setenv("HELM_DATA_HOME", helmPath)
	os.Setenv("HELM_CACHE_HOME", helmPath)
	os.Setenv("HELM_CONFIG_HOME", helmPath)

	helmClient := &Client{
		settings: cli.New(),
		cfg:      new(action.Configuration),
	}

	helmClient.settings.Debug = true
	helmClient.settings.SetNamespace(namespace)
	helmClient.settings.KubeContext = kubecontext

	if err = helmClient.cfg.Init(helmClient.settings.RESTClientGetter(), helmClient.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	return helmClient, nil
}
