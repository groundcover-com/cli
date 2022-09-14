package helm

import (
	"fmt"
	"os"

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

	helmPath := fmt.Sprintf("%s/helm", utils.PresistentStorage.BasePath)

	helmClient := &Client{
		settings: cli.New(),
		cfg:      new(action.Configuration),
	}

	helmClient.settings.Debug = true
	helmClient.settings.SetNamespace(namespace)
	helmClient.settings.KubeContext = kubecontext
	helmClient.settings.PluginsDirectory = fmt.Sprintf("%s/plugins", helmPath)
	helmClient.settings.RepositoryCache = fmt.Sprintf("%s/repository", helmPath)
	helmClient.settings.RepositoryConfig = fmt.Sprintf("%s/repositories.yaml", helmPath)
	helmClient.settings.RegistryConfig = fmt.Sprintf("%s/registry/config.json", helmPath)

	if err = helmClient.cfg.Init(helmClient.settings.RESTClientGetter(), helmClient.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	return helmClient, nil
}
