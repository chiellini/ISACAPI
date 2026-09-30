package middleware

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestValidateToolsRequestAcceptsWhitelistedToolsAndCanonicalizes(t *testing.T) {
	capabilities := service.ChatCapabilities{WebSearch: true, Image: true}
	payload := map[string]any{
		"tools": []any{
			map[string]any{"type": "function", "function": map[string]any{"name": "web_fetch", "description": "client"}},
			map[string]any{"type": "function", "function": map[string]any{"name": "generate_image"}},
			map[string]any{"type": "function", "function": map[string]any{"name": "current_time"}},
		},
		"tool_choice": "auto",
	}
	canonical, err := validateToolsRequest(payload, capabilities)
	if err != nil {
		t.Fatalf("multi-tool request rejected: %v", err)
	}
	if len(canonical) != 3 {
		t.Fatalf("canonical tool count = %d, want 3", len(canonical))
	}
	names := make([]string, 0, len(canonical))
	for _, tool := range canonical {
		function := tool.(map[string]any)["function"].(map[string]any)
		if function["description"] == "client" {
			t.Fatal("client-controlled description must be replaced")
		}
		names = append(names, function["name"].(string))
	}
	if names[0] != "web_fetch" || names[1] != "generate_image" || names[2] != "current_time" {
		t.Fatalf("unexpected canonical tool order: %v", names)
	}
}

func TestValidateToolsRequestEnforcesPerToolCapabilities(t *testing.T) {
	noCapabilities := service.ChatCapabilities{}
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{"name": "web_search"}}},
	}, noCapabilities); err == nil {
		t.Fatal("web_search must require the WebSearch capability")
	}
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{"name": "web_fetch"}}},
	}, noCapabilities); err == nil {
		t.Fatal("web_fetch must require the WebSearch capability")
	}
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{"name": "generate_image"}}},
	}, service.ChatCapabilities{WebSearch: true}); err == nil {
		t.Fatal("generate_image must require the Image capability")
	}
	// current_time 纯本地，任何档案都可用。
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{"name": "current_time"}}},
	}, noCapabilities); err != nil {
		t.Fatalf("current_time must not require capabilities: %v", err)
	}
}

func TestValidateToolsRequestRejectsUnknownOrDuplicateTools(t *testing.T) {
	capabilities := service.ChatCapabilities{WebSearch: true}
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{map[string]any{"type": "function", "function": map[string]any{"name": "rm_rf"}}},
	}, capabilities); err == nil {
		t.Fatal("unknown tool must be rejected")
	}
	if _, err := validateToolsRequest(map[string]any{
		"tools": []any{
			map[string]any{"type": "function", "function": map[string]any{"name": "web_search"}},
			map[string]any{"type": "function", "function": map[string]any{"name": "web_search"}},
		},
	}, capabilities); err == nil {
		t.Fatal("duplicate tool must be rejected")
	}
}

func TestNormalizeToolCallsValidatesArgumentsPerTool(t *testing.T) {
	capabilities := service.ChatCapabilities{WebSearch: true, Image: true}
	raw := []any{
		map[string]any{"id": "a", "type": "function", "function": map[string]any{
			"name": "web_fetch", "arguments": `{"url":"https://example.com/doc"}`,
		}},
		map[string]any{"id": "b", "type": "function", "function": map[string]any{
			"name": "generate_image", "arguments": `{"prompt":"a cat"}`,
		}},
	}
	pending := map[string]struct{}{}
	calls, _, err := normalizeToolCalls(raw, pending, capabilities)
	if err != nil {
		t.Fatalf("valid multi-tool history rejected: %v", err)
	}
	if len(calls) != 2 || len(pending) != 2 {
		t.Fatalf("unexpected normalized calls: %#v", calls)
	}
	fetch := calls[0].(map[string]any)["function"].(map[string]any)
	if fetch["arguments"] != `{"url":"https://example.com/doc"}` {
		t.Fatalf("web_fetch arguments were not canonicalized: %#v", fetch)
	}

	// 参数字段按工具校验：web_fetch 不接受 query。
	bad := []any{map[string]any{"id": "c", "type": "function", "function": map[string]any{
		"name": "web_fetch", "arguments": `{"query":"x"}`,
	}}}
	if _, _, err := normalizeToolCalls(bad, map[string]struct{}{}, capabilities); err == nil {
		t.Fatal("web_fetch must reject non-url arguments")
	}
	// 能力不匹配：无 Image 能力的档案不允许 generate_image 历史。
	if _, _, err := normalizeToolCalls(raw, map[string]struct{}{}, service.ChatCapabilities{WebSearch: true}); err == nil {
		t.Fatal("generate_image history must require the Image capability")
	}
}
