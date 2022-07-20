package helm

import (
	"os"

	"github.com/containerd/containerd/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type Client struct {
	settings *cli.EnvSettings
	cfg      *action.Configuration
}

func NewHelmClient(namespace, kubecontext string) (*Client, error) {
	var err error

	helmClient := &Client{
		settings: cli.New(),
		cfg:      new(action.Configuration),
	}

	helmClient.settings.SetNamespace(namespace)
	helmClient.settings.KubeContext = kubecontext

	if err = helmClient.cfg.Init(helmClient.settings.RESTClientGetter(), helmClient.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	return helmClient, nil
}
