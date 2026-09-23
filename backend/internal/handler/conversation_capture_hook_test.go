package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayHandlerCaptureResponsesConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		enabled        bool
		skip           bool
		wantSubmission bool
	}{
		{name: "captures request and upstream response", enabled: true, wantSubmission: true},
		{name: "archive disabled", wantSubmission: false},
		{name: "capture explicitly skipped", enabled: true, skip: true, wantSubmission: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"gpt-5","input":[{"role":"user","content":"question"}]}`))
			if tt.skip {
				c.Set(CtxSkipConversationCapture, true)
			}

			sink := &service.MemoryCaptureSink{}
			h := &OpenAIGatewayHandler{
				cfg: &config.Config{
					ConversationArchive: config.ConversationArchiveConfig{Enabled: tt.enabled},
				},
				captureSink: sink,
			}
			groupID := int64(23)
			apiKey := &service.APIKey{ID: 41, GroupID: &groupID}
			service.SetOpenAICapturedResponse(c, service.OpenAICapturedResponse{
				Text:         "assistant answer",
				ResponseID:   "resp_captured",
				FinishReason: "completed",
			})

			h.captureResponsesConversation(c, []byte(`{"model":"gpt-5","input":[{"role":"user","content":"question"}]}`), &service.OpenAIForwardResult{
				RequestID: "req_123",
				Model:     "gpt-5",
				Usage: service.OpenAIUsage{
					InputTokens:  17,
					OutputTokens: 9,
				},
			}, apiKey, 7)

			records := sink.Records()
			if !tt.wantSubmission {
				require.Empty(t, records)
				return
			}

			require.Len(t, records, 1)
			record := records[0]
			require.Equal(t, "req_123", record.Request.RequestID)
			require.Equal(t, int64(7), record.Request.UserID)
			require.Len(t, record.Request.Events, 1)
			require.Equal(t, service.ConversationRoleUser, record.Request.Events[0].Role)
			require.Equal(t, "question", record.Request.Events[0].Content)
			require.Equal(t, "resp_captured", record.Response.ResponseID)
			require.Equal(t, int64(17), record.Response.InputTokens)
			require.Equal(t, int64(9), record.Response.OutputTokens)
			require.Len(t, record.Response.Events, 1)
			require.Equal(t, service.ConversationRoleAssistant, record.Response.Events[0].Role)
			require.Equal(t, "assistant answer", record.Response.Events[0].Content)
		})
	}
}
