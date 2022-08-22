package auth

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

var DefaultClient *Client = &Client{
	httpClient: http.DefaultClient,
	Audience:   "https://groundcover",
	Scope:      "access:router offline_access",
	ClientId:   "UkQmsxoqC8OzajqptiADtAZD6GS2mG9U",
	baseUrl: &url.URL{
		Scheme: "https",
		Path:   "/oauth/",
		Host:   "auth.groundcover.com",
	},
}

type Client struct {
	Scope      string
	Audience   string
	ClientId   string
	baseUrl    *url.URL
	httpClient *http.Client
}

func (client *Client) JoinPath(endpoint string) (*url.URL, error) {
	return client.baseUrl.Parse(endpoint)
}

func (client *Client) PostForm(endpoint string, data url.Values) ([]byte, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(endpoint); err != nil {
		return nil, err
	}

	var response *http.Response
	if response, err = client.httpClient.PostForm(url.String(), data); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, NewAuth0Error(body)
	}

	return body, nil
}
