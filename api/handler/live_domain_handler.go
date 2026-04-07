package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
)

// LiveDomainHandler 直播域名处理器
type LiveDomainHandler struct {
	liveService liveService.Service
}

// NewLiveDomainHandler 创建直播域名处理器
func NewLiveDomainHandler(service liveService.Service) *LiveDomainHandler {
	return &LiveDomainHandler{
		liveService: service,
	}
}

// AddDomainRequest 添加域名请求
type AddDomainRequest struct {
	DomainName string `json:"domainName" binding:"required"`    // 域名
	DomainType int64  `json:"domainType" binding:"gte=0,lte=1"` // 0-推流域名，1-播放域名
}

// AddDomain 添加域名
// @Summary 添加直播域名
// @Description 添加推流域名或播放域名
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param request body AddDomainRequest true "添加域名请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/domains [post]
func (h *LiveDomainHandler) AddDomain(c *gin.Context) {
	var req AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	ctx := withCompanyID(c)

	err := h.liveService.AddDomain(ctx, &live.AddDomainRequest{
		DomainName: req.DomainName,
		DomainType: req.DomainType,
	})

	if err != nil {
		// 只返回简洁的错误信息，不显示腾讯云SDK的详细错误
		utils.InternalError(c, "域名添加失败")
		return
	}

	utils.Success(c, gin.H{
		"domainName": req.DomainName,
		"message":    "添加域名成功",
	})
}

// ListDomains 查询域名列表
// @Summary 查询域名列表
// @Description 查询直播域名列表
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainType query int false "域名类型：0-推流域名，1-播放域名"
// @Success 200 {object} utils.Response{data=[]live.DomainInfo}
// @Router /api/v1/live/domains [get]
func (h *LiveDomainHandler) ListDomains(c *gin.Context) {
	var domainType *int64
	if typeStr := c.Query("domainType"); typeStr != "" {
		if t, err := strconv.ParseInt(typeStr, 10, 64); err == nil {
			domainType = &t
		}
	}

	ctx := withCompanyID(c)

	domains, err := h.liveService.ListDomains(ctx, domainType)
	if err != nil {
		utils.InternalError(c, "查询域名列表失败")
		return
	}

	utils.Success(c, gin.H{
		"domains": domains,
		"total":   len(domains),
	})
}

// GetDomain 查询域名信息
// @Summary 查询域名信息
// @Description 查询指定域名的详细信息
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName path string true "域名"
// @Success 200 {object} utils.Response{data=live.DomainInfo}
// @Router /api/v1/live/domains/{domainName} [get]
func (h *LiveDomainHandler) GetDomain(c *gin.Context) {
	domainName := c.Param("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	ctx := withCompanyID(c)

	domain, err := h.liveService.DescribeDomain(ctx, domainName)
	if err != nil {
		utils.InternalError(c, "查询域名信息失败")
		return
	}

	utils.Success(c, domain)
}

// DeleteDomain 删除域名
// @Summary 删除域名
// @Description 删除指定的直播域名
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName path string true "域名"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/domains/{domainName} [delete]
func (h *LiveDomainHandler) DeleteDomain(c *gin.Context) {
	domainName := c.Param("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	ctx := withCompanyID(c)

	err := h.liveService.DeleteDomain(ctx, domainName)
	if err != nil {
		// 记录详细错误日志
		zap.L().Error("删除域名失败",
			zap.String("domain", domainName),
			zap.Error(err))
		utils.InternalError(c, "删除域名失败")
		return
	}

	utils.Success(c, gin.H{
		"domainName": domainName,
		"message":    "删除域名成功",
	})
}

// EnableDomain 启用域名
// @Summary 启用域名
// @Description 启用指定的直播域名
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName path string true "域名"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/domains/{domainName}/enable [post]
func (h *LiveDomainHandler) EnableDomain(c *gin.Context) {
	domainName := c.Param("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	ctx := withCompanyID(c)

	err := h.liveService.EnableDomain(ctx, domainName)
	if err != nil {
		utils.InternalError(c, "启用域名失败")
		return
	}

	utils.Success(c, gin.H{
		"domainName": domainName,
		"message":    "启用域名成功",
	})
}

// ForbidDomain 禁用域名
// @Summary 禁用域名
// @Description 禁用指定的直播域名
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName path string true "域名"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/domains/{domainName}/forbid [post]
func (h *LiveDomainHandler) ForbidDomain(c *gin.Context) {
	domainName := c.Param("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	ctx := withCompanyID(c)

	err := h.liveService.ForbidDomain(ctx, domainName)
	if err != nil {
		utils.InternalError(c, "禁用域名失败")
		return
	}

	utils.Success(c, gin.H{
		"domainName": domainName,
		"message":    "禁用域名成功",
	})
}

// VerifyDomainOwner 验证域名归属
// @Summary 验证域名归属
// @Description 验证域名归属，检查CNAME配置状态
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName query string true "域名"
// @Param verifyType query string false "验证类型：dnsCheck(默认) 或 fileCheck"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/domains/verify [get]
func (h *LiveDomainHandler) VerifyDomainOwner(c *gin.Context) {
	domainName := c.Query("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	verifyType := c.DefaultQuery("verifyType", "dnsCheck")

	ctx := withCompanyID(c)
	result, err := h.liveService.AuthenticateDomainOwner(ctx, domainName, verifyType)
	if err != nil {
		utils.InternalError(c, "验证域名归属失败")
		return
	}

	// 解析验证状态
	var statusText string
	switch result.Status {
	case 0:
		statusText = "未验证"
	case 1:
		statusText = "验证成功"
	case 2:
		statusText = "验证失败"
	default:
		statusText = "未知状态"
	}

	utils.Success(c, gin.H{
		"status":     result.Status,
		"statusText": statusText,
		"content":    result.Content,
		"mainDomain": result.MainDomain,
	})
}
