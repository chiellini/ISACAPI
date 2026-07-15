package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSameAffiliateWithdrawalRequestRejectsReusedKeyWithDifferentPayload(t *testing.T) {
	accountID := int64(7)
	existing := &service.AffiliateWithdrawal{Amount: 10, PaymentAccountID: &accountID}
	require.True(t, sameAffiliateWithdrawalRequest(existing, 7, 10))
	require.False(t, sameAffiliateWithdrawalRequest(existing, 8, 10))
	require.False(t, sameAffiliateWithdrawalRequest(existing, 7, 11))
}

func TestSameAffiliateAgentStatusRequestChecksRequestedStatus(t *testing.T) {
	require.True(t, sameAffiliateAgentStatusRequest(9, "active", 9, "active"))
	require.False(t, sameAffiliateAgentStatusRequest(9, "active", 9, "suspended"))
	require.False(t, sameAffiliateAgentStatusRequest(8, "active", 9, "active"))
}
