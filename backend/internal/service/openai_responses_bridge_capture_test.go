//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResponsesChatFallbackCapturesAssistantText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "json", true: "sse"}[stream], func(t *testing.T) {
			cfg := &config.Config{}
			cfg.ConversationArchive.Enabled = true
			svc := &OpenAIGatewayService{cfg: cfg}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

			var resp *http.Response
			if stream {
				resp = &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(strings.Join([]string{
					`data: {"id":"chatcmpl_capture","choices":[{"index":0,"delta":{"content":"captured"},"finish_reason":null}]}`,
					"",
					`data: {"id":"chatcmpl_capture","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
					"",
					"data: [DONE]",
					"",
				}, "\n")))}
				_, err := svc.streamChatCompletionsAsResponses(c, resp, "gpt-5.5", nil, nil, false, nil, "gpt-5.5", "gpt-5.5", nil, nil, time.Now())
				require.NoError(t, err)
			} else {
				resp = &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"id":"chatcmpl_capture","choices":[{"index":0,"message":{"role":"assistant","content":"captured"},"finish_reason":"stop"}]}`))}
				_, err := svc.bufferChatCompletionsAsResponses(c, resp, "gpt-5.5", nil, nil, false, nil, "gpt-5.5", "gpt-5.5", nil, nil, time.Now())
				require.NoError(t, err)
			}
			captured, ok := GetOpenAICapturedResponse(c)
			require.True(t, ok)
			require.Equal(t, "captured", captured.Text)
		})
	}
}

func TestResponsesNativeAnthropicCapturesAssistantText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "buffered", true: "streaming"}[stream], func(t *testing.T) {
			svc := newNativeAnthropicHangTestService(5)
			svc.cfg.ConversationArchive.Enabled = true
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(miniAnthropicSSEStream()))}

			if stream {
				_, err := svc.handleResponsesStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
				require.NoError(t, err)
			} else {
				_, err := svc.handleResponsesBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
				require.NoError(t, err)
			}
			captured, ok := GetOpenAICapturedResponse(c)
			require.True(t, ok)
			require.Equal(t, "Hello", captured.Text)
		})
	}
}

func TestResponsesWSV2CapturesAssistantText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
	c.Request.Header.Set("User-Agent", "unit-test-agent/1.0")

	cfg := newOpenAIWSV2TestConfig()
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 8
	cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 5
	cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3
	cfg.ConversationArchive.Enabled = true
	conn := &openAIWSCaptureConn{events: [][]byte{
		[]byte(`{"type":"response.created","response":{"id":"resp_capture"}}`),
		[]byte(`{"type":"response.output_text.delta","response_id":"resp_capture","delta":"captured"}`),
		[]byte(`{"type":"response.completed","response":{"id":"resp_capture","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"captured"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`),
	}}
	pool := newOpenAIWSConnPool(cfg)
	pool.setClientDialerForTest(&openAIWSCaptureDialer{conn: conn})
	svc := &OpenAIGatewayService{
		cfg:              cfg,
		httpUpstream:     &httpUpstreamRecorder{},
		cache:            &stubGatewayCache{},
		openaiWSResolver: NewOpenAIWSProtocolResolver(cfg),
		toolCorrector:    NewCodexToolCorrector(),
		openaiWSPool:     pool,
	}
	account := &Account{
		ID:          9001,
		Name:        "openai-ws-capture",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra:       map[string]any{"responses_websockets_v2_enabled": true},
	}

	result, err := svc.Forward(context.Background(), c, account, []byte(`{"model":"gpt-5.5","stream":false,"input":"hello"}`))
	require.NoError(t, err)
	require.NotNil(t, result)
	captured, ok := GetOpenAICapturedResponse(c)
	require.True(t, ok)
	require.Equal(t, "captured", captured.Text)
}
