package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	chatPolicyCacheTTL              = 5 * time.Second
	maxChatProfiles                 = 100
	maxChatSkills                   = 100
	maxChatPromptBytes              = 64 * 1024
	maxChatSkillBytes               = 32 * 1024
	maxChatTrustedInstructionsBytes = 64 * 1024
	maxChatModelIDLength            = 200
)

var chatPolicyIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// ChatPolicy is the authoritative server-side policy for the built-in chat UI.
// It deliberately supports instruction skills only; executable code and arbitrary
// outbound tool definitions are not accepted as configuration.
type ChatPolicy struct {
	Enabled  bool          `json:"enabled"`
	Profiles []ChatProfile `json:"profiles"`
	Skills   []ChatSkill   `json:"skills"`
}

type ChatProfile struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Provider      string           `json:"provider"`
	PublicModel   string           `json:"public_model"`
	UpstreamModel string           `json:"upstream_model"`
	GroupID       int64            `json:"group_id"`
	SystemPrompt  string           `json:"system_prompt"`
	SkillIDs      []string         `json:"skill_ids"`
	Enabled       bool             `json:"enabled"`
	Default       bool             `json:"default"`
	Capabilities  ChatCapabilities `json:"capabilities"`
}

type ChatCapabilities struct {
	Vision    bool `json:"vision"`
	Image     bool `json:"image"`
	WebSearch bool `json:"web_search"`
	// ContextLimit is the enforced maximum number of Unicode text characters
	// after trusted prompt/skill injection. Zero means no additional policy cap.
	ContextLimit int `json:"context_limit,omitempty"`
}

type ChatSkill struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Instructions string `json:"instructions"`
	Enabled      bool   `json:"enabled"`
}

type cachedChatPolicy struct {
	policy    *ChatPolicy
	expiresAt time.Time
}

func defaultChatPolicy() *ChatPolicy {
	return &ChatPolicy{Profiles: []ChatProfile{}, Skills: []ChatSkill{}}
}

func cloneChatPolicy(policy *ChatPolicy) *ChatPolicy {
	if policy == nil {
		return defaultChatPolicy()
	}
	clone := *policy
	clone.Profiles = append([]ChatProfile(nil), policy.Profiles...)
	for i := range clone.Profiles {
		clone.Profiles[i].SkillIDs = append([]string(nil), policy.Profiles[i].SkillIDs...)
	}
	clone.Skills = append([]ChatSkill(nil), policy.Skills...)
	return &clone
}

// GetChatPolicy returns a short-lived cached snapshot so the gateway hot path
// does not query settings for every message while still converging across nodes.
func (s *SettingService) GetChatPolicy(ctx context.Context) (*ChatPolicy, error) {
	if s == nil || s.settingRepo == nil {
		return defaultChatPolicy(), nil
	}
	if cached, ok := s.chatPolicyCache.Load().(*cachedChatPolicy); ok && cached != nil && time.Now().Before(cached.expiresAt) {
		return cloneChatPolicy(cached.policy), nil
	}
	v, err, _ := s.chatPolicySF.Do("load", func() (any, error) {
		if cached, ok := s.chatPolicyCache.Load().(*cachedChatPolicy); ok && cached != nil && time.Now().Before(cached.expiresAt) {
			return cloneChatPolicy(cached.policy), nil
		}
		raw, err := s.settingRepo.GetValue(ctx, SettingKeyChatPolicy)
		if errors.Is(err, ErrSettingNotFound) {
			policy := defaultChatPolicy()
			s.chatPolicyCache.Store(&cachedChatPolicy{policy: policy, expiresAt: time.Now().Add(chatPolicyCacheTTL)})
			return cloneChatPolicy(policy), nil
		}
		if err != nil {
			return nil, fmt.Errorf("load chat policy: %w", err)
		}
		policy := defaultChatPolicy()
		if err := json.Unmarshal([]byte(raw), policy); err != nil {
			return nil, fmt.Errorf("decode chat policy: %w", err)
		}
		if err := validateChatPolicyShape(policy); err != nil {
			return nil, fmt.Errorf("stored chat policy is invalid: %w", err)
		}
		s.chatPolicyCache.Store(&cachedChatPolicy{policy: cloneChatPolicy(policy), expiresAt: time.Now().Add(chatPolicyCacheTTL)})
		return cloneChatPolicy(policy), nil
	})
	if err != nil {
		return nil, err
	}
	policy, ok := v.(*ChatPolicy)
	if !ok || policy == nil {
		return nil, fmt.Errorf("load chat policy returned an unexpected result")
	}
	return cloneChatPolicy(policy), nil
}

// UpdateChatPolicy validates group/provider bindings before making a policy live.
func (s *SettingService) UpdateChatPolicy(ctx context.Context, policy *ChatPolicy) error {
	if s == nil || s.settingRepo == nil {
		return fmt.Errorf("chat policy storage is unavailable")
	}
	policy = cloneChatPolicy(policy)
	if err := validateChatPolicyShape(policy); err != nil {
		return err
	}
	if s.defaultSubGroupReader != nil {
		for _, profile := range policy.Profiles {
			group, err := s.defaultSubGroupReader.GetByID(ctx, profile.GroupID)
			if err != nil || group == nil {
				return fmt.Errorf("profile %q references an unavailable group", profile.ID)
			}
			if !group.IsActive() {
				return fmt.Errorf("profile %q references an inactive group", profile.ID)
			}
			if group.Platform != profile.Provider {
				return fmt.Errorf("profile %q provider %q does not match group platform %q", profile.ID, profile.Provider, group.Platform)
			}
		}
	}
	encoded, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("encode chat policy: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyChatPolicy, string(encoded)); err != nil {
		return fmt.Errorf("save chat policy: %w", err)
	}
	s.chatPolicyCache.Store(&cachedChatPolicy{policy: cloneChatPolicy(policy), expiresAt: time.Now().Add(chatPolicyCacheTTL)})
	if s.onUpdate != nil {
		s.onUpdate()
	}
	return nil
}

func validateChatPolicyShape(policy *ChatPolicy) error {
	if policy == nil {
		return fmt.Errorf("chat policy is required")
	}
	if len(policy.Profiles) > maxChatProfiles || len(policy.Skills) > maxChatSkills {
		return fmt.Errorf("chat policy exceeds the profile or skill limit")
	}
	skillIDs := make(map[string]ChatSkill, len(policy.Skills))
	for i := range policy.Skills {
		skill := &policy.Skills[i]
		skill.ID = strings.ToLower(strings.TrimSpace(skill.ID))
		skill.Name = strings.TrimSpace(skill.Name)
		skill.Description = strings.TrimSpace(skill.Description)
		skill.Instructions = strings.TrimSpace(skill.Instructions)
		if !chatPolicyIDPattern.MatchString(skill.ID) || skill.Name == "" {
			return fmt.Errorf("skill %d has an invalid id or name", i+1)
		}
		if len(skill.Name) > 100 || len(skill.Description) > 500 || len(skill.Instructions) > maxChatSkillBytes {
			return fmt.Errorf("skill %q exceeds a size limit", skill.ID)
		}
		if skill.Enabled && skill.Instructions == "" {
			return fmt.Errorf("enabled skill %q requires instructions", skill.ID)
		}
		if _, exists := skillIDs[skill.ID]; exists {
			return fmt.Errorf("duplicate skill id %q", skill.ID)
		}
		skillIDs[skill.ID] = *skill
	}

	profileIDs := make(map[string]struct{}, len(policy.Profiles))
	publicModels := make(map[string]struct{}, len(policy.Profiles))
	defaultCount := 0
	enabledCount := 0
	for i := range policy.Profiles {
		profile := &policy.Profiles[i]
		profile.ID = strings.ToLower(strings.TrimSpace(profile.ID))
		profile.Name = strings.TrimSpace(profile.Name)
		profile.Provider = strings.ToLower(strings.TrimSpace(profile.Provider))
		profile.PublicModel = strings.TrimSpace(profile.PublicModel)
		profile.UpstreamModel = strings.TrimSpace(profile.UpstreamModel)
		profile.SystemPrompt = strings.TrimSpace(profile.SystemPrompt)
		if !chatPolicyIDPattern.MatchString(profile.ID) || profile.Name == "" || profile.GroupID <= 0 {
			return fmt.Errorf("profile %d has an invalid id, name, or group", i+1)
		}
		if profile.Provider != PlatformOpenAI && profile.Provider != PlatformAnthropic && profile.Provider != PlatformGemini {
			return fmt.Errorf("profile %q has unsupported provider %q", profile.ID, profile.Provider)
		}
		if profile.PublicModel == "" || profile.UpstreamModel == "" || len(profile.PublicModel) > maxChatModelIDLength || len(profile.UpstreamModel) > maxChatModelIDLength {
			return fmt.Errorf("profile %q has an invalid model id", profile.ID)
		}
		if len(profile.Name) > 100 || len(profile.SystemPrompt) > maxChatPromptBytes {
			return fmt.Errorf("profile %q exceeds a size limit", profile.ID)
		}
		if profile.Capabilities.ContextLimit < 0 || profile.Capabilities.ContextLimit > 10_000_000 {
			return fmt.Errorf("profile %q has an invalid context limit", profile.ID)
		}
		if profile.Capabilities.Image && profile.Provider != PlatformOpenAI {
			return fmt.Errorf("profile %q enables image generation for unsupported provider %q", profile.ID, profile.Provider)
		}
		if _, exists := profileIDs[profile.ID]; exists {
			return fmt.Errorf("duplicate profile id %q", profile.ID)
		}
		profileIDs[profile.ID] = struct{}{}
		modelKey := strings.ToLower(profile.PublicModel)
		if _, exists := publicModels[modelKey]; exists {
			return fmt.Errorf("duplicate public model %q", profile.PublicModel)
		}
		publicModels[modelKey] = struct{}{}
		seenSkills := make(map[string]struct{}, len(profile.SkillIDs))
		trustedInstructionBytes := len(profile.SystemPrompt)
		for j, skillID := range profile.SkillIDs {
			skillID = strings.ToLower(strings.TrimSpace(skillID))
			profile.SkillIDs[j] = skillID
			skill, exists := skillIDs[skillID]
			if !exists || !skill.Enabled {
				return fmt.Errorf("profile %q references unavailable skill %q", profile.ID, skillID)
			}
			if _, duplicate := seenSkills[skillID]; duplicate {
				return fmt.Errorf("profile %q repeats skill %q", profile.ID, skillID)
			}
			seenSkills[skillID] = struct{}{}
			trustedInstructionBytes += len(skill.Name) + len(skill.Instructions) + len("Skill: \n\n")
		}
		if trustedInstructionBytes > maxChatTrustedInstructionsBytes {
			return fmt.Errorf("profile %q trusted instructions exceed the aggregate size limit", profile.ID)
		}
		if profile.Enabled {
			enabledCount++
			if profile.Default {
				defaultCount++
			}
		} else if profile.Default {
			return fmt.Errorf("disabled profile %q cannot be the default", profile.ID)
		}
	}
	if policy.Enabled && (enabledCount == 0 || defaultCount != 1) {
		return fmt.Errorf("an enabled policy requires exactly one enabled default profile")
	}
	return nil
}

func (p *ChatPolicy) EnabledProfileByModel(model string) (*ChatProfile, bool) {
	if p == nil || !p.Enabled {
		return nil, false
	}
	model = strings.TrimSpace(model)
	for i := range p.Profiles {
		if p.Profiles[i].Enabled && strings.EqualFold(p.Profiles[i].PublicModel, model) {
			profile := p.Profiles[i]
			profile.SkillIDs = append([]string(nil), p.Profiles[i].SkillIDs...)
			return &profile, true
		}
	}
	return nil, false
}

func (p *ChatPolicy) DefaultProfile() (*ChatProfile, bool) {
	if p == nil || !p.Enabled {
		return nil, false
	}
	for i := range p.Profiles {
		if p.Profiles[i].Enabled && p.Profiles[i].Default {
			profile := p.Profiles[i]
			profile.SkillIDs = append([]string(nil), p.Profiles[i].SkillIDs...)
			return &profile, true
		}
	}
	return nil, false
}

func (p *ChatPolicy) SystemInstructions(profile *ChatProfile) string {
	if p == nil || profile == nil {
		return ""
	}
	parts := make([]string, 0, 1+len(profile.SkillIDs))
	if value := strings.TrimSpace(profile.SystemPrompt); value != "" {
		parts = append(parts, value)
	}
	byID := make(map[string]ChatSkill, len(p.Skills))
	for _, skill := range p.Skills {
		if skill.Enabled {
			byID[skill.ID] = skill
		}
	}
	for _, skillID := range profile.SkillIDs {
		if skill, ok := byID[skillID]; ok && strings.TrimSpace(skill.Instructions) != "" {
			parts = append(parts, "Skill: "+skill.Name+"\n"+skill.Instructions)
		}
	}
	return strings.Join(parts, "\n\n")
}
