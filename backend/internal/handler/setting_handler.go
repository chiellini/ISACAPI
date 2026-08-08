package handler

import (
	"html"
	"net/http"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService           *service.SettingService
	notificationEmailService *service.NotificationEmailService
	pricingService           *service.PricingService
	version                  string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
	}
}

// SetNotificationEmailService attaches the public notification email service without
// changing the constructor signature used by existing tests.
func (h *SettingHandler) SetNotificationEmailService(notificationEmailService *service.NotificationEmailService) {
	h.notificationEmailService = notificationEmailService
}

// SetPricingService wires in dynamic model pricing for public pricing pages.
func (h *SettingHandler) SetPricingService(pricingService *service.PricingService) {
	h.pricingService = pricingService
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PublicSettings{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		ForceEmailOnThirdPartySignup:     settings.ForceEmailOnThirdPartySignup,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		PasskeyEnabled:                   settings.PasskeyEnabled,
		LoginAgreementEnabled:            settings.LoginAgreementEnabled,
		LoginAgreementMode:               settings.LoginAgreementMode,
		LoginAgreementUpdatedAt:          settings.LoginAgreementUpdatedAt,
		LoginAgreementRevision:           settings.LoginAgreementRevision,
		LoginAgreementDocuments:          publicLoginAgreementDocumentsToDTO(settings.LoginAgreementDocuments),
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		TencentCaptchaEnabled:            settings.TencentCaptchaEnabled,
		TencentCaptchaAppID:              settings.TencentCaptchaAppID,
		TencentCaptchaRegion:             settings.TencentCaptchaRegion,
		AliyunCaptchaEnabled:             settings.AliyunCaptchaEnabled,
		AliyunCaptchaSceneID:             settings.AliyunCaptchaSceneID,
		AliyunCaptchaPrefix:              settings.AliyunCaptchaPrefix,
		AliyunCaptchaRegion:              settings.AliyunCaptchaRegion,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		DocURL:                           settings.DocURL,
		HomeContent:                      settings.HomeContent,
		CompactHomeEnabled:               settings.CompactHomeEnabled,
		HideCcsImportButton:              settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:          settings.PurchaseSubscriptionURL,
		TableDefaultPageSize:             settings.TableDefaultPageSize,
		TablePageSizeOptions:             settings.TablePageSizeOptions,
		CustomMenuItems:                  dto.ParseUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  dto.ParseCustomEndpoints(settings.CustomEndpoints),
		DingTalkOAuthEnabled:             settings.DingTalkOAuthEnabled,
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		WeChatOAuthEnabled:               settings.WeChatOAuthEnabled,
		WeChatOAuthOpenEnabled:           settings.WeChatOAuthOpenEnabled,
		WeChatOAuthMPEnabled:             settings.WeChatOAuthMPEnabled,
		WeChatOAuthMobileEnabled:         settings.WeChatOAuthMobileEnabled,
		OIDCOAuthEnabled:                 settings.OIDCOAuthEnabled,
		OIDCOAuthProviderName:            settings.OIDCOAuthProviderName,
		GitHubOAuthEnabled:               settings.GitHubOAuthEnabled,
		GoogleOAuthEnabled:               settings.GoogleOAuthEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		PaymentEnabled:                   settings.PaymentEnabled,
		BalanceRechargeMultiplier:        settings.BalanceRechargeMultiplier,
		Version:                          h.version,
		ServerTimezone:                   timezone.Name(),
		ServerUTCOffset:                  timezone.UTCOffset(),
		BalanceLowNotifyEnabled:          settings.BalanceLowNotifyEnabled,
		AccountQuotaNotifyEnabled:        settings.AccountQuotaNotifyEnabled,
		BalanceLowNotifyThreshold:        settings.BalanceLowNotifyThreshold,
		BalanceLowNotifyRechargeURL:      settings.BalanceLowNotifyRechargeURL,

		ChannelMonitorEnabled:                settings.ChannelMonitorEnabled,
		ChannelMonitorDefaultIntervalSeconds: settings.ChannelMonitorDefaultIntervalSeconds,

		AvailableChannelsEnabled: settings.AvailableChannelsEnabled,

		ModelPlazaEnabled:     settings.ModelPlazaEnabled,
		ModelPlazaRequireAuth: settings.ModelPlazaRequireAuth,

		PublicStatusEnabled: settings.PublicStatusEnabled,

		AffiliateEnabled: settings.AffiliateEnabled,

		RiskControlEnabled: settings.RiskControlEnabled,

		AllowUserViewErrorRequests: settings.AllowUserViewErrorRequests,
	})
}

func publicModelFamily(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return "gpt"
	case "anthropic":
		return "claude"
	default:
		return provider
	}
}

func publicModelFamilySortFamily(family string) int {
	switch family {
	case "gpt":
		return 0
	case "claude":
		return 1
	default:
		return 2
	}
}

// GetPublicPricingModels returns public model pricing rows consumed by the pricing page.
// GET /api/v1/settings/public/pricing-models
func (h *SettingHandler) GetPublicPricingModels(c *gin.Context) {
	if h.pricingService == nil {
		response.InternalError(c, "pricing service is not configured")
		return
	}

	const usdcPerTokenToPerMillion = 1_000_000.0

	providers := []string{"openai", "anthropic"}
	models := make([]dto.PublicModelPricing, 0)
	seen := make(map[string]struct{}, 128)

	for _, provider := range providers {
		for _, model := range h.pricingService.ListModelNamesByProvider(provider) {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			if _, ok := seen[model]; ok {
				continue
			}

			pricing := h.pricingService.GetModelPricing(model)
			if pricing == nil || pricing.TokenPricingAbsent {
				continue
			}

			input := pricing.InputCostPerToken * usdcPerTokenToPerMillion
			output := pricing.OutputCostPerToken * usdcPerTokenToPerMillion
			cacheRead := pricing.CacheReadInputTokenCost * usdcPerTokenToPerMillion
			if input <= 0 && output <= 0 && cacheRead <= 0 {
				continue
			}

			models = append(models, dto.PublicModelPricing{
				ID:                              model,
				Name:                            model,
				Family:                          publicModelFamily(provider),
				BenchmarkInputUsdPerMillion:     input,
				BenchmarkOutputUsdPerMillion:    output,
				BenchmarkCacheReadUsdPerMillion: cacheRead,
			})
			seen[model] = struct{}{}
		}
	}

	sort.SliceStable(models, func(i, j int) bool {
		if models[i].Family == models[j].Family {
			return models[i].ID < models[j].ID
		}
		return publicModelFamilySortFamily(models[i].Family) < publicModelFamilySortFamily(models[j].Family)
	})

	response.Success(c, dto.PublicModelPricingListResponse{Models: models})
}

// UnsubscribeNotificationEmail handles optional notification email opt-outs.
// GET /api/v1/settings/email-unsubscribe?token=...
func (h *SettingHandler) UnsubscribeNotificationEmail(c *gin.Context) {
	if h.notificationEmailService == nil {
		response.InternalError(c, "notification email service is not configured")
		return
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		response.BadRequest(c, "token is required")
		return
	}
	result, err := h.notificationEmailService.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	body := "<!doctype html><html><head><meta charset=\"utf-8\"><title>Unsubscribed</title></head><body style=\"font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;padding:32px;\"><h1>Unsubscribed</h1><p>You have unsubscribed <strong>" + html.EscapeString(result.Email) + "</strong> from <strong>" + html.EscapeString(result.Event) + "</strong> emails.</p></body></html>"
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(body))
}

func publicLoginAgreementDocumentsToDTO(items []service.LoginAgreementDocument) []dto.LoginAgreementDocument {
	result := make([]dto.LoginAgreementDocument, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LoginAgreementDocument{
			ID:        item.ID,
			Title:     item.Title,
			ContentMD: item.ContentMD,
		})
	}
	return result
}
