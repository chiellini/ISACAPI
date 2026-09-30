package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeAssistantsRepo 是内存版助手仓储，覆盖 CRUD + 计数语义。
type fakeAssistantsRepo struct {
	byUser map[int64][]ChatAssistant
	nextID int64
}

func newFakeAssistantsRepo() *fakeAssistantsRepo {
	return &fakeAssistantsRepo{byUser: map[int64][]ChatAssistant{}, nextID: 10}
}

func (f *fakeAssistantsRepo) List(ctx context.Context, userID int64) ([]ChatAssistant, error) {
	return f.byUser[userID], nil
}

func (f *fakeAssistantsRepo) Get(ctx context.Context, userID, id int64) (*ChatAssistant, error) {
	for _, a := range f.byUser[userID] {
		if a.ID == id {
			out := a
			return &out, nil
		}
	}
	return nil, ErrChatAssistantNotFound
}

func (f *fakeAssistantsRepo) Create(ctx context.Context, userID int64, in ChatAssistantInput) (int64, error) {
	f.nextID++
	f.byUser[userID] = append(f.byUser[userID], ChatAssistant{
		ID: f.nextID, Name: in.Name, Description: in.Description,
		SystemPrompt: in.SystemPrompt, Model: in.Model, OpeningMessage: in.OpeningMessage,
	})
	return f.nextID, nil
}

func (f *fakeAssistantsRepo) Update(ctx context.Context, userID, id int64, in ChatAssistantInput) error {
	for i := range f.byUser[userID] {
		if f.byUser[userID][i].ID == id {
			f.byUser[userID][i] = ChatAssistant{
				ID: id, Name: in.Name, Description: in.Description,
				SystemPrompt: in.SystemPrompt, Model: in.Model, OpeningMessage: in.OpeningMessage,
			}
			return nil
		}
	}
	return ErrChatAssistantNotFound
}

func (f *fakeAssistantsRepo) Delete(ctx context.Context, userID, id int64) error {
	list := f.byUser[userID]
	for i := range list {
		if list[i].ID == id {
			f.byUser[userID] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return ErrChatAssistantNotFound
}

func (f *fakeAssistantsRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	return len(f.byUser[userID]), nil
}

func TestChatAssistantCreateRejectsBlankName(t *testing.T) {
	svc := NewChatAssistantsService(newFakeAssistantsRepo())
	_, err := svc.Create(context.Background(), 1, ChatAssistantInput{Name: "   "})
	require.ErrorIs(t, err, ErrChatAssistantInvalidInput)
}

func TestChatAssistantInputTrimsAndKeepsValid(t *testing.T) {
	svc := NewChatAssistantsService(newFakeAssistantsRepo())
	id, err := svc.Create(context.Background(), 1, ChatAssistantInput{
		Name:         "  翻译官  ",
		SystemPrompt: " 你是专业译员 ",
	})
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := svc.Get(context.Background(), 1, id)
	require.NoError(t, err)
	require.Equal(t, "翻译官", got.Name)
	require.Equal(t, "你是专业译员", got.SystemPrompt)
}

func TestChatAssistantRejectsOversizedPrompt(t *testing.T) {
	svc := NewChatAssistantsService(newFakeAssistantsRepo())
	_, err := svc.Create(context.Background(), 1, ChatAssistantInput{
		Name:         "a",
		SystemPrompt: strings.Repeat("x", chatAssistantMaxPromptBytes+1),
	})
	require.ErrorIs(t, err, ErrChatAssistantInvalidInput)
}

func TestChatAssistantCreateEnforcesPerUserLimit(t *testing.T) {
	repo := newFakeAssistantsRepo()
	svc := NewChatAssistantsService(repo)
	for i := 0; i < maxChatAssistantsPerUser; i++ {
		_, err := svc.Create(context.Background(), 7, ChatAssistantInput{Name: "a"})
		require.NoError(t, err)
	}
	_, err := svc.Create(context.Background(), 7, ChatAssistantInput{Name: "overflow"})
	require.ErrorIs(t, err, ErrChatAssistantLimitReached)

	// 上限按用户隔离：另一个用户仍可创建。
	_, err = svc.Create(context.Background(), 8, ChatAssistantInput{Name: "ok"})
	require.NoError(t, err)
}

func TestChatAssistantUpdateValidatesOwnership(t *testing.T) {
	repo := newFakeAssistantsRepo()
	svc := NewChatAssistantsService(repo)
	id, err := svc.Create(context.Background(), 1, ChatAssistantInput{Name: "a"})
	require.NoError(t, err)

	err = svc.Update(context.Background(), 2, id, ChatAssistantInput{Name: "b"})
	require.ErrorIs(t, err, ErrChatAssistantNotFound)
	err = svc.Delete(context.Background(), 2, id)
	require.ErrorIs(t, err, ErrChatAssistantNotFound)
}
