package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecuteCommand(command string, args ...string) (string, error) {
	output := strings.Builder{}
	cmd := exec.Command(command, args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run command: %q helm chart. exit code: %v otuput: %q", command, err.Error(), output.String())
	}

	return output.String(), nil
}
