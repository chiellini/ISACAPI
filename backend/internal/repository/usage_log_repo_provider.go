package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// GetProviderUsage aggregates immutable provider snapshots stored on usage_logs.
// The accounts join is display-only: ownership attribution never consults the
// account's current provider_id.
func (r *usageLogRepository) GetProviderUsage(ctx context.Context, providerID int64, startTime, endTime time.Time) (stats *service.ProviderUsageStats, err error) {
	rows, err := r.sql.QueryContext(ctx, `
SELECT
    GROUPING(ul.account_id) AS is_total,
    ul.account_id,
    COALESCE(MAX(a.name), '') AS account_name,
    COALESCE(MAX(a.platform), '') AS platform,
    COUNT(*)::bigint AS total_requests,
    COALESCE(SUM(ul.input_tokens), 0)::bigint AS total_input_tokens,
    COALESCE(SUM(ul.output_tokens), 0)::bigint AS total_output_tokens,
    COALESCE(SUM(ul.cache_creation_tokens), 0)::bigint AS total_cache_creation_tokens,
    COALESCE(SUM(ul.cache_read_tokens), 0)::bigint AS total_cache_read_tokens,
    COALESCE(SUM(
        ul.input_tokens + ul.output_tokens +
        ul.cache_creation_tokens + ul.cache_read_tokens
    ), 0)::bigint AS total_tokens
FROM usage_logs AS ul
LEFT JOIN accounts AS a ON a.id = ul.account_id
WHERE ul.provider_id = $1
  AND ul.created_at >= $2
  AND ul.created_at < $3
GROUP BY GROUPING SETS ((), (ul.account_id))
ORDER BY is_total DESC, total_tokens DESC, ul.account_id ASC`, providerID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			stats = nil
		}
	}()

	stats = &service.ProviderUsageStats{
		ProviderID: providerID,
		StartTime:  startTime,
		EndTime:    endTime,
		Accounts:   make([]service.ProviderAccountUsageStats, 0),
	}
	for rows.Next() {
		var (
			isTotal     int
			accountID   sql.NullInt64
			accountName string
			platform    string
			usage       service.ProviderTokenUsageStats
		)
		if err = rows.Scan(
			&isTotal,
			&accountID,
			&accountName,
			&platform,
			&usage.TotalRequests,
			&usage.TotalInputTokens,
			&usage.TotalOutputTokens,
			&usage.TotalCacheCreationTokens,
			&usage.TotalCacheReadTokens,
			&usage.TotalTokens,
		); err != nil {
			return nil, err
		}
		if isTotal != 0 {
			stats.Totals = usage
			continue
		}
		if !accountID.Valid {
			continue
		}
		stats.Accounts = append(stats.Accounts, service.ProviderAccountUsageStats{
			AccountID:               accountID.Int64,
			AccountName:             accountName,
			Platform:                platform,
			ProviderTokenUsageStats: usage,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
