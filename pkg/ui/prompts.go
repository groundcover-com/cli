package ui

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

const ASSUME_YES_FLAG = "yes"

func YesNoPrompt(message string, defaultValue bool) bool {
	if viper.GetBool(ASSUME_YES_FLAG) {
		return true
	}

	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}

	var answer bool
	survey.AskOne(prompt, &answer)

	return answer
}

func MultiSelectPrompt(message string, options, defaults []string) []string {
	prompt := &survey.MultiSelect{
		Options: options,
		Default: defaults,
		Message: message,
	}

	var response []string
	survey.AskOne(prompt, &response)

	return response
}
