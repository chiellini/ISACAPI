package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	AffiliateAgentStatusInactive  = "inactive"
	AffiliateAgentStatusActive    = "active"
	AffiliateAgentStatusSuspended = "suspended"

	AffiliatePaymentAccountAlipay   = "alipay"
	AffiliatePaymentAccountBankCard = "bank_card"
	AffiliatePaymentAccountUSDT     = "usdt"

	AffiliateWithdrawalSubmitted = "submitted"
	AffiliateWithdrawalApproved  = "approved"
	AffiliateWithdrawalPaid      = "paid"
	AffiliateWithdrawalRejected  = "rejected"
	AffiliateWithdrawalCanceled  = "canceled"

	AffiliateMinimumWithdrawalDefault = 10.0
	SettingKeyAffiliateMinimumWithdrawal = "affiliate_minimum_withdrawal"
)

var (
	ErrAffiliateAgentInactive = infraerrors.Forbidden("AFFILIATE_AGENT_INACTIVE", "affiliate agent access is not active")
	ErrAffiliatePayoutEncryptionKeyRequired = infraerrors.ServiceUnavailable("AFFILIATE_PAYOUT_ENCRYPTION_KEY_REQUIRED", "a fixed TOTP encryption key is required for affiliate payout accounts")
	ErrAffiliatePaymentAccountNotFound = infraerrors.NotFound("AFFILIATE_PAYMENT_ACCOUNT_NOT_FOUND", "affiliate payment account not found")
	ErrAffiliateWithdrawalNotFound = infraerrors.NotFound("AFFILIATE_WITHDRAWAL_NOT_FOUND", "affiliate withdrawal not found")
	ErrAffiliateWithdrawalTooSmall = infraerrors.BadRequest("AFFILIATE_WITHDRAWAL_BELOW_MINIMUM", "withdrawal amount is below the minimum")
	ErrAffiliateWithdrawalInsufficient = infraerrors.BadRequest("AFFILIATE_WITHDRAWAL_INSUFFICIENT", "insufficient available affiliate commission")
	ErrAffiliateWithdrawalDebt = infraerrors.Conflict("AFFILIATE_WITHDRAWAL_DEBT", "affiliate debt must be cleared before withdrawal")
	ErrAffiliateWithdrawalState = infraerrors.Conflict("AFFILIATE_WITHDRAWAL_INVALID_STATE", "withdrawal is not in the required state")
	ErrAffiliateIdempotencyKeyRequired = infraerrors.BadRequest("IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required")
	ErrAffiliateIdempotencyKeyConflict = infraerrors.Conflict("IDEMPOTENCY_KEY_CONFLICT", "Idempotency-Key was already used with a different request")
	ErrAffiliateBalanceTransferDisabled = infraerrors.BadRequest("AFFILIATE_BALANCE_TRANSFER_DISABLED", "affiliate commission can no longer be transferred to balance")
)

type AffiliatePaymentAccountInput struct {
	Type          string `json:"type"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	USDTNetwork   string `json:"usdt_network,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty"`
	IsDefault     bool   `json:"is_default"`
}

type AffiliatePaymentAccount struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"-"`
	Type          string    `json:"type"`
	MaskedSummary string    `json:"masked_summary"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AffiliateWithdrawal struct {
	ID                    int64      `json:"id"`
	UserID                int64      `json:"user_id,omitempty"`
	UserEmail             string     `json:"user_email,omitempty"`
	Username              string     `json:"username,omitempty"`
	Amount                float64    `json:"amount"`
	Status                string     `json:"status"`
	StatusLabel           string     `json:"status_label"`
	PaymentAccountType    string     `json:"payment_account_type"`
	PaymentAccountID      *int64     `json:"payment_account_id,omitempty"`
	PaymentAccountSummary string     `json:"payment_account_summary"`
	PaymentDetails        *AffiliatePaymentAccountInput `json:"payment_details,omitempty"`
	PaymentDetailsEncrypted string   `json:"-"`
	SubmittedAt           time.Time  `json:"submitted_at"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	RejectReason          *string    `json:"reject_reason,omitempty"`
	ActualCurrency        *string    `json:"actual_currency,omitempty"`
	ActualAmount          *float64   `json:"actual_amount,omitempty"`
	ExchangeRate          *float64   `json:"exchange_rate,omitempty"`
	ExternalReference     *string    `json:"external_reference,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type AffiliateAgentAdminEntry struct {
	UserID              int64   `json:"user_id"`
	Email               string  `json:"email"`
	Username            string  `json:"username"`
	Status              string  `json:"status"`
	AffCode             string  `json:"aff_code"`
	RebateRatePercent   float64 `json:"rebate_rate_percent"`
	InvitedCount        int     `json:"invited_count"`
	AvailableCommission float64 `json:"available_commission"`
	FrozenCommission    float64 `json:"frozen_commission"`
	WithdrawalReserved  float64 `json:"withdrawal_reserved"`
	Debt                float64 `json:"debt"`
}

type AffiliatePayoutListFilter struct {
	Search string
	Status string
	Page int
	PageSize int
}

type AffiliateMarkPaidInput struct {
	ActualCurrency    string  `json:"actual_currency"`
	ActualAmount      float64 `json:"actual_amount"`
	ExchangeRate      float64 `json:"exchange_rate"`
	ExternalReference string  `json:"external_reference"`
}

type AffiliatePayoutRepository interface {
	GetAgentStatus(ctx context.Context, userID int64) (string, error)
	ListPaymentAccounts(ctx context.Context, userID int64) ([]AffiliatePaymentAccount, error)
	CreatePaymentAccount(ctx context.Context, userID int64, accountType, encrypted, summary string, isDefault bool) (*AffiliatePaymentAccount, error)
	UpdatePaymentAccount(ctx context.Context, userID, accountID int64, accountType, encrypted, summary string, isDefault bool) (*AffiliatePaymentAccount, error)
	DeletePaymentAccount(ctx context.Context, userID, accountID int64) error
	CreateWithdrawal(ctx context.Context, userID, paymentAccountID int64, amount, minimum float64, idempotencyKey string) (*AffiliateWithdrawal, error)
	ListUserWithdrawals(ctx context.Context, userID int64, page, pageSize int) ([]AffiliateWithdrawal, int64, error)
	CancelWithdrawal(ctx context.Context, userID, withdrawalID int64) (*AffiliateWithdrawal, error)
	ListAgents(ctx context.Context, filter AffiliatePayoutListFilter) ([]AffiliateAgentAdminEntry, int64, error)
	GetAgent(ctx context.Context, userID int64) (*AffiliateAgentAdminEntry, error)
	SetAgentStatus(ctx context.Context, userID int64, status string, operatorUserID int64, idempotencyKey string) error
	ListWithdrawals(ctx context.Context, filter AffiliatePayoutListFilter) ([]AffiliateWithdrawal, int64, error)
	GetWithdrawal(ctx context.Context, withdrawalID int64) (*AffiliateWithdrawal, error)
	ApproveWithdrawal(ctx context.Context, withdrawalID, operatorUserID int64, idempotencyKey string) (*AffiliateWithdrawal, error)
	RejectWithdrawal(ctx context.Context, withdrawalID, operatorUserID int64, reason, idempotencyKey string) (*AffiliateWithdrawal, error)
	MarkWithdrawalPaid(ctx context.Context, withdrawalID, operatorUserID int64, input AffiliateMarkPaidInput, idempotencyKey string) (*AffiliateWithdrawal, error)
}

type AffiliatePayoutService struct {
	repo AffiliatePayoutRepository
	settings *SettingService
	encryptor SecretEncryptor
}

func NewAffiliatePayoutService(repo AffiliatePayoutRepository, settings *SettingService, encryptor SecretEncryptor) *AffiliatePayoutService {
	return &AffiliatePayoutService{repo: repo, settings: settings, encryptor: encryptor}
}

func (s *AffiliatePayoutService) MinimumWithdrawal(ctx context.Context) float64 {
	if s == nil || s.settings == nil || s.settings.settingRepo == nil { return AffiliateMinimumWithdrawalDefault }
	raw, err := s.settings.settingRepo.GetValue(ctx, SettingKeyAffiliateMinimumWithdrawal)
	if err != nil { return AffiliateMinimumWithdrawalDefault }
	var v float64
	if _, err := fmtSscan(strings.TrimSpace(raw), &v); err != nil || v < AffiliateMinimumWithdrawalDefault || math.IsNaN(v) || math.IsInf(v, 0) {
		return AffiliateMinimumWithdrawalDefault
	}
	return roundTo(v, 8)
}

// fmtSscan is a small variable to keep parsing testable without exposing settings internals.
var fmtSscan = func(raw string, dst *float64) (int, error) { return fmt.Sscan(raw, dst) }

func (s *AffiliatePayoutService) requireReady() error {
	if s == nil || s.repo == nil || s.encryptor == nil { return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate payout service unavailable") }
	if s.settings == nil || s.settings.cfg == nil || !s.settings.IsTotpEncryptionKeyConfigured() { return ErrAffiliatePayoutEncryptionKeyRequired }
	return nil
}

func (s *AffiliatePayoutService) ListPaymentAccounts(ctx context.Context, userID int64) ([]AffiliatePaymentAccount, error) {
	if err := s.requireReady(); err != nil { return nil, err }
	return s.repo.ListPaymentAccounts(ctx, userID)
}

func (s *AffiliatePayoutService) SavePaymentAccount(ctx context.Context, userID, accountID int64, in AffiliatePaymentAccountInput) (*AffiliatePaymentAccount, error) {
	if err := s.requireReady(); err != nil { return nil, err }
	in, err := normalizeAffiliatePaymentAccount(in)
	if err != nil { return nil, err }
	if status, err := s.repo.GetAgentStatus(ctx, userID); err != nil { return nil, err } else if status != AffiliateAgentStatusActive { return nil, ErrAffiliateAgentInactive }
	raw, _ := json.Marshal(in)
	encrypted, err := s.encryptor.Encrypt(string(raw))
	if err != nil { return nil, infraerrors.InternalServer("AFFILIATE_PAYOUT_ENCRYPT_FAILED", "failed to encrypt payout account") }
	summary := affiliatePaymentAccountSummary(in)
	if accountID <= 0 { return s.repo.CreatePaymentAccount(ctx, userID, in.Type, encrypted, summary, in.IsDefault) }
	return s.repo.UpdatePaymentAccount(ctx, userID, accountID, in.Type, encrypted, summary, in.IsDefault)
}

func (s *AffiliatePayoutService) DeletePaymentAccount(ctx context.Context, userID, accountID int64) error {
	if err := s.requireReady(); err != nil { return err }
	return s.repo.DeletePaymentAccount(ctx, userID, accountID)
}

func (s *AffiliatePayoutService) CreateWithdrawal(ctx context.Context, userID, accountID int64, amount float64, idempotencyKey string) (*AffiliateWithdrawal, error) {
	if err := s.requireReady(); err != nil { return nil, err }
	if strings.TrimSpace(idempotencyKey) == "" { return nil, ErrAffiliateIdempotencyKeyRequired }
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) { return nil, ErrAffiliateWithdrawalTooSmall }
	return s.repo.CreateWithdrawal(ctx, userID, accountID, roundTo(amount, 8), s.MinimumWithdrawal(ctx), strings.TrimSpace(idempotencyKey))
}

func (s *AffiliatePayoutService) ListUserWithdrawals(ctx context.Context, userID int64, page, pageSize int) ([]AffiliateWithdrawal, int64, error) {
	if s == nil || s.repo == nil { return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate payout service unavailable") }
	return s.repo.ListUserWithdrawals(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *AffiliatePayoutService) CancelWithdrawal(ctx context.Context, userID, withdrawalID int64) (*AffiliateWithdrawal, error) {
	if s == nil || s.repo == nil { return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate payout service unavailable") }
	return s.repo.CancelWithdrawal(ctx, userID, withdrawalID)
}

func (s *AffiliatePayoutService) AdminListAgents(ctx context.Context, filter AffiliatePayoutListFilter) ([]AffiliateAgentAdminEntry, int64, error) {
	filter.Page, filter.PageSize = normalizePage(filter.Page), normalizePageSize(filter.PageSize)
	return s.repo.ListAgents(ctx, filter)
}
func (s *AffiliatePayoutService) AdminGetAgent(ctx context.Context, userID int64) (*AffiliateAgentAdminEntry,error){ return s.repo.GetAgent(ctx,userID) }

func (s *AffiliatePayoutService) AdminSetAgentStatus(ctx context.Context, userID int64, status string, operatorID int64, key string) error {
	if strings.TrimSpace(key) == "" { return ErrAffiliateIdempotencyKeyRequired }
	status = strings.ToLower(strings.TrimSpace(status))
	if status != AffiliateAgentStatusActive && status != AffiliateAgentStatusSuspended && status != AffiliateAgentStatusInactive {
		return infraerrors.BadRequest("AFFILIATE_AGENT_STATUS_INVALID", "invalid affiliate agent status")
	}
	return s.repo.SetAgentStatus(ctx, userID, status, operatorID, strings.TrimSpace(key))
}

func (s *AffiliatePayoutService) AdminListWithdrawals(ctx context.Context, filter AffiliatePayoutListFilter) ([]AffiliateWithdrawal, int64, error) {
	filter.Page, filter.PageSize = normalizePage(filter.Page), normalizePageSize(filter.PageSize)
	return s.repo.ListWithdrawals(ctx, filter)
}
func (s *AffiliatePayoutService) AdminGetWithdrawal(ctx context.Context, id int64) (*AffiliateWithdrawal, error) {
	if err:=s.requireReady();err!=nil{return nil,err};item,err:=s.repo.GetWithdrawal(ctx,id);if err!=nil{return nil,err}
	plain,err:=s.encryptor.Decrypt(item.PaymentDetailsEncrypted);if err!=nil{return nil,infraerrors.InternalServer("AFFILIATE_PAYOUT_DECRYPT_FAILED","failed to decrypt payout account")}
	var details AffiliatePaymentAccountInput;if err=json.Unmarshal([]byte(plain),&details);err!=nil{return nil,infraerrors.InternalServer("AFFILIATE_PAYOUT_DECRYPT_FAILED","failed to decode payout account")};item.PaymentDetails=&details;return item,nil
}
func (s *AffiliatePayoutService) AdminApproveWithdrawal(ctx context.Context, id, operatorID int64, key string) (*AffiliateWithdrawal, error) {
	if strings.TrimSpace(key) == "" { return nil, ErrAffiliateIdempotencyKeyRequired }
	return s.repo.ApproveWithdrawal(ctx, id, operatorID, strings.TrimSpace(key))
}
func (s *AffiliatePayoutService) AdminRejectWithdrawal(ctx context.Context, id, operatorID int64, reason, key string) (*AffiliateWithdrawal, error) {
	if strings.TrimSpace(key) == "" { return nil, ErrAffiliateIdempotencyKeyRequired }
	if strings.TrimSpace(reason) == "" { return nil, infraerrors.BadRequest("REJECT_REASON_REQUIRED", "reject reason is required") }
	return s.repo.RejectWithdrawal(ctx, id, operatorID, strings.TrimSpace(reason), strings.TrimSpace(key))
}
func (s *AffiliatePayoutService) AdminMarkWithdrawalPaid(ctx context.Context, id, operatorID int64, in AffiliateMarkPaidInput, key string) (*AffiliateWithdrawal, error) {
	if strings.TrimSpace(key) == "" { return nil, ErrAffiliateIdempotencyKeyRequired }
	in.ActualCurrency, in.ExternalReference = strings.ToUpper(strings.TrimSpace(in.ActualCurrency)), strings.TrimSpace(in.ExternalReference)
	if in.ActualCurrency == "" || in.ExternalReference == "" || in.ActualAmount <= 0 || in.ExchangeRate <= 0 || math.IsNaN(in.ActualAmount) || math.IsNaN(in.ExchangeRate) || math.IsInf(in.ActualAmount, 0) || math.IsInf(in.ExchangeRate, 0) {
		return nil, infraerrors.BadRequest("AFFILIATE_PAID_DETAILS_INVALID", "actual currency, amount, exchange rate and external reference are required")
	}
	return s.repo.MarkWithdrawalPaid(ctx, id, operatorID, in, strings.TrimSpace(key))
}

func normalizeAffiliatePaymentAccount(in AffiliatePaymentAccountInput) (AffiliatePaymentAccountInput, error) {
	in.Type = strings.ToLower(strings.TrimSpace(in.Type)); in.AccountName = strings.TrimSpace(in.AccountName)
	in.AccountNumber = strings.TrimSpace(in.AccountNumber); in.BankName = strings.TrimSpace(in.BankName)
	in.USDTNetwork = strings.ToUpper(strings.TrimSpace(in.USDTNetwork)); in.WalletAddress = strings.TrimSpace(in.WalletAddress)
	switch in.Type {
	case AffiliatePaymentAccountAlipay:
		if in.AccountName == "" || in.AccountNumber == "" { return in, infraerrors.BadRequest("AFFILIATE_PAYMENT_ACCOUNT_INVALID", "Alipay name and account are required") }
	case AffiliatePaymentAccountBankCard:
		if in.AccountName == "" || in.AccountNumber == "" || in.BankName == "" { return in, infraerrors.BadRequest("AFFILIATE_PAYMENT_ACCOUNT_INVALID", "bank name, account name and card number are required") }
	case AffiliatePaymentAccountUSDT:
		if in.USDTNetwork == "" || in.WalletAddress == "" { return in, infraerrors.BadRequest("AFFILIATE_PAYMENT_ACCOUNT_INVALID", "USDT network and wallet address are required") }
	default:
		return in, infraerrors.BadRequest("AFFILIATE_PAYMENT_ACCOUNT_TYPE_INVALID", "payment account type must be alipay, bank_card or usdt")
	}
	return in, nil
}

func affiliatePaymentAccountSummary(in AffiliatePaymentAccountInput) string {
	switch in.Type {
	case AffiliatePaymentAccountAlipay: return "支付宝 · " + maskName(in.AccountName) + " · " + maskTail(in.AccountNumber)
	case AffiliatePaymentAccountBankCard: return in.BankName + " · " + maskName(in.AccountName) + " · " + maskTail(in.AccountNumber)
	default: return "USDT-" + in.USDTNetwork + " · " + maskTail(in.WalletAddress)
	}
}
func maskName(v string) string { r, _ := utf8.DecodeRuneInString(v); if r == utf8.RuneError || v == "" { return "***" }; return string(r)+"***" }
func maskTail(v string) string { r := []rune(v); if len(r) <= 4 { return "***"+string(r) }; return "***"+string(r[len(r)-4:]) }
func normalizePage(v int) int { if v < 1 { return 1 }; return v }
func normalizePageSize(v int) int { if v < 1 { return 20 }; if v > 100 { return 100 }; return v }
func affiliateWithdrawalStatusLabel(status string) string {
	switch status { case AffiliateWithdrawalSubmitted: return "待审核"; case AffiliateWithdrawalApproved: return "待转账"; case AffiliateWithdrawalPaid: return "已经转账"; case AffiliateWithdrawalRejected: return "已拒绝"; case AffiliateWithdrawalCanceled: return "已取消"; default: return status }
}
