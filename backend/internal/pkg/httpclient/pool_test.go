package httpclient

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildTransport_SafeDirectDialRejectsPrivateLiteral(t *testing.T) {
	transport, err := buildTransport(Options{ValidateResolvedIP: true})
	require.NoError(t, err)
	require.NotNil(t, transport.DialContext)

	conn, err := transport.DialContext(context.Background(), "tcp", "169.254.169.254:80")
	require.Nil(t, conn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestBuildTransport_ProxyModeRemainsAvailableForTrustedFixedTargetCallers(t *testing.T) {
	for _, proxyURL := range []string{
		"http://proxy.example:8080",
		"https://proxy.example:8443",
		"socks5h://proxy.example:1080",
	} {
		t.Run(strings.SplitN(proxyURL, ":", 2)[0], func(t *testing.T) {
			transport, err := buildTransport(Options{
				ProxyURL:           proxyURL,
				ValidateResolvedIP: true,
			})
			require.NoError(t, err)
			require.NotNil(t, transport)
		})
	}
}

func TestBuildTransport_TrustedNetworkOptOutAllowsProxy(t *testing.T) {
	transport, err := buildTransport(Options{
		ProxyURL:           "http://proxy.example:8080",
		ValidateResolvedIP: true,
		AllowPrivateHosts:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, transport)
	require.NotNil(t, transport.Proxy)
}
