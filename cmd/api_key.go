package cmd

import (
	"github.com/spf13/cobra"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenantUUID string
		if tenantUUID, _, err = fetchTenantOrUseFlag(); err != nil {
			return err
		}

		var apiKey *auth.ApiKey
		if apiKey, err = fetchApiKey(tenantUUID); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiKey.ApiKey)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(apiKeyCmd)
}
