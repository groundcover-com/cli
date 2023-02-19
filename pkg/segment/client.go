package segment

import (
	"time"

	"github.com/segmentio/analytics-go/v3"
)

var client analytics.Client

const (
	WRITE_KEY = "FPPzr8mdiYq9Ry2YOEVFN751DvSdwwUZ"
)

func Init(appName, version string) error {
	var err error

	config := analytics.Config{
		BatchSize: 1,
		Interval:  5 * time.Second,

		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    appName,
				Version: version,
			},
		},
	}

	if client, err = analytics.NewWithConfig(WRITE_KEY, config); err != nil {
		return err
	}

	return nil
}

func Close() error {
	return client.Close()
}
