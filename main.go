package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"groundcover.com/cmd"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/rest"
)

func main() {
	var err error

	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	environment := "prod"
	release := fmt.Sprintf("cli@%s", cmd.BinaryVersion)

	if cmd.IsDevVersion() {
		environment = "dev"
	}

	sentryClientOptions := sentry_utils.GetSentryClientOptions(environment, release)
	if err = sentry.Init(sentryClientOptions); err != nil {
		ui.PrintErrorMessageln(err.Error())
		panic(err)
	}
	defer sentry.Flush(sentry_utils.FLUSH_TIMEOUT)

	if err = cmd.Execute(); err != nil {
		sentry.CaptureException(err)
		ui.PrintErrorMessageln(err.Error())
	}
}
