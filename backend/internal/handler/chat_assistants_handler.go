package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatAssistantsHandler 处理内置聊天的用户自定义助手（JWT 鉴权，按 user 隔离）。
type ChatAssistantsHandler struct {
	svc *service.ChatAssistantsService
}

func NewChatAssistantsHandler(svc *service.ChatAssistantsService) *ChatAssistantsHandler {
	return &ChatAssistantsHandler{svc: svc}
}

func parseAssistantID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid assistant id")
		return 0, false
	}
	return id, true
}

// List GET /api/v1/chat/assistants
func (h *ChatAssistantsHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	assistants, err := h.svc.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, assistants)
}

// Get GET /api/v1/chat/assistants/:id
func (h *ChatAssistantsHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAssistantID(c)
	if !ok {
		return
	}
	assistant, err := h.svc.Get(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, assistant)
}

// Create POST /api/v1/chat/assistants —— 新建助手，返回 id。
func (h *ChatAssistantsHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req service.ChatAssistantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	id, err := h.svc.Create(c.Request.Context(), subject.UserID, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Update PUT /api/v1/chat/assistants/:id
func (h *ChatAssistantsHandler) Update(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAssistantID(c)
	if !ok {
		return
	}
	var req service.ChatAssistantInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), subject.UserID, id, req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// Delete DELETE /api/v1/chat/assistants/:id
func (h *ChatAssistantsHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAssistantID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), subject.UserID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}
