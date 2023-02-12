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

func (apiKey *ApiKey) Load() error {
	var err error

	var data []byte
	if data, err = utils.PresistentStorage.Read(API_KEY_STORAGE_KEY); err != nil {
		return err
	}

	if err = json.Unmarshal(data, &apiKey); err != nil {
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
