package handler

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/email"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService email.Service
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler(emailService email.Service) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

// SendVerificationCode 发送验证码
// @Summary 发送验证码
// @Tags Email
// @Accept json
// @Produce json
// @Param body body map[string]string true "email"
// @Success 200
// @Router /email/verify-code [post]
func (h *EmailHandler) SendVerificationCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 生成6位随机验证码
	code := generateCode(6)

	if err := h.emailService.SendVerificationCode(req.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "发送验证码失败: " + err.Error(),
		})
		return
	}

	// TODO: 将验证码存入 Redis，设置5分钟过期

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "验证码已发送",
	})
}

// generateCode 生成随机数字验证码
func generateCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		b[i] = '0' + byte(n.Int64())
	}
	return string(b)
}
