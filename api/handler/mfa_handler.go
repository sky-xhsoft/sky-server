package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/mfa"
)

// MFAHandler 多因素认证处理器
type MFAHandler struct {
	mfaService mfa.Service
}

// NewMFAHandler 创建 MFA 处理器
func NewMFAHandler(mfaService mfa.Service) *MFAHandler {
	return &MFAHandler{mfaService: mfaService}
}

// EnrollRequest 启用 MFA 请求
type EnrollRequest struct {
	AccountName string `json:"accountName" binding:"required"`
}

// VerifyRequest 验证 MFA 码请求
type VerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// EnrollResponse 启用响应
type EnrollResponse struct {
	Secret    string `json:"secret"`    // Base32 编码的密钥
	QRCodeURI string `json:"qrCodeUri"` // otpauth:// URI，可生成二维码
}

// Enroll 启用 MFA
// @Summary 启用多因素认证
// @Description 为当前用户生成 TOTP 密钥并返回 QR Code URI
// @Tags auth
// @Accept json
// @Produce json
// @Param request body EnrollRequest true "启用请求"
// @Success 200 {object} EnrollResponse
// @Router /api/v1/auth/mfa/enroll [post]
func (h *MFAHandler) Enroll(c *gin.Context) {
	var req EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		utils.InternalError(c, "用户ID类型错误")
		return
	}

	secret, qrURI, err := h.mfaService.EnrollUser(uid, req.AccountName)
	if err != nil {
		utils.FailFromError(c, "启用 MFA 失败: ", err)
		return
	}

	utils.Success(c, EnrollResponse{
		Secret:    secret,
		QRCodeURI: qrURI,
	})
}

// Verify 验证 TOTP 码
// @Summary 验证 MFA 码
// @Description 验证 TOTP 一次性密码
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyRequest true "验证请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/mfa/verify [post]
func (h *MFAHandler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		utils.InternalError(c, "用户ID类型错误")
		return
	}

	if err := h.mfaService.VerifyCode(uid, req.Code); err != nil {
		if err == mfa.ErrInvalidCode {
			utils.Unauthorized(c, "验证码无效")
			return
		}
		utils.FailFromError(c, "验证失败: ", err)
		return
	}

	utils.Success(c, gin.H{"verified": true})
}

// Disable 禁用 MFA
// @Summary 禁用多因素认证
// @Description 禁用当前用户的 MFA
// @Tags auth
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/mfa/disable [post]
func (h *MFAHandler) Disable(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		utils.InternalError(c, "用户ID类型错误")
		return
	}

	if err := h.mfaService.DisableMFA(uid); err != nil {
		utils.FailFromError(c, "禁用失败: ", err)
		return
	}

	utils.Success(c, gin.H{"disabled": true})
}

// Status 获取 MFA 启用状态
// @Summary 获取 MFA 状态
// @Description 检查当前用户是否启用 MFA
// @Tags auth
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/mfa/status [get]
func (h *MFAHandler) Status(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		utils.InternalError(c, "用户ID类型错误")
		return
	}

	enabled := h.mfaService.IsEnabled(uid)
	utils.Success(c, gin.H{"enabled": enabled})
}
