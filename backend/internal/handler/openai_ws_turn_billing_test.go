package handler

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpenAIWSTurnBillingStateChecksOnlyNewTurnsAndKeepsSnapshotsIsolated(t *testing.T) {
	initial := &service.BillingDecision{
		CallerUserID:          11,
		PayerUserID:           22,
		ResearchGroupID:       33,
		ResearchGroupMemberID: 44,
		FundingSource:         service.FundingSourceResearchGroup,
	}
	checkCalls := 0
	state := newOpenAIWSTurnBillingState(11, initial, func(context.Context) (*service.BillingDecision, error) {
		checkCalls++
		return service.SelfBillingDecision(11), nil
	})

	// Turn 1 was checked before upstream selection; the lifecycle hook must not
	// consume RPM or resolve funding a second time.
	require.NoError(t, state.Admit(context.Background(), 1))
	require.Zero(t, checkCalls)

	// A later turn is checked once, while a same-turn transport retry reuses the
	// immutable admission snapshot.
	require.NoError(t, state.Admit(context.Background(), 2))
	require.NoError(t, state.Admit(context.Background(), 2))
	require.Equal(t, 1, checkCalls)

	turn1 := service.BillingDecisionFromContext(state.Context(context.Background(), 1), 11)
	require.Equal(t, int64(22), turn1.PayerUserID)
	require.Equal(t, int64(33), turn1.ResearchGroupID)
	require.Equal(t, service.FundingSourceResearchGroup, turn1.FundingSource)

	turn2 := service.BillingDecisionFromContext(state.Context(context.Background(), 2), 11)
	require.Equal(t, int64(11), turn2.PayerUserID)
	require.Zero(t, turn2.ResearchGroupID)
	require.Equal(t, service.FundingSourceSelf, turn2.FundingSource)

	state.Forget(2)
	missing := service.BillingDecisionFromContext(state.Context(service.WithBillingDecision(context.Background(), initial), 2), 11)
	require.Equal(t, int64(11), missing.PayerUserID, "missing later-turn snapshots must not reuse turn 1 group funding")
	require.Equal(t, service.FundingSourceSelf, missing.FundingSource)
}
