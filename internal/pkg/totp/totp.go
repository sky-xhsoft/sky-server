// Package totp 提供基于时间的一次性密码 (TOTP) 实现
// 符合 RFC 6238 标准，可用于多因素认证 (MFA)
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultPeriod 默认时间步长（秒）
	DefaultPeriod = 30
	// DefaultDigits 默认验证码位数
	DefaultDigits = 6
	// SecretLength 默认密钥长度（字节）
	SecretLength = 20
)

// GenerateSecret 生成随机的 Base32 编码密钥
// 长度通常为 16 或 32 字符的 Base32（20 或 40 字节原始数据）
func GenerateSecret() (string, error) {
	bytes := make([]byte, SecretLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateCode 根据密钥和时间生成 TOTP 码
func GenerateCode(secret string, t time.Time) (string, error) {
	return GenerateCodeCustom(secret, t, DefaultPeriod, DefaultDigits)
}

// GenerateCodeCustom 生成自定义参数的 TOTP 码
func GenerateCodeCustom(secret string, t time.Time, period int, digits int) (string, error) {
	// Base32 解码
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	// 计算时间计数器
	counter := uint64(t.Unix()) / uint64(period)

	// HMAC-SHA1
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	binCode := (uint32(hash[offset])&0x7f)<<24 |
		(uint32(hash[offset+1])&0xff)<<16 |
		(uint32(hash[offset+2])&0xff)<<8 |
		(uint32(hash[offset+3]) & 0xff)

	// 取模得到 N 位数字
	mod := uint32(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	code := binCode % mod

	return fmt.Sprintf("%0*d", digits, code), nil
}

// ValidateCode 验证 TOTP 码
// 默认允许 ±1 个时间窗口的容差（即 ±30 秒）
func ValidateCode(secret string, code string, t time.Time) bool {
	return ValidateCodeCustom(secret, code, t, DefaultPeriod, DefaultDigits, 1)
}

// ValidateCodeCustom 自定义参数验证
// skew 表示前后允许的时间窗口数量
func ValidateCodeCustom(secret string, code string, t time.Time, period int, digits int, skew int) bool {
	for i := -skew; i <= skew; i++ {
		testTime := t.Add(time.Duration(i) * time.Duration(period) * time.Second)
		expected, err := GenerateCodeCustom(secret, testTime, period, digits)
		if err != nil {
			return false
		}
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// BuildURI 生成 otpauth:// URI，可被 Google Authenticator 等扫码
// issuer: 应用名称（如 "Sky-Server"）
// accountName: 用户标识（如 "user@example.com"）
func BuildURI(issuer, accountName, secret string) string {
	params := url.Values{}
	params.Set("secret", secret)
	params.Set("issuer", issuer)
	params.Set("algorithm", "SHA1")
	params.Set("digits", fmt.Sprintf("%d", DefaultDigits))
	params.Set("period", fmt.Sprintf("%d", DefaultPeriod))

	label := url.PathEscape(issuer + ":" + accountName)
	return fmt.Sprintf("otpauth://totp/%s?%s", label, params.Encode())
}
