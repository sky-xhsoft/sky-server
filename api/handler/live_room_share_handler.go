package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 直播间分享：任何人打开 /live-share/{code} 即可观看，无需登录。
// 只播放直播流（不提供回放）；直播结束超过 shareEndedGrace 后链接失效

const (
	shareEndedGrace = time.Hour     // 直播结束后链接继续有效的时间
	sharePlayTTL    = 6 * time.Hour // 分享页播放地址鉴权有效期（页面定时刷新状态时重新生成）

	ShareStatusLive     = "live"     // 直播中
	ShareStatusUpcoming = "upcoming" // 未开始
	ShareStatusEnded    = "ended"    // 已结束（1 小时内）
	ShareStatusInvalid  = "invalid"  // 链接失效：分享关闭、直播间删除、结束超过 1 小时
)

// LiveRoomShareHandler 直播间分享
type LiveRoomShareHandler struct {
	db             *gorm.DB
	liveService    live.Service
	pushURLService live.PushURLService
	now            func() time.Time
}

// NewLiveRoomShareHandler 创建直播间分享处理器
func NewLiveRoomShareHandler(db *gorm.DB, liveService live.Service, pushURLService live.PushURLService) *LiveRoomShareHandler {
	return &LiveRoomShareHandler{db: db, liveService: liveService, pushURLService: pushURLService, now: time.Now}
}

// shareInfo 管理端看到的分享设置
type shareInfo struct {
	Enabled bool   `json:"enabled"`
	Code    string `json:"code"` // 未生成时为空
}

func newShareCode() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// companyRoom 当前公司的有效直播间；失败时已写入响应
func (h *LiveRoomShareHandler) companyRoom(c *gin.Context) (*entity.LiveRoom, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.BadRequest(c, "直播间ID无效")
		return nil, false
	}
	companyID, _ := withCompanyID(c).Value(contextutil.CompanyIDKey).(uint)
	var room entity.LiveRoom
	err = h.db.Where("ID = ? AND SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", id, companyID).First(&room).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.NotFound(c, "直播间不存在")
		return nil, false
	}
	if err != nil {
		utils.FailFromError(c, "查询直播间失败: ", err)
		return nil, false
	}
	return &room, true
}

func toShareInfo(room *entity.LiveRoom) shareInfo {
	info := shareInfo{Enabled: room.ShareEnabled == "Y"}
	if room.ShareCode != nil {
		info.Code = *room.ShareCode
	}
	return info
}

// GetShare 分享设置：GET /live/rooms/:id/share
func (h *LiveRoomShareHandler) GetShare(c *gin.Context) {
	room, ok := h.companyRoom(c)
	if !ok {
		return
	}
	utils.Success(c, toShareInfo(room))
}

// UpdateShare 开启 / 关闭分享，reset 为 true 时重新生成分享码（旧链接失效）：PUT /live/rooms/:id/share
func (h *LiveRoomShareHandler) UpdateShare(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
		Reset   bool `json:"reset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	room, ok := h.companyRoom(c)
	if !ok {
		return
	}
	updates := map[string]interface{}{"SHARE_ENABLED": "N", "UPDATE_BY": c.GetString("username")}
	if req.Enabled {
		updates["SHARE_ENABLED"] = "Y"
	}
	// 开启时还没有分享码，或要求重新生成
	if (req.Enabled && room.ShareCode == nil) || req.Reset {
		code, err := newShareCode()
		if err != nil {
			utils.InternalError(c, "生成分享码失败")
			return
		}
		updates["SHARE_CODE"] = code
	}
	if err := h.db.Model(&entity.LiveRoom{}).Where("ID = ?", room.ID).Updates(updates).Error; err != nil {
		utils.FailFromError(c, "保存分享设置失败: ", err)
		return
	}
	if err := h.db.First(room, room.ID).Error; err != nil {
		utils.FailFromError(c, "查询直播间失败: ", err)
		return
	}
	utils.Success(c, toShareInfo(room))
}

// sharePublicView 分享页看到的内容（不含任何内部 ID、推流信息）
type sharePublicView struct {
	Status      string            `json:"status"`
	RoomName    string            `json:"roomName,omitempty"`
	CompanyName string            `json:"companyName,omitempty"`
	CoverImage  string            `json:"coverImage,omitempty"`
	Description string            `json:"description,omitempty"`
	DisplayMode string            `json:"displayMode,omitempty"`
	StartTime   *time.Time        `json:"startTime,omitempty"` // 计划开始时间（未开始时倒计时）
	EndedAt     *time.Time        `json:"endedAt,omitempty"`   // 实际结束时间（已结束）
	PlayURLs    map[string]string `json:"playUrls,omitempty"`  // 直播中：flv / hls
	ServerTime  time.Time         `json:"serverTime"`          // 服务器时间，页面按它算倒计时
}

// shareStatus 分享页状态：直播中；已结束且结束不到 1 小时；结束超过 1 小时或已归档为失效；其余为未开始
func shareStatus(room *entity.LiveRoom, endedAt *time.Time, now time.Time) string {
	switch room.Status {
	case "live":
		return ShareStatusLive
	case "ended":
		if endedAt != nil && now.Sub(*endedAt) <= shareEndedGrace {
			return ShareStatusEnded
		}
		return ShareStatusInvalid
	case "archived":
		return ShareStatusInvalid
	}
	return ShareStatusUpcoming
}

// roomEndedAt 实际结束时间：最后一次推流结束（断流回调），没有推流记录时取状态变为已结束的时间
func (h *LiveRoomShareHandler) roomEndedAt(room *entity.LiveRoom) *time.Time {
	var last entity.BillLiveSession
	err := h.db.Where("ROOM_ID = ? AND END_TIME IS NOT NULL", room.ID).Order("END_TIME DESC").First(&last).Error
	if err == nil && last.EndTime != nil && !last.EndTime.IsZero() {
		return last.EndTime
	}
	t := room.UpdateTime
	return &t
}

// GetPublicShare 分享页（无需登录）：GET /live/share/:code
func (h *LiveRoomShareHandler) GetPublicShare(c *gin.Context) {
	now := h.now()
	invalid := sharePublicView{Status: ShareStatusInvalid, ServerTime: now}
	code := strings.TrimSpace(c.Param("code"))
	if len(code) < 16 || len(code) > 64 {
		utils.Success(c, invalid)
		return
	}
	var room entity.LiveRoom
	if err := h.db.Where("SHARE_CODE = ? AND SHARE_ENABLED = 'Y' AND IS_ACTIVE = 'Y'", code).First(&room).Error; err != nil {
		utils.Success(c, invalid)
		return
	}
	var endedAt *time.Time
	if room.Status == "ended" {
		endedAt = h.roomEndedAt(&room)
	}
	status := shareStatus(&room, endedAt, now)
	if status == ShareStatusInvalid {
		utils.Success(c, invalid)
		return
	}

	view := sharePublicView{
		Status:      status,
		RoomName:    room.RoomName,
		CoverImage:  room.CoverImage,
		Description: room.Description,
		DisplayMode: room.DisplayMode,
		StartTime:   room.StartTime,
		ServerTime:  now,
	}
	var company entity.SysCompany
	if h.db.Select("NAME").Where("ID = ?", room.SysCompanyID).First(&company).Error == nil {
		view.CompanyName = company.Name
	}
	if status == ShareStatusEnded {
		view.EndedAt = endedAt
	}
	if status == ShareStatusLive {
		urls, err := h.playURLs(&room, now)
		if err != nil {
			logger.Warn("分享页生成播放地址失败", zap.Uint("roomId", room.ID), zap.Error(err))
		} else {
			view.PlayURLs = map[string]string{"flv": urls["flv"], "hls": urls["hls"]}
		}
	}
	utils.Success(c, view)
}

// playURLs 直播间播放地址：用直播间所属公司的播放域名与播放鉴权
func (h *LiveRoomShareHandler) playURLs(room *entity.LiveRoom, now time.Time) (map[string]string, error) {
	if h.liveService == nil || h.pushURLService == nil {
		return nil, errors.New("直播服务不可用")
	}
	ctx := context.WithValue(context.Background(), contextutil.CompanyIDKey, room.SysCompanyID)
	domains, err := h.liveService.ListDomains(ctx, int64Ptr(1))
	if err != nil {
		return nil, err
	}
	playDomain := pickRoomDomain(domains)
	if playDomain == "" {
		return nil, errors.New("未配置播放域名")
	}
	playKey := ""
	if auth, err := h.liveService.GetDomainAuthKey(ctx, playDomain, "play"); err == nil && auth != nil {
		playKey = auth.MasterKey
	}
	return h.pushURLService.GeneratePlayURL(playDomain, "live", strconv.FormatUint(uint64(room.ID), 10), playKey, now.Add(sharePlayTTL).Unix())
}
