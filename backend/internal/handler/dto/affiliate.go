package dto

type AffiliatePaymentAccountRequest struct {
	Type string `json:"type" binding:"required"`
	AccountName string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	BankName string `json:"bank_name"`
	USDTNetwork string `json:"usdt_network"`
	WalletAddress string `json:"wallet_address"`
	IsDefault bool `json:"is_default"`
}
type CreateAffiliateWithdrawalRequest struct { PaymentAccountID int64 `json:"payment_account_id" binding:"required"`; Amount float64 `json:"amount" binding:"required"` }
type UpdateAffiliateAgentStatusRequest struct { Status string `json:"status" binding:"required"` }
type RejectAffiliateWithdrawalRequest struct { Reason string `json:"reason" binding:"required"` }
type MarkAffiliateWithdrawalPaidRequest struct { ActualCurrency string `json:"actual_currency" binding:"required"`; ActualAmount float64 `json:"actual_amount" binding:"required"`; ExchangeRate float64 `json:"exchange_rate" binding:"required"`; ExternalReference string `json:"external_reference" binding:"required"` }
