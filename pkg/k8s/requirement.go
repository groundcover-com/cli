package k8s

import (
	"strings"

	"github.com/fatih/color"
	"groundcover.com/pkg/ui"
)

type Requirement struct {
	IsCompatible    bool
	IsNonCompatible bool
	Message         string   `json:"-"`
	ErrorMessages   []string `json:"-"`
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

	switch {
	case requirement.IsCompatible:
		ui.SingletonWriter.PrintSuccessMessage(messageBuffer.String())
	case requirement.IsNonCompatible:
		ui.SingletonWriter.PrintErrorMessage(messageBuffer.String())
	default:
		ui.SingletonWriter.PrintWarningMessage(messageBuffer.String())
	}
}
