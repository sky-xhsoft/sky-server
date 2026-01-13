package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/sso"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	ssoService sso.Service
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(ssoService sso.Service) *AuthHandler {
	return &AuthHandler{
		ssoService: ssoService,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录，支持多设备登录，优先使用域名识别公司
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body sso.LoginRequest true "登录请求"
// @Success 200 {object} sso.LoginResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req sso.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 优先使用域名识别的公司ID
	domainCompanyID := middleware.GetCompanyID(c)
	if domainCompanyID != nil {
		// 使用域名识别的公司ID
		req.CompanyID = domainCompanyID
	} else if req.CompanyID == nil {
		// 如果既没有域名识别，也没有传递公司ID，返回错误
		utils.BadRequest(c, "无法识别公司，请配置域名或传递 companyId")
		return
	}

	// 从请求头获取IP和UserAgent
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	// 调用SSO服务登录
	resp, err := h.ssoService.Login(&req)
	if err != nil {
		if errors.Is(err, errors.InvalidCredentials) {
			utils.Unauthorized(c, "用户名或密码错误")
			return
		}
		utils.InternalError(c, "登录失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}

// RefreshToken 刷新Token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} sso.TokenResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 调用SSO服务刷新Token
	resp, err := h.ssoService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, "刷新令牌无效或已过期")
		return
	}

	utils.Success(c, resp)
}

// Logout 登出（单个设备）
// @Summary 登出
// @Description 登出当前设备
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "登出请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID（从JWT中间件设置）
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 调用SSO服务登出
	if err := h.ssoService.Logout(userID.(uint), req.DeviceID); err != nil {
		utils.InternalError(c, "登出失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "登出成功"})
}

// LogoutAll 登出所有设备
// @Summary 登出所有设备
// @Description 登出用户的所有设备
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 调用SSO服务登出所有设备
	if err := h.ssoService.LogoutAll(userID.(uint)); err != nil {
		utils.InternalError(c, "登出失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "已登出所有设备"})
}

// GetActiveSessions 获取活跃会话
// @Summary 获取活跃会话
// @Description 获取用户所有活跃的登录会话
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {array} sso.SessionInfo
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 调用SSO服务获取活跃会话
	sessions, err := h.ssoService.GetActiveSessions(userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取会话失败: "+err.Error())
		return
	}

	utils.Success(c, sessions)
}

// KickDevice 踢出指定设备
// @Summary 踢出指定设备
// @Description 强制指定设备下线
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body KickDeviceRequest true "踢出设备请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/kick-device [post]
func (h *AuthHandler) KickDevice(c *gin.Context) {
	var req KickDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	// 调用SSO服务踢出设备
	if err := h.ssoService.KickDevice(userID.(uint), req.DeviceID); err != nil {
		utils.InternalError(c, "踢出设备失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "设备已被踢出"})
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	DeviceID string `json:"deviceId" binding:"required"`
}

// KickDeviceRequest 踢出设备请求
type KickDeviceRequest struct {
	DeviceID string `json:"deviceId" binding:"required"`
}
