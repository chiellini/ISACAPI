package domain

import (
	"time"

	"github.com/google/uuid"
)

// 协议标识。
const (
	ConversationProtocolAnthropic = "anthropic"
	ConversationProtocolOpenAI    = "openai"
	ConversationProtocolGemini    = "gemini"
)

// 会话/分支状态。
const (
	ConversationStatusActive   = "active"
	ConversationStatusArchived = "archived"
	ConversationStatusExpired  = "expired"
)

// 事件角色。
const (
	ConversationRoleSystem    = "system"
	ConversationRoleUser      = "user"
	ConversationRoleAssistant = "assistant"
	ConversationRoleTool      = "tool"
)

// 事件类型。
const (
	ConversationKindMessage    = "message"
	ConversationKindToolCall   = "tool_call"
	ConversationKindToolResult = "tool_result"
	ConversationKindImageRef   = "image_reference"
	ConversationKindFileRef    = "file_reference"
	ConversationKindError      = "error"
)

// 采集级别。
const (
	ConversationModeMetadata          = "metadata"
	ConversationModeUserAssistantText = "user_assistant_text"
	ConversationModeFull              = "full"
)

// 分支产生原因。
const (
	BranchReasonInitial              = "initial"
	BranchReasonEditedHistory        = "edited_history"
	BranchReasonContinuationMismatch = "continuation_mismatch"
	BranchReasonProviderBoundary     = "provider_boundary"
	BranchReasonManualFork           = "manual_fork"
	BranchReasonUnknownParent        = "unknown_parent"
)

// ConversationSession 表示一个逻辑会话（Session → Branch → Event 模型的根）。
type ConversationSession struct {
	ID                uuid.UUID
	UserID            int64
	APIKeyID          *int64
	GroupID           int64 // 0 表示无分组
	ArchiveKey        string
	ContextDomain     string
	Protocol          string
	Title             string
	StartedAt         time.Time
	LastActiveAt      time.Time
	ActiveBranchID    *uuid.UUID
	RequestCount      int
	TotalInputTokens  int64
	TotalOutputTokens int64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// ConversationBranch 表示会话内的一条分支。tail_* 为游标快速路径所用尾事件信息。
type ConversationBranch struct {
	ID             uuid.UUID
	SessionID      uuid.UUID
	ParentBranchID *uuid.UUID
	ForkEventID    *int64
	HeadEventID    *int64
	EventCount     int
	TailSequence   int
	TailEventHash  string
	BranchReason   string
	Status         string
	CreatedAt      time.Time
	LastActiveAt   time.Time
}

// ConversationEvent 表示某分支上的一个事件。明文绝不落库（仅密文 + 脱敏预览）。
type ConversationEvent struct {
	ID                     int64
	SessionID              uuid.UUID
	BranchID               uuid.UUID
	ParentEventID          *int64
	Sequence               int
	RequestID              string
	Role                   string
	Kind                   string
	ContentCiphertext      []byte
	ContentNonce           []byte
	EncryptionKeyVersion   int
	ContentPreview         string
	EventHash              string
	Model                  string
	Provider               string
	UpstreamResponseIDHash string
	ToolCallIDHash         string
	Partial                bool
	CreatedAt              time.Time
}

// ConversationResponseRef 记录上游 response/item id 哈希 → 会话分支尾部 的映射。
type ConversationResponseRef struct {
	ID             int64
	UserID         int64
	ContextDomain  string
	ResponseIDHash string
	SessionID      uuid.UUID
	BranchID       uuid.UUID
	TailEventID    *int64
	Durable        bool
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
