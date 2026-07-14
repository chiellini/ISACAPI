//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type providerHandlerAccountRepo struct {
	service.AccountRepository
	account        *service.Account
	err            error
	getCalls       int
	lastAccountID  int64
	lastProviderID int64
	updateCalls    int
	updateInput    *service.ProviderAccountUpdateInput
	updateErr      error
	deleteCalls    int
	deleteErr      error
}

func (r *providerHandlerAccountRepo) UpdateForProvider(_ context.Context, accountID, providerID int64, input *service.ProviderAccountUpdateInput) (*service.Account, error) {
	r.updateCalls++
	r.lastAccountID = accountID
	r.lastProviderID = providerID
	r.updateInput = input
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	if r.account != nil {
		return r.account, nil
	}
	return &service.Account{ID: accountID, ProviderID: &providerID, Status: service.StatusActive}, nil
}

func (r *providerHandlerAccountRepo) GetByIDForProvider(_ context.Context, accountID, providerID int64) (*service.Account, error) {
	r.getCalls++
	r.lastAccountID = accountID
	r.lastProviderID = providerID
	return r.account, r.err
}

func (r *providerHandlerAccountRepo) ListWithFiltersForProvider(
	_ context.Context,
	_ pagination.PaginationParams,
	_, _, _, _ string,
	_ int64,
	_ string,
	_ int64,
) ([]service.Account, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (r *providerHandlerAccountRepo) DeleteForProvider(_ context.Context, accountID, providerID int64) error {
	r.deleteCalls++
	r.lastAccountID = accountID
	r.lastProviderID = providerID
	return r.deleteErr
}

type providerHandlerAdminStub struct {
	service.AdminService
	createCalls int
	updateCalls int
	created     *service.CreateAccountInput
	updated     *service.UpdateAccountInput
}

type providerHandlerGroupRepo struct {
	service.GroupRepository
	groups []service.Group
}

func (r *providerHandlerGroupRepo) ListActive(_ context.Context) ([]service.Group, error) {
	return r.groups, nil
}

func (s *providerHandlerAdminStub) CreateAccount(_ context.Context, input *service.CreateAccountInput) (*service.Account, error) {
	s.createCalls++
	s.created = input
	return &service.Account{
		ID:         100,
		Name:       input.Name,
		Platform:   input.Platform,
		Type:       input.Type,
		Status:     service.StatusActive,
		ProviderID: input.ProviderID,
	}, nil
}

func (s *providerHandlerAdminStub) UpdateAccount(_ context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error) {
	s.updateCalls++
	s.updated = input
	return &service.Account{ID: id, Name: input.Name, Status: service.StatusActive}, nil
}

type providerHandlerUsageRepo struct {
	service.UsageLogRepository
	calls          int
	lastProviderID int64
	lastStart      time.Time
	lastEnd        time.Time
}

func (r *providerHandlerUsageRepo) GetProviderUsage(_ context.Context, providerID int64, startTime, endTime time.Time) (*service.ProviderUsageStats, error) {
	r.calls++
	r.lastProviderID = providerID
	r.lastStart = startTime
	r.lastEnd = endTime
	return &service.ProviderUsageStats{
		ProviderID: providerID,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

func newProviderHandlerTestRouter(h *ProviderHandler, providerID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: providerID})
		c.Next()
	})
	router.POST("/provider/accounts", h.CreateAccount)
	router.PUT("/provider/accounts/:id", h.UpdateAccount)
	router.DELETE("/provider/accounts/:id", h.DeleteAccount)
	router.GET("/provider/usage", h.GetUsage)
	router.GET("/admin/providers/:id/usage", h.GetAdminUsage)
	return router
}

func providerHandlerRequest(t *testing.T, router http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestProviderHandlerRejectsForbiddenAndUnknownAccountFields(t *testing.T) {
	accountRepo := &providerHandlerAccountRepo{account: &service.Account{
		ID: 12, ProviderID: int64Pointer(42), Platform: service.PlatformAnthropic,
	}}
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(service.NewAccountService(accountRepo, nil), nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	tests := []struct {
		name   string
		method string
		target string
		body   string
	}{
		{name: "create priority", method: http.MethodPost, target: "/provider/accounts", body: `{"name":"a","platform":"anthropic","type":"oauth","credentials":{"token":"x"},"priority":99}`},
		{name: "create provider id", method: http.MethodPost, target: "/provider/accounts", body: `{"name":"a","platform":"anthropic","type":"oauth","credentials":{"token":"x"},"provider_id":7}`},
		{name: "create proxy", method: http.MethodPost, target: "/provider/accounts", body: `{"name":"a","platform":"anthropic","type":"oauth","credentials":{"token":"x"},"proxy_id":2}`},
		{name: "create unknown", method: http.MethodPost, target: "/provider/accounts", body: `{"name":"a","platform":"anthropic","type":"oauth","credentials":{"token":"x"},"future_field":true}`},
		{name: "update load factor", method: http.MethodPut, target: "/provider/accounts/12", body: `{"load_factor":3}`},
		{name: "update multiplier", method: http.MethodPut, target: "/provider/accounts/12", body: `{"rate_multiplier":2}`},
		{name: "update schedulable", method: http.MethodPut, target: "/provider/accounts/12", body: `{"schedulable":false}`},
		{name: "update skip flag", method: http.MethodPut, target: "/provider/accounts/12", body: `{"skip_connectivity_check":true}`},
		{name: "update unknown", method: http.MethodPut, target: "/provider/accounts/12", body: `{"future_field":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCreates := admin.createCalls
			beforeUpdates := admin.updateCalls
			response := providerHandlerRequest(t, router, tt.method, tt.target, tt.body)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			require.Equal(t, beforeCreates, admin.createCalls)
			require.Equal(t, beforeUpdates, admin.updateCalls)
		})
	}
}

func TestProviderHandlerCreateUsesAuthenticatedProviderAndSafeDefaults(t *testing.T) {
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(nil, nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodPost, "/provider/accounts", `{"name":"owned","platform":"anthropic","type":"oauth","credentials":{"token":"x"}}`)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, admin.createCalls)
	require.NotNil(t, admin.created)
	require.NotNil(t, admin.created.ProviderID)
	require.Equal(t, int64(42), *admin.created.ProviderID)
	require.Equal(t, providerDefaultPriority, admin.created.Priority)
	require.NotNil(t, admin.created.LoadFactor)
	require.Equal(t, providerDefaultLoadFactor, *admin.created.LoadFactor)
	require.Equal(t, 1, admin.created.Concurrency)
	require.True(t, admin.created.SkipDefaultGroupBind)
}

func TestProviderHandlerWrongOwnerUsesAtomicScopedUpdate(t *testing.T) {
	accountRepo := &providerHandlerAccountRepo{updateErr: service.ErrAccountNotFound}
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(service.NewAccountService(accountRepo, nil), nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodPut, "/provider/accounts/12", `{"name":"forbidden"}`)

	require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
	require.Zero(t, accountRepo.getCalls, "update must not authorize with a stale preflight read")
	require.Equal(t, 1, accountRepo.updateCalls)
	require.Equal(t, int64(12), accountRepo.lastAccountID)
	require.Equal(t, int64(42), accountRepo.lastProviderID)
	require.Zero(t, admin.updateCalls)
}

func TestProviderHandlerUpdatePassesOnlyProviderAllowlistToScopedService(t *testing.T) {
	accountRepo := &providerHandlerAccountRepo{}
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(service.NewAccountService(accountRepo, nil), nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodPut, "/provider/accounts/12", `{"name":"renamed","notes":null,"credentials":{"label":"new"},"concurrency":4,"status":"inactive","group_ids":[7,7,8]}`)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, accountRepo.updateCalls)
	require.Zero(t, accountRepo.getCalls)
	require.Zero(t, admin.updateCalls)
	require.NotNil(t, accountRepo.updateInput)
	require.Equal(t, "renamed", *accountRepo.updateInput.Name)
	require.True(t, accountRepo.updateInput.NotesSet)
	require.Nil(t, accountRepo.updateInput.Notes)
	require.Equal(t, map[string]any{"label": "new"}, accountRepo.updateInput.Credentials)
	require.Equal(t, 4, *accountRepo.updateInput.Concurrency)
	require.Equal(t, "inactive", *accountRepo.updateInput.Status)
	require.Equal(t, []int64{7, 8}, *accountRepo.updateInput.GroupIDs)
}

func TestProviderHandlerCreateRejectsUnsupportedPlatform(t *testing.T) {
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(nil, nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodPost, "/provider/accounts", `{"name":"owned","platform":"custom","type":"oauth","credentials":{"token":"x"}}`)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	require.Zero(t, admin.createCalls)
}

func TestProviderHandlerCreateRejectsAPIKeyInOAuthOnlyGroup(t *testing.T) {
	admin := &providerHandlerAdminStub{}
	groupService := service.NewGroupService(&providerHandlerGroupRepo{groups: []service.Group{{
		ID: 7, Name: "oauth-only", Platform: service.PlatformAnthropic,
		Status: service.StatusActive, RequireOAuthOnly: true,
	}}}, nil)
	h := NewProviderHandler(nil, groupService, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodPost, "/provider/accounts", `{"name":"owned","platform":"anthropic","type":"apikey","credentials":{"api_key":"x"},"group_ids":[7]}`)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "GROUP_REQUIRES_OAUTH")
	require.Zero(t, admin.createCalls)
}

func TestProviderHandlerDeleteChecksOwnershipAtomicallyAtDeletePoint(t *testing.T) {
	accountRepo := &providerHandlerAccountRepo{deleteErr: service.ErrAccountNotFound}
	admin := &providerHandlerAdminStub{}
	h := NewProviderHandler(service.NewAccountService(accountRepo, nil), nil, admin, nil)
	router := newProviderHandlerTestRouter(h, 42)

	response := providerHandlerRequest(t, router, http.MethodDelete, "/provider/accounts/12", "")

	require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
	require.Zero(t, accountRepo.getCalls, "delete must not authorize with a stale preflight read")
	require.Equal(t, 1, accountRepo.deleteCalls)
	require.Equal(t, int64(12), accountRepo.lastAccountID)
	require.Equal(t, int64(42), accountRepo.lastProviderID)
}

func TestProviderHandlerUsageUsesSelfOrAdminProviderAndExactTimeRange(t *testing.T) {
	start := time.Date(2026, 7, 1, 1, 2, 3, 0, time.UTC)
	end := start.Add(6 * time.Hour)
	usageRepo := &providerHandlerUsageRepo{}
	usageService := service.NewUsageService(usageRepo, nil, nil, nil)
	h := NewProviderHandler(nil, nil, nil, usageService)
	router := newProviderHandlerTestRouter(h, 42)
	query := "?start_time=" + start.Format(time.RFC3339) + "&end_time=" + end.Format(time.RFC3339)

	selfResponse := providerHandlerRequest(t, router, http.MethodGet, "/provider/usage"+query, "")
	require.Equal(t, http.StatusOK, selfResponse.Code, selfResponse.Body.String())
	require.Equal(t, 1, usageRepo.calls)
	require.Equal(t, int64(42), usageRepo.lastProviderID)
	require.Equal(t, start, usageRepo.lastStart)
	require.Equal(t, end, usageRepo.lastEnd)

	adminResponse := providerHandlerRequest(t, router, http.MethodGet, "/admin/providers/77/usage"+query, "")
	require.Equal(t, http.StatusOK, adminResponse.Code, adminResponse.Body.String())
	require.Equal(t, 2, usageRepo.calls)
	require.Equal(t, int64(77), usageRepo.lastProviderID)
	require.Equal(t, start, usageRepo.lastStart)
	require.Equal(t, end, usageRepo.lastEnd)

	invalidResponse := providerHandlerRequest(t, router, http.MethodGet, "/provider/usage?start_time="+end.Format(time.RFC3339)+"&end_time="+start.Format(time.RFC3339), "")
	require.Equal(t, http.StatusBadRequest, invalidResponse.Code, invalidResponse.Body.String())
	require.Equal(t, 2, usageRepo.calls)
}

func int64Pointer(value int64) *int64 {
	return &value
}
