package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// 录制、切片列表（live_recode / live_cut）的数据由腾讯云回调写入。
// 仅处理能找到直播间的回调：公司取直播间所在公司，创建人取直播间创建人；
// 同一直播间的同一文件地址只记录一次（腾讯云可能重试回调）。

const mediaStatusReady = "ready"

// saveRecording 录制文件回调写入录制列表
func (h *LiveCallbackHandler) saveRecording(room *entity.LiveRoom, event *RecordingFileEvent, eventTime int64) {
	if room == nil || room.ID == 0 || event.VideoURL == "" {
		return
	}
	roomID := int64(room.ID)
	var exists int64
	h.db.Model(&entity.LiveRecode{}).Where("ROOM_ID = ? AND RECORD_URL = ?", roomID, event.VideoURL).Count(&exists)
	if exists > 0 {
		return
	}

	recordType := strings.ToLower(strings.TrimSpace(event.FileFormat))
	if recordType == "" {
		recordType = "flv"
	}
	endTime := event.EndTime
	if endTime == 0 {
		endTime = eventTime
	}
	duration := int(event.Duration)
	if duration == 0 && event.StartTime > 0 && endTime > event.StartTime {
		duration = int(endTime - event.StartTime)
	}

	record := &entity.LiveRecode{
		SysCompanyID: int64(room.SysCompanyID),
		RoomID:       &roomID,
		StreamID:     event.StreamID,
		RecordURL:    event.VideoURL,
		RecordType:   recordType,
		FileSize:     fileSizeOr(event.FileSize, event.VideoURL),
		Duration:     duration,
		StartTime:    event.StartTime,
		EndTime:      endTime,
		Status:       mediaStatusReady,
		CreateBy:     room.CreateBy,
		UpdateBy:     room.CreateBy,
		IsActive:     "Y",
	}
	if err := h.db.Create(record).Error; err != nil {
		logger.Error("保存录制记录失败", zap.Uint("roomId", room.ID), zap.String("url", event.VideoURL), zap.Error(err))
	}
}

// saveClip 高光切片回调写入切片列表；size 为切片文件大小（回调不提供，由调用方从 COS 查询）
func (h *LiveCallbackHandler) saveClip(room *entity.LiveRoom, streamID string, item *HighlightItem, size int64) {
	if room == nil || room.ID == 0 || item.VideoStoreURL == "" {
		return
	}
	roomID := int64(room.ID)
	var exists int64
	h.db.Model(&entity.LiveCut{}).Where("ROOM_ID = ? AND CLIP_URL = ?", roomID, item.VideoStoreURL).Count(&exists)
	if exists > 0 {
		return
	}

	name := clipName(item)
	duration := 0
	if item.EndTime > item.BeginTime {
		duration = int(item.EndTime - item.BeginTime)
	}

	clip := &entity.LiveCut{
		SysCompanyID: int64(room.SysCompanyID),
		RoomID:       &roomID,
		StreamID:     streamID,
		ClipName:     truncateRunes(name, 255),
		ClipURL:      item.VideoStoreURL,
		StartTime:    item.BeginTime,
		EndTime:      item.EndTime,
		Duration:     duration,
		FileSize:     size,
		Thumbnail:    item.CovImgStoreURL,
		Status:       mediaStatusReady,
		Description:  truncateRunes(item.Summary, 500),
		CreateBy:     room.CreateBy,
		UpdateBy:     room.CreateBy,
		IsActive:     "Y",
	}
	if err := h.db.Create(clip).Error; err != nil {
		logger.Error("保存切片记录失败", zap.Uint("roomId", room.ID), zap.String("url", item.VideoStoreURL), zap.Error(err))
	}
}

// clipName 切片名称：回调标题为空（多数情况）时用「高光切片_开始时间戳」，切片列表与云盘文件名一致
func clipName(item *HighlightItem) string {
	if name := strings.TrimSpace(item.Title); name != "" {
		return name
	}
	return fmt.Sprintf("高光切片_%d", item.BeginTime)
}

// clipFileName 云盘文件名：切片名称去掉文件名中不允许的字符，加 .mp4
func clipFileName(item *HighlightItem) string {
	name := strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', '*', '?', '"', '<', '>', '|':
			return -1
		case ':':
			return '_'
		}
		return r
	}, clipName(item))
	if strings.TrimSpace(name) == "" {
		name = fmt.Sprintf("高光切片_%d", item.BeginTime)
	}
	return name + ".mp4"
}

// mediaURLAllowed 允许查询大小的地址（测试中替换为本地服务）
var mediaURLAllowed = isTencentMediaURL

// mediaHeadClient 查询 COS 文件大小：超时短，不跟随跳转
var mediaHeadClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// mediaFileSize 用 HEAD 请求从腾讯云 COS 取文件大小（字节）。
// 高光切片回调不带文件大小；只查询腾讯云域名的地址，失败返回 0
func mediaFileSize(rawURL string) int64 {
	if !mediaURLAllowed(rawURL) {
		return 0
	}
	resp, err := mediaHeadClient.Head(strings.TrimSpace(rawURL))
	if err != nil {
		logger.Warn("查询媒体文件大小失败", zap.String("url", rawURL), zap.Error(err))
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("查询媒体文件大小失败", zap.String("url", rawURL), zap.Int("status", resp.StatusCode))
		return 0
	}
	if resp.ContentLength > 0 {
		return resp.ContentLength
	}
	n, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	return n
}

// fileSizeOr 回调带了文件大小则直接用，否则从 COS 查询
func fileSizeOr(size int64, rawURL string) int64 {
	if size > 0 {
		return size
	}
	return mediaFileSize(rawURL)
}

// truncateRunes 按字符截断，避免超出列长度
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
