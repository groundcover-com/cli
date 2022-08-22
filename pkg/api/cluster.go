package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"groundcover.com/pkg/utils"
)

const (
	CLUSTER_LIST_ENDPOINT        = "cluster/list"
	CLUSTER_POLLING_INTERVAL     = time.Second * 10
	CLUSTER_POLLING_TIMEOUT      = time.Minute * 5
	CLUSTER_POLLING_SPINNER_TYPE = 26 // ....
)

type ClusterList struct {
	ClusterIds []string `json:"clusterIds"`
}

func (client *Client) PollIsClusterExist(clusterName string) error {
	var err error

	spinner := utils.NewSpinner(CLUSTER_POLLING_SPINNER_TYPE, "Waiting until groundcover is connected to cloud platform ")

	isClusterExistInSassFunc := func() (bool, error) {
		var clusterList *ClusterList
		if clusterList, err = client.ClusterList(); err != nil {
			return false, err
		}

		for _, _clusterName := range clusterList.ClusterIds {
			if _clusterName == clusterName {
				spinner.FinalMSG = "groundcover is connected to cloud platform\n"
				return true, nil
			}
		}
		return false, nil
	}

	if err = spinner.Poll(isClusterExistInSassFunc, CLUSTER_POLLING_INTERVAL, CLUSTER_POLLING_TIMEOUT); err == nil {
		return nil
	}

	if errors.Is(err, utils.ErrSpinnerTimeout) {
		return fmt.Errorf("groundcover is yet connected to cloud platform")
	}

	return err
}

func (client *Client) ClusterList() (*ClusterList, error) {
	var err error

	var body []byte
	if body, err = client.Get(CLUSTER_LIST_ENDPOINT); err != nil {
		return nil, err
	}

	var clusterList *ClusterList
	if err = json.Unmarshal(body, &clusterList); err != nil {
		return nil, err
	}

	return clusterList, nil
}
