package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

// conversationExportMaxSessions 批量导出的安全上限，避免一次性占用过多内存。
const conversationExportMaxSessions = 5000

// ConversationService 提供对话存档的管理员读取/删除能力（含内容解密）。
//
// 仅供管理员侧使用；普通用户不暴露任何对话查看入口。
type ConversationService struct {
	repo ConversationRepository
	enc  *ContentEncryptor // 可空：未配置加密或密钥不可用时为 nil（此时只能展示脱敏预览）
	cfg  *config.Config
}

// ProvideConversationService 构造管理员对话服务。
func ProvideConversationService(cfg *config.Config, repo ConversationRepository) *ConversationService {
	var enc *ContentEncryptor
	if cfg != nil && cfg.ConversationArchive.EncryptContent {
		if e, err := NewContentEncryptor(cfg.ConversationArchive.EncryptionKey); err == nil {
			enc = e
		}
	}
	return &ConversationService{repo: repo, enc: enc, cfg: cfg}
}

// ConversationSessionView 是会话的管理员视图。
type ConversationSessionView struct {
	ID                string    `json:"id"`
	UserID            int64     `json:"user_id"`
	APIKeyID          *int64    `json:"api_key_id,omitempty"`
	GroupID           int64     `json:"group_id"`
	ContextDomain     string    `json:"context_domain"`
	Protocol          string    `json:"protocol"`
	Title             string    `json:"title"`
	StartedAt         time.Time `json:"started_at"`
	LastActiveAt      time.Time `json:"last_active_at"`
	RequestCount      int       `json:"request_count"`
	TotalInputTokens  int64     `json:"total_input_tokens"`
	TotalOutputTokens int64     `json:"total_output_tokens"`
	Status            string    `json:"status"`
}

// ConversationBranchView 是分支的管理员视图。
type ConversationBranchView struct {
	ID           string    `json:"id"`
	ParentID     *string   `json:"parent_branch_id,omitempty"`
	BranchReason string    `json:"branch_reason"`
	EventCount   int       `json:"event_count"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ConversationEventView 是事件的管理员视图，Content 为解密后的明文（不可解时回退预览）。
type ConversationEventView struct {
	ID        int64     `json:"id"`
	BranchID  string    `json:"branch_id"`
	Sequence  int       `json:"sequence"`
	Role      string    `json:"role"`
	Kind      string    `json:"kind"`
	Content   string    `json:"content"`
	Model     string    `json:"model"`
	Partial   bool      `json:"partial"`
	Encrypted bool      `json:"encrypted"`
	Decrypted bool      `json:"decrypted"`
	CreatedAt time.Time `json:"created_at"`
}

// ConversationSessionDetail 是会话详情（含分支与按序事件）。
type ConversationSessionDetail struct {
	Session  ConversationSessionView  `json:"session"`
	Branches []ConversationBranchView `json:"branches"`
	Events   []ConversationEventView  `json:"events"`
}

// ListSessions 分页查询会话列表。
func (s *ConversationService) ListSessions(ctx context.Context, params pagination.PaginationParams, filters ConversationSessionListFilters) ([]ConversationSessionView, *pagination.PaginationResult, error) {
	sessions, result, err := s.repo.ListSessions(ctx, params, filters)
	if err != nil {
		return nil, nil, err
	}
	views := make([]ConversationSessionView, 0, len(sessions))
	for i := range sessions {
		views = append(views, toConversationSessionView(&sessions[i]))
	}
	return views, result, nil
}

// GetSessionDetail 返回会话详情：会话 + 分支 + 解密后的事件（按 sequence 升序）。
func (s *ConversationService) GetSessionDetail(ctx context.Context, id uuid.UUID) (*ConversationSessionDetail, error) {
	session, err := s.repo.GetSessionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	branches, err := s.repo.ListBranches(ctx, id)
	if err != nil {
		return nil, err
	}
	events, err := s.repo.ListEvents(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	detail := &ConversationSessionDetail{
		Session:  toConversationSessionView(session),
		Branches: make([]ConversationBranchView, 0, len(branches)),
		Events:   make([]ConversationEventView, 0, len(events)),
	}
	for i := range branches {
		detail.Branches = append(detail.Branches, toConversationBranchView(&branches[i]))
	}
	for i := range events {
		detail.Events = append(detail.Events, s.toConversationEventView(&events[i]))
	}
	return detail, nil
}

// DeleteSession 软删除会话。
func (s *ConversationService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteSession(ctx, id)
}

// ExportSessionText 导出单个会话为纯文本逐字稿。
func (s *ConversationService) ExportSessionText(ctx context.Context, id uuid.UUID) (filename string, content []byte, err error) {
	detail, err := s.GetSessionDetail(ctx, id)
	if err != nil {
		return "", nil, err
	}
	text := formatConversationTranscript(detail)
	filename = "conversation-" + detail.Session.ID + ".txt"
	return filename, []byte(text), nil
}

// ExportZip 按过滤条件导出所有匹配会话为 zip（每个会话一个 .txt）。
func (s *ConversationService) ExportZip(ctx context.Context, filters ConversationSessionListFilters) (filename string, content []byte, err error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	page := 1
	const pageSize = 200
	written := 0
	for {
		params := pagination.PaginationParams{Page: page, PageSize: pageSize}
		sessions, _, listErr := s.repo.ListSessions(ctx, params, filters)
		if listErr != nil {
			_ = zw.Close()
			return "", nil, listErr
		}
		if len(sessions) == 0 {
			break
		}
		for i := range sessions {
			if written >= conversationExportMaxSessions {
				break
			}
			detail, detErr := s.GetSessionDetail(ctx, sessions[i].ID)
			if detErr != nil {
				continue // 跳过无法读取的会话，不中断整体导出
			}
			entry := "conversation-" + detail.Session.ID + ".txt"
			w, createErr := zw.Create(entry)
			if createErr != nil {
				_ = zw.Close()
				return "", nil, createErr
			}
			if _, wErr := w.Write([]byte(formatConversationTranscript(detail))); wErr != nil {
				_ = zw.Close()
				return "", nil, wErr
			}
			written++
		}
		if len(sessions) < pageSize || written >= conversationExportMaxSessions {
			break
		}
		page++
	}

	if err := zw.Close(); err != nil {
		return "", nil, err
	}
	filename = fmt.Sprintf("conversations-%s.zip", time.Now().Format("20060102-150405"))
	return filename, buf.Bytes(), nil
}

func formatConversationTranscript(detail *ConversationSessionDetail) string {
	var b strings.Builder
	sess := detail.Session
	fmt.Fprintf(&b, "Conversation %s\n", sess.ID)
	fmt.Fprintf(&b, "User: %d  Upstream: %s  Protocol: %s\n", sess.UserID, sess.ContextDomain, sess.Protocol)
	fmt.Fprintf(&b, "Started: %s  Last active: %s\n", sess.StartedAt.Format(time.RFC3339), sess.LastActiveAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Requests: %d  Tokens in/out: %d/%d  Status: %s\n", sess.RequestCount, sess.TotalInputTokens, sess.TotalOutputTokens, sess.Status)
	b.WriteString(strings.Repeat("=", 60) + "\n\n")

	for _, ev := range detail.Events {
		role := strings.ToUpper(ev.Role)
		ts := ev.CreatedAt.Format(time.RFC3339)
		header := fmt.Sprintf("[#%d] %s", ev.Sequence, role)
		if ev.Kind != "" && ev.Kind != ConversationKindMessage {
			header += " (" + ev.Kind + ")"
		}
		if ev.Partial {
			header += " [partial]"
		}
		if ev.Encrypted && !ev.Decrypted {
			header += " [preview only]"
		}
		fmt.Fprintf(&b, "%s  %s:\n%s\n\n", header, ts, ev.Content)
	}
	return b.String()
}

func toConversationSessionView(e *ConversationSession) ConversationSessionView {
	return ConversationSessionView{
		ID:                e.ID.String(),
		UserID:            e.UserID,
		APIKeyID:          e.APIKeyID,
		GroupID:           e.GroupID,
		ContextDomain:     e.ContextDomain,
		Protocol:          e.Protocol,
		Title:             e.Title,
		StartedAt:         e.StartedAt,
		LastActiveAt:      e.LastActiveAt,
		RequestCount:      e.RequestCount,
		TotalInputTokens:  e.TotalInputTokens,
		TotalOutputTokens: e.TotalOutputTokens,
		Status:            e.Status,
	}
}

func toConversationBranchView(b *ConversationBranch) ConversationBranchView {
	view := ConversationBranchView{
		ID:           b.ID.String(),
		BranchReason: b.BranchReason,
		EventCount:   b.EventCount,
		Status:       b.Status,
		CreatedAt:    b.CreatedAt,
	}
	if b.ParentBranchID != nil {
		p := b.ParentBranchID.String()
		view.ParentID = &p
	}
	return view
}

// toConversationEventView 解密事件内容；key_version=0 表示明文，>0 表示 AES-GCM 密文。
func (s *ConversationService) toConversationEventView(e *ConversationEvent) ConversationEventView {
	view := ConversationEventView{
		ID:        e.ID,
		BranchID:  e.BranchID.String(),
		Sequence:  e.Sequence,
		Role:      e.Role,
		Kind:      e.Kind,
		Model:     e.Model,
		Partial:   e.Partial,
		CreatedAt: e.CreatedAt,
	}
	switch {
	case e.EncryptionKeyVersion == 0:
		// 未加密：密文列即明文。
		view.Content = string(e.ContentCiphertext)
		view.Decrypted = true
	case s.enc != nil && len(e.ContentCiphertext) > 0 && len(e.ContentNonce) > 0:
		if plain, err := s.enc.Decrypt(e.ContentCiphertext, e.ContentNonce); err == nil {
			view.Content = string(plain)
			view.Encrypted = true
			view.Decrypted = true
		} else {
			view.Content = e.ContentPreview
			view.Encrypted = true
		}
	default:
		// 已加密但无可用密钥：仅回退脱敏预览。
		view.Content = e.ContentPreview
		view.Encrypted = true
	}
	return view
}
