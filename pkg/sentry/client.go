package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	TAINTED_TAG                  = "tainted"
	ERASE_DATA_TAG               = "erase"
	UPGRADE_TAG                  = "upgrade"
	ORGANIZATION_TAG             = "organization"
	CHART_VERSION_TAG            = "chart.version"
	FLUSH_TIMEOUT                = time.Second * 2
	DEFAULT_RESOURCES_PRESET_TAG = "resources.presets.default"
	PERSISTENT_STORAGE_TAG       = "storage.persistent"
	PROD_DSN                     = "https://a8ac7024755f47e5b5d4ae620499c7f6@o1295881.ingest.sentry.io/6521983"
	DEV_DSN                      = "https://6420be38b4544852a61df1d7ec56f442@o1295881.ingest.sentry.io/6521982"
)

func GetSentryClientOptions(environment, release string) sentry.ClientOptions {
	clientOptions := sentry.ClientOptions{
		MaxBreadcrumbs: 10,
		Dsn:            PROD_DSN,
		Release:        release,
		Environment:    environment,
	}

	if environment == "dev" {
		clientOptions.Dsn = DEV_DSN
	}

	return clientOptions
}

func SetTagOnCurrentScope(key, value string) {
	sentry.CurrentHub().Scope().SetTag(key, value)
}

func SetUserOnCurrentScope(user sentry.User) {
	sentry.CurrentHub().Scope().SetUser(user)
}

func SetLevelOnCurrentScope(level sentry.Level) {
	sentry.CurrentHub().Scope().SetLevel(level)
}

func SetTransactionOnCurrentScope(name string) {
	sentry.CurrentHub().Scope().SetTransaction(name)
}
