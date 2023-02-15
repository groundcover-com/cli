package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/auth"
)

type AuthTokenTestSuite struct {
	suite.Suite
}

func (suite *AuthTokenTestSuite) SetupSuite() {}

func (suite *AuthTokenTestSuite) TearDownSuite() {}

func TestAuthTokenTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, &AuthTokenTestSuite{})
}

func (suite *AuthTokenTestSuite) TestParseInstallationTokenSuccess() {
	//prepare
	var err error

	token := map[string]string{
		"id":     "myid",
		"apiKey": "testApiKey",
		"org":    "example.com",
		"email":  "user@example.com",
	}

	var data []byte
	data, err = json.Marshal(token)
	suite.NoError(err)

	encodedToken := base64.StdEncoding.EncodeToString(data)

	//act
	var installationToken *auth.InstallationToken
	installationToken, err = auth.NewInstallationToken(encodedToken)

	// assert

	expected := &auth.InstallationToken{
		Id:     token["id"],
		Org:    token["org"],
		Email:  token["email"],
		ApiKey: &auth.ApiKey{ApiKey: token["apiKey"]},
	}

	suite.NoError(err)

	suite.Equal(expected, installationToken)
}

func (suite *AuthTokenTestSuite) TestParseInstallationTokenValidationError() {
	//prepare
	var err error

	token := map[string]string{
		"id-bad":     "myid",
		"apiKey-bad": "testApiKey",
		"org-bad":    "example.com",
		"email-bad":  "user@example.com",
	}

	var data []byte
	data, err = json.Marshal(token)
	suite.NoError(err)

	encodedToken := base64.StdEncoding.EncodeToString(data)

	//act
	_, err = auth.NewInstallationToken(encodedToken)

	// assert
	expected := []string{
		"Key: 'InstallationToken.ApiKey' Error:Field validation for 'ApiKey' failed on the 'required' tag",
		"Key: 'InstallationToken.Id' Error:Field validation for 'Id' failed on the 'required' tag",
		"Key: 'InstallationToken.Org' Error:Field validation for 'Org' failed on the 'required' tag",
		"Key: 'InstallationToken.Email' Error:Field validation for 'Email' failed on the 'required' tag",
	}

	validationErrors, _ := err.(validator.ValidationErrors)
	suite.Len(validationErrors, 4)

	var errs []string
	for _, validationError := range validationErrors {
		errs = append(errs, validationError.Error())
	}

	suite.Equal(expected, errs)
}
