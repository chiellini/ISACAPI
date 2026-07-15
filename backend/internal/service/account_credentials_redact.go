package service

// SensitiveCredentialKeys 列出 Account.Credentials JSON map 中绝不允许返回到前端的子键。
// dto 层做响应脱敏、service 层做更新合并都引用此清单——新增凭证类型时务必同步。
var SensitiveCredentialKeys = []string{
	// OAuth
	"access_token", "refresh_token", "id_token", "agent_private_key",
	// API Key 类
	"api_key", "session_key", "cookie",
	// 云服务凭据
	"aws_secret_access_key", "aws_session_token",
	"service_account_json", "service_account", "private_key",
}

var sensitiveCredentialKeySet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(SensitiveCredentialKeys))
	for _, k := range SensitiveCredentialKeys {
		m[k] = struct{}{}
	}
	return m
}()

// IsSensitiveCredentialKey 判断指定键是否为敏感凭证子键。
func IsSensitiveCredentialKey(key string) bool {
	_, ok := sensitiveCredentialKeySet[key]
	return ok
}

// MergePreservingSensitiveCreds 把 incoming 写入 existing 之上，但敏感子键采用"incoming 没提供就保留 existing"
// 的语义。返回新的 map，不修改入参。
//
// 用途：前端编辑账号通常采用"全对象 PUT"模式；脱敏后前端 spread 旧 credentials 时不会带上敏感键，
// 直接覆盖会清空已有 token。此函数保证：
//   - 非敏感键：完全由 incoming 决定（用户可以编辑、删除非敏感字段）。
//   - 敏感键：incoming 显式提供则覆盖（用户主动旋转 token），否则保留 existing。
func MergePreservingSensitiveCreds(existing, incoming map[string]any) map[string]any {
	out := make(map[string]any, len(incoming)+len(SensitiveCredentialKeys))
	for k, v := range incoming {
		out[k] = v
	}
	for _, key := range SensitiveCredentialKeys {
		if _, hasIncoming := incoming[key]; hasIncoming {
			continue
		}
		if existingVal, ok := existing[key]; ok {
			out[key] = existingVal
		}
	}
	return out
}

// ReauthPreservedConfigKeys 列出"重新授权"时必须从旧凭据继承的非敏感配置子键。
// 重新授权（OAuth 重新走一遍授权流程）只携带新 token，不带这些管理员手工配置的字段；
// 若直接覆盖会丢失账号上已配置的模型白名单 / 模型映射。
//
// 注意：这些键仅在"重新授权"语义下保留，普通编辑（PUT /:id）仍允许通过 incoming 省略来删除，
// 因此不能并入 SensitiveCredentialKeys 的通用合并逻辑。
var ReauthPreservedConfigKeys = []string{"model_mapping", "compact_model_mapping"}

// MergePreservingReauthConfig 在"重新授权"落库前，把旧凭据中的模型白名单 / 模型映射等
// 非敏感配置继承到新凭据上（incoming 未显式提供时才继承）。返回新的 map，不修改入参。
func MergePreservingReauthConfig(existing, incoming map[string]any) map[string]any {
	out := make(map[string]any, len(incoming)+len(ReauthPreservedConfigKeys))
	for k, v := range incoming {
		out[k] = v
	}
	for _, key := range ReauthPreservedConfigKeys {
		if _, hasIncoming := incoming[key]; hasIncoming {
			continue
		}
		if existingVal, ok := existing[key]; ok {
			out[key] = existingVal
		}
	}
	return out
}
