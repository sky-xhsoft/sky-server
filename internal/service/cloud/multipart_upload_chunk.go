package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadChunk 上传单个分片
func (s *multipartUploadService) UploadChunk(ctx context.Context, req *UploadChunkRequest, userID uint) error {
	logger.Debug("上传分片",
		zap.Uint("sessionID", req.SessionID),
		zap.Int("chunkIndex", req.ChunkIndex),
		zap.Int("chunkSize", len(req.ChunkData)))

	// 1. 获取上传会话
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", req.SessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	// 2. 检查会话状态
	if session.Status != "uploading" && session.Status != "paused" {
		return errors.New(errors.ErrInvalidParam, fmt.Sprintf("上传会话状态无效: %s", session.Status))
	}

	// 3. 检查会话是否过期
	if session.ExpireTime.Before(time.Now()) {
		return errors.New(errors.ErrInvalidParam, "上传会话已过期")
	}

	// 4. 检查分片索引是否有效
	if req.ChunkIndex < 0 || req.ChunkIndex >= session.TotalChunks {
		return errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片索引无效: %d", req.ChunkIndex))
	}

	// 5. 检查分片是否已上传
	var chunkRecord entity.CloudChunkRecord
	err := s.db.WithContext(ctx).
		Where("SESSION_ID = ? AND CHUNK_INDEX = ?", session.ID, req.ChunkIndex).
		First(&chunkRecord).Error

	if err == nil && chunkRecord.Uploaded {
		logger.Debug("分片已上传，跳过", zap.Int("chunkIndex", req.ChunkIndex))
		return nil // 分片已上传，跳过
	}

	// 6. 验证分片MD5
	actualMD5 := calculateMD5(req.ChunkData)
	if actualMD5 != req.ChunkMD5 {
		logger.Error("分片MD5校验失败",
			zap.Int("chunkIndex", req.ChunkIndex),
			zap.String("expected", req.ChunkMD5),
			zap.String("actual", actualMD5))
		return errors.New(errors.ErrInvalidParam, "分片MD5校验失败")
	}

	// 7. 保存分片到临时目录
	chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, req.ChunkIndex)

	// 创建 Reader
	chunkReader := &chunkReader{data: req.ChunkData}

	if _, err := s.storage.Upload(ctx, chunkPath, chunkReader, "application/octet-stream"); err != nil {
		logger.Error("保存分片失败",
			zap.Int("chunkIndex", req.ChunkIndex),
			zap.Error(err))
		return errors.Wrap(errors.ErrInternal, "保存分片失败", err)
	}

	// 8. 更新或创建分片记录
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		chunkRecord = entity.CloudChunkRecord{
			SessionID:  session.ID,
			ChunkIndex: req.ChunkIndex,
			ChunkSize:  len(req.ChunkData),
			ChunkMD5:   req.ChunkMD5,
			ChunkPath:  chunkPath,
			Uploaded:   true,
			UploadTime: &now,
			RetryCount: 0,
		}

		if err := s.db.WithContext(ctx).Create(&chunkRecord).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "创建分片记录失败", err)
		}
	} else {
		// 更新现有记录
		if err := s.db.WithContext(ctx).Model(&chunkRecord).Updates(map[string]interface{}{
			"UPLOADED":    true,
			"UPLOAD_TIME": now,
			"RETRY_COUNT": gorm.Expr("RETRY_COUNT + 1"),
		}).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "更新分片记录失败", err)
		}
	}

	// 9. 更新上传会话的已上传分片列表
	var uploadedChunks []int
	if session.UploadedChunks != "" {
		json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)
	}

	// 检查是否已存在
	exists := false
	for _, idx := range uploadedChunks {
		if idx == req.ChunkIndex {
			exists = true
			break
		}
	}

	if !exists {
		uploadedChunks = append(uploadedChunks, req.ChunkIndex)
		sort.Ints(uploadedChunks)

		uploadedChunksJSON, _ := json.Marshal(uploadedChunks)
		if err := s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
			Where("ID = ?", session.ID).
			Update("UPLOADED_CHUNKS", string(uploadedChunksJSON)).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "更新会话状态失败", err)
		}
	}

	logger.Debug("分片上传成功",
		zap.Int("chunkIndex", req.ChunkIndex),
		zap.Int("uploaded", len(uploadedChunks)),
		zap.Int("total", session.TotalChunks))

	return nil
}

// cleanupChunks 清理临时分片文件
func (s *multipartUploadService) cleanupChunks(ctx context.Context, sessionID uint, storagePath string) {
	logger.Info("清理临时分片文件", zap.Uint("sessionID", sessionID), zap.String("path", storagePath))

	// 1. 列出所有分片文件
	objects, err := s.storage.ListObjects(ctx, storagePath, 0)
	if err != nil {
		logger.Error("列举分片文件失败", zap.Error(err))
		return
	}

	// 2. 删除所有分片文件
	for _, obj := range objects {
		if err := s.storage.Delete(ctx, obj.Key); err != nil {
			logger.Error("删除分片文件失败", zap.String("key", obj.Key), zap.Error(err))
		}
	}

	// 3. 删除分片记录
	if err := s.db.WithContext(ctx).
		Where("SESSION_ID = ?", sessionID).
		Delete(&entity.CloudChunkRecord{}).Error; err != nil {
		logger.Error("删除分片记录失败", zap.Error(err))
	}

	logger.Info("临时文件清理完成", zap.Uint("sessionID", sessionID))
}

// chunkReader 分片数据读取器
type chunkReader struct {
	data []byte
	pos  int
}

func (r *chunkReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
