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
	Thinking     string // 模型思维链（Anthropic thinking / OpenAI reasoning / Gemini thought 摘要）
	ResponseID   string
	FinishReason string
}

func (s *OpenAIGatewayService) conversationCaptureEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.ConversationArchive.Enabled
}

type openAIResponseAccumulator struct {
	text         strings.Builder
	thinking     strings.Builder
	responseID   string
	finishReason string
}

func newOpenAIResponseAccumulator() *openAIResponseAccumulator {
	return &openAIResponseAccumulator{}
}

func (a *openAIResponseAccumulator) observeSSE(data []byte) {
	a.observeSSEWithType(data, "")
}

func (a *openAIResponseAccumulator) observeSSEWithType(data []byte, eventType string) {
	if a == nil || len(data) == 0 {
		return
	}
	switch effectiveOpenAISSEEventType(data, eventType) {
	case "response.output_text.delta":
		if d := gjson.GetBytes(data, "delta").String(); d != "" {
			_, _ = a.text.WriteString(d)
		}
	case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		if d := gjson.GetBytes(data, "delta").String(); d != "" {
			_, _ = a.thinking.WriteString(d)
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
				_, _ = a.text.WriteString(text)
			}
		}
		if a.thinking.Len() == 0 {
			if thinking := openAIThinkingFromResponseOutput(data); thinking != "" {
				_, _ = a.thinking.WriteString(thinking)
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
			_, _ = a.text.WriteString(d)
		}
		if t := choice.Get("message.content").String(); t != "" && a.text.Len() == 0 {
			_, _ = a.text.WriteString(t)
		}
		// 思维链：DeepSeek 风格 reasoning_content，OpenRouter 风格 reasoning。
		if r := firstNonEmptyGJSON(choice.Get("delta.reasoning_content"), choice.Get("delta.reasoning")); r != "" {
			_, _ = a.thinking.WriteString(r)
		} else if r := firstNonEmptyGJSON(choice.Get("message.reasoning_content"), choice.Get("message.reasoning")); r != "" && a.thinking.Len() == 0 {
			_, _ = a.thinking.WriteString(r)
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
				_, _ = a.text.WriteString(text)
			}
		}
	case "content_block_start":
		if text := gjson.GetBytes(data, "content_block.text").String(); text != "" {
			_, _ = a.text.WriteString(text)
		}
		if th := gjson.GetBytes(data, "content_block.thinking").String(); th != "" {
			_, _ = a.thinking.WriteString(th)
		}
	case "content_block_delta":
		if text := gjson.GetBytes(data, "delta.text").String(); text != "" {
			_, _ = a.text.WriteString(text)
		}
		// thinking_delta 增量的字段名是 thinking；redacted_thinking 无明文可采，跳过。
		if th := gjson.GetBytes(data, "delta.thinking").String(); th != "" {
			_, _ = a.thinking.WriteString(th)
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
		Thinking:     strings.TrimSpace(a.thinking.String()),
		ResponseID:   a.responseID,
		FinishReason: a.finishReason,
	}
}

func openAITextFromResponseOutput(data []byte) string {
	var b strings.Builder
	gjson.GetBytes(data, "response.output").ForEach(func(_, item gjson.Result) bool {
		// reasoning 项的 summary/content 文本属于思维链，不混入正文。
		if item.Get("type").String() == "reasoning" {
			return true
		}
		item.Get("content").ForEach(func(_, c gjson.Result) bool {
			if strings.Contains(c.Get("type").String(), "text") {
				if t := c.Get("text").String(); t != "" {
					if b.Len() > 0 {
						_, _ = b.WriteString("\n")
					}
					_, _ = b.WriteString(t)
				}
			}
			return true
		})
		return true
	})
	return strings.TrimSpace(b.String())
}

// openAIThinkingFromResponseOutput 从终止事件的 response.output 中提取 reasoning
// 项的思维链文本（summary 摘要 + content 明文，二者取其一存在时）。
func openAIThinkingFromResponseOutput(data []byte) string {
	var b strings.Builder
	gjson.GetBytes(data, "response.output").ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() != "reasoning" {
			return true
		}
		for _, key := range []string{"summary", "content"} {
			item.Get(key).ForEach(func(_, c gjson.Result) bool {
				if strings.Contains(c.Get("type").String(), "text") {
					if t := c.Get("text").String(); t != "" {
						if b.Len() > 0 {
							_, _ = b.WriteString("\n")
						}
						_, _ = b.WriteString(t)
					}
				}
				return true
			})
		}
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
				_, _ = b.WriteString("\n")
			}
			_, _ = b.WriteString(t)
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

func firstNonEmptyGJSON(values ...gjson.Result) string {
	for _, v := range values {
		if s := v.String(); s != "" {
			return s
		}
	}
	return ""
}

// assistantTextAndThinking 从抽取器的助手事件中拆出正文与思维链文本。
func assistantTextAndThinking(events []NormalizedEvent) (text, thinking string) {
	for _, ev := range events {
		if ev.Content == "" {
			continue
		}
		if ev.Kind == ConversationKindThinking {
			if thinking != "" {
				thinking += "\n"
			}
			thinking += ev.Content
			continue
		}
		if text == "" {
			text = ev.Content
		}
	}
	return text, thinking
}

func setCapturedAssistantText(c *gin.Context, text, thinking, responseID, finishReason string) {
	text = strings.TrimSpace(text)
	thinking = strings.TrimSpace(thinking)
	if text == "" && thinking == "" && strings.TrimSpace(responseID) == "" {
		return
	}
	SetOpenAICapturedResponse(c, OpenAICapturedResponse{
		Text:         text,
		Thinking:     thinking,
		ResponseID:   strings.TrimSpace(responseID),
		FinishReason: strings.TrimSpace(finishReason),
	})
}

func captureOpenAIResponseFromJSON(c *gin.Context, body []byte) {
	ext := ExtractOpenAIResponsesResponse(body)
	text, thinking := assistantTextAndThinking(ext.AssistantEvents)
	// Replace the previous attempt even when the successful response has no text.
	SetOpenAICapturedResponse(c, OpenAICapturedResponse{
		Text: text, Thinking: thinking, ResponseID: ext.ResponseID, FinishReason: ext.FinishReason,
	})
}

func captureOpenAIResponseFromSSE(c *gin.Context, body []byte) {
	acc := newOpenAIResponseAccumulator()
	forEachOpenAISSEFrame(string(body), func(eventType string, data []byte) {
		acc.observeSSEWithType(data, eventType)
	})
	SetOpenAICapturedResponse(c, acc.result())
}

func captureOpenAIChatCompletionsResponseFromJSON(c *gin.Context, body []byte) {
	var b, tb strings.Builder
	gjson.GetBytes(body, "choices").ForEach(func(_, choice gjson.Result) bool {
		if t := choice.Get("message.content").String(); t != "" {
			if b.Len() > 0 {
				_, _ = b.WriteString("\n")
			}
			_, _ = b.WriteString(t)
		}
		if r := firstNonEmptyGJSON(choice.Get("message.reasoning_content"), choice.Get("message.reasoning")); r != "" {
			if tb.Len() > 0 {
				_, _ = tb.WriteString("\n")
			}
			_, _ = tb.WriteString(r)
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
	setCapturedAssistantText(c, b.String(), tb.String(), gjson.GetBytes(body, "id").String(), finishReason)
}

func captureAnthropicResponseFromJSON(c *gin.Context, body []byte) {
	ext := ExtractAnthropicMessagesResponse(body)
	text, thinking := assistantTextAndThinking(ext.AssistantEvents)
	setCapturedAssistantText(c, text, thinking, ext.ResponseID, ext.FinishReason)
}

func captureGeminiResponseFromJSON(c *gin.Context, body []byte) {
	root := gjson.ParseBytes(body)
	if response := root.Get("response"); response.Exists() {
		root = response
	}
	var b, tb strings.Builder
	root.Get("candidates").ForEach(func(_, candidate gjson.Result) bool {
		candidate.Get("content.parts").ForEach(func(_, part gjson.Result) bool {
			t := part.Get("text").String()
			if t == "" {
				return true
			}
			// thought:true 的 part 是思维链摘要，归档时与正文分开。
			if part.Get("thought").Bool() {
				if tb.Len() > 0 {
					_, _ = tb.WriteString("\n")
				}
				_, _ = tb.WriteString(t)
				return true
			}
			if b.Len() > 0 {
				_, _ = b.WriteString("\n")
			}
			_, _ = b.WriteString(t)
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
	setCapturedAssistantText(c, b.String(), tb.String(), firstNonEmptyString(root.Get("responseId").String(), root.Get("modelVersion").String()), finishReason)
}

func capturePlainAssistantText(c *gin.Context, text, thinking, responseID, finishReason string) {
	setCapturedAssistantText(c, text, thinking, responseID, finishReason)
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
