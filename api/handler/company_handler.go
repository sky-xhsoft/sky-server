package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/company"
)

// CompanyHandler 公司处理器
type CompanyHandler struct {
	svc company.Service
}

// NewCompanyHandler 创建公司处理器
func NewCompanyHandler(svc company.Service) *CompanyHandler {
	return &CompanyHandler{svc: svc}
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
