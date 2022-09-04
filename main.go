package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"groundcover.com/cmd"
	sentry_utils "groundcover.com/pkg/sentry"
)

func main() {
	var err error

	logrus.SetFormatter(&logrus.TextFormatter{
		PadLevelText:     true,
		DisableTimestamp: true,
	})

	environment := "prod"
	release := fmt.Sprintf("cli@%s", cmd.BinaryVersion)

	if cmd.IsDevVersion() {
		environment = "dev"
	}

	sentryClientOptions := sentry_utils.GetSentryClientOptions(environment, release)
	if err = sentry.Init(sentryClientOptions); err != nil {
		logrus.Panic(err)
	}
	defer sentry.Flush(sentry_utils.FLUSH_TIMEOUT)

	if err = cmd.Execute(); err != nil {
		sentry.CaptureException(err)
		logrus.Error(err)
	}
}
