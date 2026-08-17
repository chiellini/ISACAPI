package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalChatAPIKeyMigrationRetiresEntireReservedNamespace(t *testing.T) {
	content, err := FS.ReadFile("222_reserve_internal_chat_api_keys.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT FALSE")
	require.Contains(t, sql, "ADD CONSTRAINT api_keys_reserved_chat_name_requires_internal")
	require.Contains(t, sql, "OR is_internal")
	require.Contains(t, sql, ") NOT VALID")
	require.Contains(t, sql, "SET status = 'disabled'")
	require.Contains(t, sql, "deleted_at = NOW()")
	require.Contains(t, sql, "LEFT(name, 19) = '__chat_playground__'")
	require.NotContains(t, sql, "CREATE UNIQUE INDEX")
	require.NotContains(t, sql, "+goose Down")
	require.Less(t,
		strings.Index(sql, "ADD CONSTRAINT api_keys_reserved_chat_name_requires_internal"),
		strings.Index(sql, "SET status = 'disabled'"),
	)
	require.Less(t,
		strings.Index(sql, "SET status = 'disabled'"),
		strings.Index(sql, "VALIDATE CONSTRAINT api_keys_reserved_chat_name_requires_internal"),
	)
}

func TestInternalChatAPIKeyUniqueIndexIsNonBlockingAndCoversReservedNamespace(t *testing.T) {
	content, err := FS.ReadFile("223_internal_chat_api_key_unique_index_notx.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS api_keys_internal_chat_name_unique_active")
	require.Contains(t, sql, "WHERE deleted_at IS NULL")
	require.Contains(t, sql, "LEFT(name, 19) = '__chat_playground__'")
}
