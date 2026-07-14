package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	batchImageHoldRequestPrefix    = "batch_image_hold:"
	batchImageCaptureRequestPrefix = "batch_image_capture:"
	batchImageReleaseRequestPrefix = "batch_image_release:"
)

func BatchImageHoldRequestID(batchID string) string {
	return batchImageHoldRequestPrefix + strings.TrimSpace(batchID)
}

func BatchImageCaptureRequestID(batchID string) string {
	return batchImageCaptureRequestPrefix + strings.TrimSpace(batchID)
}

func BatchImageReleaseRequestID(batchID string) string {
	return batchImageReleaseRequestPrefix + strings.TrimSpace(batchID)
}

func buildBatchImageHoldCommand(job *BatchImageJob, requestID string, actualAmount float64, payloadHash string) (*BatchImageBalanceHoldCommand, error) {
	if job == nil {
		return nil, ErrBatchImageBillingHoldFailed
	}
	if job.APIKeyID == nil || *job.APIKeyID <= 0 {
		return nil, ErrBatchImageSettlementMissingAPIKeyID
	}
	holdAmount := job.EstimatedCost
	if job.HoldAmount != nil {
		holdAmount = *job.HoldAmount
	}
	if holdAmount < 0 {
		holdAmount = 0
	}
	if actualAmount < 0 {
		actualAmount = 0
	}
	decision := SelfBillingDecision(job.UserID)
	if job.PayerUserID != nil {
		decision.PayerUserID = *job.PayerUserID
	}
	if job.ResearchGroupID != nil {
		decision.ResearchGroupID = *job.ResearchGroupID
	}
	if job.ResearchGroupMemberID != nil {
		decision.ResearchGroupMemberID = *job.ResearchGroupMemberID
	}
	if job.FundingSource != nil {
		decision.FundingSource = *job.FundingSource
	}
	decision = decision.Normalize(job.UserID)
	return &BatchImageBalanceHoldCommand{
		RequestID:             requestID,
		APIKeyID:              *job.APIKeyID,
		UserID:                job.UserID,
		PayerUserID:           decision.PayerUserID,
		ResearchGroupID:       decision.ResearchGroupID,
		ResearchGroupMemberID: decision.ResearchGroupMemberID,
		FundingSource:         decision.FundingSource,
		BatchID:               job.BatchID,
		HoldAmount:            holdAmount,
		ActualAmount:          actualAmount,
		RequestPayloadHash:    strings.TrimSpace(payloadHash),
	}, nil
}

func reserveBatchImageBalanceHold(ctx context.Context, repo UsageBillingRepository, job *BatchImageJob, payloadHash string) error {
	if repo == nil {
		return ErrBatchImageBillingHoldFailed.WithCause(errors.New("batch image billing repository is not configured"))
	}
	cmd, err := buildBatchImageHoldCommand(job, BatchImageHoldRequestID(job.BatchID), 0, payloadHash)
	if err != nil {
		return err
	}
	if cmd.HoldAmount <= 0 {
		return nil
	}
	result, err := repo.ReserveBatchImageBalance(ctx, cmd)
	if err != nil {
		if errors.Is(err, ErrBatchImageInsufficientBalance) {
			return ErrBatchImageInsufficientBalance
		}
		if errors.Is(err, ErrResearchGroupAndPersonalBalanceInsufficient) {
			return ErrResearchGroupAndPersonalBalanceInsufficient
		}
		return ErrBatchImageBillingHoldFailed.WithCause(err)
	}
	if result != nil && result.Applied && result.PayerUserID > 0 {
		job.PayerUserID = optionalInt64Ptr(result.PayerUserID)
		job.ResearchGroupID = optionalInt64Ptr(result.ResearchGroupID)
		job.ResearchGroupMemberID = optionalInt64Ptr(result.ResearchGroupMemberID)
		job.FundingSource = optionalTrimmedStringPtr(result.FundingSource)
	}
	return nil
}

func captureBatchImageBalanceHold(ctx context.Context, repo UsageBillingRepository, job *BatchImageJob, actualAmount float64, payloadHash string) (*BatchImageBalanceHoldResult, error) {
	if repo == nil {
		return nil, ErrBatchImageSettlementBillingFailed.WithCause(errors.New("batch image billing repository is not configured"))
	}
	cmd, err := buildBatchImageHoldCommand(job, BatchImageCaptureRequestID(job.BatchID), actualAmount, payloadHash)
	if err != nil {
		return nil, err
	}
	result, err := repo.CaptureBatchImageBalance(ctx, cmd)
	if err != nil {
		return nil, ErrBatchImageSettlementBillingFailed.WithCause(err)
	}
	return result, nil
}

func releaseBatchImageBalanceHold(ctx context.Context, repo UsageBillingRepository, job *BatchImageJob, payloadHash string) error {
	if repo == nil || job == nil {
		return nil
	}
	cmd, err := buildBatchImageHoldCommand(job, BatchImageReleaseRequestID(job.BatchID), 0, payloadHash)
	if err != nil {
		return err
	}
	if cmd.HoldAmount <= 0 {
		return nil
	}
	if _, err := repo.ReleaseBatchImageBalance(ctx, cmd); err != nil {
		// 同一 release request id 出现指纹冲突，说明此前已有一次携带不同
		// payloadHash 的释放成功提交（资金已归还）。视为幂等成功，
		// 避免历史指纹不一致的 job 永远卡在释放失败的毒消息循环里。
		if errors.Is(err, ErrUsageBillingRequestConflict) {
			logger.L().Warn("batch_image.release_fingerprint_conflict_treated_as_released",
				zap.String("batch_id", job.BatchID),
			)
			return nil
		}
		return ErrBatchImageBillingHoldFailed.WithCause(err)
	}
	return nil
}

func invalidateBatchImageBillingCaches(ctx context.Context, authCache APIKeyAuthCacheInvalidator, billingCache *BillingCacheService, job *BatchImageJob) {
	if job == nil {
		return
	}
	payerUserID := job.UserID
	if job.PayerUserID != nil && *job.PayerUserID > 0 {
		payerUserID = *job.PayerUserID
	}
	if authCache != nil {
		if job.UserID > 0 {
			authCache.InvalidateAuthCacheByUserID(ctx, job.UserID)
		}
		if payerUserID > 0 && payerUserID != job.UserID {
			authCache.InvalidateAuthCacheByUserID(ctx, payerUserID)
		}
	}
	if billingCache == nil {
		return
	}
	if payerUserID > 0 {
		if err := billingCache.InvalidateUserBalance(ctx, payerUserID); err != nil {
			logger.L().Warn("batch_image.invalidate_payer_balance_failed", zap.Int64("payer_user_id", payerUserID), zap.Error(err))
		}
	}
	if job.ResearchGroupMemberID != nil && job.UserID > 0 {
		if err := billingCache.InvalidateResearchGroupFunding(ctx, job.UserID); err != nil {
			logger.L().Warn("batch_image.invalidate_research_group_funding_failed", zap.Int64("user_id", job.UserID), zap.Error(err))
		}
	}
}
