package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/utils"
)

const (
	API_KEY_GENERATE_URL     = "https://app.groundcover.com/api/cluster/list"
	CLUSTER_POLLING_INTERVAL = time.Second * 10
	CLUSTER_POLLING_TIMEOUT  = time.Minute * 5
	SPINNER_TYPE             = 27 // ▁▂▃▄▅▆▇█▉▊▋▌▍▎▏▏▎▍▌▋▊▉█▇▆▅▄▃▂▁
)

type ListResponse struct {
	ClusterIds []string `json:"clusterIds"`
}

func WaitUntilClusterConnectedToSaas(token *auth.Auth0Token, clusterToPoll string) error {
	var err error
	var clusterNames []string

	spinner := utils.NewSpinner(SPINNER_TYPE, "Waiting until groundcover connected to saas ")

	isClusterExistInSassFunc := func() (bool, error) {
		if clusterNames, err = getClusters(token); err != nil {
			return false, err
		}
		for _, clusterName := range clusterNames {
			if clusterToPoll == clusterName {
				return true, nil
			}
		}
		return false, nil
	}

	if err = spinner.Poll(isClusterExistInSassFunc, CLUSTER_POLLING_INTERVAL, CLUSTER_POLLING_TIMEOUT); err != nil {
		return fmt.Errorf("timed out while waiting for groundcover to connect to saas")
	}

	return nil
}

func getClusters(token *auth.Auth0Token) ([]string, error) {
	req, err := http.NewRequest("GET", API_KEY_GENERATE_URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("unable to get clusters")
	}

	var clustersResponse *ListResponse
	err = json.Unmarshal(body, &clustersResponse)
	if err != nil {
		return nil, err
	}

	return clustersResponse.ClusterIds, nil
}
