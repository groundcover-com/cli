package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	DEVICE_CODE_ENDPOINT         = "device/code"
	DEVICE_CODE_POLLING_TIMEOUT  = time.Minute * 1
	DEVICE_CODE_POLLING_INTERVAL = time.Second * 7
)

type DeviceCode struct {
	Interval                int    `json:"interval"`
	UserCode                string `json:"user_code"`
	ExpiresIn               int    `json:"expires_in"`
	DeviceCode              string `json:"device_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

func (deviceCode *DeviceCode) Fetch() error {
	var err error

	data := url.Values{}
	data.Set("scope", DefaultClient.Scope)
	data.Set("audience", DefaultClient.Audience)
	data.Set("client_id", DefaultClient.ClientId)

	var body []byte
	if body, err = DefaultClient.PostForm(DEVICE_CODE_ENDPOINT, data); err != nil {
		return err
	}

	return json.Unmarshal(body, &deviceCode)
}

func (deviceCode *DeviceCode) PollToken(auth0Token *Auth0Token) error {
	spinner := ui.NewSpinner("Waiting for device confirmation")
	spinner.StopMessage("Device authentication confirmed, You are successfully logged in!")

	spinner.Start()
	defer spinner.Stop()

	data := url.Values{}
	data.Set("client_id", DefaultClient.ClientId)
	data.Set("device_code", deviceCode.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	err := spinner.Poll(func() (bool, error) {
		return fetchTokenFunc(auth0Token, data)
	},
		DEVICE_CODE_POLLING_INTERVAL,
		DEVICE_CODE_POLLING_TIMEOUT)
	if err == nil {
		return nil
	}

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		spinner.StopFailMessage("timed out while waiting for your login in browser")
		spinner.StopFail()
		return fmt.Errorf("timed out while waiting for your login in browser")
	}

	return err
}

func fetchTokenFunc(auth0Token *Auth0Token, data url.Values) (bool, error) {
	err := auth0Token.Fetch(data)
	if err == nil {
		return true, nil
	}

	var auth0Err *Auth0Error
	if errors.As(err, &auth0Err) {
		if auth0Err.Type == "authorization_pending" {
			return false, nil
		}
	}

	return false, err
}
