package kubectl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
	cmdOutput := strings.Builder{}
	deleteTsdbConfigCmd := exec.Command("kubectl", "delete", "--namespace", namespace, objectName)
	deleteTsdbConfigCmd.Stdout = &cmdOutput
	deleteTsdbConfigCmd.Stderr = &cmdOutput
	if err := deleteTsdbConfigCmd.Run(); err != nil {
		// if the object doesn't exist, we don't care
		if strings.Contains(cmdOutput.String(), "not found") {
			return nil
		}

		return fmt.Errorf("failed to uninstall %q. error: %s", objectName, err.Error())
	}

	return nil
}

func DeletePvcByLabels(ctx context.Context, namespace string, labelsToDelete []string) error {
	for _, label := range labelsToDelete {
		cmdOutput := strings.Builder{}

		deleteLabeledPvcs := exec.Command("kubectl", "delete", "pvc", "--namespace", namespace, "--selector", label)
		deleteLabeledPvcs.Stdout = &cmdOutput
		deleteLabeledPvcs.Stderr = &cmdOutput

		if err := deleteLabeledPvcs.Run(); err != nil {
			// if the object doesn't exist, we don't care
			if strings.Contains(cmdOutput.String(), "No resources found in") {
				return nil
			}

			return fmt.Errorf("failed to delete all of groundcovers pvcs in namespace %q. error: %s", namespace, err.Error())
		}
	}
	return nil
}
