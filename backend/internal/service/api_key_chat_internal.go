package service

import (
	"context"
	"fmt"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

// InternalChatKeyName 是内置聊天 Playground 专用 API Key 的保留名称。
// 该名称用于识别、去重，并在用户 Key 列表中过滤隐藏（见 apiKeyRepository.ListByUserID）。
const InternalChatKeyName = "__chat_playground__"

// ErrNoAvailableChatGroup 表示用户当前没有任何可用分组，无法开启内置聊天。
var ErrNoAvailableChatGroup = infraerrors.Forbidden("NO_AVAILABLE_CHAT_GROUP", "no available group for chat")

// internalChatKeySF 串行化同一用户首次内置 Key 的并发创建，避免重复落库。
var internalChatKeySF singleflight.Group

// GetOrCreateInternalChatKey 返回用户的内置聊天 Key，不存在则惰性创建。
//
// 内置 Key 绑定到用户第一个可用分组、不限额（quota=0），真正的约束来自用户余额/订阅。
// 返回值仅保证 Key 字段可用——调用方（chatBridge）只需要 Key 字符串注入 Authorization，
// 后续 apiKeyAuth 会用该 Key 重新做完整鉴权与计费加载。
func (s *APIKeyService) GetOrCreateInternalChatKey(ctx context.Context, userID int64) (*APIKey, error) {
	if existing, err := s.apiKeyRepo.FindInternalChatKey(ctx, userID); err == nil && existing != nil {
		return existing, nil
	}

	v, err, _ := internalChatKeySF.Do(fmt.Sprintf("internal-chat-key:%d", userID), func() (interface{}, error) {
		// 双重检查：抢到 singleflight 的协程可能在等待期间已被其它协程创建。
		if existing, err := s.apiKeyRepo.FindInternalChatKey(ctx, userID); err == nil && existing != nil {
			return existing, nil
		}

		groupID, err := s.defaultChatGroupID(ctx, userID)
		if err != nil {
			return nil, err
		}

		return s.Create(ctx, userID, CreateAPIKeyRequest{
			Name:    InternalChatKeyName,
			GroupID: groupID,
			Quota:   0, // 不限额，约束来自用户余额/订阅
		})
	})
	if err != nil {
		return nil, err
	}
	return v.(*APIKey), nil
}

// defaultChatGroupID 选择用户的默认聊天分组。
// 优先选 OpenAI 平台分组（聊天与生图都依赖 OpenAI），没有则回退到第一个可用分组。
func (s *APIKeyService) defaultChatGroupID(ctx context.Context, userID int64) (*int64, error) {
	groups, err := s.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve chat group: %w", err)
	}
	if len(groups) == 0 {
		return nil, ErrNoAvailableChatGroup
	}
	for i := range groups {
		if groups[i].Platform == PlatformOpenAI {
			return &groups[i].ID, nil
		}
	}
	id := groups[0].ID
	return &id, nil
}
