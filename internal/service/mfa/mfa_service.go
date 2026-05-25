// Package mfa 多因素认证服务
// 基于 TOTP 实现，与 sso 服务配合使用
package mfa

import (
	"errors"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/totp"
)

// ErrInvalidCode 验证码无效
var ErrInvalidCode = errors.New("invalid MFA code")

// ErrSecretNotFound 用户未启用 MFA
var ErrSecretNotFound = errors.New("MFA secret not found")

// Service MFA 服务接口
type Service interface {
	// EnrollUser 为用户生成新的 MFA 密钥（开启 MFA 流程）
	// 返回 (secret, qrCodeURI, err)
	EnrollUser(userID uint, accountName string) (secret string, qrCodeURI string, err error)

	// VerifyCode 验证用户输入的 6 位 TOTP 码
	VerifyCode(userID uint, code string) error

	// SaveSecret 持久化用户的 MFA 密钥（通常在 VerifyCode 通过后调用）
	SaveSecret(userID uint, secret string) error

	// DisableMFA 禁用用户的 MFA
	DisableMFA(userID uint) error

	// IsEnabled 判断用户是否启用 MFA
	IsEnabled(userID uint) bool
}

// SecretStore 密钥存储接口（可由数据库、Redis 等实现）
type SecretStore interface {
	Get(userID uint) (string, error)
	Set(userID uint, secret string) error
	Delete(userID uint) error
}

// InMemoryStore 内存密钥存储（用于测试或单实例部署）
type InMemoryStore struct {
	mu      sync.RWMutex
	secrets map[uint]string
}

// NewInMemoryStore 创建内存存储
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		secrets: make(map[uint]string),
	}
}

// Get 获取密钥
func (s *InMemoryStore) Get(userID uint) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	secret, ok := s.secrets[userID]
	if !ok {
		return "", ErrSecretNotFound
	}
	return secret, nil
}

// Set 设置密钥
func (s *InMemoryStore) Set(userID uint, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[userID] = secret
	return nil
}

// Delete 删除密钥
func (s *InMemoryStore) Delete(userID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, userID)
	return nil
}

// service MFA 服务实现
type service struct {
	store  SecretStore
	issuer string

	// 临时密钥（用户启动 MFA 但还未确认）
	pendingMu sync.RWMutex
	pending   map[uint]string
}

// NewService 创建 MFA 服务
func NewService(store SecretStore, issuer string) Service {
	if issuer == "" {
		issuer = "Sky-Server"
	}
	return &service{
		store:   store,
		issuer:  issuer,
		pending: make(map[uint]string),
	}
}

// EnrollUser 启动 MFA 注册流程
func (s *service) EnrollUser(userID uint, accountName string) (string, string, error) {
	secret, err := totp.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	// 暂存待确认的密钥
	s.pendingMu.Lock()
	s.pending[userID] = secret
	s.pendingMu.Unlock()

	uri := totp.BuildURI(s.issuer, accountName, secret)
	return secret, uri, nil
}

// VerifyCode 验证 TOTP 码
// 如果用户尚未确认 MFA（密钥在 pending），通过后会自动保存
func (s *service) VerifyCode(userID uint, code string) error {
	// 优先使用 pending 中的密钥（用于启用流程）
	s.pendingMu.RLock()
	pendingSecret, hasPending := s.pending[userID]
	s.pendingMu.RUnlock()

	if hasPending {
		if !totp.ValidateCode(pendingSecret, code, time.Now()) {
			return ErrInvalidCode
		}
		// 验证通过，持久化并清除 pending
		if err := s.store.Set(userID, pendingSecret); err != nil {
			return err
		}
		s.pendingMu.Lock()
		delete(s.pending, userID)
		s.pendingMu.Unlock()
		return nil
	}

	// 已启用 MFA 的用户，从 store 中读取密钥
	secret, err := s.store.Get(userID)
	if err != nil {
		return err
	}
	if !totp.ValidateCode(secret, code, time.Now()) {
		return ErrInvalidCode
	}
	return nil
}

// SaveSecret 持久化密钥
func (s *service) SaveSecret(userID uint, secret string) error {
	return s.store.Set(userID, secret)
}

// DisableMFA 禁用 MFA
func (s *service) DisableMFA(userID uint) error {
	s.pendingMu.Lock()
	delete(s.pending, userID)
	s.pendingMu.Unlock()
	return s.store.Delete(userID)
}

// IsEnabled 判断 MFA 是否启用
func (s *service) IsEnabled(userID uint) bool {
	_, err := s.store.Get(userID)
	return err == nil
}
