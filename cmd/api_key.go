package cmd

import (
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenant *api.TenantInfo
		if tenant, err = fetchTenant(); err != nil {
			return err
		}

		var apiKey *auth.ApiKey
		if apiKey, err = fetchApiKey(tenant.UUID); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiKey.ApiKey)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(apiKeyCmd)
}
