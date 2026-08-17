package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

// InternalChatKeyName 是内置聊天 Playground 专用 API Key 的保留名称。
// 该名称用于识别、去重，并在用户 Key 列表中过滤隐藏（见 apiKeyRepository.ListByUserID）。
const InternalChatKeyName = "__chat_playground__"

const internalChatKeyNamePrefix = InternalChatKeyName + ":"

type internalChatKeyRepository interface {
	FindInternalChatKeyByName(ctx context.Context, userID int64, name string) (*APIKey, error)
	CreateInternalChatKey(ctx context.Context, key *APIKey) error
}

func internalChatKeyNameForGroup(groupID int64) string {
	return internalChatKeyNamePrefix + strconv.FormatInt(groupID, 10)
}

func isInternalChatKeyName(name string) bool {
	return strings.HasPrefix(name, InternalChatKeyName)
}

// ErrNoAvailableChatGroup 表示用户当前没有任何可用分组，无法开启内置聊天。
var ErrNoAvailableChatGroup = infraerrors.Forbidden("NO_AVAILABLE_CHAT_GROUP", "no available group for chat")

var errInternalChatKeyConflict = infraerrors.Conflict("INTERNAL_CHAT_KEY_CONFLICT", "an internal chat key has an invalid owner or group binding")

// internalChatKeySF 串行化同一用户首次内置 Key 的并发创建，避免重复落库。
var internalChatKeySF singleflight.Group

// GetOrCreateInternalChatKey 返回用户的内置聊天 Key，不存在则惰性创建。
//
// 内置 Key 绑定到用户第一个可用分组、不限额（quota=0），真正的约束来自用户余额/订阅。
// 返回值仅保证 Key 字段可用——调用方（chatBridge）只需要 Key 字符串注入 Authorization，
// 后续 apiKeyAuth 会用该 Key 重新做完整鉴权与计费加载。
func (s *APIKeyService) GetOrCreateInternalChatKey(ctx context.Context, userID int64) (*APIKey, error) {
	groupID, err := s.defaultChatGroupID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Never reuse the legacy unscoped name: before names were reserved, a user
	// could pre-create that row and retain the credential. Migration 222 retires
	// every historical reserved key; scoped keys are recreated lazily below.
	return s.GetOrCreateInternalChatKeyForGroup(ctx, userID, *groupID)
}

// GetOrCreateInternalChatKeyForGroup uses a stable, per-group internal key.
// Separate keys avoid a cross-request race where switching GPT/Claude/Gemini
// would otherwise mutate one shared credential's group assignment.
func (s *APIKeyService) GetOrCreateInternalChatKeyForGroup(ctx context.Context, userID, groupID int64) (*APIKey, error) {
	if groupID <= 0 {
		return nil, ErrNoAvailableChatGroup
	}
	groups, err := s.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve chat group: %w", err)
	}
	allowed := false
	for i := range groups {
		if groups[i].ID == groupID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, ErrGroupNotAllowed
	}

	repo, ok := s.apiKeyRepo.(internalChatKeyRepository)
	if !ok {
		return nil, fmt.Errorf("api key repository does not support named internal chat keys")
	}
	name := internalChatKeyNameForGroup(groupID)
	return s.getOrCreateNamedInternalChatKey(ctx, repo, userID, groupID, name)
}

func (s *APIKeyService) getOrCreateNamedInternalChatKey(
	ctx context.Context,
	repo internalChatKeyRepository,
	userID, groupID int64,
	name string,
) (*APIKey, error) {
	existing, findErr := repo.FindInternalChatKeyByName(ctx, userID, name)
	if findErr == nil {
		if err := validateInternalChatKey(existing, userID, name, groupID); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !errors.Is(findErr, ErrAPIKeyNotFound) {
		return nil, fmt.Errorf("find internal chat key: %w", findErr)
	}

	v, err, _ := internalChatKeySF.Do(fmt.Sprintf("internal-chat-key:%d:%d", userID, groupID), func() (any, error) {
		existing, findErr := repo.FindInternalChatKeyByName(ctx, userID, name)
		if findErr == nil {
			if err := validateInternalChatKey(existing, userID, name, groupID); err != nil {
				return nil, err
			}
			return existing, nil
		}
		if !errors.Is(findErr, ErrAPIKeyNotFound) {
			return nil, fmt.Errorf("find internal chat key: %w", findErr)
		}

		key, keyErr := s.GenerateKey()
		if keyErr != nil {
			return nil, fmt.Errorf("generate internal chat key: %w", keyErr)
		}
		created := &APIKey{
			UserID:  userID,
			Key:     key,
			Name:    name,
			GroupID: &groupID,
			Status:  StatusAPIKeyActive,
		}
		createErr := repo.CreateInternalChatKey(ctx, created)
		if createErr == nil {
			s.InvalidateAuthCacheByKey(ctx, created.Key)
			return created, nil
		}

		// Another process may have won the insert against the database partial
		// unique index. Re-read the winner; otherwise preserve the real create error.
		winner, winnerErr := repo.FindInternalChatKeyByName(ctx, userID, name)
		if winnerErr == nil {
			if err := validateInternalChatKey(winner, userID, name, groupID); err != nil {
				return nil, err
			}
			return winner, nil
		}
		if !errors.Is(winnerErr, ErrAPIKeyNotFound) {
			return nil, errors.Join(createErr, fmt.Errorf("re-read internal chat key: %w", winnerErr))
		}
		return nil, createErr
	})
	if err != nil {
		return nil, err
	}
	apiKey, ok := v.(*APIKey)
	if !ok || apiKey == nil {
		return nil, fmt.Errorf("internal chat key creation returned an unexpected result")
	}
	return apiKey, nil
}

func validateInternalChatKey(apiKey *APIKey, userID int64, name string, groupID int64) error {
	if apiKey == nil || apiKey.UserID != userID || apiKey.Name != name || apiKey.GroupID == nil || *apiKey.GroupID != groupID {
		return errInternalChatKeyConflict
	}
	return nil
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
