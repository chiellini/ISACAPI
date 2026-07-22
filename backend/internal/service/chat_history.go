package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/websearch"
)

// ErrChatSessionNotFound 表示会话不存在或不属于当前用户。
var ErrChatSessionNotFound = infraerrors.NotFound("CHAT_SESSION_NOT_FOUND", "chat session not found")

// ErrChatImageNotFound 表示图片不存在或不属于当前用户。
var ErrChatImageNotFound = infraerrors.NotFound("CHAT_IMAGE_NOT_FOUND", "chat image not found")

// ErrChatImageTooLarge 表示单张图片超过服务端允许的大小。
var ErrChatImageTooLarge = infraerrors.BadRequest("CHAT_IMAGE_TOO_LARGE", "chat image exceeds the maximum allowed size")

// ErrChatImageInvalid 表示上传内容不是可识别的图片数据。
var ErrChatImageInvalid = infraerrors.BadRequest("CHAT_IMAGE_INVALID", "chat image payload is not valid image data")

// chatImageMaxBytes 单张聊天图片上限（解码后）。与前端 MAX_IMAGE_BYTES 对齐。
const chatImageMaxBytes = 10 << 20 // 10 MiB

// ChatHistoryMessage 是一条聊天消息（content 可能是纯文本，也可能是携带图片引用的紧凑 JSON）。
type ChatHistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatHistorySession 是一个聊天会话。List 时不带 Messages/记忆字段，Get 时全带。
type ChatHistorySession struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	UpdatedAt time.Time `json:"updated_at"`
	Summary   string    `json:"summary"`
	Memory    string    `json:"memory"`
	// SummarizedCount 是已折叠进 Summary 的前缀消息条数（其后为短期原文上下文）。
	SummarizedCount int                  `json:"summarized_count"`
	Messages        []ChatHistoryMessage `json:"messages,omitempty"`
}

// ChatSessionUpdate 描述一次会话更新。指针字段为 nil 表示「保持原值不动」，
// 从而让「仅重命名」与「整轮保存（含消息与记忆）」共用一个入口而互不干扰。
type ChatSessionUpdate struct {
	Title           string
	Model           string
	Summary         *string
	Memory          *string
	SummarizedCount *int
	// Messages 为 nil 表示不改动消息；非 nil（含空切片）表示整体替换。
	Messages []ChatHistoryMessage
}

// ChatHistoryRepository 聊天历史仓储（原生 SQL 实现，不依赖 ent 生成）。
type ChatHistoryRepository interface {
	ListSessions(ctx context.Context, userID int64) ([]ChatHistorySession, error)
	GetSession(ctx context.Context, userID, id int64) (*ChatHistorySession, error)
	CreateSession(ctx context.Context, userID int64, title, model string) (int64, error)
	// UpdateSession 按 ChatSessionUpdate 的语义更新会话元数据 / 记忆 / 消息。
	UpdateSession(ctx context.Context, userID, id int64, in ChatSessionUpdate) error
	DeleteSession(ctx context.Context, userID, id int64) error

	// SaveImage 把一张图片写入会话（校验会话归属），失败返回 error。
	SaveImage(ctx context.Context, userID, sessionID int64, id, mime string, data []byte) error
	// GetImage 读回一张图片（按 user 隔离）。不存在返回 ErrChatImageNotFound。
	GetImage(ctx context.Context, userID int64, id string) (mime string, data []byte, err error)
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

func (s *ChatHistoryService) Update(ctx context.Context, userID, id int64, in ChatSessionUpdate) error {
	return s.repo.UpdateSession(ctx, userID, id, in)
}

func (s *ChatHistoryService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.DeleteSession(ctx, userID, id)
}

// SaveImage 解码传入的图片（data URL 或裸 base64），落库并返回生成的图片 ID。
func (s *ChatHistoryService) SaveImage(ctx context.Context, userID, sessionID int64, payload string) (id, mime string, err error) {
	data, mime, err := decodeChatImage(payload)
	if err != nil {
		return "", "", err
	}
	if len(data) == 0 {
		return "", "", ErrChatImageInvalid
	}
	if len(data) > chatImageMaxBytes {
		return "", "", ErrChatImageTooLarge
	}
	id, err = newChatImageID()
	if err != nil {
		return "", "", err
	}
	if err := s.repo.SaveImage(ctx, userID, sessionID, id, mime, data); err != nil {
		return "", "", err
	}
	return id, mime, nil
}

// GetImage 读回图片字节与其 MIME 类型。
func (s *ChatHistoryService) GetImage(ctx context.Context, userID int64, id string) (mime string, data []byte, err error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", nil, ErrChatImageNotFound
	}
	return s.repo.GetImage(ctx, userID, id)
}

// ChatSearchResult 是回给前端的一条联网搜索结果（供工具回灌与来源引用展示）。
type ChatSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// ErrChatSearchUnavailable 表示平台未配置可用的联网搜索提供方（或搜索本次失败）。
var ErrChatSearchUnavailable = infraerrors.ServiceUnavailable("CHAT_SEARCH_UNAVAILABLE", "web search is not available")

// ErrChatSearchEmptyQuery 表示搜索查询为空。
var ErrChatSearchEmptyQuery = infraerrors.BadRequest("CHAT_SEARCH_EMPTY_QUERY", "search query is required")

const (
	chatSearchDefaultMaxResults = 5
	chatSearchMaxResults        = 8
)

// SearchWeb 通过平台已配置的联网搜索提供方（Brave/Tavily，与网关搜索复用同一配额）执行一次搜索。
// 未配置或本次失败时返回 ErrChatSearchUnavailable（调用方据此让模型「无搜索」继续作答）。
func (s *ChatHistoryService) SearchWeb(ctx context.Context, query string, maxResults int) ([]ChatSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrChatSearchEmptyQuery
	}
	mgr := getWebSearchManager()
	if mgr == nil {
		return nil, ErrChatSearchUnavailable
	}
	if maxResults <= 0 {
		maxResults = chatSearchDefaultMaxResults
	}
	if maxResults > chatSearchMaxResults {
		maxResults = chatSearchMaxResults
	}
	resp, _, err := mgr.SearchWithBestProvider(ctx, websearch.SearchRequest{Query: query, MaxResults: maxResults})
	if err != nil {
		return nil, ErrChatSearchUnavailable
	}
	out := make([]ChatSearchResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		out = append(out, ChatSearchResult{Title: r.Title, URL: r.URL, Snippet: r.Snippet})
	}
	return out, nil
}

// WebSearchAvailable 报告当前是否已配置可用的联网搜索（供前端决定是否展示开关）。
func (s *ChatHistoryService) WebSearchAvailable() bool {
	return getWebSearchManager() != nil
}

// decodeChatImage 支持两种输入：完整 data URL（data:image/png;base64,xxxx）与裸 base64。
// 返回解码后的字节与推断出的 MIME 类型（默认 image/png）。
func decodeChatImage(payload string) ([]byte, string, error) {
	raw := strings.TrimSpace(payload)
	if raw == "" {
		return nil, "", ErrChatImageInvalid
	}
	mime := "image/png"
	if strings.HasPrefix(raw, "data:") {
		comma := strings.IndexByte(raw, ',')
		if comma < 0 {
			return nil, "", ErrChatImageInvalid
		}
		header := raw[len("data:"):comma]
		raw = raw[comma+1:]
		if !strings.Contains(header, "base64") {
			// 仅支持 base64 编码的 data URL（生图与前端上传均为 base64）。
			return nil, "", ErrChatImageInvalid
		}
		if mt := strings.TrimSpace(strings.Split(header, ";")[0]); strings.HasPrefix(mt, "image/") {
			mime = mt
		}
	}
	// 去除可能的空白/换行后解码。
	raw = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, raw)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", ErrChatImageInvalid
	}
	// 依据实际字节纠正 MIME（防止前端标注错误），仅当能识别为图片时覆盖。
	if detected := detectImageContentType(data); strings.HasPrefix(detected, "image/") {
		mime = detected
	}
	return data, mime, nil
}

// newChatImageID 生成一个不可枚举的随机图片 ID。
func newChatImageID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate chat image id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
