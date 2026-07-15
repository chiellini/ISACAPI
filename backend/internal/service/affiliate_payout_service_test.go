//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAffiliatePaymentAccount(t *testing.T) {
	account, err := normalizeAffiliatePaymentAccount(AffiliatePaymentAccountInput{Type: " bank_card ", AccountName: " Alice ", AccountNumber: "12345678", BankName: " Example Bank "})
	require.NoError(t, err)
	require.Equal(t, AffiliatePaymentAccountBankCard, account.Type)
	require.Equal(t, "Example Bank · A*** · ***5678", affiliatePaymentAccountSummary(account))
	_, err = normalizeAffiliatePaymentAccount(AffiliatePaymentAccountInput{Type: AffiliatePaymentAccountUSDT, USDTNetwork: "trc20"})
	require.Error(t, err)
}

func TestAffiliateAdminWritesRequireIdempotencyKey(t *testing.T) {
	svc := &AffiliatePayoutService{}
	require.ErrorIs(t, svc.AdminSetAgentStatus(context.Background(), 1, AffiliateAgentStatusActive, 2, ""), ErrAffiliateIdempotencyKeyRequired)
	_, err := svc.AdminApproveWithdrawal(context.Background(), 1, 2, "")
	require.ErrorIs(t, err, ErrAffiliateIdempotencyKeyRequired)
}
