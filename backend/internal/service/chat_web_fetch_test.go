package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatFetchValidateURL(t *testing.T) {
	for _, valid := range []string{
		"https://example.com/a?b=c",
		"http://example.com/",
		"https://example.com:443/path",
		// nip.io 形态合法：私网解析的拦截发生在拨号层而非 URL 形态层。
		"https://127.0.0.1.nip.io/",
	} {
		_, err := chatFetchValidateURL(valid)
		require.NoError(t, err, valid)
	}
	for _, invalid := range []string{
		"",
		"   ",
		"ftp://example.com/file",
		"file:///etc/passwd",
		"https://user:pass@example.com/",
		"https://example.com:8443/",
		strings.Repeat("https://example.com/", 200),
	} {
		_, err := chatFetchValidateURL(invalid)
		require.Error(t, err, invalid)
	}
}

func TestChatFetchExtractTextStripsScriptsAndTags(t *testing.T) {
	htmlBody := []byte(`<!doctype html><html><head><title>示例页</title><style>body{color:red}</style></head>
<body><script>alert(1)</script><h1>标题</h1><p>第一段 &amp; 细节</p><p>第二段</p><br/><div>尾部</div></body></html>`)
	text := chatFetchExtractText(htmlBody, true)
	require.Contains(t, text, "标题")
	require.Contains(t, text, "第一段 & 细节")
	require.Contains(t, text, "第二段")
	require.Contains(t, text, "尾部")
	require.NotContains(t, text, "alert(1)")
	require.NotContains(t, text, "color:red")
	require.NotContains(t, text, "<p>")
	require.Contains(t, text, "\n")
}

func TestChatFetchExtractTitle(t *testing.T) {
	require.Equal(t, "示例页", chatFetchExtractTitle([]byte(`<html><head><title> 示例页 </title></head></html>`)))
	require.Equal(t, "", chatFetchExtractTitle([]byte(`<html><body>no title</body></html>`)))
}

func TestTruncateUTF8BytesKeepsValidSequences(t *testing.T) {
	require.Equal(t, "abc", truncateUTF8Bytes("abcdef", 3))
	// “中文” 每字 3 字节：在 4 字节处截断必须退回到完整字符边界。
	require.Equal(t, "中", truncateUTF8Bytes("中文", 4))
	require.Equal(t, "中文", truncateUTF8Bytes("中文", 6))
	require.Equal(t, "中文x", truncateUTF8Bytes("中文xyz", 7))
}
