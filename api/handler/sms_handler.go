package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
	"github.com/sky-xhsoft/sky-server/internal/service/sms"
)

// SmsHandler 短信处理器
type SmsHandler struct {
	smsService sms.Service
}

// NewSmsHandler 创建短信处理器
func NewSmsHandler(smsService sms.Service) *SmsHandler {
	return &SmsHandler{smsService: smsService}
}

// SendVerifyCode 发送短信验证码
// @Summary 发送短信验证码
// @Tags SMS
// @Accept json
// @Produce json
// @Param body body map[string]string true "phone"
// @Success 200
// @Router /sms/send-code [post]
func (h *SmsHandler) SendVerifyCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 简单校验手机号格式（1开头的11位数字）
	if !isValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "手机号格式不正确",
		})
		return
	}

	if err := h.smsService.SendVerifyCode(c.Request.Context(), req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "发送验证码失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "验证码已发送",
	})
}

// isValidPhone 中国大陆手机号（与注册校验一致）
func isValidPhone(phone string) bool {
	return company.ValidPhone(phone)
}

// VerifySmsCode 校验短信验证码（供注册等内部使用）
func VerifySmsCode(svc sms.Service, c *gin.Context, phone, code string) bool {
	if svc == nil {
		return false
	}
	ok, _ := svc.VerifyCode(c.Request.Context(), phone, code)
	return ok
}