package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateAgentAnnouncementMigration(t *testing.T) {
	content, err := FS.ReadFile("185_seed_affiliate_agent_announcement.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "INSERT INTO announcements")
	require.Contains(t, sql, "'代理合作与佣金提现功能上线'")
	require.Contains(t, sql, "'active'")
	require.Contains(t, sql, "'popup'")
	require.Contains(t, sql, "'{\"any_of\":[]}'::jsonb")
	require.Contains(t, sql, "WHERE NOT EXISTS")
	require.Contains(t, sql, "[立即前往代理中心](/affiliate)")
	require.Contains(t, sql, "邀请返利（代理中心）")
	require.Contains(t, sql, "状态显示 **已经转账**")
}
