package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

const (
	KUBECTL_BINARY_NAME = "kubectl"
)

func GetKubectlPath() (string, error) {
	kubectlPath, err := exec.LookPath(KUBECTL_BINARY_NAME)
	if err != nil {
		return "", errors.New("Failed to find kubectl executable. make sure kubectl is installed and in your PATH")
	}

	return kubectlPath, nil
}

func Delete(ctx context.Context, namespace string, objectName string) error {
	deleteTsdbConfig := exec.Command("kubectl", "delete", "--namespace", namespace, objectName)
	if err := deleteTsdbConfig.Run(); err != nil {
		return fmt.Errorf("failed to delete: %q. error: %s", objectName, err.Error())
	}

	return nil
}
