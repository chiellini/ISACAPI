package service

import "regexp"

// RedactedPlaceholder 是脱敏占位符。
const RedactedPlaceholder = "[REDACTED]"

// SecretRedactor 在内容落库前做两层 Secret 脱敏：
//
//	第一层：结构化字段（authorization / cookie / password / secret / token / api_key / private_key 等）的值
//	第二层：常见凭证格式（Bearer / sk- / AKIA / JWT / PEM 私钥 / 数据库连接串密码）
//
// 这是在线路径上的轻量实现，不跑完整 TruffleHog；熵检测等留给后续异步任务。
type SecretRedactor struct {
	rules []redactRule
}

type redactRule struct {
	re   *regexp.Regexp
	repl string
}

const sensitiveFieldNames = `authorization|cookie|password|passwd|secret|token|api[_-]?key|access[_-]?key|secret[_-]?key|private[_-]?key|client[_-]?secret|refresh[_-]?token`

// NewSecretRedactor 构造脱敏器（编译一次，复用）。
func NewSecretRedactor() *SecretRedactor {
	rules := []redactRule{
		// PEM 私钥块（多行）。放最前，避免被其它规则切碎。
		{regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`), "[REDACTED PRIVATE KEY]"},

		// 数据库/服务连接串中的密码：scheme://user:PASSWORD@host
		{regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.\-]*://[^:/@\s]+:)[^@/\s]+(@)`), "${1}" + RedactedPlaceholder + "${2}"},

		// 结构化字段（JSON 引号形式）："token": "VALUE"
		{regexp.MustCompile(`(?i)("(?:` + sensitiveFieldNames + `)"\s*:\s*")[^"]*(")`), "${1}" + RedactedPlaceholder + "${2}"},
		// 结构化字段（无引号 key:value / key=value）
		{regexp.MustCompile(`(?i)\b(` + sensitiveFieldNames + `)(\s*[:=]\s*)("?)[^\s,;&}"']+`), "${1}${2}${3}" + RedactedPlaceholder},

		// Bearer token
		{regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._\-]+`), "Bearer " + RedactedPlaceholder},
		// OpenAI/Anthropic 风格密钥 sk-... / sk-ant-... / sk-proj-...
		{regexp.MustCompile(`\bsk-[A-Za-z0-9_\-]{12,}`), RedactedPlaceholder},
		// AWS Access Key ID
		{regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), RedactedPlaceholder},
		// JWT（三段 base64url）
		{regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+`), RedactedPlaceholder},
	}
	return &SecretRedactor{rules: rules}
}

// Redact 返回脱敏后的文本。
func (r *SecretRedactor) Redact(s string) string {
	if s == "" {
		return s
	}
	for _, rule := range r.rules {
		s = rule.re.ReplaceAllString(s, rule.repl)
	}
	return s
}
