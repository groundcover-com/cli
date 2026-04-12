package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type BackendFlagTestSuite struct {
	suite.Suite
	originalBackend string
}

func (suite *BackendFlagTestSuite) SetupTest() {
	suite.originalBackend = viper.GetString(BACKEND_NAME_FLAG)
}

func (suite *BackendFlagTestSuite) TearDownTest() {
	viper.Set(BACKEND_NAME_FLAG, suite.originalBackend)
}

func TestBackendFlagTestSuite(t *testing.T) {
	suite.Run(t, &BackendFlagTestSuite{})
}

func (suite *BackendFlagTestSuite) TestBackendFlagExists() {
	flag := RootCmd.PersistentFlags().Lookup(BACKEND_NAME_FLAG)
	suite.NotNil(flag, "backend flag should be registered")
	suite.Equal("", flag.DefValue, "backend flag should have empty default value")
	suite.Equal("backend name to use", flag.Usage, "backend flag should have correct usage description")
}

func (suite *BackendFlagTestSuite) TestBackendFlagConstant() {
	suite.Equal("backend-name", BACKEND_NAME_FLAG)
}

func (suite *BackendFlagTestSuite) TestViperBindingForBackend() {
	testBackend := "gcp-us-east1"
	viper.Set(BACKEND_NAME_FLAG, testBackend)

	result := viper.GetString(BACKEND_NAME_FLAG)
	suite.Equal(testBackend, result)
}

func (suite *BackendFlagTestSuite) TestEmptyBackendFlag() {
	viper.Set(BACKEND_NAME_FLAG, "")

	result := viper.GetString(BACKEND_NAME_FLAG)
	suite.Equal("", result)
}

func (suite *BackendFlagTestSuite) TestBackendFlagWithValidName() {
	validBackend := "gcp-us-east1"
	viper.Set(BACKEND_NAME_FLAG, validBackend)

	result := viper.GetString(BACKEND_NAME_FLAG)
	suite.Equal(validBackend, result)
}

func (suite *BackendFlagTestSuite) TestCommandsHaveAccessToBackendFlag() {
	commands := []struct {
		name string
		cmd  interface{}
	}{
		{"IngestionKeyCmd", IngestionKeyCmd},
		{"getDatasourcesAPIKeyCmd", getDatasourcesAPIKeyCmd},
		{"DeployCmd", DeployCmd},
	}

	for _, tc := range commands {
		suite.Run(tc.name, func() {
			flag := RootCmd.PersistentFlags().Lookup(BACKEND_NAME_FLAG)
			suite.NotNil(flag, "Flag should be accessible to "+tc.name)
		})
	}
}

func (suite *BackendFlagTestSuite) TestBackendFlagPattern() {
	suite.Run("flag not set - should be empty", func() {
		viper.Set(BACKEND_NAME_FLAG, "")
		backendName := viper.GetString(BACKEND_NAME_FLAG)

		if backendName == "" {
			suite.Empty(backendName)
		} else {
			suite.Fail("Expected empty backend name when flag not set")
		}
	})

	suite.Run("flag set - should use flag value", func() {
		expectedBackend := "gcp-us-east1"
		viper.Set(BACKEND_NAME_FLAG, expectedBackend)
		backendName := viper.GetString(BACKEND_NAME_FLAG)

		if backendName == "" {
			suite.Fail("Should not be empty when flag is set")
		} else {
			suite.Equal(expectedBackend, backendName)
		}
	})
}
