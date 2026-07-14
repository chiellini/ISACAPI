//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositoryProviderOwnership(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	repo := newAccountRepositoryWithSQL(tx.Client(), tx, nil)

	provider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-owner@example.com"})
	otherProvider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-other@example.com"})

	createAccount := func(name string, providerID *int64) *service.Account {
		t.Helper()
		account := &service.Account{
			Name:               name,
			Platform:           service.PlatformAnthropic,
			Type:               service.AccountTypeOAuth,
			Credentials:        map[string]any{},
			Extra:              map[string]any{},
			ProviderID:         providerID,
			Concurrency:        3,
			Priority:           50,
			Status:             service.StatusActive,
			Schedulable:        true,
			AutoPauseOnExpired: true,
		}
		require.NoError(t, repo.Create(ctx, account))
		return account
	}

	owned := createAccount("provider-owned", &provider.ID)
	createAccount("other-provider-owned", &otherProvider.ID)
	legacy := createAccount("platform-owned", nil)

	got, err := repo.GetByIDForProvider(ctx, owned.ID, provider.ID)
	require.NoError(t, err)
	require.Equal(t, &provider.ID, got.ProviderID)

	_, err = repo.GetByIDForProvider(ctx, owned.ID, otherProvider.ID)
	require.ErrorIs(t, err, service.ErrAccountNotFound)

	accounts, page, err := repo.ListWithFiltersForProvider(
		ctx,
		pagination.PaginationParams{Page: 1, PageSize: 10},
		"", "", "", "", 0, "",
		provider.ID,
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, accounts, 1)
	require.Equal(t, owned.ID, accounts[0].ID)

	legacyGot, err := repo.GetByID(ctx, legacy.ID)
	require.NoError(t, err)
	require.Nil(t, legacyGot.ProviderID)

	owned.ProviderID = &otherProvider.ID
	require.NoError(t, repo.Update(ctx, owned))
	got, err = repo.GetByIDForProvider(ctx, owned.ID, otherProvider.ID)
	require.NoError(t, err)
	require.Equal(t, &otherProvider.ID, got.ProviderID)

	owned.ProviderID = nil
	require.NoError(t, repo.Update(ctx, owned))
	got, err = repo.GetByID(ctx, owned.ID)
	require.NoError(t, err)
	require.Nil(t, got.ProviderID)
}

func TestAccountProviderForeignKeyOnDeleteSetsNull(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	repo := newAccountRepositoryWithSQL(tx.Client(), tx, nil)
	provider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-delete@example.com"})

	account := &service.Account{
		Name:               "survives-provider-delete",
		Platform:           service.PlatformAnthropic,
		Type:               service.AccountTypeOAuth,
		Credentials:        map[string]any{},
		Extra:              map[string]any{},
		ProviderID:         &provider.ID,
		Concurrency:        3,
		Priority:           50,
		Status:             service.StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
	}
	require.NoError(t, repo.Create(ctx, account))

	_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", provider.ID)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Nil(t, got.ProviderID)
}

func TestAccountRepositoryDeleteForProviderIsAtomicAndPreservesCascadeCleanup(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	cache := &schedulerCacheRecorder{}
	repo := newAccountRepositoryWithSQL(tx.Client(), tx, cache)
	provider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-atomic-delete@example.com"})
	otherProvider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-atomic-delete-other@example.com"})
	group := mustCreateGroup(t, tx.Client(), &service.Group{Name: "provider-delete-group"})

	createAccount := func(name string, parentID *int64, quotaDimension string) *service.Account {
		t.Helper()
		account := &service.Account{
			Name:               name,
			Platform:           service.PlatformOpenAI,
			Type:               service.AccountTypeOAuth,
			Credentials:        map[string]any{},
			Extra:              map[string]any{},
			ProviderID:         &provider.ID,
			Concurrency:        3,
			Priority:           50,
			Status:             service.StatusActive,
			Schedulable:        true,
			AutoPauseOnExpired: true,
			ParentAccountID:    parentID,
			QuotaDimension:     quotaDimension,
		}
		require.NoError(t, repo.Create(ctx, account))
		return account
	}

	parent := createAccount("provider-delete-parent", nil, service.QuotaDimensionGlobal)
	shadow := createAccount("provider-delete-shadow", &parent.ID, service.QuotaDimensionSpark)
	mustBindAccountToGroup(t, tx.Client(), parent.ID, group.ID, 1)
	mustBindAccountToGroup(t, tx.Client(), shadow.ID, group.ID, 1)
	_, err := tx.ExecContext(ctx, "INSERT INTO scheduled_test_plans (account_id) VALUES ($1), ($2)", parent.ID, shadow.ID)
	require.NoError(t, err)

	err = repo.DeleteForProvider(ctx, parent.ID, otherProvider.ID)
	require.ErrorIs(t, err, service.ErrAccountNotFound)
	_, err = repo.GetByID(ctx, parent.ID)
	require.NoError(t, err)
	_, err = repo.GetByID(ctx, shadow.ID)
	require.NoError(t, err)
	require.Empty(t, cache.deleteIDs)

	require.NoError(t, repo.DeleteForProvider(ctx, parent.ID, provider.ID))
	_, err = repo.GetByID(ctx, parent.ID)
	require.ErrorIs(t, err, service.ErrAccountNotFound)
	_, err = repo.GetByID(ctx, shadow.ID)
	require.ErrorIs(t, err, service.ErrAccountNotFound)
	require.Equal(t, []int64{shadow.ID, parent.ID}, cache.deleteIDs)

	var bindingCount int
	require.NoError(t, tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM account_groups WHERE account_id IN ($1, $2)", parent.ID, shadow.ID,
	).Scan(&bindingCount))
	require.Zero(t, bindingCount)
	var planCount int
	require.NoError(t, tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM scheduled_test_plans WHERE account_id IN ($1, $2)", parent.ID, shadow.ID,
	).Scan(&planCount))
	require.Zero(t, planCount)
}

func TestAccountRepositoryUpdateForProviderIsOwnedAllowlistedAndAtomic(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	repo := newAccountRepositoryWithSQL(tx.Client(), tx, nil)
	provider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-update-owner@example.com"})
	otherProvider := mustCreateUser(t, tx.Client(), &service.User{Email: "provider-update-other@example.com"})
	proxy := mustCreateProxy(t, tx.Client(), &service.Proxy{Name: "provider-update-proxy"})
	oldGroup := mustCreateGroup(t, tx.Client(), &service.Group{
		Name: "provider-update-old", Platform: service.PlatformAnthropic,
	})
	newGroup := mustCreateGroup(t, tx.Client(), &service.Group{
		Name: "provider-update-new", Platform: service.PlatformAnthropic,
	})
	wrongPlatformGroup := mustCreateGroup(t, tx.Client(), &service.Group{
		Name: "provider-update-wrong-platform", Platform: service.PlatformOpenAI,
	})
	oauthOnlyGroup := mustCreateGroup(t, tx.Client(), &service.Group{
		Name: "provider-update-oauth-only", Platform: service.PlatformAnthropic,
	})
	_, err := tx.Client().Group.UpdateOneID(oauthOnlyGroup.ID).SetRequireOauthOnly(true).Save(ctx)
	require.NoError(t, err)

	notes := "admin note"
	loadFactor := 9
	rateMultiplier := 0.75
	account := &service.Account{
		Name: "owned", Notes: &notes, Platform: service.PlatformAnthropic,
		Type: service.AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "latest-access", "refresh_token": "latest-refresh", "label": "old",
		},
		Extra: map[string]any{"admin_sentinel": "keep"}, ProxyID: &proxy.ID,
		ProviderID: &provider.ID, Concurrency: 3, Priority: 7,
		RateMultiplier: &rateMultiplier, LoadFactor: &loadFactor,
		Status: service.StatusActive, Schedulable: false, AutoPauseOnExpired: true,
	}
	require.NoError(t, repo.Create(ctx, account))
	mustBindAccountToGroup(t, tx.Client(), account.ID, oldGroup.ID, 1)

	forbiddenName := "wrong-owner-write"
	_, err = repo.UpdateForProvider(ctx, account.ID, otherProvider.ID, &service.ProviderAccountUpdateInput{
		Name: &forbiddenName,
	})
	require.ErrorIs(t, err, service.ErrAccountNotFound)
	afterWrongOwner, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "owned", afterWrongOwner.Name)
	require.Equal(t, []int64{oldGroup.ID}, afterWrongOwner.GroupIDs)

	invalidName := "must-roll-back"
	invalidGroups := []int64{wrongPlatformGroup.ID}
	_, err = repo.UpdateForProvider(ctx, account.ID, provider.ID, &service.ProviderAccountUpdateInput{
		Name: &invalidName, GroupIDs: &invalidGroups,
	})
	require.Error(t, err)
	afterInvalidGroup, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "owned", afterInvalidGroup.Name)
	require.Equal(t, []int64{oldGroup.ID}, afterInvalidGroup.GroupIDs)

	updatedName := "provider-edit"
	updatedConcurrency := 4
	updatedStatus := "inactive"
	updatedGroups := []int64{newGroup.ID}
	updated, err := repo.UpdateForProvider(ctx, account.ID, provider.ID, &service.ProviderAccountUpdateInput{
		Name: &updatedName, NotesSet: true, Notes: nil,
		Credentials: map[string]any{"label": "new"},
		Concurrency: &updatedConcurrency, Status: &updatedStatus, GroupIDs: &updatedGroups,
	})
	require.NoError(t, err)
	require.Equal(t, updatedName, updated.Name)
	require.Nil(t, updated.Notes)
	require.Equal(t, updatedConcurrency, updated.Concurrency)
	require.Equal(t, updatedStatus, updated.Status)
	require.Equal(t, []int64{newGroup.ID}, updated.GroupIDs)
	require.Equal(t, "latest-access", updated.Credentials["access_token"])
	require.Equal(t, "latest-refresh", updated.Credentials["refresh_token"])
	require.Equal(t, "new", updated.Credentials["label"])
	require.Equal(t, service.PlatformAnthropic, updated.Platform)
	require.Equal(t, service.AccountTypeOAuth, updated.Type)
	require.Equal(t, &provider.ID, updated.ProviderID)
	require.Equal(t, &proxy.ID, updated.ProxyID)
	require.Equal(t, 7, updated.Priority)
	require.Equal(t, loadFactor, *updated.LoadFactor)
	require.Equal(t, rateMultiplier, *updated.RateMultiplier)
	require.False(t, updated.Schedulable)
	require.Equal(t, map[string]any{"admin_sentinel": "keep"}, updated.Extra)

	apiKey := &service.Account{
		Name: "apikey-owned", Platform: service.PlatformAnthropic, Type: service.AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "latest-key"}, Extra: map[string]any{},
		ProviderID: &provider.ID, Concurrency: 1, Priority: 50,
		Status: service.StatusActive, Schedulable: true, AutoPauseOnExpired: true,
	}
	require.NoError(t, repo.Create(ctx, apiKey))
	mustBindAccountToGroup(t, tx.Client(), apiKey.ID, oldGroup.ID, 1)
	apiKeyInvalidName := "apikey-must-roll-back"
	oauthOnlyGroups := []int64{oauthOnlyGroup.ID}
	_, err = repo.UpdateForProvider(ctx, apiKey.ID, provider.ID, &service.ProviderAccountUpdateInput{
		Name: &apiKeyInvalidName, GroupIDs: &oauthOnlyGroups,
	})
	require.Error(t, err)
	afterOAuthOnly, err := repo.GetByID(ctx, apiKey.ID)
	require.NoError(t, err)
	require.Equal(t, "apikey-owned", afterOAuthOnly.Name)
	require.Equal(t, []int64{oldGroup.ID}, afterOAuthOnly.GroupIDs)

	shadow := &service.Account{
		Name: "provider-shadow", Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
		Credentials: map[string]any{"model_mapping": map[string]any{"spark": "gpt"}}, Extra: map[string]any{},
		ProviderID: &provider.ID, ParentAccountID: &account.ID, QuotaDimension: service.QuotaDimensionSpark,
		Concurrency: 2, Priority: 50, Status: service.StatusActive, Schedulable: true,
		AutoPauseOnExpired: true,
	}
	require.NoError(t, repo.Create(ctx, shadow))
	_, err = repo.UpdateForProvider(ctx, shadow.ID, provider.ID, &service.ProviderAccountUpdateInput{
		Credentials: map[string]any{"access_token": "must-not-persist"},
	})
	require.Error(t, err)
	afterShadowCredentialAttempt, err := repo.GetByID(ctx, shadow.ID)
	require.NoError(t, err)
	require.NotContains(t, afterShadowCredentialAttempt.Credentials, "access_token")
	require.Contains(t, afterShadowCredentialAttempt.Credentials, "model_mapping")
}
