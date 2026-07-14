//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequireSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		isSuperAdmin   any
		setContext     bool
		expectedStatus int
	}{
		{name: "super admin allowed", isSuperAdmin: true, setContext: true, expectedStatus: http.StatusOK},
		{name: "ordinary admin rejected", isSuperAdmin: false, setContext: true, expectedStatus: http.StatusForbidden},
		{name: "missing identity rejected", expectedStatus: http.StatusForbidden},
		{name: "invalid identity rejected", isSuperAdmin: "true", setContext: true, expectedStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			if tt.setContext {
				router.Use(func(c *gin.Context) {
					c.Set(string(ContextKeyIsSuperAdmin), tt.isSuperAdmin)
					c.Next()
				})
			}
			router.Use(RequireSuperAdmin())
			router.POST("/protected", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/protected", nil)
			router.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusForbidden {
				require.Contains(t, w.Body.String(), "SUPER_ADMIN_REQUIRED")
			}
		})
	}
}
