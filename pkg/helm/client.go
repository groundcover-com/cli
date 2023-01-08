package helm

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/log"
	"groundcover.com/pkg/k8s"
	"groundcover.com/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

type Client struct {
	settings *cli.EnvSettings
	cfg      *action.Configuration
}

func NewHelmClient(namespace, kubecontext string) (*Client, error) {
	var err error

	helmPath := filepath.Join(utils.PresistentStorage.BasePath, "helm")

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

	getter, ok := helmClient.settings.RESTClientGetter().(*genericclioptions.ConfigFlags)
	if !ok {
		return nil, errors.New("failed to cast helm rest client getter")
	}

	getter.WrapConfigFn = func(restConfig *rest.Config) *rest.Config {
		k8s.OverrideDepartedAuthenticationApiVersion(restConfig)
		return restConfig
	}

	if err = helmClient.cfg.Init(getter, helmClient.settings.Namespace(), os.Getenv("HELM_DRIVER"), log.L.Debugf); err != nil {
		return nil, err
	}

	return helmClient, nil
}
