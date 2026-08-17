package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestClientControlledSecurityAuditBodyExcludesInjectedChatPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	clientBody := []byte(`{"model":"public-gpt","messages":[{"role":"user","content":"hello"}]}`)
	injectedBody := []byte(`{"model":"gpt-upstream","messages":[{"role":"system","content":"trusted secret prompt"}]}`)

	request := httptest.NewRequest("POST", "/api/v1/chat/v1/chat/completions", nil)
	request = request.WithContext(context.WithValue(request.Context(), ctxkey.ChatSecurityAuditBody, clientBody))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = request

	got := clientControlledSecurityAuditBody(c, injectedBody)
	require.Equal(t, clientBody, got)
	require.NotContains(t, string(got), "trusted secret prompt")
}
