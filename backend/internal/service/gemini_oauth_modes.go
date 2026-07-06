package service

import (
	"fmt"
	"strings"
)

func isGeminiCodeAssistScopedOAuth(account *Account) bool {
	if account == nil || account.Platform != PlatformGemini || account.Type != AccountTypeOAuth {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(account.GeminiOAuthType())) {
	case "code_assist", "google_one":
		return true
	default:
		return false
	}
}

func geminiCodeAssistOAuthRequiresProjectIDError(account *Account) error {
	label := "Gemini Code Assist"
	if account != nil && strings.EqualFold(strings.TrimSpace(account.GeminiOAuthType()), "google_one") {
		label = "Google One"
	}
	return fmt.Errorf("%s OAuth token uses Gemini CLI / Code Assist scopes and cannot call generativelanguage.googleapis.com directly without project_id; add a Code Assist Project ID and re-authorize this account, or use an AI Studio OAuth/API-key account", label)
}
