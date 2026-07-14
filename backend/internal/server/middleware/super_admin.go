package middleware

import (
	"net/http"

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
		if IsSuperAdminFromContext(c) {
			c.Next()
			return
		}

		AbortWithError(c, http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "Super administrator access required")
	}
}
