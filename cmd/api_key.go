package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var apiKey *auth.ApiKey
		if apiKey, err = auth.LoadApiKey(); err != nil {
			return errors.Wrap(err, "failed to load api key")
		}

		ui.QuietWriter.Println(apiKey.ApiKey)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(apiKeyCmd)
}
