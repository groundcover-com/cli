package helm

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	GROUNDCOVER_HELM_REPO_ADDR = "https://helm.groundcover.com"
	GROUNDCOVER_HELM_REPO_NAME = "groundcover"
	GROUNDCOVER_CHART_NAME     = "groundcover/groundcover"
	HELM_VERSION_REGEX         = "version: (.*)"
	HELM_BINARY_NAME           = "helm"
)

var (
	helmVersionRegex = regexp.MustCompile(HELM_VERSION_REGEX)
)

type HelmCmd struct {
	helmPath  string
	repoName  string
	chartName string
	repoAddr  string
}

func NewHelmCmd() (*HelmCmd, error) {
	path, err := getHelmExecutablePath()
	if err != nil {
		return nil, err
	}

	return &HelmCmd{
		helmPath:  path,
		repoName:  GROUNDCOVER_HELM_REPO_NAME,
		chartName: GROUNDCOVER_CHART_NAME,
		repoAddr:  GROUNDCOVER_HELM_REPO_ADDR,
	}, nil
}

func getHelmExecutablePath() (string, error) {
	helmPath, err := exec.LookPath(HELM_BINARY_NAME)
	if err != nil {
		return "", errors.New("Failed to find helm executable. make sure helm is installed and in your PATH")
	}

	return helmPath, nil
}

func (h *HelmCmd) Upgrade(ctx context.Context, apiKey string, clusterName string, namespace string) error {
	helmUpgradeCmd := exec.Command(h.helmPath, "upgrade", "--install", h.repoName, h.chartName,
		"--set", fmt.Sprintf("global.groundcover_token=%s", apiKey),
		"--set", fmt.Sprintf("clusterId=%s", clusterName),
		"--create-namespace", "-n", namespace,
	)

	if err := helmUpgradeCmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade helm chart. error: %s", err.Error())
	}

	return nil
}

func (h *HelmCmd) RepoAdd(ctx context.Context) error {
	helmBinary, err := getHelmExecutablePath()
	if err != nil {
		return err
	}

	helmRepoAddCmd := exec.Command(helmBinary, "repo", "add", h.repoName, h.repoAddr)
	if err := helmRepoAddCmd.Run(); err != nil {
		return fmt.Errorf("failed to add helm repo. error: %s", err.Error())
	}

	return nil
}

func (h *HelmCmd) RepoUpdate(ctx context.Context) error {
	helmBinary, err := getHelmExecutablePath()
	if err != nil {
		return err
	}

	helmRepoUpdateCmd := exec.Command(helmBinary, "repo", "update", h.repoName)
	if err := helmRepoUpdateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update helm repo. error: %s", err.Error())
	}

	return nil
}

func (h *HelmCmd) GetLatestChartVersion(ctx context.Context) (string, error) {
	err := h.RepoAdd(ctx)
	if err != nil {
		return "", err
	}

	err = h.RepoUpdate(ctx)
	if err != nil {
		return "", err
	}

	chartCommandOutput, err := h.ShowChartCommand(ctx)
	if err != nil {
		return "", err
	}

	matches := helmVersionRegex.FindStringSubmatch(chartCommandOutput)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to get groundcover version. failed to parse helm version")
	}

	return matches[1], nil
}

func (h *HelmCmd) ShowChartCommand(ctx context.Context) (string, error) {
	helmGetVersionCmd := exec.Command(h.helmPath, "show", "chart", h.chartName)

	output := strings.Builder{}
	helmGetVersionCmd.Stdout = &output
	if err := helmGetVersionCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get run show chart. error: %s", err.Error())
	}

	return output.String(), nil
}

func (h *HelmCmd) BuildInstallCommand(apiKey, clusterName, namespace string) string {
	return fmt.Sprintf("helm repo add %s %s && helm repo update %s && helm upgrade --install %s %s --set global.groundcover_token=%s,clusterId=%s --create-namespace -n %s\n",
		h.repoName,
		h.repoAddr,
		h.repoName,
		h.repoName,
		h.chartName,
		apiKey,
		clusterName,
		namespace)
}

func (h *HelmCmd) Uninstall(ctx context.Context, namespace string, helmRelease string) error {
	helmUninstallCmd := exec.Command(h.helmPath, "uninstall", "--namespace", namespace, helmRelease)
	if err := helmUninstallCmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall groundcover. error: %s", err.Error())
	}

	return nil
}
