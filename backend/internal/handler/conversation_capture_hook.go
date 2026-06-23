package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// CtxSkipConversationCapture 是 gin.Context 标记位：置 true 时跳过网关侧会话存档采集。
// 内置聊天 Playground 用它来避免把对话重复写进 API 会话存档系统
// （聊天历史已单独存于 chat_sessions / chat_messages 表）。
const CtxSkipConversationCapture = "skip_conversation_capture"

// conversationHeaderSignals 汇集从 HTTP 头读取的会话标识候选（同时兼容下划线/连字符写法）。
type conversationHeaderSignals struct {
	ExplicitConversationID string
	SessionID              string
	ConversationID         string
	ThreadID               string
}

// readConversationHeaderSignals 在请求 goroutine 上读取头部信号（gin.Context 非 goroutine 安全）。
//
// 注意：Nginx 默认可能丢弃带下划线的头（如 session_id），代理层需开启 underscores_in_headers on。
func readConversationHeaderSignals(c *gin.Context) conversationHeaderSignals {
	first := func(keys ...string) string {
		for _, k := range keys {
			if v := strings.TrimSpace(c.GetHeader(k)); v != "" {
				return v
			}
		}
		return ""
	}
	return conversationHeaderSignals{
		ExplicitConversationID: first("X-Sub2API-Conversation-ID"),
		SessionID:              first("session-id", "session_id"),
		ConversationID:         first("conversation-id", "conversation_id"),
		ThreadID:               first("thread-id", "thread_id"),
	}
}

// captureResponsesConversation 旁路采集 OpenAI Responses 一轮对话（请求侧 + 响应元数据）。
//
// 安全约束：仅在启用时运行；不触碰转发写路径；落库经 worker 池 + 归档器双重 fail-open，
// 任何环节失败都不影响 API 响应。第一版只采集请求侧（用户文本 + 身份）与响应元数据
// （response_id / model / token）；assistant 文本需 SSE tee，作为后续步骤接入。
func (h *OpenAIGatewayHandler) captureResponsesConversation(
	c *gin.Context,
	body []byte,
	result *service.OpenAIForwardResult,
	apiKey *service.APIKey,
	userID int64,
) {
	if h.captureSink == nil || result == nil || apiKey == nil || !service.ConversationArchiveEnabled(h.cfg) {
		return
	}
	// 内置聊天 Playground：对话已单独存表，跳过网关侧会话存档，避免重复采集。
	if c != nil && c.GetBool(CtxSkipConversationCapture) {
		return
	}

	hdr := readConversationHeaderSignals(c)
	// assistant 文本由转发路径旁路聚合后暂存于 context，在请求 goroutine 上读取。
	captured, _ := service.GetOpenAICapturedResponse(c)

	secret := h.cfg.ConversationArchive.IdentitySecret
	groupID := derefGroupID(apiKey.GroupID)
	apiKeyID := apiKey.ID
	requestID := result.RequestID
	model := result.Model
	responseID := result.ResponseID
	if responseID == "" {
		responseID = captured.ResponseID
	}
	finishReason := captured.FinishReason
	assistantText := captured.Text
	inputTokens := int64(result.Usage.InputTokens)
	outputTokens := int64(result.Usage.OutputTokens)
	// 上游隔离域：第一版 OpenAI Responses 统一归 openai_api（codex OAuth 细分留待后续）。
	contextDomain := service.ContextDomainOpenAIAPI

	task := func(ctx context.Context) {
		reqExtract := service.ExtractOpenAIResponsesRequest(body)
		signals := reqExtract.Signals
		signals.ExplicitConversationID = hdr.ExplicitConversationID
		signals.SessionID = hdr.SessionID
		signals.ConversationID = hdr.ConversationID
		signals.ThreadID = hdr.ThreadID

		identity := service.ResolveConversationIdentity(service.ConversationIdentityInput{
			Secret:        secret,
			UserID:        userID,
			ContextDomain: contextDomain,
			Protocol:      service.ConversationProtocolOpenAI,
			Signals:       signals,
		})

		reqModel := reqExtract.Model
		if reqModel == "" {
			reqModel = model
		}
		var assistantEvents []service.NormalizedEvent
		if assistantText != "" {
			assistantEvents = []service.NormalizedEvent{{
				Role:    service.ConversationRoleAssistant,
				Kind:    service.ConversationKindMessage,
				Content: assistantText,
			}}
		}
		h.captureSink.Submit(ctx, service.CaptureRecord{
			Version:  service.CaptureRecordVersion,
			Identity: identity,
			Request: service.NormalizedRequest{
				UserID:    userID,
				APIKeyID:  &apiKeyID,
				GroupID:   groupID,
				Protocol:  service.ConversationProtocolOpenAI,
				Model:     reqModel,
				RequestID: requestID,
				Events:    reqExtract.UserEvents,
			},
			Response: service.NormalizedResponse{
				Model:        model,
				ResponseID:   responseID,
				FinishReason: finishReason,
				InputTokens:  inputTokens,
				OutputTokens: outputTokens,
				Events:       assistantEvents,
			},
			Timing: service.CaptureTiming{FinishedAt: time.Now()},
		})
	}

	// 经 worker 池提交，避免在请求热路径解析请求体；池满即丢弃（fail-open）。
	h.usageRecordWorkerPool.Submit(task)
}
