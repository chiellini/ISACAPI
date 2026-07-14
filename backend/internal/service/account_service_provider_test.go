//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type providerScopedAccountRepoStub struct {
	AccountRepository
	account      *Account
	accounts     []Account
	result       *pagination.PaginationResult
	err          error
	gotID        int64
	gotProvider  int64
	updateCalls  int
	updateInput  *ProviderAccountUpdateInput
	deleteCalls  int
	gotParams    pagination.PaginationParams
	gotPlatform  string
	gotGroupID   int64
}

func (r *providerScopedAccountRepoStub) UpdateForProvider(_ context.Context, id, providerID int64, input *ProviderAccountUpdateInput) (*Account, error) {
	r.updateCalls++
	r.gotID = id
	r.gotProvider = providerID
	r.updateInput = input
	return r.account, r.err
}

func (r *providerScopedAccountRepoStub) GetByIDForProvider(_ context.Context, id, providerID int64) (*Account, error) {
	r.gotID = id
	r.gotProvider = providerID
	return r.account, r.err
}

func (r *providerScopedAccountRepoStub) ListWithFiltersForProvider(
	_ context.Context,
	params pagination.PaginationParams,
	platform, _, _, _ string,
	groupID int64,
	_ string,
	providerID int64,
) ([]Account, *pagination.PaginationResult, error) {
	r.gotParams = params
	r.gotPlatform = platform
	r.gotGroupID = groupID
	r.gotProvider = providerID
	return r.accounts, r.result, r.err
}

func (r *providerScopedAccountRepoStub) DeleteForProvider(_ context.Context, id, providerID int64) error {
	r.deleteCalls++
	r.gotID = id
	r.gotProvider = providerID
	return r.err
}

type providerUnsupportedAccountRepoStub struct{ AccountRepository }

func TestAccountServiceProviderScopedWrappers(t *testing.T) {
	repo := &providerScopedAccountRepoStub{
		account:  &Account{ID: 41},
		accounts: []Account{{ID: 41}},
		result:   &pagination.PaginationResult{Total: 1},
	}
	svc := NewAccountService(repo, nil)

	account, err := svc.GetByIDForProvider(context.Background(), 41, 9)
	require.NoError(t, err)
	require.Same(t, repo.account, account)
	require.Equal(t, int64(41), repo.gotID)
	require.Equal(t, int64(9), repo.gotProvider)

	params := pagination.PaginationParams{Page: 2, PageSize: 25}
	accounts, result, err := svc.ListWithFiltersForProvider(
		context.Background(), params, PlatformOpenAI, AccountTypeOAuth, StatusActive, "shared", 7, "", 9,
	)
	require.NoError(t, err)
	require.Equal(t, repo.accounts, accounts)
	require.Same(t, repo.result, result)
	require.Equal(t, params, repo.gotParams)
	require.Equal(t, PlatformOpenAI, repo.gotPlatform)
	require.Equal(t, int64(7), repo.gotGroupID)
	require.Equal(t, int64(9), repo.gotProvider)

	name := "updated"
	updated, err := svc.UpdateForProvider(context.Background(), 41, 9, &ProviderAccountUpdateInput{Name: &name})
	require.NoError(t, err)
	require.Same(t, repo.account, updated)
	require.Equal(t, 1, repo.updateCalls)
	require.Equal(t, "updated", *repo.updateInput.Name)
	require.Equal(t, int64(41), repo.gotID)
	require.Equal(t, int64(9), repo.gotProvider)

	require.NoError(t, svc.DeleteForProvider(context.Background(), 41, 9))
	require.Equal(t, 1, repo.deleteCalls)
	require.Equal(t, int64(41), repo.gotID)
	require.Equal(t, int64(9), repo.gotProvider)
}

func TestAccountServiceProviderScopedWrappersPreserveOwnershipMiss(t *testing.T) {
	repo := &providerScopedAccountRepoStub{err: ErrAccountNotFound}
	svc := NewAccountService(repo, nil)

	_, err := svc.GetByIDForProvider(context.Background(), 41, 9)
	require.ErrorIs(t, err, ErrAccountNotFound)
	_, err = svc.UpdateForProvider(context.Background(), 41, 9, &ProviderAccountUpdateInput{})
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestAccountServiceProviderScopedWrappersRequireCapability(t *testing.T) {
	svc := NewAccountService(&providerUnsupportedAccountRepoStub{}, nil)

	_, err := svc.GetByIDForProvider(context.Background(), 41, 9)
	require.ErrorIs(t, err, ErrProviderAccountRepositoryUnsupported)
	_, _, err = svc.ListWithFiltersForProvider(
		context.Background(), pagination.PaginationParams{}, "", "", "", "", 0, "", 9,
	)
	require.ErrorIs(t, err, ErrProviderAccountRepositoryUnsupported)
	_, err = svc.UpdateForProvider(context.Background(), 41, 9, &ProviderAccountUpdateInput{})
	require.ErrorIs(t, err, ErrProviderAccountRepositoryUnsupported)
	err = svc.DeleteForProvider(context.Background(), 41, 9)
	require.ErrorIs(t, err, ErrProviderAccountRepositoryUnsupported)
}
