package urlvalidator

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type staticResolver struct {
	addresses []net.IPAddr
	err       error
	calls     int
}

func (r *staticResolver) LookupIPAddr(_ context.Context, _ string) ([]net.IPAddr, error) {
	r.calls++
	return r.addresses, r.err
}

type recordingDialer struct {
	addresses []string
	err       error
}

func (d *recordingDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.addresses = append(d.addresses, address)
	return nil, d.err
}

type tlsCaptureDialer struct {
	sni chan string
}

type happyEyeballsDialer struct {
	mu        sync.Mutex
	addresses []string
}

func (d *happyEyeballsDialer) DialContext(ctx context.Context, _, address string) (net.Conn, error) {
	d.mu.Lock()
	d.addresses = append(d.addresses, address)
	d.mu.Unlock()
	if strings.HasPrefix(address, "[") {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	client, server := net.Pipe()
	go func() { _ = server.Close() }()
	return client, nil
}

func (d *tlsCaptureDialer) DialContext(_ context.Context, _, _ string) (net.Conn, error) {
	clientConn, serverConn := net.Pipe()
	go func() {
		defer func() { _ = serverConn.Close() }()
		server := tls.Server(serverConn, &tls.Config{
			MinVersion: tls.VersionTLS12,
			GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
				d.sni <- hello.ServerName
				return nil, errors.New("stop after capturing SNI")
			},
		})
		_ = server.Handshake()
	}()
	return clientConn, nil
}

func TestValidateURLFormat(t *testing.T) {
	if _, err := ValidateURLFormat("", false); err == nil {
		t.Fatalf("expected empty url to fail")
	}
	if _, err := ValidateURLFormat("://bad", false); err == nil {
		t.Fatalf("expected invalid url to fail")
	}
	if _, err := ValidateURLFormat("https://user:super-secret@", false); err == nil || strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("expected invalid URL error to redact credentials, got %v", err)
	}
	if _, err := ValidateURLFormat("http://example.com", false); err == nil {
		t.Fatalf("expected http to fail when allow_insecure_http is false")
	}
	if _, err := ValidateURLFormat("https://example.com", false); err != nil {
		t.Fatalf("expected https to pass, got %v", err)
	}
	if _, err := ValidateURLFormat("http://example.com", true); err != nil {
		t.Fatalf("expected http to pass when allow_insecure_http is true, got %v", err)
	}
	if _, err := ValidateURLFormat("https://example.com:bad", true); err == nil {
		t.Fatalf("expected invalid port to fail")
	}

	// 验证末尾斜杠被移除
	normalized, err := ValidateURLFormat("https://example.com/", false)
	if err != nil {
		t.Fatalf("expected trailing slash url to pass, got %v", err)
	}
	if normalized != "https://example.com" {
		t.Fatalf("expected trailing slash to be removed, got %s", normalized)
	}

	// 验证多个末尾斜杠被移除
	normalized, err = ValidateURLFormat("https://example.com///", false)
	if err != nil {
		t.Fatalf("expected multiple trailing slashes to pass, got %v", err)
	}
	if normalized != "https://example.com" {
		t.Fatalf("expected all trailing slashes to be removed, got %s", normalized)
	}

	// 验证带路径的 URL 末尾斜杠被移除
	normalized, err = ValidateURLFormat("https://example.com/api/v1/", false)
	if err != nil {
		t.Fatalf("expected trailing slash url with path to pass, got %v", err)
	}
	if normalized != "https://example.com/api/v1" {
		t.Fatalf("expected trailing slash to be removed from path, got %s", normalized)
	}
}

func TestValidateHTTPURL(t *testing.T) {
	if _, err := ValidateHTTPURL("http://example.com", false, ValidationOptions{}); err == nil {
		t.Fatalf("expected http to fail when allow_insecure_http is false")
	}
	if _, err := ValidateHTTPURL("http://example.com", true, ValidationOptions{}); err != nil {
		t.Fatalf("expected http to pass when allow_insecure_http is true, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://example.com", false, ValidationOptions{RequireAllowlist: true}); err == nil {
		t.Fatalf("expected require allowlist to fail when empty")
	}
	if _, err := ValidateHTTPURL("https://example.com", false, ValidationOptions{AllowedHosts: []string{"api.example.com"}}); err == nil {
		t.Fatalf("expected host not in allowlist to fail")
	}
	if _, err := ValidateHTTPURL("https://api.example.com", false, ValidationOptions{AllowedHosts: []string{"api.example.com"}}); err != nil {
		t.Fatalf("expected allowlisted host to pass, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://sub.api.example.com", false, ValidationOptions{AllowedHosts: []string{"*.example.com"}}); err != nil {
		t.Fatalf("expected wildcard allowlist to pass, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://bedrock-runtime.us-east-1.amazonaws.com", false, ValidationOptions{AllowedHosts: []string{"bedrock-runtime.*.amazonaws.com"}}); err != nil {
		t.Fatalf("expected single-label regional wildcard to pass, got %v", err)
	}
	if _, err := ValidateHTTPURL("https://evil.us-east-1.amazonaws.com", false, ValidationOptions{AllowedHosts: []string{"bedrock-runtime.*.amazonaws.com"}}); err == nil {
		t.Fatal("regional wildcard must not match a different service")
	}
	if _, err := ValidateHTTPURL("https://bedrock-runtime.extra.us-east-1.amazonaws.com", false, ValidationOptions{AllowedHosts: []string{"bedrock-runtime.*.amazonaws.com"}}); err == nil {
		t.Fatal("regional wildcard must match exactly one DNS label")
	}
	if _, err := ValidateHTTPURL("https://localhost", false, ValidationOptions{AllowPrivate: false}); err == nil {
		t.Fatalf("expected localhost to be blocked when allow_private_hosts is false")
	}
	if _, err := ValidateHTTPURL("https://168.63.129.16", false, ValidationOptions{AllowPrivate: false}); err == nil {
		t.Fatalf("expected cloud metadata/platform address to be blocked")
	}
}

func TestSafeDialerPinsValidatedIPAddress(t *testing.T) {
	resolver := &staticResolver{addresses: []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}}
	dialer := &recordingDialer{err: errors.New("test dial stopped")}
	safeDialer := NewSafeDialer(resolver, dialer, time.Second)

	_, err := safeDialer.DialContext(context.Background(), "tcp", "upstream.example:443")
	if err == nil {
		t.Fatal("expected the recording dialer error")
	}
	if resolver.calls != 1 {
		t.Fatalf("expected exactly one DNS lookup, got %d", resolver.calls)
	}
	if len(dialer.addresses) != 1 || dialer.addresses[0] != "93.184.216.34:443" {
		t.Fatalf("expected the validated IP literal to be dialed, got %v", dialer.addresses)
	}
	if strings.Contains(dialer.addresses[0], "upstream.example") {
		t.Fatalf("underlying dialer must not receive a hostname: %q", dialer.addresses[0])
	}
}

func TestSafeDialerUsesStaggeredFamilyFallback(t *testing.T) {
	resolver := &staticResolver{addresses: []net.IPAddr{
		{IP: net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")},
		{IP: net.ParseIP("93.184.216.34")},
	}}
	dialer := &happyEyeballsDialer{}
	safeDialer := NewSafeDialer(resolver, dialer, time.Second)
	safeDialer.fallbackDelay = 10 * time.Millisecond

	started := time.Now()
	conn, err := safeDialer.DialContext(context.Background(), "tcp", "upstream.example:443")
	if err != nil {
		t.Fatalf("expected IPv4 fallback to connect: %v", err)
	}
	_ = conn.Close()
	if elapsed := time.Since(started); elapsed >= 500*time.Millisecond {
		t.Fatalf("fallback waited too long: %s", elapsed)
	}
	dialer.mu.Lock()
	addresses := append([]string(nil), dialer.addresses...)
	dialer.mu.Unlock()
	if len(addresses) != 2 || !strings.Contains(addresses[0], "2606:2800") || addresses[1] != "93.184.216.34:443" {
		t.Fatalf("unexpected family fallback order: %v", addresses)
	}
}

func TestSafeDialerPreservesOriginalHostForTLSSNI(t *testing.T) {
	resolver := &staticResolver{addresses: []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}}
	baseDialer := &tlsCaptureDialer{sni: make(chan string, 1)}
	safeDialer := NewSafeDialer(resolver, baseDialer, time.Second)
	transport := &http.Transport{
		DialContext:         safeDialer.DialContext,
		TLSHandshakeTimeout: time.Second,
	}
	t.Cleanup(transport.CloseIdleConnections)

	req, err := http.NewRequest(http.MethodGet, "https://upstream.example/v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transport.RoundTrip(req)
	if err == nil {
		t.Fatal("expected the test TLS server to stop the handshake")
	}
	select {
	case serverName := <-baseDialer.sni:
		if serverName != "upstream.example" {
			t.Fatalf("expected original URL hostname as SNI, got %q", serverName)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for TLS ClientHello")
	}
}

func TestSafeDialerRejectsEntireMixedDNSAnswerBeforeDial(t *testing.T) {
	resolver := &staticResolver{addresses: []net.IPAddr{
		{IP: net.ParseIP("93.184.216.34")},
		{IP: net.ParseIP("169.254.169.254")},
	}}
	dialer := &recordingDialer{err: errors.New("must not dial")}
	safeDialer := NewSafeDialer(resolver, dialer, time.Second)

	_, err := safeDialer.DialContext(context.Background(), "tcp", "rebind.example:443")
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("expected the metadata IP to be rejected, got %v", err)
	}
	if len(dialer.addresses) != 0 {
		t.Fatalf("expected no dial after an unsafe DNS answer, got %v", dialer.addresses)
	}
}

func TestSafeDialerRejectsSpecialPurposeAddresses(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1:80",
		"10.0.0.1:80",
		"100.100.100.200:80",
		"168.63.129.16:80",
		"169.254.169.254:80",
		"192.0.2.1:80",
		"[::1]:80",
		"[64:ff9b::a9fe:a9fe]:80",
		"[2001:db8::1]:80",
	} {
		t.Run(address, func(t *testing.T) {
			dialer := &recordingDialer{err: errors.New("must not dial")}
			safeDialer := NewSafeDialer(nil, dialer, time.Second)
			_, err := safeDialer.DialContext(context.Background(), "tcp", address)
			if err == nil || !strings.Contains(err.Error(), "not allowed") {
				t.Fatalf("expected %s to be rejected, got %v", address, err)
			}
			if len(dialer.addresses) != 0 {
				t.Fatalf("expected no dial for %s", address)
			}
		})
	}
}
