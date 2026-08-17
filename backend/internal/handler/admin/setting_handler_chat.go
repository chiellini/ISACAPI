package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GetChatPolicy returns the full trusted prompt/skill configuration. The route
// is super-admin-only because these instructions define privileged model behavior.
func (h *SettingHandler) GetChatPolicy(c *gin.Context) {
	policy, err := h.settingService.GetChatPolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}

// UpdateChatPolicy replaces the complete policy atomically after validation.
func (h *SettingHandler) UpdateChatPolicy(c *gin.Context) {
	var policy service.ChatPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.settingService.UpdateChatPolicy(c.Request.Context(), &policy); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updated, err := h.settingService.GetChatPolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, updated)
}
