package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDetectUpstreamContentPolicyRejection(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		want     bool
		wantCode string
	}{
		{
			name:   "anthropic content policy violation",
			status: http.StatusForbidden,
			body:   `{"error":{"code":"content_policy_violation","message":"请求涉及受限话题，已被内容安全策略阻止","type":"permission_error"},"type":"error"}`,
			want:   true, wantCode: "content_policy_violation",
		},
		{
			name:   "image moderation blocked",
			status: http.StatusBadRequest,
			body:   `{"error":{"type":"image_generation_user_error","code":"moderation_blocked","message":"rejected by safety"}}`,
			want:   true, wantCode: "moderation_blocked",
		},
		{
			name:   "sse content filter",
			status: http.StatusOK,
			body:   "data: {\"type\":\"error\",\"error\":{\"code\":\"content_filter\",\"message\":\"blocked by content policy\"}}\n\n",
			want:   true, wantCode: "content_filter",
		},
		{
			name:   "image incomplete reason carried in error message",
			status: http.StatusBadRequest,
			body:   `{"error":{"type":"image_generation_user_error","code":"response_incomplete","message":"Upstream image generation incomplete: content_filter"}}`,
			want:   true, wantCode: "content_filter",
		},
		{
			name:   "generic permission denied",
			status: http.StatusForbidden,
			body:   `{"error":{"type":"permission_error","message":"account is not entitled to this model"}}`,
		},
		{
			name:   "account policy suspension",
			status: http.StatusForbidden,
			body:   `{"error":{"code":"account_suspended","reason":"policy_violation","message":"account suspended due to policy violation"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DetectUpstreamContentPolicyRejection(tt.status, []byte(tt.body))
			require.Equal(t, tt.want, ok)
			if tt.want {
				require.Equal(t, tt.wantCode, got.Code)
			}
		})
	}
}

func TestRecordOpenAIStreamUpstreamErrorPreservesPolicyCodeWhenBodyLoggingDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	payload := []byte(`{"type":"response.failed","response":{"error":{"code":"safety_error","message":"blocked"}}}`)

	(&OpenAIGatewayService{}).recordOpenAIStreamUpstreamError(c, nil, true, "req-policy", "http_error", payload, "blocked")

	stored, exists := c.Get(OpsUpstreamPolicyPayloadKey)
	require.True(t, exists)
	require.JSONEq(t, string(payload), stored.(string))
	_, generalDetailExists := c.Get(OpsUpstreamErrorDetailKey)
	require.False(t, generalDetailExists)
}

func TestResetOpsUpstreamErrorAttemptContextClearsFlatPolicyFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(OpsUpstreamStatusCodeKey, http.StatusForbidden)
	c.Set(OpsUpstreamErrorMessageKey, "first account refused")
	c.Set(OpsUpstreamErrorDetailKey, `{"error":{"code":"content_policy_violation"}}`)
	c.Set(OpsUpstreamPolicyPayloadKey, `{"error":{"code":"content_policy_violation"}}`)

	ResetOpsUpstreamErrorAttemptContext(c)

	for _, key := range []string{
		OpsUpstreamStatusCodeKey,
		OpsUpstreamErrorMessageKey,
		OpsUpstreamErrorDetailKey,
		OpsUpstreamPolicyPayloadKey,
	} {
		value, _ := c.Get(key)
		require.Nil(t, value, key)
	}
}
