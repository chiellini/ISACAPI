package handler

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHistoryHandler 处理内置聊天的会话历史（JWT 鉴权，按 user 隔离，跨设备同步）。
type ChatHistoryHandler struct {
	svc *service.ChatHistoryService
}

func NewChatHistoryHandler(svc *service.ChatHistoryService) *ChatHistoryHandler {
	return &ChatHistoryHandler{svc: svc}
}

type chatSessionUpsertRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
	// 记忆字段用指针区分「未提供（保持原值）」与「显式置空」。
	Summary         *string                      `json:"summary"`
	Memory          *string                      `json:"memory"`
	SummarizedCount *int                         `json:"summarized_count"`
	Messages        []service.ChatHistoryMessage `json:"messages"`
}

type chatImageUploadRequest struct {
	// Image 支持完整 data URL（data:image/png;base64,...）或裸 base64。
	Image string `json:"image"`
}

type chatSearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

// List GET /api/v1/chat/sessions —— 会话列表（不含消息）。
func (h *ChatHistoryHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	sessions, err := h.svc.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, sessions)
}

// Get GET /api/v1/chat/sessions/:id —— 单个会话（含消息与记忆）。
func (h *ChatHistoryHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid session id")
		return
	}
	sess, err := h.svc.Get(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, sess)
}

// Create POST /api/v1/chat/sessions —— 新建空会话，返回 id。
func (h *ChatHistoryHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req chatSessionUpsertRequest
	_ = c.ShouldBindJSON(&req) // 允许空 body
	id, err := h.svc.Create(c.Request.Context(), subject.UserID, req.Title, req.Model)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Update PUT /api/v1/chat/sessions/:id —— 覆盖标题/模型/记忆并按需整体替换消息（每轮对话后保存）。
func (h *ChatHistoryHandler) Update(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid session id")
		return
	}
	var req chatSessionUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	update := service.ChatSessionUpdate{
		Title:           req.Title,
		Model:           req.Model,
		Summary:         req.Summary,
		Memory:          req.Memory,
		SummarizedCount: req.SummarizedCount,
		Messages:        req.Messages,
	}
	if err := h.svc.Update(c.Request.Context(), subject.UserID, id, update); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Delete DELETE /api/v1/chat/sessions/:id
func (h *ChatHistoryHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid session id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), subject.UserID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// UploadImage POST /api/v1/chat/sessions/:id/images —— 把一张（生成/上传的）图片持久化到服务端，
// 返回可回读的相对 URL。图片按会话与 user 双重隔离。
func (h *ChatHistoryHandler) UploadImage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid session id")
		return
	}
	var req chatImageUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	id, mime, err := h.svc.SaveImage(c.Request.Context(), subject.UserID, sessionID, req.Image)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"id":   id,
		"mime": mime,
		"url":  "/api/v1/chat/images/" + id,
	})
}

// GetImage GET /api/v1/chat/images/:imageId —— 回读一张图片的原始字节（JWT 鉴权，按 user 隔离）。
func (h *ChatHistoryHandler) GetImage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	imageID := c.Param("imageId")
	mime, data, err := h.svc.GetImage(c.Request.Context(), subject.UserID, imageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	// 图片内容不可变：可安全长缓存。private 避免共享缓存跨用户串图。
	c.Header("Cache-Control", "private, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	if mime == "" {
		mime = "image/png"
	}
	// 用 c.Data 直出原始字节（绕过标准 JSON 信封）。
	c.Data(http.StatusOK, mime, data)
}

// Capabilities GET /api/v1/chat/capabilities —— 告知前端当前可用的能力（如是否可联网搜索）。
func (h *ChatHistoryHandler) Capabilities(c *gin.Context) {
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	response.Success(c, gin.H{
		"web_search": h.svc.WebSearchAvailable(),
	})
}

// Search POST /api/v1/chat/search —— 执行一次联网搜索（复用网关侧搜索提供方与配额）。
// 供聊天页的工具调用（web_search）回灌结果，模型据此合成带来源的回答。
func (h *ChatHistoryHandler) Search(c *gin.Context) {
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req chatSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	results, err := h.svc.SearchWeb(c.Request.Context(), req.Query, req.MaxResults)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"results": results})
}
