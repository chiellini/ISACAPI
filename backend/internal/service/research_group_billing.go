package service

import (
	"context"
	"errors"
	"net/http"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// BillingDecision is the immutable payer snapshot captured during request admission.
// Settlement must use this snapshot even if the member relationship changes in-flight.
type BillingDecision struct {
	CallerUserID          int64
	PayerUserID           int64
	ResearchGroupID       int64
	ResearchGroupMemberID int64
	FundingSource         string
}

func SelfBillingDecision(userID int64) *BillingDecision {
	return &BillingDecision{
		CallerUserID:  userID,
		PayerUserID:   userID,
		FundingSource: FundingSourceSelf,
	}
}

func (d *BillingDecision) Normalize(callerUserID int64) *BillingDecision {
	if d == nil {
		return SelfBillingDecision(callerUserID)
	}
	copy := *d
	if copy.CallerUserID <= 0 {
		copy.CallerUserID = callerUserID
	}
	if copy.PayerUserID <= 0 {
		copy.PayerUserID = copy.CallerUserID
	}
	if copy.FundingSource != FundingSourceResearchGroup || copy.ResearchGroupID <= 0 || copy.ResearchGroupMemberID <= 0 {
		copy.PayerUserID = copy.CallerUserID
		copy.ResearchGroupID = 0
		copy.ResearchGroupMemberID = 0
		copy.FundingSource = FundingSourceSelf
	}
	return &copy
}

func (d *BillingDecision) IsResearchGroupFunded() bool {
	return d != nil && d.FundingSource == FundingSourceResearchGroup &&
		d.ResearchGroupID > 0 && d.ResearchGroupMemberID > 0 && d.PayerUserID > 0
}

type billingDecisionContextKey struct{}

func WithBillingDecision(ctx context.Context, decision *BillingDecision) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if decision == nil {
		return ctx
	}
	copy := *decision
	return context.WithValue(ctx, billingDecisionContextKey{}, &copy)
}

func BillingDecisionFromContext(ctx context.Context, callerUserID int64) *BillingDecision {
	if ctx != nil {
		if decision, ok := ctx.Value(billingDecisionContextKey{}).(*BillingDecision); ok && decision != nil {
			return decision.Normalize(callerUserID)
		}
	}
	return SelfBillingDecision(callerUserID)
}

var ErrResearchGroupAndPersonalBalanceInsufficient = infraerrors.New(
	http.StatusPaymentRequired,
	"RESEARCH_GROUP_AND_PERSONAL_BALANCE_INSUFFICIENT",
	"Research group funding and personal balance are both insufficient.",
)

func isInsufficientBalanceError(err error) bool {
	return errors.Is(err, ErrInsufficientBalance)
}

type ResearchGroupFundingCache interface {
	GetResearchGroupFunding(ctx context.Context, userID int64) (*ResearchGroupFundingContext, bool, error)
	SetResearchGroupFunding(ctx context.Context, userID int64, funding *ResearchGroupFundingContext) error
	InvalidateResearchGroupFunding(ctx context.Context, userID int64) error
}

type ResearchGroupFundingCacheInvalidator interface {
	InvalidateResearchGroupFunding(ctx context.Context, userID int64) error
}
