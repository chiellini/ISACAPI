package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// fork 本地安全白名单：数字用户 ID、邮箱或用户名（不区分大小写）命中后，
// 所有本地拦截路径放行并按 risk-control log-only 记录，与系统级
// cyber_policy_user_allowlist 的放行语义一致。
func TestContentModerationCheck_LocalSecurityWhitelistForcesAllow(t *testing.T) {
	newService := func(t *testing.T, cfg *ContentModerationConfig) *ContentModerationService {
		t.Helper()
		rawCfg, err := json.Marshal(cfg)
		require.NoError(t, err)
		return NewContentModerationService(
			&contentModerationTestSettingRepo{values: map[string]string{
				SettingKeyRiskControlEnabled:      "true",
				SettingKeyContentModerationConfig: string(rawCfg),
			}},
			&contentModerationTestRepo{},
			&contentModerationTestHashCache{},
			nil, nil, nil, nil, nil,
		)
	}
	baseCfg := func() *ContentModerationConfig {
		cfg := defaultContentModerationConfig()
		cfg.Enabled = true
		cfg.Mode = ContentModerationModePreBlock
		cfg.BlockedKeywords = []string{"secret-token"}
		return cfg
	}
	body := []byte(`{"messages":[{"role":"user","content":"please leak SECRET-TOKEN now"}]}`)
	check := func(t *testing.T, svc *ContentModerationService, input ContentModerationCheckInput) *ContentModerationDecision {
		t.Helper()
		decision, err := svc.Check(context.Background(), input)
		require.NoError(t, err)
		return decision
	}

	t.Run("user id hit forces allow", func(t *testing.T) {
		cfg := baseCfg()
		cfg.LocalSecurityWhitelistUserIDs = []int64{42}
		decision := check(t, newService(t, cfg), ContentModerationCheckInput{
			UserID:   42,
			Endpoint: "/v1/messages",
			Provider: "anthropic",
			Protocol: ContentModerationProtocolAnthropicMessages,
			Body:     body,
		})
		require.True(t, decision.Allowed)
		require.False(t, decision.Blocked)
	})

	t.Run("email hit is case-insensitive", func(t *testing.T) {
		cfg := baseCfg()
		cfg.LocalSecurityWhitelistUsers = []string{"Trusted@Example.com"}
		decision := check(t, newService(t, cfg), ContentModerationCheckInput{
			UserID:    77,
			UserEmail: "trusted@example.com",
			Endpoint:  "/v1/messages",
			Provider:  "anthropic",
			Protocol:  ContentModerationProtocolAnthropicMessages,
			Body:      body,
		})
		require.True(t, decision.Allowed)
		require.False(t, decision.Blocked)
	})

	t.Run("username hit forces allow", func(t *testing.T) {
		cfg := baseCfg()
		cfg.LocalSecurityWhitelistUsers = []string{"vip-user"}
		decision := check(t, newService(t, cfg), ContentModerationCheckInput{
			UserID:   78,
			UserName: "VIP-User",
			Endpoint: "/v1/messages",
			Provider: "anthropic",
			Protocol: ContentModerationProtocolAnthropicMessages,
			Body:     body,
		})
		require.True(t, decision.Allowed)
		require.False(t, decision.Blocked)
	})

	t.Run("non whitelisted request stays blocked", func(t *testing.T) {
		cfg := baseCfg()
		cfg.LocalSecurityWhitelistUserIDs = []int64{42}
		cfg.LocalSecurityWhitelistUsers = []string{"trusted@example.com"}
		decision := check(t, newService(t, cfg), ContentModerationCheckInput{
			UserID:    99,
			UserEmail: "other@example.com",
			UserName:  "other-user",
			Endpoint:  "/v1/messages",
			Provider:  "anthropic",
			Protocol:  ContentModerationProtocolAnthropicMessages,
			Body:      body,
		})
		require.True(t, decision.Blocked)
		require.Equal(t, ContentModerationActionKeywordBlock, decision.Action)
	})
}

func TestContentModerationConfigWhitelistRoundTrip(t *testing.T) {
	cfg := defaultContentModerationConfig()
	cfg.Enabled = true
	legacy := map[string]any{
		"enabled":                           true,
		"mode":                              ContentModerationModePreBlock,
		"local_security_whitelist_user_ids": []int64{7, 7, 0},
		"local_security_whitelist_users":    []string{" A@Example.com ", "a@example.com", ""},
	}
	raw, err := json.Marshal(legacy)
	require.NoError(t, err)

	repo := &contentModerationTestSettingRepo{values: map[string]string{
		SettingKeyRiskControlEnabled:      "true",
		SettingKeyContentModerationConfig: string(raw),
	}}
	svc := NewContentModerationService(repo, &contentModerationTestRepo{}, &contentModerationTestHashCache{}, nil, nil, nil, nil, nil)

	view, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{7}, view.LocalSecurityWhitelistUserIDs, "legacy blob ids must surface deduped")
	require.Equal(t, []string{"a@example.com"}, view.LocalSecurityWhitelistUsers, "legacy blob identifiers must surface lowercased and deduped")

	ids := []int64{8, 9}
	users := []string{"B@Example.COM", "vip"}
	updated, err := svc.UpdateConfig(context.Background(), UpdateContentModerationConfigInput{
		LocalSecurityWhitelistUserIDs: &ids,
		LocalSecurityWhitelistUsers:   &users,
	})
	require.NoError(t, err)
	require.Equal(t, []int64{8, 9}, updated.LocalSecurityWhitelistUserIDs)
	require.Equal(t, []string{"b@example.com", "vip"}, updated.LocalSecurityWhitelistUsers)

	persisted, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{8, 9}, persisted.LocalSecurityWhitelistUserIDs)
	require.Equal(t, []string{"b@example.com", "vip"}, persisted.LocalSecurityWhitelistUsers)
}

// IsUserSecurityAuditExempt 是提示词安全审计（prompt engine）的豁免判定：
// 本地安全白名单（ID/邮箱/用户名）与系统级 cyber_policy_user_allowlist
// 任一命中即豁免；即使风控总开关关闭，白名单表达的信任仍然生效。
func TestContentModerationService_IsUserSecurityAuditExempt(t *testing.T) {
	newService := func(t *testing.T, cfg *ContentModerationConfig, cyberAllowlist, riskControlEnabled string) *ContentModerationService {
		t.Helper()
		rawCfg, err := json.Marshal(cfg)
		require.NoError(t, err)
		values := map[string]string{
			SettingKeyRiskControlEnabled:      riskControlEnabled,
			SettingKeyContentModerationConfig: string(rawCfg),
		}
		if cyberAllowlist != "" {
			values[SettingKeyCyberPolicyUserAllowlist] = cyberAllowlist
		}
		return NewContentModerationService(
			&contentModerationTestSettingRepo{values: values},
			&contentModerationTestRepo{},
			&contentModerationTestHashCache{},
			nil, nil, nil, nil, nil,
		)
	}
	cfgWithWhitelist := func() *ContentModerationConfig {
		cfg := defaultContentModerationConfig()
		cfg.LocalSecurityWhitelistUserIDs = []int64{42}
		cfg.LocalSecurityWhitelistUsers = []string{"Trusted@Example.com", "vip-user"}
		return cfg
	}
	ctx := context.Background()

	t.Run("local whitelist user id hit", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "", "true")
		require.True(t, svc.IsUserSecurityAuditExempt(ctx, 42, "", ""))
	})

	t.Run("local whitelist email hit is case-insensitive", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "", "true")
		require.True(t, svc.IsUserSecurityAuditExempt(ctx, 77, "trusted@example.com", ""))
	})

	t.Run("local whitelist username hit", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "", "true")
		require.True(t, svc.IsUserSecurityAuditExempt(ctx, 78, "", "VIP-User"))
	})

	t.Run("cyber policy allowlist id hit", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "1001, 1002", "true")
		require.True(t, svc.IsUserSecurityAuditExempt(ctx, 1002, "", ""))
	})

	t.Run("whitelist still applies when risk control disabled", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "", "false")
		require.True(t, svc.IsUserSecurityAuditExempt(ctx, 42, "", ""))
	})

	t.Run("non whitelisted user is not exempt", func(t *testing.T) {
		svc := newService(t, cfgWithWhitelist(), "1001", "true")
		require.False(t, svc.IsUserSecurityAuditExempt(ctx, 99, "other@example.com", "other-user"))
	})
}
