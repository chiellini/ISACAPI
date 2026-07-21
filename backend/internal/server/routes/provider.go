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

		// OAuth link/exchange flows for provider self-service account creation.
		// Route suffixes mirror the admin endpoints so the frontend only swaps the prefix.
		oauth := provider.Group("/oauth")
		oauth.POST("/anthropic/generate-auth-url", h.ProviderOAuth.AnthropicGenerateAuthURL)
		oauth.POST("/anthropic/generate-setup-token-url", h.ProviderOAuth.AnthropicGenerateSetupTokenURL)
		oauth.POST("/anthropic/exchange-code", h.ProviderOAuth.AnthropicExchangeCode)
		oauth.POST("/anthropic/exchange-setup-token-code", h.ProviderOAuth.AnthropicExchangeCode)
		oauth.POST("/openai/generate-auth-url", h.ProviderOAuth.OpenAIGenerateAuthURL)
		oauth.POST("/openai/exchange-code", h.ProviderOAuth.OpenAIExchangeCode)
		oauth.GET("/gemini/capabilities", h.ProviderOAuth.GeminiGetCapabilities)
		oauth.POST("/gemini/auth-url", h.ProviderOAuth.GeminiGenerateAuthURL)
		oauth.POST("/gemini/exchange-code", h.ProviderOAuth.GeminiExchangeCode)
		oauth.POST("/antigravity/auth-url", h.ProviderOAuth.AntigravityGenerateAuthURL)
		oauth.POST("/antigravity/exchange-code", h.ProviderOAuth.AntigravityExchangeCode)
		oauth.POST("/grok/auth-url", h.ProviderOAuth.GrokGenerateAuthURL)
		oauth.POST("/grok/exchange-code", h.ProviderOAuth.GrokExchangeCode)
	}
}
