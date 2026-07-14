package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ResearchGroupHandler struct {
	service *service.ResearchGroupService
}

func NewResearchGroupHandler(service *service.ResearchGroupService) *ResearchGroupHandler {
	return &ResearchGroupHandler{service: service}
}

func adminResearchGroupID(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid "+key)
		return 0, false
	}
	return id, true
}

func adminResearchGroupActor(c *gin.Context) int64 {
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		return subject.UserID
	}
	return 0
}

func (h *ResearchGroupHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	result, err := h.service.AdminList(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ResearchGroupHandler) GetByID(c *gin.Context) {
	groupID, ok := adminResearchGroupID(c, "id")
	if !ok {
		return
	}
	result, err := h.service.AdminGet(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ResearchGroupHandler) Dissolve(c *gin.Context) {
	groupID, ok := adminResearchGroupID(c, "id")
	if !ok {
		return
	}
	if err := h.service.AdminDissolve(c.Request.Context(), adminResearchGroupActor(c), groupID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "research group dissolved"})
}

func (h *ResearchGroupHandler) DetachMember(c *gin.Context) {
	groupID, ok := adminResearchGroupID(c, "id")
	if !ok {
		return
	}
	memberID, ok := adminResearchGroupID(c, "member_id")
	if !ok {
		return
	}
	if err := h.service.AdminDetach(c.Request.Context(), adminResearchGroupActor(c), groupID, memberID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "research group member detached"})
}
