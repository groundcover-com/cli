package custom_sentry

import (
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	SENTRY_FLUSH_INTERVAL = 2 * time.Second
)

func Flush() {
	sentry.Flush(SENTRY_FLUSH_INTERVAL)
}

func CaptureException(err error) {
	sentry.CaptureException(err)
}

func Init(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: true,
		Environment:      runtime.GOOS,
		MaxBreadcrumbs:   10,
		SampleRate:       1.0,
		TracesSampleRate: 1.0,
	})
}
