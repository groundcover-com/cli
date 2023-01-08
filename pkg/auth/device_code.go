package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	DEVICE_CODE_ENDPOINT            = "device/code"
	DEVICE_CODE_POLLING_RETIRES     = 10
	DEVICE_CODE_POLLING_TIMEOUT     = time.Minute * 1
	DEVICE_CODE_POLLING_INTERVAL    = time.Second * 7
	AUTH0_ACCOUNT_NOT_INVITED_ERROR = "access_denied: User has yet to receive an invitation."
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

func (deviceCode *DeviceCode) PollToken(ctx context.Context, auth0Token *Auth0Token) error {
	var err error

	spinnerMessage := fmt.Sprintf("Waiting for device confirmation for: %s", deviceCode.UserCode)
	spinner := ui.SingletonWriter.NewSpinner(spinnerMessage)
	spinner.StopMessage("Device authentication confirmed by auth0")
	spinner.StopFailMessage("Device authentication failed")

	spinner.Start()
	defer spinner.Stop()

	data := url.Values{}
	data.Set("client_id", DefaultClient.ClientId)
	data.Set("device_code", deviceCode.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	fetchTokenFunc := func() error {
		err = auth0Token.Fetch(data)
		if err == nil {
			return nil
		}

		var auth0Err *Auth0Error
		if errors.As(err, &auth0Err) {
			if auth0Err.Type == "authorization_pending" {
				return ui.RetryableError(err)
			}
		}

		return err
	}

	err = spinner.Poll(ctx, fetchTokenFunc, DEVICE_CODE_POLLING_INTERVAL, DEVICE_CODE_POLLING_TIMEOUT, DEVICE_CODE_POLLING_RETIRES)

	if err == nil {
		return nil
	}

	spinner.StopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return fmt.Errorf("timed out while waiting for your login in browser")
	}

	if err.Error() == AUTH0_ACCOUNT_NOT_INVITED_ERROR {
		return errors.New("sorry, we don't support private emails, please try again with your company email")
	}

	return err
}
