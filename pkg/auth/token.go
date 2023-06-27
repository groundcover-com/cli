package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Token interface {
	GetId() string
	GetOrg() string
	GetEmail() string
	GetSessionId() string
}

type InstallationToken struct {
	*ApiKey    `validate:"required"`
	Id         string `json:"id" validate:"required"`
	Org        string `json:"org" validate:"required"`
	Email      string `json:"email" validate:"required"`
	SessionId  string `json:"sessionId" validate:"required"`
	Tenant     string `json:"tenant" validate:"required"`
	TenantUUID string `json:"tenantUUID" validate:"required"`
}

func NewInstallationToken(encodedToken string) (*InstallationToken, error) {
	var err error

	if encodedToken == "" {
		return nil, fmt.Errorf("empty input token")
	}

	var data []byte
	if data, err = base64.StdEncoding.DecodeString(encodedToken); err != nil {
		return nil, err
	}

	token := &InstallationToken{}
	if err = json.Unmarshal(data, token); err != nil {
		return nil, err
	}

	if err = validate.Struct(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (installationToken InstallationToken) GetId() string {
	return installationToken.Id
}

func (installationToken InstallationToken) GetOrg() string {
	return installationToken.Org
}

func (installationToken InstallationToken) GetEmail() string {
	return installationToken.Email
}

func (installationToken InstallationToken) GetSessionId() string {
	return installationToken.SessionId
}
