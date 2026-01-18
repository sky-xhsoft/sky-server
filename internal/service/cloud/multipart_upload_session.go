package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitUpload 初始化上传会话
func (s *multipartUploadService) InitUpload(ctx context.Context, req *InitUploadRequest, userID uint) (*UploadSessionInfo, error) {
	logger.Info("初始化分片上传",
		zap.String("fileName", req.FileName),
		zap.Int64("fileSize", req.FileSize),
		zap.String("fileMD5", req.FileMD5),
		zap.Uint("userID", userID))

	// 1. 检查配额
	if err := s.cloudService.CheckQuota(ctx, userID, req.FileSize); err != nil {
		return nil, err
	}

	// 2. 检查是否存在未完成的会话（断点续传）
	var existingSession entity.CloudUploadSession
	err := s.db.WithContext(ctx).
		Where("FILE_ID = ? AND USER_ID = ? AND STATUS IN (?, ?) AND IS_ACTIVE = ?",
			req.FileMD5, userID, "uploading", "paused", "Y").
		First(&existingSession).Error

	if err == nil {
		// 存在未完成的会话,返回已上传的分片信息
		logger.Info("找到现有上传会话,支持断点续传",
			zap.Uint("sessionID", existingSession.ID),
			zap.String("status", existingSession.Status))

		// 延长会话过期时间(重要:避免用户继续上传时会话已过期)
		newExpireTime := time.Now().Add(time.Duration(s.sessionExpireHours) * time.Hour)
		if err := s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
			Where("ID = ?", existingSession.ID).
			Updates(map[string]interface{}{
				"EXPIRE_TIME": newExpireTime,
				"STATUS":      "uploading", // 确保状态为 uploading
				"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
				"UPDATE_TIME": time.Now(),
			}).Error; err != nil {
			logger.Error("更新会话过期时间失败", zap.Error(err))
			// 不返回错误,继续返回会话信息
		}

		var uploadedChunks []int
		if existingSession.UploadedChunks != "" {
			json.Unmarshal([]byte(existingSession.UploadedChunks), &uploadedChunks)
		}

		return &UploadSessionInfo{
			SessionID:      existingSession.ID,
			FileID:         existingSession.FileID,
			FileName:       existingSession.FileName,
			FileSize:       existingSession.FileSize,
			ChunkSize:      existingSession.ChunkSize,
			TotalChunks:    existingSession.TotalChunks,
			UploadedChunks: uploadedChunks,
			Status:         "uploading", // 返回最新状态
			ExpireTime:     newExpireTime.Format(time.RFC3339),
		}, nil
	}
	// 3. 设置默认分片大小（使用配置值）
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = s.defaultChunkSize
	}

	// 4. 计算总分片数
	totalChunks := int(math.Ceil(float64(req.FileSize) / float64(chunkSize)))

	// 5. 设置存储类型
	storageType := req.StorageType
	if storageType == "" {
		storageType = "local"
	}

	// 6. 创建新的上传会话（使用配置的过期时间）
	expireTime := time.Now().Add(time.Duration(s.sessionExpireHours) * time.Hour)

	session := &entity.CloudUploadSession{
		BaseModel: entity.BaseModel{
			CreateBy: s.getUsernameByID(ctx, userID),
			UpdateBy: s.getUsernameByID(ctx, userID),
			IsActive: "Y",
		},
		FileID:         req.FileMD5,
		UserID:         userID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		FileType:       req.FileType,
		FolderID:       req.FolderID,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: "[]",
		Status:         "uploading",
		StorageType:    storageType,
		StoragePath:    fmt.Sprintf("cloud/temp/%d/%s", userID, req.FileMD5),
		ExpireTime:     &expireTime,
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建上传会话失败", err)
	}

	logger.Info("创建上传会话成功",
		zap.Uint("sessionID", session.ID),
		zap.Int("totalChunks", totalChunks))

	return &UploadSessionInfo{
		SessionID:      session.ID,
		FileID:         session.FileID,
		FileName:       session.FileName,
		FileSize:       session.FileSize,
		ChunkSize:      session.ChunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: []int{},
		Status:         "uploading",
		ExpireTime:     expireTime.Format(time.RFC3339),
	}, nil
}

// GetUploadStatus 获取上传状态
func (s *multipartUploadService) GetUploadStatus(ctx context.Context, sessionID uint, userID uint) (*UploadStatus, error) {
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	var uploadedChunks []int
	if session.UploadedChunks != "" {
		json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)
	}

	progress := float64(len(uploadedChunks)) / float64(session.TotalChunks)

	return &UploadStatus{
		SessionID:      session.ID,
		FileID:         session.FileID,
		FileName:       session.FileName,
		FileSize:       session.FileSize,
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedChunks,
		Status:         session.Status,
		Progress:       progress,
		ExpireTime:     session.ExpireTime.Format(time.RFC3339),
	}, nil
}

// ResumeUpload 恢复上传（断点续传）
func (s *multipartUploadService) ResumeUpload(ctx context.Context, fileMD5 string, userID uint) (*UploadSessionInfo, error) {
	logger.Info("恢复上传（断点续传）", zap.String("fileMD5", fileMD5), zap.Uint("userID", userID))

	// 查找未完成的会话
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("FILE_ID = ? AND USER_ID = ? AND STATUS IN (?, ?) AND IS_ACTIVE = ?",
			fileMD5, userID, "uploading", "paused", "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "未找到可恢复的上传会话")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	// 检查会话是否过期
	if session.ExpireTime.Before(time.Now()) {
		return nil, errors.New(errors.ErrInvalidParam, "上传会话已过期")
	}

	// 获取已上传的分片列表
	var uploadedChunks []int
	if session.UploadedChunks != "" {
		json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)
	}

	// 更新会话状态为 uploading 并延长过期时间
	newExpireTime := time.Now().Add(time.Duration(s.sessionExpireHours) * time.Hour)
	s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", session.ID).
		Updates(map[string]interface{}{
			"STATUS":      "uploading",
			"EXPIRE_TIME": newExpireTime,
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		})

	return &UploadSessionInfo{
		SessionID:      session.ID,
		FileID:         session.FileID,
		FileName:       session.FileName,
		FileSize:       session.FileSize,
		ChunkSize:      session.ChunkSize,
		TotalChunks:    session.TotalChunks,
		UploadedChunks: uploadedChunks,
		Status:         "uploading",
		ExpireTime:     newExpireTime.Format(time.RFC3339),
	}, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *multipartUploadService) CleanupExpiredSessions(ctx context.Context) error {
	logger.Info("开始清理过期会话")

	// 查询过期的会话
	var sessions []entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("EXPIRE_TIME < ? AND STATUS IN (?, ?, ?) AND IS_ACTIVE = ?",
			time.Now(), "uploading", "paused", "failed", "Y").
		Find(&sessions).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询过期会话失败", err)
	}

	logger.Info("找到过期会话", zap.Int("count", len(sessions)))

	// 清理每个过期会话
	for _, session := range sessions {
		// 标记为已删除
		s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
			Where("ID = ?", session.ID).
			Update("IS_ACTIVE", "N")

		// 异步清理临时文件
		go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)
	}

	return nil
}
