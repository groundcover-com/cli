package sentry

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	TAINTED_TAG                  = "tainted"
	ERASE_DATA_TAG               = "erase"
	UPGRADE_TAG                  = "upgrade"
	TOKEN_ID_TAG                 = "token.id"
	ORGANIZATION_TAG             = "organization"
	CHART_VERSION_TAG            = "chart.version"
	DEFAULT_RESOURCES_PRESET_TAG = "resources.presets.default"
	PERSISTENT_STORAGE_TAG       = "storage.persistent"
	CLUSTER_NAME_TAG             = "cluster.name"
	NODES_COUNT_TAG              = "nodes.count"
	EXPECTED_NODES_COUNT_TAG     = "nodes.expected_count"
	RUNNING_ALLIGATORS_TAG       = "nodes.running_alligators"
	FLUSH_TIMEOUT                = time.Second * 2
)

var Dsn string = "https://6420be38b4544852a61df1d7ec56f442@o1295881.ingest.sentry.io/6521982"

func GetSentryClientOptions(appName, environment, version string) sentry.ClientOptions {
	return sentry.ClientOptions{
		MaxBreadcrumbs: 10,
		Dsn:            Dsn,
		Environment:    environment,
		Release:        fmt.Sprintf("%s@%s", appName, version),
	}
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
