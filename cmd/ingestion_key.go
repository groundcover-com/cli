package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/ui"
)

// CustomTransport is defined in pkg/client/client.go

var IngestionKeyCmd = &cobra.Command{
	Use:       "get-ingestion-key",
	Short:     "get-ingestion-key",
	ValidArgs: []string{"sensor", "rum", "thirdParty"},
	Example:   "groundcover get-ingestion-key [sensor|rum|thirdParty] [optional-name]",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var tenant *api.TenantInfo
		if tenant, err = fetchTenant(); err != nil {
			return err
		}

		var auth0Token *auth.Auth0Token
		if auth0Token, err = auth.LoadAuth0Token(); err != nil {
			return err
		}

		var backendName string
		if backendName, _, err = selectBackendName(tenant.UUID, false); err != nil {
			return err
		}

		// Create API client
		apiClient := api.NewClient(auth0Token)

		// Get or create ingestion key
		var ingestionKeyType string
		var customName string

		if len(args) == 0 {
			return fmt.Errorf("ingestion key type is required")
		}
		ingestionKeyType = args[0]

		if len(args) > 1 {
			customName = args[1]
		}

		ingestionKey, err := apiClient.GetOrCreateIngestionKey(tenant.UUID, backendName, ingestionKeyType, customName)
		if err != nil {
			return err
		}

		ui.QuietWriter.Println(ingestionKey)

		return nil
	},
}

func init() {
	AuthCmd.AddCommand(IngestionKeyCmd)
}
