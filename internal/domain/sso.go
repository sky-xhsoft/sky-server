package domain

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// SSO领域事件类型
const (
	EventUserLoggedIn    = "user.logged_in"    // 用户登录
	EventUserLoggedOut   = "user.logged_out"   // 用户登出
	EventTokenRefreshed  = "token.refreshed"   // Token刷新
	EventSessionKicked   = "session.kicked"    // 会话被踢出
	EventSessionExpired  = "session.expired"   // 会话过期
)

// SSO领域实体

// Session 会话实体
type Session struct {
	ID              uint             `json:"id"`
	UserID          uint             `json:"user_id"`
	CompanyID       uint             `json:"company_id"`
	Token           string           `json:"token"`
	RefreshToken    string           `json:"refresh_token"`
	ClientType      string           `json:"client_type"` // web, mobile, desktop
	DeviceID        string           `json:"device_id"`
	DeviceName      string           `json:"device_name"`
	IPAddress       string           `json:"ip_address"`
	UserAgent       string           `json:"user_agent"`
	LoginTime       time.Time        `json:"login_time"`
	LastActiveTime  time.Time        `json:"last_active_time"`
	ExpireTime      time.Time        `json:"expire_time"`
	IsActive        string           `json:"is_active"`
}

// SSO领域服务接口
type SSOService interface {
	BaseDomainService

	// 登录
	Login(ctx context.Context, req *LoginRequest) (*LoginResult, error)

	// 刷新Token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)

	// 登出
	Logout(ctx context.Context, userID uint, deviceID string) error

	// 登出所有设备
	LogoutAll(ctx context.Context, userID uint) error

	// 获取活跃会话
	GetActiveSessions(ctx context.Context, userID uint) ([]*SessionInfo, error)

	// 踢出设备
	KickDevice(ctx context.Context, userID uint, deviceID string) error
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	CompanyID  *uint  `json:"company_id"`
	ClientType string `json:"client_type"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

// LoginResult 登录结果
type LoginResult struct {
	Token         string               `json:"token"`
	RefreshToken  string               `json:"refresh_token"`
	ExpiresIn     int                  `json:"expires_in"`
	User          *UserInfo            `json:"user"`
	Company       *entity.SysCompany     `json:"company"`
	CompanyConf   *entity.SysCompanyConf `json:"company_conf"`
}

// TokenResult Token响应
type TokenResult struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	TrueName  string `json:"true_name"`
	IsAdmin   string `json:"is_admin"`
	CompanyID uint   `json:"company_id"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name"`
	ClientType     string    `json:"client_type"`
	IPAddress      string    `json:"ip_address"`
	LoginTime      time.Time `json:"login_time"`
	LastActiveTime time.Time `json:"last_active_time"`
	IsCurrent      bool      `json:"is_current"`
}

// SSO领域仓库接口
type SSORepository interface {
	GetUserByUsername(ctx context.Context, username string) (*entity.SysUser, error)
	GetUserByID(ctx context.Context, id uint) (*entity.SysUser, error)
	GetSessionByDeviceID(ctx context.Context, userID uint, deviceID string) (*Session, error)
	GetActiveSessions(ctx context.Context, userID uint) ([]*Session, error)
	CreateSession(ctx context.Context, session *Session) error
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id uint) error
	DeleteSessionByDeviceID(ctx context.Context, userID uint, deviceID string) error
	DeleteAllSessions(ctx context.Context, userID uint) error
	GetCompanyByID(ctx context.Context, id uint) (*entity.SysCompany, error)
	GetCompanyConfByCompanyID(ctx context.Context, id uint) (*entity.SysCompanyConf, error)
}

// LoginEventData 登录事件数据
type LoginEventData struct {
	UserID    uint
	Username  string
	CompanyID uint
	ClientType string
	DeviceID  string
	IPAddress string
	LoginTime time.Time
}

// LogoutEventData 登出事件数据
type LogoutEventData struct {
	UserID    uint
	Username  string
	CompanyID uint
	DeviceID  string
	LogoutTime time.Time
}

// TokenRefreshEventData Token刷新事件数据
type TokenRefreshEventData struct {
	UserID    uint
	Username  string
	CompanyID uint
	DeviceID  string
	RefreshTime time.Time
}