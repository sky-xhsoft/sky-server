package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// Claims JWT Claims结构
type Claims struct {
	UserID     uint   `json:"userId"`
	CompanyID  uint   `json:"companyId"`
	Username   string `json:"username"`
	ClientType string `json:"clientType"` // web, mobile, desktop
	DeviceID   string `json:"deviceId"`
	TokenType  string `json:"tokenType"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// JWT JWT工具
type JWT struct {
	secret    []byte
	redisCli  *redis.Client
	keyPrefix string
}

// Option JWT选项
type Option func(*JWT)

// WithRedis 配置Redis用于黑名单
func WithRedis(redisCli *redis.Client, keyPrefix string) Option {
	return func(j *JWT) {
		j.redisCli = redisCli
		j.keyPrefix = keyPrefix
	}
}

// New 创建JWT工具
func New(secret string, opts ...Option) *JWT {
	jwt := &JWT{
		secret:    []byte(secret),
		keyPrefix: "jwt_blacklist:",
	}
	for _, opt := range opts {
		opt(jwt)
	}
	return jwt
}

// GenerateToken 生成Token
func (j *JWT) GenerateToken(userID, companyID uint, username, clientType, deviceID string, expireDuration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		CompanyID:  companyID,
		Username:   username,
		ClientType: clientType,
		DeviceID:   deviceID,
		TokenType:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken 生成刷新Token
func (j *JWT) GenerateRefreshToken(userID, companyID uint, username, clientType, deviceID string, expireDuration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		CompanyID:  companyID,
		Username:   username,
		ClientType: clientType,
		DeviceID:   deviceID,
		TokenType:  "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析Token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.InvalidToken
}

// ValidateToken 验证Token
func (j *JWT) ValidateToken(tokenString string) (*Claims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		// 判断是否过期
		if err == jwt.ErrTokenExpired {
			return nil, errors.TokenExpired
		}
		return nil, errors.InvalidToken
	}

	// 检查是否在黑名单中
	if j.IsBlacklisted(tokenString) {
		return nil, errors.InvalidToken
	}

	return claims, nil
}

// RefreshToken 刷新Token（使用刷新Token生成新的访问Token）
func (j *JWT) RefreshToken(refreshToken string, accessTokenExpire time.Duration) (string, error) {
	// 解析刷新Token
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		return "", errors.InvalidToken
	}

	// 验证是否为刷新Token
	if claims.TokenType != "refresh" {
		return "", errors.InvalidToken
	}

	// 检查刷新Token是否在黑名单中
	if j.IsBlacklisted(refreshToken) {
		return "", errors.InvalidToken
	}

	// 生成新的访问Token
	return j.GenerateToken(claims.UserID, claims.CompanyID, claims.Username, claims.ClientType, claims.DeviceID, accessTokenExpire)
}

// IsBlacklisted 检查Token是否在黑名单中
func (j *JWT) IsBlacklisted(tokenString string) bool {
	if j.redisCli == nil {
		// 如果没有Redis，暂时不检查黑名单
		return false
	}

	ctx := context.Background()
	// 新黑名单按哈希作键；同时检查原始 token 作键的旧记录（升级前加入黑名单、尚未过期的 token）
	exists, err := j.redisCli.Exists(ctx, j.keyPrefix+TokenHash(tokenString), j.keyPrefix+tokenString).Result()
	if err != nil {
		// Redis出错时，默认不阻止访问（降级处理）
		return false
	}

	return exists > 0
}

// BlacklistToken 将Token加入黑名单
func (j *JWT) BlacklistToken(tokenString string, expiration time.Duration) error {
	if j.redisCli == nil {
		// 如果没有Redis，暂时不支持黑名单
		return nil
	}

	return j.BlacklistTokenHash(TokenHash(tokenString), expiration)
}

// BlacklistTokenHash 按 token 哈希加入黑名单（会话表只保存哈希时使用）
func (j *JWT) BlacklistTokenHash(tokenHash string, expiration time.Duration) error {
	if j.redisCli == nil {
		return nil
	}
	return j.redisCli.Set(context.Background(), j.keyPrefix+tokenHash, "1", expiration).Err()
}

// TokenHash 返回 token 的 SHA-256 十六进制摘要；会话表与黑名单只保存摘要，不保存 token 原文
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// BlacklistTokenByClaims 根据Claims将Token加入黑名单
func (j *JWT) BlacklistTokenByClaims(claims *Claims) error {
	if claims.ExpiresAt == nil {
		return nil
	}

	// 计算剩余有效期
	expiresIn := time.Until(claims.ExpiresAt.Time)
	if expiresIn <= 0 {
		// Token已过期，不需要加入黑名单
		return nil
	}

	// 注意：这种方式需要保存完整的token，这里我们需要知道token字符串
	// 通常在调用处会有token字符串，所以这个方法可能不太适用
	// 保留此方法作为占位
	return nil
}
