package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var apiKeyCmd = &cobra.Command{
	Use:   "print-api-key",
	Short: "Print api-key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenantUUID string
		var tenant *api.TenantInfo
		if tenantUUID = viper.GetString(TENANT_UUID_FLAG); tenantUUID == "" {
			if tenant, err = fetchTenant(); err != nil {
				return err
			}
			tenantUUID = tenant.UUID
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
