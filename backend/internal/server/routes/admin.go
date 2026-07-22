// Package routes provides HTTP route registration and handlers.
package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 注册管理员路由
func RegisterAdminRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	adminAuth middleware.AdminAuthMiddleware,
	auditLog middleware.AuditLogMiddleware,
	stepUpAuth middleware.StepUpAuthMiddleware,
	settingService *service.SettingService,
) {
	admin := v1.Group("/admin")
	admin.Use(gin.HandlerFunc(adminAuth))
	// 审计中间件挂在认证之后：所有管理面变更类操作 + 敏感读取入审计日志
	admin.Use(gin.HandlerFunc(auditLog))
	admin.Use(middleware.AdminComplianceGuard(settingService))
	{
		// 部署与运营合规确认
		registerAdminComplianceRoutes(admin, h)

		// 仪表盘
		registerDashboardRoutes(admin, h)

		// 用户管理
		registerUserManagementRoutes(admin, h)

		// 分组管理
		registerGroupRoutes(admin, h)

		registerResearchGroupRoutes(admin, h)

		// 账号管理
		registerAccountRoutes(admin, h, stepUpAuth)
		registerAdminProviderRoutes(admin, h)

		// 公告管理
		registerAnnouncementRoutes(admin, h)

		// OpenAI OAuth
		registerOpenAIOAuthRoutes(admin, h)

		// Gemini OAuth
		registerGeminiOAuthRoutes(admin, h)

		// Antigravity OAuth
		registerAntigravityOAuthRoutes(admin, h)

		// Grok OAuth
		registerGrokOAuthRoutes(admin, h)

		// 代理管理
		registerProxyRoutes(admin, h, stepUpAuth)

		// 卡密管理
		registerRedeemCodeRoutes(admin, h)

		// 优惠码管理
		registerPromoCodeRoutes(admin, h)

		// 系统设置
		registerSettingsRoutes(admin, h)

		// 数据管理
		registerDataManagementRoutes(admin, h, stepUpAuth)

		// 数据库备份恢复
		registerBackupRoutes(admin, h, stepUpAuth)

		// 运维监控（Ops）
		registerOpsRoutes(admin, h)

		// 系统管理
		registerSystemRoutes(admin, h)

		// 订阅管理
		registerSubscriptionRoutes(admin, h)

		// 使用记录管理
		registerUsageRoutes(admin, h)

		// 用户属性管理
		registerUserAttributeRoutes(admin, h)

		// 错误透传规则管理
		registerErrorPassthroughRoutes(admin, h)

		// TLS 指纹模板管理
		registerTLSFingerprintProfileRoutes(admin, h)

		// API Key 管理
		registerAdminAPIKeyRoutes(admin, h)

		// 定时测试计划
		registerScheduledTestRoutes(admin, h)

		// 渠道管理
		registerChannelRoutes(admin, h)

		// 渠道监控
		registerChannelMonitorRoutes(admin, h)

		// 风控中心
		registerContentModerationRoutes(admin, h)

		// 独立提示词输入审计
		registerPromptAuditRoutes(admin, h)

		// 邀请返利（专属用户管理）
		registerAffiliateRoutes(admin, h)

		// 对话存档（仅超级管理员查看/删除）
		registerConversationRoutes(admin, h)

		// 操作审计日志
		registerAuditLogRoutes(admin, h, stepUpAuth)
	}
}

func registerPromptAuditRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	promptAudit := admin.Group("/prompt-audit")
	{
		promptAudit.GET("/config", h.Admin.PromptAudit.GetConfig)
		promptAudit.PUT("/config", h.Admin.PromptAudit.UpdateConfig)
		promptAudit.POST("/endpoints/probe", h.Admin.PromptAudit.ProbeEndpoint)
		promptAudit.GET("/runtime", h.Admin.PromptAudit.GetRuntime)
		promptAudit.GET("/events", h.Admin.PromptAudit.ListEvents)
		promptAudit.GET("/events/:id", h.Admin.PromptAudit.GetEvent)
		promptAudit.DELETE("/events/:id", h.Admin.PromptAudit.DeleteEvent)
		promptAudit.POST("/events/batch-delete", h.Admin.PromptAudit.BatchDelete)
		promptAudit.POST("/events/delete-preview", h.Admin.PromptAudit.DeletePreview)
		promptAudit.POST("/events/delete-by-filter", h.Admin.PromptAudit.DeleteByFilter)
	}
}

func registerAuditLogRoutes(admin *gin.RouterGroup, h *handler.Handlers, _ middleware.StepUpAuthMiddleware) {
	auditLogs := admin.Group("/audit-logs")
	{
		auditLogs.GET("", h.Admin.AuditLog.List)
		auditLogs.GET("/:id", h.Admin.AuditLog.Get)
		// 清空需现场 TOTP 校验（在 handler 内强制），不复用 step-up sudo 窗口
		auditLogs.POST("/clear", h.Admin.AuditLog.Clear)
	}
}

func superAdminOnly(group *gin.RouterGroup) *gin.RouterGroup {
	protected := group.Group("")
	protected.Use(middleware.RequireSuperAdmin())
	return protected
}

func registerConversationRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	conv := admin.Group("/conversations")
	conv.Use(middleware.RequireSuperAdmin())
	{
		conv.GET("", h.Admin.Conversation.ListSessions)
		conv.GET("/:id", h.Admin.Conversation.GetSession)
		conv.GET("/:id/export", h.Admin.Conversation.ExportSession)
		conv.DELETE("/:id", h.Admin.Conversation.DeleteSession)
	}
	// 批量导出独立路径，避免与 /conversations/:id 的参数段冲突。
	admin.GET("/conversation-exports", middleware.RequireSuperAdmin(), h.Admin.Conversation.ExportAll)
}

func registerAdminComplianceRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	compliance := admin.Group("/compliance")
	{
		compliance.GET("", h.Admin.Compliance.GetStatus)
		compliance.POST("/accept", h.Admin.Compliance.Accept)
	}
}

func registerContentModerationRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	risk := admin.Group("/risk-control")
	{
		risk.GET("/config", h.Admin.ContentModeration.GetConfig)
		risk.PUT("/config", h.Admin.ContentModeration.UpdateConfig)
		risk.POST("/api-keys/test", h.Admin.ContentModeration.TestAPIKeys)
		risk.GET("/status", h.Admin.ContentModeration.GetStatus)
		risk.GET("/logs", h.Admin.ContentModeration.ListLogs)
		risk.POST("/users/:user_id/unban", h.Admin.ContentModeration.UnbanUser)
		risk.DELETE("/hashes", h.Admin.ContentModeration.DeleteFlaggedHash)
		risk.DELETE("/hashes/all", h.Admin.ContentModeration.ClearFlaggedHashes)
	}
}

func registerAdminAPIKeyRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	apiKeys := admin.Group("/api-keys")
	apiKeys.Use(middleware.RequireSuperAdmin())
	{
		apiKeys.PUT("/:id", h.Admin.APIKey.UpdateGroup)
	}
}

func registerOpsRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	ops := admin.Group("/ops")
	{
		// Realtime ops signals
		ops.GET("/concurrency", h.Admin.Ops.GetConcurrencyStats)
		ops.GET("/user-concurrency", h.Admin.Ops.GetUserConcurrencyStats)
		ops.GET("/account-availability", h.Admin.Ops.GetAccountAvailability)
		ops.GET("/realtime-traffic", h.Admin.Ops.GetRealtimeTrafficSummary)

		// Alerts (rules + events)
		ops.GET("/alert-rules", h.Admin.Ops.ListAlertRules)
		ops.POST("/alert-rules", h.Admin.Ops.CreateAlertRule)
		ops.PUT("/alert-rules/:id", h.Admin.Ops.UpdateAlertRule)
		ops.DELETE("/alert-rules/:id", h.Admin.Ops.DeleteAlertRule)
		ops.GET("/alert-events", h.Admin.Ops.ListAlertEvents)
		ops.GET("/alert-events/:id", h.Admin.Ops.GetAlertEvent)
		ops.PUT("/alert-events/:id/status", h.Admin.Ops.UpdateAlertEventStatus)
		ops.POST("/alert-silences", h.Admin.Ops.CreateAlertSilence)

		// Email notification config (DB-backed)
		ops.GET("/email-notification/config", h.Admin.Ops.GetEmailNotificationConfig)
		ops.PUT("/email-notification/config", h.Admin.Ops.UpdateEmailNotificationConfig)

		// Runtime settings (DB-backed)
		runtime := ops.Group("/runtime")
		{
			runtime.GET("/alert", h.Admin.Ops.GetAlertRuntimeSettings)
			runtime.PUT("/alert", h.Admin.Ops.UpdateAlertRuntimeSettings)
			runtime.GET("/logging", h.Admin.Ops.GetRuntimeLogConfig)
			runtime.PUT("/logging", h.Admin.Ops.UpdateRuntimeLogConfig)
			runtime.POST("/logging/reset", h.Admin.Ops.ResetRuntimeLogConfig)
		}

		// Advanced settings (DB-backed)
		ops.GET("/advanced-settings", h.Admin.Ops.GetAdvancedSettings)
		ops.PUT("/advanced-settings", h.Admin.Ops.UpdateAdvancedSettings)

		// Settings group (DB-backed)
		settings := ops.Group("/settings")
		{
			settings.GET("/metric-thresholds", h.Admin.Ops.GetMetricThresholds)
			settings.PUT("/metric-thresholds", h.Admin.Ops.UpdateMetricThresholds)
		}

		// WebSocket realtime (QPS/TPS)
		ws := ops.Group("/ws")
		{
			ws.GET("/qps", h.Admin.Ops.QPSWSHandler)
		}

		// Error logs (legacy)
		ops.GET("/errors", h.Admin.Ops.GetErrorLogs)
		ops.GET("/errors/:id", h.Admin.Ops.GetErrorLogByID)
		ops.PUT("/errors/:id/resolve", h.Admin.Ops.UpdateErrorResolution)

		// Request errors (client-visible failures)
		ops.GET("/request-errors", h.Admin.Ops.ListRequestErrors)
		ops.GET("/request-errors/:id", h.Admin.Ops.GetRequestError)
		ops.GET("/request-errors/:id/upstream-errors", h.Admin.Ops.ListRequestErrorUpstreamErrors)
		ops.PUT("/request-errors/:id/resolve", h.Admin.Ops.ResolveRequestError)

		// Bounded ingress-admission rejection aggregates.
		ops.GET("/ingress-rejections", h.Admin.Ops.ListIngressRejects)
		ops.GET("/ingress-rejections/health", h.Admin.Ops.GetIngressRejectHealth)
		ops.GET("/auth-cache-invalidation/health", h.Admin.Ops.GetAuthCacheInvalidationHealth)

		// Upstream errors (independent upstream failures)
		ops.GET("/upstream-errors", h.Admin.Ops.ListUpstreamErrors)
		ops.GET("/upstream-errors/:id", h.Admin.Ops.GetUpstreamError)
		ops.PUT("/upstream-errors/:id/resolve", h.Admin.Ops.ResolveUpstreamError)

		// Request drilldown (success + error)
		ops.GET("/requests", h.Admin.Ops.ListRequestDetails)

		// Indexed system logs
		ops.GET("/system-logs", h.Admin.Ops.ListSystemLogs)
		ops.POST("/system-logs/cleanup", h.Admin.Ops.CleanupSystemLogs)
		ops.GET("/system-logs/health", h.Admin.Ops.GetSystemLogIngestionHealth)

		// Dashboard (vNext - raw path for MVP)
		ops.GET("/dashboard/snapshot-v2", h.Admin.Ops.GetDashboardSnapshotV2)
		ops.GET("/dashboard/overview", h.Admin.Ops.GetDashboardOverview)
		ops.GET("/dashboard/throughput-trend", h.Admin.Ops.GetDashboardThroughputTrend)
		ops.GET("/dashboard/latency-histogram", h.Admin.Ops.GetDashboardLatencyHistogram)
		ops.GET("/dashboard/error-trend", h.Admin.Ops.GetDashboardErrorTrend)
		ops.GET("/dashboard/error-distribution", h.Admin.Ops.GetDashboardErrorDistribution)
		ops.GET("/dashboard/openai-token-stats", h.Admin.Ops.GetDashboardOpenAITokenStats)
	}
}

func registerDashboardRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	dashboard := admin.Group("/dashboard")
	{
		dashboard.GET("/snapshot-v2", h.Admin.Dashboard.GetSnapshotV2)
		dashboard.GET("/stats", h.Admin.Dashboard.GetStats)
		dashboard.GET("/realtime", h.Admin.Dashboard.GetRealtimeMetrics)
		dashboard.GET("/trend", h.Admin.Dashboard.GetUsageTrend)
		dashboard.GET("/models", h.Admin.Dashboard.GetModelStats)
		dashboard.GET("/groups", h.Admin.Dashboard.GetGroupStats)
		dashboard.GET("/api-keys-trend", h.Admin.Dashboard.GetAPIKeyUsageTrend)
		dashboard.GET("/users-trend", h.Admin.Dashboard.GetUserUsageTrend)
		dashboard.GET("/users-ranking", h.Admin.Dashboard.GetUserSpendingRanking)
		dashboard.POST("/users-usage", h.Admin.Dashboard.GetBatchUsersUsage)
		dashboard.POST("/api-keys-usage", h.Admin.Dashboard.GetBatchAPIKeysUsage)
		dashboard.GET("/user-breakdown", h.Admin.Dashboard.GetUserBreakdown)
		dashboard.POST("/aggregation/backfill", h.Admin.Dashboard.BackfillAggregation)
	}
}

func registerUserManagementRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	users := admin.Group("/users")
	writes := superAdminOnly(users)
	{
		users.GET("", h.Admin.User.List)
		users.GET("/:id", h.Admin.User.GetByID)
		writes.POST("/:id/auth-identities", h.Admin.User.BindAuthIdentity)
		writes.POST("", h.Admin.User.Create)
		writes.PUT("/:id", h.Admin.User.Update)
		writes.DELETE("/:id", h.Admin.User.Delete)
		writes.POST("/:id/balance", h.Admin.User.UpdateBalance)
		users.GET("/:id/api-keys", h.Admin.User.GetUserAPIKeys)
		users.GET("/:id/usage", h.Admin.User.GetUserUsage)
		users.GET("/:id/balance-history", h.Admin.User.GetBalanceHistory)
		writes.POST("/:id/replace-group", h.Admin.User.ReplaceGroup)
		users.GET("/:id/rpm-status", h.Admin.User.GetUserRPMStatus)
		writes.POST("/batch-concurrency", h.Admin.User.BatchUpdateConcurrency)
		writes.POST("/batch-limits", h.Admin.User.BatchUpdateLimits)
		users.GET("/:id/platform-quotas", h.Admin.User.GetUserPlatformQuotas)
		writes.PUT("/:id/platform-quotas", h.Admin.User.UpdateUserPlatformQuotas)
		writes.POST("/:id/platform-quotas/reset", h.Admin.User.ResetUserPlatformQuotaWindow)

		// User attribute values
		users.GET("/:id/attributes", h.Admin.UserAttribute.GetUserAttributes)
		writes.PUT("/:id/attributes", h.Admin.UserAttribute.UpdateUserAttributes)
	}
}

func registerGroupRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	groups := admin.Group("/groups")
	writes := superAdminOnly(groups)
	{
		groups.GET("", h.Admin.Group.List)
		groups.GET("/all", h.Admin.Group.GetAll)
		groups.GET("/usage-summary", h.Admin.Group.GetUsageSummary)
		groups.GET("/capacity-summary", h.Admin.Group.GetCapacitySummary)
		writes.PUT("/sort-order", h.Admin.Group.UpdateSortOrder)
		groups.GET("/:id/models-list-candidates", h.Admin.Group.GetModelsListCandidates)
		groups.GET("/:id", h.Admin.Group.GetByID)
		writes.POST("", h.Admin.Group.Create)
		writes.POST("/:id/duplicate", h.Admin.Group.Duplicate)
		writes.PUT("/:id", h.Admin.Group.Update)
		writes.DELETE("/:id", h.Admin.Group.Delete)
		groups.GET("/:id/stats", h.Admin.Group.GetStats)
		groups.GET("/:id/rate-multipliers", h.Admin.Group.GetGroupRateMultipliers)
		writes.PUT("/:id/rate-multipliers", h.Admin.Group.BatchSetGroupRateMultipliers)
		writes.DELETE("/:id/rate-multipliers", h.Admin.Group.ClearGroupRateMultipliers)
		writes.PUT("/:id/rpm-overrides", h.Admin.Group.BatchSetGroupRPMOverrides)
		writes.DELETE("/:id/rpm-overrides", h.Admin.Group.ClearGroupRPMOverrides)
		groups.GET("/:id/api-keys", h.Admin.Group.GetGroupAPIKeys)
	}
}

func registerResearchGroupRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	groups := admin.Group("/research-groups")
	{
		groups.GET("", h.Admin.ResearchGroup.List)
		groups.GET("/:id", h.Admin.ResearchGroup.GetByID)
		groups.DELETE("/:id", h.Admin.ResearchGroup.Dissolve)
		groups.DELETE("/:id/members/:member_id", h.Admin.ResearchGroup.DetachMember)
	}
}

func registerAccountRoutes(admin *gin.RouterGroup, h *handler.Handlers, stepUpAuth middleware.StepUpAuthMiddleware) {
	accounts := admin.Group("/accounts")
	writes := superAdminOnly(accounts)
	{
		accounts.GET("", h.Admin.Account.List)
		accounts.GET("/upstream-billing-probe/settings", h.Admin.Account.GetUpstreamBillingProbeSettings)
		accounts.PUT("/upstream-billing-probe/settings", h.Admin.Account.UpdateUpstreamBillingProbeSettings)
		accounts.POST("/upstream-billing-probe/batch", h.Admin.Account.ProbeUpstreamBillingBatch)
		accounts.GET("/:id", h.Admin.Account.GetByID)
		writes.POST("", h.Admin.Account.Create)
		writes.POST("/:id/duplicate", h.Admin.Account.Duplicate)
		writes.POST("/check-mixed-channel", h.Admin.Account.CheckMixedChannel)
		writes.POST("/import/codex-session", h.Admin.Account.ImportCodexSession)
		writes.POST("/sync/crs", h.Admin.Account.SyncFromCRS)
		writes.POST("/sync/crs/preview", h.Admin.Account.PreviewFromCRS)
		writes.PUT("/:id", h.Admin.Account.Update)
		writes.PUT("/:id/provider", h.Admin.Account.UpdateProvider)
		writes.PUT("/:id/upstream-billing-probe", h.Admin.Account.SetUpstreamBillingProbeEnabled)
		writes.POST("/:id/upstream-billing-probe", h.Admin.Account.ProbeUpstreamBilling)
		writes.DELETE("/:id", h.Admin.Account.Delete)
		writes.POST("/:id/test", h.Admin.Account.Test)
		writes.POST("/:id/recover-state", h.Admin.Account.RecoverState)
		writes.POST("/:id/refresh", h.Admin.Account.Refresh)
		writes.POST("/:id/apply-oauth-credentials", h.Admin.Account.ApplyOAuthCredentials)
		writes.POST("/:id/set-privacy", h.Admin.Account.SetPrivacy)
		writes.POST("/:id/refresh-tier", h.Admin.Account.RefreshTier)
		accounts.GET("/:id/stats", h.Admin.Account.GetStats)
		writes.POST("/:id/clear-error", h.Admin.Account.ClearError)
		writes.POST("/:id/revert-proxy-fallback", h.Admin.Account.RevertProxyFallback)
		accounts.GET("/:id/usage", h.Admin.Account.GetUsage)
		accounts.GET("/:id/today-stats", h.Admin.Account.GetTodayStats)
		accounts.POST("/today-stats/batch", h.Admin.Account.GetBatchTodayStats)
		writes.POST("/:id/clear-rate-limit", h.Admin.Account.ClearRateLimit)
		writes.POST("/:id/reset-quota", h.Admin.Account.ResetQuota)
		accounts.GET("/:id/temp-unschedulable", h.Admin.Account.GetTempUnschedulable)
		writes.DELETE("/:id/temp-unschedulable", h.Admin.Account.ClearTempUnschedulable)
		writes.POST("/:id/schedulable", h.Admin.Account.SetSchedulable)
		writes.POST("/models/sync-upstream-preview", h.Admin.Account.SyncUpstreamModelsPreview)
		accounts.GET("/:id/models", h.Admin.Account.GetAvailableModels)
		writes.POST("/:id/models/sync-upstream", h.Admin.Account.SyncUpstreamModels)
		writes.POST("/batch", h.Admin.Account.BatchCreate)
		// 账号导出泄露上游凭证原文——要求 step-up 2FA
		writes.GET("/data", gin.HandlerFunc(stepUpAuth), h.Admin.Account.ExportData)
		writes.POST("/data", h.Admin.Account.ImportData)
		writes.POST("/batch-update-credentials", h.Admin.Account.BatchUpdateCredentials)
		writes.POST("/batch-refresh-tier", h.Admin.Account.BatchRefreshTier)
		writes.POST("/bulk-update", h.Admin.Account.BulkUpdate)
		writes.POST("/batch-clear-error", h.Admin.Account.BatchClearError)
		writes.POST("/batch-refresh", h.Admin.Account.BatchRefresh)

		// Antigravity 默认模型映射
		accounts.GET("/antigravity/default-model-mapping", h.Admin.Account.GetAntigravityDefaultModelMapping)

		// Spark 影子账号
		writes.POST("/:id/shadow", h.Admin.OpenAIOAuth.CreateShadow)

		// Claude OAuth routes
		writes.POST("/generate-auth-url", h.Admin.OAuth.GenerateAuthURL)
		writes.POST("/generate-setup-token-url", h.Admin.OAuth.GenerateSetupTokenURL)
		writes.POST("/exchange-code", h.Admin.OAuth.ExchangeCode)
		writes.POST("/exchange-setup-token-code", h.Admin.OAuth.ExchangeSetupTokenCode)
		writes.POST("/cookie-auth", h.Admin.OAuth.CookieAuth)
		writes.POST("/setup-token-cookie-auth", h.Admin.OAuth.SetupTokenCookieAuth)
	}
}

func registerAdminProviderRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	providers := admin.Group("/providers")
	{
		providers.GET("/:id/usage", h.Provider.GetAdminUsage)
	}
}

func registerAnnouncementRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	announcements := admin.Group("/announcements")
	{
		announcements.GET("", h.Admin.Announcement.List)
		announcements.POST("", h.Admin.Announcement.Create)
		announcements.GET("/:id", h.Admin.Announcement.GetByID)
		announcements.PUT("/:id", h.Admin.Announcement.Update)
		announcements.DELETE("/:id", h.Admin.Announcement.Delete)
		announcements.GET("/:id/read-status", h.Admin.Announcement.ListReadStatus)
	}
}

func registerOpenAIOAuthRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	openai := admin.Group("/openai")
	writes := superAdminOnly(openai)
	{
		writes.POST("/generate-auth-url", h.Admin.OpenAIOAuth.GenerateAuthURL)
		writes.POST("/exchange-code", h.Admin.OpenAIOAuth.ExchangeCode)
		writes.POST("/refresh-token", h.Admin.OpenAIOAuth.RefreshToken)
		writes.POST("/accounts/:id/refresh", h.Admin.OpenAIOAuth.RefreshAccountToken)
		writes.POST("/create-from-oauth", h.Admin.OpenAIOAuth.CreateAccountFromOAuth)
		writes.POST("/create-from-codex-pat", h.Admin.OpenAIOAuth.CreateAccountFromCodexPAT)
		openai.GET("/accounts/:id/quota", h.Admin.OpenAIOAuth.QueryQuota)
		writes.POST("/accounts/:id/reset-quota", h.Admin.OpenAIOAuth.ResetQuota)
	}
}

func registerGeminiOAuthRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	gemini := admin.Group("/gemini")
	writes := superAdminOnly(gemini)
	{
		writes.POST("/oauth/auth-url", h.Admin.GeminiOAuth.GenerateAuthURL)
		writes.POST("/oauth/exchange-code", h.Admin.GeminiOAuth.ExchangeCode)
		gemini.GET("/oauth/capabilities", h.Admin.GeminiOAuth.GetCapabilities)
	}
}

func registerAntigravityOAuthRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	antigravity := admin.Group("/antigravity")
	antigravity.Use(middleware.RequireSuperAdmin())
	{
		antigravity.POST("/oauth/auth-url", h.Admin.AntigravityOAuth.GenerateAuthURL)
		antigravity.POST("/oauth/exchange-code", h.Admin.AntigravityOAuth.ExchangeCode)
		antigravity.POST("/oauth/refresh-token", h.Admin.AntigravityOAuth.RefreshToken)
	}
}

func registerGrokOAuthRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	grok := admin.Group("/grok")
	writes := superAdminOnly(grok)
	{
		writes.POST("/oauth/auth-url", h.Admin.GrokOAuth.GenerateAuthURL)
		writes.POST("/oauth/exchange-code", h.Admin.GrokOAuth.ExchangeCode)
		writes.POST("/oauth/refresh-token", h.Admin.GrokOAuth.RefreshToken)
		writes.POST("/oauth/create-from-oauth", h.Admin.GrokOAuth.CreateAccountFromOAuth)
		writes.POST("/sso-to-oauth", h.Admin.GrokOAuth.CreateAccountsFromSSO)
		writes.POST("/oauth/reconcile", h.Admin.GrokOAuth.ReconcileOAuthAccounts)
		writes.POST("/accounts/:id/refresh", h.Admin.GrokOAuth.RefreshAccountToken)
		grok.GET("/accounts/:id/quota", h.Admin.GrokOAuth.QueryQuota)
		writes.POST("/accounts/:id/reset-quota", h.Admin.GrokOAuth.ResetQuota)
		grok.GET("/runtime-sanity", h.Admin.GrokOAuth.RuntimeSanity)
	}
}

func registerProxyRoutes(admin *gin.RouterGroup, h *handler.Handlers, stepUpAuth middleware.StepUpAuthMiddleware) {
	proxies := admin.Group("/proxies")
	writes := superAdminOnly(proxies)
	{
		proxies.GET("", h.Admin.Proxy.List)
		proxies.GET("/all", h.Admin.Proxy.GetAll)
		// 代理导出泄露账号密码原文——要求 step-up 2FA
		writes.GET("/data", gin.HandlerFunc(stepUpAuth), h.Admin.Proxy.ExportData)
		writes.POST("/data", h.Admin.Proxy.ImportData)
		proxies.GET("/:id", h.Admin.Proxy.GetByID)
		writes.POST("", h.Admin.Proxy.Create)
		writes.PUT("/:id", h.Admin.Proxy.Update)
		writes.DELETE("/:id", h.Admin.Proxy.Delete)
		writes.POST("/:id/test", h.Admin.Proxy.Test)
		writes.POST("/:id/quality-check", h.Admin.Proxy.CheckQuality)
		proxies.GET("/:id/stats", h.Admin.Proxy.GetStats)
		proxies.GET("/:id/accounts", h.Admin.Proxy.GetProxyAccounts)
		writes.POST("/batch-delete", h.Admin.Proxy.BatchDelete)
		writes.POST("/batch", h.Admin.Proxy.BatchCreate)
	}
}

func registerRedeemCodeRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	codes := admin.Group("/redeem-codes")
	{
		codes.GET("", h.Admin.Redeem.List)
		codes.GET("/stats", h.Admin.Redeem.GetStats)
		codes.GET("/export", h.Admin.Redeem.Export)
		codes.GET("/:id", h.Admin.Redeem.GetByID)
		codes.POST("/create-and-redeem", h.Admin.Redeem.CreateAndRedeem)
		codes.POST("/generate", h.Admin.Redeem.Generate)
		codes.DELETE("/:id", h.Admin.Redeem.Delete)
		codes.POST("/batch-delete", h.Admin.Redeem.BatchDelete)
		codes.POST("/batch-update", h.Admin.Redeem.BatchUpdate)
		codes.POST("/:id/expire", h.Admin.Redeem.Expire)
	}
}

func registerPromoCodeRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	promoCodes := admin.Group("/promo-codes")
	{
		promoCodes.GET("", h.Admin.Promo.List)
		promoCodes.GET("/:id", h.Admin.Promo.GetByID)
		promoCodes.POST("", h.Admin.Promo.Create)
		promoCodes.PUT("/:id", h.Admin.Promo.Update)
		promoCodes.DELETE("/:id", h.Admin.Promo.Delete)
		promoCodes.GET("/:id/usages", h.Admin.Promo.GetUsages)
	}
}

func registerSettingsRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	adminSettings := admin.Group("/settings")
	writes := superAdminOnly(adminSettings)
	{
		adminSettings.GET("", h.Admin.Setting.GetSettings)
		writes.PUT("", h.Admin.Setting.UpdateSettings)
		writes.POST("/test-smtp", h.Admin.Setting.TestSMTPConnection)
		writes.POST("/send-test-email", h.Admin.Setting.SendTestEmail)
		writes.GET("/email-templates", h.Admin.Setting.ListEmailTemplates)
		writes.POST("/email-template-preview", h.Admin.Setting.PreviewEmailTemplate)
		writes.GET("/email-templates/:event/:locale", h.Admin.Setting.GetEmailTemplate)
		writes.PUT("/email-templates/:event/:locale", h.Admin.Setting.UpdateEmailTemplate)
		writes.POST("/email-templates/:event/:locale/restore-official", h.Admin.Setting.RestoreOfficialEmailTemplate)
		// Admin API Key 管理
		writes.GET("/admin-api-key", h.Admin.Setting.GetAdminAPIKey)
		writes.POST("/admin-api-key/regenerate", h.Admin.Setting.RegenerateAdminAPIKey)
		writes.DELETE("/admin-api-key", h.Admin.Setting.DeleteAdminAPIKey)
		// 529过载冷却配置
		writes.GET("/overload-cooldown", h.Admin.Setting.GetOverloadCooldownSettings)
		writes.PUT("/overload-cooldown", h.Admin.Setting.UpdateOverloadCooldownSettings)
		// 429默认回避配置
		writes.GET("/rate-limit-429-cooldown", h.Admin.Setting.GetRateLimit429CooldownSettings)
		writes.PUT("/rate-limit-429-cooldown", h.Admin.Setting.UpdateRateLimit429CooldownSettings)
		// 流超时处理配置
		writes.GET("/stream-timeout", h.Admin.Setting.GetStreamTimeoutSettings)
		writes.PUT("/stream-timeout", h.Admin.Setting.UpdateStreamTimeoutSettings)
		// 请求整流器配置
		writes.GET("/rectifier", h.Admin.Setting.GetRectifierSettings)
		writes.PUT("/rectifier", h.Admin.Setting.UpdateRectifierSettings)
		// Beta 策略配置
		writes.GET("/beta-policy", h.Admin.Setting.GetBetaPolicySettings)
		writes.PUT("/beta-policy", h.Admin.Setting.UpdateBetaPolicySettings)
		// Web Search 模拟配置
		writes.GET("/web-search-emulation", h.Admin.Setting.GetWebSearchEmulationConfig)
		writes.PUT("/web-search-emulation", h.Admin.Setting.UpdateWebSearchEmulationConfig)
		writes.POST("/web-search-emulation/test", h.Admin.Setting.TestWebSearchEmulation)
		writes.POST("/web-search-emulation/reset-usage", h.Admin.Setting.ResetWebSearchUsage)
	}
}

func registerDataManagementRoutes(admin *gin.RouterGroup, h *handler.Handlers, stepUpAuth middleware.StepUpAuthMiddleware) {
	dataManagement := admin.Group("/data-management")
	{
		dataManagement.GET("/agent/health", h.Admin.DataManagement.GetAgentHealth)
		dataManagement.GET("/config", h.Admin.DataManagement.GetConfig)
		dataManagement.PUT("/config", h.Admin.DataManagement.UpdateConfig)
		dataManagement.GET("/sources/:source_type/profiles", h.Admin.DataManagement.ListSourceProfiles)
		dataManagement.POST("/sources/:source_type/profiles", h.Admin.DataManagement.CreateSourceProfile)
		dataManagement.PUT("/sources/:source_type/profiles/:profile_id", h.Admin.DataManagement.UpdateSourceProfile)
		dataManagement.DELETE("/sources/:source_type/profiles/:profile_id", h.Admin.DataManagement.DeleteSourceProfile)
		dataManagement.POST("/sources/:source_type/profiles/:profile_id/activate", h.Admin.DataManagement.SetActiveSourceProfile)
		dataManagement.POST("/s3/test", h.Admin.DataManagement.TestS3)
		dataManagement.GET("/s3/profiles", h.Admin.DataManagement.ListS3Profiles)
		// 修改 S3 目标可将数据备份外泄——要求 step-up 2FA
		dataManagement.POST("/s3/profiles", gin.HandlerFunc(stepUpAuth), h.Admin.DataManagement.CreateS3Profile)
		dataManagement.PUT("/s3/profiles/:profile_id", gin.HandlerFunc(stepUpAuth), h.Admin.DataManagement.UpdateS3Profile)
		dataManagement.DELETE("/s3/profiles/:profile_id", h.Admin.DataManagement.DeleteS3Profile)
		dataManagement.POST("/s3/profiles/:profile_id/activate", gin.HandlerFunc(stepUpAuth), h.Admin.DataManagement.SetActiveS3Profile)
		dataManagement.POST("/backups", gin.HandlerFunc(stepUpAuth), h.Admin.DataManagement.CreateBackupJob)
		dataManagement.GET("/backups", h.Admin.DataManagement.ListBackupJobs)
		dataManagement.GET("/backups/:job_id", h.Admin.DataManagement.GetBackupJob)
	}
}

func registerBackupRoutes(admin *gin.RouterGroup, h *handler.Handlers, stepUpAuth middleware.StepUpAuthMiddleware) {
	backup := admin.Group("/backups")
	{
		// S3 存储配置
		backup.GET("/s3-config", h.Admin.Backup.GetS3Config)
		// 修改 S3 目标可将数据库备份外泄——要求 step-up 2FA
		backup.PUT("/s3-config", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.UpdateS3Config)
		backup.POST("/s3-config/test", h.Admin.Backup.TestS3Connection)

		// 异步生图对象存储配置（与备份共用 S3 客户端，可直接复用备份凭证）
		backup.GET("/image-storage", h.Admin.Backup.GetImageStorageConfig)
		// 同 S3 配置：改写对象存储目标可将生成内容导向外部账号——要求 step-up 2FA
		backup.PUT("/image-storage", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.UpdateImageStorageConfig)
		backup.POST("/image-storage/test", h.Admin.Backup.TestImageStorageConnection)

		// 定时备份配置
		backup.GET("/schedule", h.Admin.Backup.GetSchedule)
		backup.PUT("/schedule", h.Admin.Backup.UpdateSchedule)

		// 备份操作
		backup.POST("", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.CreateBackup)
		backup.GET("", h.Admin.Backup.ListBackups)
		backup.GET("/:id", h.Admin.Backup.GetBackup)
		backup.DELETE("/:id", h.Admin.Backup.DeleteBackup)
		// 备份下载链接可直接取走整库数据——要求 step-up 2FA
		backup.GET("/:id/download-url", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.GetDownloadURL)

		// 恢复操作：整库覆盖可回滚安全设置（含 step-up 开关本身）——要求 step-up 2FA
		backup.POST("/:id/restore", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.RestoreBackup)
	}
}

func registerSystemRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	system := admin.Group("/system")
	{
		system.GET("/version", h.Admin.System.GetVersion)
		system.GET("/check-updates", h.Admin.System.CheckUpdates)
		system.GET("/rollback-versions", h.Admin.System.GetRollbackVersions)
		system.POST("/update", h.Admin.System.PerformUpdate)
		system.POST("/rollback", h.Admin.System.Rollback)
		system.POST("/restart", h.Admin.System.RestartService)
	}
}

func registerSubscriptionRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	subscriptions := admin.Group("/subscriptions")
	{
		subscriptions.GET("", h.Admin.Subscription.List)
		subscriptions.GET("/:id", h.Admin.Subscription.GetByID)
		subscriptions.GET("/:id/progress", h.Admin.Subscription.GetProgress)
		subscriptions.POST("/assign", h.Admin.Subscription.Assign)
		subscriptions.POST("/bulk-assign", h.Admin.Subscription.BulkAssign)
		subscriptions.POST("/:id/extend", h.Admin.Subscription.Extend)
		subscriptions.POST("/:id/reset-quota", h.Admin.Subscription.ResetQuota)
		subscriptions.POST("/:id/revoke", h.Admin.Subscription.Revoke)
		subscriptions.POST("/:id/restore", h.Admin.Subscription.Restore)
		subscriptions.DELETE("/:id", h.Admin.Subscription.Revoke)
	}

	// 分组下的订阅列表
	admin.GET("/groups/:id/subscriptions", h.Admin.Subscription.ListByGroup)

	// 用户下的订阅列表
	admin.GET("/users/:id/subscriptions", h.Admin.Subscription.ListByUser)
}

func registerUsageRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	usage := admin.Group("/usage")
	{
		usage.GET("", h.Admin.Usage.List)
		usage.GET("/stats", h.Admin.Usage.Stats)
		usage.GET("/search-users", h.Admin.Usage.SearchUsers)
		usage.GET("/search-api-keys", h.Admin.Usage.SearchAPIKeys)
		usage.GET("/cleanup-tasks", h.Admin.Usage.ListCleanupTasks)
		usage.POST("/cleanup-tasks", h.Admin.Usage.CreateCleanupTask)
		usage.POST("/cleanup-tasks/:id/cancel", h.Admin.Usage.CancelCleanupTask)
	}
}

func registerUserAttributeRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	attrs := admin.Group("/user-attributes")
	writes := superAdminOnly(attrs)
	{
		attrs.GET("", h.Admin.UserAttribute.ListDefinitions)
		writes.POST("", h.Admin.UserAttribute.CreateDefinition)
		attrs.POST("/batch", h.Admin.UserAttribute.GetBatchUserAttributes)
		writes.PUT("/reorder", h.Admin.UserAttribute.ReorderDefinitions)
		writes.PUT("/:id", h.Admin.UserAttribute.UpdateDefinition)
		writes.DELETE("/:id", h.Admin.UserAttribute.DeleteDefinition)
	}
}

func registerScheduledTestRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	plans := admin.Group("/scheduled-test-plans")
	writes := superAdminOnly(plans)
	{
		writes.POST("", h.Admin.ScheduledTest.Create)
		writes.PUT("/:id", h.Admin.ScheduledTest.Update)
		writes.DELETE("/:id", h.Admin.ScheduledTest.Delete)
		plans.GET("/:id/results", h.Admin.ScheduledTest.ListResults)
	}
	// Nested under accounts
	admin.GET("/accounts/:id/scheduled-test-plans", h.Admin.ScheduledTest.ListByAccount)
}

func registerErrorPassthroughRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	rules := admin.Group("/error-passthrough-rules")
	rules.Use(middleware.RequireSuperAdmin())
	{
		rules.GET("", h.Admin.ErrorPassthrough.List)
		rules.GET("/:id", h.Admin.ErrorPassthrough.GetByID)
		rules.POST("", h.Admin.ErrorPassthrough.Create)
		rules.PUT("/:id", h.Admin.ErrorPassthrough.Update)
		rules.DELETE("/:id", h.Admin.ErrorPassthrough.Delete)
	}
}

func registerTLSFingerprintProfileRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	profiles := admin.Group("/tls-fingerprint-profiles")
	profiles.Use(middleware.RequireSuperAdmin())
	{
		profiles.GET("", h.Admin.TLSFingerprintProfile.List)
		profiles.GET("/:id", h.Admin.TLSFingerprintProfile.GetByID)
		profiles.POST("", h.Admin.TLSFingerprintProfile.Create)
		profiles.PUT("/:id", h.Admin.TLSFingerprintProfile.Update)
		profiles.DELETE("/:id", h.Admin.TLSFingerprintProfile.Delete)
	}
}

func registerChannelRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	channels := admin.Group("/channels")
	{
		channels.GET("", h.Admin.Channel.List)
		channels.GET("/model-pricing", h.Admin.Channel.GetModelDefaultPricing)
		channels.GET("/pricing/sync-models", h.Admin.Channel.SyncPricingModels)
		channels.GET("/:id", h.Admin.Channel.GetByID)
		channels.POST("", h.Admin.Channel.Create)
		channels.PUT("/:id", h.Admin.Channel.Update)
		channels.DELETE("/:id", h.Admin.Channel.Delete)
	}
}

func registerChannelMonitorRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	monitors := admin.Group("/channel-monitors")
	{
		monitors.GET("", h.Admin.ChannelMonitor.List)
		monitors.POST("", h.Admin.ChannelMonitor.Create)
		monitors.GET("/:id", h.Admin.ChannelMonitor.Get)
		monitors.POST("/:id/duplicate", h.Admin.ChannelMonitor.Duplicate)
		monitors.PUT("/:id", h.Admin.ChannelMonitor.Update)
		monitors.DELETE("/:id", h.Admin.ChannelMonitor.Delete)
		monitors.POST("/:id/run", h.Admin.ChannelMonitor.Run)
		monitors.GET("/:id/history", h.Admin.ChannelMonitor.History)
	}

	templates := admin.Group("/channel-monitor-templates")
	{
		templates.GET("", h.Admin.ChannelMonitorTemplate.List)
		templates.POST("", h.Admin.ChannelMonitorTemplate.Create)
		templates.GET("/:id", h.Admin.ChannelMonitorTemplate.Get)
		templates.PUT("/:id", h.Admin.ChannelMonitorTemplate.Update)
		templates.DELETE("/:id", h.Admin.ChannelMonitorTemplate.Delete)
		templates.GET("/:id/monitors", h.Admin.ChannelMonitorTemplate.AssociatedMonitors)
		templates.POST("/:id/apply", h.Admin.ChannelMonitorTemplate.Apply)
	}
}

// registerAffiliateRoutes 注册邀请返利的管理端路由（专属用户配置）
func registerAffiliateRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	affiliates := admin.Group("/affiliates")
	writes := superAdminOnly(affiliates)
	{
		affiliates.GET("/invites", h.Admin.Affiliate.ListInviteRecords)
		affiliates.GET("/rebates", h.Admin.Affiliate.ListRebateRecords)
		affiliates.GET("/transfers", h.Admin.Affiliate.ListTransferRecords)

		users := affiliates.Group("/users")
		userWrites := superAdminOnly(users)
		{
			users.GET("", h.Admin.Affiliate.ListUsers)
			users.GET("/lookup", h.Admin.Affiliate.LookupUsers)
			userWrites.POST("/batch-rate", h.Admin.Affiliate.BatchSetRate)
			users.GET("/:user_id/overview", h.Admin.Affiliate.GetUserOverview)
			userWrites.PUT("/:user_id", h.Admin.Affiliate.UpdateUserSettings)
			userWrites.DELETE("/:user_id", h.Admin.Affiliate.ClearUserSettings)
		}
		affiliates.GET("/agents", h.Admin.Affiliate.ListAgents)
		writes.PUT("/agents/:userId/status", h.Admin.Affiliate.UpdateAgentStatus)
		affiliates.GET("/withdrawals", h.Admin.Affiliate.ListWithdrawals)
		writes.GET("/withdrawals/:id", h.Admin.Affiliate.GetWithdrawal)
		writes.POST("/withdrawals/:id/approve", h.Admin.Affiliate.ApproveWithdrawal)
		writes.POST("/withdrawals/:id/reject", h.Admin.Affiliate.RejectWithdrawal)
		writes.POST("/withdrawals/:id/mark-paid", h.Admin.Affiliate.MarkWithdrawalPaid)
	}
}
