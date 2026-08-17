package service

import (
	"context"
	"strings"
	"testing"
)

type chatPolicySettingRepo struct{ values map[string]string }

func (r *chatPolicySettingRepo) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}
func (r *chatPolicySettingRepo) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}
func (r *chatPolicySettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}
func (r *chatPolicySettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return nil, nil
}
func (r *chatPolicySettingRepo) SetMultiple(context.Context, map[string]string) error { return nil }
func (r *chatPolicySettingRepo) GetAll(context.Context) (map[string]string, error) {
	return r.values, nil
}
func (r *chatPolicySettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type chatPolicyGroupReader struct{ groups map[int64]*Group }

func (r chatPolicyGroupReader) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

func TestChatPolicyDefaultsDisabledAndPersistsValidatedProfiles(t *testing.T) {
	repo := &chatPolicySettingRepo{values: map[string]string{}}
	svc := NewSettingService(repo, nil)
	initial, err := svc.GetChatPolicy(context.Background())
	if err != nil || initial.Enabled || len(initial.Profiles) != 0 {
		t.Fatalf("unexpected default policy: %#v, %v", initial, err)
	}
	svc.SetDefaultSubscriptionGroupReader(chatPolicyGroupReader{groups: map[int64]*Group{
		1: {ID: 1, Platform: PlatformOpenAI, Status: StatusActive},
		2: {ID: 2, Platform: PlatformAnthropic, Status: StatusActive},
		3: {ID: 3, Platform: PlatformGemini, Status: StatusActive},
	}})
	policy := &ChatPolicy{
		Enabled: true,
		Skills:  []ChatSkill{{ID: "Research", Name: "Research", Instructions: "Cite sources.", Enabled: true}},
		Profiles: []ChatProfile{
			{ID: "GPT", Name: "GPT", Provider: PlatformOpenAI, PublicModel: "gpt", UpstreamModel: "gpt-5", GroupID: 1, Enabled: true, Default: true, SkillIDs: []string{"RESEARCH"}},
			{ID: "Claude", Name: "Claude", Provider: PlatformAnthropic, PublicModel: "claude", UpstreamModel: "claude-sonnet-4", GroupID: 2, Enabled: true},
			{ID: "Gemini", Name: "Gemini", Provider: PlatformGemini, PublicModel: "gemini", UpstreamModel: "gemini-2.5-pro", GroupID: 3, Enabled: true},
		},
	}
	if err := svc.UpdateChatPolicy(context.Background(), policy); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetChatPolicy(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Profiles[0].ID != "gpt" || got.Profiles[0].SkillIDs[0] != "research" || got.Skills[0].ID != "research" {
		t.Fatalf("policy was not normalized: %#v", got)
	}
	if instructions := got.SystemInstructions(&got.Profiles[0]); instructions != "Skill: Research\nCite sources." {
		t.Fatalf("unexpected trusted instructions: %q", instructions)
	}
}

func TestChatPolicyRejectsOversizedAggregateInstructions(t *testing.T) {
	policy := &ChatPolicy{
		Enabled: true,
		Skills: []ChatSkill{
			{ID: "one", Name: "One", Instructions: strings.Repeat("a", 32*1024), Enabled: true},
			{ID: "two", Name: "Two", Instructions: strings.Repeat("b", 32*1024), Enabled: true},
		},
		Profiles: []ChatProfile{{
			ID: "gpt", Name: "GPT", Provider: PlatformOpenAI, PublicModel: "gpt", UpstreamModel: "gpt", GroupID: 1,
			Enabled: true, Default: true, SystemPrompt: "extra", SkillIDs: []string{"one", "two"},
		}},
	}
	if err := validateChatPolicyShape(policy); err == nil {
		t.Fatal("expected aggregate trusted instruction size to be rejected")
	}
}

func TestChatPolicyRejectsProviderMismatchAndUnsafeShape(t *testing.T) {
	repo := &chatPolicySettingRepo{values: map[string]string{}}
	svc := NewSettingService(repo, nil)
	svc.SetDefaultSubscriptionGroupReader(chatPolicyGroupReader{groups: map[int64]*Group{
		1: {ID: 1, Platform: PlatformOpenAI, Status: StatusActive},
	}})
	policy := &ChatPolicy{Enabled: true, Profiles: []ChatProfile{{
		ID: "claude", Name: "Claude", Provider: PlatformAnthropic, PublicModel: "claude", UpstreamModel: "claude", GroupID: 1, Enabled: true, Default: true,
	}}}
	if err := svc.UpdateChatPolicy(context.Background(), policy); err == nil {
		t.Fatal("expected provider/group mismatch to be rejected")
	}
	policy.Profiles[0].Provider = PlatformOpenAI
	policy.Profiles[0].SkillIDs = []string{"missing"}
	if err := svc.UpdateChatPolicy(context.Background(), policy); err == nil {
		t.Fatal("expected an unknown skill to be rejected")
	}
	policy.Profiles[0].SkillIDs = nil
	policy.Profiles[0].Default = false
	if err := svc.UpdateChatPolicy(context.Background(), policy); err == nil {
		t.Fatal("expected an enabled policy without a default to be rejected")
	}
}
