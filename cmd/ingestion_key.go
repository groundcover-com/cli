package cmd

import (
	"context"

	ingestionKeysClient "github.com/groundcover-com/groundcover-sdk-go/pkg/client/ingestionkeys"
	"github.com/groundcover-com/groundcover-sdk-go/pkg/models"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/client"
	"groundcover.com/pkg/ui"

	sdkClient "github.com/groundcover-com/groundcover-sdk-go/pkg/client"
)

const (
	CLI_INGESTION_KEY_NAME = "cli-generated-ingestion-key"
	CLI_INGESTION_KEY_TYPE = "sensor"
)

// CustomTransport is defined in pkg/client/client.go

var IngestionKeyCmd = &cobra.Command{
	Use:   "get-ingestion-key",
	Short: "get-ingestion-key",
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
		if backendName, err = selectBackendName(tenant); err != nil {
			return err
		}

		// Create SDK client factory and client with tenant UUID
		sdkClient, err := client.NewDefaultClient(auth0Token.AccessToken, backendName, tenant.UUID)
		if err != nil {
			return err
		}

		// Get or create ingestion key
		ingestionKey, err := getOrCreateIngestionKey(sdkClient)
		if err != nil {
			return err
		}

		ui.QuietWriter.Println(ingestionKey)

		return nil
	},
}

// getOrCreateIngestionKey retrieves an existing CLI ingestion key or creates a new one
func getOrCreateIngestionKey(sdkClient *sdkClient.GroundcoverAPI) (string, error) {
	// Check if ingestion key already exists
	listParams := ingestionKeysClient.NewListIngestionKeysParamsWithContext(context.Background()).WithBody(&models.ListIngestionKeysRequest{
		Name: CLI_INGESTION_KEY_NAME,
	})
	existingKeys, err := sdkClient.Ingestionkeys.ListIngestionKeys(listParams, nil)
	if err != nil {
		return "", err
	}

	// Look for existing CLI ingestion key
	for _, key := range existingKeys.Payload {
		if key.Name == CLI_INGESTION_KEY_NAME {
			return key.Key, nil
		}
	}

	// Create new ingestion key if none exists
	ingestionKeyName := CLI_INGESTION_KEY_NAME
	ingestionKeyType := CLI_INGESTION_KEY_TYPE
	ingestionKeyReq := models.CreateIngestionKeyRequest{
		Name: &ingestionKeyName,
		Type: &ingestionKeyType,
	}

	ingestionCreateParams := ingestionKeysClient.NewCreateIngestionKeyParamsWithContext(context.Background()).WithBody(&ingestionKeyReq)
	keyRes, err := sdkClient.Ingestionkeys.CreateIngestionKey(ingestionCreateParams, nil)
	if err != nil {
		return "", err
	}

	return keyRes.Payload.Key, nil
}

func init() {
	AuthCmd.AddCommand(IngestionKeyCmd)
}
