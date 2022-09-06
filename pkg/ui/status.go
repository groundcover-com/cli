package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	greenStatusOk = color.GreenString(statusOK)
	redStatusErr  = color.RedString(statusErr)
)

func PrintStatus(condition bool, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	if condition {
		PrintSuccess(message)
	} else {
		PrintError(message)
	}
}

func PrintSuccess(message string) {
	fmt.Printf("%s %s", greenStatusOk, message)
}

func PrintError(message string) {
	fmt.Printf("%s %s", redStatusErr, message)
}
