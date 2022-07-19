package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

func ExecuteCommand(command string, args ...string) (string, error) {
	output := strings.Builder{}
	cmd := exec.Command(command, args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(output.String())
	}

	return output.String(), nil
}

func OpenBrowser(url string) {
	if err := open.Run(url); err != nil {
		fmt.Printf("You can browse to: %s", url)
	}
}
