package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// Claims JWT Claims结构
type Claims struct {
	UserID     uint   `json:"userId"`
	CompanyID  uint   `json:"companyId"`
	Username   string `json:"username"`
	ClientType string `json:"clientType"` // web, mobile, desktop
	DeviceID   string `json:"deviceId"`
	jwt.RegisteredClaims
}

// JWT JWT工具
type JWT struct {
	secret []byte
}

// New 创建JWT工具
func New(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
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
	// 刷新Token使用相同的Claims结构，但过期时间更长
	return j.GenerateToken(userID, companyID, username, clientType, deviceID, expireDuration)
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

	return claims, nil
}

// RefreshToken 刷新Token（使用刷新Token生成新的访问Token）
func (j *JWT) RefreshToken(refreshToken string, accessTokenExpire time.Duration) (string, error) {
	// 解析刷新Token
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		return "", errors.InvalidToken
	}

	// 生成新的访问Token
	return j.GenerateToken(claims.UserID, claims.CompanyID, claims.Username, claims.ClientType, claims.DeviceID, accessTokenExpire)
}
