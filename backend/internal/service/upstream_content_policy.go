package service

import (
	"encoding/json"
	"strings"
)

// UpstreamContentPolicyRejection is a provider's explicit, request-scoped
// safety refusal. It deliberately excludes account, entitlement, and generic
// permission failures.
type UpstreamContentPolicyRejection struct {
	Code    string
	Message string
}

// DetectUpstreamContentPolicyRejection recognizes structured policy codes and
// a narrow set of unambiguous provider messages across JSON and SSE payloads.
func DetectUpstreamContentPolicyRejection(statusCode int, responseBody []byte) (UpstreamContentPolicyRejection, bool) {
	if len(responseBody) == 0 {
		return UpstreamContentPolicyRejection{}, false
	}
	raw := strings.TrimSpace(string(responseBody))
	if raw == "" || grokAccountAccessMessage(raw) {
		return UpstreamContentPolicyRejection{}, false
	}

	var payload any
	if json.Unmarshal(responseBody, &payload) == nil {
		if grokStructuredAccountAccessMarker(payload) {
			return UpstreamContentPolicyRejection{}, false
		}
		if code := findUpstreamPolicyCode(payload); code != "" {
			return UpstreamContentPolicyRejection{Code: code, Message: findUpstreamPolicyMessage(payload)}, true
		}
	}

	if code := upstreamPolicyCodeFromText(raw); code != "" && (statusCode >= 400 || upstreamPolicyMessageIsExplicit(raw)) {
		return UpstreamContentPolicyRejection{Code: code, Message: strings.TrimSpace(ExtractUpstreamErrorMessage([]byte(raw)))}, true
	}
	if upstreamPolicyMessageIsExplicit(raw) {
		return UpstreamContentPolicyRejection{Code: "content_policy_violation", Message: strings.TrimSpace(ExtractUpstreamErrorMessage([]byte(raw)))}, true
	}
	return UpstreamContentPolicyRejection{}, false
}

func findUpstreamPolicyCode(value any) string {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			switch normalizeGrokErrorMarker(key) {
			case "code", "error_code", "type", "category", "reason":
				if marker, ok := child.(string); ok {
					if code := normalizeUpstreamPolicyCode(marker); code != "" {
						return code
					}
				}
			}
			if code := findUpstreamPolicyCode(child); code != "" {
				return code
			}
		}
	case []any:
		for _, child := range node {
			if code := findUpstreamPolicyCode(child); code != "" {
				return code
			}
		}
	}
	return ""
}

func normalizeUpstreamPolicyCode(value string) string {
	normalized := normalizeGrokErrorMarker(value)
	if isGrokContentPolicyCode(normalized) {
		return normalized
	}
	switch normalized {
	case "moderation_blocked", "safety_blocked", "safety_violation", "safety_error", "prompt_blocked":
		return normalized
	default:
		return ""
	}
}

func findUpstreamPolicyMessage(value any) string {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			switch normalizeGrokErrorMarker(key) {
			case "message", "detail", "error_description":
				if message, ok := child.(string); ok && strings.TrimSpace(message) != "" {
					return strings.TrimSpace(message)
				}
			}
		}
		for _, child := range node {
			if message := findUpstreamPolicyMessage(child); message != "" {
				return message
			}
		}
	case []any:
		for _, child := range node {
			if message := findUpstreamPolicyMessage(child); message != "" {
				return message
			}
		}
	}
	return ""
}

func upstreamPolicyCodeFromText(value string) string {
	lower := strings.ToLower(value)
	for _, code := range []string{
		"content_policy_violation",
		"moderation_blocked",
		"safety_violation",
		"safety_blocked",
		"cyber_policy",
		"content_filter",
		"content_moderation",
	} {
		if strings.Contains(lower, code) {
			return code
		}
	}
	return ""
}

func upstreamPolicyMessageIsExplicit(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, phrase := range []string{
		"blocked by content policy",
		"content policy violation",
		"content policy rejection",
		"request blocked by content moderation",
		"request rejected by content moderation",
		"request rejected by the safety system",
		"prompt violates content policy",
		"input violates content policy",
		"prohibited content",
		"forbidden content",
		"请求涉及受限话题",
		"已被内容安全策略阻止",
		"内容安全策略拦截",
	} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
