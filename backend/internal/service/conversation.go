package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

const (
	ConversationProtocolAnthropic = domain.ConversationProtocolAnthropic
	ConversationProtocolOpenAI    = domain.ConversationProtocolOpenAI
	ConversationProtocolGemini    = domain.ConversationProtocolGemini

	ConversationStatusActive   = domain.ConversationStatusActive
	ConversationStatusArchived = domain.ConversationStatusArchived
	ConversationStatusExpired  = domain.ConversationStatusExpired

	ConversationRoleSystem    = domain.ConversationRoleSystem
	ConversationRoleUser      = domain.ConversationRoleUser
	ConversationRoleAssistant = domain.ConversationRoleAssistant
	ConversationRoleTool      = domain.ConversationRoleTool

	ConversationKindMessage    = domain.ConversationKindMessage
	ConversationKindToolCall   = domain.ConversationKindToolCall
	ConversationKindToolResult = domain.ConversationKindToolResult
	ConversationKindImageRef   = domain.ConversationKindImageRef
	ConversationKindFileRef    = domain.ConversationKindFileRef
	ConversationKindError      = domain.ConversationKindError

	ConversationModeMetadata          = domain.ConversationModeMetadata
	ConversationModeUserAssistantText = domain.ConversationModeUserAssistantText
	ConversationModeFull              = domain.ConversationModeFull

	BranchReasonInitial              = domain.BranchReasonInitial
	BranchReasonEditedHistory        = domain.BranchReasonEditedHistory
	BranchReasonContinuationMismatch = domain.BranchReasonContinuationMismatch
	BranchReasonProviderBoundary     = domain.BranchReasonProviderBoundary
	BranchReasonManualFork           = domain.BranchReasonManualFork
	BranchReasonUnknownParent        = domain.BranchReasonUnknownParent
)

var (
	ErrConversationNotFound = infraerrors.NotFound("CONVERSATION_NOT_FOUND", "conversation not found")
	ErrConversationDisabled = infraerrors.BadRequest("CONVERSATION_DISABLED", "conversation archive is disabled")
)

type ConversationSession = domain.ConversationSession

type ConversationBranch = domain.ConversationBranch

type ConversationEvent = domain.ConversationEvent

type ConversationResponseRef = domain.ConversationResponseRef

// ConversationSessionAggregate 表示一次请求对会话聚合计数的增量。
type ConversationSessionAggregate struct {
	RequestDelta      int
	InputTokensDelta  int64
	OutputTokensDelta int64
	LastActiveAt      time.Time
}

// ConversationBranchCursor 表示分支游标的更新。
type ConversationBranchCursor struct {
	TailSequence  int
	TailEventHash string
	HeadEventID   *int64
	EventCount    int
	LastActiveAt  time.Time
}

// ConversationSessionListFilters 是会话列表过滤条件。Keyword 仅匹配脱敏预览。
type ConversationSessionListFilters struct {
	UserID        *int64
	GroupID       *int64
	ContextDomain string
	Protocol      string
	Status        string
	StartTime     *time.Time
	EndTime       *time.Time
	Keyword       string
}

// ConversationRepository 是对话存档（Session/Branch/Event/ResponseRef）的持久化接口。
//
// 实现须保证：AppendEvents 基于 (branch_id, request_id, sequence) 唯一约束幂等；
// CleanupExpired 批删除避免长事务。
type ConversationRepository interface {
	// --- 会话 ---
	// GetSessionByIdentity 按 (UserID, GroupID, ContextDomain, ArchiveKey) 查找会话，未找到返回 (nil, nil)。
	GetSessionByIdentity(ctx context.Context, userID, groupID int64, contextDomain, archiveKey string) (*ConversationSession, error)
	CreateSession(ctx context.Context, session *ConversationSession) error
	SetActiveBranch(ctx context.Context, sessionID, branchID uuid.UUID) error
	ApplySessionAggregate(ctx context.Context, sessionID uuid.UUID, delta ConversationSessionAggregate) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*ConversationSession, error)
	ListSessions(ctx context.Context, params pagination.PaginationParams, filters ConversationSessionListFilters) ([]ConversationSession, *pagination.PaginationResult, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// --- 分支 ---
	CreateBranch(ctx context.Context, branch *ConversationBranch) error
	GetBranchByID(ctx context.Context, id uuid.UUID) (*ConversationBranch, error)
	ListBranches(ctx context.Context, sessionID uuid.UUID) ([]ConversationBranch, error)
	UpdateBranchCursor(ctx context.Context, branchID uuid.UUID, cursor ConversationBranchCursor) error

	// --- 事件 ---
	// AppendEvents 幂等批量追加事件，返回实际新插入行数（冲突跳过）。
	AppendEvents(ctx context.Context, events []*ConversationEvent) (inserted int, err error)
	// ListEvents 返回会话事件；branchID 非 nil 时仅返回该分支，按 sequence 升序。
	ListEvents(ctx context.Context, sessionID uuid.UUID, branchID *uuid.UUID) ([]ConversationEvent, error)
	// GetMaxSequence 返回分支当前最大序号，无事件时返回 -1。
	GetMaxSequence(ctx context.Context, branchID uuid.UUID) (int, error)
	// HasRequestEvents 判断该分支是否已存在某 request_id 的事件（用于请求级幂等：重试跳过）。
	HasRequestEvents(ctx context.Context, branchID uuid.UUID, requestID string) (bool, error)

	// --- response_id 映射 ---
	UpsertResponseRef(ctx context.Context, ref *ConversationResponseRef) error
	// LookupResponseRef 解析 previous_response_id，未找到返回 (nil, nil)。
	LookupResponseRef(ctx context.Context, userID int64, contextDomain, responseIDHash string) (*ConversationResponseRef, error)

	// --- 清理 ---
	// CleanupExpired 物理删除 last_active_at 早于 cutoff 的会话（分支/事件经 FK 级联删除），返回删除会话数。
	CleanupExpired(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
}
