//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type providerOwnershipAccountRepoStub struct {
	AccountRepository
	accounts map[int64]*Account
	nextID   int64
	updates  []int64
}

type providerOwnershipGroupRepoStub struct {
	GroupRepository
	groups map[int64]*Group
}

func (r *providerOwnershipGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

func (r *providerOwnershipAccountRepoStub) GetByID(_ context.Context, id int64) (*Account, error) {
	account, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (r *providerOwnershipAccountRepoStub) Create(_ context.Context, account *Account) error {
	if account.ID == 0 {
		r.nextID++
		account.ID = r.nextID
	}
	r.accounts[account.ID] = account
	return nil
}

func (r *providerOwnershipAccountRepoStub) Update(_ context.Context, account *Account) error {
	r.accounts[account.ID] = account
	r.updates = append(r.updates, account.ID)
	return nil
}

func (r *providerOwnershipAccountRepoStub) ListShadowsByParent(_ context.Context, parentID int64) ([]*Account, error) {
	shadows := make([]*Account, 0)
	for _, account := range r.accounts {
		if account.ParentAccountID != nil && *account.ParentAccountID == parentID && account.IsCredentialShadow() {
			shadows = append(shadows, account)
		}
	}
	return shadows, nil
}

func (r *providerOwnershipAccountRepoStub) BindGroups(_ context.Context, accountID int64, groupIDs []int64) error {
	r.accounts[accountID].GroupIDs = append([]int64(nil), groupIDs...)
	return nil
}

func providerIDUpdate(value *int64) **int64 { return &value }

func TestAdminUpdateAccountProviderFreezesLoadAndPropagatesToShadow(t *testing.T) {
	parentID := int64(1)
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{
		parentID: {ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 6},
		2: {
			ID: 2, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark, Concurrency: 2,
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}
	providerID := int64(9)

	updated, err := svc.UpdateAccount(context.Background(), parentID, &UpdateAccountInput{
		ProviderID: providerIDUpdate(&providerID),
	})
	require.NoError(t, err)
	require.Equal(t, &providerID, updated.ProviderID)
	require.NotNil(t, updated.LoadFactor)
	require.Equal(t, 6, *updated.LoadFactor)
	require.Equal(t, &providerID, repo.accounts[2].ProviderID)
	require.NotNil(t, repo.accounts[2].LoadFactor)
	require.Equal(t, 2, *repo.accounts[2].LoadFactor)
}

func TestAdminUpdateAccountProviderPreservesExplicitLoadFactor(t *testing.T) {
	loadFactor := 8
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{
		1: {ID: 1, Concurrency: 6, LoadFactor: &loadFactor},
	}}
	svc := &adminServiceImpl{accountRepo: repo}
	providerID := int64(9)

	updated, err := svc.UpdateAccount(context.Background(), 1, &UpdateAccountInput{
		ProviderID: providerIDUpdate(&providerID),
	})
	require.NoError(t, err)
	require.Equal(t, 8, *updated.LoadFactor)
}

func TestAdminUpdateAccountProviderClearPropagatesWithoutFreezingLoad(t *testing.T) {
	parentID := int64(1)
	oldProviderID := int64(7)
	frozenShadowLoad := 4
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{
		parentID: {ID: parentID, ProviderID: &oldProviderID, Concurrency: 4},
		2: {
			ID: 2, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			ProviderID: &oldProviderID, ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark,
			LoadFactor: &frozenShadowLoad,
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}

	updated, err := svc.UpdateAccount(context.Background(), parentID, &UpdateAccountInput{
		ProviderID: providerIDUpdate(nil),
	})
	require.NoError(t, err)
	require.Nil(t, updated.ProviderID)
	require.Nil(t, updated.LoadFactor)
	require.Nil(t, repo.accounts[2].ProviderID)
	require.NotNil(t, repo.accounts[2].LoadFactor)
	require.Equal(t, frozenShadowLoad, *repo.accounts[2].LoadFactor)
}

func TestAdminUpdateAccountProviderOwnershipGuardAndShadowAssignment(t *testing.T) {
	ownerID := int64(7)
	parentID := int64(1)
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{
		parentID: {ID: parentID, ProviderID: &ownerID},
		2: {
			ID: 2, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			ProviderID: &ownerID, ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark,
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}
	wrongOwnerID := int64(8)

	_, err := svc.UpdateAccount(context.Background(), parentID, &UpdateAccountInput{RequiredProviderID: &wrongOwnerID})
	require.ErrorIs(t, err, ErrAccountNotFound)
	require.Empty(t, repo.updates)

	newProviderID := int64(9)
	_, err = svc.UpdateAccount(context.Background(), 2, &UpdateAccountInput{
		ProviderID: providerIDUpdate(&newProviderID),
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "provider is inherited")
}

func TestAdminCreateShadowInheritsProvider(t *testing.T) {
	parentID := int64(10)
	providerID := int64(7)
	repo := &providerOwnershipAccountRepoStub{
		accounts: map[int64]*Account{
			parentID: {
				ID: parentID, Name: "parent", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				ProviderID: &providerID, Concurrency: 3, Priority: 50, GroupIDs: []int64{4},
			},
		},
		nextID: 20,
	}
	svc := &adminServiceImpl{accountRepo: repo}

	shadow, err := svc.CreateShadow(context.Background(), parentID, ShadowOptions{})
	require.NoError(t, err)
	require.Equal(t, &providerID, shadow.ProviderID)
	require.Equal(t, []int64{4}, shadow.GroupIDs)
	require.NotNil(t, shadow.LoadFactor)
	require.Equal(t, 3, *shadow.LoadFactor)
}

func TestAdminProviderPropagationPreservesExplicitShadowLoadFactor(t *testing.T) {
	parentID := int64(1)
	explicitShadowLoad := 9
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{
		parentID: {ID: parentID, Concurrency: 6},
		2: {
			ID: 2, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark,
			Concurrency: 2, LoadFactor: &explicitShadowLoad,
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}
	providerID := int64(9)

	_, err := svc.UpdateAccount(context.Background(), parentID, &UpdateAccountInput{
		ProviderID: providerIDUpdate(&providerID),
	})
	require.NoError(t, err)
	require.Equal(t, explicitShadowLoad, *repo.accounts[2].LoadFactor)
}

func TestAdminCreateAccountRejectsAPIKeyOAuthOnlyGroupBeforeWrite(t *testing.T) {
	repo := &providerOwnershipAccountRepoStub{accounts: map[int64]*Account{}, nextID: 10}
	groupRepo := &providerOwnershipGroupRepoStub{groups: map[int64]*Group{
		7: {ID: 7, Platform: PlatformAnthropic, Status: StatusActive, RequireOAuthOnly: true},
	}}
	svc := &adminServiceImpl{accountRepo: repo, groupRepo: groupRepo}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name: "apikey", Platform: PlatformAnthropic, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "secret"}, Concurrency: 1, Priority: 50,
		GroupIDs: []int64{7}, SkipMixedChannelCheck: true,
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "only accepts non-apikey accounts")
	require.Empty(t, repo.accounts, "OAuth-only validation must run before account persistence")
}
