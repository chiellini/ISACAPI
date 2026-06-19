package service

import (
	"strings"
	"testing"
)

func TestSecretRedactor(t *testing.T) {
	r := NewSecretRedactor()
	cases := []struct {
		name       string
		in         string
		mustRedact []string // 不应再出现的子串
		mustKeep   []string // 应保留的子串
	}{
		{
			name:       "openai api key",
			in:         "here is my key sk-abc123DEF456ghi789jkl please use it",
			mustRedact: []string{"sk-abc123DEF456ghi789jkl"},
			mustKeep:   []string{"here is my key", "please use it"},
		},
		{
			name:       "bearer token",
			in:         "Authorization header was Bearer eyJ0abc.def.ghi rest",
			mustRedact: []string{"eyJ0abc.def.ghi"},
		},
		{
			name:       "aws access key",
			in:         "aws key AKIAIOSFODNN7EXAMPLE done",
			mustRedact: []string{"AKIAIOSFODNN7EXAMPLE"},
		},
		{
			name:       "jwt",
			in:         "token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123 end",
			mustRedact: []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123"},
		},
		{
			name:       "db connection string password",
			in:         "postgresql://user:real-password@host:5432/db",
			mustRedact: []string{"real-password"},
			mustKeep:   []string{"postgresql://user:", "@host:5432/db"},
		},
		{
			name:       "json structured secret",
			in:         `{"api_key": "supersecretvalue", "model": "gpt-5"}`,
			mustRedact: []string{"supersecretvalue"},
			mustKeep:   []string{"gpt-5", "api_key"},
		},
		{
			name:       "unquoted password field",
			in:         "password=hunter2 and more text",
			mustRedact: []string{"hunter2"},
			mustKeep:   []string{"and more text"},
		},
		{
			name:       "pem private key",
			in:         "-----BEGIN RSA PRIVATE KEY-----\nMIIabc123\n-----END RSA PRIVATE KEY-----",
			mustRedact: []string{"MIIabc123"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := r.Redact(tc.in)
			for _, s := range tc.mustRedact {
				if strings.Contains(out, s) {
					t.Fatalf("secret not redacted: %q still in %q", s, out)
				}
			}
			for _, s := range tc.mustKeep {
				if !strings.Contains(out, s) {
					t.Fatalf("expected to keep %q, got %q", s, out)
				}
			}
			if !strings.Contains(out, "[REDACTED") {
				t.Fatalf("expected a redaction marker in %q", out)
			}
		})
	}
}

func TestSecretRedactor_CleanTextUnchanged(t *testing.T) {
	r := NewSecretRedactor()
	clean := "How do I sort a slice of structs in Go by a field?"
	if got := r.Redact(clean); got != clean {
		t.Fatalf("clean text changed: %q", got)
	}
}
