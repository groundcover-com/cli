package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	greenStatusOk = color.GreenString(statusOK)
	redStatusErr  = color.RedString(statusErr)
)

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
	fmt.Printf("%s %s", statusWarning, message)
}

func UrlLink(url string) string {
	return color.New(color.FgBlue).Add(color.Underline).Sprint(url)
}
