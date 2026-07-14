//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetProviderUsageAggregatesSnapshotRows(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newUsageLogRepositoryWithSQL(nil, db)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"is_total", "account_id", "account_name", "platform", "total_requests",
		"total_input_tokens", "total_output_tokens", "total_cache_creation_tokens",
		"total_cache_read_tokens", "total_tokens",
	}).
		AddRow(1, nil, "", "", int64(3), int64(100), int64(50), int64(10), int64(20), int64(180)).
		AddRow(0, int64(42), "shared-claude", "claude", int64(3), int64(100), int64(50), int64(10), int64(20), int64(180))

	mock.ExpectQuery(`(?s)FROM usage_logs AS ul.*LEFT JOIN accounts AS a ON a.id = ul.account_id.*WHERE ul.provider_id = \$1.*ul.created_at >= \$2.*ul.created_at < \$3.*GROUP BY GROUPING SETS`).
		WithArgs(int64(7), start, end).
		WillReturnRows(rows)

	stats, err := repo.GetProviderUsage(context.Background(), 7, start, end)
	require.NoError(t, err)
	require.Equal(t, int64(7), stats.ProviderID)
	require.Equal(t, int64(180), stats.Totals.TotalTokens)
	require.Equal(t, int64(3), stats.Totals.TotalRequests)
	require.Len(t, stats.Accounts, 1)
	require.Equal(t, int64(42), stats.Accounts[0].AccountID)
	require.Equal(t, "shared-claude", stats.Accounts[0].AccountName)
	require.Equal(t, "claude", stats.Accounts[0].Platform)
	require.Equal(t, int64(180), stats.Accounts[0].TotalTokens)
	require.NoError(t, mock.ExpectationsWereMet())
}
