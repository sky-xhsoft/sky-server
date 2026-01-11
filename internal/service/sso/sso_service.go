package sso

import (
	"time"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Service SSO服务接口
type Service interface {
	// 登录
	Login(req *LoginRequest) (*LoginResponse, error)

	// 刷新Token
	RefreshToken(refreshToken string) (*TokenResponse, error)

	// 登出（单个设备）
	Logout(userID uint, deviceID string) error

	// 登出所有设备
	LogoutAll(userID uint) error

	// 获取用户所有活跃会话
	GetActiveSessions(userID uint) ([]*SessionInfo, error)

	// 踢出指定设备
	KickDevice(userID uint, deviceID string) error
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CompanyID  uint   `json:"companyId" binding:"required"`
	ClientType string `json:"clientType" binding:"required"` // web, mobile, desktop
	DeviceID   string `json:"deviceId"`                      // 设备唯一标识
	DeviceName string `json:"deviceName"`                    // 设备名称
	IPAddress  string `json:"ipAddress"`
	UserAgent  string `json:"userAgent"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	ExpiresIn    int         `json:"expiresIn"`
	User         *UserInfo   `json:"user"`
}

// TokenResponse Token响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	TrueName  string `json:"trueName"`
	IsAdmin   string `json:"isAdmin"`
	CompanyID uint   `json:"companyId"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	DeviceID       string    `json:"deviceId"`
	DeviceName     string    `json:"deviceName"`
	ClientType     string    `json:"clientType"`
	IPAddress      string    `json:"ipAddress"`
	LoginTime      time.Time `json:"loginTime"`
	LastActiveTime time.Time `json:"lastActiveTime"`
	IsCurrent      bool      `json:"isCurrent"`
}

// service SSO服务实现
type service struct {
	userRepo            repository.UserRepository
	jwtUtil             *jwt.JWT
	accessTokenExpire   time.Duration
	refreshTokenExpire  time.Duration
}

// NewService 创建SSO服务
func NewService(userRepo repository.UserRepository, jwtSecret string, accessTokenExpire, refreshTokenExpire int) Service {
	return &service{
		userRepo:           userRepo,
		jwtUtil:            jwt.New(jwtSecret),
		accessTokenExpire:  time.Duration(accessTokenExpire) * time.Second,
		refreshTokenExpire: time.Duration(refreshTokenExpire) * time.Second,
	}
}

// Login 登录
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.InvalidCredentials
	}

	// 验证公司ID
	if user.SysCompanyID != req.CompanyID {
		return nil, errors.InvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.InvalidCredentials
	}

	// 生成设备ID（如果未提供）
	deviceID := req.DeviceID
	if deviceID == "" {
		deviceID = uuid.New().String()
	}

	// 生成Token
	token, err := s.jwtUtil.GenerateToken(user.ID, user.SysCompanyID, user.Username, req.ClientType, deviceID, s.accessTokenExpire)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成Token失败", err)
	}

	// 生成刷新Token
	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, user.SysCompanyID, user.Username, req.ClientType, deviceID, s.refreshTokenExpire)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成刷新Token失败", err)
	}

	// 创建或更新会话
	now := time.Now()
	session := &entity.SysUserSession{
		UserID:         user.ID,
		CompanyID:      user.SysCompanyID,
		Token:          token,
		RefreshToken:   refreshToken,
		ClientType:     req.ClientType,
		DeviceID:       deviceID,
		DeviceName:     req.DeviceName,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		LoginTime:      now,
		LastActiveTime: now,
		ExpireTime:     now.Add(s.refreshTokenExpire),
		IsActive:       "Y",
	}

	// 检查是否已存在该设备的会话
	existingSession, err := s.userRepo.GetSessionByDeviceID(user.ID, deviceID)
	if err == nil {
		// 更新现有会话
		session.ID = existingSession.ID
		if err := s.userRepo.UpdateSession(session); err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "更新会话失败", err)
		}
	} else {
		// 创建新会话
		if err := s.userRepo.CreateSession(session); err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "创建会话失败", err)
		}
	}

	// 返回登录响应
	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTokenExpire.Seconds()),
		User: &UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			TrueName:  user.TrueName,
			IsAdmin:   user.IsAdmin,
			CompanyID: user.SysCompanyID,
		},
	}, nil
}

// RefreshToken 刷新Token
func (s *service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// 验证刷新Token
	claims, err := s.jwtUtil.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 生成新的访问Token
	newToken, err := s.jwtUtil.GenerateToken(claims.UserID, claims.CompanyID, claims.Username, claims.ClientType, claims.DeviceID, s.accessTokenExpire)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成Token失败", err)
	}

	// 更新会话中的Token
	session, err := s.userRepo.GetSessionByDeviceID(claims.UserID, claims.DeviceID)
	if err == nil {
		session.Token = newToken
		session.LastActiveTime = time.Now()
		s.userRepo.UpdateSession(session)
	}

	return &TokenResponse{
		Token:     newToken,
		ExpiresIn: int(s.accessTokenExpire.Seconds()),
	}, nil
}

// Logout 登出（单个设备）
func (s *service) Logout(userID uint, deviceID string) error {
	session, err := s.userRepo.GetSessionByDeviceID(userID, deviceID)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "会话不存在", err)
	}

	return s.userRepo.DeleteSession(session.ID)
}

// LogoutAll 登出所有设备
func (s *service) LogoutAll(userID uint) error {
	return s.userRepo.DeleteAllSessions(userID)
}

// GetActiveSessions 获取用户所有活跃会话
func (s *service) GetActiveSessions(userID uint) ([]*SessionInfo, error) {
	sessions, err := s.userRepo.GetActiveSessions(userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询会话失败", err)
	}

	result := make([]*SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, &SessionInfo{
			DeviceID:       session.DeviceID,
			DeviceName:     session.DeviceName,
			ClientType:     session.ClientType,
			IPAddress:      session.IPAddress,
			LoginTime:      session.LoginTime,
			LastActiveTime: session.LastActiveTime,
			IsCurrent:      false, // 需要根据当前请求的Token判断
		})
	}

	return result, nil
}

// KickDevice 踢出指定设备
func (s *service) KickDevice(userID uint, deviceID string) error {
	return s.Logout(userID, deviceID)
}
