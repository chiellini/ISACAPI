package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type researchGroupRepository struct {
	db *sql.DB
}

// Serialize all transitions that can make a user either a group owner or an
// effective member. The per-table partial unique indexes cannot by themselves
// prevent a concurrent create-group / invite-member race across two tables.
const researchGroupUserLockNamespace int64 = 0x5247000000000000

func lockResearchGroupUser(ctx context.Context, tx sqlQueryExecutor, userID int64) error {
	key := researchGroupUserLockNamespace | (userID & 0x0000ffffffffffff)
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, key)
	return err
}

func NewResearchGroupRepository(db *sql.DB) service.ResearchGroupRepository {
	return &researchGroupRepository{db: db}
}

type researchGroupRowScanner interface {
	Scan(dest ...any) error
}

const researchGroupSelect = `
SELECT rg.id, rg.name, COALESCE(rg.owner_user_id, 0), COALESCE(owner.email, ''),
       COALESCE(owner.username, ''), COALESCE(owner.balance, 0),
       rg.status, rg.created_at, rg.updated_at, rg.dissolved_at
FROM research_groups rg
LEFT JOIN users owner ON owner.id = rg.owner_user_id`

func scanResearchGroup(row researchGroupRowScanner) (*service.ResearchGroup, error) {
	group := &service.ResearchGroup{}
	var ownerBalance float64
	if err := row.Scan(
		&group.ID, &group.Name, &group.OwnerUserID, &group.OwnerEmail,
		&group.OwnerUsername, &ownerBalance, &group.Status,
		&group.CreatedAt, &group.UpdatedAt, &group.DissolvedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupNotFound
		}
		return nil, err
	}
	group.OwnerBalance = &ownerBalance
	return group, nil
}

const researchGroupMemberSelect = `
SELECT m.id, m.research_group_id, m.user_id, member_user.email, member_user.username,
       rg.name, COALESCE(owner.email, ''), COALESCE(owner.username, ''), m.status,
       m.monthly_limit_usd, m.monthly_usage_usd, m.monthly_reserved_usd,
       m.usage_window_start, m.invited_at, m.accepted_at, m.paused_at,
       m.removed_at, m.created_at, m.updated_at
FROM research_group_members m
JOIN research_groups rg ON rg.id = m.research_group_id
JOIN users member_user ON member_user.id = m.user_id
LEFT JOIN users owner ON owner.id = rg.owner_user_id`

func scanResearchGroupMember(row researchGroupRowScanner) (*service.ResearchGroupMember, error) {
	member := &service.ResearchGroupMember{}
	if err := row.Scan(
		&member.ID, &member.ResearchGroupID, &member.UserID, &member.Email,
		&member.Username, &member.ResearchGroupName, &member.OwnerEmail,
		&member.OwnerUsername, &member.Status, &member.MonthlyLimitUSD,
		&member.MonthlyUsageUSD, &member.MonthlyReservedUSD,
		&member.UsageWindowStart, &member.InvitedAt, &member.AcceptedAt,
		&member.PausedAt, &member.RemovedAt, &member.CreatedAt, &member.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupMemberNotFound
		}
		return nil, err
	}
	member.FillDerived()
	return member, nil
}

func (r *researchGroupRepository) Create(ctx context.Context, ownerUserID int64, name string) (*service.ResearchGroup, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockResearchGroupUser(ctx, tx, ownerUserID); err != nil {
		return nil, err
	}
	var ownerRole, ownerStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT role, status FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, ownerUserID).
		Scan(&ownerRole, &ownerStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupForbidden
		}
		return nil, err
	}
	if ownerRole != service.RoleUser || ownerStatus != service.StatusActive {
		return nil, service.ErrResearchGroupForbidden
	}
	var hasEffectiveMembership bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM research_group_members m
			JOIN research_groups rg ON rg.id = m.research_group_id
			WHERE m.user_id = $1 AND m.status IN ('pending', 'active', 'paused')
			  AND rg.status <> 'dissolved'
		)`, ownerUserID).Scan(&hasEffectiveMembership); err != nil {
		return nil, err
	}
	if hasEffectiveMembership {
		return nil, service.ErrResearchGroupAlreadyExists
	}

	var groupID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO research_groups (name, owner_user_id)
		VALUES ($1, $2)
		RETURNING id`, name, ownerUserID).Scan(&groupID)
	if err != nil {
		if isResearchGroupUniqueViolation(err) {
			return nil, service.ErrResearchGroupAlreadyExists
		}
		return nil, err
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, nil, ownerUserID, "group_created", nil, nil, nil, map[string]any{"name": name}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, groupID)
}

func (r *researchGroupRepository) GetByID(ctx context.Context, id int64) (*service.ResearchGroup, error) {
	return scanResearchGroup(r.db.QueryRowContext(ctx, researchGroupSelect+` WHERE rg.id = $1`, id))
}

func (r *researchGroupRepository) GetByOwnerUserID(ctx context.Context, ownerUserID int64) (*service.ResearchGroup, error) {
	return scanResearchGroup(r.db.QueryRowContext(ctx, researchGroupSelect+` WHERE rg.owner_user_id = $1 AND rg.status <> 'dissolved'`, ownerUserID))
}

func (r *researchGroupRepository) Update(ctx context.Context, groupID, actorUserID int64, name, status string) (*service.ResearchGroup, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var previousName, previousStatus string
	if err := tx.QueryRowContext(ctx, `SELECT name, status FROM research_groups WHERE id = $1 AND status <> 'dissolved' FOR UPDATE`, groupID).Scan(&previousName, &previousStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupNotFound
		}
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE research_groups SET name = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, groupID, name, status); err != nil {
		return nil, err
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, nil, actorUserID, "group_updated", nil, nil, nil, map[string]any{
		"previous_name": previousName, "name": name, "previous_status": previousStatus, "status": status,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, groupID)
}

func (r *researchGroupRepository) Dissolve(ctx context.Context, groupID, actorUserID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
		UPDATE research_groups
		SET status = 'dissolved', dissolved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status <> 'dissolved'`, groupID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return service.ErrResearchGroupNotFound
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE research_group_members
		SET status = 'removed', removed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE research_group_id = $1 AND status IN ('pending', 'active', 'paused')`, groupID); err != nil {
		return err
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, nil, actorUserID, "group_dissolved", nil, nil, nil, nil); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *researchGroupRepository) InviteMember(ctx context.Context, groupID, userID, actorUserID int64, limit float64) (*service.ResearchGroupMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockResearchGroupUser(ctx, tx, userID); err != nil {
		return nil, err
	}
	var memberRole, memberStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT role, status FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, userID).
		Scan(&memberRole, &memberStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupMemberNotEligible
		}
		return nil, err
	}
	if memberRole != service.RoleUser || memberStatus != service.StatusActive {
		return nil, service.ErrResearchGroupMemberNotEligible
	}
	var unavailable bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM research_groups WHERE owner_user_id = $1 AND status <> 'dissolved'
			UNION ALL
			SELECT 1 FROM research_group_members m
			JOIN research_groups rg ON rg.id = m.research_group_id
			WHERE m.user_id = $1 AND m.status IN ('pending', 'active', 'paused')
			  AND rg.status <> 'dissolved'
		)`, userID).Scan(&unavailable); err != nil {
		return nil, err
	}
	if unavailable {
		return nil, service.ErrResearchGroupAlreadyExists
	}
	var lockedGroupID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM research_groups WHERE id = $1 AND status <> 'dissolved' FOR UPDATE`, groupID).
		Scan(&lockedGroupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupNotFound
		}
		return nil, err
	}

	var memberID int64
	monthStart := timezone.StartOfMonth(timezone.Now())
	err = tx.QueryRowContext(ctx, `
		INSERT INTO research_group_members
		    (research_group_id, user_id, monthly_limit_usd, usage_window_start)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (SELECT 1 FROM research_groups WHERE id = $1 AND status <> 'dissolved')
		RETURNING id`, groupID, userID, limit, monthStart).Scan(&memberID)
	if err != nil {
		if isResearchGroupUniqueViolation(err) {
			return nil, service.ErrResearchGroupAlreadyExists
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupNotFound
		}
		return nil, err
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, &memberID, actorUserID, "member_invited", nil, nil, &limit, nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMemberByID(ctx, memberID)
}

func (r *researchGroupRepository) GetMemberByID(ctx context.Context, memberID int64) (*service.ResearchGroupMember, error) {
	return scanResearchGroupMember(r.db.QueryRowContext(ctx, researchGroupMemberSelect+` WHERE m.id = $1`, memberID))
}

func (r *researchGroupRepository) GetEffectiveMemberByUserID(ctx context.Context, userID int64) (*service.ResearchGroupMember, error) {
	return scanResearchGroupMember(r.db.QueryRowContext(ctx, researchGroupMemberSelect+`
		WHERE m.user_id = $1 AND m.status IN ('pending', 'active', 'paused') AND rg.status <> 'dissolved'`, userID))
}

func (r *researchGroupRepository) ListMembers(ctx context.Context, groupID int64) ([]service.ResearchGroupMember, error) {
	rows, err := r.db.QueryContext(ctx, researchGroupMemberSelect+`
		WHERE m.research_group_id = $1 AND m.status IN ('pending', 'active', 'paused') ORDER BY m.id`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]service.ResearchGroupMember, 0)
	for rows.Next() {
		member, err := scanResearchGroupMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *member)
	}
	return out, rows.Err()
}

func (r *researchGroupRepository) ListInvitations(ctx context.Context, userID int64) ([]service.ResearchGroupMember, error) {
	rows, err := r.db.QueryContext(ctx, researchGroupMemberSelect+`
		WHERE m.user_id = $1 AND m.status = 'pending' AND rg.status <> 'dissolved' ORDER BY m.invited_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]service.ResearchGroupMember, 0)
	for rows.Next() {
		member, err := scanResearchGroupMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *member)
	}
	return out, rows.Err()
}

func (r *researchGroupRepository) RespondInvitation(ctx context.Context, memberID, userID int64, accept bool) (*service.ResearchGroupMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if accept {
		if err := lockResearchGroupUser(ctx, tx, userID); err != nil {
			return nil, err
		}
		var memberRole, memberStatus string
		if err := tx.QueryRowContext(ctx, `
			SELECT role, status FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, userID).
			Scan(&memberRole, &memberStatus); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, service.ErrResearchGroupInvitationInvalid
			}
			return nil, err
		}
		if memberRole != service.RoleUser || memberStatus != service.StatusActive {
			return nil, service.ErrResearchGroupInvitationInvalid
		}
		var ownsGroup bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (SELECT 1 FROM research_groups WHERE owner_user_id = $1 AND status <> 'dissolved')`, userID).
			Scan(&ownsGroup); err != nil {
			return nil, err
		}
		if ownsGroup {
			return nil, service.ErrResearchGroupInvitationInvalid
		}
	}

	status, action := service.ResearchGroupMemberStatusRemoved, "invitation_rejected"
	if accept {
		status, action = service.ResearchGroupMemberStatusActive, "invitation_accepted"
	}
	var groupID int64
	err = tx.QueryRowContext(ctx, `
		UPDATE research_group_members m
		SET status = $3,
		    accepted_at = CASE WHEN $3 = 'active' THEN CURRENT_TIMESTAMP ELSE accepted_at END,
		    removed_at = CASE WHEN $3 = 'removed' THEN CURRENT_TIMESTAMP ELSE removed_at END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE m.id = $1 AND m.user_id = $2 AND m.status = 'pending'
		  AND EXISTS (SELECT 1 FROM research_groups rg WHERE rg.id = m.research_group_id AND rg.status <> 'dissolved')
		RETURNING research_group_id`, memberID, userID, status).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupInvitationInvalid
		}
		return nil, err
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, &memberID, userID, action, nil, nil, nil, nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMemberByID(ctx, memberID)
}

func (r *researchGroupRepository) UpdateMember(ctx context.Context, groupID, memberID, actorUserID int64, limit *float64, status *string) (*service.ResearchGroupMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var previousLimit float64
	var previousStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT monthly_limit_usd, status FROM research_group_members
		WHERE id = $1 AND research_group_id = $2 AND status IN ('pending', 'active', 'paused') FOR UPDATE`, memberID, groupID).
		Scan(&previousLimit, &previousStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupMemberNotFound
		}
		return nil, err
	}
	// Only the student may activate a pending invitation.
	if status != nil && previousStatus == service.ResearchGroupMemberStatusPending {
		return nil, service.ErrResearchGroupInvitationInvalid
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE research_group_members
		SET monthly_limit_usd = COALESCE($3, monthly_limit_usd),
		    status = COALESCE($4, status),
		    paused_at = CASE
		        WHEN $4 = 'paused' THEN CURRENT_TIMESTAMP
		        WHEN $4 = 'active' THEN NULL
		        ELSE paused_at
		    END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND research_group_id = $2
		  AND status IN ('pending', 'active', 'paused')
		  AND ($4::text IS NULL OR status IN ('active', 'paused'))`, memberID, groupID, limit, status)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, service.ErrResearchGroupInvitationInvalid
	}
	newLimit, newStatus := previousLimit, previousStatus
	if limit != nil {
		newLimit = *limit
	}
	if status != nil {
		newStatus = *status
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, &memberID, actorUserID, "member_updated", nil, &previousLimit, &newLimit, map[string]any{
		"previous_status": previousStatus, "status": newStatus,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMemberByID(ctx, memberID)
}

func (r *researchGroupRepository) ResetMemberMonth(ctx context.Context, groupID, memberID, actorUserID int64) (*service.ResearchGroupMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var previousUsage float64
	err = tx.QueryRowContext(ctx, `
		SELECT monthly_usage_usd FROM research_group_members
		WHERE id = $1 AND research_group_id = $2 AND status IN ('pending', 'active', 'paused')
		FOR UPDATE`, memberID, groupID).Scan(&previousUsage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrResearchGroupMemberNotFound
		}
		return nil, err
	}
	monthStart := timezone.StartOfMonth(timezone.Now())
	if _, err := tx.ExecContext(ctx, `
		UPDATE research_group_members
		SET monthly_usage_usd = 0, usage_window_start = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND research_group_id = $2`, memberID, groupID, monthStart); err != nil {
		return nil, err
	}
	zero := float64(0)
	if err := appendResearchGroupAudit(ctx, tx, groupID, &memberID, actorUserID, "member_month_reset", &previousUsage, &previousUsage, &zero, nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMemberByID(ctx, memberID)
}

func (r *researchGroupRepository) RemoveMember(ctx context.Context, groupID, memberID, actorUserID int64, action string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `
		UPDATE research_group_members
		SET status = 'removed', removed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND research_group_id = $2 AND status IN ('pending', 'active', 'paused')`, memberID, groupID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return service.ErrResearchGroupMemberNotFound
	}
	if strings.TrimSpace(action) == "" {
		action = "member_removed"
	}
	if err := appendResearchGroupAudit(ctx, tx, groupID, &memberID, actorUserID, action, nil, nil, nil, nil); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *researchGroupRepository) ResetExpiredMemberWindows(ctx context.Context, groupID, userID int64) error {
	monthStart := timezone.StartOfMonth(timezone.Now())
	_, err := r.db.ExecContext(ctx, `
		UPDATE research_group_members
		SET monthly_usage_usd = 0,
		    usage_window_start = $3, updated_at = CURRENT_TIMESTAMP
		WHERE research_group_id = $1
		  AND ($2 = 0 OR user_id = $2)
		  AND status IN ('pending', 'active', 'paused')
		  AND usage_window_start < $3`, groupID, userID, monthStart)
	return err
}

func (r *researchGroupRepository) GetFundingContextByUserID(ctx context.Context, userID int64) (*service.ResearchGroupFundingContext, error) {
	monthStart := timezone.StartOfMonth(timezone.Now())
	_, err := r.db.ExecContext(ctx, `
		UPDATE research_group_members
		SET monthly_usage_usd = 0,
		    usage_window_start = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND status = 'active'
		  AND usage_window_start < $2`, userID, monthStart)
	if err != nil {
		return nil, err
	}
	funding := &service.ResearchGroupFundingContext{}
	err = r.db.QueryRowContext(ctx, `
		SELECT rg.id, m.id, m.user_id, rg.owner_user_id, rg.status, m.status,
		       m.monthly_limit_usd, m.monthly_usage_usd, m.monthly_reserved_usd,
		       m.usage_window_start, owner.balance
		FROM research_group_members m
		JOIN research_groups rg ON rg.id = m.research_group_id
		JOIN users owner ON owner.id = rg.owner_user_id
		WHERE m.user_id = $1 AND m.status = 'active' AND rg.status = 'active'
		  AND owner.status = 'active' AND owner.deleted_at IS NULL`, userID).Scan(
		&funding.ResearchGroupID, &funding.ResearchGroupMemberID,
		&funding.MemberUserID, &funding.OwnerUserID, &funding.GroupStatus,
		&funding.MemberStatus, &funding.MonthlyLimitUSD, &funding.MonthlyUsageUSD,
		&funding.MonthlyReservedUSD, &funding.UsageWindowStart, &funding.PayerBalance,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	funding.CallerUserID = funding.MemberUserID
	funding.PayerUserID = funding.OwnerUserID
	return funding, nil
}

func (r *researchGroupRepository) GetUsageSummary(ctx context.Context, groupID, payerUserID int64) (*service.ResearchGroupUsageSummary, error) {
	summary := &service.ResearchGroupUsageSummary{}
	monthStart := timezone.StartOfMonth(timezone.Now())
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(ul.actual_cost), 0), COUNT(*),
		       (SELECT COUNT(*) FROM research_group_members m WHERE m.research_group_id = $1 AND m.status = 'active')
		FROM usage_logs ul
		WHERE ul.research_group_id = $1 AND ($2 = 0 OR ul.payer_user_id = $2)
		  AND ul.funding_source = 'research_group' AND ul.created_at >= $3`, groupID, payerUserID, monthStart).
		Scan(&summary.TotalCostUSD, &summary.RequestCount, &summary.ActiveMemberCount)
	return summary, err
}

func (r *researchGroupRepository) ListFundedUsage(ctx context.Context, groupID, payerUserID int64, filter service.ResearchGroupUsageFilter) (*service.ResearchGroupUsagePage, error) {
	where := []string{"ul.research_group_id = $1", "ul.payer_user_id = $2", "ul.funding_source = 'research_group'"}
	args := []any{groupID, payerUserID}
	if filter.MemberID != nil {
		args = append(args, *filter.MemberID)
		where = append(where, fmt.Sprintf("ul.research_group_member_id = $%d", len(args)))
	}
	if filter.Start != nil {
		args = append(args, *filter.Start)
		where = append(where, fmt.Sprintf("ul.created_at >= $%d", len(args)))
	}
	if filter.End != nil {
		args = append(args, *filter.End)
		where = append(where, fmt.Sprintf("ul.created_at < $%d", len(args)))
	}
	whereSQL := strings.Join(where, " AND ")
	page := &service.ResearchGroupUsagePage{Page: filter.Page, PageSize: filter.PageSize, Items: []service.ResearchGroupUsageItem{}}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM usage_logs ul WHERE `+whereSQL, args...).Scan(&page.Total); err != nil {
		return nil, err
	}
	page.Pages = int(math.Ceil(float64(page.Total) / float64(page.PageSize)))
	if page.Pages < 1 {
		page.Pages = 1
	}
	args = append(args, page.PageSize, (page.Page-1)*page.PageSize)
	query := `
		SELECT ul.id, ul.user_id, ul.research_group_member_id, ul.request_id, ul.model,
		       ul.actual_cost, ul.created_at, u.email, u.username
		FROM usage_logs ul JOIN users u ON u.id = ul.user_id
		WHERE ` + whereSQL + fmt.Sprintf(" ORDER BY ul.created_at DESC, ul.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item service.ResearchGroupUsageItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.MemberID, &item.RequestID, &item.Model, &item.TotalCost, &item.CreatedAt, &item.Email, &item.Username); err != nil {
			return nil, err
		}
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (r *researchGroupRepository) AdminList(ctx context.Context, pageNumber, pageSize int) (*service.ResearchGroupAdminPage, error) {
	page := &service.ResearchGroupAdminPage{Page: pageNumber, PageSize: pageSize, Items: []service.ResearchGroup{}}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM research_groups`).Scan(&page.Total); err != nil {
		return nil, err
	}
	page.Pages = int(math.Ceil(float64(page.Total) / float64(pageSize)))
	if page.Pages < 1 {
		page.Pages = 1
	}
	rows, err := r.db.QueryContext(ctx, researchGroupSelect+` ORDER BY rg.created_at DESC, rg.id DESC LIMIT $1 OFFSET $2`, pageSize, (pageNumber-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		group, err := scanResearchGroup(rows)
		if err != nil {
			return nil, err
		}
		page.Items = append(page.Items, *group)
	}
	return page, rows.Err()
}

func appendResearchGroupAudit(ctx context.Context, tx *sql.Tx, groupID int64, memberID *int64, actorUserID int64, action string, amount, previous, next *float64, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	var actor any
	if actorUserID > 0 {
		actor = actorUserID
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO research_group_quota_audits
		    (research_group_id, member_id, actor_user_id, action, amount_usd, previous_value_usd, new_value_usd, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)`,
		groupID, memberID, actor, action, amount, previous, next, string(payload))
	return err
}

func isResearchGroupUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

var _ service.ResearchGroupRepository = (*researchGroupRepository)(nil)
