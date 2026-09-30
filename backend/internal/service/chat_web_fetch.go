package service

import (
	"context"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ChatFetchPage 是回给前端的一个网页正文（供 web_fetch 工具回灌与来源引用展示）。
type ChatFetchPage struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Text  string `json:"text"`
}

var (
	// ErrChatFetchInvalidURL 表示 URL 形态非法（非 http/https、带用户信息、端口受限等）。
	ErrChatFetchInvalidURL = infraerrors.BadRequest("CHAT_FETCH_INVALID_URL", "web fetch url is invalid")
	// ErrChatFetchBlocked 表示目标命中 SSRF 防护（私网/环回/链路本地等）。
	ErrChatFetchBlocked = infraerrors.BadRequest("CHAT_FETCH_BLOCKED", "web fetch target is not allowed")
	// ErrChatFetchFailed 表示下载或解析失败（网络错误、非 2xx、非文本类型等）。
	ErrChatFetchFailed = infraerrors.ServiceUnavailable("CHAT_FETCH_FAILED", "web fetch failed")
	// ErrChatFetchTooLarge 表示目标页面超过下载上限。
	ErrChatFetchTooLarge = infraerrors.BadRequest("CHAT_FETCH_TOO_LARGE", "web fetch target exceeds the size limit")
)

const (
	chatFetchMaxURLRunes  = 2048
	chatFetchTimeout      = 10 * time.Second
	chatFetchMaxRedirects = 3
	// chatFetchMaxDownload 单次下载上限（超出直接拒绝，不做截断下载）。
	chatFetchMaxDownload = 2 << 20
	// chatFetchMaxTextBytes 回灌给模型的正文上限。
	chatFetchMaxTextBytes = 64 << 10
)

// chatFetchValidateURL 校验 URL 形态：仅 http/https、仅默认端口、拒绝 userinfo。
// 主机是否解析到公网 IP 在拨号时二次校验（见 chatFetchDialContext）。
func chatFetchValidateURL(raw string) (*url.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > chatFetchMaxURLRunes {
		return nil, ErrChatFetchInvalidURL
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, ErrChatFetchInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrChatFetchInvalidURL
	}
	if u.Hostname() == "" || u.User != nil {
		return nil, ErrChatFetchInvalidURL
	}
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		return nil, ErrChatFetchBlocked
	}
	return u, nil
}

// chatFetchDialContext 先解析再拨号：所有解析结果必须是公网单播地址，
// 阻断 SSRF（私网 / 环回 / 链路本地 / 未指定 / 组播），并直接拨向已校验的
// IP，关闭「校验后 DNS 再变」的重绑窗口。
func chatFetchDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, ErrChatFetchBlocked
	}
	for _, ip := range ips {
		parsed, err := netip.ParseAddr(ip.IP.String())
		if err != nil {
			return nil, ErrChatFetchBlocked
		}
		parsed = parsed.Unmap()
		if parsed.IsPrivate() || parsed.IsLoopback() || parsed.IsLinkLocalUnicast() ||
			parsed.IsLinkLocalMulticast() || parsed.IsMulticast() || parsed.IsUnspecified() ||
			!parsed.IsGlobalUnicast() {
			return nil, ErrChatFetchBlocked
		}
	}
	var d net.Dialer
	var lastErr error
	for _, ip := range ips {
		conn, dialErr := d.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no dialable address for %s", host)
	}
	return nil, lastErr
}

// chatFetchHTTPClient 出站抓取客户端：直连（不走代理），重定向前逐跳复检。
var chatFetchHTTPClient = &http.Client{
	Transport: &http.Transport{
		DialContext: chatFetchDialContext,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= chatFetchMaxRedirects {
			return fmt.Errorf("too many redirects")
		}
		if _, err := chatFetchValidateURL(req.URL.String()); err != nil {
			return err
		}
		return nil
	},
	Timeout: chatFetchTimeout,
}

// FetchPage 抓取一个公网页面并抽取可读正文。供聊天页 web_fetch 工具调用。
func (s *ChatHistoryService) FetchPage(ctx context.Context, rawURL string) (*ChatFetchPage, error) {
	u, err := chatFetchValidateURL(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, ErrChatFetchInvalidURL
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ISACAI-Chat/Bot)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain,application/json;q=0.9,*/*;q=0.5")

	resp, err := chatFetchHTTPClient.Do(req)
	if err != nil {
		return nil, ErrChatFetchFailed
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrChatFetchFailed
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	baseType := strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch {
	case baseType == "", strings.HasPrefix(baseType, "text/"),
		baseType == "application/json", baseType == "application/xhtml+xml":
	default:
		return nil, ErrChatFetchFailed
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, chatFetchMaxDownload+1))
	if err != nil {
		return nil, ErrChatFetchFailed
	}
	if len(body) > chatFetchMaxDownload {
		return nil, ErrChatFetchTooLarge
	}

	isHTML := strings.Contains(contentType, "html")
	title := ""
	if isHTML {
		title = chatFetchExtractTitle(body)
	}
	text := truncateUTF8Bytes(chatFetchExtractText(body, isHTML), chatFetchMaxTextBytes)
	if strings.TrimSpace(text) == "" {
		return nil, ErrChatFetchFailed
	}
	return &ChatFetchPage{Title: title, URL: resp.Request.URL.String(), Text: text}, nil
}

var (
	chatFetchScriptRe = regexp.MustCompile(`(?is)<(script|style|noscript|template|svg)\b[^>]*>.*?</(script|style|noscript|template|svg)>`)
	chatFetchBlockRe  = regexp.MustCompile(`(?i)</?(?:p|div|section|article|aside|header|footer|nav|br|li|ul|ol|table|tr|td|th|h[1-6]|blockquote|pre|hr)\b[^>]*>`)
	chatFetchTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	chatFetchSpaceRe  = regexp.MustCompile(`[ \t\f\v]+`)
	chatFetchNLinesRe = regexp.MustCompile(`\n{3,}`)
	chatFetchTitleRe  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

// chatFetchExtractTitle 抽取 <title>（仅 HTML）。
func chatFetchExtractTitle(body []byte) string {
	if m := chatFetchTitleRe.FindSubmatch(body); len(m) == 2 {
		return strings.TrimSpace(chatFetchSpaceRe.ReplaceAllString(html.UnescapeString(string(m[1])), " "))
	}
	return ""
}

// chatFetchExtractText 把 HTML（或纯文本）压成单行间距的可读正文：
// 去掉脚本/样式与标签，块级标签转换行，实体解码，空白折叠。
func chatFetchExtractText(body []byte, isHTML bool) string {
	content := string(body)
	if isHTML {
		content = chatFetchScriptRe.ReplaceAllString(content, " ")
		content = chatFetchBlockRe.ReplaceAllString(content, "\n")
		content = chatFetchTagRe.ReplaceAllString(content, " ")
		content = html.UnescapeString(content)
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = chatFetchSpaceRe.ReplaceAllString(content, " ")
	content = chatFetchNLinesRe.ReplaceAllString(content, "\n")
	return strings.TrimSpace(content)
}

// truncateUTF8Bytes 按字节截断并保证不切断 UTF-8 序列。
func truncateUTF8Bytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := s[:maxBytes]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return strings.TrimRight(cut, " \n\t")
}
