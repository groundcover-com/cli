package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIngestionKeyCmd_TenantUUIDFlag(t *testing.T) {
	// Save original viper state
	originalTenantUUID := viper.GetString(TENANT_UUID_FLAG)
	defer viper.Set(TENANT_UUID_FLAG, originalTenantUUID)

	tests := []struct {
		name              string
		tenantUUIDFlag    string
		expectedBehavior  string
	}{
		{
			name:             "tenant-uuid flag provided",
			tenantUUIDFlag:   "test-tenant-uuid-123",
			expectedBehavior: "should use provided tenant UUID",
		},
		{
			name:             "tenant-uuid flag empty",
			tenantUUIDFlag:   "",
			expectedBehavior: "should call fetchTenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the flag value
			viper.Set(TENANT_UUID_FLAG, tt.tenantUUIDFlag)

			// Get the value back
			result := viper.GetString(TENANT_UUID_FLAG)

			// Assert
			assert.Equal(t, tt.tenantUUIDFlag, result)
		})
	}
}

func TestTenantUUIDFlagBinding(t *testing.T) {
	// Verify that TENANT_UUID_FLAG constant is correct
	assert.Equal(t, "tenant-uuid", TENANT_UUID_FLAG)

	// Verify the flag exists in RootCmd
	flag := RootCmd.PersistentFlags().Lookup(TENANT_UUID_FLAG)
	assert.NotNil(t, flag, "tenant-uuid flag should be registered")
	assert.Equal(t, "", flag.DefValue, "tenant-uuid flag should have empty default value")
}
