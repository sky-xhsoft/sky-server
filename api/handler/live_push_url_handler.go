package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

// LivePushURLHandler 推流地址生成处理器
type LivePushURLHandler struct {
	pushURLService live.PushURLService
}

// NewLivePushURLHandler 创建推流地址生成处理器
func NewLivePushURLHandler(pushURLService live.PushURLService) *LivePushURLHandler {
	return &LivePushURLHandler{
		pushURLService: pushURLService,
	}
}

// GeneratePushURLRequest 生成推流地址请求
type GeneratePushURLRequest struct {
	DomainName string `json:"domainName" binding:"required"` // 推流域名
	AppName    string `json:"appName" binding:"required"`    // 应用名称
	StreamName string `json:"streamName" binding:"required"` // 流名称
	StreamKey  string `json:"streamKey"`                     // 推流密钥（可选）
	ExpireTime int64  `json:"expireTime"`                    // 过期时间（Unix时间戳，秒，可选）
}

// GeneratePlayURLRequest 生成播放地址请求
type GeneratePlayURLRequest struct {
	PlayDomain string `json:"playDomain" binding:"required"` // 播放域名
	AppName    string `json:"appName" binding:"required"`    // 应用名称
	StreamName string `json:"streamName" binding:"required"` // 流名称
	PlayKey    string `json:"playKey"`                       // 播放密钥（可选）
	ExpireTime int64  `json:"expireTime"`                    // 过期时间（Unix时间戳，秒，可选）
}

// GeneratePushURL 生成推流地址
// @Summary 生成推流地址
// @Description 根据域名、应用名、流名称等参数生成推流地址
// @Tags Live
// @Accept json
// @Produce json
// @Param request body GeneratePushURLRequest true "生成推流地址请求"
// @Success 200 {object} map[string]string
// @Router /api/v1/live/push-url/generate [post]
func (h *LivePushURLHandler) GeneratePushURL(c *gin.Context) {
	var req GeneratePushURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 生成推流地址
	pushURLResponse, err := h.pushURLService.GeneratePushURL(
		req.DomainName,
		req.AppName,
		req.StreamName,
		req.StreamKey,
		req.ExpireTime,
	)
	if err != nil {
		utils.InternalError(c, "生成推流地址失败: "+err.Error())
		return
	}

	utils.Success(c, pushURLResponse)
}

// GeneratePlayURL 生成播放地址
// @Summary 生成播放地址
// @Description 根据播放域名、应用名、流名称等参数生成播放地址（RTMP、FLV、HLS）
// @Tags Live
// @Accept json
// @Produce json
// @Param request body GeneratePlayURLRequest true "生成播放地址请求"
// @Success 200 {object} map[string]string
// @Router /api/v1/live/play-url/generate [post]
func (h *LivePushURLHandler) GeneratePlayURL(c *gin.Context) {
	var req GeneratePlayURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 生成播放地址
	playURLs, err := h.pushURLService.GeneratePlayURL(
		req.PlayDomain,
		req.AppName,
		req.StreamName,
		req.PlayKey,
		req.ExpireTime,
	)
	if err != nil {
		utils.InternalError(c, "生成播放地址失败: "+err.Error())
		return
	}

	utils.Success(c, playURLs)
}
