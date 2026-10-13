package sso

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service SSO服务接口
type Service interface {
	// 登录
	Login(req *LoginRequest) (*LoginResponse, error)

	// 刷新Token
	RefreshToken(refreshToken string) (*TokenResponse, error)

	// 登出（单个设备）
	Logout(userID uint, deviceID string, token string) error

	// 登出所有设备
	LogoutAll(userID uint) error

	// 获取用户所有活跃会话
	GetActiveSessions(userID uint) ([]*SessionInfo, error)

	// 踢出指定设备
	KickDevice(userID uint, deviceID string) error

	// 修改密码
	ChangePassword(userID uint, oldPassword string, newPassword string) error

	// 获取当前登录用户信息
	GetCurrentUser(userID uint) (*CurrentUserResponse, error)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CompanyID  *uint  `json:"companyId"`                     // 可选，优先使用域名识别的公司ID
	ClientType string `json:"clientType" binding:"required"` // web, mobile, desktop
	DeviceID   string `json:"deviceId"`                      // 设备唯一标识
	DeviceName string `json:"deviceName"`                    // 设备名称
	IPAddress  string `json:"ipAddress"`
	UserAgent  string `json:"userAgent"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refreshToken"`
	ExpiresIn    int                    `json:"expiresIn"`
	User         *UserInfo              `json:"user"`
	Company      *entity.SysCompany     `json:"company"`
	CompanyConf  *entity.SysCompanyConf `json:"companyConf"`
}

// TokenResponse Token响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

// CurrentUserResponse 当前登录用户信息响应
type CurrentUserResponse struct {
	User        *UserInfo              `json:"user"`
	Company     *entity.SysCompany     `json:"company"`
	CompanyConf *entity.SysCompanyConf `json:"companyConf"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	TrueName  string `json:"trueName"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Language  string `json:"language"`
	Timezone  string `json:"timezone"`
	Currency  string `json:"currency"`
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
	userRepo           repository.UserRepository
	jwtUtil            *jwt.JWT
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
}

// NewService 创建SSO服务
func NewService(userRepo repository.UserRepository, jwtUtil *jwt.JWT, accessTokenExpire, refreshTokenExpire int) Service {
	return &service{
		userRepo:           userRepo,
		jwtUtil:            jwtUtil,
		accessTokenExpire:  time.Duration(accessTokenExpire) * time.Second,
		refreshTokenExpire: time.Duration(refreshTokenExpire) * time.Second,
	}
}

// isPhone 判断登录输入是否为手机号（中国大陆 11 位纯数字）
func isPhone(s string) bool {
	if len(s) != 11 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// Login 登录。支持用户名或手机号登录，自动定位所属公司。
func (s *service) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		if isPhone(req.Username) {
			user, err = s.userRepo.GetUserByPhone(req.Username)
			if err != nil {
				return nil, errors.InvalidCredentials
			}
		} else {
			return nil, errors.InvalidCredentials
		}
	}

	// 验证公司ID（仅用户名登录时校验；手机号登录唯一标识用户，跳过域名强制校验）
	if req.CompanyID != nil && user.SysCompanyID != *req.CompanyID && !isPhone(req.Username) {
		// 检查公司是否配置了域名，未配置则不校验
		company, _ := s.userRepo.GetCompanyByID(user.SysCompanyID)
		if company != nil && company.Domain != nil && *company.Domain != "" {
			return nil, errors.InvalidCredentials
		}
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
		Token:          jwt.TokenHash(token), // 只存摘要，不存 token 原文
		RefreshToken:   jwt.TokenHash(refreshToken),
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

	// 先删除该设备的所有旧会话，避免唯一索引冲突
	s.userRepo.DeleteSessionByDeviceID(user.ID, deviceID)

	// 检查是否已存在该用户的会话（包括非活跃的），更新或创建
	existingSession, err := s.userRepo.GetSessionByDeviceID(user.ID, deviceID)
	if err == nil && existingSession != nil {
		session.ID = existingSession.ID
		if err := s.userRepo.UpdateSession(session); err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "更新会话失败", err)
		}
	} else {
		if err := s.userRepo.CreateSession(session); err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "创建会话失败", err)
		}
	}

	// 查询公司信息
	var company *entity.SysCompany
	var companyConf *entity.SysCompanyConf
	if user.SysCompanyID != 0 {
		company, err = s.userRepo.GetCompanyByID(user.SysCompanyID)
		if err != nil {
			logger.Warn("Failed to get company info for user",
				zap.Uint("userId", user.ID),
				zap.Uint("companyId", user.SysCompanyID),
				zap.Error(err))
			// 不阻塞登录，继续返回
			company = nil
		} else if company != nil {
			// 查询公司配置
			companyConf, err = s.userRepo.GetCompanyConfByCompanyID(user.SysCompanyID)
			if err != nil {
				logger.Warn("Failed to get company config for user",
					zap.Uint("userId", user.ID),
					zap.Uint("companyId", user.SysCompanyID),
					zap.Error(err))
				// 不阻塞登录
				companyConf = nil
			}
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
			Phone:     user.Phone,
			Email:     user.Email,
			Language:  user.Language,
			Timezone:  user.Timezone,
			Currency:  user.Currency,
			IsAdmin:   user.IsAdmin,
			CompanyID: user.SysCompanyID,
		},
		Company:     company,
		CompanyConf: companyConf.Masked(), // 密钥脱敏，避免明文下发到浏览器
	}, nil
}

// RefreshToken 刷新Token
func (s *service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// 验证刷新Token
	claims, err := s.jwtUtil.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}
	// 只接受刷新令牌：访问令牌不能用来换新令牌（否则可无限续期）
	if claims.TokenType != "refresh" {
		return nil, errors.InvalidToken
	}

	// 用户名、所属公司从数据库读取当前值，不沿用旧令牌里的：用户改名、调整公司后，新令牌立即生效
	// （业务表的 CREATE_BY 取自令牌中的用户名）。用户已停用或删除时不再续期
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		return nil, errors.InvalidToken
	}

	// 生成新的访问Token
	newToken, err := s.jwtUtil.GenerateToken(user.ID, user.SysCompanyID, user.Username, claims.ClientType, claims.DeviceID, s.accessTokenExpire)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成Token失败", err)
	}

	// 更新会话中的Token
	session, err := s.userRepo.GetSessionByDeviceID(claims.UserID, claims.DeviceID)
	if err == nil {
		session.Token = jwt.TokenHash(newToken)
		session.LastActiveTime = time.Now()
		s.userRepo.UpdateSession(session)
	}

	return &TokenResponse{
		Token:     newToken,
		ExpiresIn: int(s.accessTokenExpire.Seconds()),
	}, nil
}

// blacklistStoredToken 将会话表中保存的 token 加入黑名单：新记录保存的是摘要，升级前的旧记录保存的是 JWT 原文
func (s *service) blacklistStoredToken(stored string, expiration time.Duration) error {
	if strings.HasPrefix(stored, "eyJ") {
		return s.jwtUtil.BlacklistToken(stored, expiration)
	}
	return s.jwtUtil.BlacklistTokenHash(stored, expiration)
}

// Logout 登出（单个设备）
func (s *service) Logout(userID uint, deviceID string, token string) error {
	session, err := s.userRepo.GetSessionByDeviceID(userID, deviceID)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "会话不存在", err)
	}

	// 将token加入黑名单
	if token != "" {
		// 先尝试解析token获取过期时间
		if claims, err := s.jwtUtil.ParseToken(token); err == nil && claims.ExpiresAt != nil {
			expiresIn := time.Until(claims.ExpiresAt.Time)
			if expiresIn > 0 {
				_ = s.jwtUtil.BlacklistToken(token, expiresIn)
			}
		}
	}

	// 同时将刷新token加入黑名单
	if session.RefreshToken != "" {
		_ = s.blacklistStoredToken(session.RefreshToken, s.refreshTokenExpire)
	}

	return s.userRepo.DeleteSession(session.ID)
}

// LogoutAll 登出所有设备
func (s *service) LogoutAll(userID uint) error {
	// 先获取所有活跃会话
	sessions, err := s.userRepo.GetActiveSessions(userID)
	if err == nil {
		// 将所有会话的 token 加入黑名单
		for _, session := range sessions {
			if session.Token != "" {
				_ = s.blacklistStoredToken(session.Token, s.accessTokenExpire)
			}
			if session.RefreshToken != "" {
				_ = s.blacklistStoredToken(session.RefreshToken, s.refreshTokenExpire)
			}
		}
	}

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
	session, err := s.userRepo.GetSessionByDeviceID(userID, deviceID)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "会话不存在", err)
	}

	// 将刷新token加入黑名单
	if session.RefreshToken != "" {
		_ = s.blacklistStoredToken(session.RefreshToken, s.refreshTokenExpire)
	}

	// 也将会话中的token加入黑名单
	if session.Token != "" {
		_ = s.blacklistStoredToken(session.Token, s.accessTokenExpire)
	}

	return s.userRepo.DeleteSession(session.ID)
}

// ChangePassword 修改密码
func (s *service) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	// 查询用户
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "用户不存在", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.Wrap(errors.ErrInvalidCredentials, "旧密码不正确", err)
	}

	// 验证密码复杂度
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "密码加密失败", err)
	}

	// 更新密码
	if err := s.userRepo.UpdateUserPassword(userID, string(hashedPassword)); err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新密码失败", err)
	}

	// 退出所有会话，强制用户重新登录
	if err := s.LogoutAll(userID); err != nil {
		logger.Warn("Failed to logout all sessions after password change",
			zap.Uint("userId", userID),
			zap.Error(err))
	}

	return nil
}

// GetCurrentUser 获取当前登录用户信息
func (s *service) GetCurrentUser(userID uint) (*CurrentUserResponse, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "用户不存在", err)
	}

	resp := &CurrentUserResponse{
		User: &UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			TrueName:  user.TrueName,
			Phone:     user.Phone,
			Email:     user.Email,
			Language:  user.Language,
			Timezone:  user.Timezone,
			Currency:  user.Currency,
			IsAdmin:   user.IsAdmin,
			CompanyID: user.SysCompanyID,
		},
	}

	// 查询公司信息及配置
	if user.SysCompanyID != 0 {
		company, err := s.userRepo.GetCompanyByID(user.SysCompanyID)
		if err != nil {
			logger.Warn("Failed to get company info for current user",
				zap.Uint("userId", user.ID),
				zap.Uint("companyId", user.SysCompanyID),
				zap.Error(err))
		} else if company != nil {
			resp.Company = company

			companyConf, err := s.userRepo.GetCompanyConfByCompanyID(user.SysCompanyID)
			if err != nil {
				logger.Warn("Failed to get company config for current user",
					zap.Uint("userId", user.ID),
					zap.Uint("companyId", user.SysCompanyID),
					zap.Error(err))
			} else if companyConf != nil {
				resp.CompanyConf = companyConf.Masked()
			}
		}
	}

	return resp, nil
}

var (
	passwordUpperRegex = regexp.MustCompile(`[A-Z]`)
	passwordLowerRegex = regexp.MustCompile(`[a-z]`)
	passwordDigitRegex = regexp.MustCompile(`[0-9]`)
)

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New(errors.ErrInvalidParam, "密码长度至少8位")
	}
	if !passwordUpperRegex.MatchString(password) {
		return errors.New(errors.ErrInvalidParam, "密码需包含大写字母")
	}
	if !passwordLowerRegex.MatchString(password) {
		return errors.New(errors.ErrInvalidParam, "密码需包含小写字母")
	}
	if !passwordDigitRegex.MatchString(password) {
		return errors.New(errors.ErrInvalidParam, "密码需包含数字")
	}
	return nil
}
