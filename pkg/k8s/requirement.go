package k8s

import (
	"strings"

	"github.com/fatih/color"
	"groundcover.com/pkg/ui"
)

type Requirement struct {
	IsCompatible  bool
	Message       string   `json:"-"`
	ErrorMessages []string `json:"-"`
}

func (requirement Requirement) PrintStatus() {
	var messageBuffer strings.Builder

	messageBuffer.WriteString(requirement.Message)
	messageBuffer.WriteString("\n")

	for _, errorMessage := range requirement.ErrorMessages {
		messageBuffer.WriteString(color.RedString(ui.Bullet))
		messageBuffer.WriteString(" ")
		messageBuffer.WriteString(errorMessage)
		messageBuffer.WriteString("\n")
	}

	ui.PrintStatus(requirement.IsCompatible, messageBuffer.String())
}
