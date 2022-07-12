package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
	"groundcover.com/pkg/auth"
)

const (
	API_KEY_GENERATE_URL                       = "https://app.groundcover.com/api/cluster/list"
	CLUSTER_POLLING_INTERVAL                   = time.Second * 10
	WAIT_FOR_CLUSTER_CONNECTED_TO_SAAS_TIMEOUT = time.Minute * 5
	SPINNER_TYPE                               = 27 // ▁▂▃▄▅▆▇█▉▊▋▌▍▎▏▏▎▍▌▋▊▉█▇▆▅▄▃▂▁
)

type ListResponse struct {
	ClusterIds []string `json:"clusterIds"`
}

func WaitUntilClusterConnectedToSaas(ctx context.Context, token *auth.Auth0Token, clusterToPoll string) error {
	ctx, cancel := context.WithTimeout(ctx, WAIT_FOR_CLUSTER_CONNECTED_TO_SAAS_TIMEOUT)
	defer cancel()

	s := spinner.New(spinner.CharSets[SPINNER_TYPE], 100*time.Millisecond)
	s.Suffix = " Waiting until groundcover connected to saas"
	s.Color("red")
	s.Start()
	defer s.Stop()

	ticker := time.NewTicker(CLUSTER_POLLING_INTERVAL)

	for {
		select {
		case <-ticker.C:
			clusterExists, err := doesClusterExistsInSass(token, clusterToPoll)
			if err != nil {
				return err
			}

			if clusterExists {
				return nil
			}
		case <-ctx.Done():
			return errors.New("timed out while waiting for groundcover to connect to saas")
		}
	}
}

func doesClusterExistsInSass(token *auth.Auth0Token, clusterToPoll string) (bool, error) {
	clusters, err := getClusters(token)
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if clusterToPoll == cluster {
			return true, nil
		}
	}

	return false, nil
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
