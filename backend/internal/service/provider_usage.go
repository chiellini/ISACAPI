package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrProviderUsageInvalidProviderID = infraerrors.BadRequest(
		"PROVIDER_USAGE_INVALID_PROVIDER_ID",
		"provider ID must be positive",
	)
	ErrProviderUsageInvalidTimeRange = infraerrors.BadRequest(
		"PROVIDER_USAGE_INVALID_TIME_RANGE",
		"usage start time must be before end time",
	)
	ErrProviderUsageUnsupported = errors.New("provider usage repository is not supported")
)

// ProviderTokenUsageStats contains the token counts recorded by usage_logs.
// TotalTokens includes input, output, cache creation, and cache read tokens.
type ProviderTokenUsageStats struct {
	TotalRequests            int64 `json:"total_requests"`
	TotalInputTokens         int64 `json:"total_input_tokens"`
	TotalOutputTokens        int64 `json:"total_output_tokens"`
	TotalCacheCreationTokens int64 `json:"total_cache_creation_tokens"`
	TotalCacheReadTokens     int64 `json:"total_cache_read_tokens"`
	TotalTokens              int64 `json:"total_tokens"`
}

// ProviderAccountUsageStats is a provider usage breakdown for one account.
type ProviderAccountUsageStats struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Platform    string `json:"platform"`
	ProviderTokenUsageStats
}

// ProviderUsageStats is the usage attributed to a provider when each request
// was recorded. The time range follows the project-wide [start, end) convention.
type ProviderUsageStats struct {
	ProviderID int64                       `json:"provider_id"`
	StartTime  time.Time                   `json:"start_time"`
	EndTime    time.Time                   `json:"end_time"`
	Totals     ProviderTokenUsageStats     `json:"totals"`
	Accounts   []ProviderAccountUsageStats `json:"accounts"`
}

// ProviderUsageRepository is the optional usage-log repository capability used
// by provider and admin usage views. Authorization is enforced by the caller.
type ProviderUsageRepository interface {
	GetProviderUsage(ctx context.Context, providerID int64, startTime, endTime time.Time) (*ProviderUsageStats, error)
}

// GetProviderUsage returns consumption attributed by usage_logs.provider_id.
// Authorization (provider self-query vs. admin query) is left to the caller.
func (s *UsageService) GetProviderUsage(ctx context.Context, providerID int64, startTime, endTime time.Time) (*ProviderUsageStats, error) {
	if providerID <= 0 {
		return nil, ErrProviderUsageInvalidProviderID
	}
	if !startTime.Before(endTime) {
		return nil, ErrProviderUsageInvalidTimeRange
	}

	repo, ok := s.usageRepo.(ProviderUsageRepository)
	if !ok {
		return nil, ErrProviderUsageUnsupported
	}

	stats, err := repo.GetProviderUsage(ctx, providerID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("get provider usage: %w", err)
	}
	return stats, nil
}
