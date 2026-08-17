package securityaudit

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeBaseURLUsesSecureDefaults(t *testing.T) {
	t.Setenv(promptAuditAllowInsecureHTTPEnv, "")
	t.Setenv(promptAuditAllowPrivateEndpointsEnv, "")
	allowed := []string{
		"https://guard.example.com", "https://guard.example.com/v1", "https://guard.example.com/audit",
	}
	for _, raw := range allowed {
		_, err := NormalizeBaseURL(raw)
		require.NoError(t, err, raw)
	}
	blocked := []string{
		"http://guard.example.com", "ftp://guard.example.com", "https://user:pass@guard.example.com",
		"https://guard.example.com?q=secret", "https://guard.example.com/#fragment",
		"https://127.0.0.1:8080", "https://[::1]", "https://10.0.0.8:8080", "https://172.16.0.5",
		"https://169.254.169.254", "https://168.63.129.16", "https://100.100.100.200", "https://metadata.google.internal",
		"https://localhost", "https://guard.local:8080",
	}
	for _, raw := range blocked {
		_, err := NormalizeBaseURL(raw)
		require.Error(t, err, raw)
	}
	url, err := ChatCompletionsURL("https://guard.example.com/v1")
	require.NoError(t, err)
	require.Equal(t, "https://guard.example.com/v1/chat/completions", url)
}

func TestNormalizeBaseURLAllowsExplicitPrivateHTTPOptIn(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	allowed := []string{
		"http://127.0.0.1:8080", "http://10.0.0.8:8080", "https://172.16.0.5",
		"http://internal-admin.local", "http://guard.local:8080",
	}
	for _, raw := range allowed {
		_, err := NormalizeBaseURL(raw)
		require.NoError(t, err, raw)
	}
	for _, raw := range []string{"http://169.254.169.254", "https://168.63.129.16", "https://100.100.100.200", "https://metadata.google.internal"} {
		_, err := NormalizeBaseURL(raw)
		require.Error(t, err, raw)
	}
}

func TestHTTPClientUsesPinnedSecureDialer(t *testing.T) {
	client, err := NewSecureHTTPClient(ActiveEndpoint{BaseURL: "https://guard.example.com", TimeoutMS: 1000})
	require.NoError(t, err)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.Nil(t, transport.Proxy)
	require.NotNil(t, transport.DialContext)
	require.NotNil(t, client.CheckRedirect)
}

func TestPromptAuditSecureDialerRejectsMixedDNSAnswersBeforeDial(t *testing.T) {
	var dialCalls atomic.Int64
	dialer := &promptAuditSecureDialer{
		resolver: staticPromptAuditResolver{addresses: []netip.Addr{
			netip.MustParseAddr("93.184.216.34"),
			netip.MustParseAddr("10.0.0.8"),
		}},
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			dialCalls.Add(1)
			return nil, errors.New("unexpected dial")
		},
	}

	_, err := dialer.DialContext(context.Background(), "tcp", "guard.example.com:443")
	require.Error(t, err)
	require.Zero(t, dialCalls.Load())
}

func TestPromptAuditSecureDialerPinsValidatedDNSAddress(t *testing.T) {
	stop := errors.New("stop after recording dial")
	var dialed string
	dialer := &promptAuditSecureDialer{
		resolver: staticPromptAuditResolver{addresses: []netip.Addr{netip.MustParseAddr("93.184.216.34")}},
		dialContext: func(_ context.Context, _ string, address string) (net.Conn, error) {
			dialed = address
			return nil, stop
		},
	}

	_, err := dialer.DialContext(context.Background(), "tcp", "guard.example.com:443")
	require.ErrorIs(t, err, stop)
	require.Equal(t, "93.184.216.34:443", dialed)
}

func TestOpenAICompatibleScannerRequestContract(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, DefaultGuardModel, payload["model"])
		require.Equal(t, float64(0), payload["temperature"])
		require.Equal(t, float64(64), payload["max_tokens"])
		require.Equal(t, float64(42), payload["seed"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Safety: Safe\nCategories: None"}}]}`))
	}))
	defer server.Close()
	scanner := NewOpenAICompatibleScanner()
	result, err := scanner.Scan(context.Background(), ActiveEndpoint{ID: "one", BaseURL: server.URL, Model: DefaultGuardModel, Token: "token", TimeoutMS: 1000}, "hello", AllScannerIDs)
	require.NoError(t, err)
	require.Equal(t, EventPass, result.Decision)
}

func TestOpenAICompatibleScannerAllowsSameOriginRedirect(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			http.Redirect(w, r, "/guard-result", http.StatusTemporaryRedirect)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Safety: Safe\nCategories: None"}}]}`))
	}))
	defer server.Close()
	result, err := NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "redirect", BaseURL: server.URL, Model: DefaultGuardModel, TimeoutMS: 1000}, "hello", AllScannerIDs)
	require.NoError(t, err)
	require.Equal(t, EventPass, result.Decision)
}

func TestOpenAICompatibleScannerRejectsCrossOriginRedirect(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	var targetCalls atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		targetCalls.Add(1)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Safety: Safe\nCategories: None"}}]}`))
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()

	_, err := NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "redirect", BaseURL: redirect.URL, Model: DefaultGuardModel, TimeoutMS: 1000}, "hello", AllScannerIDs)
	require.Error(t, err)
	require.Zero(t, targetCalls.Load())
}

func TestOpenAICompatibleScannerRejectsOversizeResponse(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	oversize := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(maxGuardResponseBytes)+1)))
	}))
	defer oversize.Close()
	_, err := NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "large", BaseURL: oversize.URL, Model: DefaultGuardModel, TimeoutMS: 1000}, "hello", AllScannerIDs)
	require.Error(t, err)
}

func TestOpenAICompatibleScannerClassifiesHTTPConnectionAndTimeoutFailures(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	tests := []struct {
		name      string
		status    int
		retryable bool
	}{
		{name: "authentication", status: http.StatusUnauthorized, retryable: false},
		{name: "forbidden", status: http.StatusForbidden, retryable: false},
		{name: "rate limited", status: http.StatusTooManyRequests, retryable: true},
		{name: "server failure", status: http.StatusBadGateway, retryable: true},
		{name: "other client error", status: http.StatusBadRequest, retryable: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()
			_, err := NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "status", BaseURL: server.URL, Model: DefaultGuardModel, TimeoutMS: 1000}, "hello", AllScannerIDs)
			var guardErr *GuardError
			require.ErrorAs(t, err, &guardErr)
			require.Equal(t, ErrorCodeUnavailable, guardErr.Code)
			require.Equal(t, tt.status, guardErr.HTTPStatus)
			require.Equal(t, tt.retryable, guardErr.Retryable)
			require.NotContains(t, err.Error(), server.URL)
		})
	}

	closed := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	_, err := NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "closed", BaseURL: closedURL, Model: DefaultGuardModel, TimeoutMS: 100}, "hello", AllScannerIDs)
	var connectionErr *GuardError
	require.ErrorAs(t, err, &connectionErr)
	require.True(t, connectionErr.Retryable)

	timeout := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer timeout.Close()
	_, err = NewOpenAICompatibleScanner().Scan(context.Background(), ActiveEndpoint{ID: "timeout", BaseURL: timeout.URL, Model: DefaultGuardModel, TimeoutMS: 20}, "hello", AllScannerIDs)
	var timeoutErr *GuardError
	require.ErrorAs(t, err, &timeoutErr)
	require.True(t, timeoutErr.Retryable)
	require.True(t, timeoutErr.Timeout)
}

func TestPromptAuditProbeModelsFallbackAndResponseSafety(t *testing.T) {
	allowLocalPromptAuditEndpoint(t)
	t.Run("models contains configured model", func(t *testing.T) {
		var chatCalls atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "Bearer temporary-token", r.Header.Get("Authorization"))
			if r.URL.Path == "/v1/models" {
				_, _ = w.Write([]byte(`{"data":[{"id":"` + DefaultGuardModel + `"}]}`))
				return
			}
			chatCalls.Add(1)
		}))
		defer server.Close()
		result := newProbeTestService().Probe(context.Background(), ProbeRequest{Endpoint: probeEndpoint(server.URL, "temporary-token")})
		require.True(t, result.OK)
		require.True(t, result.TokenApplied)
		require.Equal(t, http.StatusOK, result.HTTPStatus)
		require.Zero(t, chatCalls.Load())
	})

	t.Run("invalid models response performs real guard fallback", func(t *testing.T) {
		var chatCalls atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/models" {
				_, _ = w.Write([]byte(`{"unexpected":true}`))
				return
			}
			chatCalls.Add(1)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Safety: Safe\nCategories: None"}}]}`))
		}))
		defer server.Close()
		result := newProbeTestService().Probe(context.Background(), ProbeRequest{Endpoint: probeEndpoint(server.URL, "temporary-token")})
		require.True(t, result.OK)
		require.Equal(t, int64(1), chatCalls.Load())
	})

	t.Run("fallback authentication failure is stable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/models" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()
		result := newProbeTestService().Probe(context.Background(), ProbeRequest{Endpoint: probeEndpoint(server.URL, "temporary-token")})
		require.False(t, result.OK)
		require.Equal(t, ErrorCodeUnavailable, result.ErrorCode)
		require.Equal(t, http.StatusUnauthorized, result.HTTPStatus)
		require.False(t, result.Retryable)
	})

	t.Run("oversized models response is rejected without fallback", func(t *testing.T) {
		var chatCalls atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/models" {
				chatCalls.Add(1)
			}
			_, _ = w.Write([]byte(strings.Repeat("x", int(maxGuardResponseBytes)+1)))
		}))
		defer server.Close()
		result := newProbeTestService().Probe(context.Background(), ProbeRequest{Endpoint: probeEndpoint(server.URL, "temporary-token")})
		require.False(t, result.OK)
		require.Equal(t, "response_too_large", result.ErrorCode)
		require.Zero(t, chatCalls.Load())
	})
}

func TestResolveProbeEndpointReusesTokenOnlyForMatchingBaseURL(t *testing.T) {
	manager := &ConfigManager{}
	manager.snapshot.Store(&activeConfigSnapshot{active: ActiveConfig{Endpoints: []ActiveEndpoint{{
		ID: "guard-1", BaseURL: "https://guard.example.com", Token: "STORED_GUARD_TOKEN", TimeoutMS: 1000, InputLimit: 1024, Enabled: true,
	}}}})
	service := &PromptService{config: manager}

	matched, applied, err := service.resolveProbeEndpoint(UpdateEndpoint{
		ID: "guard-1", BaseURL: "https://guard.example.com/v1", TimeoutMS: 1000, InputLimit: 1024,
	})
	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, "STORED_GUARD_TOKEN", matched.Token)

	mismatched, applied, err := service.resolveProbeEndpoint(UpdateEndpoint{
		ID: "guard-1", BaseURL: "https://attacker.example.com", TimeoutMS: 1000, InputLimit: 1024,
	})
	require.NoError(t, err)
	require.False(t, applied)
	require.Empty(t, mismatched.Token)
}

func newProbeTestService() *PromptService {
	return &PromptService{
		config: &ConfigManager{}, scanner: NewOpenAICompatibleScanner(), clock: realClock{},
		probes: map[string]ProbeResult{},
	}
}

func probeEndpoint(baseURL, token string) UpdateEndpoint {
	return UpdateEndpoint{
		ID: "probe-one", Name: "Probe One", Protocol: "openai_compatible", BaseURL: baseURL,
		Model: DefaultGuardModel, Token: token, TimeoutMS: 1000, InputLimit: 1024, Enabled: true,
	}
}

func allowLocalPromptAuditEndpoint(t *testing.T) {
	t.Helper()
	t.Setenv(promptAuditAllowInsecureHTTPEnv, "true")
	t.Setenv(promptAuditAllowPrivateEndpointsEnv, "true")
}

type staticPromptAuditResolver struct {
	addresses []netip.Addr
	err       error
}

func (r staticPromptAuditResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.addresses, r.err
}
