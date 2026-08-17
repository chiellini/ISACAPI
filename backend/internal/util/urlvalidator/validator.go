package urlvalidator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ValidationOptions struct {
	AllowedHosts     []string
	RequireAllowlist bool
	AllowPrivate     bool
}

// Resolver is the subset of net.Resolver used by SafeDialer. It is exported so
// callers can provide the resolver that matches their runtime environment and
// tests can deterministically exercise DNS rebinding scenarios.
type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// ContextDialer is the subset of net.Dialer used by SafeDialer.
type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// SafeDialer resolves a host exactly once per connection attempt, validates
// every returned address, and then dials an IP literal. Passing an IP literal to
// the underlying dialer prevents a second DNS lookup between validation and the
// socket connection. HTTP transports still retain the original URL hostname,
// so HTTPS certificate verification and TLS SNI are not changed.
type SafeDialer struct {
	resolver      Resolver
	dialer        ContextDialer
	timeout       time.Duration
	fallbackDelay time.Duration
}

const (
	defaultSafeDialFallbackDelay = 250 * time.Millisecond
	maxSafeDialAddresses         = 64
)

// NewSafeDialer builds a DNS-rebinding-safe socket dialer. Nil dependencies use
// the process defaults. timeout covers both DNS resolution and TCP connection.
func NewSafeDialer(resolver Resolver, dialer ContextDialer, timeout time.Duration) *SafeDialer {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	if dialer == nil {
		dialer = &net.Dialer{Timeout: timeout}
	}
	return &SafeDialer{resolver: resolver, dialer: dialer, timeout: timeout, fallbackDelay: defaultSafeDialFallbackDelay}
}

// Dial implements net.Dialer-style and golang.org/x/net/proxy.Dialer APIs.
func (d *SafeDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext validates and pins the resolved IP used for the socket.
func (d *SafeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d == nil || d.resolver == nil || d.dialer == nil {
		return nil, errors.New("safe dialer is not configured")
	}
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unsupported network for safe dialer: %s", network)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid dial address: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber <= 0 || portNumber > 65535 {
		return nil, fmt.Errorf("invalid dial port: %s", port)
	}

	if d.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.timeout)
		defer cancel()
	}

	ips, err := resolveAndValidateIPs(ctx, d.resolver, host)
	if err != nil {
		return nil, err
	}

	candidates := make([]netip.Addr, 0, len(ips))
	for _, ip := range interleaveIPFamilies(ips) {
		if !networkAllowsIP(network, ip) {
			continue
		}
		candidates = append(candidates, ip)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("dns resolution returned no addresses usable with %s", network)
	}

	type dialResult struct {
		ip   netip.Addr
		conn net.Conn
		err  error
	}
	dialCtx, cancelDials := context.WithCancel(ctx)
	defer cancelDials()
	results := make(chan dialResult)
	delay := d.fallbackDelay
	if delay <= 0 {
		delay = defaultSafeDialFallbackDelay
	}
	for index, ip := range candidates {
		go func(ip netip.Addr, startDelay time.Duration) {
			if startDelay > 0 {
				timer := time.NewTimer(startDelay)
				defer timer.Stop()
				select {
				case <-dialCtx.Done():
					return
				case <-timer.C:
				}
			}
			pinnedAddress := net.JoinHostPort(ip.String(), port)
			conn, dialErr := d.dialer.DialContext(dialCtx, network, pinnedAddress)
			result := dialResult{ip: ip, conn: conn, err: dialErr}
			select {
			case results <- result:
			case <-dialCtx.Done():
				if conn != nil {
					_ = conn.Close()
				}
			}
		}(ip, time.Duration(index)*delay)
	}

	dialErrors := make([]error, 0, len(candidates))
	for range candidates {
		select {
		case <-ctx.Done():
			cancelDials()
			return nil, errors.Join(append(dialErrors, ctx.Err())...)
		case result := <-results:
			if result.err == nil && result.conn != nil {
				cancelDials()
				return result.conn, nil
			}
			if result.conn != nil {
				_ = result.conn.Close()
			}
			dialErrors = append(dialErrors, fmt.Errorf("dial validated ip %s: %w", result.ip, result.err))
		}
	}
	return nil, errors.Join(dialErrors...)
}

func interleaveIPFamilies(ips []netip.Addr) []netip.Addr {
	if len(ips) < 2 {
		return append([]netip.Addr(nil), ips...)
	}
	v4 := make([]netip.Addr, 0, len(ips))
	v6 := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		if ip.Is4() || ip.Is4In6() {
			v4 = append(v4, ip)
		} else {
			v6 = append(v6, ip)
		}
	}
	preferV4 := ips[0].Is4() || ips[0].Is4In6()
	result := make([]netip.Addr, 0, len(ips))
	for len(v4) > 0 || len(v6) > 0 {
		if preferV4 && len(v4) > 0 {
			result = append(result, v4[0])
			v4 = v4[1:]
		} else if !preferV4 && len(v6) > 0 {
			result = append(result, v6[0])
			v6 = v6[1:]
		} else if len(v4) > 0 {
			result = append(result, v4[0])
			v4 = v4[1:]
		} else {
			result = append(result, v6[0])
			v6 = v6[1:]
		}
		preferV4 = !preferV4
	}
	return result
}

// ValidateHTTPURL validates an outbound HTTP/HTTPS URL.
//
// It provides a single validation entry point that supports:
// - scheme 校验（https 或可选允许 http）
// - 可选 allowlist（支持 *.example.com 通配）
// - allow_private_hosts 策略（阻断 localhost/私网字面量 IP）
//
// 注意：DNS Rebinding 防护（解析后 IP 校验）应在实际发起请求时执行，避免 TOCTOU。
func ValidateHTTPURL(raw string, allowInsecureHTTP bool, opts ValidationOptions) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid url")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && (!allowInsecureHTTP || scheme != "http") {
		return "", fmt.Errorf("invalid url scheme: %s", parsed.Scheme)
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return "", errors.New("invalid host")
	}
	if !opts.AllowPrivate && isBlockedHost(host) {
		return "", fmt.Errorf("host is not allowed: %s", host)
	}

	if port := parsed.Port(); port != "" {
		num, err := strconv.Atoi(port)
		if err != nil || num <= 0 || num > 65535 {
			return "", fmt.Errorf("invalid port: %s", port)
		}
	}

	allowlist := normalizeAllowlist(opts.AllowedHosts)
	if opts.RequireAllowlist && len(allowlist) == 0 {
		return "", errors.New("allowlist is not configured")
	}
	if len(allowlist) > 0 && !isAllowedHost(host, allowlist) {
		return "", fmt.Errorf("host is not allowed: %s", host)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func ValidateURLFormat(raw string, allowInsecureHTTP bool) (string, error) {
	// 最小格式校验：仅保证 URL 可解析且 scheme 合规，不做白名单/私网/SSRF 校验
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("url is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid url")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && (!allowInsecureHTTP || scheme != "http") {
		return "", fmt.Errorf("invalid url scheme: %s", parsed.Scheme)
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return "", errors.New("invalid host")
	}

	if port := parsed.Port(); port != "" {
		num, err := strconv.Atoi(port)
		if err != nil || num <= 0 || num > 65535 {
			return "", fmt.Errorf("invalid port: %s", port)
		}
	}

	return strings.TrimRight(trimmed, "/"), nil
}

func ValidateHTTPSURL(raw string, opts ValidationOptions) (string, error) {
	return ValidateHTTPURL(raw, false, opts)
}

// ValidateResolvedIP validates the current DNS answer. It is suitable for
// diagnostics, but a preflight call alone cannot prevent DNS rebinding because
// a later network dial may resolve the hostname again. Outbound HTTP transports
// that require SSRF protection must use SafeDialer.
func ValidateResolvedIP(host string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := resolveAndValidateIPs(ctx, net.DefaultResolver, host)
	return err
}

func resolveAndValidateIPs(ctx context.Context, resolver Resolver, host string) ([]netip.Addr, error) {
	host = strings.TrimSpace(strings.TrimSuffix(host, "."))
	if host == "" {
		return nil, errors.New("host is empty")
	}

	if literal, err := netip.ParseAddr(host); err == nil {
		literal = literal.Unmap()
		if err := validateIP(literal); err != nil {
			return nil, err
		}
		return []netip.Addr{literal}, nil
	}

	resolved, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("dns resolution failed: %w", err)
	}
	if len(resolved) == 0 {
		return nil, errors.New("dns resolution returned no addresses")
	}
	if len(resolved) > maxSafeDialAddresses {
		return nil, fmt.Errorf("dns resolution returned too many addresses: %d", len(resolved))
	}

	ips := make([]netip.Addr, 0, len(resolved))
	seen := make(map[netip.Addr]struct{}, len(resolved))
	for _, resolvedIP := range resolved {
		ip, ok := netip.AddrFromSlice(resolvedIP.IP)
		if !ok {
			return nil, fmt.Errorf("dns returned invalid ip: %q", resolvedIP.IP.String())
		}
		ip = ip.Unmap()
		if err := validateIP(ip); err != nil {
			// Mixed public/private answers are rejected as a set. Trying only the
			// public member can make behavior depend on DNS ordering and permits
			// an attacker-controlled answer set to reach the dial stage.
			return nil, err
		}
		if _, duplicate := seen[ip]; duplicate {
			continue
		}
		seen[ip] = struct{}{}
		ips = append(ips, ip)
	}
	return ips, nil
}

func validateIP(ip netip.Addr) error {
	if !ip.IsValid() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return fmt.Errorf("resolved ip %s is not allowed", ip)
	}
	// Block special-purpose/reserved ranges not covered by the convenience
	// predicates above, including IPv4 documentation, benchmarking, protocol
	// assignment, and the IPv6 documentation prefix. Global-unicast alone is
	// not sufficient because netip reports several of these as global.
	if !ip.IsGlobalUnicast() || isSpecialPurposeIP(ip) {
		return fmt.Errorf("resolved ip %s is not allowed", ip)
	}
	return nil
}

func isSpecialPurposeIP(ip netip.Addr) bool {
	for _, prefix := range blockedIPPrefixes {
		if prefix.Contains(ip) {
			return true
		}
	}
	return false
}

func networkAllowsIP(network string, ip netip.Addr) bool {
	switch network {
	case "tcp4":
		return ip.Is4()
	case "tcp6":
		return ip.Is6()
	default:
		return true
	}
}

var blockedIPPrefixes = mustPrefixes(
	"0.0.0.0/8",
	"100.64.0.0/10",
	"168.63.129.16/32",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"100::/64",
	"2001::/32",
	"2001:db8::/32",
	"2001:2::/48",
	"2001:10::/28",
	"2001:20::/28",
	"2002::/16",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
)

func mustPrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefixes = append(prefixes, netip.MustParsePrefix(value))
	}
	return prefixes
}

func normalizeAllowlist(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, v := range values {
		entry := strings.ToLower(strings.TrimSpace(v))
		if entry == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(entry); err == nil {
			entry = host
		}
		normalized = append(normalized, entry)
	}
	return normalized
}

func isAllowedHost(host string, allowlist []string) bool {
	for _, entry := range allowlist {
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, "*.") {
			suffix := strings.TrimPrefix(entry, "*.")
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return true
			}
			continue
		}
		// A full-label wildcard in the middle of a hostname is useful for
		// provider endpoints whose region is encoded as one DNS label, such as
		// bedrock-runtime.<region>.amazonaws.com. Unlike a leading wildcard,
		// it deliberately matches exactly one label and cannot widen to all of
		// amazonaws.com.
		if strings.Contains(entry, "*") {
			hostLabels := strings.Split(host, ".")
			entryLabels := strings.Split(entry, ".")
			if len(hostLabels) != len(entryLabels) {
				continue
			}
			matched := true
			for i := range entryLabels {
				if entryLabels[i] != "*" && entryLabels[i] != hostLabels[i] {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
			continue
		}
		if host == entry {
			return true
		}
	}
	return false
}

func isBlockedHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		return validateIP(ip.Unmap()) != nil
	}
	return false
}
