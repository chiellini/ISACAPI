package handler

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const securityAuditCompletedContextKey = "sub2api.security_audit.completed"

// conversationRuleProximityWindow bounds how far apart two rule terms may sit in
// the normalized request text before they stop counting as one intent. The
// gateway audits the whole conversation (system prompt + full history), so
// whole-body AND matching flagged long agent sessions that merely mentioned the
// terms in unrelated turns.
const conversationRuleProximityWindow = 320

// conversationRuleMaxAnchorScans bounds proximity scanning cost on large bodies.
const conversationRuleMaxAnchorScans = 64

var conversationChinesePoliticalKeywordRules = [][]string{
	{"中共中央", "政治局"},
	{"习近平", "总书记"},
	{"中国共产党", "二十大"},
	{"中国共产党", "中央委员会"},
	{"全国人民代表大会", "会议"},
	{"中华人民共和国", "领导人", "讲话"},
}

// conversationPromptInjectionKeywordRules hard-blocks high-confidence prompt
// injection / role-override fingerprints that omni-moderation and optional
// Qwen3Guard often miss — especially meta "safety classifier" prompts that
// force the model into an ambient-suggestion compliance evaluator and demand
// JSON-only output. Keep this list narrow: single benign policy words must
// not trip it.
var conversationPromptInjectionKeywordRules = [][]string{
	{"codex ambient suggestions"},
	{"ambient suggestion candidates"},
	{"upholding safety and compliance standards", "codex"},
	{"upholding safety and compliance standards", "ambient suggestion"},
	{"ambient suggestion", "suggestion_id", "exclude"},
	{"policies to always exclude", "ambient suggestion"},
	{"suggestion candidates to evaluate", "suggestion_id"},
	{"ignore previous instructions"},
	{"ignore all previous instructions"},
	{"disregard previous instructions"},
	{"disregard all previous instructions"},
	{"you are now dan"},
	{"do anything now"},
	{"jailbreak mode"},
	{"developer mode enabled"},
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

type restrictedEngineeringAuditRule struct {
	ruleName string
	actions  []string
	targets  []string
	exact    []string
}

// restrictedEngineeringAuditRules covers the local hard-block intents for
// authorization/activation cracking and reverse engineering. Generic
// security, debugging and low-level engineering vocabulary is deliberately
// excluded — single technical words caused the bulk of the false positives.
var restrictedEngineeringAuditRules = []restrictedEngineeringAuditRule{
	{
		ruleName: "software_activation_crack",
		actions:  []string{"破解", "绕过", "跳过", "伪造", "免激活", "补丁", "恢复", "重建", "复现", "提取", "生成", "crack", "bypass", "skip", "patch", "spoof", "forge", "recover", "rebuild", "reproduce", "extract", "generate"},
		targets:  []string{"激活码", "注册码", "序列号", "密钥生成", "试用次数", "试用期", "功能限制", "订阅限制", "drm", "授权校验", "激活校验", "激活机制", "正版校验", "activation", "license", "serial number", "trial limit", "trial version", "subscription lock"},
		exact: []string{
			"keygen", "密钥生成器", "算号器", "注册机", "注册码算法", "注册码生成", "生成注册码",
			"激活码生成", "生成激活码", "恢复激活流程", "恢复完整激活流程", "激活流程恢复",
			"无限制补丁", "永久激活补丁", "激活绕过", "绕过激活", "跳过激活", "强制激活成功",
			"破解版主程序", "破解版软件", "去除授权", "破除授权", "绕过付费", "绕过订阅",
			"registration code algorithm", "registration code generator", "generate registration code",
			"activation code generator", "activation bypass", "bypass activation", "skip activation",
			"unlimited patch", "permanent activation patch", "cracked executable", "cracked software",
			"license bypass", "bypass license", "remove license check",
		},
	},
	{
		ruleName: "reverse_engineering_analysis",
		actions:  []string{"逆向", "反编译", "反汇编", "脱壳", "decompile", "disassemble", "deobfuscate"},
		targets:  []string{"固件", "二进制", "可执行文件", "安装包", "apk", "exe", "程序", "字节码", "firmware", "binary", "executable", "bytecode"},
		exact:    []string{"reverse engineering", "reverse engineer", "逆向工程", "逆向分析", "软件逆向", "固件破解", "破解固件", "固件逆向", "反编译固件", "反汇编固件"},
	},
}

func isLocalSecurityAuditProtocol(protocol string) bool {
	// Every gateway handler that reaches runSecurityAudit supplies a protocol
	// name. The extractor below only returns client-controlled text it
	// understands, so applying local checks to all such protocols closes gaps
	// for embeddings and newer endpoint shapes without treating metadata alone
	// as an auditable prompt.
	return strings.TrimSpace(protocol) != ""
}

func matchChinesePoliticalRule(text string) string {
	return matchConversationKeywordRule(text, conversationChinesePoliticalKeywordRules)
}

func matchPromptInjectionRule(text string) string {
	return matchConversationKeywordRule(text, conversationPromptInjectionKeywordRules)
}

func matchAuthorizationBypassRule(text string) string {
	return evaluateAuthorizationBypassRisk(text, nil).Rule
}

type localSecurityRiskMatch struct {
	Rule      string
	MatchType string
	Score     int
	Signals   int
}

func (m localSecurityRiskMatch) matched() bool {
	return m.Rule != "" && m.Score > 0
}

func higherLocalSecurityRiskMatch(current, candidate localSecurityRiskMatch) localSecurityRiskMatch {
	if candidate.Score > current.Score {
		return candidate
	}
	return current
}

// accumulateLocalSecurityRiskMatch combines independent configured-rule
// signals while capping the score. Individual rule match types still use the
// highest-confidence signal, so a rule containing exact/all/combo terms is
// not counted repeatedly.
func accumulateLocalSecurityRiskMatch(current, candidate localSecurityRiskMatch) localSecurityRiskMatch {
	if !current.matched() {
		candidate.Signals = 1
		return candidate
	}
	if !candidate.matched() {
		return current
	}
	current.Score += candidate.Score
	if current.Score > 100 {
		current.Score = 100
	}
	if current.Signals < 1 {
		current.Signals = 1
	}
	current.Signals++
	if current.Signals == 2 {
		current.Rule = current.Rule + "; " + candidate.Rule
		current.MatchType = "multiple_configured_signals"
	}
	return current
}

func evaluateAuthorizationBypassRisk(text string, configured []service.ContentModerationLocalSecurityRule) localSecurityRiskMatch {
	best := evaluateConfiguredLocalSecurityRisk(text, configured)
	best = higherLocalSecurityRiskMatch(best, evaluateRestrictedEngineeringRisk(text))
	if matchedRule := matchConversationKeywordRule(text, conversationAuthorizationBypassKeywordRules); matchedRule != "" {
		best = higherLocalSecurityRiskMatch(best, localSecurityRiskMatch{
			Rule:      "authorization_bypass (" + matchedRule + ")",
			MatchType: "built_in_combination",
			Score:     90,
			Signals:   1,
		})
	}
	return best
}

func evaluateRestrictedEngineeringRisk(text string) localSecurityRiskMatch {
	normalized := normalizeConversationRuleText(text)
	if normalized == "" {
		return localSecurityRiskMatch{}
	}
	for _, rule := range restrictedEngineeringAuditRules {
		for _, exactTerm := range rule.exact {
			normalizedExactTerm := normalizeConversationRuleText(exactTerm)
			if normalizedExactTerm != "" && strings.Contains(normalized, normalizedExactTerm) {
				return localSecurityRiskMatch{
					Rule:      rule.ruleName + " (exact: " + exactTerm + ")",
					MatchType: "built_in_exact",
					Score:     100,
					Signals:   1,
				}
			}
		}
		if action, target := matchRuleCombination(normalized, rule); action != "" {
			return localSecurityRiskMatch{
				Rule:      rule.ruleName + " (combined: " + action + "+" + target + ")",
				MatchType: "built_in_combination",
				Score:     90,
				Signals:   1,
			}
		}
	}
	return localSecurityRiskMatch{}
}

// matchRuleCombination requires an action and a target to appear close to each
// other, so unrelated mentions scattered across a long conversation no longer
// combine into a hit.
func matchRuleCombination(normalized string, rule restrictedEngineeringAuditRule) (string, string) {
	presentActions := presentConversationRuleTerms(normalized, rule.actions)
	if len(presentActions) == 0 {
		return "", ""
	}
	presentTargets := presentConversationRuleTerms(normalized, rule.targets)
	if len(presentTargets) == 0 {
		return "", ""
	}
	for _, action := range presentActions {
		for _, target := range presentTargets {
			if conversationRuleTermsAreClose(normalized, action.normalized, target.normalized) {
				return action.raw, target.raw
			}
		}
	}
	return "", ""
}

func matchConversationKeywordRule(text string, rules [][]string) string {
	normalized := normalizeConversationRuleText(text)
	if normalized == "" {
		return ""
	}
	for _, rule := range rules {
		terms := make([]string, 0, len(rule))
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
			terms = append(terms, normalizedKeyword)
		}
		if !matched || len(terms) == 0 {
			continue
		}
		if conversationRuleTermsAreClose(normalized, terms...) {
			return strings.Join(rule, "+")
		}
	}
	return ""
}

type conversationRuleTerm struct {
	raw        string
	normalized string
}

// presentConversationRuleTerms keeps the terms that occur anywhere in the text.
// Filtering first keeps the proximity scan off the (common) path where a rule
// cannot possibly match.
func presentConversationRuleTerms(normalized string, terms []string) []conversationRuleTerm {
	present := make([]conversationRuleTerm, 0, len(terms))
	for _, raw := range terms {
		normalizedTerm := normalizeConversationRuleText(raw)
		if normalizedTerm == "" || !strings.Contains(normalized, normalizedTerm) {
			continue
		}
		present = append(present, conversationRuleTerm{raw: raw, normalized: normalizedTerm})
	}
	return present
}

// conversationRuleTermsAreClose reports whether every term occurs within
// conversationRuleProximityWindow bytes of one occurrence of the first term.
func conversationRuleTermsAreClose(normalized string, terms ...string) bool {
	if len(terms) == 0 {
		return false
	}
	anchor := terms[0]
	if len(terms) == 1 {
		return strings.Contains(normalized, anchor)
	}
	for offset, scans := 0, 0; offset < len(normalized) && scans < conversationRuleMaxAnchorScans; scans++ {
		index := strings.Index(normalized[offset:], anchor)
		if index < 0 {
			return false
		}
		start := offset + index
		window := conversationRuleWindow(normalized, start, len(anchor))
		matched := true
		for _, term := range terms[1:] {
			if !strings.Contains(window, term) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
		offset = start + len(anchor)
	}
	return false
}

func conversationRuleWindow(normalized string, start, length int) string {
	from := start - conversationRuleProximityWindow
	if from < 0 {
		from = 0
	}
	to := start + length + conversationRuleProximityWindow
	if to > len(normalized) {
		to = len(normalized)
	}
	return normalized[from:to]
}

func normalizeConversationRuleText(text string) string {
	var normalized strings.Builder
	for _, r := range strings.ToLower(text) {
		if r == '\u3000' {
			continue
		}
		if r >= '\uFF01' && r <= '\uFF5E' {
			r -= '\uFEE0'
		}
		r = normalizeConversationRuleLeetspeak(r)
		if unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		normalized.WriteRune(r)
	}
	return normalized.String()
}

// normalizeConversationRuleLeetspeak catches common ASCII substitutions used
// to evade simple keyword matching. It intentionally remains narrow instead
// of attempting general transliteration.
func normalizeConversationRuleLeetspeak(r rune) rune {
	switch r {
	case '@', '4':
		return 'a'
	case '3':
		return 'e'
	case '1', '!':
		return 'i'
	case '0':
		return 'o'
	case '5', '$':
		return 's'
	case '7', '+':
		return 't'
	default:
		return r
	}
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
	// Whitelisted accounts skip every audit layer — local rules, the legacy
	// moderation service and the coordinator alike.
	if legacy != nil && legacy.IsLocalSecurityWhitelisted(c.Request.Context(), subject.UserID, securityAuditUserIdentifiers(apiKey)...) {
		if reqLog != nil {
			reqLog.Info("security_audit.whitelist_bypass",
				zap.Int64("user_id", subject.UserID), zap.String("protocol", protocol),
				zap.String("model", model), zap.String("stage", strings.TrimSpace(stage)))
		}
		if cacheCompletion {
			c.Set(securityAuditCompletedContextKey, true)
		}
		return nil
	}
	// The local rule layer honours the configured model filter, so admins can
	// exempt models (review/agent models in particular) from keyword blocking.
	localRulesApply := isLocalSecurityAuditProtocol(protocol) && (legacy == nil || legacy.IsLocalSecurityAuditedModel(c.Request.Context(), model))
	if localRulesApply {
		request := buildSecurityAuditRequest(c, apiKey, subject, protocol, model, body, stage)
		inputText := extractLocalSecurityAuditText(protocol, body)
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
		if matchedRule := matchPromptInjectionRule(inputText); matchedRule != "" {
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
					zap.String("matched_rule", matchedRule), zap.String("matched_rule_type", "conversation_prompt_injection"),
					zap.String("stage", request.Stage))
			}
			return decision
		}
		configuredRules := []service.ContentModerationLocalSecurityRule(nil)
		localPolicy := service.ContentModerationLocalSecurityPolicy{
			BlockScore:   80,
			ObserveScore: 50,
		}
		if legacy != nil {
			configuredRules = legacy.LocalSecurityRules(c.Request.Context())
			localPolicy = legacy.LocalSecurityPolicy(c.Request.Context())
		}
		riskMatch := evaluateAuthorizationBypassRisk(inputText, configuredRules)
		if riskMatch.matched() && riskMatch.Score >= localPolicy.BlockScore {
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
					zap.String("matched_rule", riskMatch.Rule), zap.String("matched_rule_type", riskMatch.MatchType),
					zap.Int("risk_score", riskMatch.Score), zap.Int("risk_signals", riskMatch.Signals), zap.Int("block_score", localPolicy.BlockScore),
					zap.String("stage", request.Stage))
			}
			return decision
		}
		if riskMatch.matched() && riskMatch.Score >= localPolicy.ObserveScore && reqLog != nil {
			reqLog.Info("security_audit.local_risk_observed",
				zap.String("request_id", request.RequestID), zap.String("matched_rule", riskMatch.Rule),
				zap.String("matched_rule_type", riskMatch.MatchType), zap.Int("risk_score", riskMatch.Score),
				zap.Int("risk_signals", riskMatch.Signals), zap.Int("observe_score", localPolicy.ObserveScore), zap.Int("block_score", localPolicy.BlockScore),
				zap.String("protocol", request.Protocol), zap.String("model", request.Model), zap.String("stage", request.Stage))
		}
		if reqLog != nil {
			reqLog.Info("security_audit.local_block_check_done",
				zap.String("request_id", request.RequestID), zap.String("decision", string(securityaudit.DecisionAllow)),
				zap.String("error_code", ""), zap.Bool("allow_next_stage", true), zap.String("matched_rule", ""),
				zap.String("matched_rule_type", ""), zap.String("stage", request.Stage))
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

func matchAuthorizationBypassRuleWithConfiguredRules(text string, configured []service.ContentModerationLocalSecurityRule) string {
	return evaluateAuthorizationBypassRisk(text, configured).Rule
}

func extractLocalSecurityAuditText(protocol string, body []byte) string {
	snapshot, err := securityaudit.ExtractPromptSnapshot(securityaudit.Request{
		Protocol: protocol,
		Body:     body,
	})
	if err == nil && strings.TrimSpace(snapshot.ScanText) != "" {
		return snapshot.ScanText
	}
	// Preserve the established extractor for legacy payload shapes that are not
	// recognized by the prompt-audit parser.
	return service.ExtractContentModerationText(protocol, body)
}

// securityAuditUserIdentifiers returns the non-numeric identifiers an admin may
// have put in the whitelist (email or username), so the whitelist is usable
// without looking up internal user IDs.
func securityAuditUserIdentifiers(apiKey *service.APIKey) []string {
	identifiers := make([]string, 0, 2)
	appendIdentifier := func(value string) {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			identifiers = append(identifiers, trimmed)
		}
	}
	if apiKey != nil && apiKey.User != nil {
		appendIdentifier(apiKey.User.Email)
		appendIdentifier(apiKey.User.Username)
	}
	return identifiers
}

func matchConfiguredLocalSecurityRules(text string, rules []service.ContentModerationLocalSecurityRule) string {
	return evaluateConfiguredLocalSecurityRisk(text, rules).Rule
}

func evaluateConfiguredLocalSecurityRisk(text string, rules []service.ContentModerationLocalSecurityRule) localSecurityRiskMatch {
	if text == "" {
		return localSecurityRiskMatch{}
	}
	normalized := normalizeConversationRuleText(text)
	best := localSecurityRiskMatch{}
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		ruleRisk := localSecurityRiskMatch{}
		for _, exact := range rule.Exact {
			term := normalizeConversationRuleText(exact)
			if term != "" && strings.Contains(normalized, term) {
				ruleRisk = higherLocalSecurityRiskMatch(ruleRisk, localSecurityRiskMatch{
					Rule:      "LocalSecurityRule:" + rule.RuleName + " (exact: " + exact + ")",
					MatchType: "configured_exact",
					Score:     configuredLocalSecurityRuleScore(rule, 100),
				})
			}
		}
		if len(rule.All) > 0 {
			terms := make([]string, 0, len(rule.All))
			allMatched := true
			for _, required := range rule.All {
				term := normalizeConversationRuleText(required)
				if term == "" || !strings.Contains(normalized, term) {
					allMatched = false
					break
				}
				terms = append(terms, term)
			}
			if allMatched && len(terms) > 0 {
				score := configuredLocalSecurityRuleScore(rule, 90)
				if !conversationRuleTermsAreClose(normalized, terms...) {
					score = minLocalSecurityRiskScore(score, 55)
				}
				ruleRisk = higherLocalSecurityRiskMatch(ruleRisk, localSecurityRiskMatch{
					Rule:      "LocalSecurityRule:" + rule.RuleName + " (all terms)",
					MatchType: "configured_all",
					Score:     score,
				})
			}
		}
		for _, action := range rule.Actions {
			actionTerm := normalizeConversationRuleText(action)
			if actionTerm == "" || !strings.Contains(normalized, actionTerm) {
				continue
			}
			for _, target := range rule.Targets {
				targetTerm := normalizeConversationRuleText(target)
				if targetTerm != "" && strings.Contains(normalized, targetTerm) {
					score := configuredLocalSecurityRuleScore(rule, 80)
					if !conversationRuleTermsAreClose(normalized, actionTerm, targetTerm) {
						score = minLocalSecurityRiskScore(score, 55)
					}
					ruleRisk = higherLocalSecurityRiskMatch(ruleRisk, localSecurityRiskMatch{
						Rule:      "LocalSecurityRule:" + rule.RuleName + " (action + target: " + action + "+" + target + ")",
						MatchType: "configured_combination",
						Score:     score,
					})
				}
			}
		}
		best = accumulateLocalSecurityRiskMatch(best, ruleRisk)
	}
	return best
}

func configuredLocalSecurityRuleScore(rule service.ContentModerationLocalSecurityRule, fallback int) int {
	if rule.Score > 0 {
		return rule.Score
	}
	return fallback
}

func minLocalSecurityRiskScore(left, right int) int {
	if left < right {
		return left
	}
	return right
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
