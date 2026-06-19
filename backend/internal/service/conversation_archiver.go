package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ConversationArchiverOptions 是归档器的运行参数（与 config 包解耦，由 provider 注入）。
type ConversationArchiverOptions struct {
	Mode             string
	QueueSize        int
	Workers          int
	MaxEventBytes    int
	MaxSessionEvents int
	PreviewRunes     int
	RedactSecrets    bool
	EncryptContent   bool
	CaptureSystem    bool
}

// ConversationArchiverMetrics 是归档器的可观测计数快照。
type ConversationArchiverMetrics struct {
	Submitted       uint64
	Dropped         uint64
	Succeeded       uint64
	Failed          uint64
	BranchesCreated uint64
	LowConfidence   uint64
	QueueDepth      int
}

// ConversationArchiver 是基于有界内存队列的对话采集落库实现（CaptureSink）。
//
// Submit 非阻塞：队列满即丢弃并计数，绝不阻塞转发主路径（fail-open）。后台 worker 串行
// 消费，将一次请求/响应编排为 会话 → 分支 → 事件 落库；内容经脱敏 + AES-GCM 加密。
//
// 第一版只追加到活跃分支（含 previous_response_id 续传定位）；编辑旧历史的分叉去重留待后续。
type ConversationArchiver struct {
	repo        ConversationRepository
	enc         *ContentEncryptor
	redactor    *SecretRedactor
	opts        ConversationArchiverOptions
	queue       chan CaptureRecord
	wg          sync.WaitGroup
	stopOnce    sync.Once
	baseCtx     context.Context
	cancel      context.CancelFunc
	procTimeout time.Duration

	submitted       atomic.Uint64
	dropped         atomic.Uint64
	succeeded       atomic.Uint64
	failed          atomic.Uint64
	branchesCreated atomic.Uint64
	lowConfidence   atomic.Uint64
}

// NewConversationArchiver 构造并启动归档器。redactor 为 nil 时不脱敏；enc 为 nil 时不加密。
func NewConversationArchiver(repo ConversationRepository, enc *ContentEncryptor, redactor *SecretRedactor, opts ConversationArchiverOptions) *ConversationArchiver {
	if opts.QueueSize <= 0 {
		opts.QueueSize = 1000
	}
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	if opts.PreviewRunes <= 0 {
		opts.PreviewRunes = 200
	}
	if opts.MaxEventBytes <= 0 {
		opts.MaxEventBytes = 65536
	}
	if opts.Mode == "" {
		opts.Mode = ConversationModeUserAssistantText
	}
	baseCtx, cancel := context.WithCancel(context.Background())
	a := &ConversationArchiver{
		repo:        repo,
		enc:         enc,
		redactor:    redactor,
		opts:        opts,
		queue:       make(chan CaptureRecord, opts.QueueSize),
		baseCtx:     baseCtx,
		cancel:      cancel,
		procTimeout: 10 * time.Second,
	}
	for i := 0; i < opts.Workers; i++ {
		a.wg.Add(1)
		go a.worker()
	}
	return a
}

// Submit 实现 CaptureSink：非阻塞入队，满则丢弃。
func (a *ConversationArchiver) Submit(_ context.Context, record CaptureRecord) {
	a.submitted.Add(1)
	if record.Identity.Confidence == IdentityConfidenceLow {
		a.lowConfidence.Add(1)
	}
	select {
	case a.queue <- record:
	default:
		a.dropped.Add(1)
		slog.Warn("conversation.capture.dropped", "queue_size", a.opts.QueueSize)
	}
}

// Metrics 返回计数快照。
func (a *ConversationArchiver) Metrics() ConversationArchiverMetrics {
	return ConversationArchiverMetrics{
		Submitted:       a.submitted.Load(),
		Dropped:         a.dropped.Load(),
		Succeeded:       a.succeeded.Load(),
		Failed:          a.failed.Load(),
		BranchesCreated: a.branchesCreated.Load(),
		LowConfidence:   a.lowConfidence.Load(),
		QueueDepth:      len(a.queue),
	}
}

// Stop 停止 worker 并等待退出。
func (a *ConversationArchiver) Stop() {
	a.stopOnce.Do(func() {
		close(a.queue)
		a.wg.Wait()
		a.cancel()
	})
}

func (a *ConversationArchiver) worker() {
	defer a.wg.Done()
	for record := range a.queue {
		a.safeProcess(record)
	}
}

func (a *ConversationArchiver) safeProcess(record CaptureRecord) {
	defer func() {
		if r := recover(); r != nil {
			a.failed.Add(1)
			slog.Error("conversation.capture.panic", "recover", r)
		}
	}()
	ctx, cancel := context.WithTimeout(a.baseCtx, a.procTimeout)
	defer cancel()
	if err := a.process(ctx, record); err != nil {
		a.failed.Add(1)
		slog.Warn("conversation.capture.failed", "error", err.Error())
		return
	}
	a.succeeded.Add(1)
}

func (a *ConversationArchiver) process(ctx context.Context, rec CaptureRecord) error {
	now := rec.Timing.FinishedAt
	if now.IsZero() {
		now = time.Now()
	}

	session, branchID, err := a.resolveSessionBranch(ctx, rec, now)
	if err != nil {
		return err
	}

	// 请求级幂等：该分支已存在此 request_id 的事件 → 视为重试，整体跳过（含聚合），避免重复计数。
	if rec.Request.RequestID != "" {
		if exists, err := a.repo.HasRequestEvents(ctx, branchID, rec.Request.RequestID); err == nil && exists {
			return nil
		}
	}

	// 聚合（所有模式都更新）。
	defer func() {
		_ = a.repo.ApplySessionAggregate(ctx, session.ID, ConversationSessionAggregate{
			RequestDelta:      1,
			InputTokensDelta:  rec.Response.InputTokens,
			OutputTokensDelta: rec.Response.OutputTokens,
			LastActiveAt:      now,
		})
	}()

	// metadata 模式不存内容。
	if a.opts.Mode == ConversationModeMetadata {
		return nil
	}

	maxSeq, err := a.repo.GetMaxSequence(ctx, branchID)
	if err != nil {
		return err
	}
	if a.opts.MaxSessionEvents > 0 && maxSeq+1 >= a.opts.MaxSessionEvents {
		return nil // 会话过大，停止追加
	}

	events := a.buildEvents(rec, session.ID, branchID)
	if len(events) == 0 {
		return nil
	}
	seq := maxSeq
	for _, ev := range events {
		seq++
		ev.Sequence = seq
		ev.CreatedAt = now
	}
	lastHash := events[len(events)-1].EventHash

	inserted, err := a.repo.AppendEvents(ctx, events)
	if err != nil {
		return err
	}
	if inserted == 0 {
		return nil // 全部为重试冲突，幂等跳过
	}

	var headID *int64
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].ID != 0 {
			id := events[i].ID
			headID = &id
			break
		}
	}
	if err := a.repo.UpdateBranchCursor(ctx, branchID, ConversationBranchCursor{
		TailSequence:  seq,
		TailEventHash: lastHash,
		HeadEventID:   headID,
		EventCount:    seq + 1,
		LastActiveAt:  now,
	}); err != nil {
		return err
	}

	// 记录 response_id → 分支尾部 映射，供下一轮 previous_response_id 续传定位。
	if rec.Response.ResponseID != "" {
		_ = a.repo.UpsertResponseRef(ctx, &ConversationResponseRef{
			UserID:         rec.Request.UserID,
			ContextDomain:  rec.Identity.ContextDomain,
			ResponseIDHash: HashConversationRef(rec.Response.ResponseID),
			SessionID:      session.ID,
			BranchID:       branchID,
			TailEventID:    headID,
			Durable:        DeriveResponseDurability(rec.Identity.ContextDomain),
		})
	}
	return nil
}

// resolveSessionBranch 解析（或创建）会话与活跃分支。
func (a *ConversationArchiver) resolveSessionBranch(ctx context.Context, rec CaptureRecord, now time.Time) (*ConversationSession, uuid.UUID, error) {
	id := rec.Identity

	// 1. previous_response_id 续传：定位既有会话/分支。
	if id.ParentRef != "" {
		ref, err := a.repo.LookupResponseRef(ctx, rec.Request.UserID, id.ContextDomain, HashConversationRef(id.ParentRef))
		if err == nil && ref != nil {
			if sess, err := a.repo.GetSessionByID(ctx, ref.SessionID); err == nil && sess != nil {
				return sess, ref.BranchID, nil
			}
		}
	}

	// 2. 强信号且可合并：按 archive_key 查既有会话。
	if id.ArchiveKey != "" && id.IsMergeable() {
		sess, err := a.repo.GetSessionByIdentity(ctx, rec.Request.UserID, rec.Request.GroupID, id.ContextDomain, id.ArchiveKey)
		if err != nil {
			return nil, uuid.Nil, err
		}
		if sess != nil {
			branchID, err := a.ensureActiveBranch(ctx, sess, now)
			return sess, branchID, err
		}
		return a.createSessionWithBranch(ctx, rec, id.ArchiveKey, now)
	}

	// 3. 低置信度 / 无信号：建立临时会话（不与既有合并）。
	key := id.ArchiveKey
	if key == "" {
		key = NewTemporaryArchiveKey()
	}
	return a.createSessionWithBranch(ctx, rec, key, now)
}

func (a *ConversationArchiver) ensureActiveBranch(ctx context.Context, sess *ConversationSession, now time.Time) (uuid.UUID, error) {
	if sess.ActiveBranchID != nil {
		return *sess.ActiveBranchID, nil
	}
	branch := &ConversationBranch{
		SessionID:    sess.ID,
		BranchReason: BranchReasonInitial,
		Status:       ConversationStatusActive,
		TailSequence: -1,
		LastActiveAt: now,
	}
	if err := a.repo.CreateBranch(ctx, branch); err != nil {
		return uuid.Nil, err
	}
	a.branchesCreated.Add(1)
	if err := a.repo.SetActiveBranch(ctx, sess.ID, branch.ID); err != nil {
		return uuid.Nil, err
	}
	sess.ActiveBranchID = &branch.ID
	return branch.ID, nil
}

func (a *ConversationArchiver) createSessionWithBranch(ctx context.Context, rec CaptureRecord, archiveKey string, now time.Time) (*ConversationSession, uuid.UUID, error) {
	sess := &ConversationSession{
		UserID:        rec.Request.UserID,
		APIKeyID:      rec.Request.APIKeyID,
		GroupID:       rec.Request.GroupID,
		ArchiveKey:    archiveKey,
		ContextDomain: rec.Identity.ContextDomain,
		Protocol:      rec.Request.Protocol,
		StartedAt:     now,
		LastActiveAt:  now,
		Status:        ConversationStatusActive,
	}
	if err := a.repo.CreateSession(ctx, sess); err != nil {
		return nil, uuid.Nil, err
	}
	branch := &ConversationBranch{
		SessionID:    sess.ID,
		BranchReason: BranchReasonInitial,
		Status:       ConversationStatusActive,
		TailSequence: -1,
		LastActiveAt: now,
	}
	if err := a.repo.CreateBranch(ctx, branch); err != nil {
		return nil, uuid.Nil, err
	}
	a.branchesCreated.Add(1)
	if err := a.repo.SetActiveBranch(ctx, sess.ID, branch.ID); err != nil {
		return nil, uuid.Nil, err
	}
	sess.ActiveBranchID = &branch.ID
	return sess, branch.ID, nil
}

func (a *ConversationArchiver) buildEvents(rec CaptureRecord, sessionID, branchID uuid.UUID) []*ConversationEvent {
	out := make([]*ConversationEvent, 0, len(rec.Request.Events)+len(rec.Response.Events))

	add := func(src NormalizedEvent, responseID string, partial bool) {
		kind := src.Kind
		if kind == "" {
			kind = ConversationKindMessage
		}
		if src.Role == ConversationRoleSystem && !a.opts.CaptureSystem {
			return
		}
		content := src.Content
		if content == "" && kind == ConversationKindMessage {
			return
		}
		if a.opts.RedactSecrets && a.redactor != nil {
			content = a.redactor.Redact(content)
		}
		content = truncateUTF8(content, a.opts.MaxEventBytes)

		ev := &ConversationEvent{
			SessionID:              sessionID,
			BranchID:               branchID,
			RequestID:              rec.Request.RequestID,
			Role:                   src.Role,
			Kind:                   kind,
			EventHash:              conversationEventHash(src.Role, kind, content, src.ToolCallID),
			Model:                  rec.Response.Model,
			Provider:               rec.Request.Protocol,
			Partial:                partial,
			UpstreamResponseIDHash: HashConversationRef(responseID),
			ToolCallIDHash:         HashConversationRef(src.ToolCallID),
			ContentPreview:         previewOf(content, a.opts.PreviewRunes),
		}
		a.applyContent(ev, content)
		out = append(out, ev)
	}

	for _, e := range rec.Request.Events {
		add(e, "", false)
	}
	for _, e := range rec.Response.Events {
		add(e, rec.Response.ResponseID, rec.Response.Partial)
	}
	return out
}

// applyContent 把明文写入事件：加密则存密文 + nonce + 版本；否则存明文字节并以 key_version=0 标记。
func (a *ConversationArchiver) applyContent(ev *ConversationEvent, content string) {
	if a.opts.EncryptContent && a.enc != nil {
		ct, nonce, err := a.enc.Encrypt([]byte(content))
		if err == nil {
			ev.ContentCiphertext = ct
			ev.ContentNonce = nonce
			ev.EncryptionKeyVersion = a.enc.KeyVersion()
			return
		}
		slog.Warn("conversation.capture.encrypt_failed", "error", err.Error())
	}
	// 未加密：明文字节，key_version=0 表示未加密。
	ev.ContentCiphertext = []byte(content)
	ev.EncryptionKeyVersion = 0
}

func conversationEventHash(role, kind, content, toolCallID string) string {
	h := sha256.New()
	h.Write([]byte(role))
	h.Write([]byte{0})
	h.Write([]byte(kind))
	h.Write([]byte{0})
	h.Write([]byte(content))
	h.Write([]byte{0})
	h.Write([]byte(toolCallID))
	return hex.EncodeToString(h.Sum(nil))
}

func previewOf(s string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = 200
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

func truncateUTF8(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	b := []byte(s)[:maxBytes]
	for len(b) > 0 && !utf8.Valid(b) {
		b = b[:len(b)-1]
	}
	return string(b)
}

var _ CaptureSink = (*ConversationArchiver)(nil)
