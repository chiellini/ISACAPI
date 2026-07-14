package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authResponseResearchGroupRepo struct {
	service.ResearchGroupRepository
	group *service.ResearchGroup
}

func (r *authResponseResearchGroupRepo) GetByOwnerUserID(context.Context, int64) (*service.ResearchGroup, error) {
	return r.group, nil
}

func newResearchGroupAuthResponseHandler() *AuthHandler {
	ownerBalance := 42.5
	researchGroupService := service.NewResearchGroupService(&authResponseResearchGroupRepo{
		group: &service.ResearchGroup{
			ID:           17,
			Name:         "AI Lab",
			OwnerUserID:  11,
			OwnerBalance: &ownerBalance,
			Status:       service.ResearchGroupStatusActive,
		},
	}, nil)
	return &AuthHandler{researchGroupService: researchGroupService}
}

func authResponseResearchGroupOwner() *service.User {
	return &service.User{
		ID:       11,
		Email:    "owner@example.com",
		Username: "owner",
		Role:     service.RoleUser,
		Status:   service.StatusActive,
	}
}

func TestAuthResponseFromTokenPairIncludesResearchGroupContext(t *testing.T) {
	h := newResearchGroupAuthResponseHandler()

	payload, err := h.authResponseFromTokenPair(context.Background(), authResponseResearchGroupOwner(), &service.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	})
	require.NoError(t, err)
	require.Equal(t, "access-token", payload.AccessToken)
	require.NotNil(t, payload.User)
	require.NotNil(t, payload.User.ResearchGroup)
	require.Equal(t, "owner", payload.User.ResearchGroup.Role)
	require.Equal(t, int64(17), payload.User.ResearchGroup.Group.ID)
}

func TestWriteOAuthTokenPairResponseUsesUnifiedAuthResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newResearchGroupAuthResponseHandler()
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest("POST", "/api/v1/auth/oauth/oidc/complete-registration", nil)

	h.writeOAuthTokenPairResponse(ginContext, authResponseResearchGroupOwner(), &service.TokenPair{
		AccessToken:  "oauth-access-token",
		RefreshToken: "oauth-refresh-token",
		ExpiresIn:    7200,
	})

	require.Equal(t, 200, recorder.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, "oauth-access-token", payload["access_token"])
	user, ok := payload["user"].(map[string]any)
	require.True(t, ok)
	researchGroup, ok := user["research_group"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "owner", researchGroup["role"])
}
