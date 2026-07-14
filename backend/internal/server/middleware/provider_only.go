package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ProviderOnly restricts a route to authenticated provider users.
// It must be installed after JWTAuth so the role is present in context.
func ProviderOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRoleFromContext(c)
		if !ok {
			AbortWithError(c, 401, "UNAUTHORIZED", "User not found in context")
			return
		}

		if role != service.RoleProvider {
			AbortWithError(c, 403, "FORBIDDEN", "Provider access required")
			return
		}

		c.Next()
	}
}
