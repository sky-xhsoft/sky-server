package fieldcrypto

import (
	"encoding/base64"
	"errors"
)

// ErrInvalidToken 令牌无效（格式错误、被篡改或密钥不匹配）
var ErrInvalidToken = errors.New("fieldcrypto: invalid token")

// SealToken 将载荷加密为 URL 安全的不透明令牌（AES-GCM 同时保证机密性与防篡改）
func SealToken(payload string) (string, error) {
	if globalEncryptor == nil {
		return "", ErrNotConfigured
	}
	sealed, err := globalEncryptor.EncryptBytes([]byte(payload))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

// OpenToken 解密 SealToken 生成的令牌，返回原始载荷
func OpenToken(token string) (string, error) {
	if globalEncryptor == nil {
		return "", ErrNotConfigured
	}
	sealed, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(sealed) == 0 {
		return "", ErrInvalidToken
	}
	payload, err := globalEncryptor.DecryptBytes(sealed)
	if err != nil {
		return "", ErrInvalidToken
	}
	return string(payload), nil
}
