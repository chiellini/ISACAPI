package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ResearchGroupHandler struct {
	service *service.ResearchGroupService
}

func NewResearchGroupHandler(service *service.ResearchGroupService) *ResearchGroupHandler {
	return &ResearchGroupHandler{service: service}
}

type createResearchGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateResearchGroupRequest struct {
	Name   *string `json:"name"`
	Status *string `json:"status"`
}

type inviteResearchGroupMemberRequest struct {
	Email           string  `json:"email" binding:"required,email"`
	MonthlyLimitUSD float64 `json:"monthly_limit_usd"`
}

type updateResearchGroupMemberRequest struct {
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd"`
	Status          *string  `json:"status"`
}

func researchGroupSubject(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}

func researchGroupPathID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid "+name)
		return 0, false
	}
	return id, true
}

func (h *ResearchGroupHandler) GetContext(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	context, err := h.service.GetContext(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, context)
}

func (h *ResearchGroupHandler) Create(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	var request createResearchGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	context, err := h.service.Create(c.Request.Context(), userID, request.Name)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, context)
}

func (h *ResearchGroupHandler) Update(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	var request updateResearchGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	context, err := h.service.Update(c.Request.Context(), userID, request.Name, request.Status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, context)
}

func (h *ResearchGroupHandler) Dissolve(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	if err := h.service.Dissolve(c.Request.Context(), userID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "research group dissolved"})
}

func (h *ResearchGroupHandler) InviteMember(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	var request inviteResearchGroupMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	member, err := h.service.Invite(c.Request.Context(), userID, request.Email, request.MonthlyLimitUSD)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, member)
}

func (h *ResearchGroupHandler) UpdateMember(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	memberID, ok := researchGroupPathID(c, "id")
	if !ok {
		return
	}
	var request updateResearchGroupMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	member, err := h.service.UpdateMember(c.Request.Context(), userID, memberID, request.MonthlyLimitUSD, request.Status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, member)
}

func (h *ResearchGroupHandler) ResetMemberMonth(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	memberID, ok := researchGroupPathID(c, "id")
	if !ok {
		return
	}
	member, err := h.service.ResetMemberMonth(c.Request.Context(), userID, memberID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, member)
}

func (h *ResearchGroupHandler) RemoveMember(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	memberID, ok := researchGroupPathID(c, "id")
	if !ok {
		return
	}
	if err := h.service.RemoveMember(c.Request.Context(), userID, memberID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "research group member removed"})
}

func (h *ResearchGroupHandler) ListInvitations(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	invitations, err := h.service.ListInvitations(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, invitations)
}

func (h *ResearchGroupHandler) AcceptInvitation(c *gin.Context) {
	h.respondInvitation(c, true)
}

func (h *ResearchGroupHandler) RejectInvitation(c *gin.Context) {
	h.respondInvitation(c, false)
}

func (h *ResearchGroupHandler) respondInvitation(c *gin.Context, accept bool) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	memberID, ok := researchGroupPathID(c, "id")
	if !ok {
		return
	}
	context, err := h.service.RespondInvitation(c.Request.Context(), userID, memberID, accept)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, context)
}

func (h *ResearchGroupHandler) Leave(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	if err := h.service.Leave(c.Request.Context(), userID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "left research group"})
}

func (h *ResearchGroupHandler) ListUsage(c *gin.Context) {
	userID, ok := researchGroupSubject(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	filter := service.ResearchGroupUsageFilter{Page: page, PageSize: pageSize}
	if raw := c.Query("member_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid member_id")
			return
		}
		filter.MemberID = &id
	}
	for key, target := range map[string]**time.Time{"start_time": &filter.Start, "end_time": &filter.End} {
		if raw := c.Query(key); raw != "" {
			value, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				response.BadRequest(c, "Invalid "+key)
				return
			}
			*target = &value
		}
	}
	result, err := h.service.ListUsage(c.Request.Context(), userID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
