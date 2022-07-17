package k8s

import (
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Kuber struct {
	kubecontext string
	config      clientcmd.ClientConfig
	client      *kubernetes.Clientset
}

func NewKuber(kubeconfig, kubecontext string) (*Kuber, error) {
	var err error

	kuber := new(Kuber)

	if err = kuber.loadConfig(kubeconfig, kubecontext); err != nil {
		return nil, err
	}

	if err = kuber.loadClient(); err != nil {
		return nil, err
	}

	return kuber, nil
}

func (kuber *Kuber) loadConfig(kubeconfig, kubecontext string) error {
	var err error

	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	configLoader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	kuber.config = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoader, configOverrides)

	if kubecontext != "" {
		kuber.kubecontext = kubecontext
		return nil
	}

	if kuber.kubecontext, err = kuber.defaultContext(); err != nil {
		return err
	}

	return nil
}

func (kuber *Kuber) defaultContext() (string, error) {
	var err error
	var rawConfig clientcmdapi.Config

	if rawConfig, err = kuber.config.RawConfig(); err != nil {
		return "", err
	}

	return rawConfig.CurrentContext, nil
}

func (kuber *Kuber) loadClient() error {
	var err error
	var restConfig *restclient.Config

	if restConfig, err = kuber.config.ClientConfig(); err != nil {
		return err
	}

	if kuber.client, err = kubernetes.NewForConfig(restConfig); err != nil {
		return err
	}

	return nil
}
