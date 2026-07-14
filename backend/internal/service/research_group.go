package service

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const (
	ResearchGroupStatusActive    = domain.ResearchGroupStatusActive
	ResearchGroupStatusPaused    = domain.ResearchGroupStatusPaused
	ResearchGroupStatusDissolved = domain.ResearchGroupStatusDissolved

	ResearchGroupMemberStatusPending = domain.ResearchGroupMemberStatusPending
	ResearchGroupMemberStatusActive  = domain.ResearchGroupMemberStatusActive
	ResearchGroupMemberStatusPaused  = domain.ResearchGroupMemberStatusPaused
	ResearchGroupMemberStatusRemoved = domain.ResearchGroupMemberStatusRemoved

	FundingSourceSelf          = domain.FundingSourceSelf
	FundingSourceResearchGroup = domain.FundingSourceResearchGroup
)

var (
	ErrResearchGroupNotFound       = infraerrors.NotFound("RESEARCH_GROUP_NOT_FOUND", "research group not found")
	ErrResearchGroupAlreadyExists  = infraerrors.Conflict("RESEARCH_GROUP_ALREADY_EXISTS", "user already owns a research group")
	ErrResearchGroupForbidden      = infraerrors.Forbidden("RESEARCH_GROUP_FORBIDDEN", "research group operation is not permitted")
	ErrResearchGroupMemberNotFound = infraerrors.NotFound("RESEARCH_GROUP_MEMBER_NOT_FOUND", "research group member not found")
	// This deliberately covers not-found, disabled, non-user, owner and already-assigned
	// targets so the invitation endpoint cannot be used to enumerate accounts.
	ErrResearchGroupMemberNotEligible = infraerrors.New(http.StatusBadRequest, "RESEARCH_GROUP_MEMBER_NOT_ELIGIBLE", "member cannot be added to this research group")
	ErrResearchGroupInvitationInvalid = infraerrors.Conflict("RESEARCH_GROUP_INVITATION_INVALID", "research group invitation is no longer available")
	ErrResearchGroupInvalidQuota      = infraerrors.New(http.StatusBadRequest, "RESEARCH_GROUP_INVALID_QUOTA", "monthly limit must be a finite non-negative amount")
	ErrResearchGroupInvalidName       = infraerrors.New(http.StatusBadRequest, "RESEARCH_GROUP_INVALID_NAME", "research group name is required and must not exceed 100 characters")
	ErrResearchGroupInvalidStatus     = infraerrors.New(http.StatusBadRequest, "RESEARCH_GROUP_INVALID_STATUS", "invalid research group status")
	ErrResearchGroupOwnerMustDissolve = infraerrors.Conflict("RESEARCH_GROUP_OWNER_MUST_DISSOLVE", "dissolve the research group before deleting its owner")
)

type ResearchGroup struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	OwnerUserID   int64      `json:"owner_user_id"`
	OwnerEmail    string     `json:"owner_email,omitempty"`
	OwnerUsername string     `json:"owner_username,omitempty"`
	OwnerBalance  *float64   `json:"owner_balance,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DissolvedAt   *time.Time `json:"dissolved_at,omitempty"`
}

type ResearchGroupMember struct {
	ID                  int64      `json:"id"`
	ResearchGroupID     int64      `json:"research_group_id"`
	UserID              int64      `json:"user_id"`
	Email               string     `json:"email,omitempty"`
	Username            string     `json:"username,omitempty"`
	ResearchGroupName   string     `json:"research_group_name,omitempty"`
	OwnerEmail          string     `json:"owner_email,omitempty"`
	OwnerUsername       string     `json:"owner_username,omitempty"`
	Status              string     `json:"status"`
	MonthlyLimitUSD     float64    `json:"monthly_limit_usd"`
	MonthlyUsageUSD     float64    `json:"monthly_usage_usd"`
	MonthlyReservedUSD  float64    `json:"monthly_reserved_usd"`
	MonthlyRemainingUSD float64    `json:"monthly_remaining_usd"`
	UsageWindowStart    time.Time  `json:"usage_window_start"`
	ResetsAt            time.Time  `json:"resets_at"`
	InvitedAt           time.Time  `json:"invited_at"`
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
	PausedAt            *time.Time `json:"paused_at,omitempty"`
	RemovedAt           *time.Time `json:"removed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (m *ResearchGroupMember) FillDerived() {
	if m == nil {
		return
	}
	m.MonthlyRemainingUSD = math.Max(0, m.MonthlyLimitUSD-m.MonthlyUsageUSD-m.MonthlyReservedUSD)
	windowStart := timezone.StartOfMonth(m.UsageWindowStart)
	m.UsageWindowStart = windowStart
	m.ResetsAt = time.Date(windowStart.Year(), windowStart.Month()+1, 1, 0, 0, 0, 0, timezone.Location())
}

type ResearchGroupUsageSummary struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	RequestCount      int64   `json:"request_count"`
	ActiveMemberCount int64   `json:"active_member_count"`
}

type ResearchGroupContext struct {
	Role         string                     `json:"role"`
	Group        *ResearchGroup             `json:"group"`
	Member       *ResearchGroupMember       `json:"member,omitempty"`
	Members      []ResearchGroupMember      `json:"members,omitempty"`
	UsageSummary *ResearchGroupUsageSummary `json:"usage_summary,omitempty"`
}

// ResearchGroupFundingContext is the stable integration boundary used by
// billing preflight. A nil result means the caller should use self funding.
type ResearchGroupFundingContext struct {
	ResearchGroupID       int64     `json:"research_group_id"`
	ResearchGroupMemberID int64     `json:"research_group_member_id"`
	MemberUserID          int64     `json:"member_user_id"`
	OwnerUserID           int64     `json:"owner_user_id"`
	CallerUserID          int64     `json:"caller_user_id"`
	PayerUserID           int64     `json:"payer_user_id"`
	GroupStatus           string    `json:"group_status"`
	MemberStatus          string    `json:"member_status"`
	MonthlyLimitUSD       float64   `json:"monthly_limit_usd"`
	MonthlyUsageUSD       float64   `json:"monthly_usage_usd"`
	MonthlyReservedUSD    float64   `json:"monthly_reserved_usd"`
	UsageWindowStart      time.Time `json:"usage_window_start"`
	PayerBalance          float64   `json:"payer_balance"`
}

func (c *ResearchGroupFundingContext) RemainingAt(now time.Time) float64 {
	if c == nil || c.MonthlyLimitUSD <= 0 {
		return 0
	}
	usage, reserved := c.MonthlyUsageUSD, c.MonthlyReservedUSD
	monthStart := timezone.StartOfMonth(now)
	if c.UsageWindowStart.Before(monthStart) {
		usage = 0
	}
	return math.Max(0, c.MonthlyLimitUSD-usage-reserved)
}

type ResearchGroupUsageFilter struct {
	MemberID *int64
	Start    *time.Time
	End      *time.Time
	Page     int
	PageSize int
}

type ResearchGroupUsageItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	MemberID  *int64    `json:"member_id,omitempty"`
	RequestID string    `json:"request_id"`
	Model     string    `json:"model"`
	TotalCost float64   `json:"total_cost"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
}

type ResearchGroupUsagePage struct {
	Items    []ResearchGroupUsageItem `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Pages    int                      `json:"pages"`
}

type ResearchGroupAdminPage struct {
	Items    []ResearchGroup `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Pages    int             `json:"pages"`
}

type ResearchGroupRepository interface {
	Create(ctx context.Context, ownerUserID int64, name string) (*ResearchGroup, error)
	GetByID(ctx context.Context, id int64) (*ResearchGroup, error)
	GetByOwnerUserID(ctx context.Context, ownerUserID int64) (*ResearchGroup, error)
	Update(ctx context.Context, groupID, actorUserID int64, name, status string) (*ResearchGroup, error)
	Dissolve(ctx context.Context, groupID, actorUserID int64) error

	InviteMember(ctx context.Context, groupID, userID, actorUserID int64, monthlyLimitUSD float64) (*ResearchGroupMember, error)
	GetMemberByID(ctx context.Context, memberID int64) (*ResearchGroupMember, error)
	GetEffectiveMemberByUserID(ctx context.Context, userID int64) (*ResearchGroupMember, error)
	ListMembers(ctx context.Context, groupID int64) ([]ResearchGroupMember, error)
	ListInvitations(ctx context.Context, userID int64) ([]ResearchGroupMember, error)
	RespondInvitation(ctx context.Context, memberID, userID int64, accept bool) (*ResearchGroupMember, error)
	UpdateMember(ctx context.Context, groupID, memberID, actorUserID int64, monthlyLimitUSD *float64, status *string) (*ResearchGroupMember, error)
	ResetMemberMonth(ctx context.Context, groupID, memberID, actorUserID int64) (*ResearchGroupMember, error)
	RemoveMember(ctx context.Context, groupID, memberID, actorUserID int64, action string) error
	ResetExpiredMemberWindows(ctx context.Context, groupID, userID int64) error

	GetFundingContextByUserID(ctx context.Context, userID int64) (*ResearchGroupFundingContext, error)
	GetUsageSummary(ctx context.Context, groupID, payerUserID int64) (*ResearchGroupUsageSummary, error)
	ListFundedUsage(ctx context.Context, groupID, payerUserID int64, filter ResearchGroupUsageFilter) (*ResearchGroupUsagePage, error)
	AdminList(ctx context.Context, page, pageSize int) (*ResearchGroupAdminPage, error)
}

// ResearchGroupUserDeletionPreparer is an optional UserRepository capability.
// It blocks active owners and detaches students before users are soft-deleted.
type ResearchGroupUserDeletionPreparer interface {
	PrepareResearchGroupUserDeletion(ctx context.Context, userID, actorUserID int64) error
}

type ResearchGroupService struct {
	repo         ResearchGroupRepository
	userRepo     UserRepository
	fundingCache ResearchGroupFundingCacheInvalidator
}

func NewResearchGroupService(repo ResearchGroupRepository, userRepo UserRepository) *ResearchGroupService {
	return &ResearchGroupService{repo: repo, userRepo: userRepo}
}

func ProvideResearchGroupService(repo ResearchGroupRepository, userRepo UserRepository, billingCache *BillingCacheService) *ResearchGroupService {
	service := NewResearchGroupService(repo, userRepo)
	service.fundingCache = billingCache
	return service
}

func (s *ResearchGroupService) SetFundingCacheInvalidator(cache ResearchGroupFundingCacheInvalidator) {
	s.fundingCache = cache
}

func (s *ResearchGroupService) invalidateFunding(ctx context.Context, userID int64) {
	if s != nil && s.fundingCache != nil && userID > 0 {
		_ = s.fundingCache.InvalidateResearchGroupFunding(ctx, userID)
	}
}

func (s *ResearchGroupService) invalidateMembers(ctx context.Context, members []ResearchGroupMember) {
	for i := range members {
		s.invalidateFunding(ctx, members[i].UserID)
	}
}

func normalizeResearchGroupName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 100 {
		return "", ErrResearchGroupInvalidName
	}
	return name, nil
}

func validateResearchGroupQuota(value float64) error {
	// Backed by DECIMAL(20,10): at most ten digits may appear before the
	// decimal point. Reject overflow as a stable client error instead of a DB 500.
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value >= 1e10 {
		return ErrResearchGroupInvalidQuota
	}
	return nil
}

func (s *ResearchGroupService) Create(ctx context.Context, ownerUserID int64, name string) (*ResearchGroupContext, error) {
	name, err := normalizeResearchGroupName(name)
	if err != nil {
		return nil, err
	}
	owner, err := s.userRepo.GetByID(ctx, ownerUserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrResearchGroupForbidden
		}
		return nil, err
	}
	if owner == nil || owner.Role != RoleUser || !owner.IsActive() {
		return nil, ErrResearchGroupForbidden
	}
	if _, err = s.repo.GetEffectiveMemberByUserID(ctx, ownerUserID); err == nil {
		return nil, ErrResearchGroupAlreadyExists
	} else if !errors.Is(err, ErrResearchGroupMemberNotFound) {
		return nil, err
	}
	if _, err = s.repo.Create(ctx, ownerUserID, name); err != nil {
		return nil, err
	}
	return s.GetContext(ctx, ownerUserID)
}

// GetAuthContext returns the compact research-group context embedded in login,
// /auth/me and profile responses. Owner member lists and funded-usage summaries
// remain on the dedicated research-group endpoint so authentication stays O(1).
func (s *ResearchGroupService) GetAuthContext(ctx context.Context, userID int64) (*ResearchGroupContext, error) {
	group, err := s.repo.GetByOwnerUserID(ctx, userID)
	if err == nil {
		return &ResearchGroupContext{Role: "owner", Group: group}, nil
	}
	if !errors.Is(err, ErrResearchGroupNotFound) {
		return nil, err
	}
	return s.getMemberContext(ctx, userID)
}

func (s *ResearchGroupService) GetContext(ctx context.Context, userID int64) (*ResearchGroupContext, error) {
	group, err := s.repo.GetByOwnerUserID(ctx, userID)
	if err == nil {
		if err := s.repo.ResetExpiredMemberWindows(ctx, group.ID, 0); err != nil {
			return nil, err
		}
		members, err := s.repo.ListMembers(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		summary, err := s.repo.GetUsageSummary(ctx, group.ID, group.OwnerUserID)
		if err != nil {
			return nil, err
		}
		return &ResearchGroupContext{Role: "owner", Group: group, Members: members, UsageSummary: summary}, nil
	}
	if !errors.Is(err, ErrResearchGroupNotFound) {
		return nil, err
	}
	return s.getMemberContext(ctx, userID)
}

func (s *ResearchGroupService) getMemberContext(ctx context.Context, userID int64) (*ResearchGroupContext, error) {
	member, err := s.repo.GetEffectiveMemberByUserID(ctx, userID)
	if errors.Is(err, ErrResearchGroupMemberNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := s.repo.ResetExpiredMemberWindows(ctx, member.ResearchGroupID, userID); err != nil {
		return nil, err
	}
	member, err = s.repo.GetMemberByID(ctx, member.ID)
	if err != nil {
		return nil, err
	}
	member.FillDerived()
	group, err := s.repo.GetByID(ctx, member.ResearchGroupID)
	if err != nil {
		return nil, err
	}
	// A sponsored student may see the owner identity, but not the owner's live
	// wallet balance. Their dashboard uses the member quota and personal balance.
	group.OwnerBalance = nil
	return &ResearchGroupContext{Role: "member", Group: group, Member: member}, nil
}

func (s *ResearchGroupService) Update(ctx context.Context, ownerUserID int64, name *string, status *string) (*ResearchGroupContext, error) {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	newName, newStatus := group.Name, group.Status
	if name != nil {
		newName, err = normalizeResearchGroupName(*name)
		if err != nil {
			return nil, err
		}
	}
	if status != nil {
		if *status != ResearchGroupStatusActive && *status != ResearchGroupStatusPaused {
			return nil, ErrResearchGroupInvalidStatus
		}
		newStatus = *status
	}
	var members []ResearchGroupMember
	if newStatus != group.Status {
		members, err = s.repo.ListMembers(ctx, group.ID)
		if err != nil {
			return nil, err
		}
	}
	if _, err := s.repo.Update(ctx, group.ID, ownerUserID, newName, newStatus); err != nil {
		return nil, err
	}
	s.invalidateMembers(ctx, members)
	return s.GetContext(ctx, ownerUserID)
}

func (s *ResearchGroupService) Dissolve(ctx context.Context, ownerUserID int64) error {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	members, err := s.repo.ListMembers(ctx, group.ID)
	if err != nil {
		return err
	}
	if err := s.repo.Dissolve(ctx, group.ID, ownerUserID); err != nil {
		return err
	}
	s.invalidateMembers(ctx, members)
	return nil
}

func (s *ResearchGroupService) Invite(ctx context.Context, ownerUserID int64, email string, limit float64) (*ResearchGroupMember, error) {
	if err := validateResearchGroupQuota(limit); err != nil {
		return nil, err
	}
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	target, err := s.userRepo.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrResearchGroupMemberNotEligible
		}
		return nil, err
	}
	if target == nil || target.ID == ownerUserID || target.Role != RoleUser || !target.IsActive() {
		return nil, ErrResearchGroupMemberNotEligible
	}
	if _, err = s.repo.GetByOwnerUserID(ctx, target.ID); err == nil {
		return nil, ErrResearchGroupMemberNotEligible
	} else if !errors.Is(err, ErrResearchGroupNotFound) {
		return nil, err
	}
	if _, err = s.repo.GetEffectiveMemberByUserID(ctx, target.ID); err == nil {
		return nil, ErrResearchGroupMemberNotEligible
	} else if !errors.Is(err, ErrResearchGroupMemberNotFound) {
		return nil, err
	}
	member, err := s.repo.InviteMember(ctx, group.ID, target.ID, ownerUserID, limit)
	if errors.Is(err, ErrResearchGroupAlreadyExists) {
		return nil, ErrResearchGroupMemberNotEligible
	}
	if err == nil {
		s.invalidateFunding(ctx, target.ID)
	}
	return member, err
}

func (s *ResearchGroupService) ListInvitations(ctx context.Context, userID int64) ([]ResearchGroupMember, error) {
	return s.repo.ListInvitations(ctx, userID)
}

func (s *ResearchGroupService) RespondInvitation(ctx context.Context, userID, memberID int64, accept bool) (*ResearchGroupContext, error) {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, ErrResearchGroupMemberNotFound) {
			return nil, ErrResearchGroupInvitationInvalid
		}
		return nil, err
	}
	if member.UserID != userID {
		return nil, ErrResearchGroupInvitationInvalid
	}
	if _, err := s.repo.RespondInvitation(ctx, memberID, userID, accept); err != nil {
		return nil, err
	}
	s.invalidateFunding(ctx, userID)
	if !accept {
		return nil, nil
	}
	return s.GetContext(ctx, userID)
}

func (s *ResearchGroupService) UpdateMember(ctx context.Context, ownerUserID, memberID int64, limit *float64, status *string) (*ResearchGroupMember, error) {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	if limit != nil {
		if err := validateResearchGroupQuota(*limit); err != nil {
			return nil, err
		}
	}
	if status != nil && *status != ResearchGroupMemberStatusActive && *status != ResearchGroupMemberStatusPaused {
		return nil, ErrResearchGroupInvalidStatus
	}
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member.ResearchGroupID != group.ID {
		return nil, ErrResearchGroupMemberNotFound
	}
	if status != nil {
		if member.Status != ResearchGroupMemberStatusActive && member.Status != ResearchGroupMemberStatusPaused {
			return nil, ErrResearchGroupInvitationInvalid
		}
	}
	updated, err := s.repo.UpdateMember(ctx, group.ID, memberID, ownerUserID, limit, status)
	if err == nil {
		s.invalidateFunding(ctx, member.UserID)
	}
	return updated, err
}

func (s *ResearchGroupService) ResetMemberMonth(ctx context.Context, ownerUserID, memberID int64) (*ResearchGroupMember, error) {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member.ResearchGroupID != group.ID {
		return nil, ErrResearchGroupMemberNotFound
	}
	updated, err := s.repo.ResetMemberMonth(ctx, group.ID, memberID, ownerUserID)
	if err == nil {
		s.invalidateFunding(ctx, member.UserID)
	}
	return updated, err
}

func (s *ResearchGroupService) RemoveMember(ctx context.Context, ownerUserID, memberID int64) error {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.ResearchGroupID != group.ID {
		return ErrResearchGroupMemberNotFound
	}
	if err := s.repo.RemoveMember(ctx, group.ID, memberID, ownerUserID, "member_removed"); err != nil {
		return err
	}
	s.invalidateFunding(ctx, member.UserID)
	return nil
}

func (s *ResearchGroupService) Leave(ctx context.Context, userID int64) error {
	member, err := s.repo.GetEffectiveMemberByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.repo.RemoveMember(ctx, member.ResearchGroupID, member.ID, userID, "member_left"); err != nil {
		return err
	}
	s.invalidateFunding(ctx, userID)
	return nil
}

func (s *ResearchGroupService) ListUsage(ctx context.Context, ownerUserID int64, filter ResearchGroupUsageFilter) (*ResearchGroupUsagePage, error) {
	group, err := s.requireOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.ListFundedUsage(ctx, group.ID, ownerUserID, filter)
}

func (s *ResearchGroupService) AdminList(ctx context.Context, page, pageSize int) (*ResearchGroupAdminPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.AdminList(ctx, page, pageSize)
}

func (s *ResearchGroupService) AdminGet(ctx context.Context, groupID int64) (*ResearchGroupContext, error) {
	group, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ResetExpiredMemberWindows(ctx, group.ID, 0); err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	summary, err := s.repo.GetUsageSummary(ctx, group.ID, group.OwnerUserID)
	if err != nil {
		return nil, err
	}
	return &ResearchGroupContext{Role: "admin", Group: group, Members: members, UsageSummary: summary}, nil
}

func (s *ResearchGroupService) AdminDissolve(ctx context.Context, adminUserID, groupID int64) error {
	if _, err := s.repo.GetByID(ctx, groupID); err != nil {
		return err
	}
	members, err := s.repo.ListMembers(ctx, groupID)
	if err != nil {
		return err
	}
	if err := s.repo.Dissolve(ctx, groupID, adminUserID); err != nil {
		return err
	}
	s.invalidateMembers(ctx, members)
	return nil
}

func (s *ResearchGroupService) AdminDetach(ctx context.Context, adminUserID, groupID, memberID int64) error {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.ResearchGroupID != groupID {
		return ErrResearchGroupMemberNotFound
	}
	if err := s.repo.RemoveMember(ctx, groupID, memberID, adminUserID, "admin_detached"); err != nil {
		return err
	}
	s.invalidateFunding(ctx, member.UserID)
	return nil
}

func (s *ResearchGroupService) requireOwner(ctx context.Context, ownerUserID int64) (*ResearchGroup, error) {
	group, err := s.repo.GetByOwnerUserID(ctx, ownerUserID)
	if errors.Is(err, ErrResearchGroupNotFound) {
		return nil, ErrResearchGroupForbidden
	}
	return group, err
}
