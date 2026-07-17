package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateAgentsAndWithdrawalsMigration(t *testing.T) {
	content, err := FS.ReadFile("181_affiliate_agents_and_withdrawals.sql")
	require.NoError(t, err)
	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "agent_status VARCHAR(16) NOT NULL DEFAULT 'inactive'")
	require.Contains(t, sql, "aff_withdrawal_pending DECIMAL(20,8) NOT NULL DEFAULT 0")
	require.Contains(t, sql, "aff_debt DECIMAL(20,8) NOT NULL DEFAULT 0")
	require.Contains(t, sql, "WHERE ua.aff_count > 0 OR ua.aff_history_quota > 0 OR EXISTS")
	require.NotContains(t, sql, "OR ua.inviter_id IS NOT NULL")
	require.Contains(t, sql, "idx_user_affiliate_withdrawals_user_idempotency")
	require.Contains(t, sql, "idx_user_affiliate_ledger_accrue_source_order")
	require.Contains(t, sql, "cancel_reason TEXT NULL")
	require.Contains(t, sql, "affiliate_minimum_withdrawal', '10'")
	require.Contains(t, sql, "affiliate_rebate_freeze_hours', '168'")
	require.Contains(t, sql, "INSERT INTO settings (key, value, updated_at)")
	require.NotContains(t, sql, "INSERT INTO settings (key, value, created_at")
}
