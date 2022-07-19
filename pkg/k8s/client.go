package k8s

import (
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubeClient struct {
	*kubernetes.Clientset
	kubecontext string
	config      clientcmd.ClientConfig
}

func NewKubeClient(kubeconfig, kubecontext string) (*KubeClient, error) {
	var err error

	kubeClient := new(KubeClient)

	if err = kubeClient.loadConfig(kubeconfig, kubecontext); err != nil {
		return nil, err
	}

	if err = kubeClient.loadClient(); err != nil {
		return nil, err
	}

	return kubeClient, nil
}

func (kubeClient *KubeClient) loadConfig(kubeconfig, kubecontext string) error {
	var err error

	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	configLoader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	kubeClient.config = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoader, configOverrides)

	if kubecontext != "" {
		kubeClient.kubecontext = kubecontext
		return nil
	}

	if kubeClient.kubecontext, err = kubeClient.defaultContext(); err != nil {
		return err
	}

	return nil
}

func (kubeClient *KubeClient) defaultContext() (string, error) {
	var err error
	var rawConfig clientcmdapi.Config

	if rawConfig, err = kubeClient.config.RawConfig(); err != nil {
		return "", err
	}

	return rawConfig.CurrentContext, nil
}

func (kubeClient *KubeClient) loadClient() error {
	var err error
	var restConfig *restclient.Config

	if restConfig, err = kubeClient.config.ClientConfig(); err != nil {
		return err
	}

	if kubeClient.Clientset, err = kubernetes.NewForConfig(restConfig); err != nil {
		return err
	}

	return nil
}
