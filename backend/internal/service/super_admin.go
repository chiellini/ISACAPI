package service

import (
	"os"
	"strings"
	"sync/atomic"
)

var configuredSuperAdminEmailFallback atomic.Value

// SetConfiguredSuperAdminEmailFallback supplies the installation-time admin
// identity persisted in config.yaml. ADMIN_EMAIL remains the explicit runtime
// override for container and orchestrated deployments.
func SetConfiguredSuperAdminEmailFallback(email string) {
	configuredSuperAdminEmailFallback.Store(strings.TrimSpace(email))
}

// ConfiguredSuperAdminEmail returns the super-admin identity configured by the
// runtime environment or, when absent, by the owner-only installation config.
func ConfiguredSuperAdminEmail() string {
	if configured := strings.TrimSpace(os.Getenv("ADMIN_EMAIL")); configured != "" {
		return configured
	}
	if configured := configuredSuperAdminEmailFallback.Load(); configured != nil {
		if email, ok := configured.(string); ok {
			return strings.TrimSpace(email)
		}
	}
	return ""
}

func IsSuperAdminEmail(email string) bool {
	configured := ConfiguredSuperAdminEmail()
	return configured != "" && strings.EqualFold(strings.TrimSpace(email), configured)
}
