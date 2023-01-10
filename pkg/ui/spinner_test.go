package ui_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/ui"
)

type SpinnerTestSuite struct {
	suite.Suite
}

func (suite *SpinnerTestSuite) SetupSuite() {}

func (suite *SpinnerTestSuite) TearDownSuite() {}

func TestSpinnerSuite(t *testing.T) {
	suite.Run(t, &SpinnerTestSuite{})
}

func (suite *SpinnerTestSuite) TestPollFuncSuccues() {
	//prepare
	ctx := context.Background()
	spinner := ui.NewWriter().NewSpinner("test")

	//act
	testFunc := func() error {
		return nil
	}

	err := spinner.Poll(ctx, testFunc, time.Millisecond, time.Second, 0)

	// assert
	suite.NoError(err)
}

func (suite *SpinnerTestSuite) TestPollFuncMaxRetries() {
	//prepare
	ctx := context.Background()
	spinner := ui.NewWriter().NewSpinner("test")
	myError := errors.New("test")

	//act
	var attempts int
	testFunc := func() error {
		attempts++
		return ui.RetryableError(myError)
	}

	err := spinner.Poll(ctx, testFunc, time.Millisecond, time.Second, 1)

	// assert
	suite.Equal(2, attempts)
	suite.ErrorContains(err, "test")
}

func (suite *SpinnerTestSuite) TestPollFuncTimeout() {
	//prepare
	ctx := context.Background()
	myError := errors.New("test")
	spinner := ui.NewWriter().NewSpinner("test")

	//act
	testFunc := func() error {
		time.Sleep(time.Millisecond * 500)
		return ui.RetryableError(myError)
	}

	err := spinner.Poll(ctx, testFunc, time.Millisecond, time.Second, 100)

	// assert
	suite.ErrorIs(err, ui.ErrSpinnerTimeout)
}

func (suite *SpinnerTestSuite) TestPollFuncNonRetryableError() {
	//prepare
	ctx := context.Background()
	myError := errors.New("test")
	spinner := ui.NewWriter().NewSpinner("test")

	//act
	var attempts int
	testFunc := func() error {
		attempts++
		return myError
	}

	err := spinner.Poll(ctx, testFunc, time.Millisecond, time.Second, 100)

	// assert
	suite.Equal(1, attempts)
	suite.ErrorIs(err, myError)
}
