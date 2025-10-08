package cmd

import (
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var getDatasourcesAPIKeyCmd = &cobra.Command{
	Use:   "get-datasources-api-key",
	Short: "Get the API key for datasources",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenantUUID string
		var tenant *api.TenantInfo
		if tenantUUID, tenant, err = fetchTenantOrUseFlag(); err != nil {
			return err
		}

		var backendName string
		if backendName, _, err = selectBackendName(tenantUUID, false); err != nil {
			return err
		}

		var apiToken *auth.ApiKey
		if apiToken, err = fetchDatasourcesAPIKey(tenant, backendName); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiToken.ApiKey)
		return nil
	},
}

func fetchDatasourcesAPIKey(tenant *api.TenantInfo, backendName string) (*auth.ApiKey, error) {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return nil, err
	}

	apiClient := api.NewClient(auth0Token)

	var apiToken *auth.ApiKey
	if apiToken, err = apiClient.GetDatasourcesAPIKey(tenant, backendName); err != nil {
		return nil, err
	}

	return apiToken, nil
}

func init() {
	AuthCmd.AddCommand(getDatasourcesAPIKeyCmd)
}
