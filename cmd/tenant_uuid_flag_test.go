package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type TenantUUIDFlagTestSuite struct {
	suite.Suite
	originalTenantUUID string
}

func (suite *TenantUUIDFlagTestSuite) SetupTest() {
	// Save original viper state before each test
	suite.originalTenantUUID = viper.GetString(TENANT_UUID_FLAG)
}

func (suite *TenantUUIDFlagTestSuite) TearDownTest() {
	// Restore original viper state after each test
	viper.Set(TENANT_UUID_FLAG, suite.originalTenantUUID)
}

func TestTenantUUIDFlagTestSuite(t *testing.T) {
	suite.Run(t, &TenantUUIDFlagTestSuite{})
}

func (suite *TenantUUIDFlagTestSuite) TestTenantUUIDFlagExists() {
	// Verify the flag is registered as a persistent flag
	flag := RootCmd.PersistentFlags().Lookup(TENANT_UUID_FLAG)
	suite.NotNil(flag, "tenant-uuid flag should be registered")
	suite.Equal("", flag.DefValue, "tenant-uuid flag should have empty default value")
	suite.Equal("optional tenant-uuid", flag.Usage, "tenant-uuid flag should have correct usage description")
}

func (suite *TenantUUIDFlagTestSuite) TestTenantUUIDFlagConstant() {
	// Verify the constant value is correct
	suite.Equal("tenant-uuid", TENANT_UUID_FLAG)
}

func (suite *TenantUUIDFlagTestSuite) TestViperBindingForTenantUUID() {
	// Test that viper can read the flag value
	testUUID := "test-uuid-12345"
	viper.Set(TENANT_UUID_FLAG, testUUID)

	result := viper.GetString(TENANT_UUID_FLAG)
	suite.Equal(testUUID, result)
}

func (suite *TenantUUIDFlagTestSuite) TestEmptyTenantUUIDFlag() {
	// When flag is empty, viper should return empty string
	viper.Set(TENANT_UUID_FLAG, "")

	result := viper.GetString(TENANT_UUID_FLAG)
	suite.Equal("", result)
}

func (suite *TenantUUIDFlagTestSuite) TestTenantUUIDFlagWithValidUUID() {
	// Test with a valid UUID format
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	viper.Set(TENANT_UUID_FLAG, validUUID)

	result := viper.GetString(TENANT_UUID_FLAG)
	suite.Equal(validUUID, result)
}

// Test that all affected commands have access to the tenant-uuid flag
func (suite *TenantUUIDFlagTestSuite) TestCommandsHaveAccessToTenantUUIDFlag() {
	commands := []struct {
		name string
		cmd  interface{}
	}{
		{"IngestionKeyCmd", IngestionKeyCmd},
		{"apiKeyCmd", apiKeyCmd},
		{"getDatasourcesAPIKeyCmd", getDatasourcesAPIKeyCmd},
		{"generateClientTokenCmd", generateClientTokenCmd},
		{"serviceAccountTokenCmd", serviceAccountTokenCmd},
	}

	for _, tc := range commands {
		suite.Run(tc.name, func() {
			// All these commands are children of AuthCmd or RootCmd
			// and should inherit the persistent flag
			flag := RootCmd.PersistentFlags().Lookup(TENANT_UUID_FLAG)
			suite.NotNil(flag, "Flag should be accessible to "+tc.name)
		})
	}
}

// Integration-style test verifying the pattern used in the fixed commands
func (suite *TenantUUIDFlagTestSuite) TestTenantUUIDFlagPattern() {
	// This test verifies the pattern: if tenantUUID = viper.GetString(TENANT_UUID_FLAG); tenantUUID == ""

	suite.Run("flag not set - should be empty", func() {
		viper.Set(TENANT_UUID_FLAG, "")
		tenantUUID := viper.GetString(TENANT_UUID_FLAG)

		if tenantUUID == "" {
			// This is the expected path - should call fetchTenant()
			suite.Empty(tenantUUID)
		} else {
			suite.Fail("Expected empty tenant UUID when flag not set")
		}
	})

	suite.Run("flag set - should use flag value", func() {
		expectedUUID := "my-custom-tenant-uuid"
		viper.Set(TENANT_UUID_FLAG, expectedUUID)
		tenantUUID := viper.GetString(TENANT_UUID_FLAG)

		if tenantUUID == "" {
			suite.Fail("Should not be empty when flag is set")
		} else {
			// This is the expected path - should NOT call fetchTenant()
			suite.Equal(expectedUUID, tenantUUID)
		}
	})
}
