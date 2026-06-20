package repository

import (
	"context"
	"database/sql"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbconvbranch "github.com/Wei-Shaw/sub2api/ent/conversationbranch"
	dbconvevent "github.com/Wei-Shaw/sub2api/ent/conversationevent"
	dbconvref "github.com/Wei-Shaw/sub2api/ent/conversationresponseref"
	dbconvsession "github.com/Wei-Shaw/sub2api/ent/conversationsession"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
)

// conversationRepository 是对话存档（Session/Branch/Event/ResponseRef）持久化实现。
type conversationRepository struct {
	client *dbent.Client
	sql    *sql.DB
}

// NewConversationRepository 创建对话存档仓储。
func NewConversationRepository(client *dbent.Client, sqlDB *sql.DB) service.ConversationRepository {
	return &conversationRepository{client: client, sql: sqlDB}
}

// ---- 会话 ----

func (r *conversationRepository) GetSessionByIdentity(ctx context.Context, userID, groupID int64, contextDomain, archiveKey string) (*service.ConversationSession, error) {
	if archiveKey == "" {
		return nil, nil
	}
	row, err := r.client.ConversationSession.Query().
		Where(
			dbconvsession.UserID(userID),
			dbconvsession.GroupID(groupID),
			dbconvsession.ContextDomain(contextDomain),
			dbconvsession.ArchiveKey(archiveKey),
		).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toServiceConversationSession(row), nil
}

func (r *conversationRepository) CreateSession(ctx context.Context, s *service.ConversationSession) error {
	client := clientFromContext(ctx, r.client)
	now := s.LastActiveAt
	if now.IsZero() {
		now = time.Now()
	}
	builder := client.ConversationSession.Create().
		SetUserID(s.UserID).
		SetGroupID(s.GroupID).
		SetArchiveKey(s.ArchiveKey).
		SetContextDomain(s.ContextDomain).
		SetProtocol(s.Protocol).
		SetTitle(s.Title).
		SetStartedAt(firstNonZeroTime(s.StartedAt, now)).
		SetLastActiveAt(now).
		SetRequestCount(s.RequestCount).
		SetTotalInputTokens(s.TotalInputTokens).
		SetTotalOutputTokens(s.TotalOutputTokens).
		SetStatus(orDefault(s.Status, service.ConversationStatusActive))
	if s.ID != uuid.Nil {
		builder.SetID(s.ID)
	}
	if s.APIKeyID != nil {
		builder.SetAPIKeyID(*s.APIKeyID)
	}
	if s.ActiveBranchID != nil {
		builder.SetActiveBranchID(*s.ActiveBranchID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*s = *toServiceConversationSession(created)
	return nil
}

func (r *conversationRepository) SetActiveBranch(ctx context.Context, sessionID, branchID uuid.UUID) error {
	client := clientFromContext(ctx, r.client)
	err := client.ConversationSession.UpdateOneID(sessionID).
		SetActiveBranchID(branchID).
		Exec(ctx)
	if dbent.IsNotFound(err) {
		return service.ErrConversationNotFound
	}
	return err
}

func (r *conversationRepository) ApplySessionAggregate(ctx context.Context, sessionID uuid.UUID, delta service.ConversationSessionAggregate) error {
	client := clientFromContext(ctx, r.client)
	update := client.ConversationSession.UpdateOneID(sessionID).
		AddRequestCount(delta.RequestDelta).
		AddTotalInputTokens(delta.InputTokensDelta).
		AddTotalOutputTokens(delta.OutputTokensDelta)
	if !delta.LastActiveAt.IsZero() {
		update.SetLastActiveAt(delta.LastActiveAt)
	}
	err := update.Exec(ctx)
	if dbent.IsNotFound(err) {
		return service.ErrConversationNotFound
	}
	return err
}

func (r *conversationRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*service.ConversationSession, error) {
	row, err := r.client.ConversationSession.Get(ctx, id)
	if dbent.IsNotFound(err) {
		return nil, service.ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	return toServiceConversationSession(row), nil
}

func (r *conversationRepository) ListSessions(ctx context.Context, params pagination.PaginationParams, filters service.ConversationSessionListFilters) ([]service.ConversationSession, *pagination.PaginationResult, error) {
	q := r.client.ConversationSession.Query()
	if filters.UserID != nil {
		q = q.Where(dbconvsession.UserID(*filters.UserID))
	}
	if filters.GroupID != nil {
		q = q.Where(dbconvsession.GroupID(*filters.GroupID))
	}
	if filters.ContextDomain != "" {
		q = q.Where(dbconvsession.ContextDomain(filters.ContextDomain))
	}
	if filters.Protocol != "" {
		q = q.Where(dbconvsession.Protocol(filters.Protocol))
	}
	if filters.Status != "" {
		q = q.Where(dbconvsession.Status(filters.Status))
	}
	if filters.StartTime != nil {
		q = q.Where(dbconvsession.LastActiveAtGTE(*filters.StartTime))
	}
	if filters.EndTime != nil {
		q = q.Where(dbconvsession.LastActiveAtLTE(*filters.EndTime))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	rows, err := q.
		Order(dbent.Desc(dbconvsession.FieldLastActiveAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	sessions := make([]service.ConversationSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, *toServiceConversationSession(row))
	}
	result := &pagination.PaginationResult{Total: int64(total), Page: params.Page, PageSize: params.PageSize}
	return sessions, result, nil
}

func (r *conversationRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	client := clientFromContext(ctx, r.client)
	err := client.ConversationSession.DeleteOneID(id).Exec(ctx)
	if dbent.IsNotFound(err) {
		return service.ErrConversationNotFound
	}
	return err
}

// ---- 分支 ----

func (r *conversationRepository) CreateBranch(ctx context.Context, b *service.ConversationBranch) error {
	client := clientFromContext(ctx, r.client)
	now := b.LastActiveAt
	if now.IsZero() {
		now = time.Now()
	}
	builder := client.ConversationBranch.Create().
		SetSessionID(b.SessionID).
		SetEventCount(b.EventCount).
		SetTailSequence(orDefaultInt(b.TailSequence, -1)).
		SetTailEventHash(b.TailEventHash).
		SetBranchReason(orDefault(b.BranchReason, service.BranchReasonInitial)).
		SetStatus(orDefault(b.Status, service.ConversationStatusActive)).
		SetLastActiveAt(now)
	if b.ID != uuid.Nil {
		builder.SetID(b.ID)
	}
	if b.ParentBranchID != nil {
		builder.SetParentBranchID(*b.ParentBranchID)
	}
	if b.ForkEventID != nil {
		builder.SetForkEventID(*b.ForkEventID)
	}
	if b.HeadEventID != nil {
		builder.SetHeadEventID(*b.HeadEventID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*b = *toServiceConversationBranch(created)
	return nil
}

func (r *conversationRepository) GetBranchByID(ctx context.Context, id uuid.UUID) (*service.ConversationBranch, error) {
	row, err := r.client.ConversationBranch.Get(ctx, id)
	if dbent.IsNotFound(err) {
		return nil, service.ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	return toServiceConversationBranch(row), nil
}

func (r *conversationRepository) ListBranches(ctx context.Context, sessionID uuid.UUID) ([]service.ConversationBranch, error) {
	rows, err := r.client.ConversationBranch.Query().
		Where(dbconvbranch.SessionID(sessionID)).
		Order(dbent.Asc(dbconvbranch.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	branches := make([]service.ConversationBranch, 0, len(rows))
	for _, row := range rows {
		branches = append(branches, *toServiceConversationBranch(row))
	}
	return branches, nil
}

func (r *conversationRepository) UpdateBranchCursor(ctx context.Context, branchID uuid.UUID, cursor service.ConversationBranchCursor) error {
	client := clientFromContext(ctx, r.client)
	update := client.ConversationBranch.UpdateOneID(branchID).
		SetTailSequence(cursor.TailSequence).
		SetTailEventHash(cursor.TailEventHash).
		SetEventCount(cursor.EventCount)
	if cursor.HeadEventID != nil {
		update.SetHeadEventID(*cursor.HeadEventID)
	}
	if !cursor.LastActiveAt.IsZero() {
		update.SetLastActiveAt(cursor.LastActiveAt)
	}
	err := update.Exec(ctx)
	if dbent.IsNotFound(err) {
		return service.ErrConversationNotFound
	}
	return err
}

// ---- 事件 ----

func (r *conversationRepository) AppendEvents(ctx context.Context, events []*service.ConversationEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}
	client := clientFromContext(ctx, r.client)
	inserted := 0
	for _, ev := range events {
		builder := client.ConversationEvent.Create().
			SetSessionID(ev.SessionID).
			SetBranchID(ev.BranchID).
			SetSequence(ev.Sequence).
			SetRequestID(ev.RequestID).
			SetRole(ev.Role).
			SetKind(ev.Kind).
			SetEncryptionKeyVersion(ev.EncryptionKeyVersion).
			SetContentPreview(ev.ContentPreview).
			SetEventHash(ev.EventHash).
			SetModel(ev.Model).
			SetProvider(ev.Provider).
			SetUpstreamResponseIDHash(ev.UpstreamResponseIDHash).
			SetToolCallIDHash(ev.ToolCallIDHash).
			SetPartial(ev.Partial)
		if ev.ParentEventID != nil {
			builder.SetParentEventID(*ev.ParentEventID)
		}
		if len(ev.ContentCiphertext) > 0 {
			builder.SetContentCiphertext(ev.ContentCiphertext)
		}
		if len(ev.ContentNonce) > 0 {
			builder.SetContentNonce(ev.ContentNonce)
		}
		if !ev.CreatedAt.IsZero() {
			builder.SetCreatedAt(ev.CreatedAt)
		}

		id, err := builder.
			OnConflictColumns(
				dbconvevent.FieldBranchID,
				dbconvevent.FieldRequestID,
				dbconvevent.FieldSequence,
			).
			DoNothing().
			ID(ctx)
		if err != nil {
			if isSQLNoRowsError(err) {
				continue // 冲突：幂等跳过
			}
			return inserted, err
		}
		ev.ID = id
		inserted++
	}
	return inserted, nil
}

func (r *conversationRepository) ListEvents(ctx context.Context, sessionID uuid.UUID, branchID *uuid.UUID) ([]service.ConversationEvent, error) {
	q := r.client.ConversationEvent.Query().Where(dbconvevent.SessionID(sessionID))
	if branchID != nil {
		q = q.Where(dbconvevent.BranchID(*branchID))
	}
	rows, err := q.Order(dbent.Asc(dbconvevent.FieldSequence)).All(ctx)
	if err != nil {
		return nil, err
	}
	events := make([]service.ConversationEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, *toServiceConversationEvent(row))
	}
	return events, nil
}

func (r *conversationRepository) HasRequestEvents(ctx context.Context, branchID uuid.UUID, requestID string) (bool, error) {
	if requestID == "" {
		return false, nil
	}
	return r.client.ConversationEvent.Query().
		Where(
			dbconvevent.BranchID(branchID),
			dbconvevent.RequestID(requestID),
		).
		Exist(ctx)
}

func (r *conversationRepository) GetMaxSequence(ctx context.Context, branchID uuid.UUID) (int, error) {
	row, err := r.client.ConversationEvent.Query().
		Where(dbconvevent.BranchID(branchID)).
		Order(dbent.Desc(dbconvevent.FieldSequence)).
		Select(dbconvevent.FieldSequence).
		First(ctx)
	if dbent.IsNotFound(err) {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return row.Sequence, nil
}

// ---- response_id 映射 ----

func (r *conversationRepository) UpsertResponseRef(ctx context.Context, ref *service.ConversationResponseRef) error {
	client := clientFromContext(ctx, r.client)
	builder := client.ConversationResponseRef.Create().
		SetUserID(ref.UserID).
		SetContextDomain(ref.ContextDomain).
		SetResponseIDHash(ref.ResponseIDHash).
		SetSessionID(ref.SessionID).
		SetBranchID(ref.BranchID).
		SetDurable(ref.Durable)
	if ref.TailEventID != nil {
		builder.SetTailEventID(*ref.TailEventID)
	}
	if ref.ExpiresAt != nil {
		builder.SetExpiresAt(*ref.ExpiresAt)
	}
	return builder.
		OnConflictColumns(
			dbconvref.FieldUserID,
			dbconvref.FieldContextDomain,
			dbconvref.FieldResponseIDHash,
		).
		Update(func(u *dbent.ConversationResponseRefUpsert) {
			u.SetSessionID(ref.SessionID)
			u.SetBranchID(ref.BranchID)
			u.SetDurable(ref.Durable)
			u.UpdateUpdatedAt()
			if ref.TailEventID != nil {
				u.SetTailEventID(*ref.TailEventID)
			}
			if ref.ExpiresAt != nil {
				u.SetExpiresAt(*ref.ExpiresAt)
			}
		}).
		Exec(ctx)
}

func (r *conversationRepository) LookupResponseRef(ctx context.Context, userID int64, contextDomain, responseIDHash string) (*service.ConversationResponseRef, error) {
	if responseIDHash == "" {
		return nil, nil
	}
	row, err := r.client.ConversationResponseRef.Query().
		Where(
			dbconvref.UserID(userID),
			dbconvref.ContextDomain(contextDomain),
			dbconvref.ResponseIDHash(responseIDHash),
		).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toServiceConversationResponseRef(row), nil
}

// ---- 清理 ----

func (r *conversationRepository) CleanupExpired(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	if r.sql == nil {
		return 0, nil
	}
	if batchSize <= 0 {
		batchSize = 5000
	}
	var deleted int64
	for {
		res, err := r.sql.ExecContext(ctx, `
			WITH victims AS (
				SELECT ctid
				FROM conversation_sessions
				WHERE last_active_at < $1
				LIMIT $2
			)
			DELETE FROM conversation_sessions
			WHERE ctid IN (SELECT ctid FROM victims)
		`, cutoff.UTC(), batchSize)
		if err != nil {
			return deleted, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return deleted, err
		}
		deleted += affected
		if affected < int64(batchSize) {
			break
		}
	}
	return deleted, nil
}

// ---- 映射辅助 ----

func toServiceConversationSession(e *dbent.ConversationSession) *service.ConversationSession {
	if e == nil {
		return nil
	}
	return &service.ConversationSession{
		ID:                e.ID,
		UserID:            e.UserID,
		APIKeyID:          e.APIKeyID,
		GroupID:           e.GroupID,
		ArchiveKey:        e.ArchiveKey,
		ContextDomain:     e.ContextDomain,
		Protocol:          e.Protocol,
		Title:             e.Title,
		StartedAt:         e.StartedAt,
		LastActiveAt:      e.LastActiveAt,
		ActiveBranchID:    e.ActiveBranchID,
		RequestCount:      e.RequestCount,
		TotalInputTokens:  e.TotalInputTokens,
		TotalOutputTokens: e.TotalOutputTokens,
		Status:            e.Status,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		DeletedAt:         e.DeletedAt,
	}
}

func toServiceConversationBranch(e *dbent.ConversationBranch) *service.ConversationBranch {
	if e == nil {
		return nil
	}
	return &service.ConversationBranch{
		ID:             e.ID,
		SessionID:      e.SessionID,
		ParentBranchID: e.ParentBranchID,
		ForkEventID:    e.ForkEventID,
		HeadEventID:    e.HeadEventID,
		EventCount:     e.EventCount,
		TailSequence:   e.TailSequence,
		TailEventHash:  e.TailEventHash,
		BranchReason:   e.BranchReason,
		Status:         e.Status,
		CreatedAt:      e.CreatedAt,
		LastActiveAt:   e.LastActiveAt,
	}
}

func toServiceConversationEvent(e *dbent.ConversationEvent) *service.ConversationEvent {
	if e == nil {
		return nil
	}
	out := &service.ConversationEvent{
		ID:                     e.ID,
		SessionID:              e.SessionID,
		BranchID:               e.BranchID,
		ParentEventID:          e.ParentEventID,
		Sequence:               e.Sequence,
		RequestID:              e.RequestID,
		Role:                   e.Role,
		Kind:                   e.Kind,
		EncryptionKeyVersion:   e.EncryptionKeyVersion,
		ContentPreview:         e.ContentPreview,
		EventHash:              e.EventHash,
		Model:                  e.Model,
		Provider:               e.Provider,
		UpstreamResponseIDHash: e.UpstreamResponseIDHash,
		ToolCallIDHash:         e.ToolCallIDHash,
		Partial:                e.Partial,
		CreatedAt:              e.CreatedAt,
	}
	if e.ContentCiphertext != nil {
		out.ContentCiphertext = *e.ContentCiphertext
	}
	if e.ContentNonce != nil {
		out.ContentNonce = *e.ContentNonce
	}
	return out
}

func toServiceConversationResponseRef(e *dbent.ConversationResponseRef) *service.ConversationResponseRef {
	if e == nil {
		return nil
	}
	return &service.ConversationResponseRef{
		ID:             e.ID,
		UserID:         e.UserID,
		ContextDomain:  e.ContextDomain,
		ResponseIDHash: e.ResponseIDHash,
		SessionID:      e.SessionID,
		BranchID:       e.BranchID,
		TailEventID:    e.TailEventID,
		Durable:        e.Durable,
		ExpiresAt:      e.ExpiresAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func firstNonZeroTime(a, b time.Time) time.Time {
	if !a.IsZero() {
		return a
	}
	return b
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func orDefaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
