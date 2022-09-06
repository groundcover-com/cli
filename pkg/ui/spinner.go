package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/theckman/yacspin"
)

const (
	statusOK       = "\u2714"
	statusErr      = "\u2715"
	spinnerCharset = 11
)

var ErrSpinnerTimeout = fmt.Errorf("spinner timeout")

type Spinner struct {
	*yacspin.Spinner
}

func NewSpinner(message string) *Spinner {
	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		Colors:            []string{"fgGreen"},
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

func (d *Spinner) Decor(success bool) string {
	if !success {
		return color.RedString(statusErr)
	}

	return color.GreenString(statusOK)
}

func (spinner *Spinner) Poll(function func() (bool, error), interval, duration time.Duration) error {
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
