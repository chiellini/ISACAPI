package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

type researchGroupServiceRepoStub struct {
	ResearchGroupRepository
	ownerGroup        *ResearchGroup
	ownerErr          error
	effectiveMember   *ResearchGroupMember
	effectiveErr      error
	member            *ResearchGroupMember
	memberErr         error
	group             *ResearchGroup
	members           []ResearchGroupMember
	listMembersCalled bool
	updateCalled      bool
}

type researchGroupUserRepoStub struct {
	UserRepository
	byID       *User
	byIDErr    error
	byEmail    *User
	byEmailErr error
}

func (r *researchGroupUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return r.byID, r.byIDErr
}

func (r *researchGroupUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	return r.byEmail, r.byEmailErr
}

func (r *researchGroupServiceRepoStub) GetByOwnerUserID(context.Context, int64) (*ResearchGroup, error) {
	return r.ownerGroup, r.ownerErr
}

func (r *researchGroupServiceRepoStub) GetEffectiveMemberByUserID(context.Context, int64) (*ResearchGroupMember, error) {
	return r.effectiveMember, r.effectiveErr
}

func (r *researchGroupServiceRepoStub) GetMemberByID(context.Context, int64) (*ResearchGroupMember, error) {
	return r.member, r.memberErr
}

func (r *researchGroupServiceRepoStub) GetByID(context.Context, int64) (*ResearchGroup, error) {
	return r.group, nil
}

func (r *researchGroupServiceRepoStub) ResetExpiredMemberWindows(context.Context, int64, int64) error {
	return nil
}

func (r *researchGroupServiceRepoStub) ListMembers(context.Context, int64) ([]ResearchGroupMember, error) {
	r.listMembersCalled = true
	return r.members, nil
}

func (r *researchGroupServiceRepoStub) GetUsageSummary(context.Context, int64, int64) (*ResearchGroupUsageSummary, error) {
	return &ResearchGroupUsageSummary{}, nil
}

func (r *researchGroupServiceRepoStub) UpdateMember(context.Context, int64, int64, int64, *float64, *string) (*ResearchGroupMember, error) {
	r.updateCalled = true
	return r.member, nil
}

func TestResearchGroupServiceOwnerCannotActivatePendingInvitation(t *testing.T) {
	repo := &researchGroupServiceRepoStub{
		ownerGroup: &ResearchGroup{ID: 10, OwnerUserID: 1, Status: ResearchGroupStatusActive},
		member: &ResearchGroupMember{
			ID:              20,
			ResearchGroupID: 10,
			UserID:          2,
			Status:          ResearchGroupMemberStatusPending,
		},
	}
	service := NewResearchGroupService(repo, nil)
	status := ResearchGroupMemberStatusActive

	_, err := service.UpdateMember(context.Background(), 1, 20, nil, &status)

	require.ErrorIs(t, err, ErrResearchGroupInvitationInvalid)
	require.False(t, repo.updateCalled)
}

func TestResearchGroupServiceMemberContextRedactsOwnerBalance(t *testing.T) {
	window := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	ownerBalance := float64(999)
	member := &ResearchGroupMember{
		ID:               20,
		ResearchGroupID:  10,
		UserID:           2,
		Status:           ResearchGroupMemberStatusActive,
		MonthlyLimitUSD:  50,
		MonthlyUsageUSD:  12,
		UsageWindowStart: window,
	}
	repo := &researchGroupServiceRepoStub{
		ownerErr:        ErrResearchGroupNotFound,
		effectiveMember: member,
		member:          member,
		group: &ResearchGroup{
			ID:           10,
			OwnerUserID:  1,
			OwnerBalance: &ownerBalance,
			Status:       ResearchGroupStatusActive,
		},
	}
	service := NewResearchGroupService(repo, nil)

	context, err := service.GetContext(context.Background(), 2)

	require.NoError(t, err)
	require.Equal(t, "member", context.Role)
	require.Nil(t, context.Group.OwnerBalance)
	require.Equal(t, float64(38), context.Member.MonthlyRemainingUSD)
	encoded, err := json.Marshal(context)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "owner_balance")
}

func TestResearchGroupOwnerContextSerializesZeroBalance(t *testing.T) {
	zero := float64(0)
	encoded, err := json.Marshal(&ResearchGroupContext{
		Role:  "owner",
		Group: &ResearchGroup{ID: 10, OwnerBalance: &zero},
	})

	require.NoError(t, err)
	require.Contains(t, string(encoded), `"owner_balance":0`)
}

func TestResearchGroupOwnerAuthContextIsLightweight(t *testing.T) {
	repo := &researchGroupServiceRepoStub{
		ownerGroup: &ResearchGroup{ID: 10, OwnerUserID: 1, Status: ResearchGroupStatusActive},
	}
	service := NewResearchGroupService(repo, nil)

	context, err := service.GetAuthContext(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, "owner", context.Role)
	require.Nil(t, context.Members)
	require.Nil(t, context.UsageSummary)
	require.False(t, repo.listMembersCalled)
}

func TestResearchGroupFundingRemainingAcrossMonthPreservesOutstandingReservation(t *testing.T) {
	funding := &ResearchGroupFundingContext{
		MonthlyLimitUSD:    100,
		MonthlyUsageUSD:    75,
		MonthlyReservedUSD: 20,
		UsageWindowStart:   time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
	}

	remaining := funding.RemainingAt(time.Date(2026, time.July, 2, 12, 0, 0, 0, time.UTC))

	require.Equal(t, float64(80), remaining)
}

func TestResearchGroupMemberDerivedResetUsesAccountingTimezone(t *testing.T) {
	require.NoError(t, timezone.Init("Asia/Taipei"))
	t.Cleanup(func() { _ = timezone.Init("UTC") })
	member := &ResearchGroupMember{
		MonthlyLimitUSD:    50,
		MonthlyUsageUSD:    12,
		MonthlyReservedUSD: 3,
		UsageWindowStart:   time.Date(2026, time.June, 30, 16, 0, 0, 0, time.UTC),
	}

	member.FillDerived()

	require.Equal(t, time.July, member.UsageWindowStart.Month())
	require.Equal(t, 1, member.UsageWindowStart.Day())
	require.Equal(t, 0, member.UsageWindowStart.Hour())
	require.Equal(t, time.August, member.ResetsAt.Month())
	require.Equal(t, 1, member.ResetsAt.Day())
	require.Equal(t, 0, member.ResetsAt.Hour())
	require.Equal(t, timezone.Location(), member.ResetsAt.Location())
	require.Equal(t, float64(35), member.MonthlyRemainingUSD)
}

func TestResearchGroupServiceReturnsRepositoryErrorsFromContext(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := NewResearchGroupService(&researchGroupServiceRepoStub{ownerErr: repoErr}, nil)

	_, err := service.GetContext(context.Background(), 2)

	require.ErrorIs(t, err, repoErr)
}

func TestResearchGroupQuotaRejectsDatabaseNumericOverflow(t *testing.T) {
	require.ErrorIs(t, validateResearchGroupQuota(1e10), ErrResearchGroupInvalidQuota)
}

func TestResearchGroupCreatePropagatesUserRepositoryFailure(t *testing.T) {
	repoErr := errors.New("database unavailable")
	service := NewResearchGroupService(&researchGroupServiceRepoStub{}, &researchGroupUserRepoStub{byIDErr: repoErr})

	_, err := service.Create(context.Background(), 1, "Lab")

	require.ErrorIs(t, err, repoErr)
}

func TestResearchGroupInvitePropagatesUserRepositoryFailure(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &researchGroupServiceRepoStub{ownerGroup: &ResearchGroup{ID: 10, OwnerUserID: 1, Status: ResearchGroupStatusActive}}
	service := NewResearchGroupService(repo, &researchGroupUserRepoStub{byEmailErr: repoErr})

	_, err := service.Invite(context.Background(), 1, "student@example.com", 10)

	require.ErrorIs(t, err, repoErr)
}
