package auth

import (
	"encoding/json"
)

const (
	GENERATE_SERVICE_ACCOUNT_TOKEN_ENDPOINT = "system/generate-service-account-token"
)

type SAToken struct {
	Token string `json:"token" validate:"required"`
}

func (sa *SAToken) ParseBody(body []byte) error {
	var err error

	if err = json.Unmarshal(body, &sa); err != nil {
		return err
	}

	if err = validate.Struct(sa); err != nil {
		return err
	}

	return nil
}
