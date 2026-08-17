package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfiguredSuperAdminEmailUsesPersistedFallbackAndEnvironmentOverride(t *testing.T) {
	t.Setenv("ADMIN_EMAIL", "")
	SetConfiguredSuperAdminEmailFallback("owner@example.com")
	t.Cleanup(func() { SetConfiguredSuperAdminEmailFallback("") })

	require.Equal(t, "owner@example.com", ConfiguredSuperAdminEmail())
	require.True(t, IsSuperAdminEmail(" OWNER@example.com "))

	t.Setenv("ADMIN_EMAIL", "override@example.com")
	require.Equal(t, "override@example.com", ConfiguredSuperAdminEmail())
	require.False(t, IsSuperAdminEmail("owner@example.com"))
}
