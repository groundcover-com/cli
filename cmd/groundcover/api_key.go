package cmd

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/auth"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var apiKey *auth.ApiKey
		if apiKey, err = auth.LoadApiKey(); err != nil {
			return err
		}

		fmt.Println(apiKey.ApiKey)

		sentry.CaptureMessage("print-api-key executed successfully")
		return nil
	},
}

func init() {
	authCmd.AddCommand(apiKeyCmd)
}
