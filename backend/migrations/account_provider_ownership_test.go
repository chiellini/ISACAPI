package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountProviderOwnershipMigration(t *testing.T) {
	content, err := FS.ReadFile("175_account_provider_ownership.sql")
	require.NoError(t, err)

	sql := strings.ToUpper(string(content))
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS PROVIDER_ID BIGINT")
	require.Contains(t, sql, "FOREIGN KEY (PROVIDER_ID)")
	require.Contains(t, sql, "REFERENCES USERS(ID)")
	require.Contains(t, sql, "ON DELETE SET NULL")
	require.Contains(t, sql, "CREATE INDEX IF NOT EXISTS ACCOUNTS_PROVIDER_ID_IDX")
}
