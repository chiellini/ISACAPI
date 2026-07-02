package service

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const openAICapturedResponseKey = "conv_openai_captured_response"

// OpenAICapturedResponse is assistant output captured on the forwarding path.
// It is used only by conversation archival and does not affect billing.
type OpenAICapturedResponse struct {
	Text         string
	ResponseID   string
	FinishReason string
}

func (s *OpenAIGatewayService) conversationCaptureEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.ConversationArchive.Enabled
}

type openAIResponseAccumulator struct {
	text         strings.Builder
	responseID   string
	finishReason string
}

func newOpenAIResponseAccumulator() *openAIResponseAccumulator {
	return &openAIResponseAccumulator{}
}

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
		if a.text.Len() == 0 {
			if text := openAITextFromResponseOutput(data); text != "" {
				a.text.WriteString(text)
			}
		}
	}
}

func (a *openAIResponseAccumulator) observeChatCompletionsSSE(data []byte) {
	if a == nil || len(data) == 0 {
		return
	}
	if id := gjson.GetBytes(data, "id").String(); id != "" {
		a.responseID = id
	}
	gjson.GetBytes(data, "choices").ForEach(func(_, choice gjson.Result) bool {
		if d := choice.Get("delta.content").String(); d != "" {
			a.text.WriteString(d)
		}
		if t := choice.Get("message.content").String(); t != "" && a.text.Len() == 0 {
			a.text.WriteString(t)
		}
		if fr := choice.Get("finish_reason").String(); fr != "" {
			a.finishReason = fr
		}
		return true
	})
}

func (a *openAIResponseAccumulator) observeAnthropicSSE(data []byte) {
	if a == nil || len(data) == 0 {
		return
	}
	switch strings.TrimSpace(gjson.GetBytes(data, "type").String()) {
	case "message_start":
		if id := gjson.GetBytes(data, "message.id").String(); id != "" {
			a.responseID = id
		}
		if a.text.Len() == 0 {
			if text := anthropicTextFromContent(gjson.GetBytes(data, "message.content")); text != "" {
				a.text.WriteString(text)
			}
		}
	case "content_block_start":
		if text := gjson.GetBytes(data, "content_block.text").String(); text != "" {
			a.text.WriteString(text)
		}
	case "content_block_delta":
		if text := gjson.GetBytes(data, "delta.text").String(); text != "" {
			a.text.WriteString(text)
		}
	case "message_delta":
		if fr := gjson.GetBytes(data, "delta.stop_reason").String(); fr != "" {
			a.finishReason = fr
		}
	case "message_stop":
		if a.finishReason == "" {
			a.finishReason = "end_turn"
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

func anthropicTextFromContent(content gjson.Result) string {
	var b strings.Builder
	content.ForEach(func(_, c gjson.Result) bool {
		if c.Get("type").String() != "text" {
			return true
		}
		if t := c.Get("text").String(); t != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(t)
		}
		return true
	})
	return strings.TrimSpace(b.String())
}

func SetOpenAICapturedResponseAccumulator(c *gin.Context, acc *openAIResponseAccumulator) {
	if c == nil || acc == nil {
		return
	}
	c.Set(openAICapturedResponseKey, acc)
}

func SetOpenAICapturedResponse(c *gin.Context, r OpenAICapturedResponse) {
	if c == nil {
		return
	}
	c.Set(openAICapturedResponseKey, r)
}

func setCapturedAssistantText(c *gin.Context, text, responseID, finishReason string) {
	text = strings.TrimSpace(text)
	if text == "" && strings.TrimSpace(responseID) == "" {
		return
	}
	SetOpenAICapturedResponse(c, OpenAICapturedResponse{
		Text:         text,
		ResponseID:   strings.TrimSpace(responseID),
		FinishReason: strings.TrimSpace(finishReason),
	})
}

func captureOpenAIResponseFromJSON(c *gin.Context, body []byte) {
	ext := ExtractOpenAIResponsesResponse(body)
	text := ""
	if len(ext.AssistantEvents) > 0 {
		text = ext.AssistantEvents[0].Content
	}
	setCapturedAssistantText(c, text, ext.ResponseID, ext.FinishReason)
}

func captureOpenAIChatCompletionsResponseFromJSON(c *gin.Context, body []byte) {
	var b strings.Builder
	gjson.GetBytes(body, "choices").ForEach(func(_, choice gjson.Result) bool {
		if t := choice.Get("message.content").String(); t != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(t)
		}
		return true
	})
	finishReason := ""
	gjson.GetBytes(body, "choices").ForEach(func(_, choice gjson.Result) bool {
		if fr := choice.Get("finish_reason").String(); fr != "" {
			finishReason = fr
			return false
		}
		return true
	})
	setCapturedAssistantText(c, b.String(), gjson.GetBytes(body, "id").String(), finishReason)
}

func captureAnthropicResponseFromJSON(c *gin.Context, body []byte) {
	ext := ExtractAnthropicMessagesResponse(body)
	text := ""
	if len(ext.AssistantEvents) > 0 {
		text = ext.AssistantEvents[0].Content
	}
	setCapturedAssistantText(c, text, ext.ResponseID, ext.FinishReason)
}

func captureGeminiResponseFromJSON(c *gin.Context, body []byte) {
	root := gjson.ParseBytes(body)
	if response := root.Get("response"); response.Exists() {
		root = response
	}
	var b strings.Builder
	root.Get("candidates").ForEach(func(_, candidate gjson.Result) bool {
		candidate.Get("content.parts").ForEach(func(_, part gjson.Result) bool {
			if t := part.Get("text").String(); t != "" {
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(t)
			}
			return true
		})
		return true
	})
	finishReason := ""
	root.Get("candidates").ForEach(func(_, candidate gjson.Result) bool {
		if fr := candidate.Get("finishReason").String(); fr != "" {
			finishReason = fr
			return false
		}
		return true
	})
	setCapturedAssistantText(c, b.String(), firstNonEmptyString(root.Get("responseId").String(), root.Get("modelVersion").String()), finishReason)
}

func capturePlainAssistantText(c *gin.Context, text, responseID, finishReason string) {
	setCapturedAssistantText(c, text, responseID, finishReason)
}

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
