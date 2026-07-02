package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// CtxSkipConversationCapture marks requests whose chat history is stored by a
// dedicated subsystem, such as the built-in Chat Playground.
const CtxSkipConversationCapture = "skip_conversation_capture"

type conversationHeaderSignals struct {
	ExplicitConversationID string
	SessionID              string
	ConversationID         string
	ThreadID               string
}

type conversationCaptureForwardResult struct {
	RequestID    string
	ResponseID   string
	Model        string
	FinishReason string
	InputTokens  int64
	OutputTokens int64
	Partial      bool
}

func readConversationHeaderSignals(c *gin.Context) conversationHeaderSignals {
	if c == nil {
		return conversationHeaderSignals{}
	}
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

func (h *OpenAIGatewayHandler) captureResponsesConversation(
	c *gin.Context,
	body []byte,
	result *service.OpenAIForwardResult,
	apiKey *service.APIKey,
	userID int64,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolOpenAI,
		service.ContextDomainOpenAIAPI,
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			ResponseID:   result.ResponseID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractOpenAIResponsesRequest,
	)
}

func (h *OpenAIGatewayHandler) captureChatCompletionsConversation(
	c *gin.Context,
	body []byte,
	result *service.OpenAIForwardResult,
	apiKey *service.APIKey,
	userID int64,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolOpenAI,
		service.ContextDomainOpenAIAPI,
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			ResponseID:   result.ResponseID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractOpenAIChatCompletionsRequest,
	)
}

func (h *GatewayHandler) captureChatCompletionsConversation(
	c *gin.Context,
	body []byte,
	result *service.ForwardResult,
	apiKey *service.APIKey,
	userID int64,
	account *service.Account,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolOpenAI,
		conversationContextDomainForAccount(account, service.ConversationProtocolOpenAI),
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractOpenAIChatCompletionsRequest,
	)
}

func (h *GatewayHandler) captureGatewayResponsesConversation(
	c *gin.Context,
	body []byte,
	result *service.ForwardResult,
	apiKey *service.APIKey,
	userID int64,
	account *service.Account,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolOpenAI,
		conversationContextDomainForAccount(account, service.ConversationProtocolOpenAI),
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractOpenAIResponsesRequest,
	)
}

func (h *GatewayHandler) captureAnthropicMessagesConversation(
	c *gin.Context,
	body []byte,
	result *service.ForwardResult,
	apiKey *service.APIKey,
	userID int64,
	account *service.Account,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolAnthropic,
		conversationContextDomainForAccount(account, service.ConversationProtocolAnthropic),
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractAnthropicMessagesRequest,
	)
}

func (h *GatewayHandler) captureGeminiGenerateContentConversation(
	c *gin.Context,
	body []byte,
	result *service.ForwardResult,
	apiKey *service.APIKey,
	userID int64,
	account *service.Account,
) {
	if result == nil {
		return
	}
	submitConversationCapture(
		c,
		h.cfg,
		h.captureSink,
		h.usageRecordWorkerPool,
		body,
		apiKey,
		userID,
		service.ConversationProtocolGemini,
		conversationContextDomainForAccount(account, service.ConversationProtocolGemini),
		conversationCaptureForwardResult{
			RequestID:    result.RequestID,
			Model:        result.Model,
			InputTokens:  int64(result.Usage.InputTokens),
			OutputTokens: int64(result.Usage.OutputTokens),
			Partial:      result.ClientDisconnect,
		},
		service.ExtractGeminiGenerateContentRequest,
	)
}

func submitConversationCapture(
	c *gin.Context,
	cfg *config.Config,
	captureSink service.CaptureSink,
	usageRecordWorkerPool *service.UsageRecordWorkerPool,
	body []byte,
	apiKey *service.APIKey,
	userID int64,
	protocol string,
	contextDomain string,
	result conversationCaptureForwardResult,
	extract func([]byte) service.ConversationRequestExtract,
) {
	if captureSink == nil || apiKey == nil || extract == nil || !service.ConversationArchiveEnabled(cfg) {
		return
	}
	if c != nil && c.GetBool(CtxSkipConversationCapture) {
		return
	}

	hdr := readConversationHeaderSignals(c)
	captured, _ := service.GetOpenAICapturedResponse(c)

	secret := cfg.ConversationArchive.IdentitySecret
	groupID := derefGroupID(apiKey.GroupID)
	apiKeyID := apiKey.ID
	requestID := result.RequestID
	model := result.Model
	responseID := result.ResponseID
	if responseID == "" {
		responseID = captured.ResponseID
	}
	finishReason := result.FinishReason
	if finishReason == "" {
		finishReason = captured.FinishReason
	}
	assistantText := captured.Text
	inputTokens := result.InputTokens
	outputTokens := result.OutputTokens
	finishedAt := time.Now()

	task := func(ctx context.Context) {
		reqExtract := extract(body)
		signals := reqExtract.Signals
		signals.ExplicitConversationID = hdr.ExplicitConversationID
		signals.SessionID = hdr.SessionID
		signals.ConversationID = hdr.ConversationID
		signals.ThreadID = hdr.ThreadID

		identity := service.ResolveConversationIdentity(service.ConversationIdentityInput{
			Secret:        secret,
			UserID:        userID,
			ContextDomain: contextDomain,
			Protocol:      protocol,
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
		captureSink.Submit(ctx, service.CaptureRecord{
			Version:  service.CaptureRecordVersion,
			Identity: identity,
			Request: service.NormalizedRequest{
				UserID:    userID,
				APIKeyID:  &apiKeyID,
				GroupID:   groupID,
				Protocol:  protocol,
				Model:     reqModel,
				RequestID: requestID,
				Events:    reqExtract.UserEvents,
			},
			Response: service.NormalizedResponse{
				Model:        model,
				ResponseID:   responseID,
				FinishReason: finishReason,
				Partial:      result.Partial,
				InputTokens:  inputTokens,
				OutputTokens: outputTokens,
				Events:       assistantEvents,
			},
			Timing: service.CaptureTiming{FinishedAt: finishedAt},
		})
	}

	submitConversationCaptureTask(c, usageRecordWorkerPool, task)
}

func submitConversationCaptureTask(c *gin.Context, pool *service.UsageRecordWorkerPool, task service.UsageRecordTask) {
	if task == nil {
		return
	}
	var parent context.Context
	if c != nil && c.Request != nil {
		parent = c.Request.Context()
	}
	task = wrapUsageRecordTaskContext(parent, task)
	if pool != nil {
		pool.Submit(task)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	task(ctx)
}

func conversationContextDomainForAccount(account *service.Account, protocol string) string {
	if account == nil {
		return service.ContextDomainUnknown
	}
	switch account.Platform {
	case service.PlatformOpenAI, service.PlatformGrok:
		return service.ContextDomainOpenAIAPI
	case service.PlatformGemini:
		return service.ContextDomainGeminiNative
	case service.PlatformAntigravity:
		if protocol == service.ConversationProtocolGemini {
			return service.ContextDomainAntigravityGemini
		}
		return service.ContextDomainAntigravityClaude
	case service.PlatformAnthropic:
		return service.ContextDomainAnthropicNative
	default:
		return service.ContextDomainUnknown
	}
}
