package repository

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

func TestBuildUpstreamTransportPolicy_PinsDirectDialAndRejectsPrivateIP(t *testing.T) {
	transport, err := buildUpstreamTransportPolicy(defaultPoolSettings(nil), nil, upstreamProtocolModeDefault, true)
	require.NoError(t, err)
	require.NotNil(t, transport.DialContext)

	conn, err := transport.DialContext(context.Background(), "tcp", "169.254.169.254:80")
	require.Nil(t, conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestBuildUpstreamTransportPolicy_RejectsProxy(t *testing.T) {
	proxyURL, err := url.Parse("http://proxy.example:8080")
	require.NoError(t, err)

	transport, err := buildUpstreamTransportPolicy(defaultPoolSettings(nil), proxyURL, upstreamProtocolModeDefault, true)
	require.Nil(t, transport)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be guaranteed through an upstream proxy")
}

func TestBuildTLSFingerprintTransportPolicy_SafeDialRunsBeforeTLS(t *testing.T) {
	transport, err := buildUpstreamTransportWithTLSFingerprintPolicy(
		defaultPoolSettings(nil),
		nil,
		&tlsfingerprint.Profile{Name: "test"},
		true,
	)
	require.NoError(t, err)
	require.NotNil(t, transport.DialTLSContext)
	require.NotNil(t, transport.DialContext)

	conn, err := transport.DialTLSContext(context.Background(), "tcp", "127.0.0.1:443")
	require.Nil(t, conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")

	conn, err = transport.DialContext(context.Background(), "tcp", "127.0.0.1:80")
	require.Nil(t, conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestHTTPUpstreamSecurityModeRejectsProxyBeforeRequest(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{
				Enabled:       true,
				UpstreamHosts: []string{"api.example.com"},
			},
		},
	}
	upstream := NewHTTPUpstream(cfg)
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1", nil)
	require.NoError(t, err)

	resp, err := upstream.Do(req, "http://proxy.example:8080", 1, 1)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "security.url_allowlist.trust_upstream_proxy=true")
}

func TestHTTPUpstreamExplicitTrustedProxyKeepsURLPolicy(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{
				Enabled:            true,
				TrustUpstreamProxy: true,
				UpstreamHosts:      []string{"api.example.com"},
			},
		},
	}
	svc := NewHTTPUpstream(cfg).(*httpUpstreamService)
	require.NoError(t, svc.validateProxySecurity("http://127.0.0.1:8080"))
	require.True(t, svc.shouldValidateResolvedIP(), "direct connections must remain IP-pinned")

	allowed, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1", nil)
	require.NoError(t, err)
	require.NoError(t, svc.validateRequestHost(allowed))

	blocked, err := http.NewRequest(http.MethodGet, "https://evil.example/v1", nil)
	require.NoError(t, err)
	err = svc.validateRequestHost(blocked)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host is not allowed")
}

func TestHTTPUpstreamAllowPrivateDoesNotTrustProxyOrDisableURLPolicy(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{
				Enabled:           true,
				AllowPrivateHosts: true,
				UpstreamHosts:     []string{"api.example.com"},
			},
		},
	}
	svc := NewHTTPUpstream(cfg).(*httpUpstreamService)
	err := svc.validateProxySecurity("http://127.0.0.1:8080")
	require.Error(t, err)
	require.Contains(t, err.Error(), "trust_upstream_proxy=true")

	allowed, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1", nil)
	require.NoError(t, err)
	require.NoError(t, svc.validateRequestHost(allowed))

	blocked, err := http.NewRequest(http.MethodGet, "https://evil.example/v1", nil)
	require.NoError(t, err)
	err = svc.validateRequestHost(blocked)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host is not allowed")
}

func TestHTTPUpstreamRedirectCheckerEnforcesUpstreamAllowlist(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{
				Enabled:       true,
				UpstreamHosts: []string{"api.example.com"},
			},
		},
	}
	svc := NewHTTPUpstream(cfg).(*httpUpstreamService)
	req, err := http.NewRequest(http.MethodGet, "https://evil.example/v1", nil)
	require.NoError(t, err)

	err = svc.redirectChecker(req, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host is not allowed")
}
