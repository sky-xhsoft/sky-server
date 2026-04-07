package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// HandleSmartErase 处理智能擦除事件
// @Summary 智能擦除事件回调
// @Description 接收腾讯云智能擦除事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body SmartEraseEvent true "智能擦除事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/smart-erase [post]
func (h *LiveCallbackHandler) HandleSmartErase(c *gin.Context) {
	var event SmartEraseEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析智能擦除事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("智能擦除事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到智能擦除事件",
		zap.String("streamId", event.StreamID),
		zap.String("taskId", event.TaskID),
		zap.String("status", event.Status),
		zap.String("outputUrl", event.OutputURL))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "smart_erase",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存智能擦除事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleSubtitle 处理直播字幕事件
// @Summary 直播字幕事件回调
// @Description 接收腾讯云直播字幕事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body SubtitleEvent true "直播字幕事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/subtitle [post]
func (h *LiveCallbackHandler) HandleSubtitle(c *gin.Context) {
	var event SubtitleEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析直播字幕事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("直播字幕事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到直播字幕事件",
		zap.String("streamId", event.StreamID),
		zap.String("text", event.Text),
		zap.String("language", event.Language))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "subtitle",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存直播字幕事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleSummary 处理直播摘要事件
// @Summary 直播摘要事件回调
// @Description 接收腾讯云直播摘要事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body SummaryEvent true "直播摘要事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/summary [post]
func (h *LiveCallbackHandler) HandleSummary(c *gin.Context) {
	var event SummaryEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析直播摘要事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("直播摘要事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到直播摘要事件",
		zap.String("streamId", event.StreamID),
		zap.String("summary", event.Summary),
		zap.Strings("keywords", event.Keywords))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "summary",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存直播摘要事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleHighlight 处理高光切片事件
// @Summary 高光切片事件回调
// @Description 接收腾讯云高光切片事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body HighlightEvent true "高光切片事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/highlight [post]
func (h *LiveCallbackHandler) HandleHighlight(c *gin.Context) {
	// 读取原始请求体
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("读取高光切片事件请求体失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "failed to read request body"})
		return
	}
	// 恢复请求体，以便后续的 ShouldBindJSON 可以读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var event HighlightEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析高光切片事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if event.Sign != "" && h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("高光切片事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到高光切片事件",
		zap.String("streamId", event.StreamID),
		zap.String("domain", event.Domain),
		zap.Int("itemCount", len(event.Items)))

	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	// 为每个高光切片创建一条记录
	for i, item := range event.Items {
		// 构建单个切片的数据结构（用于保存到 event_data）
		itemData := map[string]interface{}{
			"event_type":        event.EventType,
			"stream_id":         event.StreamID,
			"appid":             event.AppID,
			"domain":            event.Domain,
			"path":              event.Path,
			"begin_time":        item.BeginTime,
			"end_time":          item.EndTime,
			"video_store_url":   item.VideoStoreURL,
			"cov_img_store_url": item.CovImgStoreURL,
			"title":             item.Title,
			"summary":           item.Summary,
			"key_words":         item.KeyWords,
			"clip_url":          item.VideoStoreURL, // 兼容旧字段名
			"start_time":        item.BeginTime,     // 兼容旧字段名
		}

		eventData, _ := json.Marshal(itemData)
		callbackEvent := &entity.LiveCallbackEvent{
			EventType:    "highlight",
			EventTime:    item.BeginTime,
			DomainName:   event.Domain,
			AppName:      event.Path,
			StreamName:   "", // 从 stream_id 中提取或留空
			StreamID:     event.StreamID,
			EventData:    string(eventData),
			ClientIP:     c.ClientIP(),
			Sign:         event.Sign,
			TValue:       event.T,
			SysCompanyID: 1,
			IsActive:     "Y",
			RoomName:     room.RoomName,
		}

		if err := h.db.Create(callbackEvent).Error; err != nil {
			logger.Error("保存高光切片事件失败",
				zap.Error(err),
				zap.Int("itemIndex", i),
				zap.String("title", item.Title))
		} else {
			logger.Info("保存高光切片成功",
				zap.Int("itemIndex", i),
				zap.String("title", item.Title),
				zap.String("videoUrl", item.VideoStoreURL))
		}

		// 将高光切片保存到云盘（如果找到了对应的直播间）
		if room.RoomName != "" {
			// 获取创建者的用户ID（从 CreateBy 字段映射到 User ID）
			var creatorUser entity.SysUser
			if err := h.db.Where("USERNAME = ? AND IS_ACTIVE = ?", room.CreateBy, "Y").First(&creatorUser).Error; err == nil {
				// 创建文件夹："直播间名称/直播切片"
				var roomFolder, highlightFolder entity.CloudItem

				// 检查直播间文件夹是否存在
				if err := h.db.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID IS NULL AND IS_ACTIVE = ?",
					"folder", room.RoomName, creatorUser.ID, "Y").First(&roomFolder).Error; err != nil {
					// 直播间文件夹不存在，创建
					roomFolder = entity.CloudItem{
						BaseModel: entity.BaseModel{
							CreateBy: room.CreateBy,
							UpdateBy: room.CreateBy,
							IsActive: "Y",
						},
						ItemType: "folder",
						Name:     room.RoomName,
						ParentID: nil,
						Path:     "/" + room.RoomName,
						OwnerID:  creatorUser.ID,
					}
					if err := h.db.Create(&roomFolder).Error; err != nil {
						logger.Error("创建直播间文件夹失败", zap.String("roomName", room.RoomName), zap.Error(err))
					}
				}

				// 检查"直播切片"文件夹是否存在
				if err := h.db.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = ?",
					"folder", "直播切片", creatorUser.ID, roomFolder.ID, "Y").First(&highlightFolder).Error; err != nil {
					// 直播切片文件夹不存在，创建
					highlightFolder = entity.CloudItem{
						BaseModel: entity.BaseModel{
							CreateBy: room.CreateBy,
							UpdateBy: room.CreateBy,
							IsActive: "Y",
						},
						ItemType: "folder",
						Name:     "直播切片",
						ParentID: &roomFolder.ID,
						Path:     roomFolder.Path + "/直播切片",
						OwnerID:  creatorUser.ID,
					}
					if err := h.db.Create(&highlightFolder).Error; err != nil {
						logger.Error("创建直播切片文件夹失败", zap.String("roomName", room.RoomName), zap.Error(err))
					}
				}

				// 创建高光切片文件记录
				fileName := fmt.Sprintf("高光_%d_%s.mp4", item.BeginTime, strings.TrimSpace(item.Title))
				// 清理文件名中的无效字符
				fileName = strings.ReplaceAll(fileName, "\\", "")
				fileName = strings.ReplaceAll(fileName, "/", "")
				fileName = strings.ReplaceAll(fileName, ":", "_")
				fileName = strings.ReplaceAll(fileName, "*", "")
				fileName = strings.ReplaceAll(fileName, "?", "")
				fileName = strings.ReplaceAll(fileName, "\"", "")
				fileName = strings.ReplaceAll(fileName, "<", "")
				fileName = strings.ReplaceAll(fileName, ">", "")
				fileName = strings.ReplaceAll(fileName, "|", "")

				ext := ".mp4"
				fileType := "video/mp4"

				// 检查文件是否已存在（防止重复创建）
				var existingFile entity.CloudItem
				fileExists := h.db.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = ?",
					"file", fileName, creatorUser.ID, highlightFolder.ID, "Y").First(&existingFile).Error == nil

				if !fileExists {
					fileItem := entity.CloudItem{
						BaseModel: entity.BaseModel{
							CreateBy: room.CreateBy,
							UpdateBy: room.CreateBy,
							IsActive: "Y",
						},
						ItemType:    "file",
						Name:        fileName,
						ParentID:    &highlightFolder.ID,
						Path:        highlightFolder.Path + "/" + fileName,
						OwnerID:     creatorUser.ID,
						FileSize:    func() *int64 { size := int64(0); return &size }(), // 高光切片大小通常在回调中不提供，设置为0
						FileType:    &fileType,
						FileExt:     &ext,
						AccessURL:   &item.VideoStoreURL,                            // 使用回调中的 VideoStoreURL 作为访问地址
						StorageType: func() *string { str := "oss"; return &str }(), // 标记为腾讯云存储
						StoragePath: &item.VideoStoreURL,                            // 将 VideoStoreURL 作为存储路径
					}

					if err := h.cloudService.CreateItem(context.Background(), &fileItem); err != nil {
						logger.Error("保存高光切片到云盘失败",
							zap.String("roomName", room.RoomName),
							zap.String("fileName", fileName),
							zap.Error(err))
					} else {
						logger.Info("高光切片已保存到云盘",
							zap.String("roomName", room.RoomName),
							zap.String("fileName", fileName),
							zap.String("path", fileItem.Path),
							zap.String("accessURL", *fileItem.AccessURL))
					}
				} else {
					logger.Warn("高光切片已存在，跳过创建",
						zap.String("roomName", room.RoomName),
						zap.String("fileName", fileName))
				}
			} else {
				logger.Warn("未找到直播间创建者信息，无法保存到云盘", zap.String("roomName", room.RoomName))
			}
		}
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandlePushException 处理推流异常事件
// @Summary 推流异常事件回调
// @Description 接收腾讯云推流异常事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body PushExceptionEvent true "推流异常事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/push-exception [post]
func (h *LiveCallbackHandler) HandlePushException(c *gin.Context) {
	var event PushExceptionEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析推流异常事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("推流异常事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Error("收到推流异常事件",
		zap.String("streamId", event.StreamID),
		zap.Int("errorCode", event.ErrorCode),
		zap.String("errorMsg", event.ErrorMsg))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "push_exception",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存推流异常事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleRecordException 处理录制异常事件
// @Summary 录制异常事件回调
// @Description 接收腾讯云录制异常事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body RecordExceptionEvent true "录制异常事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/record-exception [post]
func (h *LiveCallbackHandler) HandleRecordException(c *gin.Context) {
	var event RecordExceptionEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析录制异常事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("录制异常事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Error("收到录制异常事件",
		zap.String("streamId", event.StreamID),
		zap.Int("errorCode", event.ErrorCode),
		zap.String("errorMsg", event.ErrorMsg))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "record_exception",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存录制异常事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandlePullStream 处理拉流转推事件
// @Summary 拉流转推事件回调
// @Description 接收腾讯云拉流转推事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body PullStreamEvent true "拉流转推事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/pull-stream [post]
func (h *LiveCallbackHandler) HandlePullStream(c *gin.Context) {
	var event PullStreamEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析拉流转推事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("拉流转推事件签名验证失败", zap.String("taskId", event.TaskID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到拉流转推事件",
		zap.String("taskId", event.TaskID),
		zap.String("status", event.Status),
		zap.String("sourceUrl", event.SourceURL),
		zap.String("targetUrl", event.TargetURL))

	eventData, _ := json.Marshal(event)
	// 通过流名称获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("STREAM_NAME = ? AND IS_ACTIVE = ?", event.StreamName, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamName", event.StreamName),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamName", event.StreamName))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "pull_stream",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存拉流转推事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}

// HandleMonitor 处理监播事件
// @Summary 监播事件回调
// @Description 接收腾讯云监播事件通知
// @Tags Live-Callback
// @Accept json
// @Produce json
// @Param event body MonitorEvent true "监播事件"
// @Success 200 {object} map[string]interface{} "{"code":0}"
// @Router /api/v1/live/callback/monitor [post]
func (h *LiveCallbackHandler) HandleMonitor(c *gin.Context) {
	var event MonitorEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		logger.Error("解析监播事件失败", zap.Error(err))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	if h.callbackKey != "" && !h.verifySign(event.T, event.Sign) {
		logger.Warn("监播事件签名验证失败", zap.String("streamId", event.StreamID))
		c.JSON(200, gin.H{"code": 1, "msg": "invalid signature"})
		return
	}

	logger.Info("收到监播事件",
		zap.String("streamId", event.StreamID),
		zap.String("alertType", event.AlertType),
		zap.String("alertLevel", event.AlertLevel),
		zap.String("description", event.Description))

	eventData, _ := json.Marshal(event)
	// 通过 stream_id 获取直播间信息
	var room entity.LiveRoom
	if err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
		logger.Info("找到对应的直播间",
			zap.String("streamId", event.StreamID),
			zap.String("roomName", room.RoomName))
	} else {
		logger.Warn("未找到对应的直播间",
			zap.String("streamId", event.StreamID))
	}

	callbackEvent := &entity.LiveCallbackEvent{
		EventType:    "monitor",
		EventTime:    event.EventTime,
		DomainName:   event.PushDomain,
		AppName:      event.AppName,
		StreamName:   event.StreamName,
		StreamID:     event.StreamID,
		EventData:    string(eventData),
		Sign:         event.Sign,
		TValue:       event.T,
		SysCompanyID: 1,
		RoomName:     room.RoomName,
		IsActive:     "Y",
	}

	if err := h.db.Create(callbackEvent).Error; err != nil {
		logger.Error("保存监播事件失败", zap.Error(err))
	}

	c.JSON(200, gin.H{"code": 0})
}
