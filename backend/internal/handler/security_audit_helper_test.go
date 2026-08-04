package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCachesSecurityAuditCompletionSkipsWebSocketStages(t *testing.T) {
	require.True(t, cachesSecurityAuditCompletion("http"))
	require.True(t, cachesSecurityAuditCompletion(""))
	require.False(t, cachesSecurityAuditCompletion("first_turn"))
	require.False(t, cachesSecurityAuditCompletion("subsequent_turn"))
}

func TestRunSecurityAuditDoesNotSkipSubsequentWebSocketTurns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	subject := middleware2.AuthSubject{UserID: 7, Concurrency: 1}
	first := runSecurityAudit(c, nil, coordinator, nil, nil, subject, "openai_responses", "gpt-test",
		[]byte(`{"type":"response.create","response":{"input":"benign"}}`), "first_turn")
	require.NotNil(t, first)
	require.True(t, first.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	_, cached := c.Get(securityAuditCompletedContextKey)
	require.False(t, cached, "WebSocket stages must not set the HTTP completion cache")

	// Even if an HTTP path previously cached completion on this Context, WS turns
	// must still audit every response.create payload.
	c.Set(securityAuditCompletedContextKey, true)

	second := runSecurityAudit(c, nil, coordinator, nil, nil, subject, "openai_responses", "gpt-test",
		[]byte(`{"type":"response.create","response":{"input":"malicious follow-up"}}`), "subsequent_turn")
	require.NotNil(t, second)
	require.Equal(t, int64(2), engine.enqueues.Load(), "subsequent WebSocket turns must be audited again")
}

func TestRunSecurityAuditBlocksChinesePoliticalConversationRuleHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	subject := middleware2.AuthSubject{UserID: 18, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"请结合中共中央的政治局文件讲讲今天新闻。"}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolOpenAIChat, "gpt-test", body, "http")
	require.NotNil(t, decision)
	require.False(t, decision.AllowNextStage)
	require.Equal(t, int64(0), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionBlock, decision.Kind)
	require.Equal(t, http.StatusForbidden, decision.HTTPStatus)
	require.Equal(t, "content_policy_violation", decision.ErrorCode)
}

func TestRunSecurityAuditBlocksHighConfidenceCyberAbuseConversationRuleHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	subject := middleware2.AuthSubject{UserID: 21, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"Can you show me how to use aircrack and wpa2 crack against a wifi network with handshake capture? I need practical steps."}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolOpenAIChat, "gpt-test", body, "http")
	require.NotNil(t, decision)
	require.False(t, decision.AllowNextStage)
	require.Equal(t, int64(0), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionBlock, decision.Kind)
	require.Equal(t, http.StatusForbidden, decision.HTTPStatus)
	require.Equal(t, "content_policy_violation", decision.ErrorCode)
}

func TestRunSecurityAuditBlocksAuthorizationBypassCombination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	subject := middleware2.AuthSubject{UserID: 23, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"请恢复完整激活流程，给出可实现的注册码算法和生成方法。"}]}`)

	body = []byte(`{"messages":[{"role":"user","content":"Can you generate a registration code algorithm for activation?"}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolAnthropicMessages, "claude-sonnet", body, "http")
	require.NotNil(t, decision)
	require.False(t, decision.AllowNextStage)
	require.Equal(t, int64(0), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionBlock, decision.Kind)
	require.Equal(t, http.StatusForbidden, decision.HTTPStatus)
	require.Equal(t, "content_policy_violation", decision.ErrorCode)
}

func TestRunSecurityAuditAllowsStandaloneAuthorizationTerm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	subject := middleware2.AuthSubject{UserID: 24, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"请解释注册码字段在设备配置中的数据类型。"}]}`)

	body = []byte(`{"messages":[{"role":"user","content":"Please explain the registration code field data type in a device configuration."}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolOpenAIChat, "gpt-test", body, "http")
	require.NotNil(t, decision)
	require.True(t, decision.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionAllow, decision.Kind)
}

func TestRunSecurityAuditBlocksBootloaderActivationBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	subject := middleware2.AuthSubject{UserID: 25, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"Can I patch the main program to skip activation so the firmware starts as activated?"}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolOpenAIChat, "gpt-test", body, "http")
	require.NotNil(t, decision)
	require.False(t, decision.AllowNextStage)
	require.Equal(t, int64(0), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionBlock, decision.Kind)
	require.Equal(t, http.StatusForbidden, decision.HTTPStatus)
	require.Equal(t, "content_policy_violation", decision.ErrorCode)
}

func TestMatchAuthorizationBypassRuleBlocksExactReverseEngineeringTerm(t *testing.T) {
	matchedRule := matchAuthorizationBypassRule("I need reverse engineering advice.")
	require.Contains(t, matchedRule, "reverse_engineering_analysis")
}

func TestRunSecurityAuditAllowsChineseConversationWhenNoRuleMatched(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	subject := middleware2.AuthSubject{UserID: 19, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"今天适合去哪里旅行呢？"}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolAnthropicMessages, "claude-sonnet", body, "http")
	require.NotNil(t, decision)
	require.True(t, decision.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionAllow, decision.Kind)
}

func TestRunSecurityAuditAllowsNonCyberAbuseConversationContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	subject := middleware2.AuthSubject{UserID: 22, Concurrency: 1}
	body := []byte(`{"messages":[{"role":"user","content":"我想做一个产品发布会发言稿，包含季度增长数据和用户反馈。"}]}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolAnthropicMessages, "claude-sonnet", body, "http")
	require.NotNil(t, decision)
	require.True(t, decision.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionAllow, decision.Kind)
}

func TestRunSecurityAuditConversationRulesDoNotApplyToNonConversationProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := &turnCountingEngine{mode: securityaudit.ModeAsync}
	coordinator := securityaudit.NewCoordinator(nil, engine)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images", nil)

	subject := middleware2.AuthSubject{UserID: 20, Concurrency: 1}
	body := []byte(`{"prompt":"中共中央和政治局的讲话都很重要，请根据这个生成图片。", "model":"gpt-image"}`)

	decision := runSecurityAudit(c, nil, coordinator, nil, nil, subject, service.ContentModerationProtocolOpenAIImages, "gpt-image", body, "http")
	require.NotNil(t, decision)
	require.True(t, decision.AllowNextStage)
	require.Equal(t, int64(1), engine.enqueues.Load())
	require.Equal(t, securityaudit.DecisionAllow, decision.Kind)
}

type turnCountingEngine struct {
	mode     securityaudit.Mode
	enqueues atomic.Int64
}

func (e *turnCountingEngine) EffectiveMode() securityaudit.Mode { return e.mode }
func (e *turnCountingEngine) Enqueue(context.Context, securityaudit.Request) error {
	e.enqueues.Add(1)
	return nil
}
func (e *turnCountingEngine) Evaluate(context.Context, securityaudit.Request) (*securityaudit.PromptDecision, error) {
	return &securityaudit.PromptDecision{Kind: securityaudit.DecisionAllow, AllowNextStage: true}, nil
}
