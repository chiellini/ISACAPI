package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestResearchGroupRepositoryGetFundingContextUsesApplicationMonthBoundary(t *testing.T) {
	require.NoError(t, timezone.Init("Asia/Taipei"))
	t.Cleanup(func() { _ = timezone.Init("UTC") })
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &researchGroupRepository{db: db}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE research_group_members")).
		WithArgs(int64(7), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	window := time.Date(2026, time.July, 1, 0, 0, 0, 0, timezone.Location())
	mock.ExpectQuery(`(?s)SELECT rg\.id, m\.id, m\.user_id.*owner\.status = 'active' AND owner\.deleted_at IS NULL`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"group_id", "member_id", "member_user_id", "owner_user_id", "group_status", "member_status",
			"monthly_limit", "monthly_usage", "monthly_reserved", "window_start", "owner_balance",
		}).AddRow(3, 4, 7, 9, "active", "active", 100.0, 20.0, 5.0, window, 300.0))

	funding, err := repo.GetFundingContextByUserID(context.Background(), 7)

	require.NoError(t, err)
	require.Equal(t, int64(9), funding.PayerUserID)
	require.Equal(t, int64(7), funding.CallerUserID)
	require.Equal(t, float64(75), funding.RemainingAt(timezone.Now()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResearchGroupRepositoryUsageSummaryIsCurrentMonthOnly(t *testing.T) {
	require.NoError(t, timezone.Init("Asia/Taipei"))
	t.Cleanup(func() { _ = timezone.Init("UTC") })
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &researchGroupRepository{db: db}

	mock.ExpectQuery(`(?s)SUM\(ul\.actual_cost\).*FROM usage_logs ul.*ul\.created_at >= \$3`).
		WithArgs(int64(3), int64(9), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "request_count", "active_members"}).AddRow(12.5, 4, 2))

	summary, err := repo.GetUsageSummary(context.Background(), 3, 9)

	require.NoError(t, err)
	require.Equal(t, &service.ResearchGroupUsageSummary{TotalCostUSD: 12.5, RequestCount: 4, ActiveMemberCount: 2}, summary)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResearchGroupRepositoryCreateRejectsEffectiveMemberInsideUserLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &researchGroupRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT role, status FROM users.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"role", "status"}).AddRow(service.RoleUser, service.StatusActive))
	mock.ExpectQuery(`(?s)SELECT EXISTS.*research_group_members`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	_, err = repo.Create(context.Background(), 7, "Lab")

	require.ErrorIs(t, err, service.ErrResearchGroupAlreadyExists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResearchGroupRepositoryInviteLocksGroupAgainstConcurrentDissolve(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &researchGroupRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)SELECT role, status FROM users.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"role", "status"}).AddRow(service.RoleUser, service.StatusActive))
	mock.ExpectQuery(`(?s)SELECT EXISTS.*research_groups.*research_group_members`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`(?s)SELECT id FROM research_groups.*FOR UPDATE`).
		WithArgs(int64(3)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = repo.InviteMember(context.Background(), 3, 7, 9, 100)

	require.ErrorIs(t, err, service.ErrResearchGroupNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareResearchGroupUserDeletionRejectsUndissolvedOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newUserRepositoryWithSQL(nil, db)

	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT 1 FROM research_groups`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))

	err = repo.PrepareResearchGroupUserDeletion(context.Background(), 5, 99)

	require.ErrorIs(t, err, service.ErrResearchGroupOwnerMustDissolve)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareResearchGroupUserDeletionDetachesStudentAndAudits(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newUserRepositoryWithSQL(nil, db)

	mock.ExpectExec(`SELECT pg_advisory_xact_lock\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT 1 FROM research_groups`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"one"}))
	mock.ExpectExec(`(?s)WITH removed AS.*member_user_deleted`).
		WithArgs(int64(5), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.PrepareResearchGroupUserDeletion(context.Background(), 5, 99)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
