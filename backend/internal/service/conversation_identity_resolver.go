package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// IdentityConfidence 表示会话身份的置信度。
//
// 调度（粘性）可接受较弱信号以提高命中率；归档合并必须谨慎：低置信度默认不与既有
// 会话合并，而是建立临时会话，避免把两个相似但不同的对话错误归并。
type IdentityConfidence int

const (
	IdentityConfidenceNone IdentityConfidence = iota // 无任何信号
	IdentityConfidenceLow                            // 仅内容前缀指纹
	IdentityConfidenceHigh                           // 强显式信号
)

// IdentitySource 标识身份来源（可观测）。
type IdentitySource string

const (
	IdentitySourceExplicitInternal   IdentitySource = "explicit_internal"
	IdentitySourceSessionID          IdentitySource = "session_id"
	IdentitySourceConversationID     IdentitySource = "conversation_id"
	IdentitySourcePromptCacheKey     IdentitySource = "prompt_cache_key"
	IdentitySourcePreviousResponseID IdentitySource = "previous_response_id"
	IdentitySourceThreadID           IdentitySource = "thread_id"
	IdentitySourceContentPrefix      IdentitySource = "content_prefix"
	IdentitySourceNone               IdentitySource = "none"
)

// 上游隔离域。即便客户端给出相同 session_id，不同 context_domain 也不归并到同一会话。
const (
	ContextDomainOpenAIAPI         = "openai_api"
	ContextDomainOpenAIOAuthCodex  = "openai_oauth_codex"
	ContextDomainAnthropicNative   = "anthropic_native"
	ContextDomainAntigravityClaude = "antigravity_claude"
	ContextDomainGeminiNative      = "gemini_native"
	ContextDomainAntigravityGemini = "antigravity_gemini"
	ContextDomainUnknown           = "unknown"
)

// ConversationIdentitySignals 是从请求中提取的身份候选值。
//
// Header 同时存在下划线/连字符写法时，调用方应在提取阶段归一后再填入。
// ContentSeed 为兜底种子（通常是「首条用户消息 + system + tools」的规范化串）。
type ConversationIdentitySignals struct {
	ExplicitConversationID string // 内部显式头 X-Sub2API-Conversation-ID
	SessionID              string // session_id / session-id
	ConversationID         string // conversation_id / conversation-id
	PromptCacheKey         string
	PreviousResponseID     string // 父响应引用，需经 response_refs 映射解析，非稳定会话 ID
	ThreadID               string
	MetadataNativeID       string // 协议原生 ID（如 metadata 中的会话标识）
	ContentSeed            string // 内容前缀兜底种子
}

// ConversationIdentityInput 是身份解析输入。
type ConversationIdentityInput struct {
	Secret        string // HMAC 密钥（archive_key 派生用）
	UserID        int64
	ContextDomain string
	Protocol      string
	Signals       ConversationIdentitySignals
}

// ConversationIdentity 是结构化的会话身份结果，供调度、归档、隔离、分支判断共享。
type ConversationIdentity struct {
	// ArchiveKey：归档会话唯一键（配合 user/group/context_domain）。
	// 强信号为 HMAC 派生的稳定键；低置信度/无信号为一次性临时键，不与既有会话合并。
	ArchiveKey string
	// RoutingKey：供粘性调度使用的弱键，尽量非空以提高命中。
	RoutingKey string
	// ContextDomain：上游隔离域。
	ContextDomain string
	Confidence    IdentityConfidence
	Source        IdentitySource
	// ParentRef：客户端带回的 previous_response_id（原值），由调用方经 response_refs 解析父分支。
	// 解析失败应安全降级为新建临时分支，绝不据此向请求注入任何 id。
	ParentRef string
}

// IsMergeable 表示该身份是否可安全地与既有归档会话合并（按 archive_key upsert）。
// 低置信度内容指纹默认不合并。
func (id ConversationIdentity) IsMergeable() bool {
	return id.Confidence == IdentityConfidenceHigh
}

// ResolveConversationIdentity 按优先级解析结构化会话身份。
//
// 优先级：内部显式 ID → session_id → conversation_id → prompt_cache_key →
// previous_response_id（父引用，交调用方映射）→ thread_id/原生 ID → 内容前缀（低置信度）→ 无。
func ResolveConversationIdentity(in ConversationIdentityInput) ConversationIdentity {
	s := in.Signals
	contextDomain := in.ContextDomain
	if contextDomain == "" {
		contextDomain = ContextDomainUnknown
	}

	out := ConversationIdentity{
		ContextDomain: contextDomain,
		RoutingKey:    firstNonEmptyRoutingKey(s),
		// ParentRef 始终透传 previous_response_id，无论哪个信号胜出，供分支父引用解析。
		ParentRef: strings.TrimSpace(s.PreviousResponseID),
	}

	strong := func(source IdentitySource, value string) ConversationIdentity {
		out.Source = source
		out.Confidence = IdentityConfidenceHigh
		out.ArchiveKey = deriveArchiveKey(in.Secret, in.UserID, contextDomain, source, value)
		return out
	}

	switch {
	case strings.TrimSpace(s.ExplicitConversationID) != "":
		return strong(IdentitySourceExplicitInternal, strings.TrimSpace(s.ExplicitConversationID))
	case strings.TrimSpace(s.SessionID) != "":
		return strong(IdentitySourceSessionID, strings.TrimSpace(s.SessionID))
	case strings.TrimSpace(s.ConversationID) != "":
		return strong(IdentitySourceConversationID, strings.TrimSpace(s.ConversationID))
	case strings.TrimSpace(s.PromptCacheKey) != "":
		return strong(IdentitySourcePromptCacheKey, strings.TrimSpace(s.PromptCacheKey))
	case strings.TrimSpace(s.PreviousResponseID) != "":
		// 父引用本身不是稳定会话键：留空 ArchiveKey，由调用方经 response_refs 映射定位会话/分支；
		// 命中前视为高置信度（有明确父引用），未命中则安全降级建临时会话。
		out.Source = IdentitySourcePreviousResponseID
		out.Confidence = IdentityConfidenceHigh
		out.ArchiveKey = ""
		return out
	case strings.TrimSpace(s.ThreadID) != "":
		return strong(IdentitySourceThreadID, strings.TrimSpace(s.ThreadID))
	case strings.TrimSpace(s.MetadataNativeID) != "":
		return strong(IdentitySourceThreadID, strings.TrimSpace(s.MetadataNativeID))
	case strings.TrimSpace(s.ContentSeed) != "":
		// 内容前缀：低置信度，默认不合并；archive_key 仍给出（临时会话用）。
		out.Source = IdentitySourceContentPrefix
		out.Confidence = IdentityConfidenceLow
		out.ArchiveKey = deriveArchiveKey(in.Secret, in.UserID, contextDomain, IdentitySourceContentPrefix, hashString(s.ContentSeed))
		return out
	default:
		out.Source = IdentitySourceNone
		out.Confidence = IdentityConfidenceNone
		out.ArchiveKey = NewTemporaryArchiveKey()
		return out
	}
}

// NewTemporaryArchiveKey 生成一次性临时归档键（永不与既有会话碰撞）。
func NewTemporaryArchiveKey() string {
	return "tmp:" + uuid.NewString()
}

// deriveArchiveKey 计算稳定归档键：HMAC-SHA256(secret, user|context_domain|source|value) 的十六进制。
// 不直接存客户端原始 ID，防止跨用户串会话与数据库泄露暴露原始标识。
func deriveArchiveKey(secret string, userID int64, contextDomain string, source IdentitySource, value string) string {
	if secret == "" {
		// 无密钥时退化为不可逆哈希（仍不暴露原值），保证功能可用。
		return hashString(strconv.FormatInt(userID, 10) + "|" + contextDomain + "|" + string(source) + "|" + value)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(userID, 10)))
	mac.Write([]byte("|"))
	mac.Write([]byte(contextDomain))
	mac.Write([]byte("|"))
	mac.Write([]byte(string(source)))
	mac.Write([]byte("|"))
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func hashString(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

// HashConversationRef 对上游 id（response_id / tool_call_id 等）做不可逆哈希用于追踪与匹配。
// 空串返回空串。
func HashConversationRef(v string) string {
	if strings.TrimSpace(v) == "" {
		return ""
	}
	return hashString(strings.TrimSpace(v))
}

// DeriveResponseDurability 返回某 context_domain 下上游 response id 是否可持久复用。
// openai_api 通常可持久；codex OAuth（store:false）下 reasoning item 不可持久；
// Anthropic 走完整历史续传，response id 不作持久延续依据。
func DeriveResponseDurability(contextDomain string) bool {
	switch contextDomain {
	case ContextDomainOpenAIAPI:
		return true
	default:
		return false
	}
}

func firstNonEmptyRoutingKey(s ConversationIdentitySignals) string {
	for _, v := range []string{
		s.ExplicitConversationID,
		s.SessionID,
		s.ConversationID,
		s.PromptCacheKey,
		s.PreviousResponseID,
		s.ThreadID,
		s.MetadataNativeID,
	} {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if strings.TrimSpace(s.ContentSeed) != "" {
		return hashString(s.ContentSeed)
	}
	return ""
}
