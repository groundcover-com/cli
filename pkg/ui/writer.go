package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

const (
	ASSUME_YES_FLAG = "yes"
)

type Writer struct {
	writen []string
}

var (
	greenStatusOk    = color.GreenString(statusOK)
	redStatusErr     = color.RedString(statusErr)
	writenStatusOk   = "V"
	writenStatusErr  = "X"
	writenStatusWarn = "!"
)

func NewWriter() *Writer {
	return &Writer{
		writen: []string{},
	}
}

var SingletonWriter = NewWriter()

func (w *Writer) MarshalJSON() ([]byte, error) {
	return json.Marshal((w.Dump()))
}

func (w *Writer) Println(message string) {
	w.writen = append(w.writen, fmt.Sprintln(message))
	fmt.Println(message)
}

func (w *Writer) Printf(format string, args ...interface{}) {
	formatted := fmt.Sprintf(format, args...)
	w.writen = append(w.writen, formatted)
	fmt.Print(formatted)
}

func (w *Writer) Errorf(format string, args ...interface{}) error {
	formatted := fmt.Sprintf(format, args...)
	w.writen = append(w.writen, formatted)
	return errors.New(formatted)
}

func (w *Writer) PrintSuccessMessage(message string) {
	w.writen = append(w.writen, fmt.Sprintf("%s %s", writenStatusOk, message))
	fmt.Printf("%s %s", greenStatusOk, message)
}

func (w *Writer) PrintErrorMessage(message string) {
	w.writen = append(w.writen, fmt.Sprintf("%s %s", writenStatusErr, message))
	fmt.Printf("%s %s", redStatusErr, message)
}

func (w *Writer) PrintErrorMessageln(message string) {
	w.writen = append(w.writen, fmt.Sprintf("%s %s\n", writenStatusErr, message))
	fmt.Printf("%s %s\n", redStatusErr, message)
}

func (w *Writer) PrintWarningMessage(message string) {
	w.writen = append(w.writen, fmt.Sprintf("%s %s", writenStatusWarn, message))
	fmt.Printf("%s %s", statusWarning, message)
}

func (w *Writer) PrintNoticeMessage(message string) {
	w.writen = append(w.writen, message)
	fmt.Printf("🚨 %s", message)
}

func (w *Writer) UrlLink(url string) string {
	return color.New(color.FgBlue).Add(color.Underline).Sprint(url)
}

func (w *Writer) NewSpinner(message string) *Spinner {
	w.writen = append(w.writen, fmt.Sprintln(message))
	return newSpinner(message)
}

func (w *Writer) SprintfScrub(format string, args ...interface{}) string {
	w.writen = append(w.writen, format)
	return fmt.Sprintf(format, args...)
}

func (w *Writer) YesNoPrompt(message string, defaultValue bool) bool {
	if viper.GetBool(ASSUME_YES_FLAG) {
		return true
	}

	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}

	var answer bool
	survey.AskOne(prompt, &answer)
	w.writen = append(w.writen, fmt.Sprintf("%s %t\n", message, answer))
	return answer
}

func (w *Writer) MultiSelectPrompt(message string, options, defaults []string) []string {
	if viper.GetBool(ASSUME_YES_FLAG) {
		return defaults
	}

	prompt := &survey.MultiSelect{
		Options: options,
		Default: defaults,
		Message: message,
	}

	var response []string
	survey.AskOne(prompt, &response)
	w.writen = append(w.writen, fmt.Sprintf("%s %v\n", message, response))

	return response
}

func (w *Writer) Dump() string {
	return strings.Join(w.writen, "")
}
