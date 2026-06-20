package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// 当前内容加密密钥版本。第一版固定为 1，不实现自动轮换。
const ConversationKeyVersion = 1

var (
	// ErrConversationNoKey 表示未配置加密密钥。
	ErrConversationNoKey = errors.New("conversation: encryption key not configured")
	// ErrConversationKeySize 表示密钥长度不是 32 字节（AES-256）。
	ErrConversationKeySize = errors.New("conversation: encryption key must be 32 bytes")
	// ErrConversationDecrypt 表示解密失败（密文损坏或密钥不匹配）。
	ErrConversationDecrypt = errors.New("conversation: decrypt failed")
)

// ContentEncryptor 用 AES-256-GCM 加密会话内容。明文绝不落库。
//
// nonce 随机生成并与密文分列存储（content_ciphertext / content_nonce）。
type ContentEncryptor struct {
	gcm        cipher.AEAD
	keyVersion int
}

// NewContentEncryptor 从配置的密钥字符串构造加密器。
// 密钥支持 base64 或 hex 编码，解码后必须为 32 字节。空密钥返回 ErrConversationNoKey。
func NewContentEncryptor(key string) (*ContentEncryptor, error) {
	raw, err := decodeConversationKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &ContentEncryptor{gcm: gcm, keyVersion: ConversationKeyVersion}, nil
}

// KeyVersion 返回当前密钥版本。
func (e *ContentEncryptor) KeyVersion() int {
	return e.keyVersion
}

// Encrypt 加密明文，返回密文与 nonce。
func (e *ContentEncryptor) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, e.gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = e.gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt 用密文与 nonce 还原明文。
func (e *ContentEncryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	if len(nonce) != e.gcm.NonceSize() {
		return nil, ErrConversationDecrypt
	}
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrConversationDecrypt
	}
	return plaintext, nil
}

func decodeConversationKey(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrConversationNoKey
	}
	// 优先按 hex（64 字符）解析，其次 base64。
	if len(key) == 64 {
		if raw, err := hex.DecodeString(key); err == nil {
			if len(raw) != 32 {
				return nil, ErrConversationKeySize
			}
			return raw, nil
		}
	}
	if raw, err := base64.StdEncoding.DecodeString(key); err == nil {
		if len(raw) != 32 {
			return nil, ErrConversationKeySize
		}
		return raw, nil
	}
	if raw, err := base64.RawStdEncoding.DecodeString(key); err == nil {
		if len(raw) != 32 {
			return nil, ErrConversationKeySize
		}
		return raw, nil
	}
	// 末路：按原始字节（要求恰好 32 字节）。
	if len(key) == 32 {
		return []byte(key), nil
	}
	return nil, ErrConversationKeySize
}
