package api

import (
	"io"
	"net/http"
	"net/url"

	"groundcover.com/pkg/auth"
)

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

func (client *Client) GetDatasourcesAPIKey(tenant *TenantInfo) (*auth.ApiKey, error) {
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
