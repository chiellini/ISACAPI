package service

import (
	"encoding/json"
	"strings"
)

// ConversationRequestExtract 是从请求体提取的可归档部分（协议相关）。
type ConversationRequestExtract struct {
	Model      string
	Signals    ConversationIdentitySignals
	UserEvents []NormalizedEvent
}

// ConversationResponseExtract 是从非流式响应体提取的可归档部分（协议相关）。
type ConversationResponseExtract struct {
	Model           string
	ResponseID      string
	FinishReason    string
	InputTokens     int64
	OutputTokens    int64
	AssistantEvents []NormalizedEvent
}

// V1 关键取舍：每次请求只取“最新一条 user 消息”，不存客户端重发的完整历史，
// 从根本上避免历史重复与近似平方增长。完整分支 reconcile（编辑旧历史分叉）留待后续。

// --- OpenAI Responses ---

type openaiResponsesRequest struct {
	Model              string          `json:"model"`
	PreviousResponseID string          `json:"previous_response_id"`
	PromptCacheKey     string          `json:"prompt_cache_key"`
	Instructions       string          `json:"instructions"`
	Input              json.RawMessage `json:"input"`
}

type openaiInputItem struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ExtractOpenAIResponsesRequest 从 OpenAI Responses 请求体提取归档信息。容错：解析失败返回空结构。
func ExtractOpenAIResponsesRequest(body []byte) ConversationRequestExtract {
	var out ConversationRequestExtract
	var req openaiResponsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return out
	}
	out.Model = req.Model
	out.Signals.PreviousResponseID = req.PreviousResponseID
	out.Signals.PromptCacheKey = req.PromptCacheKey

	userText := lastOpenAIUserText(req.Input)
	if userText != "" {
		out.UserEvents = append(out.UserEvents, NormalizedEvent{
			Role: ConversationRoleUser, Kind: ConversationKindMessage, Content: userText,
		})
	}
	// 内容前缀种子（兜底身份信号）：system/instructions + 最新用户文本。
	out.Signals.ContentSeed = strings.TrimSpace(req.Instructions + "\n" + userText)
	return out
}

func lastOpenAIUserText(input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}
	// input 可能是纯字符串。
	var s string
	if err := json.Unmarshal(input, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var items []openaiInputItem
	if err := json.Unmarshal(input, &items); err != nil {
		return ""
	}
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].Role != "" && items[i].Role != ConversationRoleUser {
			continue
		}
		if text := extractOpenAIContentText(items[i].Content); text != "" {
			return text
		}
	}
	return ""
}

func extractOpenAIContentText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &parts); err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Text != "" && (strings.Contains(p.Type, "text") || p.Type == "") {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(p.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

type openaiResponsesResponse struct {
	ID     string `json:"id"`
	Model  string `json:"model"`
	Status string `json:"status"`
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int64 `json:"input_tokens"`
		OutputTokens int64 `json:"output_tokens"`
	} `json:"usage"`
}

// ExtractOpenAIResponsesResponse 从 OpenAI Responses 非流式响应体提取归档信息。
func ExtractOpenAIResponsesResponse(body []byte) ConversationResponseExtract {
	var out ConversationResponseExtract
	var resp openaiResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return out
	}
	out.Model = resp.Model
	out.ResponseID = resp.ID
	out.FinishReason = resp.Status
	out.InputTokens = resp.Usage.InputTokens
	out.OutputTokens = resp.Usage.OutputTokens

	var b strings.Builder
	for _, item := range resp.Output {
		if item.Type != "" && item.Type != "message" {
			continue
		}
		for _, c := range item.Content {
			if c.Text != "" && strings.Contains(c.Type, "text") {
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(c.Text)
			}
		}
	}
	if text := strings.TrimSpace(b.String()); text != "" {
		out.AssistantEvents = append(out.AssistantEvents, NormalizedEvent{
			Role: ConversationRoleAssistant, Kind: ConversationKindMessage, Content: text,
		})
	}
	return out
}

// --- Anthropic Messages ---

type anthropicMessagesRequest struct {
	Model    string          `json:"model"`
	System   json.RawMessage `json:"system"`
	Messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
	Metadata struct {
		UserID string `json:"user_id"`
	} `json:"metadata"`
}

// ExtractAnthropicMessagesRequest 从 Anthropic Messages 请求体提取归档信息。
func ExtractAnthropicMessagesRequest(body []byte) ConversationRequestExtract {
	var out ConversationRequestExtract
	var req anthropicMessagesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return out
	}
	out.Model = req.Model
	out.Signals.MetadataNativeID = req.Metadata.UserID

	systemText := extractAnthropicText(req.System)
	if systemText != "" {
		out.UserEvents = append(out.UserEvents, NormalizedEvent{
			Role: ConversationRoleSystem, Kind: ConversationKindMessage, Content: systemText,
		})
	}
	userText := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role != ConversationRoleUser {
			continue
		}
		if t := extractAnthropicText(req.Messages[i].Content); t != "" {
			userText = t
			break
		}
	}
	if userText != "" {
		out.UserEvents = append(out.UserEvents, NormalizedEvent{
			Role: ConversationRoleUser, Kind: ConversationKindMessage, Content: userText,
		})
	}
	out.Signals.ContentSeed = strings.TrimSpace(systemText + "\n" + userText)
	return out
}

func extractAnthropicText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var b strings.Builder
	for _, bl := range blocks {
		if bl.Text != "" && (bl.Type == "text" || bl.Type == "") {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(bl.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

type anthropicMessagesResponse struct {
	ID         string `json:"id"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int64 `json:"input_tokens"`
		OutputTokens int64 `json:"output_tokens"`
	} `json:"usage"`
}

// ExtractAnthropicMessagesResponse 从 Anthropic Messages 非流式响应体提取归档信息。
func ExtractAnthropicMessagesResponse(body []byte) ConversationResponseExtract {
	var out ConversationResponseExtract
	var resp anthropicMessagesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return out
	}
	out.Model = resp.Model
	out.ResponseID = resp.ID
	out.FinishReason = resp.StopReason
	out.InputTokens = resp.Usage.InputTokens
	out.OutputTokens = resp.Usage.OutputTokens

	var b strings.Builder
	for _, c := range resp.Content {
		if c.Text != "" && c.Type == "text" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(c.Text)
		}
	}
	if text := strings.TrimSpace(b.String()); text != "" {
		out.AssistantEvents = append(out.AssistantEvents, NormalizedEvent{
			Role: ConversationRoleAssistant, Kind: ConversationKindMessage, Content: text,
		})
	}
	return out
}
