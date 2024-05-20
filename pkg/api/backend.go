package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	BACKEND_LIST_ENDPOINT    = "backends/list"
	BACKEND_POLLING_RETRIES  = 18
	BACKEND_POLLING_TIMEOUT  = time.Minute * 3
	BACKEND_POLLING_INTERVAL = time.Second * 10
)

type BackendInfo struct {
	Name     string `json:"name"`
	Online   bool   `json:"online"`
	Licensed bool   `json:"licensed"`
	Status   string `json:"status"`
}

func (client *Client) PollIsBackendExist(ctx context.Context, tenantUUID, backendName string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Waiting until groundcover is connected to cloud platform")
	spinner.SetStopMessage("groundcover is connected to cloud platform")
	spinner.SetStopFailMessage("groundcover is yet connected to cloud platform")

	spinner.Start()
	defer spinner.WriteStop()

	isBackendExistInSassFunc := func() error {
		var backendsList []BackendInfo
		if backendsList, err = client.BackendsList(tenantUUID); err != nil {
			return err
		}

		for _, backend := range backendsList {
			if backend.Name == backendName && backend.Online {
				return nil
			}
		}
		return ui.RetryableError(err)
	}

	if err = spinner.Poll(ctx, isBackendExistInSassFunc, BACKEND_POLLING_INTERVAL, BACKEND_POLLING_TIMEOUT, BACKEND_POLLING_RETRIES); err == nil {
		return nil
	}

	if err == nil {
		return nil
	}

	spinner.WriteStopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return errors.New("timeout waiting for groundcover to connect cloud platform")
	}

	return err
}

func (client *Client) BackendsList(tenantUUID string) ([]BackendInfo, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(BACKEND_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodGet, url.String(), nil); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenantUUID)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	var backendList []BackendInfo
	if err = json.Unmarshal(body, &backendList); err != nil {
		return nil, err
	}

	return backendList, nil
}
