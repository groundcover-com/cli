package auth

import (
	"encoding/json"
)

const (
	GenerateAPIKeyEndpoint       = "system/generate-api-key"
	GetDatasourcesAPIKeyEndpoint = "system/get-datasources-api-key"
)

type ApiKey struct {
	ApiKey string `json:"apiKey" validate:"required"`
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
