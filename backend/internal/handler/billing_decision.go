package handler

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func checkAndAttachBillingDecision(
	c *gin.Context,
	billing *service.BillingCacheService,
	user *service.User,
	apiKey *service.APIKey,
	group *service.Group,
	subscription *service.UserSubscription,
	platform string,
) error {
	_, err := checkAndAttachBillingDecisionToContext(c, c.Request.Context(), billing, user, apiKey, group, subscription, platform)
	return err
}

func checkAndAttachBillingDecisionToContext(
	c *gin.Context,
	ctx context.Context,
	billing *service.BillingCacheService,
	user *service.User,
	apiKey *service.APIKey,
	group *service.Group,
	subscription *service.UserSubscription,
	platform string,
) (context.Context, error) {
	decision, err := billing.CheckBillingEligibility(ctx, user, apiKey, group, subscription, platform)
	if err != nil {
		return ctx, err
	}
	requestCtx := service.WithBillingDecision(c.Request.Context(), decision)
	c.Request = c.Request.WithContext(requestCtx)
	return service.WithBillingDecision(ctx, decision), nil
}
