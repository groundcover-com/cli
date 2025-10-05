package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var generateClientTokenCmd = &cobra.Command{
	Use:   "generate-client-token",
	Short: "Get Client Token for Grafana API",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenantUUID string
		var tenant *api.TenantInfo
		if tenantUUID = viper.GetString(TENANT_UUID_FLAG); tenantUUID == "" {
			if tenant, err = fetchTenant(); err != nil {
				return err
			}
			tenantUUID = tenant.UUID
		} else {
			tenant = &api.TenantInfo{UUID: tenantUUID}
		}

		var apiToken *auth.ApiKey
		if apiToken, err = fetchClientToken(tenant); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiToken.ApiKey)

		return nil
	},
}

func fetchClientToken(tenant *api.TenantInfo) (*auth.ApiKey, error) {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return nil, err
	}

	apiClient := api.NewClient(auth0Token)

	var clientToken *auth.ApiKey
	if clientToken, err = apiClient.GetOrCreateClientToken(tenant); err != nil {
		return nil, err
	}

	return clientToken, nil
}

func init() {
	AuthCmd.AddCommand(generateClientTokenCmd)
}
