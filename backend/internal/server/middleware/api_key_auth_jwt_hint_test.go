package middleware

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestLooksLikeJWT(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"hs256 dashboard token", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln", true},
		{"rs256 oauth token", "eyJhbGciOiJSUzI1NiIsImtpZCI6IngifQ.eyJhIjoxfQ.zzz", true},
		{"sk gateway key", "sk-abcdef0123456789ABCDEF", false},
		{"empty", "", false},
		{"eyJ prefix but no segments", "eyJhbGci", false},
		{"three segments but not base64 json header", "a.b.c", false},
	}
	for _, tc := range cases {
		if got := looksLikeJWT(tc.in); got != tc.want {
			t.Errorf("%s: looksLikeJWT(%q) = %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestInvalidAPIKeyMessage(t *testing.T) {
	cfg := &config.Config{}
	cfg.Default.APIKeyPrefix = "sk-"

	jwt := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln"
	msg := invalidAPIKeyMessage(cfg, jwt)
	if !strings.Contains(strings.ToLower(msg), "jwt") {
		t.Errorf("JWT-shaped credential should produce a JWT hint, got %q", msg)
	}
	if !strings.Contains(msg, `"sk-"`) {
		t.Errorf("hint should mention the configured key prefix, got %q", msg)
	}

	// Non-JWT credentials keep the generic message verbatim.
	if got := invalidAPIKeyMessage(cfg, "totally-wrong-key"); got != "Invalid API key" {
		t.Errorf("non-JWT credential = %q, want %q", got, "Invalid API key")
	}

	// Custom prefix is reflected; nil cfg falls back to sk-.
	cfg.Default.APIKeyPrefix = "isac-"
	if got := invalidAPIKeyMessage(cfg, jwt); !strings.Contains(got, `"isac-"`) {
		t.Errorf("custom prefix not reflected: %q", got)
	}
	if got := invalidAPIKeyMessage(nil, jwt); !strings.Contains(got, `"sk-"`) {
		t.Errorf("nil cfg should fall back to sk-, got %q", got)
	}
}
