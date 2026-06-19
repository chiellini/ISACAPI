package service

import "testing"

func TestResolveConversationIdentity_Priority(t *testing.T) {
	base := ConversationIdentityInput{
		Secret:        "secret",
		UserID:        7,
		ContextDomain: ContextDomainAnthropicNative,
		Protocol:      ConversationProtocolAnthropic,
	}

	t.Run("explicit internal wins and is high confidence", func(t *testing.T) {
		in := base
		in.Signals = ConversationIdentitySignals{ExplicitConversationID: "conv_x", SessionID: "s"}
		id := ResolveConversationIdentity(in)
		if id.Source != IdentitySourceExplicitInternal || id.Confidence != IdentityConfidenceHigh {
			t.Fatalf("got source=%s conf=%d", id.Source, id.Confidence)
		}
		if id.ArchiveKey == "" || !id.IsMergeable() {
			t.Fatalf("strong signal must yield mergeable archive key")
		}
	})

	t.Run("previous_response_id leaves archive key empty but sets parent ref", func(t *testing.T) {
		in := base
		in.Signals = ConversationIdentitySignals{PreviousResponseID: "resp_123"}
		id := ResolveConversationIdentity(in)
		if id.Source != IdentitySourcePreviousResponseID {
			t.Fatalf("source = %s", id.Source)
		}
		if id.ArchiveKey != "" {
			t.Fatalf("previous_response_id must not self-derive an archive key")
		}
		if id.ParentRef != "resp_123" {
			t.Fatalf("parent ref = %q", id.ParentRef)
		}
	})

	t.Run("parent ref is passed through even when stronger signal wins", func(t *testing.T) {
		in := base
		in.Signals = ConversationIdentitySignals{SessionID: "s", PreviousResponseID: "resp_9"}
		id := ResolveConversationIdentity(in)
		if id.Source != IdentitySourceSessionID {
			t.Fatalf("source = %s", id.Source)
		}
		if id.ParentRef != "resp_9" {
			t.Fatalf("parent ref should still carry previous_response_id, got %q", id.ParentRef)
		}
	})

	t.Run("content prefix is low confidence and not mergeable", func(t *testing.T) {
		in := base
		in.Signals = ConversationIdentitySignals{ContentSeed: "system + first message"}
		id := ResolveConversationIdentity(in)
		if id.Source != IdentitySourceContentPrefix || id.Confidence != IdentityConfidenceLow {
			t.Fatalf("got source=%s conf=%d", id.Source, id.Confidence)
		}
		if id.IsMergeable() {
			t.Fatalf("low-confidence content fingerprint must not be mergeable")
		}
		if id.ArchiveKey == "" {
			t.Fatalf("content prefix should still produce a (temp) archive key")
		}
	})

	t.Run("no signal yields unique temp key", func(t *testing.T) {
		id1 := ResolveConversationIdentity(base)
		id2 := ResolveConversationIdentity(base)
		if id1.Source != IdentitySourceNone || id1.Confidence != IdentityConfidenceNone {
			t.Fatalf("expected none source")
		}
		if id1.ArchiveKey == id2.ArchiveKey {
			t.Fatalf("temp keys must be unique per request")
		}
	})
}

func TestResolveConversationIdentity_ContextDomainIsolation(t *testing.T) {
	mk := func(domain string) ConversationIdentity {
		return ResolveConversationIdentity(ConversationIdentityInput{
			Secret:        "secret",
			UserID:        1,
			ContextDomain: domain,
			Signals:       ConversationIdentitySignals{SessionID: "abc"},
		})
	}
	// 官方 Anthropic 与 Antigravity Claude 相同 session_id 仍隔离。
	a := mk(ContextDomainAnthropicNative)
	b := mk(ContextDomainAntigravityClaude)
	if a.ArchiveKey == b.ArchiveKey {
		t.Fatalf("same session_id across context domains must not share archive key")
	}
}

func TestDeriveArchiveKey_UserAndStability(t *testing.T) {
	k1 := deriveArchiveKey("secret", 1, ContextDomainOpenAIAPI, IdentitySourceSessionID, "abc")
	if len(k1) != 64 {
		t.Fatalf("hmac archive key should be 64 hex chars, got %d", len(k1))
	}
	if k1 != deriveArchiveKey("secret", 1, ContextDomainOpenAIAPI, IdentitySourceSessionID, "abc") {
		t.Fatalf("archive key must be stable")
	}
	if k1 == deriveArchiveKey("secret", 2, ContextDomainOpenAIAPI, IdentitySourceSessionID, "abc") {
		t.Fatalf("different users must not collide")
	}
}
