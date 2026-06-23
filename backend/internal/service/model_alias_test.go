package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestLookupModelAlias(t *testing.T) {
	cases := []struct {
		name      string
		model     string
		wantOK    bool
		wantAlias string
	}{
		{"fast exact", "isac-gpt-fast", true, "isac-gpt-fast"},
		{"best exact", "isac-gpt-best", true, "isac-gpt-best"},
		{"case insensitive", "ISAC-GPT-Fast", true, "isac-gpt-fast"},
		{"provider prefix", "openai/isac-gpt-best", true, "isac-gpt-best"},
		{"whitespace", "  isac-gpt-fast  ", true, "isac-gpt-fast"},
		{"unknown", "gpt-5.5", false, ""},
		{"empty", "", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := LookupModelAlias(tc.model)
			require.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				require.Equal(t, tc.wantAlias, spec.Alias)
			}
		})
	}
}

func TestApplyModelAlias_ResponsesFast(t *testing.T) {
	body := []byte(`{"model":"isac-gpt-fast","input":"hi"}`)
	out, spec, applied := ApplyModelAlias(body, ModelAliasFormatResponses)

	require.True(t, applied)
	require.Equal(t, "isac-gpt-fast", spec.Alias)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(out, "model").String())
	require.Equal(t, "xhigh", gjson.GetBytes(out, "reasoning.effort").String())
	require.Equal(t, "priority", gjson.GetBytes(out, "service_tier").String())
	// 原有字段保留
	require.Equal(t, "hi", gjson.GetBytes(out, "input").String())
}

func TestApplyModelAlias_ResponsesBest_NoServiceTier(t *testing.T) {
	body := []byte(`{"model":"isac-gpt-best","input":"hi"}`)
	out, _, applied := ApplyModelAlias(body, ModelAliasFormatResponses)

	require.True(t, applied)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(out, "model").String())
	require.Equal(t, "xhigh", gjson.GetBytes(out, "reasoning.effort").String())
	// best 不注入 service_tier
	require.False(t, gjson.GetBytes(out, "service_tier").Exists())
}

func TestApplyModelAlias_ChatCompletionsUsesFlatEffort(t *testing.T) {
	body := []byte(`{"model":"isac-gpt-fast","messages":[]}`)
	out, _, applied := ApplyModelAlias(body, ModelAliasFormatChatCompletions)

	require.True(t, applied)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(out, "model").String())
	// Chat Completions 使用平铺 reasoning_effort，而非嵌套 reasoning.effort
	require.Equal(t, "xhigh", gjson.GetBytes(out, "reasoning_effort").String())
	require.False(t, gjson.GetBytes(out, "reasoning.effort").Exists())
	require.Equal(t, "priority", gjson.GetBytes(out, "service_tier").String())
}

func TestApplyModelAlias_UnknownModelUntouched(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","input":"hi"}`)
	out, _, applied := ApplyModelAlias(body, ModelAliasFormatResponses)

	require.False(t, applied)
	require.Equal(t, "gpt-5.5", gjson.GetBytes(out, "model").String())
	require.False(t, gjson.GetBytes(out, "reasoning.effort").Exists())
	require.False(t, gjson.GetBytes(out, "service_tier").Exists())
}

func TestModelAlias_RateMultiplierOnlyOnFast(t *testing.T) {
	fast, ok := LookupModelAlias("isac-gpt-fast")
	require.True(t, ok)
	require.Equal(t, 1.5, fast.RateMultiplier)

	best, ok := LookupModelAlias("isac-gpt-best")
	require.True(t, ok)
	require.Equal(t, 0.0, best.RateMultiplier, "best 不改计费")
}

func TestModelAliasRateContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	require.Equal(t, 0.0, ModelAliasRateFromContext(ctx))

	ctx = WithModelAliasRate(ctx, 1.5)
	require.Equal(t, 1.5, ModelAliasRateFromContext(ctx))

	// <=0 不绑定（无计费覆盖）
	require.Equal(t, 0.0, ModelAliasRateFromContext(WithModelAliasRate(context.Background(), 0)))
	// nil ctx 安全
	require.Equal(t, 0.0, ModelAliasRateFromContext(nil))
}

func TestModelAliasConfigOverrideAndFallback(t *testing.T) {
	defer SetModelAliasRegistry(nil) // 测试结束还原内置默认表

	specs := ModelAliasSpecsFromConfig([]config.ModelAliasConfig{
		{Alias: "isac-gpt-fast", TargetModel: "gpt-6", ReasoningEffort: "high", ServiceTier: "priority", RateMultiplier: 2.0},
		{Alias: "", TargetModel: "x"},                       // 跳过：alias 空
		{Alias: "y", TargetModel: ""},                       // 跳过：target 空
		{Alias: "custom-mini", TargetModel: "gpt-5.5-mini"}, // platform 缺省 openai
	})
	require.Len(t, specs, 2)

	SetModelAliasRegistry(specs)
	got, ok := LookupModelAlias("isac-gpt-fast")
	require.True(t, ok)
	require.Equal(t, "gpt-6", got.TargetModel)
	require.Equal(t, 2.0, got.RateMultiplier)

	mini, ok := LookupModelAlias("custom-mini")
	require.True(t, ok)
	require.Equal(t, PlatformOpenAI, mini.Platform)

	// 内置 best 已被配置覆盖、不再存在
	_, ok = LookupModelAlias("isac-gpt-best")
	require.False(t, ok)

	// 空配置回退内置默认表
	SetModelAliasRegistry(nil)
	_, ok = LookupModelAlias("isac-gpt-best")
	require.True(t, ok)
}

func TestModelAliasNamesForPlatform(t *testing.T) {
	openaiNames := ModelAliasNamesForPlatform(PlatformOpenAI)
	require.ElementsMatch(t, []string{"isac-gpt-fast", "isac-gpt-best"}, openaiNames)

	require.Empty(t, ModelAliasNamesForPlatform(PlatformAnthropic))
	require.Empty(t, ModelAliasNamesForPlatform(""))
}

func TestAppendModelAliasNames_DedupAndPlatformScoped(t *testing.T) {
	base := []string{"gpt-5.5", "isac-gpt-fast"}
	out := AppendModelAliasNames(PlatformOpenAI, base)
	// 已存在的 isac-gpt-fast 不重复，新增 isac-gpt-best
	require.ElementsMatch(t, []string{"gpt-5.5", "isac-gpt-fast", "isac-gpt-best"}, out)

	// 非 openai 平台不追加
	anthropicBase := []string{"claude-opus-4"}
	require.Equal(t, anthropicBase, AppendModelAliasNames(PlatformAnthropic, anthropicBase))
}
