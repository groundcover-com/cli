package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

func (client *Client) ApiKey() (*auth.ApiKey, error) {
	var err error

	var body []byte
	if body, err = client.post(auth.GENERATE_API_KEY_ENDPOINT, "", nil); err != nil {
		return nil, err
	}

	var apiKey *auth.ApiKey
	if err = json.Unmarshal(body, &apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (client *Client) JoinPath(endpoint string) (*url.URL, error) {
	return client.baseUrl.Parse(endpoint)
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

	return ioutil.ReadAll(response.Body)
}

func (client *Client) post(endpoint, contentType string, payload io.Reader) ([]byte, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(endpoint); err != nil {
		return nil, err
	}

	var response *http.Response
	if response, err = client.httpClient.Post(url.String(), contentType, payload); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, NewResponseError(response)
	}

	return ioutil.ReadAll(response.Body)
}
