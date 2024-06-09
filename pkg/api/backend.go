package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
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
	InCloud  bool   `json:"inCloud"`
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
