package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterProviderRoutes registers provider self-service routes.
func RegisterProviderRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	provider := v1.Group("/provider")
	provider.Use(gin.HandlerFunc(jwtAuth))
	provider.Use(middleware.BackendModeUserGuard(settingService))
	provider.Use(middleware.ProviderOnly())
	{
		accounts := provider.Group("/accounts")
		accounts.GET("", h.Provider.ListAccounts)
		accounts.GET("/:id", h.Provider.GetAccount)
		accounts.POST("", h.Provider.CreateAccount)
		accounts.PUT("/:id", h.Provider.UpdateAccount)
		accounts.DELETE("/:id", h.Provider.DeleteAccount)

		provider.GET("/groups", h.Provider.ListGroups)
		provider.GET("/usage", h.Provider.GetUsage)
	}
}
