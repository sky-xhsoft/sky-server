package sso

import (
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/repository"
)

// MockUserRepository 简单的Mock User Repository实现
type MockUserRepository struct {
	repository.UserRepository
}

func TestSSO_ServiceCreation(t *testing.T) {
	// 测试服务创建
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo, "test-secret", 3600, 86400)

	if service == nil {
		t.Fatal("服务创建失败，返回nil")
	}
}

func TestSSO_LoginRequest_Validation(t *testing.T) {
	// 测试LoginRequest结构体
	req := &LoginRequest{
		Username:   "testuser",
		Password:   "password",
		ClientType: "web",
		DeviceID:   "device123",
	}

	if req.Username != "testuser" {
		t.Errorf("用户名不正确: %s", req.Username)
	}

	if req.Password != "password" {
		t.Errorf("密码不正确: %s", req.Password)
	}

	if req.ClientType != "web" {
		t.Errorf("客户端类型不正确: %s", req.ClientType)
	}
}

func TestSSO_TokenResponse_Validation(t *testing.T) {
	// 测试TokenResponse结构体
	resp := &TokenResponse{
		Token:     "test-token",
		ExpiresIn: 3600,
	}

	if resp.Token != "test-token" {
		t.Errorf("Token不正确: %s", resp.Token)
	}

	if resp.ExpiresIn != 3600 {
		t.Errorf("过期时间不正确: %d", resp.ExpiresIn)
	}
}

func TestSSO_UserInfo_Validation(t *testing.T) {
	// 测试UserInfo结构体
	user := &UserInfo{
		ID:        1,
		Username:  "testuser",
		TrueName:  "测试用户",
		IsAdmin:   "N",
		CompanyID: 1,
	}

	if user.ID != 1 {
		t.Errorf("用户ID不正确: %d", user.ID)
	}

	if user.Username != "testuser" {
		t.Errorf("用户名不正确: %s", user.Username)
	}

	if user.TrueName != "测试用户" {
		t.Errorf("真实姓名不正确: %s", user.TrueName)
	}

	if user.IsAdmin != "N" {
		t.Errorf("管理员标志不正确: %s", user.IsAdmin)
	}

	if user.CompanyID != 1 {
		t.Errorf("公司ID不正确: %d", user.CompanyID)
	}
}

func TestSSO_SessionInfo_Validation(t *testing.T) {
	// 测试SessionInfo结构体
	session := &SessionInfo{
		DeviceID:   "device123",
		DeviceName: "Test Device",
		ClientType: "web",
		IPAddress:  "192.168.1.1",
		IsCurrent:  true,
	}

	if session.DeviceID != "device123" {
		t.Errorf("设备ID不正确: %s", session.DeviceID)
	}

	if session.DeviceName != "Test Device" {
		t.Errorf("设备名称不正确: %s", session.DeviceName)
	}

	if session.ClientType != "web" {
		t.Errorf("客户端类型不正确: %s", session.ClientType)
	}

	if session.IPAddress != "192.168.1.1" {
		t.Errorf("IP地址不正确: %s", session.IPAddress)
	}

	if !session.IsCurrent {
		t.Error("当前会话标志不正确")
	}
}

func TestSSO_LoginResponse_Validation(t *testing.T) {
	// 测试LoginResponse结构体
	resp := &LoginResponse{
		Token:        "test-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
	}

	if resp.Token != "test-token" {
		t.Errorf("Token不正确: %s", resp.Token)
	}

	if resp.RefreshToken != "test-refresh-token" {
		t.Errorf("RefreshToken不正确: %s", resp.RefreshToken)
	}

	if resp.ExpiresIn != 3600 {
		t.Errorf("过期时间不正确: %d", resp.ExpiresIn)
	}
}
