package sso

import (
	"errors"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// mockUserRepository 模拟用户仓库
type mockUserRepository struct {
	user           *entity.SysUser
	userErr        error
	company        *entity.SysCompany
	companyErr     error
	companyConf    *entity.SysCompanyConf
	companyConfErr error
	sessions       []*entity.SysUserSession
	sessionErr     error
}

func (m *mockUserRepository) GetUserByUsername(username string) (*entity.SysUser, error) {
	if m.userErr != nil {
		return nil, m.userErr
	}
	return m.user, nil
}

func (m *mockUserRepository) GetUserByID(id uint) (*entity.SysUser, error) {
	if m.userErr != nil {
		return nil, m.userErr
	}
	return m.user, nil
}

func (m *mockUserRepository) GetCompanyByID(id uint) (*entity.SysCompany, error) {
	return m.company, m.companyErr
}

func (m *mockUserRepository) GetCompanyConfByCompanyID(companyID uint) (*entity.SysCompanyConf, error) {
	return m.companyConf, m.companyConfErr
}

func (m *mockUserRepository) CreateSession(session *entity.SysUserSession) error {
	return m.sessionErr
}

func (m *mockUserRepository) UpdateSession(session *entity.SysUserSession) error {
	return m.sessionErr
}

func (m *mockUserRepository) GetSessionByDeviceID(userID uint, deviceID string) (*entity.SysUserSession, error) {
	if m.sessionErr != nil {
		return nil, m.sessionErr
	}
	if len(m.sessions) > 0 {
		return m.sessions[0], nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) GetSessionByToken(token string) (*entity.SysUserSession, error) {
	if m.sessionErr != nil {
		return nil, m.sessionErr
	}
	if len(m.sessions) > 0 {
		return m.sessions[0], nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepository) DeleteSessionByDeviceID(userID uint, deviceID string) error {
	return nil
}

func (m *mockUserRepository) DeleteSession(sessionID uint) error {
	return nil
}

func (m *mockUserRepository) DeleteAllSessions(userID uint) error {
	return nil
}

func (m *mockUserRepository) GetActiveSessions(userID uint) ([]*entity.SysUserSession, error) {
	return m.sessions, m.sessionErr
}

func (m *mockUserRepository) UpdateUserPassword(userID uint, hashedPassword string) error {
	return nil
}

// setupTestSSOService 创建测试用 SSO 服务
func setupTestSSOService(repo *mockUserRepository) Service {
	jwtUtil := jwt.New("test-secret-key-for-unit-tests-only")
	return NewService(repo, jwtUtil, 3600, 7200)
}

func TestSSO_Login_InvalidUsername(t *testing.T) {
	repo := &mockUserRepository{
		userErr: errors.New("not found"),
	}
	svc := setupTestSSOService(repo)

	_, err := svc.Login(&LoginRequest{
		Username:   "nobody",
		Password:   "anything",
		ClientType: "web",
	})

	if err == nil {
		t.Fatal("Expected error for unknown user")
	}
}

func TestSSO_Login_WrongPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		user: &entity.SysUser{
			BaseModel: entity.BaseModel{ID: 1, SysCompanyID: 1},
			Username:  "alice",
			Password:  string(hashed),
		},
	}
	svc := setupTestSSOService(repo)

	_, err := svc.Login(&LoginRequest{
		Username:   "alice",
		Password:   "wrong-password",
		ClientType: "web",
	})

	if err == nil {
		t.Fatal("Expected error for wrong password")
	}
}

func TestSSO_Login_Success_GeneratesDeviceID(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		user: &entity.SysUser{
			BaseModel: entity.BaseModel{ID: 1, SysCompanyID: 1},
			Username:  "alice",
			Password:  string(hashed),
		},
	}
	svc := setupTestSSOService(repo)

	resp, err := svc.Login(&LoginRequest{
		Username:   "alice",
		Password:   "Password123",
		ClientType: "web",
		// 不提供 DeviceID，服务应自动生成
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response")
	}
	if resp.Token == "" {
		t.Error("Token should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
}

func TestSSO_Login_CompanyMismatchWithoutDomain(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		user: &entity.SysUser{
			BaseModel: entity.BaseModel{ID: 1, SysCompanyID: 1},
			Username:  "alice",
			Password:  string(hashed),
		},
		// 公司未配置 Domain（Domain 为 nil），允许 CompanyID 不匹配
		company: &entity.SysCompany{
			BaseModel: entity.BaseModel{ID: 1},
			Domain:    nil,
		},
	}
	svc := setupTestSSOService(repo)

	mismatchedID := uint(999)
	resp, err := svc.Login(&LoginRequest{
		Username:   "alice",
		Password:   "Password123",
		ClientType: "web",
		CompanyID:  &mismatchedID,
	})

	if err != nil {
		t.Fatalf("Expected success even with mismatched CompanyID when no domain: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response")
	}
}

func TestSSO_RefreshToken_InvalidToken(t *testing.T) {
	repo := &mockUserRepository{}
	svc := setupTestSSOService(repo)

	_, err := svc.RefreshToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid refresh token")
	}
}

func TestSSO_ChangePassword_WeakPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		user: &entity.SysUser{
			BaseModel: entity.BaseModel{ID: 1},
			Username:  "alice",
			Password:  string(hashed),
		},
	}
	svc := setupTestSSOService(repo)

	// 弱密码：长度不足
	err := svc.ChangePassword(1, "OldPass123", "weak")
	if err == nil {
		t.Error("Expected error for weak password")
	}
}

func TestSSO_ChangePassword_WrongOldPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("CorrectOld123"), bcrypt.DefaultCost)
	repo := &mockUserRepository{
		user: &entity.SysUser{
			BaseModel: entity.BaseModel{ID: 1},
			Username:  "alice",
			Password:  string(hashed),
		},
	}
	svc := setupTestSSOService(repo)

	err := svc.ChangePassword(1, "WrongOld123", "NewStrongPass456")
	if err == nil {
		t.Error("Expected error for wrong old password")
	}
}

func TestSSO_GetActiveSessions(t *testing.T) {
	repo := &mockUserRepository{
		sessions: []*entity.SysUserSession{
			{
				DeviceID:   "device-1",
				DeviceName: "Chrome on macOS",
				ClientType: "web",
			},
			{
				DeviceID:   "device-2",
				DeviceName: "iPhone",
				ClientType: "mobile",
			},
		},
	}
	svc := setupTestSSOService(repo)

	sessions, err := svc.GetActiveSessions(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}
