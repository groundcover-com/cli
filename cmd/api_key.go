package cmd

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var apiKey api.ApiKey
		if err = apiKey.Load(); err != nil {
			return errors.Wrap(err, "failed to load api key")
		}

		fmt.Println(apiKey.ApiKey)

		sentry.CaptureMessage("print-api-key executed successfully")
		return nil
	},
}

func init() {
	authCmd.AddCommand(apiKeyCmd)
}
