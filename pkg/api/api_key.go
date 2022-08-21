package api

import (
	"encoding/json"

	"groundcover.com/pkg/utils"
)

const (
	API_KEY_STORAGE_KEY       = "api_key.json"
	GENERATE_API_KEY_ENDPOINT = "system/generate-api-key"
)

type ApiKey struct {
	ApiKey string `json:"apiKey"`
}

func (apiKey *ApiKey) Load() error {
	var err error

	var data []byte
	if data, err = utils.PresistentStorage.Read(API_KEY_STORAGE_KEY); err != nil {
		return err
	}

	if err = json.Unmarshal(data, &apiKey); err != nil {
		return err
	}

	return nil
}

func (apiKey *ApiKey) Save() error {
	var err error

	var data []byte
	if data, err = json.Marshal(apiKey); err != nil {
		return err
	}

	utils.PresistentStorage.Write(API_KEY_STORAGE_KEY, data)

	return nil
}

func (client *Client) ApiKey() (*ApiKey, error) {
	var err error

	var body []byte
	if body, err = client.Post(GENERATE_API_KEY_ENDPOINT, "", nil); err != nil {
		return nil, err
	}

	var apiKey *ApiKey
	if err = json.Unmarshal(body, &apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}
