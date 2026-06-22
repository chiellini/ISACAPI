package handler

import (
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
	Title    string                       `json:"title"`
	Model    string                       `json:"model"`
	Messages []service.ChatHistoryMessage `json:"messages"`
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

// Get GET /api/v1/chat/sessions/:id —— 单个会话（含消息）。
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

// Update PUT /api/v1/chat/sessions/:id —— 覆盖标题/模型并整体替换消息（每轮对话后保存）。
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
	if err := h.svc.Update(c.Request.Context(), subject.UserID, id, req.Title, req.Model, req.Messages); err != nil {
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
