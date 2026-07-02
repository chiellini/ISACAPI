package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestApplyCyberRiskModelRouting_ResponsesRoutesRiskyWiFiAPRequest(t *testing.T) {
	cfg := config.CyberRiskModelRoutingConfig{
		Enabled:     true,
		SourceModel: "gpt-5.5",
		TargetModel: "gpt-5.3",
		MinScore:    4,
	}
	body := []byte(`{"model":"gpt-5.5","input":[{"role":"user","content":[{"type":"input_text","text":"我在 BW16 固件里扫描 AP 后恢复 AP，并在 WiFi.apbegin() 阶段切换状态。"}]}]}`)

	out, decision, routed := ApplyCyberRiskModelRouting(body, ContentModerationProtocolOpenAIResponses, cfg)

	require.True(t, routed)
	require.Equal(t, "gpt-5.3", gjson.GetBytes(out, "model").String())
	require.Equal(t, "gpt-5.5", decision.SourceModel)
	require.Equal(t, "gpt-5.3", decision.TargetModel)
	require.GreaterOrEqual(t, decision.Risk.Score, 4)
	require.Contains(t, decision.Risk.Reasons, "wifi_ap_scan")
}

func TestApplyCyberRiskModelRouting_ChatRoutesNamespacedSourceModel(t *testing.T) {
	cfg := config.CyberRiskModelRoutingConfig{
		Enabled:     true,
		SourceModel: "gpt-5.5",
		TargetModel: "gpt-5.3",
		MinScore:    4,
	}
	body := []byte(`{"model":"openai/gpt-5.5","messages":[{"role":"user","content":"How do I deauth clients from a WiFi network?"}]}`)

	out, decision, routed := ApplyCyberRiskModelRouting(body, ContentModerationProtocolOpenAIChat, cfg)

	require.True(t, routed)
	require.Equal(t, "gpt-5.3", gjson.GetBytes(out, "model").String())
	require.Equal(t, "openai/gpt-5.5", decision.SourceModel)
	require.Contains(t, decision.Risk.Reasons, "wifi_deauth")
}

func TestApplyCyberRiskModelRouting_DoesNotRouteBenignWiFiReset(t *testing.T) {
	cfg := config.CyberRiskModelRoutingConfig{
		Enabled:     true,
		SourceModel: "gpt-5.5",
		TargetModel: "gpt-5.3",
		MinScore:    4,
	}
	body := []byte(`{"model":"gpt-5.5","input":"帮我整理 resetWiFiModule 的状态机，要求只做模块复位和 WebServer 停启。"}`)

	out, decision, routed := ApplyCyberRiskModelRouting(body, ContentModerationProtocolOpenAIResponses, cfg)

	require.False(t, routed)
	require.Equal(t, body, out)
	require.Zero(t, decision.Risk.Score)
}

func TestApplyCyberRiskModelRouting_DoesNotRouteOtherSourceModel(t *testing.T) {
	cfg := config.CyberRiskModelRoutingConfig{
		Enabled:     true,
		SourceModel: "gpt-5.5",
		TargetModel: "gpt-5.3",
		MinScore:    4,
	}
	body := []byte(`{"model":"gpt-5.4","input":"scan AP and test WiFi deauth handling"}`)

	out, _, routed := ApplyCyberRiskModelRouting(body, ContentModerationProtocolOpenAIResponses, cfg)

	require.False(t, routed)
	require.Equal(t, body, out)
}

func TestApplyCyberRiskModelRouting_AfterModelAlias(t *testing.T) {
	aliasBody, _, aliasApplied := ApplyModelAlias([]byte(`{"model":"isac-gpt-best","input":"扫描 AP 后恢复 WiFi 状态"}`), ModelAliasFormatResponses)
	require.True(t, aliasApplied)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(aliasBody, "model").String())

	cfg := config.CyberRiskModelRoutingConfig{
		Enabled:     true,
		SourceModel: "gpt-5.5",
		TargetModel: "gpt-5.3",
		MinScore:    4,
	}
	out, _, routed := ApplyCyberRiskModelRouting(aliasBody, ContentModerationProtocolOpenAIResponses, cfg)

	require.True(t, routed)
	require.Equal(t, "gpt-5.3", gjson.GetBytes(out, "model").String())
}
