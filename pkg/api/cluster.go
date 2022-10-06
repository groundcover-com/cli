package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"groundcover.com/pkg/ui"
)

const (
	CLUSTER_LIST_ENDPOINT    = "cluster/list"
	CLUSTER_POLLING_TIMEOUT  = time.Minute * 3
	CLUSTER_POLLING_INTERVAL = time.Second * 10
)

func (client *Client) PollIsClusterExist(ctx context.Context, clusterName string) error {
	var err error

	spinner := ui.NewSpinner("Waiting until groundcover is connected to cloud platform")
	spinner.StopMessage("groundcover is connected to cloud platform")
	spinner.StopFailMessage("groundcover is yet connected to cloud platform")

	spinner.Start()
	defer spinner.Stop()

	isClusterExistInSassFunc := func() (bool, error) {
		var clusterList map[string]interface{}
		if clusterList, err = client.ClusterList(); err != nil {
			return false, err
		}

		for _clusterName := range clusterList {
			if _clusterName == clusterName {
				return true, nil
			}
		}
		return false, nil
	}

	if err = spinner.Poll(ctx, isClusterExistInSassFunc, CLUSTER_POLLING_INTERVAL, CLUSTER_POLLING_TIMEOUT); err == nil {
		return nil
	}

	if err == nil {
		return nil
	}

	spinner.StopFail()

	if errors.Is(err, ui.ErrSpinnerTimeout) {
		return fmt.Errorf("timeout waiting for groundcover to connect cloud platform")
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
