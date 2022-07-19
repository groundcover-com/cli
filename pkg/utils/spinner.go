package utils

import (
	"fmt"
	"time"

	_spinner "github.com/briandowns/spinner"
)

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
	var err error
	var functionDone bool

	timeout := time.After(duration)
	ticker := time.NewTicker(interval)

	spinner.Start()
	defer spinner.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("spinner timeout after %s", duration.String())
		case <-ticker.C:
			functionDone, err = function()
			switch {
			case err != nil:
				return err
			case functionDone:
				return nil
			}
		}
	}
}
