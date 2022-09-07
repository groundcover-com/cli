package ui

import (
	"fmt"
	"time"

	"github.com/theckman/yacspin"
)

const (
	statusOK       = "\u2714"
	statusErr      = "\u2715"
	statusWarning  = "\u270B"
	spinnerCharset = 11
)

var ErrSpinnerTimeout = fmt.Errorf("spinner timeout")

type Spinner struct {
	*yacspin.Spinner
}

func NewSpinner(message string) *Spinner {
	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		Colors:            []string{"fgBlue"},
		CharSet:           yacspin.CharSets[spinnerCharset],
		SuffixAutoColon:   true,
		Suffix:            " ",
		Message:           message,
		StopCharacter:     statusOK,
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: statusErr,
		StopFailColors:    []string{"fgRed"},
	}

	s, _ := yacspin.New(cfg)

	spinner := new(Spinner)
	spinner.Spinner = s

	return spinner
}

func (s *Spinner) SetWarningSign() {
	s.StopFailCharacter(statusWarning)
	s.StopFailColors("fgYellow")
}

func (s *Spinner) Poll(function func() (bool, error), interval, duration time.Duration) error {
	timeout := time.After(duration)
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-timeout:
			return ErrSpinnerTimeout
		case <-ticker.C:
			success, err := function()
			if err != nil {
				return err
			}
			if success {
				return nil
			}
		}
	}
}
