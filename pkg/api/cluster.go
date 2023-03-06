package api

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	CLUSTER_LIST_ENDPOINT    = "cluster/list"
	CLUSTER_POLLING_RETRIES  = 20
	CLUSTER_POLLING_TIMEOUT  = time.Minute * 3
	CLUSTER_POLLING_INTERVAL = time.Second * 10
)

type ClusterInfo struct {
	Name     string `json:"name"`
	Online   bool   `json:"online"`
	Licensed bool   `json:"licensed"`
	Status   string `json:"status"`
}

func (client *Client) PollIsClusterExist(ctx context.Context, clusterName string) error {
	var err error

	spinner := ui.GlobalWriter.NewSpinner("Waiting until groundcover is connected to cloud platform")
	spinner.SetStopMessage("groundcover is connected to cloud platform")
	spinner.SetStopFailMessage("groundcover is yet connected to cloud platform")

	spinner.Start()
	defer spinner.WriteStop()

	isClusterExistInSassFunc := func() error {
		var clusterList []ClusterInfo
		if clusterList, err = client.ClusterList(); err != nil {
			return err
		}

		for _, cluster := range clusterList {
			if cluster.Name == clusterName && cluster.Online {
				return nil
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

func (client *Client) ClusterList() ([]ClusterInfo, error) {
	var err error

	var body []byte
	if body, err = client.get(CLUSTER_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var clusterList []ClusterInfo
	if err = json.Unmarshal(body, &clusterList); err != nil {
		return nil, err
	}

	return clusterList, nil
}
