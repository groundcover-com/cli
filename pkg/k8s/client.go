package k8s

import (
	"errors"
	"fmt"
	"io/fs"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	kubernetes.Interface
	clientcmd.ClientConfig
	kubecontext string
}

func NewKubeClient(kubeconfig, kubecontext string) (*Client, error) {
	var err error

	kubeClient := new(Client)

	if err = kubeClient.loadConfig(kubeconfig, kubecontext); err != nil {
		return nil, err
	}

	if err = kubeClient.loadClient(); err != nil {
		return nil, err
	}

	if err = kubeClient.validateClusterConnectivity(); err != nil {
		return nil, fmt.Errorf("couldn't connect to context: %s. maybe do you need to connect via VPN?", kubeClient.kubecontext)
	}

	return kubeClient, nil
}

func (kubeClient *Client) loadConfig(kubeconfig, kubecontext string) error {
	var err error

	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	configLoader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	kubeClient.ClientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoader, configOverrides)

	if kubecontext != "" {
		kubeClient.kubecontext = kubecontext
		return nil
	}

	if kubeClient.kubecontext, err = kubeClient.defaultContext(); err != nil {
		return err
	}

	return nil
}

func (kubeClient *Client) defaultContext() (string, error) {
	var err error
	var rawConfig clientcmdapi.Config

	if rawConfig, err = kubeClient.RawConfig(); err == nil {
		return rawConfig.CurrentContext, nil
	}

	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return "", fmt.Errorf("kubeconfig not found in %s, you can override the path with --kubeconfig flag", pathErr.Path)
	}

	return "", err
}

func (kubeClient *Client) loadClient() error {
	var err error
	var restConfig *restclient.Config

	if restConfig, err = kubeClient.ClientConfig.ClientConfig(); err != nil {
		return err
	}

	OverrideDepartedAuthenticationApiVersion(restConfig)

	if kubeClient.Interface, err = kubernetes.NewForConfig(restConfig); err != nil {
		return kubeClient.printHintIfAuthError(err)
	}

	return nil
}

func (kubeClient *Client) validateClusterConnectivity() error {
	_, err := kubeClient.Discovery().ServerVersion()
	return err
}
