package utils

import (
	"fmt"
	"time"

	_spinner "github.com/briandowns/spinner"
)

type SpinnerTimeoutError struct {
	error
}

func NewSpinnerTimeoutError(duration time.Duration) SpinnerTimeoutError {
	return SpinnerTimeoutError{
		fmt.Errorf("spinner timeout after %s", duration.String()),
	}
}

type Spinner struct {
	*_spinner.Spinner
}

func NewSpinner(charset int, prefix string) *Spinner {
	spinner := new(Spinner)
	spinner.Spinner = _spinner.New(_spinner.CharSets[charset], 100*time.Millisecond)
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
			return NewSpinnerTimeoutError(duration)
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
