//go:build unit

package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

type balanceEligibilityCacheStub struct {
	billingCacheWorkerStub

	balance                  float64
	cacheMissAfterInvalidate bool
	invalidated              atomic.Bool
	deductCalls              atomic.Int64
	invalidateCalls          atomic.Int64
}

type researchGroupBillingCacheStub struct {
	billingCacheWorkerStub
	balances        map[int64]float64
	fundingReadErr  error
	cachedFunding   *ResearchGroupFundingContext
	fundingFound    bool
	setFunding      *ResearchGroupFundingContext
	setFundingCalls atomic.Int64
}

func (s *researchGroupBillingCacheStub) GetUserBalance(_ context.Context, userID int64) (float64, error) {
	return s.balances[userID], nil
}

func (s *researchGroupBillingCacheStub) GetResearchGroupFunding(context.Context, int64) (*ResearchGroupFundingContext, bool, error) {
	return s.cachedFunding, s.fundingFound, s.fundingReadErr
}

func (s *researchGroupBillingCacheStub) SetResearchGroupFunding(_ context.Context, _ int64, funding *ResearchGroupFundingContext) error {
	s.setFundingCalls.Add(1)
	s.setFunding = funding
	return nil
}

func (s *researchGroupBillingCacheStub) InvalidateResearchGroupFunding(context.Context, int64) error {
	return nil
}

type researchGroupFundingRepoStub struct {
	ResearchGroupRepository
	funding *ResearchGroupFundingContext
	calls   atomic.Int64
}

func (s *researchGroupFundingRepoStub) GetFundingContextByUserID(context.Context, int64) (*ResearchGroupFundingContext, error) {
	s.calls.Add(1)
	return s.funding, nil
}

func (s *balanceEligibilityCacheStub) GetUserBalance(context.Context, int64) (float64, error) {
	if s.cacheMissAfterInvalidate && s.invalidated.Load() {
		return 0, errors.New("cache miss")
	}
	return s.balance, nil
}

func (s *balanceEligibilityCacheStub) DeductUserBalance(context.Context, int64, float64) error {
	s.deductCalls.Add(1)
	return nil
}

func (s *balanceEligibilityCacheStub) InvalidateUserBalance(context.Context, int64) error {
	s.invalidateCalls.Add(1)
	s.invalidated.Store(true)
	return nil
}

func TestCheckBillingEligibility_RejectsBalanceBelowMinimumReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.005}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.01
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	_, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestCheckBillingEligibility_AllowsBalanceAtMinimumReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.01}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.01
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	_, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.NoError(t, err)
}

func TestCheckBillingEligibility_ResearchGroupRedisFailureFallsBackToDB(t *testing.T) {
	funding := &ResearchGroupFundingContext{
		ResearchGroupID:       81,
		ResearchGroupMemberID: 91,
		MemberUserID:          1,
		OwnerUserID:           2,
		MonthlyLimitUSD:       100,
		UsageWindowStart:      timezone.StartOfMonth(timezone.Now()),
	}
	cache := &researchGroupBillingCacheStub{
		balances:       map[int64]float64{1: 10, 2: 20},
		fundingReadErr: errors.New("redis unavailable"),
	}
	repo := &researchGroupFundingRepoStub{funding: funding}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil, repo)
	t.Cleanup(svc.Stop)

	decision, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), decision.CallerUserID)
	require.Equal(t, int64(2), decision.PayerUserID)
	require.Equal(t, int64(81), decision.ResearchGroupID)
	require.Equal(t, FundingSourceResearchGroup, decision.FundingSource)
	require.Equal(t, int64(1), repo.calls.Load())
	require.Same(t, funding, cache.setFunding)
}

func TestCheckBillingEligibility_QuotaAndPersonalBalanceInsufficientUsesCombinedError(t *testing.T) {
	funding := &ResearchGroupFundingContext{
		ResearchGroupID:       81,
		ResearchGroupMemberID: 91,
		MemberUserID:          1,
		OwnerUserID:           2,
		MonthlyLimitUSD:       5,
		MonthlyUsageUSD:       5,
		UsageWindowStart:      timezone.StartOfMonth(timezone.Now()),
	}
	cache := &researchGroupBillingCacheStub{balances: map[int64]float64{1: 0, 2: 20}}
	repo := &researchGroupFundingRepoStub{funding: funding}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil, repo)
	t.Cleanup(svc.Stop)

	_, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.ErrorIs(t, err, ErrResearchGroupAndPersonalBalanceInsufficient)
}

func TestCheckBillingEligibility_StalePositiveFundingCacheIsRevalidated(t *testing.T) {
	staleFunding := &ResearchGroupFundingContext{
		ResearchGroupID:       81,
		ResearchGroupMemberID: 91,
		MemberUserID:          1,
		OwnerUserID:           2,
		MonthlyLimitUSD:       100,
		UsageWindowStart:      timezone.StartOfMonth(timezone.Now()),
	}
	cache := &researchGroupBillingCacheStub{
		balances:      map[int64]float64{1: 10, 2: 20},
		cachedFunding: staleFunding,
		fundingFound:  true,
	}
	repo := &researchGroupFundingRepoStub{funding: nil}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil, repo)
	t.Cleanup(svc.Stop)

	decision, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), decision.PayerUserID)
	require.Equal(t, FundingSourceSelf, decision.FundingSource)
	require.Equal(t, int64(1), repo.calls.Load())
	require.Nil(t, cache.setFunding)
}

func TestCheckBillingEligibility_NegativeFundingCacheIsRevalidatedAndRefreshed(t *testing.T) {
	funding := &ResearchGroupFundingContext{
		ResearchGroupID:       81,
		ResearchGroupMemberID: 91,
		MemberUserID:          1,
		OwnerUserID:           2,
		MonthlyLimitUSD:       100,
		UsageWindowStart:      timezone.StartOfMonth(timezone.Now()),
	}
	cache := &researchGroupBillingCacheStub{
		balances:      map[int64]float64{1: 10, 2: 20},
		cachedFunding: nil,
		fundingFound:  true,
	}
	repo := &researchGroupFundingRepoStub{funding: funding}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil, repo)
	t.Cleanup(svc.Stop)

	decision, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), decision.PayerUserID)
	require.Equal(t, int64(81), decision.ResearchGroupID)
	require.Equal(t, int64(91), decision.ResearchGroupMemberID)
	require.Equal(t, FundingSourceResearchGroup, decision.FundingSource)
	require.Equal(t, int64(1), repo.calls.Load())
	require.Equal(t, int64(1), cache.setFundingCalls.Load())
	require.Same(t, funding, cache.setFunding)
}

func TestSyncBalanceCacheAfterDeduction_InvalidatesExhaustedBalance(t *testing.T) {
	cache := &balanceEligibilityCacheStub{
		balance:                  0.50,
		cacheMissAfterInvalidate: true,
	}
	userRepo := &balanceLoadUserRepoStub{balance: -0.25}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.01
	svc := NewBillingCacheService(cache, userRepo, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	newBalance := -0.25
	syncBalanceCacheAfterDeduction(context.Background(), &postUsageBillingParams{
		Cost: &CostBreakdown{ActualCost: 0.75},
		User: &User{ID: 1},
	}, &billingDeps{billingCacheService: svc}, &UsageBillingApplyResult{
		NewBalance:         &newBalance,
		BalanceOverdrafted: true,
	})

	require.Equal(t, int64(1), cache.invalidateCalls.Load())
	require.Equal(t, int64(0), cache.deductCalls.Load())

	_, err := svc.CheckBillingEligibility(context.Background(), &User{ID: 1}, nil, nil, nil, "")
	require.ErrorIs(t, err, ErrInsufficientBalance)
	require.Equal(t, int64(1), userRepo.calls.Load())
}

func TestSyncBalanceCacheAfterDeduction_InvalidatesWhenBalanceFallsBelowReserve(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 0.50}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.01
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	newBalance := 0.005
	syncBalanceCacheAfterDeduction(context.Background(), &postUsageBillingParams{
		Cost: &CostBreakdown{ActualCost: 0.495},
		User: &User{ID: 1},
	}, &billingDeps{billingCacheService: svc}, &UsageBillingApplyResult{NewBalance: &newBalance})

	require.Equal(t, int64(1), cache.invalidateCalls.Load())
	require.Equal(t, int64(0), cache.deductCalls.Load())
}

func TestSyncBalanceCacheAfterDeduction_QueuesDeductWhenBalanceStillEligible(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 1}
	cfg := &config.Config{}
	cfg.Billing.MinimumBalanceReserve = 0.01
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	newBalance := 0.75
	syncBalanceCacheAfterDeduction(context.Background(), &postUsageBillingParams{
		Cost: &CostBreakdown{ActualCost: 0.25},
		User: &User{ID: 1},
	}, &billingDeps{billingCacheService: svc}, &UsageBillingApplyResult{NewBalance: &newBalance})

	require.Equal(t, int64(0), cache.invalidateCalls.Load())
	require.Eventually(t, func() bool {
		return cache.deductCalls.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)
}
