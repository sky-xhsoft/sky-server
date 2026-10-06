package fieldcrypto

import (
	"crypto/sha256"
	"errors"
	"strings"

	"gorm.io/gorm/schema"
)

// EncPrefix 加密字段值前缀，用于区分密文与历史明文（兼容未迁移数据）
const EncPrefix = "enc:v1:"

// ErrNotConfigured 未配置字段加密密钥
var ErrNotConfigured = errors.New("fieldcrypto: encryptor not configured")

func init() {
	// 启动即注册序列化器，保证实体上的 serializer:encrypted 标签总能被 GORM 解析
	schema.RegisterSerializer("encrypted", &EncryptedString{})
}

// Setup 使用任意长度的密钥字符串初始化全局加密器（SHA-256 派生 32 字节 AES-256 密钥）
func Setup(keyStr string) {
	key := sha256.Sum256([]byte(keyStr))
	globalEncryptor = &Encryptor{key: key[:]}
}

// IsEncrypted 判断值是否已是密文
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncPrefix)
}

// EncryptField 加密字段值；空值或已加密值原样返回
func EncryptField(plaintext string) (string, error) {
	if plaintext == "" || IsEncrypted(plaintext) {
		return plaintext, nil
	}
	if globalEncryptor == nil {
		return "", ErrNotConfigured
	}
	ciphertext, err := globalEncryptor.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return EncPrefix + ciphertext, nil
}

// DecryptField 解密字段值；无前缀的历史明文原样返回
func DecryptField(value string) (string, error) {
	if !IsEncrypted(value) {
		return value, nil
	}
	if globalEncryptor == nil {
		return "", ErrNotConfigured
	}
	return globalEncryptor.Decrypt(strings.TrimPrefix(value, EncPrefix))
}
