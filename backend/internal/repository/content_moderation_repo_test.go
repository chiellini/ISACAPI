package repository

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildContentModerationLogWhere_BlockedIncludesAllBlockActions(t *testing.T) {
	where, args := buildContentModerationLogWhere(service.ContentModerationLogFilter{Result: "blocked"})

	require.Empty(t, args)
	sql := strings.Join(where, " AND ")
	require.Contains(t, sql, "l.action IN ('block', 'keyword_block', 'hash_block', 'upstream_policy_block', 'prompt_guard_block', 'session_policy_block', 'cyber_policy')")
	require.NotContains(t, sql, "l.action = 'block'")
}

func TestBuildContentModerationLogWhere_SearchIncludesSecondaryReviewFailureFields(t *testing.T) {
	where, args := buildContentModerationLogWhere(service.ContentModerationLogFilter{Search: "secondary_review"})

	require.Len(t, args, 8)
	for _, arg := range args {
		require.Equal(t, "%secondary_review%", arg)
	}
	sql := strings.Join(where, " AND ")
	require.Contains(t, sql, "l.matched_keyword ILIKE")
	require.Contains(t, sql, "l.highest_category ILIKE")
	require.Contains(t, sql, "l.error ILIKE")
}

func TestContentModerationRepositoryCountFlaggedByUserSince_ExcludesHashBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewContentModerationRepository(db)
	since := time.Now().Add(-time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("AND action <> 'hash_block'")).
		WithArgs(int64(1001), since, false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := repo.CountFlaggedByUserSince(context.Background(), 1001, since, false)

	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentModerationRepositoryCountFlaggedByUserSince_ExcludesAuditOnlyPolicyActions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewContentModerationRepository(db)
	since := time.Now().Add(-time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("AND action <> 'upstream_policy_block'\n  AND action <> 'prompt_guard_block'\n  AND action <> 'session_policy_block'")).
		WithArgs(int64(1001), since, false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := repo.CountFlaggedByUserSince(context.Background(), 1001, since, false)

	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentModerationRepositoryCountFlaggedByUserSince_ExcludesCyberPolicyWhenRequested(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewContentModerationRepository(db)
	since := time.Now().Add(-time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("AND ($3::bool IS FALSE OR action <> 'cyber_policy')")).
		WithArgs(int64(1001), since, true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountFlaggedByUserSince(context.Background(), 1001, since, true)

	require.NoError(t, err)
	require.Equal(t, 3, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentModerationRepositoryCleanupExpiredLogs_RetainsErrorsAsAuditRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewContentModerationRepository(db)
	hitBefore := time.Now().Add(-180 * 24 * time.Hour)
	nonHitBefore := time.Now().Add(-3 * 24 * time.Hour)
	mock.ExpectExec(regexp.QuoteMeta("WHERE (flagged = TRUE OR action = 'error') AND created_at < $1")).
		WithArgs(hitBefore).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("WHERE flagged = FALSE AND action <> 'error' AND created_at < $1")).
		WithArgs(nonHitBefore).
		WillReturnResult(sqlmock.NewResult(0, 5))

	result, err := repo.CleanupExpiredLogs(context.Background(), hitBefore, nonHitBefore)

	require.NoError(t, err)
	require.Equal(t, int64(2), result.DeletedHit)
	require.Equal(t, int64(5), result.DeletedNonHit)
	require.NoError(t, mock.ExpectationsWereMet())
}
