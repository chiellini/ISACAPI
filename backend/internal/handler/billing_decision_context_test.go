package handler

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageRecordContextCopiesBillingDecision(t *testing.T) {
	parent := service.WithBillingDecision(context.Background(), &service.BillingDecision{
		CallerUserID:          11,
		PayerUserID:           22,
		ResearchGroupID:       33,
		ResearchGroupMemberID: 44,
		FundingSource:         service.FundingSourceResearchGroup,
	})

	workerCtx := usageRecordContext(parent, context.Background())
	decision := service.BillingDecisionFromContext(workerCtx, 11)
	require.Equal(t, int64(11), decision.CallerUserID)
	require.Equal(t, int64(22), decision.PayerUserID)
	require.Equal(t, int64(33), decision.ResearchGroupID)
	require.Equal(t, int64(44), decision.ResearchGroupMemberID)
	require.Equal(t, service.FundingSourceResearchGroup, decision.FundingSource)
}

func TestCyberPolicyRecordContextCopiesBillingDecision(t *testing.T) {
	parent := service.WithBillingDecision(context.Background(), &service.BillingDecision{
		CallerUserID:          11,
		PayerUserID:           22,
		ResearchGroupID:       33,
		ResearchGroupMemberID: 44,
		FundingSource:         service.FundingSourceResearchGroup,
	})

	workerCtx, cancel := cyberPolicyRecordContext(parent)
	defer cancel()
	decision := service.BillingDecisionFromContext(workerCtx, 11)
	require.Equal(t, int64(22), decision.PayerUserID)
	require.Equal(t, int64(33), decision.ResearchGroupID)
	require.Equal(t, int64(44), decision.ResearchGroupMemberID)
	require.Equal(t, service.FundingSourceResearchGroup, decision.FundingSource)
}
