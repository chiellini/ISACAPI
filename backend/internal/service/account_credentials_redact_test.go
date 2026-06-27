//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergePreservingSensitiveCreds_PreservesSensitiveWhenIncomingMissing(t *testing.T) {
	existing := map[string]any{
		"refresh_token": "rt-old",
		"access_token":  "at-old",
		"api_key":       "sk-old",
		"base_url":      "https://old.example.com",
	}
	incoming := map[string]any{
		"base_url":      "https://new.example.com",
		"model_mapping": map[string]any{"foo": "bar"},
	}

	out := MergePreservingSensitiveCreds(existing, incoming)

	require.Equal(t, "rt-old", out["refresh_token"], "incoming 没传 refresh_token，应保留 existing")
	require.Equal(t, "at-old", out["access_token"])
	require.Equal(t, "sk-old", out["api_key"])
	require.Equal(t, "https://new.example.com", out["base_url"], "非敏感键由 incoming 决定")
	require.Equal(t, map[string]any{"foo": "bar"}, out["model_mapping"])
}

func TestMergePreservingSensitiveCreds_OverwritesWhenIncomingProvidesSensitive(t *testing.T) {
	existing := map[string]any{
		"refresh_token": "rt-old",
		"api_key":       "sk-old",
	}
	incoming := map[string]any{
		"refresh_token": "rt-new",
		// 显式没传 api_key —— 应保留
	}
	out := MergePreservingSensitiveCreds(existing, incoming)
	require.Equal(t, "rt-new", out["refresh_token"], "incoming 显式传入应覆盖")
	require.Equal(t, "sk-old", out["api_key"], "incoming 没传应保留")
}

func TestMergePreservingSensitiveCreds_DoesNotMutateInputs(t *testing.T) {
	existing := map[string]any{"refresh_token": "rt"}
	incoming := map[string]any{"base_url": "x"}

	_ = MergePreservingSensitiveCreds(existing, incoming)

	require.Equal(t, "rt", existing["refresh_token"])
	require.NotContains(t, existing, "base_url")
	require.Equal(t, "x", incoming["base_url"])
	require.NotContains(t, incoming, "refresh_token")
}

func TestMergePreservingSensitiveCreds_NilInputs(t *testing.T) {
	out := MergePreservingSensitiveCreds(nil, map[string]any{"base_url": "x"})
	require.Equal(t, "x", out["base_url"])
	require.NotContains(t, out, "refresh_token")

	out2 := MergePreservingSensitiveCreds(map[string]any{"refresh_token": "rt"}, nil)
	require.Equal(t, "rt", out2["refresh_token"])
}

func TestMergePreservingSensitiveCreds_NonSensitiveDeletionAllowed(t *testing.T) {
	existing := map[string]any{
		"refresh_token": "rt",
		"base_url":      "https://old",
		"project_id":    "p1",
	}
	incoming := map[string]any{
		"base_url": "https://new",
		// 不带 project_id —— 等同删除（非敏感键由 incoming 决定）
	}
	out := MergePreservingSensitiveCreds(existing, incoming)
	require.Equal(t, "rt", out["refresh_token"], "敏感键保留")
	require.Equal(t, "https://new", out["base_url"])
	require.NotContains(t, out, "project_id", "非敏感键 incoming 不传 = 删除")
}

func TestMergePreservingReauthConfig_PreservesModelConfigWhenIncomingMissing(t *testing.T) {
	existing := map[string]any{
		"refresh_token":         "rt-old",
		"model_mapping":         map[string]any{"gpt-5.5": "gpt-5.5"},
		"compact_model_mapping": map[string]any{"a": "b"},
	}
	// 重新授权只带新 token
	incoming := map[string]any{
		"access_token":  "at-new",
		"refresh_token": "rt-new",
	}

	out := MergePreservingReauthConfig(existing, incoming)

	require.Equal(t, "at-new", out["access_token"], "新 token 由 incoming 决定")
	require.Equal(t, "rt-new", out["refresh_token"])
	require.Equal(t, map[string]any{"gpt-5.5": "gpt-5.5"}, out["model_mapping"], "重新授权应继承旧模型白名单/映射")
	require.Equal(t, map[string]any{"a": "b"}, out["compact_model_mapping"])
}

func TestMergePreservingReauthConfig_IncomingModelConfigWins(t *testing.T) {
	existing := map[string]any{
		"model_mapping": map[string]any{"old": "old"},
	}
	incoming := map[string]any{
		"model_mapping": map[string]any{"new": "new"},
	}
	out := MergePreservingReauthConfig(existing, incoming)
	require.Equal(t, map[string]any{"new": "new"}, out["model_mapping"], "incoming 显式提供应覆盖")
}

func TestMergePreservingReauthConfig_DoesNotMutateInputs(t *testing.T) {
	existing := map[string]any{"model_mapping": map[string]any{"a": "a"}}
	incoming := map[string]any{"access_token": "at"}

	_ = MergePreservingReauthConfig(existing, incoming)

	require.NotContains(t, incoming, "model_mapping")
	require.NotContains(t, existing, "access_token")
}

func TestIsSensitiveCredentialKey(t *testing.T) {
	require.True(t, IsSensitiveCredentialKey("refresh_token"))
	require.True(t, IsSensitiveCredentialKey("api_key"))
	require.True(t, IsSensitiveCredentialKey("private_key"))
	require.False(t, IsSensitiveCredentialKey("base_url"))
	require.False(t, IsSensitiveCredentialKey(""))
	require.False(t, IsSensitiveCredentialKey("model_mapping"))
}
