package service

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

// fakeConversationRepo 是测试用的内存仓储实现。
type fakeConversationRepo struct {
	mu        sync.Mutex
	sessions  map[uuid.UUID]*ConversationSession
	branches  map[uuid.UUID]*ConversationBranch
	events    []*ConversationEvent
	refs      map[string]*ConversationResponseRef
	nextEvent int64
	seen      map[string]bool // 幂等键 branch|request|seq
}

func newFakeConversationRepo() *fakeConversationRepo {
	return &fakeConversationRepo{
		sessions: map[uuid.UUID]*ConversationSession{},
		branches: map[uuid.UUID]*ConversationBranch{},
		refs:     map[string]*ConversationResponseRef{},
		seen:     map[string]bool{},
	}
}

func intToStr(v int64) string { return strconv.FormatInt(v, 10) }

func (r *fakeConversationRepo) GetSessionByIdentity(_ context.Context, userID, groupID int64, contextDomain, archiveKey string) (*ConversationSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.sessions {
		if s.UserID == userID && s.GroupID == groupID && s.ContextDomain == contextDomain && s.ArchiveKey == archiveKey {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeConversationRepo) CreateSession(_ context.Context, s *ConversationSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	cp := *s
	r.sessions[s.ID] = &cp
	return nil
}

func (r *fakeConversationRepo) SetActiveBranch(_ context.Context, sessionID, branchID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.sessions[sessionID]; ok {
		b := branchID
		s.ActiveBranchID = &b
	}
	return nil
}

func (r *fakeConversationRepo) ApplySessionAggregate(_ context.Context, sessionID uuid.UUID, delta ConversationSessionAggregate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.sessions[sessionID]; ok {
		s.RequestCount += delta.RequestDelta
		s.TotalInputTokens += delta.InputTokensDelta
		s.TotalOutputTokens += delta.OutputTokensDelta
		if !delta.LastActiveAt.IsZero() {
			s.LastActiveAt = delta.LastActiveAt
		}
	}
	return nil
}

func (r *fakeConversationRepo) GetSessionByID(_ context.Context, id uuid.UUID) (*ConversationSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.sessions[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, ErrConversationNotFound
}

func (r *fakeConversationRepo) ListSessions(_ context.Context, params pagination.PaginationParams, filters ConversationSessionListFilters) ([]ConversationSession, *pagination.PaginationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var all []ConversationSession
	for _, s := range r.sessions {
		if filters.UserID != nil && s.UserID != *filters.UserID {
			continue
		}
		if filters.Status != "" && s.Status != filters.Status {
			continue
		}
		all = append(all, *s)
	}
	total := len(all)
	start := params.Offset()
	if start > total {
		start = total
	}
	end := start + params.Limit()
	if end > total {
		end = total
	}
	return all[start:end], &pagination.PaginationResult{Total: int64(total), Page: params.Page, PageSize: params.PageSize}, nil
}

func (r *fakeConversationRepo) DeleteSession(context.Context, uuid.UUID) error { return nil }

func (r *fakeConversationRepo) CreateBranch(_ context.Context, b *ConversationBranch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	cp := *b
	r.branches[b.ID] = &cp
	return nil
}

func (r *fakeConversationRepo) GetBranchByID(_ context.Context, id uuid.UUID) (*ConversationBranch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.branches[id]; ok {
		cp := *b
		return &cp, nil
	}
	return nil, ErrConversationNotFound
}

func (r *fakeConversationRepo) ListBranches(context.Context, uuid.UUID) ([]ConversationBranch, error) {
	return nil, nil
}

func (r *fakeConversationRepo) UpdateBranchCursor(_ context.Context, branchID uuid.UUID, cursor ConversationBranchCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.branches[branchID]; ok {
		b.TailSequence = cursor.TailSequence
		b.TailEventHash = cursor.TailEventHash
		b.HeadEventID = cursor.HeadEventID
		b.EventCount = cursor.EventCount
	}
	return nil
}

func (r *fakeConversationRepo) AppendEvents(_ context.Context, events []*ConversationEvent) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inserted := 0
	for _, ev := range events {
		key := ev.BranchID.String() + "|" + ev.RequestID + "|" + intToStr(int64(ev.Sequence))
		if r.seen[key] {
			continue
		}
		r.seen[key] = true
		r.nextEvent++
		ev.ID = r.nextEvent
		cp := *ev
		r.events = append(r.events, &cp)
		inserted++
	}
	return inserted, nil
}

func (r *fakeConversationRepo) ListEvents(_ context.Context, sessionID uuid.UUID, branchID *uuid.UUID) ([]ConversationEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []ConversationEvent
	for _, ev := range r.events {
		if ev.SessionID != sessionID {
			continue
		}
		if branchID != nil && ev.BranchID != *branchID {
			continue
		}
		out = append(out, *ev)
	}
	return out, nil
}

func (r *fakeConversationRepo) HasRequestEvents(_ context.Context, branchID uuid.UUID, requestID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ev := range r.events {
		if ev.BranchID == branchID && ev.RequestID == requestID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeConversationRepo) GetMaxSequence(_ context.Context, branchID uuid.UUID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	max := -1
	for _, ev := range r.events {
		if ev.BranchID == branchID && ev.Sequence > max {
			max = ev.Sequence
		}
	}
	return max, nil
}

func (r *fakeConversationRepo) UpsertResponseRef(_ context.Context, ref *ConversationResponseRef) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := ref.ContextDomain + "|" + ref.ResponseIDHash + "|" + intToStr(ref.UserID)
	cp := *ref
	r.refs[key] = &cp
	return nil
}

func (r *fakeConversationRepo) LookupResponseRef(_ context.Context, userID int64, contextDomain, responseIDHash string) (*ConversationResponseRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := contextDomain + "|" + responseIDHash + "|" + intToStr(userID)
	if ref, ok := r.refs[key]; ok {
		cp := *ref
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeConversationRepo) CleanupExpired(context.Context, time.Time, int) (int64, error) {
	return 0, nil
}

// --- 测试 ---

func newTestArchiver(repo ConversationRepository, mode string) *ConversationArchiver {
	return NewConversationArchiver(repo, nil, NewSecretRedactor(), ConversationArchiverOptions{
		Mode:           mode,
		QueueSize:      64,
		Workers:        1,
		RedactSecrets:  true,
		EncryptContent: false,
	})
}

func mkRecord(userID int64, archiveKey, contextDomain, requestID, userText, assistantText, responseID, parentRef string) CaptureRecord {
	return CaptureRecord{
		Version: CaptureRecordVersion,
		Identity: ConversationIdentity{
			ArchiveKey:    archiveKey,
			ContextDomain: contextDomain,
			Confidence:    IdentityConfidenceHigh,
			Source:        IdentitySourceSessionID,
			ParentRef:     parentRef,
		},
		Request: NormalizedRequest{
			UserID:    userID,
			Protocol:  ConversationProtocolOpenAI,
			RequestID: requestID,
			Events:    []NormalizedEvent{{Role: ConversationRoleUser, Kind: ConversationKindMessage, Content: userText}},
		},
		Response: NormalizedResponse{
			Model:      "gpt-5",
			ResponseID: responseID,
			Events:     []NormalizedEvent{{Role: ConversationRoleAssistant, Kind: ConversationKindMessage, Content: assistantText}},
		},
		Timing: CaptureTiming{FinishedAt: time.Now()},
	}
}

func TestArchiver_StoresConversationAndAppends(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)

	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "hello", "hi there", "resp-1", ""))
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-2", "second turn", "second answer", "resp-2", ""))
	a.Stop()

	if len(repo.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(repo.sessions))
	}
	var sess *ConversationSession
	for _, s := range repo.sessions {
		sess = s
	}
	if sess.RequestCount != 2 {
		t.Fatalf("request count = %d, want 2", sess.RequestCount)
	}
	// 2 turns * (user+assistant) = 4 events.
	evs, _ := repo.ListEvents(context.Background(), sess.ID, nil)
	if len(evs) != 4 {
		t.Fatalf("expected 4 events, got %d", len(evs))
	}
	// 序号连续 0..3
	for i, ev := range evs {
		if ev.Sequence != i {
			t.Fatalf("event %d has sequence %d", i, ev.Sequence)
		}
	}
	m := a.Metrics()
	if m.Succeeded != 2 || m.Dropped != 0 {
		t.Fatalf("metrics: %+v", m)
	}
}

func TestArchiver_PreviousResponseIDContinuation(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)

	// 第一轮建立会话并记录 resp-1 映射。
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "q1", "a1", "resp-1", ""))
	// 第二轮无 archive_key，仅靠 previous_response_id=resp-1 续传。
	rec := mkRecord(1, "", ContextDomainOpenAIAPI, "req-2", "q2", "a2", "resp-2", "resp-1")
	rec.Identity.Source = IdentitySourcePreviousResponseID
	a.Submit(context.Background(), rec)
	a.Stop()

	if len(repo.sessions) != 1 {
		t.Fatalf("previous_response_id should continue same session, got %d sessions", len(repo.sessions))
	}
}

func TestArchiver_IdempotentRetry(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)
	rec := mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "hello", "hi", "resp-1", "")
	a.Submit(context.Background(), rec)
	a.Submit(context.Background(), rec) // 同 request_id 重试
	a.Stop()

	var sess *ConversationSession
	for _, s := range repo.sessions {
		sess = s
	}
	evs, _ := repo.ListEvents(context.Background(), sess.ID, nil)
	if len(evs) != 2 {
		t.Fatalf("idempotent retry should keep 2 events, got %d", len(evs))
	}
}

func TestArchiver_RedactsSecretsInStoredContent(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "my key is sk-abcdefgh12345678", "ok", "resp-1", ""))
	a.Stop()

	var sess *ConversationSession
	for _, s := range repo.sessions {
		sess = s
	}
	evs, _ := repo.ListEvents(context.Background(), sess.ID, nil)
	for _, ev := range evs {
		if strings.Contains(string(ev.ContentCiphertext), "sk-abcdefgh12345678") {
			t.Fatalf("secret leaked into stored content: %q", ev.ContentCiphertext)
		}
		if strings.Contains(ev.ContentPreview, "sk-abcdefgh12345678") {
			t.Fatalf("secret leaked into preview")
		}
	}
}

func TestArchiver_MetadataModeStoresNoEvents(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeMetadata)
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "hello", "hi", "resp-1", ""))
	a.Stop()

	if len(repo.events) != 0 {
		t.Fatalf("metadata mode must not store events, got %d", len(repo.events))
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("metadata mode should still create the session")
	}
}
