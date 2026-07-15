//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProviderOnly(t *testing.T) {
	tests := []struct {
		name       string
		role       any
		setRole    bool
		wantStatus int
	}{
		{name: "provider_allowed", role: service.RoleProvider, setRole: true, wantStatus: http.StatusOK},
		{name: "admin_provider_allowed", role: service.RoleAdminProvider, setRole: true, wantStatus: http.StatusOK},
		{name: "admin_forbidden", role: service.RoleAdmin, setRole: true, wantStatus: http.StatusForbidden},
		{name: "user_forbidden", role: service.RoleUser, setRole: true, wantStatus: http.StatusForbidden},
		{name: "missing_role_unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "invalid_role_type_unauthorized", role: int64(1), setRole: true, wantStatus: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			if tc.setRole {
				router.Use(func(c *gin.Context) {
					c.Set(string(ContextKeyUserRole), tc.role)
					c.Next()
				})
			}
			router.Use(ProviderOnly())
			router.GET("/provider", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/provider", nil)
			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestAdminOnlyAllowsCombinedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), service.RoleAdminProvider)
		c.Next()
	})
	router.Use(AdminOnly())
	router.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
