package ui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/theckman/yacspin"
)

const (
	statusOK       = "\u2714"
	statusErr      = "\u2715"
	statusWarning  = "\u270B"
	Bullet         = "\u2022"
	spinnerCharset = 11
)

var ErrSpinnerTimeout = fmt.Errorf("spinner timeout")

type retryableError struct {
	error
}

func RetryableError(err error) error {
	return &retryableError{err}
}

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

func (s *Spinner) Poll(ctx context.Context, function func() error, interval, duration time.Duration, maxRetries int) error {
	var attempts int

	timeout := time.After(duration)
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return ErrSpinnerTimeout
		case <-ticker.C:
			err := function()

			if err == nil {
				return nil
			}

			var rerr *retryableError
			if !errors.As(err, &rerr) {
				return err
			}

			if attempts >= maxRetries {
				return rerr
			}

			attempts++
		}
	}
}
