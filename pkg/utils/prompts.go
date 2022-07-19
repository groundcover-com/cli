package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const ASSUME_YES_FLAG = "yes"

// This file has components that interact with the user via prompts.

// Prompter proves a user input dialog.
type Prompter struct {
	message string
	choices []string
	dv      string
}

// NewPrompter creates a Prompter with the given message, choices and defaults.
func NewPrompter(message string, choices []string, defaultValue string) *Prompter {
	return &Prompter{
		message: message,
		choices: choices,
		dv:      defaultValue,
	}
}

// Prompt prompts the user and return the value. If the config parameter "y" is set we will return the default.
func (p *Prompter) Prompt() string {
	if p.skip() {
		return p.dv
	}
	fmt.Print(p.msg())
	input := ""
	s := bufio.NewScanner(os.Stdin)
	ok := s.Scan()
	if ok {
		input = strings.TrimRight(s.Text(), "\r\n")
	}
	if input == "" {
		return p.dv
	}
	if !p.validInput(input) {
		fmt.Println(p.errorMsg())
		return p.Prompt()
	}
	return input
}

func (p *Prompter) validInput(s string) bool {
	// Compare values ignoring case.
	ls := strings.ToLower(s)
	for i := range p.choices {
		if ls == strings.ToLower(p.choices[i]) {
			return true
		}
	}
	return false
}

func (p *Prompter) msg() string {
	defaultValue := ""
	if p.dv != "" {
		defaultValue = fmt.Sprintf("[%s] ", p.dv)
	}
	return fmt.Sprintf("%s (%s) %s: ", p.message, strings.Join(p.choices, "/"), defaultValue)
}

func (p *Prompter) errorMsg() string {
	return fmt.Sprintf("Invalid input, must be one of: [%s]", strings.Join(p.choices, ", "))
}

func (p *Prompter) skip() bool {
	return viper.GetBool("y")
}

// YesNoPrompt is a helper function that prompts the user for a Y/N response.
func YesNoPrompt(message string, defaultValue bool) bool {
	if viper.GetBool(ASSUME_YES_FLAG) {
		return true
	}
	defaultChoice := "n"
	if defaultValue {
		defaultChoice = "y"
	}
	return strings.ToLower(NewPrompter(message, []string{"y", "n"}, defaultChoice).Prompt()) == "y"
}
