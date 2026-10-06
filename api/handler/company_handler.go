package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
	"github.com/sky-xhsoft/sky-server/internal/service/sms"
)

// CompanyHandler 公司处理器
type CompanyHandler struct {
	svc company.Service
	sms sms.Service
}

// NewCompanyHandler 创建公司处理器
func NewCompanyHandler(svc company.Service, smsService sms.Service) *CompanyHandler {
	return &CompanyHandler{svc: svc, sms: smsService}
}

// Register 公司注册
// @Summary 公司注册
// @Tags Company
// @Accept json
// @Produce json
// @Param body body company.RegisterRequest true "注册信息"
// @Success 200
// @Router /company/register [post]
func (h *CompanyHandler) Register(c *gin.Context) {
	var req company.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 若填写了手机号，则校验短信验证码
	if req.Phone != "" {
		if h.sms == nil || !VerifySmsCode(h.sms, c, req.Phone, strings.TrimSpace(req.VerifyCode)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40001,
				"message": "短信验证码错误或已过期",
			})
			return
		}
	}

	if err := h.svc.Register(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "注册失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "注册成功",
	})
}

// GetCompanyConfig 获取当前公司配置
// @Summary 获取当前公司配置
// @Description 获取当前登录用户所属公司的配置（敏感字段已脱敏）
// @Tags Company
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=entity.SysCompanyConf}
// @Router /company/config [get]
// @Security BearerAuth
func (h *CompanyHandler) GetCompanyConfig(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}
	id, ok := companyID.(uint)
	if !ok {
		utils.BadRequest(c, "公司ID格式错误")
		return
	}

	conf, err := h.svc.GetCompanyConfig(c.Request.Context(), id)
	if err != nil {
		utils.FailFromError(c, "获取公司配置失败: ", err)
		return
	}

	if conf == nil {
		utils.Success(c, nil)
		return
	}

	utils.Success(c, maskCompanyConfig(conf))
}

// UpdateCompanyConfig 更新当前公司配置
// @Summary 更新当前公司配置
// @Description 更新当前登录用户所属公司的配置
// @Tags Company
// @Accept json
// @Produce json
// @Param body body company.UpdateCompanyConfigRequest true "公司配置"
// @Success 200 {object} utils.Response{data=entity.SysCompanyConf}
// @Router /company/config [put]
// @Security BearerAuth
func (h *CompanyHandler) UpdateCompanyConfig(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}
	id, ok := companyID.(uint)
	if !ok {
		utils.BadRequest(c, "公司ID格式错误")
		return
	}
	fmt.Printf("[DBG] UpdateCompanyConfig companyID=%d host=%s\n", id, c.Request.Host)

	var req company.UpdateCompanyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	conf, err := h.svc.UpdateCompanyConfig(withCompanyID(c), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			utils.Error(c, http.StatusBadRequest, appErr)
			return
		}
		utils.FailFromError(c, "更新公司配置失败: ", err)
		return
	}

	utils.Success(c, maskCompanyConfig(conf))
}

// maskCompanyConfig 对公司配置敏感字段脱敏
func maskCompanyConfig(conf *entity.SysCompanyConf) *entity.SysCompanyConf {
	return conf.Masked()
}

// maskSecret 对密钥进行脱敏（保留首尾各4位，中间替换为 ****）
func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) > 8 {
		return value[:4] + "****" + value[len(value)-4:]
	}
	return "****"
}

// maskSecretID 对 ID 类凭证进行轻度脱敏（保留首尾各4位）
func maskSecretID(value string) string {
	return maskSecret(value)
}
