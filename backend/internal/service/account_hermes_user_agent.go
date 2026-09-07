package service

import (
	"net/http"
	"strings"
)

// HermesUserAgentExtraKey is the per-account switch stored in accounts.extra.
// It is intentionally scoped to OpenAI API-key accounts; OAuth/Codex accounts
// must keep their protocol-specific identity headers intact.
const HermesUserAgentExtraKey = "hermes_user_agent_enabled"

// HermesUserAgentValue is the fixed User-Agent sent when the account switch is
// enabled.  Keeping this value constant makes the setting safe to expose as a
// boolean toggle rather than an arbitrary header-injection mechanism.
const HermesUserAgentValue = "hermes-agent/1.0"

// IsHermesUserAgentEnabled reports whether this account opts into the fixed
// Hermes outbound identity.  Malformed/missing values are treated as disabled
// and the guard deliberately excludes OAuth, setup-token, and non-OpenAI
// accounts so their authentication identity cannot be changed accidentally.
func (a *Account) IsHermesUserAgentEnabled() bool {
	if a == nil || !a.IsOpenAIApiKey() || a.Extra == nil {
		return false
	}
	enabled, ok := a.Extra[HermesUserAgentExtraKey].(bool)
	return ok && enabled
}

// ApplyHermesUserAgent applies the fixed Hermes User-Agent to an outbound
// request when enabled for this account.  Remove all casing variants first:
// requests are assembled from several header sources and Go permits raw map
// keys such as "user-agent" to coexist with the canonical "User-Agent" key.
func (a *Account) ApplyHermesUserAgent(headers http.Header) {
	if headers == nil || !a.IsHermesUserAgentEnabled() {
		return
	}
	for key := range headers {
		if strings.EqualFold(key, "User-Agent") {
			delete(headers, key)
		}
	}
	headers.Set("User-Agent", HermesUserAgentValue)
}
