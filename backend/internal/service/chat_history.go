package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ErrChatSessionNotFound 表示会话不存在或不属于当前用户。
var ErrChatSessionNotFound = infraerrors.NotFound("CHAT_SESSION_NOT_FOUND", "chat session not found")

// ChatHistoryMessage 是一条聊天消息（仅文本，附件原文不持久化）。
type ChatHistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatHistorySession 是一个聊天会话。List 时不带 Messages，Get 时带。
type ChatHistorySession struct {
	ID        int64                `json:"id"`
	Title     string               `json:"title"`
	Model     string               `json:"model"`
	UpdatedAt time.Time            `json:"updated_at"`
	Messages  []ChatHistoryMessage `json:"messages,omitempty"`
}

// ChatHistoryRepository 聊天历史仓储（原生 SQL 实现，不依赖 ent 生成）。
type ChatHistoryRepository interface {
	ListSessions(ctx context.Context, userID int64) ([]ChatHistorySession, error)
	GetSession(ctx context.Context, userID, id int64) (*ChatHistorySession, error)
	CreateSession(ctx context.Context, userID int64, title, model string) (int64, error)
	// UpdateSession 覆盖会话标题/模型，并整体替换消息列表（每轮对话结束后保存）。
	UpdateSession(ctx context.Context, userID, id int64, title, model string, msgs []ChatHistoryMessage) error
	DeleteSession(ctx context.Context, userID, id int64) error
}

// ChatHistoryService 内置聊天会话历史服务（JWT 会话鉴权，按 user 隔离）。
type ChatHistoryService struct {
	repo ChatHistoryRepository
}

func NewChatHistoryService(repo ChatHistoryRepository) *ChatHistoryService {
	return &ChatHistoryService{repo: repo}
}

func (s *ChatHistoryService) List(ctx context.Context, userID int64) ([]ChatHistorySession, error) {
	return s.repo.ListSessions(ctx, userID)
}

func (s *ChatHistoryService) Get(ctx context.Context, userID, id int64) (*ChatHistorySession, error) {
	return s.repo.GetSession(ctx, userID, id)
}

func (s *ChatHistoryService) Create(ctx context.Context, userID int64, title, model string) (int64, error) {
	return s.repo.CreateSession(ctx, userID, title, model)
}

func (s *ChatHistoryService) Update(ctx context.Context, userID, id int64, title, model string, msgs []ChatHistoryMessage) error {
	return s.repo.UpdateSession(ctx, userID, id, title, model, msgs)
}

func (s *ChatHistoryService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.DeleteSession(ctx, userID, id)
}
