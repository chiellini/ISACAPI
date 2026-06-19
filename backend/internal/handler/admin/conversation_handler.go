package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConversationHandler 提供管理员侧的对话存档查看与删除接口。
type ConversationHandler struct {
	service *service.ConversationService
}

func NewConversationHandler(svc *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{service: svc}
}

// ListSessions GET /admin/conversations
func (h *ConversationHandler) ListSessions(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filters := service.ConversationSessionListFilters{
		ContextDomain: strings.TrimSpace(c.Query("context_domain")),
		Protocol:      strings.TrimSpace(c.Query("protocol")),
		Status:        strings.TrimSpace(c.Query("status")),
		Keyword:       strings.TrimSpace(c.Query("keyword")),
	}
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		userID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || userID <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filters.UserID = &userID
	}
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		groupID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filters.GroupID = &groupID
	}
	if raw := strings.TrimSpace(c.Query("from")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			response.BadRequest(c, "Invalid from (expect RFC3339)")
			return
		}
		filters.StartTime = &t
	}
	if raw := strings.TrimSpace(c.Query("to")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			response.BadRequest(c, "Invalid to (expect RFC3339)")
			return
		}
		filters.EndTime = &t
	}

	items, pageResult, err := h.service.ListSessions(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, pageResult.Total, pageResult.Page, pageResult.PageSize)
}

// GetSession GET /admin/conversations/:id
func (h *ConversationHandler) GetSession(c *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.BadRequest(c, "Invalid conversation id")
		return
	}
	detail, err := h.service.GetSessionDetail(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}

// ExportSession GET /admin/conversations/:id/export — 单会话 .txt 下载。
func (h *ConversationHandler) ExportSession(c *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.BadRequest(c, "Invalid conversation id")
		return
	}
	filename, content, err := h.service.ExportSessionText(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "text/plain; charset=utf-8", content)
}

// ExportAll GET /admin/conversation-exports — 按当前过滤条件批量导出为 .zip。
func (h *ConversationHandler) ExportAll(c *gin.Context) {
	filters := service.ConversationSessionListFilters{
		ContextDomain: strings.TrimSpace(c.Query("context_domain")),
		Protocol:      strings.TrimSpace(c.Query("protocol")),
		Status:        strings.TrimSpace(c.Query("status")),
		Keyword:       strings.TrimSpace(c.Query("keyword")),
	}
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		userID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || userID <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filters.UserID = &userID
	}
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		groupID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filters.GroupID = &groupID
	}
	filename, content, err := h.service.ExportZip(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "application/zip", content)
}

// DeleteSession DELETE /admin/conversations/:id
func (h *ConversationHandler) DeleteSession(c *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.BadRequest(c, "Invalid conversation id")
		return
	}
	if err := h.service.DeleteSession(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
