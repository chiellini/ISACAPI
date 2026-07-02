package service

import "testing"

func TestOpenAIResponseAccumulator_Deltas(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeSSE([]byte(`{"type":"response.created","response":{"id":"resp_abc","status":"in_progress"}}`))
	acc.observeSSE([]byte(`{"type":"response.output_text.delta","delta":"Hel"}`))
	acc.observeSSE([]byte(`{"type":"response.output_text.delta","delta":"lo"}`))
	acc.observeSSE([]byte(`{"type":"response.completed","response":{"id":"resp_abc","status":"completed"}}`))

	got := acc.result()
	if got.Text != "Hello" {
		t.Fatalf("text = %q, want Hello", got.Text)
	}
	if got.ResponseID != "resp_abc" {
		t.Fatalf("response id = %q", got.ResponseID)
	}
	if got.FinishReason != "completed" {
		t.Fatalf("finish reason = %q", got.FinishReason)
	}
}

func TestOpenAIResponseAccumulator_TerminalOutputFallback(t *testing.T) {
	// 无增量文本时，从终止事件的 response.output 兜底提取。
	acc := newOpenAIResponseAccumulator()
	acc.observeSSE([]byte(`{"type":"response.completed","response":{"id":"resp_x","status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"final only"}]}]}}`))
	got := acc.result()
	if got.Text != "final only" {
		t.Fatalf("text = %q, want 'final only'", got.Text)
	}
}

func TestOpenAITextFromResponseOutput(t *testing.T) {
	data := []byte(`{"response":{"output":[{"type":"message","content":[{"type":"output_text","text":"a"},{"type":"output_text","text":"b"}]}]}}`)
	if got := openAITextFromResponseOutput(data); got != "a\nb" {
		t.Fatalf("got %q", got)
	}
}

func TestOpenAIResponseAccumulator_ChatCompletions(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeChatCompletionsSSE([]byte(`{"id":"chatcmpl_1","choices":[{"delta":{"content":"Hel"}}]}`))
	acc.observeChatCompletionsSSE([]byte(`{"id":"chatcmpl_1","choices":[{"delta":{"content":"lo"},"finish_reason":"stop"}]}`))
	got := acc.result()
	if got.Text != "Hello" {
		t.Fatalf("text = %q, want Hello", got.Text)
	}
	if got.ResponseID != "chatcmpl_1" || got.FinishReason != "stop" {
		t.Fatalf("got %+v", got)
	}
}

func TestOpenAIResponseAccumulator_Anthropic(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeAnthropicSSE([]byte(`{"type":"message_start","message":{"id":"msg_1","content":[]}}`))
	acc.observeAnthropicSSE([]byte(`{"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi"}}`))
	acc.observeAnthropicSSE([]byte(`{"type":"message_delta","delta":{"stop_reason":"end_turn"}}`))
	got := acc.result()
	if got.Text != "Hi" {
		t.Fatalf("text = %q, want Hi", got.Text)
	}
	if got.ResponseID != "msg_1" || got.FinishReason != "end_turn" {
		t.Fatalf("got %+v", got)
	}
}
