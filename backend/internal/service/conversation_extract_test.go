package service

import "testing"

func TestExtractOpenAIResponsesRequest(t *testing.T) {
	body := []byte(`{
		"model": "gpt-5",
		"previous_response_id": "resp_prev",
		"prompt_cache_key": "pck-1",
		"instructions": "you are helpful",
		"input": [
			{"role":"user","content":[{"type":"input_text","text":"old question"}]},
			{"role":"assistant","content":[{"type":"output_text","text":"old answer"}]},
			{"role":"user","content":[{"type":"input_text","text":"new question"}]}
		]
	}`)
	got := ExtractOpenAIResponsesRequest(body)
	if got.Model != "gpt-5" {
		t.Fatalf("model = %q", got.Model)
	}
	if got.Signals.PreviousResponseID != "resp_prev" || got.Signals.PromptCacheKey != "pck-1" {
		t.Fatalf("signals = %+v", got.Signals)
	}
	// 只取最新一条 user 消息，不含历史。
	if len(got.UserEvents) != 1 || got.UserEvents[0].Content != "new question" {
		t.Fatalf("user events = %+v", got.UserEvents)
	}
}

func TestExtractOpenAIResponsesRequest_StringInput(t *testing.T) {
	got := ExtractOpenAIResponsesRequest([]byte(`{"model":"gpt-5","input":"just a string"}`))
	if len(got.UserEvents) != 1 || got.UserEvents[0].Content != "just a string" {
		t.Fatalf("user events = %+v", got.UserEvents)
	}
}

func TestExtractOpenAIResponsesResponse(t *testing.T) {
	body := []byte(`{
		"id":"resp_123","model":"gpt-5","status":"completed",
		"output":[{"type":"message","content":[{"type":"output_text","text":"hello world"}]}],
		"usage":{"input_tokens":10,"output_tokens":5}
	}`)
	got := ExtractOpenAIResponsesResponse(body)
	if got.ResponseID != "resp_123" || got.InputTokens != 10 || got.OutputTokens != 5 {
		t.Fatalf("got %+v", got)
	}
	if len(got.AssistantEvents) != 1 || got.AssistantEvents[0].Content != "hello world" {
		t.Fatalf("assistant events = %+v", got.AssistantEvents)
	}
}

func TestExtractAnthropicMessagesRequest(t *testing.T) {
	body := []byte(`{
		"model":"claude-x",
		"system":"sys prompt",
		"metadata":{"user_id":"user_session_abc"},
		"messages":[
			{"role":"user","content":[{"type":"text","text":"first"}]},
			{"role":"assistant","content":[{"type":"text","text":"reply"}]},
			{"role":"user","content":"latest user"}
		]
	}`)
	got := ExtractAnthropicMessagesRequest(body)
	if got.Model != "claude-x" {
		t.Fatalf("model = %q", got.Model)
	}
	if got.Signals.MetadataNativeID != "user_session_abc" {
		t.Fatalf("native id = %q", got.Signals.MetadataNativeID)
	}
	// system + 最新 user
	var roles []string
	for _, e := range got.UserEvents {
		roles = append(roles, e.Role+":"+e.Content)
	}
	if len(got.UserEvents) != 2 || got.UserEvents[1].Content != "latest user" {
		t.Fatalf("events = %v", roles)
	}
}

func TestExtractAnthropicMessagesResponse(t *testing.T) {
	body := []byte(`{
		"id":"msg_1","model":"claude-x","stop_reason":"end_turn",
		"content":[{"type":"text","text":"hi"},{"type":"text","text":"there"}],
		"usage":{"input_tokens":3,"output_tokens":2}
	}`)
	got := ExtractAnthropicMessagesResponse(body)
	if got.ResponseID != "msg_1" || got.FinishReason != "end_turn" {
		t.Fatalf("got %+v", got)
	}
	if len(got.AssistantEvents) != 1 || got.AssistantEvents[0].Content != "hi\nthere" {
		t.Fatalf("assistant = %+v", got.AssistantEvents)
	}
}

func TestExtract_GarbageDoesNotPanic(t *testing.T) {
	_ = ExtractOpenAIResponsesRequest([]byte(`not json`))
	_ = ExtractOpenAIResponsesResponse([]byte(``))
	_ = ExtractAnthropicMessagesRequest([]byte(`{`))
	_ = ExtractAnthropicMessagesResponse([]byte(`[]`))
}
