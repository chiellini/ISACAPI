package service

import (
	"encoding/hex"
	"testing"
)

func TestConversationService_EventViewDecryption(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	enc, err := NewContentEncryptor(hex.EncodeToString(key))
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	svc := &ConversationService{enc: enc}

	t.Run("plaintext key_version 0", func(t *testing.T) {
		ev := &ConversationEvent{Role: "user", ContentCiphertext: []byte("hello plain"), EncryptionKeyVersion: 0}
		v := svc.toConversationEventView(ev)
		if v.Content != "hello plain" || v.Encrypted || !v.Decrypted {
			t.Fatalf("got %+v", v)
		}
	})

	t.Run("encrypted decrypts", func(t *testing.T) {
		ct, nonce, _ := enc.Encrypt([]byte("secret answer"))
		ev := &ConversationEvent{Role: "assistant", ContentCiphertext: ct, ContentNonce: nonce, EncryptionKeyVersion: 1, ContentPreview: "preview"}
		v := svc.toConversationEventView(ev)
		if v.Content != "secret answer" || !v.Encrypted || !v.Decrypted {
			t.Fatalf("got %+v", v)
		}
	})

	t.Run("encrypted without key falls back to preview", func(t *testing.T) {
		noKey := &ConversationService{enc: nil}
		ev := &ConversationEvent{Role: "assistant", ContentCiphertext: []byte("garbage"), ContentNonce: []byte("n"), EncryptionKeyVersion: 1, ContentPreview: "redacted preview"}
		v := noKey.toConversationEventView(ev)
		if v.Content != "redacted preview" || !v.Encrypted || v.Decrypted {
			t.Fatalf("got %+v", v)
		}
	})
}
