package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// NewChatBridgeMiddleware 构造「会话 → 内置 Key」桥接中间件。
//
// 它必须夹在 jwtAuth（验证会话、写入 ContextKeyUser）与 apiKeyAuth（按 Key 做完整鉴权
// 与计费执行）之间：取出会话用户的内置聊天 Key，覆盖 Authorization 头，使下游 apiKeyAuth
// 把这次请求当作该用户用自己 Key 发起的一次普通 API 调用——计费/配额/限流/用量记录全部复用。
func NewChatBridgeMiddleware(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := GetAuthSubjectFromContext(c)
		if !ok {
			AbortWithError(c, 401, "UNAUTHORIZED", "session authentication required")
			return
		}

		internalKey, err := apiKeyService.GetOrCreateInternalChatKey(c.Request.Context(), subject.UserID)
		if err != nil {
			AbortWithError(c, 403, "CHAT_UNAVAILABLE", "chat is not available for this account")
			return
		}

		// 用内置 Key 覆盖鉴权头，交给下游 apiKeyAuth。
		c.Request.Header.Set("Authorization", "Bearer "+internalKey.Key)
		c.Request.Header.Del("x-api-key")
		c.Request.Header.Del("x-goog-api-key")

		c.Next()
	}
}
