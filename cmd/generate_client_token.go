package cmd

import (
	"github.com/spf13/cobra"
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
		if tenantUUID, err = getTenantUUID(); err != nil {
			return err
		}

		var apiToken *auth.ApiKey
		if apiToken, err = fetchClientToken(tenantUUID); err != nil {
			return err
		}

		ui.QuietWriter.Println(apiToken.ApiKey)

		return nil
	},
}

func fetchClientToken(tenantUUID string) (*auth.ApiKey, error) {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return nil, err
	}

	apiClient := api.NewClient(auth0Token)

	var clientToken *auth.ApiKey
	if clientToken, err = apiClient.GetOrCreateClientToken(tenantUUID); err != nil {
		return nil, err
	}

	return clientToken, nil
}

func init() {
	AuthCmd.AddCommand(generateClientTokenCmd)
}
