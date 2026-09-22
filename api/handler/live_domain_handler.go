package handler

import (
	"fmt"
	"strconv"
	"strings"

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
		// 腾讯云业务错误: 域名已存在 -> 明确提示; 其他继续走 500
		if strings.Contains(err.Error(), "DomainAlreadyExist") {
			utils.BadRequest(c, fmt.Sprintf("域名 %s 已存在，请勿重复添加", req.DomainName))
			return
		}
		// 返回腾讯云SDK的详细错误信息，便于前端定位
		utils.InternalError(c, fmt.Sprintf("域名添加失败: %s", err.Error()))
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

	// 解析验证状态（腾讯云文档 80524：Status >= 0 为已验证归属，< 0 为未验证归属权）
	var statusText string
	if result.Status >= 0 {
		statusText = "验证成功"
	} else {
		statusText = "未验证"
	}

	utils.Success(c, gin.H{
		"status":     result.Status,
		"statusText": statusText,
		"content":    result.Content,
		"mainDomain": result.MainDomain,
	})
}

// GetDomainAuthKey 查询域名鉴权配置（推流/播放）
// @Summary 查询域名鉴权配置
// @Description 按域名类型查询推流或播放鉴权 Key
// @Tags Live-Domain
// @Accept json
// @Produce json
// @Param domainName path string true "域名"
// @Param keyType query string false "鉴权类型：push(默认)-推流, play-播放"
// @Success 200 {object} utils.Response{data=live.AuthKeyInfo}
// @Router /api/v1/live/domains/{domainName}/auth [get]
func (h *LiveDomainHandler) GetDomainAuthKey(c *gin.Context) {
	domainName := c.Param("domainName")
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}

	keyType := c.DefaultQuery("keyType", "push")

	ctx := withCompanyID(c)
	info, err := h.liveService.GetDomainAuthKey(ctx, domainName, keyType)
	if err != nil {
		utils.InternalError(c, "获取鉴权配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"domainName": info.DomainName,
		"keyType":    keyType,
		"enable":     info.Enable,
		"masterKey":  info.MasterKey,
		"backupKey":  info.BackupKey,
		"authDelta":  info.AuthDelta,
	})
}
