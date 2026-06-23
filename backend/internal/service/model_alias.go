package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ModelAliasFormat 标识请求体的协议形态，决定 reasoning effort 注入到哪个字段：
//   - Responses 原生：reasoning.effort（嵌套对象）
//   - Chat Completions：reasoning_effort（顶层平铺字段）
//
// service_tier 在两种形态下都是顶层字段，无需区分。
type ModelAliasFormat int

const (
	ModelAliasFormatResponses ModelAliasFormat = iota
	ModelAliasFormatChatCompletions
)

// ModelAliasSpec 定义一个对外暴露的"虚拟模型别名"到真实上游模型 + 预设参数的映射。
//
// 别名是一个纯加法层：命中后把请求体改写成一个普通的 TargetModel 请求（带上预设的
// reasoning effort / service_tier），后续所有既有逻辑（渠道映射、账号调度、fast 策略、
// 计费）都按改写后的请求处理，无需感知别名的存在。
//
// 计费说明（与定价配置解耦）：别名自带 RateMultiplier，由 usage 记账侧直接把该倍率
// 叠加到请求计费倍率上，并把别名注入的 service_tier 在计费侧中和，确保 RateMultiplier
// 是唯一加价来源——无需配置 gpt-5.5 的 priority 单价即可得到精确的 1.5x。上游仍收到
// service_tier=priority 以走 fast 提速，只有"计费"这一侧不再叠加 priority 倍率。
type ModelAliasSpec struct {
	Alias           string  // 对外模型名，如 "isac-gpt-fast"
	TargetModel     string  // 改写后的真实上游模型，如 "gpt-5.5"
	ReasoningEffort string  // 预设推理档位（"" = 不注入），如 "xhigh"(extra high)
	ServiceTier     string  // 预设服务档位（"" = 不注入），如 "priority"(fast)
	RateMultiplier  float64 // 计费倍率（<=0 = 不改计费），如 1.5；命中后中和注入的 service_tier 计费
	Platform        string  // 所属平台（仅在该平台的 /v1/models 中暴露）
}

// modelAliasRegistry 是别名注册表（唯一权威来源）。
//
// defaultModelAliasRegistry 是代码内置默认表（未在 config.yaml 配置 model_aliases 时生效）。
// 当前两条对应需求：
//   - isac-gpt-fast → gpt-5.5 + extra high + fast(priority)，计费 1.5x（仅该别名加价）
//   - isac-gpt-best → gpt-5.5 + extra high（计费基础价，RateMultiplier=0 不改计费）
var defaultModelAliasRegistry = []ModelAliasSpec{
	{
		Alias:           "isac-gpt-fast",
		TargetModel:     "gpt-5.5",
		ReasoningEffort: "xhigh",
		ServiceTier:     "priority",
		RateMultiplier:  1.5,
		Platform:        PlatformOpenAI,
	},
	{
		Alias:           "isac-gpt-best",
		TargetModel:     "gpt-5.5",
		ReasoningEffort: "xhigh",
		ServiceTier:     "",
		RateMultiplier:  0,
		Platform:        PlatformOpenAI,
	},
}

// modelAliasRegistry 是当前生效的别名表。默认指向内置表；启动时若 config.yaml 配置了
// model_aliases，则由 SetModelAliasRegistry 覆盖。
//
// 并发约定：仅在进程启动阶段（开始 serving 之前）通过 SetModelAliasRegistry 写入一次，
// 之后为只读，故无需加锁。
var modelAliasRegistry = defaultModelAliasRegistry

// SetModelAliasRegistry 用配置覆盖生效的别名表。specs 为空时回退到内置默认表
// （便于"未配置即用默认"语义，也便于测试还原）。仅应在启动阶段调用一次。
func SetModelAliasRegistry(specs []ModelAliasSpec) {
	if len(specs) == 0 {
		modelAliasRegistry = defaultModelAliasRegistry
		return
	}
	modelAliasRegistry = specs
}

// ModelAliasSpecsFromConfig 把 config.yaml 的 model_aliases 配置转换为别名定义。
// 跳过 alias / target_model 为空的非法条目；platform 缺省为 openai。
func ModelAliasSpecsFromConfig(items []config.ModelAliasConfig) []ModelAliasSpec {
	specs := make([]ModelAliasSpec, 0, len(items))
	for _, it := range items {
		alias := strings.TrimSpace(it.Alias)
		target := strings.TrimSpace(it.TargetModel)
		if alias == "" || target == "" {
			continue
		}
		platform := strings.TrimSpace(it.Platform)
		if platform == "" {
			platform = PlatformOpenAI
		}
		specs = append(specs, ModelAliasSpec{
			Alias:           alias,
			TargetModel:     target,
			ReasoningEffort: strings.TrimSpace(it.ReasoningEffort),
			ServiceTier:     strings.TrimSpace(it.ServiceTier),
			RateMultiplier:  it.RateMultiplier,
			Platform:        platform,
		})
	}
	return specs
}

// modelAliasRateCtxKey 是别名计费倍率在 context 中的 key。
type modelAliasRateCtxKey struct{}

// WithModelAliasRate 把别名计费倍率绑定到 context，供 usage 记账侧读取。
// multiplier <= 0 时不绑定（视为无计费覆盖）。
func WithModelAliasRate(ctx context.Context, multiplier float64) context.Context {
	if ctx == nil || multiplier <= 0 {
		return ctx
	}
	return context.WithValue(ctx, modelAliasRateCtxKey{}, multiplier)
}

// ModelAliasRateFromContext 读取别名计费倍率；未设置返回 0。
func ModelAliasRateFromContext(ctx context.Context) float64 {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(modelAliasRateCtxKey{}).(float64); ok {
		return v
	}
	return 0
}

// modelAliasLookupKey 归一化一个客户端请求的模型名，得到用于注册表匹配的 key：
// 去空白、去 "provider/" 前缀、转小写。
func modelAliasLookupKey(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return ""
	}
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
		trimmed = trimmed[idx+1:]
	}
	return strings.ToLower(strings.TrimSpace(trimmed))
}

// LookupModelAlias 按模型名（大小写不敏感、容忍 provider/ 前缀）查找别名定义。
func LookupModelAlias(model string) (ModelAliasSpec, bool) {
	key := modelAliasLookupKey(model)
	if key == "" {
		return ModelAliasSpec{}, false
	}
	for _, spec := range modelAliasRegistry {
		if strings.ToLower(spec.Alias) == key {
			return spec, true
		}
	}
	return ModelAliasSpec{}, false
}

// ApplyModelAlias 检查请求体的 model 是否命中别名；命中则返回改写后的请求体、
// 命中的别名定义、applied=true。未命中时原样返回 applied=false（零拷贝、零开销）。
//
// 改写内容：
//   - model → spec.TargetModel
//   - reasoning effort → spec.ReasoningEffort（强制覆盖：别名即代表固定档位）
//   - service_tier → spec.ServiceTier（仅在 spec 非空时设置；为空则不动客户端原值）
//
// 任一 sjson 写入失败时保守降级（返回已成功改写的部分或原体），保证 fail-open。
func ApplyModelAlias(body []byte, format ModelAliasFormat) ([]byte, ModelAliasSpec, bool) {
	if len(body) == 0 {
		return body, ModelAliasSpec{}, false
	}
	model := gjson.GetBytes(body, "model").String()
	spec, ok := LookupModelAlias(model)
	if !ok {
		return body, ModelAliasSpec{}, false
	}

	out := body
	if spec.TargetModel != "" {
		if rewritten, err := sjson.SetBytes(out, "model", spec.TargetModel); err == nil {
			out = rewritten
		}
	}

	if effort := strings.TrimSpace(spec.ReasoningEffort); effort != "" {
		path := "reasoning.effort"
		if format == ModelAliasFormatChatCompletions {
			path = "reasoning_effort"
		}
		if rewritten, err := sjson.SetBytes(out, path, effort); err == nil {
			out = rewritten
		}
	}

	if tier := strings.TrimSpace(spec.ServiceTier); tier != "" {
		if rewritten, err := sjson.SetBytes(out, "service_tier", tier); err == nil {
			out = rewritten
		}
	}

	return out, spec, true
}

// ModelAliasNamesForPlatform 返回指定平台下所有别名的对外模型名，用于 /v1/models 暴露。
// 平台不匹配时返回 nil。
func ModelAliasNamesForPlatform(platform string) []string {
	if platform == "" {
		return nil
	}
	var names []string
	for _, spec := range modelAliasRegistry {
		if spec.Platform == platform {
			names = append(names, spec.Alias)
		}
	}
	return names
}

// AppendModelAliasNames 把指定平台的别名追加到给定模型列表尾部（去重，大小写不敏感）。
// 用于在 /v1/models 输出中并入别名，同时保留原有模型。
func AppendModelAliasNames(platform string, models []string) []string {
	aliases := ModelAliasNamesForPlatform(platform)
	if len(aliases) == 0 {
		return models
	}
	seen := make(map[string]struct{}, len(models))
	for _, m := range models {
		seen[strings.ToLower(m)] = struct{}{}
	}
	out := models
	for _, a := range aliases {
		if _, dup := seen[strings.ToLower(a)]; dup {
			continue
		}
		seen[strings.ToLower(a)] = struct{}{}
		out = append(out, a)
	}
	return out
}
