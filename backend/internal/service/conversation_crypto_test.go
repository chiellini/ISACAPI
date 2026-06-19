package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func newTestKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return k
}

func TestContentEncryptor_RoundTrip(t *testing.T) {
	key := newTestKey(t)
	for _, enc := range []string{hex.EncodeToString(key), base64.StdEncoding.EncodeToString(key)} {
		e, err := NewContentEncryptor(enc)
		if err != nil {
			t.Fatalf("new encryptor: %v", err)
		}
		plain := []byte("hello 中转站 conversation content")
		ct, nonce, err := e.Encrypt(plain)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		if bytes.Contains(ct, plain) {
			t.Fatalf("ciphertext must not contain plaintext")
		}
		got, err := e.Decrypt(ct, nonce)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}
		if !bytes.Equal(got, plain) {
			t.Fatalf("round trip mismatch: %q", got)
		}
		if e.KeyVersion() != ConversationKeyVersion {
			t.Fatalf("key version = %d", e.KeyVersion())
		}
	}
}

func TestContentEncryptor_NonceIsRandom(t *testing.T) {
	e, _ := NewContentEncryptor(hex.EncodeToString(newTestKey(t)))
	_, n1, _ := e.Encrypt([]byte("x"))
	_, n2, _ := e.Encrypt([]byte("x"))
	if bytes.Equal(n1, n2) {
		t.Fatalf("nonce must be random per encryption")
	}
}

func TestContentEncryptor_Errors(t *testing.T) {
	if _, err := NewContentEncryptor(""); err != ErrConversationNoKey {
		t.Fatalf("empty key err = %v", err)
	}
	if _, err := NewContentEncryptor("tooshort"); err != ErrConversationKeySize {
		t.Fatalf("short key err = %v", err)
	}
	e, _ := NewContentEncryptor(hex.EncodeToString(newTestKey(t)))
	ct, nonce, _ := e.Encrypt([]byte("secret"))
	ct[0] ^= 0xFF // 篡改密文
	if _, err := e.Decrypt(ct, nonce); err != ErrConversationDecrypt {
		t.Fatalf("tampered decrypt err = %v", err)
	}
}
