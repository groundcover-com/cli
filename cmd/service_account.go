package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

var serviceAccountTokenCmd = &cobra.Command{
	Use:   "generate-service-account-token",
	Short: "Generate Grafana service account token",
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

		var saToken *auth.SAToken
		if saToken, err = fetchServiceAccountToken(tenantUUID); err != nil {
			return err
		}

		ui.QuietWriter.Println(saToken.Token)

		return nil
	},
}

func fetchServiceAccountToken(tenantUUID string) (*auth.SAToken, error) {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = auth.LoadAuth0Token(); err != nil {
		return nil, err
	}

	apiClient := api.NewClient(auth0Token)

	var saToken *auth.SAToken
	if saToken, err = apiClient.ServiceAccountToken(tenantUUID); err != nil {
		return nil, err
	}

	return saToken, nil
}

func init() {
	AuthCmd.AddCommand(serviceAccountTokenCmd)
}
