//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type providerUsageRepoStub struct {
	UsageLogRepository
	stats      *ProviderUsageStats
	err        error
	providerID int64
	startTime  time.Time
	endTime    time.Time
}

func (r *providerUsageRepoStub) GetProviderUsage(_ context.Context, providerID int64, startTime, endTime time.Time) (*ProviderUsageStats, error) {
	r.providerID = providerID
	r.startTime = startTime
	r.endTime = endTime
	return r.stats, r.err
}

func TestUsageServiceGetProviderUsageValidatesAndDelegates(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	repo := &providerUsageRepoStub{stats: &ProviderUsageStats{ProviderID: 9}}
	svc := &UsageService{usageRepo: repo}

	_, err := svc.GetProviderUsage(context.Background(), 0, start, end)
	require.ErrorIs(t, err, ErrProviderUsageInvalidProviderID)
	_, err = svc.GetProviderUsage(context.Background(), 9, end, start)
	require.ErrorIs(t, err, ErrProviderUsageInvalidTimeRange)
	_, err = svc.GetProviderUsage(context.Background(), 9, start, start)
	require.ErrorIs(t, err, ErrProviderUsageInvalidTimeRange)

	stats, err := svc.GetProviderUsage(context.Background(), 9, start, end)
	require.NoError(t, err)
	require.Same(t, repo.stats, stats)
	require.Equal(t, int64(9), repo.providerID)
	require.Equal(t, start, repo.startTime)
	require.Equal(t, end, repo.endTime)
}

func TestUsageServiceGetProviderUsageWrapsRepositoryError(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	wantErr := errors.New("query failed")
	svc := &UsageService{usageRepo: &providerUsageRepoStub{err: wantErr}}

	_, err := svc.GetProviderUsage(context.Background(), 9, start, start.Add(time.Hour))
	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "get provider usage")
}
