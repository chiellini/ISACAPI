package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	providerDefaultPriority   = 50
	providerDefaultLoadFactor = 1
	providerUsageDefaultRange = 7 * 24 * time.Hour
	providerUsageMaxRange     = 30 * 24 * time.Hour
)

// ProviderHandler exposes the deliberately narrow provider self-service API.
// Scheduling strength and ownership are always supplied by trusted server code.
type ProviderHandler struct {
	accountService *service.AccountService
	groupService   *service.GroupService
	adminService   service.AdminService
	usageService   *service.UsageService
}

func NewProviderHandler(
	accountService *service.AccountService,
	groupService *service.GroupService,
	adminService service.AdminService,
	usageService *service.UsageService,
) *ProviderHandler {
	return &ProviderHandler{
		accountService: accountService,
		groupService:   groupService,
		adminService:   adminService,
		usageService:   usageService,
	}
}

// ProviderCreateAccountRequest intentionally excludes provider_id and every
// administrator-controlled scheduling/proxy field.
type ProviderCreateAccountRequest struct {
	Name        string         `json:"name" binding:"required"`
	Notes       *string        `json:"notes"`
	Platform    string         `json:"platform" binding:"required,oneof=anthropic openai gemini antigravity grok"`
	Type        string         `json:"type" binding:"required,oneof=oauth setup-token apikey upstream bedrock service_account"`
	Credentials map[string]any `json:"credentials" binding:"required"`
	Concurrency int            `json:"concurrency" binding:"omitempty,min=1"`
	GroupIDs    []int64        `json:"group_ids"`
}

type providerNullableStringField struct {
	Set   bool
	Value *string
}

func (f *providerNullableStringField) UnmarshalJSON(data []byte) error {
	f.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	f.Value = &value
	return nil
}

// ProviderUpdateAccountRequest intentionally excludes platform/type changes,
// provider assignment, proxy, priority, load factor, billing multiplier,
// schedulable, and all internal skip/confirmation flags.
type ProviderUpdateAccountRequest struct {
	Name        string                      `json:"name" binding:"omitempty,min=1"`
	Notes       providerNullableStringField `json:"notes"`
	Credentials map[string]any              `json:"credentials"`
	Concurrency *int                        `json:"concurrency" binding:"omitempty,min=1"`
	Status      string                      `json:"status" binding:"omitempty,oneof=active inactive"`
	GroupIDs    *[]int64                    `json:"group_ids"`
}

type providerGroupSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    string `json:"platform"`
	Status      string `json:"status"`
}

func providerSubject(c *gin.Context) (middleware.AuthSubject, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return middleware.AuthSubject{}, false
	}
	return subject, true
}

func bindProviderJSON(c *gin.Context, target any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body must contain exactly one JSON value")
		}
		return err
	}
	return binding.Validator.ValidateStruct(target)
}

func parseProviderAccountID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return 0, false
	}
	return id, true
}

func normalizeProviderGroupIDs(groupIDs []int64) []int64 {
	if len(groupIDs) == 0 {
		return groupIDs
	}
	seen := make(map[int64]struct{}, len(groupIDs))
	result := make([]int64, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		if _, exists := seen[groupID]; exists {
			continue
		}
		seen[groupID] = struct{}{}
		result = append(result, groupID)
	}
	return result
}

func (h *ProviderHandler) validateActivePlatformGroups(ctx context.Context, platform, accountType string, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	groups, err := h.groupService.ListActive(ctx)
	if err != nil {
		return err
	}
	activeByID := make(map[int64]service.Group, len(groups))
	for _, group := range groups {
		if group.Status == service.StatusActive {
			activeByID[group.ID] = group
		}
	}
	for _, groupID := range groupIDs {
		group, ok := activeByID[groupID]
		if groupID <= 0 || !ok {
			return service.ErrGroupNotFound
		}
		if group.Platform != platform {
			return infraerrors.BadRequest(
				"GROUP_PLATFORM_MISMATCH",
				fmt.Sprintf("group %d platform %q does not match account platform %q", groupID, group.Platform, platform),
			)
		}
		if accountType == service.AccountTypeAPIKey && group.RequireOAuthOnly {
			return infraerrors.BadRequest(
				"GROUP_REQUIRES_OAUTH",
				fmt.Sprintf("group %d only accepts non-apikey accounts", groupID),
			)
		}
	}
	return nil
}

// ListAccounts handles GET /api/v1/provider/accounts.
func (h *ProviderHandler) ListAccounts(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "name"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
	}
	accounts, result, err := h.accountService.ListWithFiltersForProvider(
		c.Request.Context(), params, c.Query("platform"), c.Query("type"), c.Query("status"), search, 0, "", subject.UserID,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items := make([]*dto.Account, 0, len(accounts))
	for i := range accounts {
		items = append(items, dto.AccountFromServiceShallow(&accounts[i]))
	}
	var total int64
	if result != nil {
		total = result.Total
	}
	response.Paginated(c, items, total, page, pageSize)
}

// GetAccount handles GET /api/v1/provider/accounts/:id.
func (h *ProviderHandler) GetAccount(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	accountID, ok := parseProviderAccountID(c)
	if !ok {
		return
	}
	account, err := h.accountService.GetByIDForProvider(c.Request.Context(), accountID, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromServiceShallow(account))
}

// CreateAccount handles POST /api/v1/provider/accounts.
func (h *ProviderHandler) CreateAccount(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	var req ProviderCreateAccountRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	req.GroupIDs = normalizeProviderGroupIDs(req.GroupIDs)
	if err := h.validateActivePlatformGroups(c.Request.Context(), req.Platform, req.Type, req.GroupIDs); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	loadFactor := providerDefaultLoadFactor
	providerID := subject.UserID
	account, err := h.adminService.CreateAccount(c.Request.Context(), &service.CreateAccountInput{
		Name:         req.Name,
		Notes:        req.Notes,
		Platform:     req.Platform,
		Type:         req.Type,
		Credentials:  req.Credentials,
		Concurrency:  concurrency,
		Priority:     providerDefaultPriority,
		LoadFactor:   &loadFactor,
		GroupIDs:     req.GroupIDs,
		ProviderID:   &providerID,
		SkipDefaultGroupBind: true,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromServiceShallow(account))
}

// UpdateAccount handles PUT /api/v1/provider/accounts/:id.
func (h *ProviderHandler) UpdateAccount(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	accountID, ok := parseProviderAccountID(c)
	if !ok {
		return
	}
	var req ProviderUpdateAccountRequest
	if err := bindProviderJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.GroupIDs != nil {
		groupIDs := normalizeProviderGroupIDs(*req.GroupIDs)
		req.GroupIDs = &groupIDs
	}
	var name *string
	if req.Name != "" {
		name = &req.Name
	}
	var status *string
	if req.Status != "" {
		status = &req.Status
	}
	updated, err := h.accountService.UpdateForProvider(c.Request.Context(), accountID, subject.UserID, &service.ProviderAccountUpdateInput{
		Name:        name,
		NotesSet:    req.Notes.Set,
		Notes:       req.Notes.Value,
		Credentials: req.Credentials,
		Concurrency: req.Concurrency,
		Status:      status,
		GroupIDs:    req.GroupIDs,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AccountFromServiceShallow(updated))
}

// DeleteAccount handles DELETE /api/v1/provider/accounts/:id.
func (h *ProviderHandler) DeleteAccount(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	accountID, ok := parseProviderAccountID(c)
	if !ok {
		return
	}
	if err := h.accountService.DeleteForProvider(c.Request.Context(), accountID, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Account deleted successfully"})
}

// ListGroups handles GET /api/v1/provider/groups.
func (h *ProviderHandler) ListGroups(c *gin.Context) {
	if _, ok := providerSubject(c); !ok {
		return
	}
	groups, err := h.groupService.ListActive(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := make([]providerGroupSummary, 0, len(groups))
	for _, group := range groups {
		if group.Status != service.StatusActive {
			continue
		}
		result = append(result, providerGroupSummary{
			ID: group.ID, Name: group.Name, Description: group.Description,
			Platform: group.Platform, Status: group.Status,
		})
	}
	response.Success(c, result)
}

// GetUsage handles GET /api/v1/provider/usage.
func (h *ProviderHandler) GetUsage(c *gin.Context) {
	subject, ok := providerSubject(c)
	if !ok {
		return
	}
	h.writeUsage(c, subject.UserID)
}

// GetAdminUsage handles GET /api/v1/admin/providers/:id/usage.
func (h *ProviderHandler) GetAdminUsage(c *gin.Context) {
	providerID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || providerID <= 0 {
		response.BadRequest(c, "Invalid provider ID")
		return
	}
	h.writeUsage(c, providerID)
}

func (h *ProviderHandler) writeUsage(c *gin.Context, providerID int64) {
	startTime, endTime, err := parseProviderUsageTimeRange(c)
	if err != nil {
		response.BadRequest(c, "Invalid time range: "+err.Error())
		return
	}
	stats, err := h.usageService.GetProviderUsage(c.Request.Context(), providerID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

func parseProviderUsageTimeRange(c *gin.Context) (time.Time, time.Time, error) {
	parse := func(raw string) (time.Time, error) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return time.Time{}, nil
		}
		if value, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			return value, nil
		}
		return time.Parse(time.RFC3339, raw)
	}
	startTime, err := parse(c.Query("start_time"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endTime, err := parse(c.Query("end_time"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}
	if startTime.IsZero() {
		startTime = endTime.Add(-providerUsageDefaultRange)
	}
	if !startTime.Before(endTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_time must be before end_time")
	}
	if endTime.Sub(startTime) > providerUsageMaxRange {
		return time.Time{}, time.Time{}, fmt.Errorf("max window is 30 days")
	}
	return startTime, endTime, nil
}
