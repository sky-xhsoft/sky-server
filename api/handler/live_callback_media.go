package handler

import (
	"fmt"
	"strings"

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
		FileSize:     event.FileSize,
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

// saveClip 高光切片回调写入切片列表
func (h *LiveCallbackHandler) saveClip(room *entity.LiveRoom, streamID string, item *HighlightItem) {
	if room == nil || room.ID == 0 || item.VideoStoreURL == "" {
		return
	}
	roomID := int64(room.ID)
	var exists int64
	h.db.Model(&entity.LiveCut{}).Where("ROOM_ID = ? AND CLIP_URL = ?", roomID, item.VideoStoreURL).Count(&exists)
	if exists > 0 {
		return
	}

	name := strings.TrimSpace(item.Title)
	if name == "" {
		name = fmt.Sprintf("高光切片_%d", item.BeginTime)
	}
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

// truncateRunes 按字符截断，避免超出列长度
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
