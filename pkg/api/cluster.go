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

func (client *Client) PollIsClusterExist(ctx context.Context, clusterName string) error {
	var err error

	spinner := ui.SingletonWriter.NewSpinner("Waiting until groundcover is connected to cloud platform")
	spinner.StopMessage("groundcover is connected to cloud platform")
	spinner.StopFailMessage("groundcover is yet connected to cloud platform")

	spinner.Start()
	defer spinner.Stop()

	isClusterExistInSassFunc := func() error {
		var clusterList map[string]interface{}
		if clusterList, err = client.ClusterList(); err != nil {
			return err
		}

		for _clusterName := range clusterList {
			if _clusterName == clusterName {
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

	spinner.StopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return errors.New("timeout waiting for groundcover to connect cloud platform")
	}

	return err
}

func (client *Client) ClusterList() (map[string]interface{}, error) {
	var err error

	var body []byte
	if body, err = client.Get(CLUSTER_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var clusterList map[string]interface{}
	if err = json.Unmarshal(body, &clusterList); err != nil {
		return nil, err
	}

	return clusterList, nil
}
