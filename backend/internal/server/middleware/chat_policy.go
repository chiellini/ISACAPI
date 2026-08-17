package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const chatPolicyGroupIDContextKey = "chat_policy_group_id"

const (
	maxChatPolicyMessages  = 500
	maxChatPolicyToolCalls = 100
)

var errChatContextLimitExceeded = errors.New("chat context limit exceeded")

// ChatPolicyGroupIDFromContext returns the profile-selected group for the
// built-in chat bridge. It is intentionally not read from client headers.
func ChatPolicyGroupIDFromContext(c *gin.Context) (int64, bool) {
	value, ok := c.Get(chatPolicyGroupIDContextKey)
	if !ok {
		return 0, false
	}
	groupID, ok := value.(int64)
	return groupID, ok && groupID > 0
}

// NewChatPolicyMiddleware resolves public model aliases to a super-admin
// managed provider/group and injects trusted prompts and instruction skills.
func NewChatPolicyMiddleware(settingService *service.SettingService, apiKeyService *service.APIKeyService, maxBodySize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		policy, err := settingService.GetChatPolicy(c.Request.Context())
		if err != nil {
			AbortWithError(c, http.StatusServiceUnavailable, "CHAT_POLICY_UNAVAILABLE", "chat configuration is temporarily unavailable")
			return
		}
		if policy == nil || !policy.Enabled {
			if c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/v1/models") {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
				return
			}
			AbortWithError(c, http.StatusForbidden, "CHAT_DISABLED", "chat is disabled by the administrator")
			return
		}

		subject, ok := GetAuthSubjectFromContext(c)
		if !ok {
			AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "session authentication required")
			return
		}
		groups, err := apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
		if err != nil {
			AbortWithError(c, http.StatusServiceUnavailable, "CHAT_GROUPS_UNAVAILABLE", "chat groups are temporarily unavailable")
			return
		}

		if c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/v1/models") {
			profiles := availableChatProfiles(policy, groups)
			if len(profiles) == 0 {
				AbortWithError(c, http.StatusForbidden, "CHAT_UNAVAILABLE", "no configured chat profile is available for this account")
				return
			}
			// The bridge still requires one valid internal key before reaching the
			// handler. The response itself is filtered across every available group.
			c.Set(chatPolicyGroupIDContextKey, profiles[0].GroupID)
			c.Next()
			return
		}

		isChatCompletion := c.Request.Method == http.MethodPost && strings.HasSuffix(c.Request.URL.Path, "/v1/chat/completions")
		isImageGeneration := c.Request.Method == http.MethodPost && strings.HasSuffix(c.Request.URL.Path, "/v1/images/generations")
		if !isChatCompletion && !isImageGeneration {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				AbortWithError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "chat request exceeds the configured size limit")
				return
			}
			AbortWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "failed to read chat request")
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			AbortWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "chat request must be valid JSON")
			return
		}
		if payload == nil {
			AbortWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "chat request must be a JSON object")
			return
		}
		model, _ := payload["model"].(string)
		var profile *service.ChatProfile
		var found bool
		if strings.TrimSpace(model) == "" {
			profile, found = firstAvailableDefaultProfile(policy, groups)
		} else {
			profile, found = policy.EnabledProfileByModel(model)
			if found && !chatProfileGroupAvailable(profile, groups) {
				found = false
			}
		}
		if !found || profile == nil {
			AbortWithError(c, http.StatusBadRequest, "CHAT_MODEL_UNAVAILABLE", "the selected chat model is not available")
			return
		}
		if isChatCompletion && !profile.Capabilities.Vision && chatPayloadContainsImage(payload["messages"]) {
			AbortWithError(c, http.StatusBadRequest, "CHAT_VISION_UNAVAILABLE", "the selected model is not configured for image input")
			return
		}
		if isImageGeneration && (profile.Provider != service.PlatformOpenAI || !profile.Capabilities.Image) {
			AbortWithError(c, http.StatusBadRequest, "CHAT_IMAGE_MODEL_UNAVAILABLE", "the selected model is not configured for image generation")
			return
		}
		// Security auditing must inspect only client-controlled input. Do not let
		// the trusted prompt/skills injected below reach external guards or
		// risk-control excerpts, including for image generation.
		ctx := context.WithValue(c.Request.Context(), ctxkey.ChatSecurityAuditBody, body)
		c.Request = c.Request.WithContext(ctx)

		trustedInstructions := policy.SystemInstructions(profile)
		var normalized map[string]any
		if isChatCompletion {
			normalized, err = normalizeChatCompletionPayload(payload, profile, trustedInstructions)
		} else {
			normalized, err = normalizeImageGenerationPayload(payload, profile, trustedInstructions)
		}
		if err != nil {
			code := "INVALID_REQUEST"
			if errors.Is(err, errChatContextLimitExceeded) {
				code = "CHAT_CONTEXT_LIMIT_EXCEEDED"
			}
			AbortWithError(c, http.StatusBadRequest, code, err.Error())
			return
		}
		encoded, err := json.Marshal(normalized)
		if err != nil {
			AbortWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "failed to normalize chat request")
			return
		}
		if maxBodySize > 0 && int64(len(encoded)) > maxBodySize {
			AbortWithError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "chat request exceeds the configured size limit after policy instructions are applied")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(encoded))
		c.Request.ContentLength = int64(len(encoded))
		c.Set(chatPolicyGroupIDContextKey, profile.GroupID)
		c.Next()
	}
}

func chatPayloadContainsImage(raw any) bool {
	messages, _ := raw.([]any)
	for _, item := range messages {
		message, ok := item.(map[string]any)
		if !ok {
			continue
		}
		parts, ok := message["content"].([]any)
		if !ok {
			continue
		}
		for _, part := range parts {
			value, ok := part.(map[string]any)
			if !ok {
				continue
			}
			partType, _ := value["type"].(string)
			switch strings.ToLower(strings.TrimSpace(partType)) {
			case "image_url", "input_image", "image":
				return true
			}
			// Fail closed for malformed/provider-specific image blocks that carry an
			// image field without using the canonical OpenAI type discriminator.
			if _, exists := value["image_url"]; exists {
				return true
			}
		}
	}
	return false
}

func normalizeChatCompletionPayload(payload map[string]any, profile *service.ChatProfile, trustedInstructions string) (map[string]any, error) {
	if profile == nil {
		return nil, errors.New("chat profile is required")
	}
	if err := rejectUnexpectedJSONFields(payload, "model", "messages", "stream", "tools", "tool_choice"); err != nil {
		return nil, err
	}
	if rawModel, exists := payload["model"]; exists {
		if _, ok := rawModel.(string); !ok {
			return nil, errors.New("model must be a string")
		}
	}
	rawMessages, ok := payload["messages"].([]any)
	if !ok || len(rawMessages) == 0 {
		return nil, errors.New("messages must be a non-empty array")
	}
	if len(rawMessages) > maxChatPolicyMessages {
		return nil, fmt.Errorf("messages exceeds the %d item limit", maxChatPolicyMessages)
	}
	requestedWebSearch, err := validateWebSearchRequest(payload, profile.Capabilities.WebSearch)
	if err != nil {
		return nil, err
	}
	messages, textCharacters, err := normalizeChatMessages(rawMessages, trustedInstructions, profile.Capabilities)
	if err != nil {
		return nil, err
	}
	if err := enforceConfiguredContextLimit(profile.Capabilities.ContextLimit, textCharacters); err != nil {
		return nil, err
	}

	stream := true
	if rawStream, exists := payload["stream"]; exists {
		value, ok := rawStream.(bool)
		if !ok {
			return nil, errors.New("stream must be a boolean")
		}
		stream = value
	}
	normalized := map[string]any{
		"model":    profile.UpstreamModel,
		"messages": messages,
		"stream":   stream,
	}
	if requestedWebSearch {
		normalized["tools"] = canonicalWebSearchTools()
		normalized["tool_choice"] = "auto"
	}
	return normalized, nil
}

func normalizeImageGenerationPayload(payload map[string]any, profile *service.ChatProfile, trustedInstructions string) (map[string]any, error) {
	if profile == nil {
		return nil, errors.New("chat profile is required")
	}
	if err := rejectUnexpectedJSONFields(payload, "model", "prompt", "stream", "response_format"); err != nil {
		return nil, err
	}
	if rawModel, exists := payload["model"]; exists {
		if _, ok := rawModel.(string); !ok {
			return nil, errors.New("model must be a string")
		}
	}
	prompt, ok := payload["prompt"].(string)
	if !ok || strings.TrimSpace(prompt) == "" {
		return nil, errors.New("prompt must be a non-empty string")
	}
	if rawStream, exists := payload["stream"]; exists {
		value, ok := rawStream.(bool)
		if !ok || !value {
			return nil, errors.New("image stream must be true")
		}
	}
	if rawFormat, exists := payload["response_format"]; exists {
		value, ok := rawFormat.(string)
		if !ok || !strings.EqualFold(strings.TrimSpace(value), "b64_json") {
			return nil, errors.New("image response_format must be b64_json")
		}
	}

	finalPrompt := prompt
	if trusted := strings.TrimSpace(trustedInstructions); trusted != "" {
		finalPrompt = trusted + "\n\nUser image request:\n" + prompt
	}
	if err := enforceConfiguredContextLimit(profile.Capabilities.ContextLimit, utf8.RuneCountInString(finalPrompt)); err != nil {
		return nil, err
	}
	return map[string]any{
		"model":           profile.UpstreamModel,
		"prompt":          finalPrompt,
		"stream":          true,
		"response_format": "b64_json",
	}, nil
}

func rejectUnexpectedJSONFields(value map[string]any, allowed ...string) error {
	allow := make(map[string]struct{}, len(allowed))
	for _, field := range allowed {
		allow[field] = struct{}{}
	}
	for field := range value {
		if _, ok := allow[field]; !ok {
			return fmt.Errorf("field %q is not supported by the built-in chat", field)
		}
	}
	return nil
}

func validateWebSearchRequest(payload map[string]any, allowed bool) (bool, error) {
	rawTools, toolsExist := payload["tools"]
	if !toolsExist {
		if _, choiceExists := payload["tool_choice"]; choiceExists {
			return false, errors.New("tool_choice requires the web_search tool")
		}
		return false, nil
	}
	tools, ok := rawTools.([]any)
	if !ok || len(tools) != 1 {
		return false, errors.New("tools must contain exactly one web_search tool")
	}
	tool, ok := tools[0].(map[string]any)
	if !ok {
		return false, errors.New("web_search tool must be an object")
	}
	if err := rejectUnexpectedJSONFields(tool, "type", "function"); err != nil {
		return false, err
	}
	if toolType, ok := tool["type"].(string); !ok || toolType != "function" {
		return false, errors.New("only the web_search function tool is supported")
	}
	function, ok := tool["function"].(map[string]any)
	if !ok {
		return false, errors.New("web_search function definition must be an object")
	}
	if err := rejectUnexpectedJSONFields(function, "name", "description", "parameters"); err != nil {
		return false, err
	}
	if name, ok := function["name"].(string); !ok || name != "web_search" {
		return false, errors.New("only the web_search function tool is supported")
	}
	if description, exists := function["description"]; exists {
		if _, ok := description.(string); !ok {
			return false, errors.New("web_search description must be a string")
		}
	}
	if parameters, exists := function["parameters"]; exists {
		if _, ok := parameters.(map[string]any); !ok {
			return false, errors.New("web_search parameters must be an object")
		}
	}
	if !allowed {
		return false, errors.New("web search is not enabled for the selected model")
	}
	if rawChoice, exists := payload["tool_choice"]; exists {
		choice, ok := rawChoice.(string)
		if !ok || choice != "auto" {
			return false, errors.New("tool_choice must be auto for web_search")
		}
	}
	return true, nil
}

func normalizeChatMessages(rawMessages []any, trustedInstructions string, capabilities service.ChatCapabilities) ([]any, int, error) {
	result := make([]any, 0, len(rawMessages)+1)
	textCharacters := 0
	if trusted := strings.TrimSpace(trustedInstructions); trusted != "" {
		result = append(result, map[string]any{"role": "system", "content": trusted})
		textCharacters += utf8.RuneCountInString(trusted)
	}
	pendingToolCalls := make(map[string]struct{})
	totalToolCalls := 0
	for index, raw := range rawMessages {
		message, ok := raw.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("message %d must be an object", index+1)
		}
		role, ok := message["role"].(string)
		if !ok {
			return nil, 0, fmt.Errorf("message %d role must be a string", index+1)
		}
		role = strings.ToLower(strings.TrimSpace(role))
		switch role {
		case "system", "developer", "user":
			if err := rejectUnexpectedJSONFields(message, "role", "content"); err != nil {
				return nil, 0, fmt.Errorf("message %d: %w", index+1, err)
			}
			content, count, err := normalizeUserMessageContent(message["content"], capabilities.Vision)
			if err != nil {
				return nil, 0, fmt.Errorf("message %d: %w", index+1, err)
			}
			// Only the server-managed policy may use privileged instruction roles.
			if role == "system" || role == "developer" {
				role = "user"
			}
			result = append(result, map[string]any{"role": role, "content": content})
			textCharacters += count
		case "assistant":
			if err := rejectUnexpectedJSONFields(message, "role", "content", "tool_calls"); err != nil {
				return nil, 0, fmt.Errorf("message %d: %w", index+1, err)
			}
			content, ok := message["content"].(string)
			if !ok {
				return nil, 0, fmt.Errorf("message %d assistant content must be a string", index+1)
			}
			normalized := map[string]any{"role": "assistant", "content": content}
			textCharacters += utf8.RuneCountInString(content)
			if rawCalls, exists := message["tool_calls"]; exists {
				if !capabilities.WebSearch {
					return nil, 0, fmt.Errorf("message %d contains tool calls but web search is disabled", index+1)
				}
				calls, count, err := normalizeWebSearchToolCalls(rawCalls, pendingToolCalls)
				if err != nil {
					return nil, 0, fmt.Errorf("message %d: %w", index+1, err)
				}
				totalToolCalls += len(calls)
				if totalToolCalls > maxChatPolicyToolCalls {
					return nil, 0, fmt.Errorf("tool call history exceeds the %d item limit", maxChatPolicyToolCalls)
				}
				normalized["tool_calls"] = calls
				textCharacters += count
			} else if content == "" {
				return nil, 0, fmt.Errorf("message %d assistant content cannot be empty", index+1)
			}
			result = append(result, normalized)
		case "tool":
			if !capabilities.WebSearch {
				return nil, 0, fmt.Errorf("message %d contains a tool result but web search is disabled", index+1)
			}
			if err := rejectUnexpectedJSONFields(message, "role", "tool_call_id", "content"); err != nil {
				return nil, 0, fmt.Errorf("message %d: %w", index+1, err)
			}
			callID, ok := message["tool_call_id"].(string)
			if !ok || strings.TrimSpace(callID) == "" || len(callID) > 256 {
				return nil, 0, fmt.Errorf("message %d has an invalid tool_call_id", index+1)
			}
			if _, exists := pendingToolCalls[callID]; !exists {
				return nil, 0, fmt.Errorf("message %d references an unknown tool call", index+1)
			}
			content, ok := message["content"].(string)
			if !ok {
				return nil, 0, fmt.Errorf("message %d tool content must be a string", index+1)
			}
			delete(pendingToolCalls, callID)
			result = append(result, map[string]any{"role": "tool", "tool_call_id": callID, "content": content})
			textCharacters += utf8.RuneCountInString(content)
		default:
			return nil, 0, fmt.Errorf("message %d has unsupported role %q", index+1, role)
		}
	}
	if len(pendingToolCalls) != 0 {
		return nil, 0, errors.New("every web_search tool call must have a matching tool result")
	}
	return result, textCharacters, nil
}

func normalizeUserMessageContent(raw any, allowVision bool) (any, int, error) {
	if text, ok := raw.(string); ok {
		if text == "" {
			return nil, 0, errors.New("content cannot be empty")
		}
		return text, utf8.RuneCountInString(text), nil
	}
	parts, ok := raw.([]any)
	if !ok || len(parts) == 0 {
		return nil, 0, errors.New("content must be a string or a non-empty content array")
	}
	normalized := make([]any, 0, len(parts))
	textCharacters := 0
	for index, rawPart := range parts {
		part, ok := rawPart.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("content part %d must be an object", index+1)
		}
		partType, ok := part["type"].(string)
		if !ok {
			return nil, 0, fmt.Errorf("content part %d type must be a string", index+1)
		}
		switch partType {
		case "text":
			if err := rejectUnexpectedJSONFields(part, "type", "text"); err != nil {
				return nil, 0, fmt.Errorf("content part %d: %w", index+1, err)
			}
			text, ok := part["text"].(string)
			if !ok {
				return nil, 0, fmt.Errorf("content part %d text must be a string", index+1)
			}
			normalized = append(normalized, map[string]any{"type": "text", "text": text})
			textCharacters += utf8.RuneCountInString(text)
		case "image_url":
			if !allowVision {
				return nil, 0, errors.New("image input is not enabled for the selected model")
			}
			if err := rejectUnexpectedJSONFields(part, "type", "image_url"); err != nil {
				return nil, 0, fmt.Errorf("content part %d: %w", index+1, err)
			}
			image, ok := part["image_url"].(map[string]any)
			if !ok {
				return nil, 0, fmt.Errorf("content part %d image_url must be an object", index+1)
			}
			if err := rejectUnexpectedJSONFields(image, "url"); err != nil {
				return nil, 0, fmt.Errorf("content part %d image_url: %w", index+1, err)
			}
			imageURL, ok := image["url"].(string)
			if !ok || !validRasterImageDataURL(imageURL) {
				return nil, 0, fmt.Errorf("content part %d must use a base64 PNG, JPEG, WEBP, or GIF data URL", index+1)
			}
			normalized = append(normalized, map[string]any{"type": "image_url", "image_url": map[string]any{"url": imageURL}})
		default:
			return nil, 0, fmt.Errorf("content part %d has unsupported type %q", index+1, partType)
		}
	}
	return normalized, textCharacters, nil
}

func validRasterImageDataURL(value string) bool {
	comma := strings.IndexByte(value, ',')
	if comma <= 0 || comma == len(value)-1 {
		return false
	}
	header := strings.ToLower(value[:comma])
	switch header {
	case "data:image/png;base64", "data:image/jpeg;base64", "data:image/jpg;base64", "data:image/webp;base64", "data:image/gif;base64":
	default:
		return false
	}
	encoded := value[comma+1:]
	if len(encoded)%4 != 0 {
		return false
	}
	padding := false
	for _, char := range encoded {
		switch {
		case char == '=':
			padding = true
		case char >= 'A' && char <= 'Z', char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '+', char == '/':
			if padding {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func normalizeWebSearchToolCalls(raw any, pending map[string]struct{}) ([]any, int, error) {
	calls, ok := raw.([]any)
	if !ok || len(calls) == 0 {
		return nil, 0, errors.New("tool_calls must be a non-empty array")
	}
	result := make([]any, 0, len(calls))
	textCharacters := 0
	for index, rawCall := range calls {
		call, ok := rawCall.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("tool call %d must be an object", index+1)
		}
		if err := rejectUnexpectedJSONFields(call, "id", "type", "function"); err != nil {
			return nil, 0, fmt.Errorf("tool call %d: %w", index+1, err)
		}
		callID, ok := call["id"].(string)
		if !ok || strings.TrimSpace(callID) == "" || len(callID) > 256 {
			return nil, 0, fmt.Errorf("tool call %d has an invalid id", index+1)
		}
		if _, exists := pending[callID]; exists {
			return nil, 0, fmt.Errorf("tool call %d repeats id %q", index+1, callID)
		}
		callType, ok := call["type"].(string)
		if !ok || callType != "function" {
			return nil, 0, fmt.Errorf("tool call %d must be a function", index+1)
		}
		function, ok := call["function"].(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("tool call %d function must be an object", index+1)
		}
		if err := rejectUnexpectedJSONFields(function, "name", "arguments"); err != nil {
			return nil, 0, fmt.Errorf("tool call %d: %w", index+1, err)
		}
		name, ok := function["name"].(string)
		if !ok || name != "web_search" {
			return nil, 0, fmt.Errorf("tool call %d is not web_search", index+1)
		}
		arguments, ok := function["arguments"].(string)
		if !ok {
			return nil, 0, fmt.Errorf("tool call %d arguments must be a string", index+1)
		}
		var decoded map[string]any
		if err := json.Unmarshal([]byte(arguments), &decoded); err != nil || decoded == nil {
			return nil, 0, fmt.Errorf("tool call %d arguments must be valid JSON", index+1)
		}
		if err := rejectUnexpectedJSONFields(decoded, "query"); err != nil {
			return nil, 0, fmt.Errorf("tool call %d arguments: %w", index+1, err)
		}
		query, ok := decoded["query"].(string)
		if !ok || strings.TrimSpace(query) == "" || utf8.RuneCountInString(query) > 2000 {
			return nil, 0, fmt.Errorf("tool call %d query is invalid", index+1)
		}
		canonicalArguments, _ := json.Marshal(map[string]string{"query": query})
		pending[callID] = struct{}{}
		result = append(result, map[string]any{
			"id":       callID,
			"type":     "function",
			"function": map[string]any{"name": "web_search", "arguments": string(canonicalArguments)},
		})
		textCharacters += utf8.RuneCountInString(query)
	}
	return result, textCharacters, nil
}

func canonicalWebSearchTools() []any {
	return []any{map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "web_search",
			"description": "Search the web for current information when it is necessary to answer the user.",
			"parameters": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "A focused web search query."},
				},
				"required": []string{"query"},
			},
		},
	}}
}

func enforceConfiguredContextLimit(limit, textCharacters int) error {
	if limit > 0 && textCharacters > limit {
		return fmt.Errorf("%w: request contains %d Unicode text characters; configured maximum is %d", errChatContextLimitExceeded, textCharacters, limit)
	}
	return nil
}

func firstAvailableDefaultProfile(policy *service.ChatPolicy, groups []service.Group) (*service.ChatProfile, bool) {
	profile, ok := policy.DefaultProfile()
	if ok && chatProfileGroupAvailable(profile, groups) {
		return profile, true
	}
	for i := range policy.Profiles {
		candidate := policy.Profiles[i]
		if candidate.Enabled && chatProfileGroupAvailable(&candidate, groups) {
			return &candidate, true
		}
	}
	return nil, false
}

func availableChatProfiles(policy *service.ChatPolicy, groups []service.Group) []service.ChatProfile {
	if policy == nil || !policy.Enabled {
		return nil
	}
	profiles := make([]service.ChatProfile, 0, len(policy.Profiles))
	for i := range policy.Profiles {
		profile := policy.Profiles[i]
		if profile.Enabled && chatProfileGroupAvailable(&profile, groups) {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

func chatProfileGroupAvailable(profile *service.ChatProfile, groups []service.Group) bool {
	if profile == nil {
		return false
	}
	for i := range groups {
		if groups[i].ID == profile.GroupID && groups[i].Platform == profile.Provider {
			return true
		}
	}
	return false
}

// ChatProfileGroupAvailable applies the same group/provider authorization used
// by the Chat gateway to auxiliary endpoints such as web search.
func ChatProfileGroupAvailable(profile *service.ChatProfile, groups []service.Group) bool {
	return chatProfileGroupAvailable(profile, groups)
}

func AvailableChatProfileModels(policy *service.ChatPolicy, groups []service.Group) []gin.H {
	if policy == nil || !policy.Enabled {
		return nil
	}
	profiles := availableChatProfiles(policy, groups)
	// Put the configured default first for compatibility with clients that only
	// understand the OpenAI model list fields and select the first item.
	for i := range profiles {
		if profiles[i].Default && i > 0 {
			profiles[0], profiles[i] = profiles[i], profiles[0]
			break
		}
	}
	result := make([]gin.H, 0, len(profiles))
	for i := range profiles {
		profile := &profiles[i]
		if !profile.Enabled || !chatProfileGroupAvailable(profile, groups) {
			continue
		}
		result = append(result, gin.H{
			"id":           profile.PublicModel,
			"object":       "model",
			"created":      0,
			"owned_by":     profile.Provider,
			"profile_id":   profile.ID,
			"display_name": profile.Name,
			"default":      profile.Default,
			"capabilities": profile.Capabilities,
		})
	}
	return result
}
