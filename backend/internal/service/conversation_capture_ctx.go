package service

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// openAICapturedResponseKey 是暂存采集到的 assistant 响应的 gin.Context 键。
const openAICapturedResponseKey = "conv_openai_captured_response"

// OpenAICapturedResponse 是从响应（流式聚合或非流式 JSON）提取的 assistant 输出。
// 仅用于对话存档旁路；与转发/计费无关。
type OpenAICapturedResponse struct {
	Text         string
	ResponseID   string
	FinishReason string
}

// conversationCaptureEnabled 报告是否启用对话存档（用于在转发路径上廉价短路）。
func (s *OpenAIGatewayService) conversationCaptureEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.ConversationArchive.Enabled
}

// openAIResponseAccumulator 在流式转发时旁路聚合 assistant 文本与 response id。
// 仅追加，不影响向客户端的写出。
type openAIResponseAccumulator struct {
	text         strings.Builder
	responseID   string
	finishReason string
}

func newOpenAIResponseAccumulator() *openAIResponseAccumulator {
	return &openAIResponseAccumulator{}
}

// observeSSE 观察一条 SSE data 负载，累积输出文本并捕获 response id/状态。
func (a *openAIResponseAccumulator) observeSSE(data []byte) {
	if a == nil || len(data) == 0 {
		return
	}
	switch strings.TrimSpace(gjson.GetBytes(data, "type").String()) {
	case "response.output_text.delta":
		if d := gjson.GetBytes(data, "delta").String(); d != "" {
			a.text.WriteString(d)
		}
	case "response.created", "response.in_progress", "response.completed", "response.done", "response.incomplete":
		if id := gjson.GetBytes(data, "response.id").String(); id != "" {
			a.responseID = id
		}
		if st := gjson.GetBytes(data, "response.status").String(); st != "" {
			a.finishReason = st
		}
		// 终止事件若无增量文本（罕见），从最终 response.output 兜底提取。
		if a.text.Len() == 0 {
			if text := openAITextFromResponseOutput(data); text != "" {
				a.text.WriteString(text)
			}
		}
	}
}

func (a *openAIResponseAccumulator) result() OpenAICapturedResponse {
	if a == nil {
		return OpenAICapturedResponse{}
	}
	return OpenAICapturedResponse{
		Text:         strings.TrimSpace(a.text.String()),
		ResponseID:   a.responseID,
		FinishReason: a.finishReason,
	}
}

// openAITextFromResponseOutput 从 {response:{output:[{content:[{type:output_text,text}]}]}} 提取文本。
func openAITextFromResponseOutput(data []byte) string {
	var b strings.Builder
	gjson.GetBytes(data, "response.output").ForEach(func(_, item gjson.Result) bool {
		item.Get("content").ForEach(func(_, c gjson.Result) bool {
			if strings.Contains(c.Get("type").String(), "text") {
				if t := c.Get("text").String(); t != "" {
					if b.Len() > 0 {
						b.WriteString("\n")
					}
					b.WriteString(t)
				}
			}
			return true
		})
		return true
	})
	return strings.TrimSpace(b.String())
}

// SetOpenAICapturedResponseAccumulator 在流式路径上暂存聚合器指针（同一请求 goroutine 上稍后读取其结果）。
func SetOpenAICapturedResponseAccumulator(c *gin.Context, acc *openAIResponseAccumulator) {
	if c == nil || acc == nil {
		return
	}
	c.Set(openAICapturedResponseKey, acc)
}

// SetOpenAICapturedResponse 在非流式路径上直接暂存已提取的结果。
func SetOpenAICapturedResponse(c *gin.Context, r OpenAICapturedResponse) {
	if c == nil {
		return
	}
	c.Set(openAICapturedResponseKey, r)
}

// captureOpenAIResponseFromJSON 从非流式最终 JSON 响应体提取 assistant 输出并暂存进 context。
func captureOpenAIResponseFromJSON(c *gin.Context, body []byte) {
	ext := ExtractOpenAIResponsesResponse(body)
	text := ""
	if len(ext.AssistantEvents) > 0 {
		text = ext.AssistantEvents[0].Content
	}
	if text == "" && ext.ResponseID == "" {
		return
	}
	SetOpenAICapturedResponse(c, OpenAICapturedResponse{
		Text:         text,
		ResponseID:   ext.ResponseID,
		FinishReason: ext.FinishReason,
	})
}

// GetOpenAICapturedResponse 取出采集到的 assistant 响应（兼容流式聚合器与非流式结果）。
func GetOpenAICapturedResponse(c *gin.Context) (OpenAICapturedResponse, bool) {
	if c == nil {
		return OpenAICapturedResponse{}, false
	}
	v, ok := c.Get(openAICapturedResponseKey)
	if !ok {
		return OpenAICapturedResponse{}, false
	}
	switch t := v.(type) {
	case *openAIResponseAccumulator:
		return t.result(), true
	case OpenAICapturedResponse:
		return t, true
	}
	return OpenAICapturedResponse{}, false
}
