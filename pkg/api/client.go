package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	ingestionKeysClient "github.com/groundcover-com/groundcover-sdk-go/pkg/client/ingestionkeys"
	"github.com/groundcover-com/groundcover-sdk-go/pkg/models"
	"groundcover.com/pkg/auth"
	clientpkg "groundcover.com/pkg/client"
)

const CLI_INGESTION_KEY_NAME = "cli-generated-ingestion-key-%s"

type TransportWithAuth0Token struct {
	http.RoundTripper
	auth0Token *auth.Auth0Token
}

func (transport *TransportWithAuth0Token) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error

	var bearerToken string
	if bearerToken, err = transport.auth0Token.BearerToken(); err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", bearerToken)
	return transport.RoundTripper.RoundTrip(request)
}

type Client struct {
	baseUrl    *url.URL
	httpClient *http.Client
}

func NewClient(auth0Token *auth.Auth0Token) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &TransportWithAuth0Token{
				auth0Token:   auth0Token,
				RoundTripper: http.DefaultTransport,
			},
		},
		baseUrl: &url.URL{
			Scheme: "https",
			Path:   "/api/",
			Host:   "app.groundcover.com",
		},
	}
}

func (client *Client) ApiKey(tenantUUID string) (*auth.ApiKey, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(auth.GenerateAPIKeyEndpoint); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodPost, url.String(), nil); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenantUUID)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	apiKey := &auth.ApiKey{}
	if err = apiKey.ParseBody(body); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (client *Client) ServiceAccountToken(tenantUUID string) (*auth.SAToken, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(auth.GENERATE_SERVICE_ACCOUNT_TOKEN_ENDPOINT); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodPost, url.String(), nil); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenantUUID)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	saToken := &auth.SAToken{}
	if err = saToken.ParseBody(body); err != nil {
		return nil, err
	}

	return saToken, nil
}
func (client *Client) GetDatasourcesAPIKey(tenant *TenantInfo, backendName string) (*auth.ApiKey, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(auth.GetDatasourcesAPIKeyEndpoint); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodPost, url.String(), nil); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenant.UUID)
	request.Header.Add(BackendIDHeader, backendName)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	key := &auth.ApiKey{}
	if err = key.ParseBody(body); err != nil {
		return nil, err
	}

	return key, nil
}

func (client *Client) GetOrCreateClientToken(tenant *TenantInfo) (*auth.ApiKey, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(auth.GenerateClientTokenAPIKeyEndpoint); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodPost, url.String(), nil); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenant.UUID)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	clientToken := &auth.ApiKey{}
	if err = clientToken.ParseBody(body); err != nil {
		return nil, err
	}

	return clientToken, nil
}

func (client *Client) JoinPath(endpoint string) (*url.URL, error) {
	return client.baseUrl.Parse(endpoint)
}

func (client *Client) do(request *http.Request) ([]byte, error) {
	var err error
	var response *http.Response

	if response, err = client.httpClient.Do(request); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, NewResponseError(response)
	}

	return io.ReadAll(io.Reader(response.Body))
}

func (client *Client) get(endpoint string) ([]byte, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(endpoint); err != nil {
		return nil, err
	}

	var response *http.Response
	if response, err = client.httpClient.Get(url.String()); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, NewResponseError(response)
	}

	return io.ReadAll(io.Reader(response.Body))
}

// GetOrCreateIngestionKey retrieves an existing CLI ingestion key or creates a new one
func (client *Client) GetOrCreateIngestionKey(tenantUUID, backendName, ingestionKeyType, customName string) (string, error) {
	// Create SDK client
	auth0Token, err := auth.LoadAuth0Token()
	if err != nil {
		return "", err
	}

	sdkClient, err := clientpkg.NewDefaultClient(auth0Token.AccessToken, backendName, tenantUUID)
	if err != nil {
		return "", err
	}

	// Use provided name if available, otherwise use default naming pattern
	var ingestionKeyName string
	if customName != "" {
		ingestionKeyName = customName
	} else {
		ingestionKeyName = strings.ToLower(fmt.Sprintf(CLI_INGESTION_KEY_NAME, ingestionKeyType))
	}

	listParams := ingestionKeysClient.NewListIngestionKeysParamsWithContext(context.Background()).WithBody(&models.ListIngestionKeysRequest{
		Name: ingestionKeyName,
	})

	existingKeys, err := sdkClient.Ingestionkeys.ListIngestionKeys(listParams, nil)
	if err != nil {
		return "", err
	}

	// Look for existing CLI ingestion key
	for _, key := range existingKeys.Payload {
		if key.Name == ingestionKeyName {
			return key.Key, nil
		}
	}

	// Create new ingestion key if none exists
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
