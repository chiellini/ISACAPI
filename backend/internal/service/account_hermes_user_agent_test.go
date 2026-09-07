//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func hermesTestAccount(accountType, platform string, enabled any, legacyUA string) *Account {
	account := &Account{
		Platform: platform,
		Type:     accountType,
		Credentials: map[string]any{
			"user_agent": legacyUA,
		},
	}
	if enabled != nil {
		account.Extra = map[string]any{HermesUserAgentExtraKey: enabled}
	}
	return account
}

func TestHermesUserAgentSwitchScopeAndLegacyUA(t *testing.T) {
	tests := []struct {
		name       string
		account    *Account
		wantEnable bool
		wantUA     string
	}{
		{
			name:       "enabled OpenAI API key uses fixed UA",
			account:    hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, true, "legacy-client/1.0"),
			wantEnable: true,
			wantUA:     HermesUserAgentValue,
		},
		{
			name:       "disabled OpenAI API key preserves legacy UA",
			account:    hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, false, "legacy-client/1.0"),
			wantUA:     "legacy-client/1.0",
		},
		{
			name:       "missing switch preserves legacy UA",
			account:    hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, nil, "legacy-client/1.0"),
			wantUA:     "legacy-client/1.0",
		},
		{
			name:       "malformed switch is disabled",
			account:    hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, "true", "legacy-client/1.0"),
			wantUA:     "legacy-client/1.0",
		},
		{
			name:       "OpenAI OAuth is never eligible",
			account:    hermesTestAccount(AccountTypeOAuth, PlatformOpenAI, true, "codex_cli_rs/0.1"),
			wantUA:     "codex_cli_rs/0.1",
		},
		{
			name:       "non OpenAI API key is never eligible",
			account:    hermesTestAccount(AccountTypeAPIKey, PlatformAnthropic, true, "claude/1.0"),
			wantUA:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantEnable, tt.account.IsHermesUserAgentEnabled())
			require.Equal(t, tt.wantUA, tt.account.GetOpenAIUserAgent())
		})
	}

	var nilAccount *Account
	require.False(t, nilAccount.IsHermesUserAgentEnabled())
}

func TestApplyHermesUserAgentRemovesCaseVariantsAndWinsOverLegacyHeaders(t *testing.T) {
	account := hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, true, "legacy-client/1.0")
	headers := http.Header{
		"User-Agent": {"OpenAI/Python 1.0"},
		"user-agent": {"openai-python-lowercase"},
		"USER-AGENT": {"openai-python-uppercase"},
		"X-Unrelated": {"preserved"},
	}

	account.ApplyHermesUserAgent(headers)

	require.Equal(t, HermesUserAgentValue, headers.Get("User-Agent"))
	require.Equal(t, []string{HermesUserAgentValue}, headers["User-Agent"])
	require.NotContains(t, headers, "user-agent")
	require.NotContains(t, headers, "USER-AGENT")
	require.Equal(t, "preserved", headers.Get("X-Unrelated"))

	// An arbitrary header override may run before this helper; the fixed switch
	// must still be authoritative.
	account.Credentials[credKeyHeaderOverrideEnabled] = true
	account.Credentials[credKeyHeaderOverrides] = map[string]any{"User-Agent": "overridden-by-admin"}
	account.ApplyHeaderOverrides(headers)
	account.ApplyHermesUserAgent(headers)
	require.Equal(t, HermesUserAgentValue, headers.Get("User-Agent"))
}

func TestApplyHermesUserAgentNoOpWhenDisabledOrNonAPIKey(t *testing.T) {
	for _, account := range []*Account{
		hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, false, "legacy"),
		hermesTestAccount(AccountTypeOAuth, PlatformOpenAI, true, "legacy"),
		hermesTestAccount(AccountTypeAPIKey, PlatformAnthropic, true, "legacy"),
	} {
		headers := make(http.Header)
		headers.Set("User-Agent", "existing")
		account.ApplyHermesUserAgent(headers)
		require.Equal(t, "existing", headers.Get("User-Agent"))
	}

	var nilAccount *Account
	headers := make(http.Header)
	headers.Set("User-Agent", "existing")
	nilAccount.ApplyHermesUserAgent(headers)
	require.Equal(t, "existing", headers.Get("User-Agent"))
}

func TestOpenAIGatewayTransportBoundaryAppliesHermesUserAgent(t *testing.T) {
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	account := hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, true, "legacy-client/1.0")
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://upstream.example/v1/responses", nil)
	require.NoError(t, err)
	request.Header.Set("User-Agent", "OpenAI/Python 1.0")

	_, err = svc.doOpenAIUpstream(request, "", account)
	require.NoError(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, HermesUserAgentValue, upstream.lastReq.Header.Get("User-Agent"))
}

func TestOpenAIBuildUpstreamRequestAppliesHermesUserAgentAfterOverrides(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"gpt-5","input":"hello"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("User-Agent", "OpenAI/Python 1.0")

	account := hermesTestAccount(AccountTypeAPIKey, PlatformOpenAI, true, "legacy-client/1.0")
	account.Credentials["api_key"] = "sk-test"
	account.Credentials["base_url"] = "https://api.liangrekui.com/v1"
	account.Credentials[credKeyHeaderOverrideEnabled] = true
	account.Credentials[credKeyHeaderOverrides] = map[string]any{"User-Agent": "admin-override/1.0"}

	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "token", false, "", false)
	require.NoError(t, err)
	require.Equal(t, "https://api.liangrekui.com/v1/responses", req.URL.String())
	require.Equal(t, HermesUserAgentValue, req.Header.Get("User-Agent"))
}
