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
		PrintSuccessMessage(message)
	} else {
		PrintErrorMessage(message)
	}
}

func PrintSuccessMessage(message string) {
	fmt.Printf("%s %s", greenStatusOk, message)
}

func PrintErrorMessage(message string) {
	fmt.Printf("%s %s", redStatusErr, message)
}

func PrintErrorMessageln(message string) {
	fmt.Printf("%s %s\n", redStatusErr, message)
}

func PrintWarningMessage(message string) {
	fmt.Print(color.RedString(message))
}
