package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ErrChatAssistantNotFound 表示助手不存在或不属于当前用户。
var ErrChatAssistantNotFound = infraerrors.NotFound("CHAT_ASSISTANT_NOT_FOUND", "chat assistant not found")

// ErrChatAssistantLimitReached 表示用户创建的助手数量已达上限。
var ErrChatAssistantLimitReached = infraerrors.BadRequest("CHAT_ASSISTANT_LIMIT_REACHED", "chat assistant limit reached")

// ErrChatAssistantInvalidInput 表示助手字段校验失败。
var ErrChatAssistantInvalidInput = infraerrors.BadRequest("CHAT_ASSISTANT_INVALID_INPUT", "chat assistant input is invalid")

const (
	// maxChatAssistantsPerUser 每用户可保存的助手数量上限。
	maxChatAssistantsPerUser = 20
	// chatAssistantMaxPromptBytes 人设 prompt 的字节上限（与 maxChatPromptBytes 同量级，
	// 但属于用户自填内容，占用更小）。
	chatAssistantMaxPromptBytes = 8 << 10
)

const (
	chatAssistantMaxNameRunes        = 100
	chatAssistantMaxDescriptionRunes = 500
	chatAssistantMaxModelRunes       = 100
	chatAssistantMaxOpeningRunes     = 2000
)

// ChatAssistant 是一个用户自定义助手（人设 + 默认模型 + 开场白）。
type ChatAssistant struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	SystemPrompt   string    `json:"system_prompt"`
	Model          string    `json:"model"`
	OpeningMessage string    `json:"opening_message"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ChatAssistantInput 是创建/更新助手的请求字段。
type ChatAssistantInput struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	SystemPrompt   string `json:"system_prompt"`
	Model          string `json:"model"`
	OpeningMessage string `json:"opening_message"`
}

// ChatAssistantsRepository 用户助手仓储（原生 SQL 实现，不依赖 ent 生成）。
type ChatAssistantsRepository interface {
	List(ctx context.Context, userID int64) ([]ChatAssistant, error)
	Get(ctx context.Context, userID, id int64) (*ChatAssistant, error)
	Create(ctx context.Context, userID int64, in ChatAssistantInput) (int64, error)
	Update(ctx context.Context, userID, id int64, in ChatAssistantInput) error
	Delete(ctx context.Context, userID, id int64) error
	CountByUser(ctx context.Context, userID int64) (int, error)
}

// ChatAssistantsService 内置聊天的用户助手服务（JWT 会话鉴权，按 user 隔离）。
//
// 模型字段不做服务端策略校验：真正的边界在对话时由 ChatPolicyMiddleware 强制
// （公开别名 → 上游模型，未配置的模型直接拒绝），助手里的过期模型只会被前端
// 回退到可用模型，无法绕过策略。
type ChatAssistantsService struct {
	repo ChatAssistantsRepository
}

func NewChatAssistantsService(repo ChatAssistantsRepository) *ChatAssistantsService {
	return &ChatAssistantsService{repo: repo}
}

// normalizeChatAssistantInput 裁剪并校验字段，返回规范化后的输入。
func normalizeChatAssistantInput(in ChatAssistantInput) (ChatAssistantInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.SystemPrompt = strings.TrimSpace(in.SystemPrompt)
	in.Model = strings.TrimSpace(in.Model)
	in.OpeningMessage = strings.TrimSpace(in.OpeningMessage)

	if in.Name == "" {
		return in, ErrChatAssistantInvalidInput
	}
	if utf8.RuneCountInString(in.Name) > chatAssistantMaxNameRunes ||
		utf8.RuneCountInString(in.Description) > chatAssistantMaxDescriptionRunes ||
		len(in.SystemPrompt) > chatAssistantMaxPromptBytes ||
		utf8.RuneCountInString(in.Model) > chatAssistantMaxModelRunes ||
		utf8.RuneCountInString(in.OpeningMessage) > chatAssistantMaxOpeningRunes {
		return in, ErrChatAssistantInvalidInput
	}
	return in, nil
}

func (s *ChatAssistantsService) List(ctx context.Context, userID int64) ([]ChatAssistant, error) {
	return s.repo.List(ctx, userID)
}

func (s *ChatAssistantsService) Get(ctx context.Context, userID, id int64) (*ChatAssistant, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *ChatAssistantsService) Create(ctx context.Context, userID int64, in ChatAssistantInput) (int64, error) {
	normalized, err := normalizeChatAssistantInput(in)
	if err != nil {
		return 0, err
	}
	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if count >= maxChatAssistantsPerUser {
		return 0, ErrChatAssistantLimitReached
	}
	return s.repo.Create(ctx, userID, normalized)
}

func (s *ChatAssistantsService) Update(ctx context.Context, userID, id int64, in ChatAssistantInput) error {
	normalized, err := normalizeChatAssistantInput(in)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, userID, id, normalized)
}

func (s *ChatAssistantsService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}
