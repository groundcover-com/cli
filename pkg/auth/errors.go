package auth

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Auth0Error struct {
	error
	Type        string `json:"error"`
	Description string `json:"error_description"`
}

func NewAuth0Error(body []byte) error {
	var auth0Error *Auth0Error

	if err := json.Unmarshal(body, &auth0Error); err != nil {
		return errors.Wrap(err, "failed to decode Auth0 error response")
	}

	auth0Error.error = fmt.Errorf("%s: %s", auth0Error.Type, auth0Error.Description)

	return auth0Error
}
