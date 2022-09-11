package k8s

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	kubernetes.Interface
	clientcmd.ClientConfig
	kubecontext string
}

type Requirement struct {
	IsCompatible  bool     `json:",omitempty"`
	Message       string   `json:"-"`
	ErrorMessages []string `json:"-"`
}

func (requirement Requirement) PrintStatus() {
	var messageBuffer strings.Builder

	messageBuffer.WriteString(requirement.Message)
	messageBuffer.WriteString("\n")

	for _, errorMessage := range requirement.ErrorMessages {
		messageBuffer.WriteString(color.RedString(ui.Bullet))
		messageBuffer.WriteString(errorMessage)
		messageBuffer.WriteString("\n")
	}

	ui.PrintStatus(requirement.IsCompatible, messageBuffer.String())
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

	if kubeClient.Interface, err = kubernetes.NewForConfig(restConfig); err != nil {
		return kubeClient.printHintIfAuthError(err)
	}

	return nil
}
