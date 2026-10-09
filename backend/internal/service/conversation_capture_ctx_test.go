package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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

func TestOpenAIResponseAccumulator_EventHeaderAndJSONCapture(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeSSEWithType([]byte(`{"delta":"captured from event header"}`), "response.output_text.delta")
	acc.observeSSEWithType([]byte(`{"response":{"id":"resp_header","status":"completed"}}`), "response.completed")

	got := acc.result()
	if got.Text != "captured from event header" || got.ResponseID != "resp_header" {
		t.Fatalf("header event capture = %+v", got)
	}
}

func TestCaptureOpenAIResponseFromJSON_ReplacesPreviousAttempt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	SetOpenAICapturedResponse(c, OpenAICapturedResponse{Text: "failed attempt"})

	captureOpenAIResponseFromJSON(c, []byte(`{"id":"resp_ok","status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"final answer"}]}]}`))
	got, ok := GetOpenAICapturedResponse(c)
	if !ok || got.Text != "final answer" || got.ResponseID != "resp_ok" {
		t.Fatalf("captured response = %+v, ok=%v", got, ok)
	}

	captureOpenAIResponseFromJSON(c, []byte(`{"id":"resp_empty","status":"completed","output":[]}`))
	got, ok = GetOpenAICapturedResponse(c)
	if !ok || got.Text != "" || got.ResponseID != "resp_empty" {
		t.Fatalf("empty successful response did not replace prior attempt: %+v, ok=%v", got, ok)
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

func TestOpenAIResponseAccumulator_ReasoningDeltas(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeSSEWithType([]byte(`{"delta":"thin"}`), "response.reasoning_summary_text.delta")
	acc.observeSSEWithType([]byte(`{"delta":"king"}`), "response.reasoning_summary_text.delta")
	acc.observeSSEWithType([]byte(`{"delta":" rd"}`), "response.reasoning_text.delta")
	acc.observeSSE([]byte(`{"type":"response.output_text.delta","delta":"answer"}`))
	got := acc.result()
	if got.Thinking != "thinking rd" || got.Text != "answer" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestOpenAIResponseAccumulator_ReasoningTerminalFallback(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeSSE([]byte(`{"type":"response.completed","response":{"id":"resp_t","status":"completed","output":[{"type":"reasoning","summary":[{"type":"summary_text","text":"because"}]},{"type":"message","content":[{"type":"output_text","text":"ok"}]}]}}`))
	got := acc.result()
	if got.Thinking != "because" || got.Text != "ok" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestOpenAIResponseAccumulator_ChatCompletionsReasoning(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeChatCompletionsSSE([]byte(`{"id":"c1","choices":[{"delta":{"reasoning_content":"deep"}}]}`))
	acc.observeChatCompletionsSSE([]byte(`{"id":"c1","choices":[{"delta":{"content":"ans"},"finish_reason":"stop"}]}`))
	got := acc.result()
	if got.Thinking != "deep" || got.Text != "ans" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestOpenAIResponseAccumulator_AnthropicThinking(t *testing.T) {
	acc := newOpenAIResponseAccumulator()
	acc.observeAnthropicSSE([]byte(`{"type":"content_block_delta","delta":{"type":"thinking_delta","thinking":"hmm"}}`))
	acc.observeAnthropicSSE([]byte(`{"type":"content_block_delta","delta":{"type":"text_delta","text":"done"}}`))
	got := acc.result()
	if got.Thinking != "hmm" || got.Text != "done" {
		t.Fatalf("captured = %+v", got)
	}
}

func TestCaptureOpenAIResponseFromJSON_Reasoning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	captureOpenAIResponseFromJSON(c, []byte(`{"id":"resp_ok","status":"completed","output":[{"type":"reasoning","summary":[{"type":"summary_text","text":"chain"}]},{"type":"message","content":[{"type":"output_text","text":"final answer"}]}]}`))
	got, ok := GetOpenAICapturedResponse(c)
	if !ok || got.Text != "final answer" || got.Thinking != "chain" {
		t.Fatalf("captured = %+v, ok=%v", got, ok)
	}
}

func TestCaptureOpenAIChatCompletionsResponseFromJSON_Reasoning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	captureOpenAIChatCompletionsResponseFromJSON(c, []byte(`{"id":"c9","choices":[{"message":{"content":"ans","reasoning_content":"why"},"finish_reason":"stop"}]}`))
	got, ok := GetOpenAICapturedResponse(c)
	if !ok || got.Text != "ans" || got.Thinking != "why" {
		t.Fatalf("captured = %+v, ok=%v", got, ok)
	}
}

func TestCaptureGeminiResponseFromJSON_ThoughtSplit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	captureGeminiResponseFromJSON(c, []byte(`{"candidates":[{"content":{"parts":[{"text":"pondering","thought":true},{"text":"answer"}]},"finishReason":"STOP"}]}`))
	got, ok := GetOpenAICapturedResponse(c)
	if !ok || got.Text != "answer" || got.Thinking != "pondering" {
		t.Fatalf("captured = %+v, ok=%v", got, ok)
	}
}
