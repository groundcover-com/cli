package auth

import (
	"encoding/json"

	"groundcover.com/pkg/utils"
)

const (
	API_KEY_STORAGE_KEY       = "api_key.json"
	GENERATE_API_KEY_ENDPOINT = "system/generate-api-key"
)

type ApiKey struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func NewApiKey() (*ApiKey, error) {
	var err error

	var data []byte
	if data, err = utils.PresistentStorage.Read(API_KEY_STORAGE_KEY); err != nil {
		return nil, err
	}

	apiKey := &ApiKey{}
	if err = apiKey.ParseBody(data); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (apiKey *ApiKey) ParseBody(body []byte) error {
	var err error

	if err = json.Unmarshal(body, &apiKey); err != nil {
		return err
	}

	if err = validate.Struct(apiKey); err != nil {
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

	return utils.PresistentStorage.Write(API_KEY_STORAGE_KEY, data)
}
