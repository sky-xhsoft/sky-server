package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

// LivePushURLHandler 推流地址生成处理器
type LivePushURLHandler struct {
	pushURLService live.PushURLService
	liveService    live.Service
	roomService    live.LiveRoomService
}

// NewLivePushURLHandler 创建推流地址生成处理器
func NewLivePushURLHandler(pushURLService live.PushURLService, liveService live.Service, roomService live.LiveRoomService) *LivePushURLHandler {
	return &LivePushURLHandler{
		pushURLService: pushURLService,
		liveService:    liveService,
		roomService:    roomService,
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

// GeneratePushURLByRoomRequest 根据直播间ID生成推流地址请求
type GeneratePushURLByRoomRequest struct {
	RoomID flexibleID `json:"roomId" binding:"required"` // 直播间ID（数字或字符串均可）
}

// GeneratePushURLByRoom 根据直播间ID生成推流地址
// @Summary 根据直播间ID生成推流地址
// @Description 根据直播间ID获取该直播间的推流地址（推流域名取自当前公司直播配置，鉴权主KEY取自推流域名鉴权配置，过期时间取直播间开始时间）
// @Tags Live
// @Accept json
// @Produce json
// @Param request body GeneratePushURLByRoomRequest true "直播间ID"
// @Success 200 {object} PushURLResponse
// @Router /api/v1/live/push-url/by-room [post]
func (h *LivePushURLHandler) GeneratePushURLByRoom(c *gin.Context) {
	var req GeneratePushURLByRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if h.roomService == nil {
		utils.InternalError(c, "直播间服务不可用")
		return
	}
	if h.liveService == nil {
		utils.InternalError(c, "直播服务不可用")
		return
	}

	// 正确注入公司ID到上下文（与 GetDomainAuthKey HTTP 接口一致），
	// 使 ListDomains 缓存 key 与取推流鉴权主KEY都用当前登录公司，避免 fallback 到无密钥的默认客户端
	ctx := withCompanyID(c)

	// RoomID 字符串转 uint
	roomID, err := strconv.ParseUint(string(req.RoomID), 10, 64)
	if err != nil || roomID == 0 {
		utils.BadRequest(c, "直播间ID无效: "+string(req.RoomID))
		return
	}

	// 查询直播间，取开始时间
	room, err := h.roomService.GetRoom(ctx, uint(roomID))
	if err != nil {
		utils.FailFromError(c, "查询直播间失败: ", err)
		return
	}

	// 取当前公司的推流域名（Type=0）：从 live:domains:list 缓存读取，排除 tlivecloud.com 尾缀的系统域名
	var pushDomain string
	domains, err := h.liveService.ListDomains(ctx, int64Ptr(0))
	if err == nil {
		for _, d := range domains {
			if d.Name != "" && !strings.HasSuffix(strings.ToLower(d.Name), ".tlivecloud.com") {
				pushDomain = d.Name
				break
			}
		}
	}
	if pushDomain == "" {
		utils.InternalError(c, "当前公司未配置推流域名")
		return
	}

	// 取推流域名的鉴权配置主KEY（取不到则地址不带鉴权参数）
	streamKey := ""
	if auth, err := h.liveService.GetDomainAuthKey(ctx, pushDomain, "push"); err == nil && auth != nil {
		streamKey = auth.MasterKey
	}

	// streamName 用直播间ID；appName 固定 live
	streamName := string(req.RoomID)
	appName := "live"

	// expireTime 用直播间结束时间（Unix秒）；已过期则默认7天后，保证推流鉴权 txTime 在有效期内
	var expireTime int64
	if room.EndTime != nil && room.EndTime.After(time.Now()) {
		expireTime = room.EndTime.Unix()
	} else {
		expireTime = time.Now().Unix() + 7*86400
	}

	pushResp, err := h.pushURLService.GeneratePushURL(pushDomain, appName, streamName, streamKey, expireTime)
	if err != nil {
		utils.FailFromError(c, "生成推流地址失败: ", err)
		return
	}

	utils.Success(c, pushResp)
}

func int64Ptr(v int64) *int64 { return &v }

// GeneratePlayURLByRoomRequest 根据直播间ID生成播放地址请求
type GeneratePlayURLByRoomRequest struct {
	RoomID flexibleID `json:"roomId" binding:"required"` // 直播间ID（数字或字符串均可）
}

// GeneratePlayURLByRoom 根据直播间ID生成播放地址
// @Summary 根据直播间ID生成播放地址
// @Description 根据直播间ID从 live:domains:list 缓存取当前公司播放域名（过滤 tlivecloud.com 尾缀），生成 RTMP/FLV/HLS 播放地址
// @Tags Live
// @Accept json
// @Produce json
// @Param request body GeneratePlayURLByRoomRequest true "直播间ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/live/pull-url/by-room [post]
func (h *LivePushURLHandler) GeneratePlayURLByRoom(c *gin.Context) {
	var req GeneratePlayURLByRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if h.liveService == nil {
		utils.InternalError(c, "直播服务不可用")
		return
	}

	// 正确注入公司ID到上下文，使 ListDomains 缓存 key 用当前登录公司
	ctx := withCompanyID(c)

	roomID, err := strconv.ParseUint(string(req.RoomID), 10, 64)
	if err != nil || roomID == 0 {
		utils.BadRequest(c, "直播间ID无效: "+string(req.RoomID))
		return
	}

	// 取当前公司的播放域名（Type=1）：从 live:domains:list 缓存读取，排除 tlivecloud.com 尾缀
	var playDomain string
	domains, err := h.liveService.ListDomains(ctx, int64Ptr(1))
	if err == nil {
		for _, d := range domains {
			if d.Name != "" && !strings.HasSuffix(strings.ToLower(d.Name), ".tlivecloud.com") {
				playDomain = d.Name
				break
			}
		}
	}
	if playDomain == "" {
		utils.InternalError(c, "当前公司未配置播放域名")
		return
	}

	// 取播放域名的鉴权主KEY（取不到则地址不带鉴权参数）
	playKey := ""
	if auth, err := h.liveService.GetDomainAuthKey(ctx, playDomain, "play"); err == nil && auth != nil {
		playKey = auth.MasterKey
	}

	// streamName 用直播间ID；appName 固定 live
	streamName := string(req.RoomID)
	appName := "live"
	var expireTime int64
	if h.roomService != nil {
		if room, err := h.roomService.GetRoom(ctx, uint(roomID)); err == nil {
			if room.EndTime != nil && room.EndTime.After(time.Now()) {
				expireTime = room.EndTime.Unix()
			} else {
				expireTime = time.Now().Unix() + 7*86400
			}
		}
	}

	playURLs, err := h.pushURLService.GeneratePlayURL(playDomain, appName, streamName, playKey, expireTime)
	if err != nil {
		utils.FailFromError(c, "生成播放地址失败: ", err)
		return
	}

	utils.Success(c, playURLs)
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
		utils.FailFromError(c, "生成推流地址失败: ", err)
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
		utils.FailFromError(c, "生成播放地址失败: ", err)
		return
	}

	utils.Success(c, playURLs)
}

// flexibleID 兼容数字与字符串两种 JSON 写法的 ID（直播间接口返回的 id 为数字）
type flexibleID string

// UnmarshalJSON 接受 123 或 "123"
func (f *flexibleID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*f = flexibleID(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*f = flexibleID(num.String())
	return nil
}
