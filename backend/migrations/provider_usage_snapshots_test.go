package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderUsageSnapshotMigrations(t *testing.T) {
	snapshotContent, err := FS.ReadFile("176_provider_usage_snapshots.sql")
	require.NoError(t, err)
	snapshotSQL := strings.ToUpper(string(snapshotContent))
	require.Contains(t, snapshotSQL, "ALTER TABLE USAGE_LOGS")
	require.Contains(t, snapshotSQL, "ADD COLUMN IF NOT EXISTS PROVIDER_ID BIGINT NULL")
	require.Contains(t, snapshotSQL, "ALTER TABLE BATCH_IMAGE_JOBS")
	require.Contains(t, snapshotSQL, "ADD COLUMN IF NOT EXISTS ACCOUNT_PROVIDER_ID BIGINT NULL")
	require.NotContains(t, snapshotSQL, "UPDATE USAGE_LOGS")
	require.NotContains(t, snapshotSQL, "REFERENCES USERS")

	indexContent, err := FS.ReadFile("176a_usage_logs_provider_created_at_index_notx.sql")
	require.NoError(t, err)
	indexSQL := strings.Join(strings.Fields(string(indexContent)), " ")
	require.Contains(t, indexSQL, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_provider_created_at")
	require.Contains(t, indexSQL, "ON usage_logs (provider_id, created_at)")
	require.Contains(t, indexSQL, "WHERE provider_id IS NOT NULL")
}
