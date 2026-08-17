package middleware

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestNormalizeChatCompletionPayloadKeepsOnlyTrustedInstructionsAndCanonicalSearch(t *testing.T) {
	profile := &service.ChatProfile{
		UpstreamModel: "upstream-model",
		Capabilities:  service.ChatCapabilities{WebSearch: true},
	}
	payload := map[string]any{
		"model":  "public-model",
		"stream": true,
		"messages": []any{
			map[string]any{"role": "system", "content": "client override"},
			map[string]any{"role": "developer", "content": "client override 2"},
			map[string]any{"role": "user", "content": "hello"},
		},
		"tools": []any{
			map[string]any{"type": "function", "function": map[string]any{"name": "web_search", "description": "client controlled"}},
		},
		"tool_choice": "auto",
	}
	got, err := normalizeChatCompletionPayload(payload, profile, "trusted policy")
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if got["model"] != "upstream-model" {
		t.Fatalf("model was not resolved server-side: %#v", got)
	}
	messages := got["messages"].([]any)
	if len(messages) != 4 {
		t.Fatalf("message count = %d, want 4", len(messages))
	}
	first := messages[0].(map[string]any)
	if first["role"] != "system" || first["content"] != "trusted policy" {
		t.Fatalf("trusted system message missing: %#v", first)
	}
	for _, index := range []int{1, 2, 3} {
		message := messages[index].(map[string]any)
		if message["role"] != "user" {
			t.Fatalf("untrusted privileged message was not downgraded: %#v", message)
		}
	}
	tools := got["tools"].([]any)
	function := tools[0].(map[string]any)["function"].(map[string]any)
	if function["name"] != "web_search" || function["description"] == "client controlled" || got["tool_choice"] != "auto" {
		t.Fatalf("client tools were not replaced: %#v", got)
	}
}

func TestNormalizeChatCompletionPayloadRejectsPolicyBypasses(t *testing.T) {
	profile := &service.ChatProfile{UpstreamModel: "upstream", Capabilities: service.ChatCapabilities{}}
	base := func() map[string]any {
		return map[string]any{"model": "public", "messages": []any{map[string]any{"role": "user", "content": "hello"}}}
	}
	for _, field := range []string{"web_search_options", "service_tier", "metadata", "store", "reasoning_effort", "moderation"} {
		payload := base()
		payload[field] = map[string]any{"enabled": true}
		if _, err := normalizeChatCompletionPayload(payload, profile, "trusted"); err == nil {
			t.Fatalf("expected unknown field %q to be rejected", field)
		}
	}
	payload := base()
	payload["messages"] = []any{map[string]any{
		"role":    "user",
		"content": []any{map[string]any{"type": "input_image", "image_url": "https://example.test/a.png"}},
	}}
	if _, err := normalizeChatCompletionPayload(payload, profile, ""); err == nil {
		t.Fatal("expected provider-native image content to be rejected")
	}
}

func TestNormalizeChatCompletionPayloadEnforcesVisionToolHistoryAndContextLimit(t *testing.T) {
	profile := &service.ChatProfile{
		UpstreamModel: "upstream",
		Capabilities:  service.ChatCapabilities{Vision: true, WebSearch: true, ContextLimit: 1000},
	}
	payload := map[string]any{
		"messages": []any{
			map[string]any{"role": "user", "content": []any{
				map[string]any{"type": "text", "text": "find it"},
				map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64,AA=="}},
			}},
			map[string]any{"role": "assistant", "content": "", "tool_calls": []any{
				map[string]any{"id": "call_1", "type": "function", "function": map[string]any{"name": "web_search", "arguments": `{"query":"current facts"}`}},
			}},
			map[string]any{"role": "tool", "tool_call_id": "call_1", "content": `[{"title":"Result"}]`},
		},
	}
	if _, err := normalizeChatCompletionPayload(payload, profile, "trusted"); err != nil {
		t.Fatalf("valid multimodal/tool history rejected: %v", err)
	}
	profile.Capabilities.ContextLimit = 5
	if _, err := normalizeChatCompletionPayload(payload, profile, "trusted"); err == nil {
		t.Fatal("expected configured Unicode character limit to be enforced")
	}
}

func TestNormalizeImageGenerationPayloadAppliesPolicyAndRejectsCostSafetyControls(t *testing.T) {
	profile := &service.ChatProfile{
		UpstreamModel: "gpt-image-upstream",
		Capabilities:  service.ChatCapabilities{Image: true, ContextLimit: 100},
	}
	got, err := normalizeImageGenerationPayload(map[string]any{
		"model": "image-public", "prompt": "draw a cat", "stream": true, "response_format": "b64_json",
	}, profile, "trusted image policy")
	if err != nil {
		t.Fatalf("normalize image failed: %v", err)
	}
	if got["model"] != "gpt-image-upstream" || got["prompt"] != "trusted image policy\n\nUser image request:\ndraw a cat" {
		t.Fatalf("image model or trusted prompt was not enforced: %#v", got)
	}
	for _, field := range []string{"moderation", "n", "quality", "size", "output_compression"} {
		payload := map[string]any{"prompt": "draw", field: "low"}
		if _, err := normalizeImageGenerationPayload(payload, profile, "trusted"); err == nil {
			t.Fatalf("expected image field %q to be rejected", field)
		}
	}
}

func TestAvailableChatProfileModelsFiltersGroupAccess(t *testing.T) {
	policy := &service.ChatPolicy{Enabled: true, Profiles: []service.ChatProfile{
		{ID: "gpt", Name: "GPT", Provider: service.PlatformOpenAI, PublicModel: "gpt", GroupID: 1, Enabled: true},
		{ID: "claude", Name: "Claude", Provider: service.PlatformAnthropic, PublicModel: "claude", GroupID: 2, Enabled: true},
	}}
	models := AvailableChatProfileModels(policy, []service.Group{{ID: 2, Platform: service.PlatformAnthropic}})
	if len(models) != 1 || models[0]["id"] != "claude" {
		t.Fatalf("unexpected filtered models: %#v", models)
	}
}

func TestChatPayloadContainsImage(t *testing.T) {
	messages := []any{map[string]any{
		"role": "user",
		"content": []any{
			map[string]any{"type": "text", "text": "look"},
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64,AA=="}},
		},
	}}
	if !chatPayloadContainsImage(messages) {
		t.Fatal("expected image_url content to be detected")
	}
	for _, part := range []map[string]any{
		{"type": "input_image", "image_url": "data:image/png;base64,AA=="},
		{"type": "image", "source": map[string]any{"data": "AA=="}},
		{"type": "text", "image_url": "https://example.test/image.png"},
	} {
		candidate := []any{map[string]any{"role": "user", "content": []any{part}}}
		if !chatPayloadContainsImage(candidate) {
			t.Fatalf("expected provider-specific or malformed image block to be detected: %#v", part)
		}
	}
	if chatPayloadContainsImage([]any{map[string]any{"role": "user", "content": "text only"}}) {
		t.Fatal("text-only messages must not be treated as image input")
	}
}
