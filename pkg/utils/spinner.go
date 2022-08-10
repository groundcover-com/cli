package utils

import (
	"fmt"
	"time"

	_spinner "github.com/briandowns/spinner"
)

var ErrSpinnerTimeout = fmt.Errorf("spinner timeout")

type Spinner struct {
	*_spinner.Spinner
}

func NewSpinner(charset int, prefix string) *Spinner {
	spinner := new(Spinner)
	spinner.Spinner = _spinner.New(_spinner.CharSets[charset], time.Millisecond*500)
	spinner.Prefix = prefix
	spinner.Color("green")

	return spinner
}

func (spinner *Spinner) Poll(function func() (bool, error), interval, duration time.Duration) error {
	timeout := time.After(duration)
	ticker := time.NewTicker(interval)

	spinner.Start()
	defer spinner.Stop()

	for {
		select {
		case <-timeout:
			return ErrSpinnerTimeout
		case <-ticker.C:
			functionDone, err := function()
			if err != nil {
				return err
			}
			if functionDone {
				return nil
			}
		}
	}
}
