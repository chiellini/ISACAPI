package repository

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	migrationfs "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestMigrationSQLForExecution(t *testing.T) {
	t.Run("unmarked migration is unchanged", func(t *testing.T) {
		const content = "-- ordinary comment\nSELECT '-- +goose Down';"
		require.Equal(t, content, migrationSQLForExecution(content))
	})

	t.Run("only goose up section is returned", func(t *testing.T) {
		const content = "header\r\n  -- +goose Up  \r\nSELECT 1;\r\n-- +goose Down\r\nSELECT 2;"
		require.Equal(t, "SELECT 1;", migrationSQLForExecution(content))
	})
}

func TestMigrationSQLForExecution_LegacyGooseFiles(t *testing.T) {
	tests := []struct {
		name    string
		upSQL   string
		downSQL string
	}{
		{
			name:    "019_migrate_wechat_to_attributes.sql",
			upSQL:   "ALTER TABLE users DROP COLUMN IF EXISTS wechat",
			downSQL: "ALTER TABLE users ADD COLUMN IF NOT EXISTS wechat",
		},
		{
			name:    "024_add_gemini_tier_id.sql",
			upSQL:   "jsonb_set(",
			downSQL: "credentials = credentials - 'tier_id'",
		},
		{
			name:    "037_ops_alert_silences.sql",
			upSQL:   "CREATE TABLE IF NOT EXISTS ops_alert_silences",
			downSQL: "DROP TABLE IF EXISTS ops_alert_silences",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := migrationfs.FS.ReadFile(tt.name)
			require.NoError(t, err)

			executionSQL := migrationSQLForExecution(strings.TrimSpace(string(content)))
			require.Contains(t, executionSQL, tt.upSQL)
			require.NotContains(t, executionSQL, tt.downSQL)
			require.NotContains(t, strings.ToLower(executionSQL), "-- +goose down")
		})
	}
}

func TestApplyMigrationsFS_GooseExecutesUpAndChecksumsFullFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	const (
		name    = "001_goose.sql"
		content = `-- +goose Up
CREATE TABLE only_up(id int);
-- +goose Down
DROP TABLE only_up;`
		upSQL = "CREATE TABLE only_up(id int);"
	)

	prepareMigrationsBootstrapExpectations(mock)
	mock.ExpectQuery("SELECT checksum FROM schema_migrations WHERE filename = \\$1").
		WithArgs(name).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(upSQL)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO schema_migrations \\(filename, checksum\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(name, migrationChecksum(content)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\)").
		WithArgs(migrationsAdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	fSys := fstest.MapFS{name: {Data: []byte(content)}}
	require.NoError(t, applyMigrationsFS(context.Background(), db, fSys))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLegacyGooseRepairMigrationIsForwardOnly(t *testing.T) {
	content, err := migrationfs.FS.ReadFile("221_repair_legacy_goose_migrations.sql")
	require.NoError(t, err)
	sql := string(content)

	require.NotContains(t, strings.ToLower(sql), "-- +goose down")
	require.Contains(t, sql, "user_attribute_definitions")
	require.Contains(t, sql, "credentials->>'tier_id'")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS ops_alert_silences")
}
