package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type internalChatAPIKeyRepoStub struct {
	APIKeyRepository
	key         *APIKey
	findErr     error
	findCalls   int
	createErr   error
	createCalls int
	deleted     bool
}

func (r *internalChatAPIKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	if r.key == nil {
		return nil, ErrAPIKeyNotFound
	}
	clone := *r.key
	return &clone, nil
}

func (r *internalChatAPIKeyRepoStub) FindInternalChatKey(context.Context, int64) (*APIKey, error) {
	return r.findByName(InternalChatKeyName)
}

func (r *internalChatAPIKeyRepoStub) FindInternalChatKeyByName(_ context.Context, _ int64, name string) (*APIKey, error) {
	return r.findByName(name)
}

func (r *internalChatAPIKeyRepoStub) findByName(name string) (*APIKey, error) {
	r.findCalls++
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.key == nil || r.key.Name != name {
		return nil, ErrAPIKeyNotFound
	}
	clone := *r.key
	return &clone, nil
}

func (r *internalChatAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	panic("generic API key creation must not be used for internal chat keys")
}

func (r *internalChatAPIKeyRepoStub) CreateInternalChatKey(_ context.Context, key *APIKey) error {
	r.createCalls++
	if r.createErr != nil {
		// Simulate another application instance winning the unique-index race.
		clone := *key
		clone.ID = 99
		r.key = &clone
		return r.createErr
	}
	r.key = key
	return nil
}

func (r *internalChatAPIKeyRepoStub) Update(context.Context, *APIKey, APIKeyUpdateFields) error {
	return nil
}

func (r *internalChatAPIKeyRepoStub) DeleteWithAudit(context.Context, int64) error {
	r.deleted = true
	return nil
}

type internalChatUserRepoStub struct {
	UserRepository
	user *User
}

func (r internalChatUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return r.user, nil
}

type internalChatGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (r internalChatGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	return r.group, nil
}

func TestAPIKeyServiceRejectsReservedChatNamesFromUserManagement(t *testing.T) {
	svc := &APIKeyService{}
	for _, name := range []string{InternalChatKeyName, internalChatKeyNameForGroup(7), InternalChatKeyName + "evil"} {
		_, err := svc.Create(context.Background(), 1, CreateAPIKeyRequest{Name: name})
		require.ErrorIs(t, err, ErrReservedAPIKeyName)
	}

	groupID := int64(7)
	repo := &internalChatAPIKeyRepoStub{key: &APIKey{
		ID: 12, UserID: 1, Key: "sk-internal", Name: internalChatKeyNameForGroup(groupID), GroupID: &groupID,
	}}
	svc = &APIKeyService{apiKeyRepo: repo}
	ordinaryName := "ordinary"
	_, err := svc.Update(context.Background(), 12, 1, UpdateAPIKeyRequest{Name: &ordinaryName})
	require.ErrorIs(t, err, ErrReservedAPIKeyName)
	require.ErrorIs(t, svc.Delete(context.Background(), 12, 1), ErrReservedAPIKeyName)
	require.False(t, repo.deleted)
	_, err = svc.GetByID(context.Background(), 12)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)

	repo.key.Name = ordinaryName
	reservedName := internalChatKeyNameForGroup(groupID)
	_, err = svc.Update(context.Background(), 12, 1, UpdateAPIKeyRequest{Name: &reservedName})
	require.ErrorIs(t, err, ErrReservedAPIKeyName)
}

func TestNamedInternalChatKeyRejectsInvalidReuseAndRepositoryErrors(t *testing.T) {
	groupID := int64(7)
	name := internalChatKeyNameForGroup(groupID)
	repo := &internalChatAPIKeyRepoStub{key: &APIKey{
		ID: 1, UserID: 1, Name: name, GroupID: func() *int64 { id := int64(8); return &id }(),
	}}
	svc := &APIKeyService{apiKeyRepo: repo}

	_, err := svc.getOrCreateNamedInternalChatKey(context.Background(), repo, 1, groupID, name)
	require.ErrorIs(t, err, errInternalChatKeyConflict)
	require.Zero(t, repo.createCalls)

	dbErr := errors.New("database unavailable")
	repo.key = nil
	repo.findErr = dbErr
	_, err = svc.getOrCreateNamedInternalChatKey(context.Background(), repo, 1, groupID, name)
	require.ErrorIs(t, err, dbErr)
	require.Zero(t, repo.createCalls)
}

func TestNamedInternalChatKeyRecoversDatabaseInsertRace(t *testing.T) {
	groupID := int64(7)
	name := internalChatKeyNameForGroup(groupID)
	uniqueErr := errors.New("unique constraint violation")
	repo := &internalChatAPIKeyRepoStub{createErr: uniqueErr}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   internalChatUserRepoStub{user: &User{ID: 1}},
		groupRepo:  internalChatGroupRepoStub{group: &Group{ID: groupID, Status: StatusActive}},
		cfg:        &config.Config{},
	}

	key, err := svc.getOrCreateNamedInternalChatKey(context.Background(), repo, 1, groupID, name)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, int64(99), key.ID)
	require.Equal(t, &groupID, key.GroupID)
	require.Equal(t, 1, repo.createCalls)
	require.GreaterOrEqual(t, repo.findCalls, 3)
}
