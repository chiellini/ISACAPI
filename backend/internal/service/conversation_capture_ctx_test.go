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
