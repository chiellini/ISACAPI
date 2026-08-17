package middleware

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// IsSuperAdminFromContext reports whether admin authentication identified the
// current account as the ADMIN_EMAIL account.
func IsSuperAdminFromContext(c *gin.Context) bool {
	value, exists := c.Get(string(ContextKeyIsSuperAdmin))
	if !exists {
		return false
	}
	isSuperAdmin, ok := value.(bool)
	return ok && isSuperAdmin
}

// RequireSuperAdmin protects operations reserved for the ADMIN_EMAIL account.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// The shared admin API key is a machine credential, not proof that the
		// caller is the ADMIN_EMAIL account. adminAuth associates that credential
		// with the first admin for legacy audit attribution, so trusting only the
		// derived super-admin flag would let every key holder impersonate that
		// account on identity-restricted routes.
		if c.GetString("auth_method") != service.AuditAuthMethodAdminAPIKey && IsSuperAdminFromContext(c) {
			c.Next()
			return
		}

		AbortWithError(c, http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "Super administrator access required")
	}
}
