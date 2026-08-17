package securityaudit

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	maxGuardResponseBytes int64 = 256 * 1024

	// These process-level escape hatches are intentionally scoped to prompt
	// auditing. Production endpoints remain public HTTPS destinations unless an
	// operator explicitly opts in to a private node or plain HTTP.
	promptAuditAllowPrivateEndpointsEnv = "PROMPT_AUDIT_ALLOW_PRIVATE_ENDPOINTS"
	promptAuditAllowInsecureHTTPEnv     = "PROMPT_AUDIT_ALLOW_INSECURE_HTTP"
)

var (
	promptAuditSharedAddressPrefix = netip.MustParsePrefix("100.64.0.0/10")
	promptAuditMetadataAddresses   = map[netip.Addr]struct{}{
		netip.MustParseAddr("100.100.100.200"): {}, // Alibaba Cloud
		netip.MustParseAddr("169.254.169.254"): {}, // AWS, Azure, GCP and others
		netip.MustParseAddr("169.254.170.2"):   {}, // AWS container credentials
		netip.MustParseAddr("168.63.129.16"):   {}, // Azure WireServer / platform virtual IP
		netip.MustParseAddr("fd00:ec2::254"):   {}, // AWS IMDS IPv6
	}
)

type promptAuditIPResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type promptAuditDialContext func(ctx context.Context, network, address string) (net.Conn, error)

type promptAuditSecureDialer struct {
	resolver     promptAuditIPResolver
	dialContext  promptAuditDialContext
	allowPrivate bool
}

func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Opaque != "" {
		return "", infraerrors.BadRequest("prompt_audit_invalid_base_url", "审计节点地址无效")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && promptAuditEnvEnabled(promptAuditAllowInsecureHTTPEnv)) {
		return "", infraerrors.BadRequest("prompt_audit_invalid_base_url_scheme", "审计节点默认仅支持 HTTPS；本地测试必须显式启用不安全 HTTP")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", infraerrors.BadRequest("prompt_audit_unsafe_base_url", "审计节点地址不能包含凭据、查询参数或片段")
	}
	host := canonicalPromptAuditHost(parsed.Hostname())
	if host == "" {
		return "", infraerrors.BadRequest("prompt_audit_invalid_base_url", "审计节点地址无效")
	}
	if err := validatePromptAuditHost(host, promptAuditEnvEnabled(promptAuditAllowPrivateEndpointsEnv)); err != nil {
		return "", err
	}
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	if strings.EqualFold(path, "/v1") {
		path = ""
	}
	parsed.Path = path
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func ChatCompletionsURL(base string) (string, error) {
	normalized, err := NormalizeBaseURL(base)
	if err != nil {
		return "", err
	}
	return normalized + "/v1/chat/completions", nil
}

func ModelsURL(base string) (string, error) {
	normalized, err := NormalizeBaseURL(base)
	if err != nil {
		return "", err
	}
	return normalized + "/v1/models", nil
}

func NewSecureHTTPClient(endpoint ActiveEndpoint) (*http.Client, error) {
	if _, err := NormalizeBaseURL(endpoint.BaseURL); err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	secureDialer := &promptAuditSecureDialer{
		resolver:     net.DefaultResolver,
		dialContext:  dialer.DialContext,
		allowPrivate: promptAuditEnvEnabled(promptAuditAllowPrivateEndpointsEnv),
	}
	transport := &http.Transport{
		// Do not inherit HTTP(S)_PROXY. A proxy would move the actual destination
		// dial outside promptAuditSecureDialer and bypass DNS/IP validation.
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: time.Duration(endpoint.TimeoutMS) * time.Millisecond,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		DialContext:           secureDialer.DialContext,
	}
	timeout := time.Duration(endpoint.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = DefaultTimeoutMS * time.Millisecond
	}
	return &http.Client{
		Transport:     transport,
		Timeout:       timeout,
		CheckRedirect: checkPromptAuditRedirect,
	}, nil
}

func (d *promptAuditSecureDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid prompt audit destination: %w", err)
	}
	host = canonicalPromptAuditHost(host)
	if err := validatePromptAuditHost(host, d.allowPrivate); err != nil {
		return nil, err
	}

	addresses, err := d.resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	// Validate the complete DNS answer before opening any connection. This both
	// rejects mixed public/private answers and pins the subsequent dial to the
	// exact checked IP, closing the usual DNS rebinding/TOCTOU window.
	for _, addr := range addresses {
		if err := validatePromptAuditIP(addr, d.allowPrivate); err != nil {
			return nil, err
		}
	}

	var lastErr error
	for _, addr := range addresses {
		if network == "tcp4" && !addr.Is4() {
			continue
		}
		if network == "tcp6" && !addr.Is6() {
			continue
		}
		conn, dialErr := d.dialContext(ctx, network, net.JoinHostPort(addr.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no compatible IP address")
	}
	return nil, fmt.Errorf("prompt audit destination connection failed: %w", lastErr)
}

func (d *promptAuditSecureDialer) resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	if addr, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{addr}, nil
	}
	addresses, err := d.resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("prompt audit destination DNS lookup failed: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("prompt audit destination DNS lookup returned no addresses")
	}
	return addresses, nil
}

func validatePromptAuditHost(host string, allowPrivate bool) error {
	if isPromptAuditMetadataHostname(host) {
		return infraerrors.BadRequest("prompt_audit_metadata_destination_forbidden", "审计节点不能使用云平台元数据地址")
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		return validatePromptAuditIP(addr, allowPrivate)
	}
	if !allowPrivate && (host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local")) {
		return infraerrors.BadRequest("prompt_audit_private_destination_forbidden", "审计节点不能使用本地或私有地址")
	}
	return nil
}

func validatePromptAuditIP(addr netip.Addr, allowPrivate bool) error {
	addr = addr.Unmap()
	if _, metadata := promptAuditMetadataAddresses[addr]; metadata {
		return infraerrors.BadRequest("prompt_audit_metadata_destination_forbidden", "审计节点不能使用云平台元数据地址")
	}
	if allowPrivate {
		return nil
	}
	if !addr.IsValid() || !addr.IsGlobalUnicast() || addr.IsUnspecified() || addr.IsLoopback() || addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || promptAuditSharedAddressPrefix.Contains(addr) {
		return infraerrors.BadRequest("prompt_audit_private_destination_forbidden", "审计节点不能使用本地、私有、链路本地或保留地址")
	}
	return nil
}

func checkPromptAuditRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	if len(via) >= 10 {
		return fmt.Errorf("prompt audit redirect limit exceeded")
	}
	if req.URL.User != nil || !samePromptAuditOrigin(via[0].URL, req.URL) {
		return fmt.Errorf("prompt audit cross-origin redirect rejected")
	}
	return nil
}

func samePromptAuditOrigin(left, right *url.URL) bool {
	if left == nil || right == nil || !strings.EqualFold(left.Scheme, right.Scheme) {
		return false
	}
	return canonicalPromptAuditHost(left.Hostname()) == canonicalPromptAuditHost(right.Hostname()) &&
		effectivePromptAuditPort(left) == effectivePromptAuditPort(right)
}

func effectivePromptAuditPort(value *url.URL) string {
	if port := value.Port(); port != "" {
		return port
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	if strings.EqualFold(value.Scheme, "http") {
		return "80"
	}
	return ""
}

func canonicalPromptAuditHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func isPromptAuditMetadataHostname(host string) bool {
	switch canonicalPromptAuditHost(host) {
	case "metadata", "metadata.google", "metadata.goog", "metadata.google.internal", "instance-data.ec2.internal", "metadata.azure.internal":
		return true
	default:
		return false
	}
}

func promptAuditEnvEnabled(name string) bool {
	enabled, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(name)))
	return err == nil && enabled
}
