package main

import (
	"os"

	"github.com/sirupsen/logrus"
	cmd "groundcover.com/cmd/groundcover"
	sentry "groundcover.com/pkg/custom_sentry"
)

const (
	SENTRY_PROD_DSN = "https://a8ac7024755f47e5b5d4ae620499c7f6@o1295881.ingest.sentry.io/6521983"
)

func main() {
	sentryDsn := SENTRY_PROD_DSN
	if cmd.IsDevVersion() {
		sentryDsn = ""
	}

	err := sentry.Init(sentryDsn)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	defer sentry.Flush()
	err = cmd.Execute()
	if err != nil {
		sentry.CaptureException(err)
		logrus.Error(err)
	}
}
