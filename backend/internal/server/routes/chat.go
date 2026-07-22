package routes

import (
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	ratemw "github.com/Wei-Shaw/sub2api/internal/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterChatRoutes 注册内置聊天 Playground 路由。
//
// 这是面向小白用户的会话代理：聊天页用 JWT 登录态访问，桥接中间件把请求改写成「该用户用
// 自己的内置 Key 发起的一次普通网关调用」，从而 100% 复用现有 chat/completions 处理与计费链路。
//
// 中间件链刻意对齐网关组（gateway.go），仅把 apiKeyAuth 之前的 JWT 段替换为
// jwtAuth + chatBridge：
//
//	bodyLimit → clientRequestID → opsErrorLogger → endpointNorm
//	  → jwtAuth → chatBridge → apiKeyAuth → requireGroupAnthropic → handler
//
// 路由路径含 "/v1/chat/completions"，使 NormalizeInboundEndpoint 正确归一为
// chat_completions，用量归因与真实网关一致。
func RegisterChatRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
	redisClient *redis.Client,
) {
	bodyLimit := middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize)
	clientRequestID := middleware.ClientRequestID()
	opsErrorLogger := handler.OpsErrorLoggerMiddleware(opsService)
	endpointNorm := handler.InboundEndpointMiddleware()
	chatBridge := middleware.NewChatBridgeMiddleware(apiKeyService)
	requireGroupAnthropic := middleware.RequireGroupAssignment(settingService, middleware.AnthropicErrorWriter)

	chat := r.Group("/api/v1/chat")
	// 标记跳过网关侧会话存档：聊天历史已单独存于 chat_sessions/chat_messages，
	// 不重复写入 API 会话存档系统。
	chat.Use(func(c *gin.Context) {
		c.Set(handler.CtxSkipConversationCapture, true)
		c.Next()
	})
	chat.Use(bodyLimit)
	chat.Use(clientRequestID)
	chat.Use(opsErrorLogger)
	chat.Use(endpointNorm)
	chat.Use(gin.HandlerFunc(jwtAuth))
	chat.Use(chatBridge)
	chat.Use(gin.HandlerFunc(apiKeyAuth))
	chat.Use(requireGroupAnthropic)
	{
		// 主对话入口：按 group 平台自动路由，完全复用现有 handler（含 SSE 流式）。
		chat.POST("/v1/chat/completions", func(c *gin.Context) {
			if getGroupPlatform(c) == service.PlatformOpenAI {
				h.OpenAIGateway.ChatCompletions(c)
				return
			}
			h.Gateway.ChatCompletions(c)
		})
		// 生图入口（GPT Image 等）：仅 OpenAI 平台分组支持。
		chat.POST("/v1/images/generations", func(c *gin.Context) {
			if getGroupPlatform(c) != service.PlatformOpenAI {
				service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Image generation is not supported for this group",
					},
				})
				return
			}
			h.OpenAIGateway.Images(c)
		})
		// 模型下拉列表（复用网关 Models）。
		chat.GET("/v1/models", h.Gateway.Models)
	}

	// 会话历史（仅 JWT 鉴权，按 user 隔离，跨设备同步；无需桥接/网关计费）。
	// bodyLimit 兜住图片上传等较大请求体（上限沿用网关 max_body_size）。
	sessions := r.Group("/api/v1/chat/sessions")
	sessions.Use(bodyLimit)
	sessions.Use(gin.HandlerFunc(jwtAuth))
	{
		sessions.GET("", h.ChatHistory.List)
		sessions.POST("", h.ChatHistory.Create)
		sessions.GET("/:id", h.ChatHistory.Get)
		sessions.PUT("/:id", h.ChatHistory.Update)
		sessions.DELETE("/:id", h.ChatHistory.Delete)
		// 生成/上传图片的服务端持久化：落库后可跨设备回读，替代仅存浏览器的方案。
		sessions.POST("/:id/images", h.ChatHistory.UploadImage)
	}

	// 图片回读（JWT 鉴权，按 user 隔离；直出原始字节，不走 JSON 信封）。
	images := r.Group("/api/v1/chat/images")
	images.Use(gin.HandlerFunc(jwtAuth))
	{
		images.GET("/:imageId", h.ChatHistory.GetImage)
	}

	// 联网搜索能力 + 搜索接口（JWT 鉴权）。搜索会命中外部提供方，按 IP 限流兜住滥用，
	// 真正的预算保护仍由提供方配额（Brave/Tavily quota_limit）承担。
	searchLimiter := ratemw.NewRateLimiter(redisClient)
	chatMeta := r.Group("/api/v1/chat")
	chatMeta.Use(gin.HandlerFunc(jwtAuth))
	{
		chatMeta.GET("/capabilities", h.ChatHistory.Capabilities)
		chatMeta.POST("/search", searchLimiter.Limit("chat-web-search", 30, time.Minute), h.ChatHistory.Search)
	}
}
