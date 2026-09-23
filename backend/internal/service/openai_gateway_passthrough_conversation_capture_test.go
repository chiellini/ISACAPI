package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func passthroughConversationCaptureContext(t *testing.T, enabled bool) (*OpenAIGatewayService, *gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	cfg := &config.Config{}
	cfg.ConversationArchive.Enabled = enabled
	return &OpenAIGatewayService{cfg: cfg}, c, recorder
}

func passthroughConversationCaptureResponse(contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestPassthroughConversationCaptureStreaming(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		text string
	}{
		{
			name: "deltas",
			body: "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_capture\",\"status\":\"in_progress\"}}\n\n" +
				"data: {\"type\":\"response.output_text.delta\",\"delta\":\"Hello \"}\n\n" +
				"data: {\"type\":\"response.output_text.delta\",\"delta\":\"world\"}\n\n" +
				"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_capture\",\"status\":\"completed\",\"usage\":{\"input_tokens\":2,\"output_tokens\":3}}}\n\n",
			text: "Hello world",
		},
		{
			name: "event headers without payload type",
			body: "event: response.output_text.delta\ndata: {\"delta\":\"Header event\"}\n\n" +
				"event: response.completed\ndata: {\"response\":{\"id\":\"resp_capture\",\"status\":\"completed\",\"usage\":{\"input_tokens\":2,\"output_tokens\":3}}}\n\n",
			text: "Header event",
		},
		{
			name: "terminal output without deltas",
			body: "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_capture\",\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"content\":[{\"type\":\"output_text\",\"text\":\"Final output\"}]}],\"usage\":{\"input_tokens\":2,\"output_tokens\":3}}}\n\n",
			text: "Final output",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, enabled := range []bool{false, true} {
				svc, c, recorder := passthroughConversationCaptureContext(t, enabled)
				resp := passthroughConversationCaptureResponse("text/event-stream", tc.body)
				result, err := svc.handleStreamingResponsePassthrough(context.Background(), resp, c, &Account{ID: 1, Platform: PlatformOpenAI}, time.Now(), "", "")
				require.NoError(t, err)
				require.Equal(t, tc.body, recorder.Body.String())
				require.Equal(t, 2, result.usage.InputTokens)
				require.Equal(t, 3, result.usage.OutputTokens)
				captured, ok := GetOpenAICapturedResponse(c)
				require.Equal(t, enabled, ok)
				if enabled {
					require.Equal(t, OpenAICapturedResponse{Text: tc.text, ResponseID: "resp_capture", FinishReason: "completed"}, captured)
				}
			}
		})
	}
}

func TestPassthroughConversationCaptureDrainsAfterClientDisconnect(t *testing.T) {
	svc, c, recorder := passthroughConversationCaptureContext(t, true)
	writer := &passthroughFlushTestWriter{ResponseWriter: c.Writer, recorder: recorder, failAfterWrites: 2}
	c.Writer = writer
	firstOutput := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"First\"}\n\n"
	body := firstOutput +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\" and remaining\"}\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_drain_capture\",\"status\":\"completed\",\"usage\":{\"input_tokens\":7,\"output_tokens\":5}}}\n\n"
	resp := passthroughConversationCaptureResponse("text/event-stream", body)
	result, err := svc.handleStreamingResponsePassthrough(context.Background(), resp, c, &Account{ID: 1, Platform: PlatformOpenAI}, time.Now(), "", "")
	require.NoError(t, err)
	require.Equal(t, firstOutput, recorder.Body.String())
	require.Equal(t, 1, writer.failedWrites)
	require.Equal(t, 7, result.usage.InputTokens)
	require.Equal(t, 5, result.usage.OutputTokens)
	captured, ok := GetOpenAICapturedResponse(c)
	require.True(t, ok)
	require.Equal(t, OpenAICapturedResponse{Text: "First and remaining", ResponseID: "resp_drain_capture", FinishReason: "completed"}, captured)
}

func TestPassthroughConversationCaptureRetryReplacesFailedAttempt(t *testing.T) {
	svc, c, recorder := passthroughConversationCaptureContext(t, true)
	account := &Account{ID: 1, Platform: PlatformOpenAI}
	failedBody := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_failed_attempt\",\"status\":\"in_progress\"}}\n\n" +
		"data: {\"type\":\"response.failed\",\"error\":{\"code\":\"server_error\",\"message\":\"upstream processing failed\"}}\n\n"
	_, err := svc.handleStreamingResponsePassthrough(context.Background(), passthroughConversationCaptureResponse("text/event-stream", failedBody), c, account, time.Now(), "", "")
	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Empty(t, recorder.Body.String())

	retryBody := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Retried output\"}\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_retry\",\"status\":\"completed\",\"usage\":{\"input_tokens\":2,\"output_tokens\":3}}}\n\n"
	_, err = svc.handleStreamingResponsePassthrough(context.Background(), passthroughConversationCaptureResponse("text/event-stream", retryBody), c, account, time.Now(), "", "")
	require.NoError(t, err)
	require.Equal(t, retryBody, recorder.Body.String())
	captured, ok := GetOpenAICapturedResponse(c)
	require.True(t, ok)
	require.Equal(t, OpenAICapturedResponse{Text: "Retried output", ResponseID: "resp_retry", FinishReason: "completed"}, captured)
}

func TestPassthroughConversationCaptureNonStreaming(t *testing.T) {
	for _, tc := range []struct {
		name        string
		contentType string
		body        string
		text        string
		responseID  string
		status      string
	}{
		{
			name:        "JSON response",
			contentType: "application/json",
			body:        `{"id":"resp_json_capture","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"JSON output"}]}],"usage":{"input_tokens":2,"output_tokens":3}}`,
			text:        "JSON output",
			responseID:  "resp_json_capture",
			status:      "completed",
		},
		{
			name:        "SSE converted to JSON with reconstructed output",
			contentType: "text/event-stream",
			body: "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Reconstructed output\"}\n\n" +
				"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_sse_capture\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":2,\"output_tokens\":3}}}\n\n",
			text:       "Reconstructed output",
			responseID: "resp_sse_capture",
			status:     "completed",
		},
		{
			name:        "SSE fallback without terminal response",
			contentType: "text/event-stream",
			body: "event: response.created\ndata: {\"response\":{\"id\":\"resp_partial_capture\",\"status\":\"in_progress\"}}\n\n" +
				"event: response.output_text.delta\ndata: {\"delta\":\"Partial output\"}\n\n",
			text:       "Partial output",
			responseID: "resp_partial_capture",
			status:     "in_progress",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var baselineBody string
			var baselineResult *openaiNonStreamingResultPassthrough
			for _, enabled := range []bool{false, true} {
				svc, c, recorder := passthroughConversationCaptureContext(t, enabled)
				resp := passthroughConversationCaptureResponse(tc.contentType, tc.body)
				result, err := svc.handleNonStreamingResponsePassthrough(context.Background(), resp, c, &Account{ID: 1, Platform: PlatformOpenAI}, "", "")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, recorder.Code)
				captured, ok := GetOpenAICapturedResponse(c)
				require.Equal(t, enabled, ok)
				if enabled {
					require.Equal(t, OpenAICapturedResponse{Text: tc.text, ResponseID: tc.responseID, FinishReason: tc.status}, captured)
					require.Equal(t, baselineBody, recorder.Body.String())
					require.Equal(t, baselineResult, result)
				} else {
					baselineBody = recorder.Body.String()
					baselineResult = result
				}
			}
		})
	}
}
