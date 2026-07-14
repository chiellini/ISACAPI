package service

import (
	"os"
	"strings"
)

// ConfiguredSuperAdminEmail returns the super-admin identity configured by .env.
func ConfiguredSuperAdminEmail() string {
	return strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
}

func IsSuperAdminEmail(email string) bool {
	configured := ConfiguredSuperAdminEmail()
	return configured != "" && strings.EqualFold(strings.TrimSpace(email), configured)
}
