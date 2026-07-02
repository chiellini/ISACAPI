package service

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const defaultCyberRiskModelRoutingMinScore = 4

type cyberRiskPattern struct {
	pattern string
	reason  string
	score   int
}

type CyberRiskAssessment struct {
	Score   int
	Reasons []string
}

type CyberRiskModelRouteDecision struct {
	Risk        CyberRiskAssessment
	SourceModel string
	TargetModel string
}

var cyberRiskStrongPatterns = []cyberRiskPattern{
	{pattern: "deauth", reason: "wifi_deauth", score: 5},
	{pattern: "deauthentication", reason: "wifi_deauth", score: 5},
	{pattern: "evil twin", reason: "wifi_evil_twin", score: 5},
	{pattern: "rogue ap", reason: "wifi_rogue_ap", score: 5},
	{pattern: "fake ap", reason: "wifi_fake_ap", score: 5},
	{pattern: "scan ap", reason: "wifi_ap_scan", score: 4},
	{pattern: "ap scan", reason: "wifi_ap_scan", score: 4},
	{pattern: "扫描 ap", reason: "wifi_ap_scan", score: 4},
	{pattern: "扫描ap", reason: "wifi_ap_scan", score: 4},
	{pattern: "handshake capture", reason: "wifi_handshake_capture", score: 5},
	{pattern: "capture handshake", reason: "wifi_handshake_capture", score: 5},
	{pattern: "aircrack", reason: "wifi_attack_tool", score: 5},
	{pattern: "airodump", reason: "wifi_attack_tool", score: 5},
	{pattern: "aireplay", reason: "wifi_attack_tool", score: 5},
	{pattern: "hashcat", reason: "password_cracking_tool", score: 4},
	{pattern: "wpa crack", reason: "wifi_password_cracking", score: 5},
	{pattern: "wpa2 crack", reason: "wifi_password_cracking", score: 5},
	{pattern: "wifi password crack", reason: "wifi_password_cracking", score: 5},
	{pattern: "强制断连", reason: "wifi_forced_disconnect", score: 5},
	{pattern: "踢下线", reason: "wifi_forced_disconnect", score: 5},
	{pattern: "伪造 ap", reason: "wifi_fake_ap", score: 5},
	{pattern: "伪造ap", reason: "wifi_fake_ap", score: 5},
	{pattern: "metasploit", reason: "exploit_framework", score: 5},
	{pattern: "reverse shell", reason: "reverse_shell", score: 5},
	{pattern: "ransomware", reason: "malware", score: 5},
	{pattern: "keylogger", reason: "malware", score: 5},
	{pattern: "credential theft", reason: "credential_theft", score: 5},
	{pattern: "steal cookies", reason: "credential_theft", score: 5},
	{pattern: "steal token", reason: "credential_theft", score: 5},
	{pattern: "exfiltrat", reason: "data_exfiltration", score: 5},
	{pattern: "sqlmap", reason: "offensive_tool", score: 4},
	{pattern: "masscan", reason: "network_scanner_tool", score: 4},
	{pattern: "nmap", reason: "network_scanner_tool", score: 2},
}

var cyberRiskWifiTerms = []string{
	"wifi", "wi-fi", "wlan", "wireless", "ssid", "bssid", "wpa", "wpa2", "802.11",
	"ap mode", "access point", "apbegin", "热点", "无线", "路由器",
}

var cyberRiskNetworkTerms = []string{
	"port", "subnet", "host", "router", "firewall", "endpoint", "cve", "漏洞", "端口", "网段", "主机", "防火墙",
}

var cyberRiskActionTerms = []string{
	"scan", "scanning", "sniff", "sniffing", "exploit", "exploitation", "attack", "bypass",
	"inject", "injection", "spoof", "forge", "fake", "brute", "bruteforce", "crack",
	"enumerate", "recon", "fuzz", "dump", "扫描", "嗅探", "抓包", "爆破", "破解", "攻击", "绕过", "注入", "伪造", "枚举", "提权",
}

var cyberRiskWifiActionTerms = []string{
	"deauth", "deauthentication", "scan", "scanning", "sniff", "sniffing", "spoof", "forge", "fake",
	"inject", "capture", "crack", "brute", "强制", "断连", "踢下线", "伪造", "扫描", "嗅探", "抓包", "破解", "爆破",
}

func ApplyCyberRiskModelRouting(body []byte, protocol string, cfg config.CyberRiskModelRoutingConfig) ([]byte, CyberRiskModelRouteDecision, bool) {
	if !cfg.Enabled || len(body) == 0 {
		return body, CyberRiskModelRouteDecision{}, false
	}

	sourceModel := strings.TrimSpace(cfg.SourceModel)
	targetModel := strings.TrimSpace(cfg.TargetModel)
	if sourceModel == "" || targetModel == "" || strings.EqualFold(sourceModel, targetModel) {
		return body, CyberRiskModelRouteDecision{}, false
	}

	currentModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if !cyberRiskModelMatches(sourceModel, currentModel) {
		return body, CyberRiskModelRouteDecision{}, false
	}

	input := ExtractContentModerationInput(protocol, body)
	risk := AssessCyberAbuseRiskText(input.Text)
	minScore := cfg.MinScore
	if minScore <= 0 {
		minScore = defaultCyberRiskModelRoutingMinScore
	}
	if risk.Score < minScore {
		return body, CyberRiskModelRouteDecision{}, false
	}

	out, err := sjson.SetBytes(body, "model", targetModel)
	if err != nil {
		return body, CyberRiskModelRouteDecision{}, false
	}
	return out, CyberRiskModelRouteDecision{
		Risk:        risk,
		SourceModel: currentModel,
		TargetModel: targetModel,
	}, true
}

func AssessCyberAbuseRiskText(text string) CyberRiskAssessment {
	normalized := normalizeCyberRiskText(text)
	if normalized == "" {
		return CyberRiskAssessment{}
	}

	score := 0
	var reasons []string
	seenReasons := make(map[string]struct{})
	addReason := func(reason string) {
		if reason == "" {
			return
		}
		if _, ok := seenReasons[reason]; ok {
			return
		}
		seenReasons[reason] = struct{}{}
		reasons = append(reasons, reason)
	}

	for _, pattern := range cyberRiskStrongPatterns {
		if strings.Contains(normalized, pattern.pattern) {
			if _, ok := seenReasons[pattern.reason]; ok {
				continue
			}
			score += pattern.score
			addReason(pattern.reason)
		}
	}

	hasWifiContext := containsAnyCyberRiskTerm(normalized, cyberRiskWifiTerms)
	hasNetworkContext := hasWifiContext || containsAnyCyberRiskTerm(normalized, cyberRiskNetworkTerms)
	hasAction := containsAnyCyberRiskTerm(normalized, cyberRiskActionTerms)
	hasWifiAction := containsAnyCyberRiskTerm(normalized, cyberRiskWifiActionTerms)

	if hasWifiContext && hasWifiAction {
		score += 4
		addReason("wifi_abuse_context")
	}
	if hasNetworkContext && hasAction {
		score += 3
		addReason("cyber_action_target")
	}
	if containsAnyCyberRiskTerm(normalized, []string{"cve", "漏洞"}) &&
		containsAnyCyberRiskTerm(normalized, []string{"exploit", "poc", "利用", "复现"}) {
		score += 4
		addReason("vulnerability_exploit_context")
	}

	return CyberRiskAssessment{Score: score, Reasons: reasons}
}

func cyberRiskModelMatches(sourceModel string, currentModel string) bool {
	sourceModel = strings.TrimSpace(sourceModel)
	currentModel = strings.TrimSpace(currentModel)
	if sourceModel == "" || currentModel == "" {
		return false
	}
	if sourceModel == "*" {
		return true
	}
	return modelAliasLookupKey(sourceModel) == modelAliasLookupKey(currentModel)
}

func normalizeCyberRiskText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"\t", " ",
		"_", "-",
		"：", ":",
		"，", " ",
		"。", " ",
		"、", " ",
		"；", " ",
		"（", " ",
		"）", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
	)
	return strings.Join(strings.Fields(replacer.Replace(text)), " ")
}

func containsAnyCyberRiskTerm(text string, terms []string) bool {
	for _, term := range terms {
		if containsCyberRiskTerm(text, term) {
			return true
		}
	}
	return false
}

func containsCyberRiskTerm(text string, term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return false
	}
	if isASCIIToken(term) {
		return containsASCIIToken(text, term)
	}
	return strings.Contains(text, term)
}

func isASCIIToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isASCIIAlphaNum(s[i]) {
			return false
		}
	}
	return true
}

func containsASCIIToken(text string, token string) bool {
	for start := 0; start < len(text); {
		idx := strings.Index(text[start:], token)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isASCIIAlphaNum(text[idx-1])
		after := idx + len(token)
		afterOK := after >= len(text) || !isASCIIAlphaNum(text[after])
		if beforeOK && afterOK {
			return true
		}
		start = idx + len(token)
	}
	return false
}

func isASCIIAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
