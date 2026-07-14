package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResearchGroupMigrationsContainRequiredConstraintsAndOnlineIndexes(t *testing.T) {
	foundation, err := FS.ReadFile("177_research_groups.sql")
	require.NoError(t, err)
	foundationSQL := string(foundation)
	require.Contains(t, foundationSQL, "CREATE TABLE IF NOT EXISTS research_groups")
	require.Contains(t, foundationSQL, "CREATE TABLE IF NOT EXISTS research_group_members")
	require.Contains(t, foundationSQL, "CREATE TABLE IF NOT EXISTS research_group_quota_audits")
	require.Contains(t, foundationSQL, "research_group_members_user_effective_uidx")
	require.Contains(t, foundationSQL, "status IN ('pending', 'active', 'paused')")
	require.Contains(t, foundationSQL, "owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL")
	require.Contains(t, foundationSQL, "research_groups_active_owner_check")
	require.Contains(t, foundationSQL, "research_group_quota_audits_append_only")
	require.NotContains(t, foundationSQL, "member_id BIGINT REFERENCES")
	require.NotContains(t, foundationSQL, "actor_user_id BIGINT REFERENCES")

	attribution, err := FS.ReadFile("178_research_group_billing_attribution.sql")
	require.NoError(t, err)
	attributionSQL := string(attribution)
	for _, column := range []string{"payer_user_id", "research_group_id", "research_group_member_id", "funding_source"} {
		require.Contains(t, attributionSQL, "ADD COLUMN IF NOT EXISTS "+column)
	}
	require.Contains(t, attributionSQL, "usage_logs_funding_source_check")
	require.Contains(t, attributionSQL, "research_group')) NOT VALID")
	require.NotContains(t, attributionSQL, "fk_usage_logs_payer_user_id")
	require.NotContains(t, attributionSQL, "fk_usage_logs_research_group_member_id")
	require.NotContains(t, attributionSQL, "CREATE INDEX IF NOT EXISTS usage_logs_")

	indexes, err := FS.ReadFile("179_research_group_usage_indexes_notx.sql")
	require.NoError(t, err)
	indexesSQL := string(indexes)
	require.Equal(t, 3, strings.Count(indexesSQL, "CREATE INDEX CONCURRENTLY IF NOT EXISTS"))
}
