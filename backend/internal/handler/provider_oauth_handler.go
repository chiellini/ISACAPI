package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ProviderOAuthHandler exposes the OAuth link/exchange flows to provider
// self-service account creation. It mirrors the admin OAuth handlers but
// deliberately accepts no proxy_id: proxies are an administrator-controlled
// resource and provider OAuth sessions always run without one.
type ProviderOAuthHandler struct {
	oauthService            *service.OAuthService
	openaiOAuthService      *service.OpenAIOAuthService
	geminiOAuthService      *service.GeminiOAuthService
	antigravityOAuthService *service.AntigravityOAuthService
	grokOAuthService        *service.GrokOAuthService
}

func NewProviderOAuthHandler(
	oauthService *service.OAuthService,
	openaiOAuthService *service.OpenAIOAuthService,
	geminiOAuthService *service.GeminiOAuthService,
	antigravityOAuthService *service.AntigravityOAuthService,
	grokOAuthService *service.GrokOAuthService,
) *ProviderOAuthHandler {
	return &ProviderOAuthHandler{
		oauthService:            oauthService,
		openaiOAuthService:      openaiOAuthService,
		geminiOAuthService:      geminiOAuthService,
		antigravityOAuthService: antigravityOAuthService,
		grokOAuthService:        grokOAuthService,
	}
}

// ========== Anthropic (Claude) ==========

// AnthropicGenerateAuthURL handles POST /api/v1/provider/oauth/anthropic/generate-auth-url.
func (h *ProviderOAuthHandler) AnthropicGenerateAuthURL(c *gin.Context) {
	result, err := h.oauthService.GenerateAuthURL(c.Request.Context(), nil)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// AnthropicGenerateSetupTokenURL handles POST /api/v1/provider/oauth/anthropic/generate-setup-token-url.
func (h *ProviderOAuthHandler) AnthropicGenerateSetupTokenURL(c *gin.Context) {
	result, err := h.oauthService.GenerateSetupTokenURL(c.Request.Context(), nil)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

type providerAnthropicExchangeCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// AnthropicExchangeCode handles POST /api/v1/provider/oauth/anthropic/exchange-code
// and POST /api/v1/provider/oauth/anthropic/exchange-setup-token-code.
func (h *ProviderOAuthHandler) AnthropicExchangeCode(c *gin.Context) {
	var req providerAnthropicExchangeCodeRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	tokenInfo, err := h.oauthService.ExchangeCode(c.Request.Context(), &service.ExchangeCodeInput{
		SessionID: req.SessionID,
		Code:      req.Code,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, tokenInfo)
}

// ========== OpenAI ==========

// OpenAIGenerateAuthURL handles POST /api/v1/provider/oauth/openai/generate-auth-url.
func (h *ProviderOAuthHandler) OpenAIGenerateAuthURL(c *gin.Context) {
	result, err := h.openaiOAuthService.GenerateAuthURL(c.Request.Context(), nil, "", service.PlatformOpenAI)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

type providerOpenAIExchangeCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
	State     string `json:"state" binding:"required"`
}

// OpenAIExchangeCode handles POST /api/v1/provider/oauth/openai/exchange-code.
func (h *ProviderOAuthHandler) OpenAIExchangeCode(c *gin.Context) {
	var req providerOpenAIExchangeCodeRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	tokenInfo, err := h.openaiOAuthService.ExchangeCode(c.Request.Context(), &service.OpenAIExchangeCodeInput{
		SessionID: req.SessionID,
		Code:      req.Code,
		State:     req.State,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, tokenInfo)
}

// ========== Gemini ==========

// GeminiGetCapabilities handles GET /api/v1/provider/oauth/gemini/capabilities.
func (h *ProviderOAuthHandler) GeminiGetCapabilities(c *gin.Context) {
	response.Success(c, h.geminiOAuthService.GetOAuthConfig())
}

type providerGeminiGenerateAuthURLRequest struct {
	ProjectID string `json:"project_id"`
	OAuthType string `json:"oauth_type"`
	TierID    string `json:"tier_id"`
}

// GeminiGenerateAuthURL handles POST /api/v1/provider/oauth/gemini/auth-url.
func (h *ProviderOAuthHandler) GeminiGenerateAuthURL(c *gin.Context) {
	var req providerGeminiGenerateAuthURLRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	oauthType, ok := normalizeGeminiOAuthType(c, req.OAuthType)
	if !ok {
		return
	}
	if service.GeminiOAuthTypeRequiresProjectID(oauthType) && strings.TrimSpace(req.ProjectID) == "" {
		response.BadRequest(c, "Project ID is required for "+oauthType+" OAuth")
		return
	}
	redirectURI := admin.DeriveGeminiRedirectURI(c)
	result, err := h.geminiOAuthService.GenerateAuthURL(c.Request.Context(), nil, redirectURI, req.ProjectID, oauthType, req.TierID)
	if err != nil {
		response.BadRequest(c, "Failed to generate auth URL: "+err.Error())
		return
	}
	response.Success(c, result)
}

type providerGeminiExchangeCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	State     string `json:"state" binding:"required"`
	Code      string `json:"code" binding:"required"`
	OAuthType string `json:"oauth_type"`
	TierID    string `json:"tier_id"`
}

// GeminiExchangeCode handles POST /api/v1/provider/oauth/gemini/exchange-code.
func (h *ProviderOAuthHandler) GeminiExchangeCode(c *gin.Context) {
	var req providerGeminiExchangeCodeRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	oauthType, ok := normalizeGeminiOAuthType(c, req.OAuthType)
	if !ok {
		return
	}
	tokenInfo, err := h.geminiOAuthService.ExchangeCode(c.Request.Context(), &service.GeminiExchangeCodeInput{
		SessionID: req.SessionID,
		State:     req.State,
		Code:      req.Code,
		OAuthType: oauthType,
		TierID:    req.TierID,
	})
	if err != nil {
		response.BadRequest(c, "Failed to exchange code: "+err.Error())
		return
	}
	response.Success(c, tokenInfo)
}

// normalizeGeminiOAuthType 与管理员端保持一致:空值回退 code_assist,非法值直接 400。
func normalizeGeminiOAuthType(c *gin.Context, raw string) (string, bool) {
	oauthType := strings.TrimSpace(raw)
	if oauthType == "" {
		oauthType = "code_assist"
	}
	if oauthType != "code_assist" && oauthType != "google_one" && oauthType != "ai_studio" {
		response.BadRequest(c, "Invalid oauth_type: must be 'code_assist', 'google_one', or 'ai_studio'")
		return "", false
	}
	return oauthType, true
}

// ========== Antigravity ==========

// AntigravityGenerateAuthURL handles POST /api/v1/provider/oauth/antigravity/auth-url.
func (h *ProviderOAuthHandler) AntigravityGenerateAuthURL(c *gin.Context) {
	result, err := h.antigravityOAuthService.GenerateAuthURL(c.Request.Context(), nil)
	if err != nil {
		response.InternalError(c, "Failed to generate auth URL: "+err.Error())
		return
	}
	response.Success(c, result)
}

type providerAntigravityExchangeCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	State     string `json:"state" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// AntigravityExchangeCode handles POST /api/v1/provider/oauth/antigravity/exchange-code.
func (h *ProviderOAuthHandler) AntigravityExchangeCode(c *gin.Context) {
	var req providerAntigravityExchangeCodeRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	tokenInfo, err := h.antigravityOAuthService.ExchangeCode(c.Request.Context(), &service.AntigravityExchangeCodeInput{
		SessionID: req.SessionID,
		State:     req.State,
		Code:      req.Code,
	})
	if err != nil {
		response.BadRequest(c, "Failed to exchange code: "+err.Error())
		return
	}
	response.Success(c, tokenInfo)
}

// ========== Grok ==========

// GrokGenerateAuthURL handles POST /api/v1/provider/oauth/grok/auth-url.
func (h *ProviderOAuthHandler) GrokGenerateAuthURL(c *gin.Context) {
	result, err := h.grokOAuthService.GenerateAuthURL(c.Request.Context(), nil, "")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

type providerGrokExchangeCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
	State     string `json:"state"`
}

// GrokExchangeCode handles POST /api/v1/provider/oauth/grok/exchange-code.
func (h *ProviderOAuthHandler) GrokExchangeCode(c *gin.Context) {
	var req providerGrokExchangeCodeRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	tokenInfo, err := h.grokOAuthService.ExchangeCode(c.Request.Context(), &service.GrokExchangeCodeInput{
		SessionID: req.SessionID,
		Code:      req.Code,
		State:     req.State,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, tokenInfo)
}
