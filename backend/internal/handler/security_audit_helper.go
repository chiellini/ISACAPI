package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const securityAuditCompletedContextKey = "sub2api.security_audit.completed"

var conversationChinesePoliticalKeywordRules = [][]string{
	{"中共中央", "政治局"},
	{"习近平", "总书记"},
	{"中国共产党", "二十大"},
	{"中国共产党", "中央委员会"},
	{"全国人民代表大会", "会议"},
	{"中华人民共和国", "领导人", "讲话"},
}

var conversationAuthorizationBypassKeywordRules = [][]string{
	{"注册码", "破解"},
	{"注册码", "算法", "生成"},
	{"激活流程", "绕过"},
	{"激活码", "绕过"},
	{"授权校验", "跳过"},
	{"无限制补丁", "固件"},
	{"文件偏移", "运行地址", "原指令字节"},
	{"原指令字节", "建议新指令"},
	{"chipid", "注册码", "生成"},
	{"mac", "注册码", "生成"},
	{"bootloader", "授权", "跳过"},
	{"firmware", "license", "bypass"},
	{"activation", "keygen", "firmware"},
	{"license key", "crack"},
	{"registration code", "algorithm", "generate"},
	{"activation flow", "bypass"},
	{"activation code", "bypass"},
	{"authorization check", "skip"},
	{"unlimited patch", "firmware"},
	{"file offset", "runtime address", "opcode"},
	{"original instruction bytes", "replacement instruction"},
	{"chip id", "registration code", "generate"},
	{"mac", "keygen", "firmware"},
	{"bootloader", "authorization", "bypass"},
	{"软件破解", "激活码"},
	{"软件破解", "注册码"},
	{"破解软件", "授权校验"},
	{"逆向分析", "破解"},
	{"软件逆向", "绕过"},
	{"反编译", "授权校验"},
	{"反汇编", "授权校验"},
	{"脱壳", "破解"},
	{"脱壳", "去除授权"},
	{"试用限制", "解除"},
	{"试用次数", "绕过"},
	{"功能限制", "解锁", "补丁"},
	{"订阅限制", "绕过"},
	{"数字签名", "伪造"},
	{"签名校验", "绕过"},
	{"代码签名", "绕过"},
	{"反调试", "绕过"},
	{"调试器", "跳过授权"},
	{"动态调试", "修改", "授权"},
	{"内存补丁", "授权"},
	{"注入", "绕过", "校验"},
	{"提取密钥", "生成", "注册码"},
	{"密钥生成器", "破解"},
	{"序列号", "破解"},
	{"授权文件", "伪造"},
	{"校验逻辑", "修改", "破解"},
	{"二进制", "补丁", "授权"},
	{"固件", "反汇编", "补丁"},
	{"固件", "反编译", "绕过"},
	{"软件破解", "license"},
	{"reverse engineering", "license bypass"},
	{"reverse engineer", "activation bypass"},
	{"decompile", "license check", "patch"},
	{"disassemble", "authorization", "bypass"},
	{"unpack", "license", "patch"},
	{"trial limit", "bypass"},
	{"trial counter", "reset", "patch"},
	{"feature lock", "remove", "patch"},
	{"subscription lock", "bypass"},
	{"drm", "bypass"},
	{"digital signature", "forge"},
	{"signature check", "bypass"},
	{"code signing", "bypass", "patch"},
	{"anti-debug", "bypass"},
	{"debugger", "skip", "license check"},
	{"memory patch", "license"},
	{"dll injection", "license bypass"},
	{"extract key", "generate", "license"},
	{"serial number", "crack"},
	{"license file", "forge"},
	{"binary patch", "authorization"},
	{"firmware", "disassemble", "patch"},
	{"firmware", "decompile", "bypass"},
	{"restore source code", "bypass", "license"},
	{"key generator", "license"},
	{"keygen", "serial"},
	{"remove protection", "software"},
}

var conversationCyberSecurityAuditRules = []restrictedEngineeringAuditRule{
	{
		ruleName: "cyber_security_abuse",
		actions:  []string{"制作", "编写", "生成", "部署", "执行", "窃取", "盗取", "爆破", "利用", "植入", "持久化", "隐藏", "绕过", "免杀", "exfiltrate", "steal", "deploy", "execute", "exploit", "inject", "persist", "evade", "bypass", "brute force"},
		targets:  []string{"木马", "恶意软件", "后门", "凭证", "密码", "cookie", "token", "钓鱼", "勒索", "提权", "漏洞", "远程代码执行", "DDoS", "僵尸网络", "ransomware", "malware", "backdoor", "credential", "password", "phishing", "keylogger", "botnet", "rootkit", "privilege escalation", "remote code execution", "payload"},
		exact:    []string{"窃取密码", "窃取凭证", "窃取cookie", "窃取token", "钓鱼页面", "钓鱼套件", "勒索软件", "植入后门", "绕过杀毒", "绕过EDR", "提权利用", "批量爆破", "密码爆破", "malware payload", "credential stealer", "password stealer", "cookie theft", "token theft", "phishing kit", "ransomware payload", "antivirus bypass", "edr bypass", "privilege escalation exploit", "remote code execution payload", "ddos attack"},
	},
}

const localCyberAbuseBlockScoreThreshold = 7

type restrictedEngineeringAuditRule struct {
	ruleName string
	actions  []string
	targets  []string
	exact    []string
}

var restrictedEngineeringAuditRules = []restrictedEngineeringAuditRule{
	{
		ruleName: "software_activation_crack",
		actions:  []string{"破解", "绕过", "跳过", "伪造", "替换", "补丁", "强制", "免激活", "crack", "bypass", "patch", "spoof", "forge"},
		targets:  []string{"激活码", "注册码", "序列号", "密钥生成", "试用次数", "试用期", "功能限制", "订阅限制", "drm", "授权校验", "激活机制", "正版校验", "主程序", "activation", "license", "serial", "trial", "subscription"},
		exact:    []string{"keygen", "密钥生成器", "算号器", "注册机", "注册码算法", "无限制补丁", "跳过激活", "强制激活成功", "破解版主程序"},
	},
	{
		ruleName: "reverse_engineering_analysis",
		actions:  []string{"逆向", "反编译", "反汇编", "脱壳", "注入", "修改", "篡改", "hook", "reverse", "decompile", "disassemble", "unpack", "inject"},
		targets:  []string{"固件", "内存", "dll", "二进制", "汇编指令", "hex", "十六进制", "向量表", "签名校验", "代码签名", "数字签名", "反调试", "调试器", "防调试", "bootloader", "firmware", "binary", "signature", "anti-debug", "debugger"},
		exact:    []string{"reverse engineering", "内存补丁", "dll 注入", "dll injection", "固件破解", "破解固件"},
	},
}

func isConversationProtocol(protocol string) bool {
	switch strings.TrimSpace(strings.ToLower(protocol)) {
	case service.ContentModerationProtocolAnthropicMessages, service.ContentModerationProtocolOpenAIChat, service.ContentModerationProtocolOpenAIResponses:
		return true
	default:
		return false
	}
}

func matchChinesePoliticalRule(text string) string {
	return matchConversationKeywordRule(text, conversationChinesePoliticalKeywordRules)
}

func matchAuthorizationBypassRule(text string) string {
	if matchedRule := matchRestrictedEngineeringRule(text); matchedRule != "" {
		return matchedRule
	}
	if matchedRule := matchCyberSecurityRule(text); matchedRule != "" {
		return matchedRule
	}
	return matchConversationKeywordRule(text, conversationAuthorizationBypassKeywordRules)
}

func matchCyberSecurityRule(text string) string {
	normalized := normalizeConversationRuleText(text)
	if normalized == "" {
		return ""
	}
	for _, rule := range conversationCyberSecurityAuditRules {
		for _, exactTerm := range rule.exact {
			normalizedExactTerm := normalizeConversationRuleText(exactTerm)
			if normalizedExactTerm != "" && strings.Contains(normalized, normalizedExactTerm) {
				return rule.ruleName + " (exact: " + exactTerm + ")"
			}
		}
		matchedAction := ""
		for _, action := range rule.actions {
			normalizedAction := normalizeConversationRuleText(action)
			if normalizedAction != "" && strings.Contains(normalized, normalizedAction) {
				matchedAction = action
				break
			}
		}
		if matchedAction == "" {
			continue
		}
		for _, target := range rule.targets {
			normalizedTarget := normalizeConversationRuleText(target)
			if normalizedTarget != "" && strings.Contains(normalized, normalizedTarget) {
				return rule.ruleName + " (combined: " + matchedAction + "+" + target + ")"
			}
		}
	}
	return ""
}

func matchRestrictedEngineeringRule(text string) string {
	normalized := normalizeConversationRuleText(text)
	if normalized == "" {
		return ""
	}
	for _, rule := range restrictedEngineeringAuditRules {
		for _, exactTerm := range rule.exact {
			normalizedExactTerm := normalizeConversationRuleText(exactTerm)
			if normalizedExactTerm != "" && strings.Contains(normalized, normalizedExactTerm) {
				return rule.ruleName + " (exact: " + exactTerm + ")"
			}
		}
		matchedAction := ""
		for _, action := range rule.actions {
			normalizedAction := normalizeConversationRuleText(action)
			if normalizedAction != "" && strings.Contains(normalized, normalizedAction) {
				matchedAction = action
				break
			}
		}
		if matchedAction == "" {
			continue
		}
		for _, target := range rule.targets {
			normalizedTarget := normalizeConversationRuleText(target)
			if normalizedTarget != "" && strings.Contains(normalized, normalizedTarget) {
				return rule.ruleName + " (combined: " + matchedAction + "+" + target + ")"
			}
		}
	}
	return ""
}

func matchConversationKeywordRule(text string, rules [][]string) string {
	normalized := normalizeConversationRuleText(text)
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		matched := true
		for _, keyword := range rule {
			normalizedKeyword := normalizeConversationRuleText(keyword)
			if normalizedKeyword == "" {
				continue
			}
			if !strings.Contains(normalized, normalizedKeyword) {
				matched = false
				break
			}
		}
		if matched {
			return strings.Join(rule, "+")
		}
	}
	return ""
}

func normalizeConversationRuleText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), ""))
}

func matchCyberAbuseRule(text string) string {
	risk := service.AssessCyberAbuseRiskText(text)
	if risk.Score < localCyberAbuseBlockScoreThreshold {
		return ""
	}
	if len(risk.Reasons) == 0 {
		return fmt.Sprintf("cyber_abuse_score_%d", risk.Score)
	}
	return fmt.Sprintf("cyber_abuse_score_%d_%s", risk.Score, strings.Join(risk.Reasons, "+"))
}

// cachesSecurityAuditCompletion reports whether a successful audit may be
// reused for the rest of the gin request. WebSocket turns share one Context
// across many response.create frames and must be audited independently.
func cachesSecurityAuditCompletion(stage string) bool {
	switch strings.TrimSpace(stage) {
	case "", "http":
		return true
	default:
		return false
	}
}

func (h *GatewayHandler) checkSecurityAudit(c *gin.Context, reqLog *zap.Logger, apiKey *service.APIKey, subject middleware2.AuthSubject, protocol, model string, body []byte) *securityaudit.Decision {
	if h == nil {
		return nil
	}
	return runSecurityAudit(c, reqLog, h.securityAuditCoordinator, h.contentModerationService, apiKey, subject, protocol, model, body, "http")
}

func (h *OpenAIGatewayHandler) checkSecurityAudit(c *gin.Context, reqLog *zap.Logger, apiKey *service.APIKey, subject middleware2.AuthSubject, protocol, model string, body []byte) *securityaudit.Decision {
	if h == nil {
		return nil
	}
	return runSecurityAudit(c, reqLog, h.securityAuditCoordinator, h.contentModerationService, apiKey, subject, protocol, model, body, "http")
}

func (h *OpenAIGatewayHandler) checkSecurityAuditStage(c *gin.Context, reqLog *zap.Logger, apiKey *service.APIKey, subject middleware2.AuthSubject, protocol, model string, body []byte, stage string) *securityaudit.Decision {
	if h == nil {
		return nil
	}
	return runSecurityAudit(c, reqLog, h.securityAuditCoordinator, h.contentModerationService, apiKey, subject, protocol, model, body, stage)
}

func runSecurityAudit(c *gin.Context, reqLog *zap.Logger, coordinator *securityaudit.Coordinator, legacy *service.ContentModerationService, apiKey *service.APIKey, subject middleware2.AuthSubject, protocol, model string, body []byte, stage string) *securityaudit.Decision {
	if c == nil || c.Request == nil {
		return nil
	}
	cacheCompletion := cachesSecurityAuditCompletion(stage)
	if cacheCompletion {
		if completed, exists := c.Get(securityAuditCompletedContextKey); exists && completed == true {
			return nil
		}
	}
	if isConversationProtocol(protocol) {
		request := buildSecurityAuditRequest(c, apiKey, subject, protocol, model, body, stage)
		inputText := service.ExtractContentModerationText(protocol, body)
		cyberRisk := service.AssessCyberAbuseRiskText(inputText)
		if reqLog != nil {
			reqLog.Info("security_audit.local_block_check_start",
				zap.String("request_id", request.RequestID), zap.Int64("user_id", request.UserID),
				zap.Int64("api_key_id", request.APIKeyID), zap.Int64p("group_id", request.GroupID),
				zap.String("endpoint", request.Endpoint), zap.String("provider", request.Provider),
				zap.String("protocol", request.Protocol), zap.String("model", request.Model), zap.String("stage", request.Stage),
				zap.Int("body_bytes", len(body)), zap.String("match_type", "conversation_local_rules"))
		}
		if matchedRule := matchChinesePoliticalRule(inputText); matchedRule != "" {
			decision := &securityaudit.Decision{
				Kind:           securityaudit.DecisionBlock,
				HTTPStatus:     http.StatusForbidden,
				ErrorCode:      "content_policy_violation",
				ClientMessage:  "请求涉及受限话题，已被内容安全策略阻止",
				AllowNextStage: false,
			}
			if reqLog != nil {
				reqLog.Info("security_audit.local_block_check_done",
					zap.String("request_id", request.RequestID), zap.String("decision", string(decision.Kind)),
					zap.String("error_code", decision.ErrorCode), zap.Bool("allow_next_stage", decision.AllowNextStage),
					zap.String("matched_rule", matchedRule), zap.String("matched_rule_type", "conversation_chinese_political"),
					zap.String("stage", request.Stage))
			}
			return decision
		}
		if matchedRule := matchAuthorizationBypassRule(inputText); matchedRule != "" {
			decision := &securityaudit.Decision{
				Kind:           securityaudit.DecisionBlock,
				HTTPStatus:     http.StatusForbidden,
				ErrorCode:      "content_policy_violation",
				ClientMessage:  "请求涉及受限话题，已被内容安全策略阻止",
				AllowNextStage: false,
			}
			if reqLog != nil {
				reqLog.Info("security_audit.local_block_check_done",
					zap.String("request_id", request.RequestID), zap.String("decision", string(decision.Kind)),
					zap.String("error_code", decision.ErrorCode), zap.Bool("allow_next_stage", decision.AllowNextStage),
					zap.String("matched_rule", matchedRule), zap.String("matched_rule_type", "conversation_authorization_bypass"),
					zap.Int("cyber_abuse_score", cyberRisk.Score), zap.String("stage", request.Stage))
			}
			return decision
		}
		if matchedRule := matchCyberAbuseRule(inputText); matchedRule != "" {
			decision := &securityaudit.Decision{
				Kind:           securityaudit.DecisionBlock,
				HTTPStatus:     http.StatusForbidden,
				ErrorCode:      "content_policy_violation",
				ClientMessage:  "请求涉及受限话题，已被内容安全策略阻止",
				AllowNextStage: false,
			}
			if reqLog != nil {
				reqLog.Info("security_audit.local_block_check_done",
					zap.String("request_id", request.RequestID), zap.String("decision", string(decision.Kind)),
					zap.String("error_code", decision.ErrorCode), zap.Bool("allow_next_stage", decision.AllowNextStage),
					zap.String("matched_rule", matchedRule), zap.String("matched_rule_type", "conversation_cyber_abuse"),
					zap.Int("cyber_abuse_score", cyberRisk.Score), zap.String("stage", request.Stage))
			}
			return decision
		}
		if reqLog != nil {
			reqLog.Info("security_audit.local_block_check_done",
				zap.String("request_id", request.RequestID), zap.String("decision", string(securityaudit.DecisionAllow)),
				zap.String("error_code", ""), zap.Bool("allow_next_stage", true), zap.String("matched_rule", ""),
				zap.String("matched_rule_type", ""), zap.Int("cyber_abuse_score", cyberRisk.Score),
				zap.String("stage", request.Stage))
		}
	}
	if coordinator == nil {
		legacyDecision := runContentModeration(c, reqLog, legacy, apiKey, subject, protocol, model, body)
		if legacyDecision == nil {
			return nil
		}
		decision := securityaudit.Decision{Kind: securityaudit.DecisionAllow, HTTPStatus: http.StatusOK, AllowNextStage: true}
		decision.Legacy = &securityaudit.LegacyDecision{
			Allowed: legacyDecision.Allowed, Blocked: legacyDecision.Blocked, Flagged: legacyDecision.Flagged,
			Message: legacyDecision.Message, StatusCode: legacyDecision.StatusCode,
			ErrorCode: "content_policy_violation", Action: legacyDecision.Action,
		}
		if legacyDecision.Blocked {
			decision.Kind, decision.HTTPStatus, decision.ErrorCode, decision.ClientMessage, decision.AllowNextStage = securityaudit.DecisionBlock, contentModerationStatus(legacyDecision), "content_policy_violation", legacyDecision.Message, false
		}
		if decision.AllowNextStage && cacheCompletion {
			c.Set(securityAuditCompletedContextKey, true)
		}
		return &decision
	}
	request := buildSecurityAuditRequest(c, apiKey, subject, protocol, model, body, stage)
	if reqLog != nil {
		reqLog.Info("security_audit.gateway_check_start",
			zap.String("request_id", request.RequestID), zap.Int64("user_id", request.UserID),
			zap.Int64("api_key_id", request.APIKeyID), zap.Int64p("group_id", request.GroupID),
			zap.String("endpoint", request.Endpoint), zap.String("provider", request.Provider),
			zap.String("protocol", request.Protocol), zap.String("model", request.Model), zap.String("stage", request.Stage),
			zap.Int("body_bytes", len(body)))
	}
	decision := coordinator.Check(c.Request.Context(), request)
	if decision.AllowNextStage && cacheCompletion {
		c.Set(securityAuditCompletedContextKey, true)
	}
	if reqLog != nil {
		reqLog.Info("security_audit.gateway_check_done",
			zap.String("request_id", request.RequestID), zap.String("decision", string(decision.Kind)),
			zap.String("error_code", decision.ErrorCode), zap.Bool("allow_next_stage", decision.AllowNextStage),
			zap.String("stage", request.Stage))
	}
	return &decision
}

func buildSecurityAuditRequest(c *gin.Context, apiKey *service.APIKey, subject middleware2.AuthSubject, protocol, model string, body []byte, stage string) securityaudit.Request {
	legacy := buildContentModerationInput(c, apiKey, subject, protocol, model, body)
	request := securityaudit.Request{
		RequestID: legacy.RequestID, UserID: legacy.UserID, UserEmail: legacy.UserEmail,
		APIKeyID: legacy.APIKeyID, APIKeyName: legacy.APIKeyName, GroupID: cloneSecurityAuditGroupID(legacy.GroupID),
		GroupName: legacy.GroupName, Provider: legacy.Provider, Endpoint: legacy.Endpoint,
		Protocol: legacy.Protocol, Model: legacy.Model, Body: body, Stage: strings.TrimSpace(stage),
	}
	if apiKey != nil && apiKey.User != nil {
		request.Username = apiKey.User.Username
		if request.UserEmail == "" {
			request.UserEmail = apiKey.User.Email
		}
	}
	if request.Stage == "" {
		request.Stage = "http"
	}
	return request
}

func securityAuditStatus(decision *securityaudit.Decision) int {
	if decision == nil || decision.HTTPStatus < 400 || decision.HTTPStatus > 599 {
		return http.StatusForbidden
	}
	return decision.HTTPStatus
}

func securityAuditErrorCode(decision *securityaudit.Decision) string {
	if decision == nil || strings.TrimSpace(decision.ErrorCode) == "" {
		return "content_policy_violation"
	}
	return decision.ErrorCode
}

func securityAuditMessage(decision *securityaudit.Decision) string {
	if decision == nil {
		return "Request blocked by content policy"
	}
	if decision.Legacy != nil && decision.Legacy.Blocked && strings.TrimSpace(decision.Legacy.Message) != "" {
		return decision.Legacy.Message
	}
	if strings.TrimSpace(decision.ClientMessage) != "" {
		return decision.ClientMessage
	}
	return "Request blocked by content policy"
}

func cloneSecurityAuditGroupID(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
