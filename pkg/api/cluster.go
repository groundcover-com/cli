package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	SOURCES_LIST_ENDPOINT    = "sources/list"
	CLUSTER_POLLING_RETRIES  = 18
	CLUSTER_POLLING_TIMEOUT  = time.Minute * 3
	CLUSTER_POLLING_INTERVAL = time.Second * 10
)

type Conditions struct {
	Conditions []interface{} `json:"conditions"`
}

type SourceList struct {
	Env     []string `json:"env"`
	Cluster []string `json:"cluster"`
}

func (client *Client) PollIsClusterExist(ctx context.Context, tenantUUID, backendName, clusterName string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Waiting until groundcover is connected to cloud platform")
	spinner.SetStopMessage("groundcover is connected to cloud platform")
	spinner.SetStopFailMessage("groundcover is yet connected to cloud platform")

	spinner.Start()
	defer spinner.WriteStop()

	isClusterExistInSassFunc := func() error {
		var backendList []BackendInfo
		if backendList, err = client.BackendsList(tenantUUID); err != nil {
			return err
		}

		for _, backend := range backendList {
			if backend.Name == backendName && backend.Online {
				return nil
			}

			if backend.InCloud {
				var clusterNames []string
				if clusterNames, err = client.clusterList(tenantUUID, backend.Name); err != nil {
					return err
				}

				for _, name := range clusterNames {
					if name == clusterName {
						return nil
					}
				}
			}
		}

		return ui.RetryableError(err)
	}

	if err = spinner.Poll(ctx, isClusterExistInSassFunc, CLUSTER_POLLING_INTERVAL, CLUSTER_POLLING_TIMEOUT, CLUSTER_POLLING_RETRIES); err == nil {
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

func (client *Client) clusterList(tenantUUID, backendName string) ([]string, error) {
	var err error

	var url *url.URL
	if url, err = client.JoinPath(SOURCES_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var payload []byte
	if payload, err = json.Marshal(Conditions{}); err != nil {
		return nil, err
	}

	var request *http.Request
	if request, err = http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(payload)); err != nil {
		return nil, err
	}

	request.Header.Add(TenantUUIDHeader, tenantUUID)
	request.Header.Add(BackendIDHeader, backendName)
	request.Header.Add(ContentTypeHeader, ContentTypeJSON)

	var body []byte
	if body, err = client.do(request); err != nil {
		return nil, err
	}

	var sourceList SourceList
	if err = json.Unmarshal(body, &sourceList); err != nil {
		return nil, err
	}

	return sourceList.Cluster, nil
}
