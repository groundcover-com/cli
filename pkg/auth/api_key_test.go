package auth_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/peterbourgon/diskv/v3"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/utils"
)

type AuthApiKeyTestSuite struct {
	suite.Suite
	diskv *diskv.Diskv
}

func (suite *AuthApiKeyTestSuite) SetupSuite() {
	suite.diskv = diskv.New(diskv.Options{
		BasePath:  filepath.Join(os.TempDir(), uuid.NewString()),
		Transform: func(s string) []string { return []string{} },
	})

	utils.PresistentStorage = suite.diskv
}

func (suite *AuthApiKeyTestSuite) TearDownSuite() {
	suite.diskv.EraseAll()
}

func TestAuthApiTokenTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, &AuthApiKeyTestSuite{})
}

func (suite *AuthApiKeyTestSuite) TestSaveAndLoadApiKeySuccess() {
	//prepare
	apiKeyValue := uuid.NewString()
	apiKey := &auth.ApiKey{
		ApiKey: apiKeyValue,
	}

	//act
	err := apiKey.Save()
	suite.NoError(err)

	loadedApiKey, err := auth.NewApiKey()
	suite.NoError(err)

	// assert
	expected, err := json.Marshal(apiKey)
	suite.NoError(err)

	storageData, err := suite.diskv.Read(auth.API_KEY_STORAGE_KEY)
	suite.NoError(err)

	suite.Equal(apiKey, loadedApiKey)
	suite.Equal(expected, storageData)
}

func (suite *AuthApiKeyTestSuite) TestApiKeyValidationError() {
	//prepare
	apiKey := "{\"apiKeyBad\":\"test\"}"

	err := suite.diskv.WriteString(auth.API_KEY_STORAGE_KEY, apiKey)
	suite.NoError(err)

	//act
	_, err = auth.NewApiKey()

	// assert
	expected := []string{
		"Key: 'ApiKey.ApiKey' Error:Field validation for 'ApiKey' failed on the 'required' tag",
	}

	validationErrors, _ := err.(validator.ValidationErrors)
	suite.Len(validationErrors, 1)

	var errs []string
	for _, validationError := range validationErrors {
		errs = append(errs, validationError.Error())
	}

	suite.Equal(expected, errs)
}
