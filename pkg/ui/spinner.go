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

var ErrSpinnerTimeout = errors.New("spinner timeout")

type retryableError struct {
	error
}

func RetryableError(err error) error {
	return &retryableError{err}
}

type Spinner struct {
	*yacspin.Spinner
	writer       *Writer
	stopFailChar string
	stopFailMsg  string
	stopChar     string
	stopMsg      string
	wroteError   bool
}

func newSpinner(writer *Writer, message string) *Spinner {
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

	spinner := Spinner{
		Spinner:      s,
		writer:       writer,
		stopFailChar: statusErr,
		stopChar:     statusOK,
	}
	return &spinner
}

func (s *Spinner) SetWarningSign() {
	s.stopFailChar = statusWarning
	s.StopFailCharacter(statusWarning)
	s.StopFailColors("fgYellow")
}

func (s *Spinner) WriteMessage(message string) {
	s.writer.Writeln(message)
	s.Message(message)
}

func (s *Spinner) SetStopMessage(message string) {
	s.stopMsg = message
	s.StopMessage(message)
}

func (s *Spinner) WriteStop() {
	if s.wroteError {
		return
	}

	s.writer.Writeln(fmt.Sprintf("%v %v", s.stopChar, s.stopMsg))
	s.Stop()
}

func (s *Spinner) SetStopFailMessage(message string) {
	s.stopFailMsg = message
	s.StopFailMessage(message)
}

func (s *Spinner) WriteStopFail() {
	s.wroteError = true
	s.writer.Writeln(fmt.Sprintf("%v %v", s.stopFailChar, s.stopFailMsg))
	s.StopFail()
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

			var retryableErr *retryableError
			if !errors.As(err, &retryableErr) {
				return err
			}

			if attempts >= maxRetries {
				return retryableErr
			}

			attempts++
		}
	}
}
