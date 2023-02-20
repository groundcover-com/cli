package segment

import (
	"github.com/segmentio/analytics-go/v3"
)

var (
	client   analytics.Client
	WriteKey string = "FPPzr8mdiYq9Ry2YOEVFN751DvSdwwUZ"
)

func GetConfig(appName, version string) analytics.Config {
	return analytics.Config{
		BatchSize: 1,

		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    appName,
				Version: version,
			},
		},
	}
}

func Init(config analytics.Config) error {
	var err error

	if client, err = analytics.NewWithConfig(WriteKey, config); err != nil {
		return err
	}

	return nil
}

func Close() error {
	return client.Close()
}
