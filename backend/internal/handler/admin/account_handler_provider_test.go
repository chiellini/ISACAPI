//go:build unit

package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type providerAssignmentAdminStub struct {
	service.AdminService
	users          map[int64]*service.User
	getUserCalls   int
	updateCalls    int
	lastAccountID  int64
	lastUpdate     *service.UpdateAccountInput
}

func (s *providerAssignmentAdminStub) GetUser(_ context.Context, id int64) (*service.User, error) {
	s.getUserCalls++
	return s.users[id], nil
}

func (s *providerAssignmentAdminStub) UpdateAccount(_ context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error) {
	s.updateCalls++
	s.lastAccountID = id
	s.lastUpdate = input
	var providerID *int64
	if input.ProviderID != nil {
		providerID = *input.ProviderID
	}
	return &service.Account{ID: id, Name: "account", Status: service.StatusActive, ProviderID: providerID}, nil
}

func newProviderAssignmentRouter(svc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.PUT("/admin/accounts/:id/provider", h.UpdateProvider)
	return router
}

func providerAssignmentRequest(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/admin/accounts/15/provider", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestAccountHandlerUpdateProviderAssignsActiveProvider(t *testing.T) {
	svc := &providerAssignmentAdminStub{users: map[int64]*service.User{
		7: {ID: 7, Role: service.RoleProvider, Status: service.StatusActive},
	}}
	router := newProviderAssignmentRouter(svc)

	response := providerAssignmentRequest(t, router, `{"provider_id":7}`)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, svc.getUserCalls)
	require.Equal(t, 1, svc.updateCalls)
	require.Equal(t, int64(15), svc.lastAccountID)
	require.NotNil(t, svc.lastUpdate)
	require.NotNil(t, svc.lastUpdate.ProviderID)
	require.NotNil(t, *svc.lastUpdate.ProviderID)
	require.Equal(t, int64(7), **svc.lastUpdate.ProviderID)
}

func TestAccountHandlerUpdateProviderClearsOwnership(t *testing.T) {
	svc := &providerAssignmentAdminStub{}
	router := newProviderAssignmentRouter(svc)

	response := providerAssignmentRequest(t, router, `{"provider_id":null}`)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Zero(t, svc.getUserCalls)
	require.Equal(t, 1, svc.updateCalls)
	require.NotNil(t, svc.lastUpdate.ProviderID)
	require.Nil(t, *svc.lastUpdate.ProviderID)
}

func TestAccountHandlerUpdateProviderRejectsInvalidAssignmentsBeforeUpdate(t *testing.T) {
	tests := []struct {
		name string
		body string
		user *service.User
	}{
		{name: "omitted", body: `{}`},
		{name: "non positive", body: `{"provider_id":0}`},
		{name: "regular user", body: `{"provider_id":7}`, user: &service.User{ID: 7, Role: service.RoleUser, Status: service.StatusActive}},
		{name: "inactive provider", body: `{"provider_id":7}`, user: &service.User{ID: 7, Role: service.RoleProvider, Status: service.StatusDisabled}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &providerAssignmentAdminStub{}
			if tt.user != nil {
				svc.users = map[int64]*service.User{tt.user.ID: tt.user}
			}
			router := newProviderAssignmentRouter(svc)

			response := providerAssignmentRequest(t, router, tt.body)

			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			require.Zero(t, svc.updateCalls)
		})
	}
}
