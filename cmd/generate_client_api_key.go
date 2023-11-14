package cmd

import (
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var generateClientAPIKeyCmd = &cobra.Command{
	Use:   "generate-client-api-key",
	Short: "Get Client Token for Grafana API",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var tenant *api.TenantInfo
		if tenant, err = fetchTenant(); err != nil {
			return err
		}

		var apiToken *auth.ApiKey
		if apiToken, err = fetchClientAPIKey(tenant); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiToken.ApiKey)

		return nil
	},
}

func fetchClientAPIKey(tenant *api.TenantInfo) (*auth.ApiKey, error) {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return nil, err
	}

	apiClient := api.NewClient(auth0Token)

	var apiToken *auth.ApiKey
	if apiToken, err = apiClient.GetOrCreateClientToken(tenant); err != nil {
		return nil, err
	}

	return apiToken, nil
}

func init() {
	AuthCmd.AddCommand(generateClientAPIKeyCmd)
}
