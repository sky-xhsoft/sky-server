package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UploadChunk 上传单个分片（后端中转模式）
func (s *multipartUploadService) UploadChunk(ctx context.Context, req *UploadChunkRequest, userID uint) error {
	logger.Debug("上传分片",
		zap.Uint("sessionID", req.SessionID),
		zap.Int("chunkIndex", req.ChunkIndex),
		zap.Int("chunkSize", len(req.ChunkData)))

	// 1. 验证上传会话
	session, err := s.ValidateUploadSession(ctx, req.SessionID, userID)
	if err != nil {
		return err
	}

	// 2. 验证分片索引
	if err := s.ValidateChunkIndex(req.ChunkIndex, session); err != nil {
		return err
	}

	// 5. 检查分片是否已上传
	var chunkRecord entity.CloudChunkRecord
	err = s.db.WithContext(ctx).
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
			SessionID:   session.ID,
			ChunkIndex:  req.ChunkIndex,
			ChunkSize:   len(req.ChunkData),
			ChunkMD5:    req.ChunkMD5,
			ChunkPath:   chunkPath,
			Uploaded:    true,
			UploadTime:  &now,
			RetryCount:  0,
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

// ValidateUploadSession 验证上传会话
func (s *multipartUploadService) ValidateUploadSession(ctx context.Context, sessionID uint, userID uint) (*entity.CloudUploadSession, error) {
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	if session.Status != "uploading" && session.Status != "paused" {
		return nil, errors.New(errors.ErrInvalidParam, "上传会话状态无效: "+session.Status)
	}

	if session.ExpireTime.Before(time.Now()) {
		return nil, errors.New(errors.ErrInvalidParam, "上传会话已过期")
	}

	return &session, nil
}

// ValidateChunkIndex 验证分片索引
func (s *multipartUploadService) ValidateChunkIndex(chunkIndex int, session *entity.CloudUploadSession) error {
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return errors.New(errors.ErrInvalidParam, "分片索引无效: "+strconv.Itoa(chunkIndex))
	}
	return nil
}

// GetChunkPresignedURL 获取分片预签名上传URL（前端直传模式）
func (s *multipartUploadService) GetChunkPresignedURL(ctx context.Context, sessionID uint, chunkIndex int, userID uint) (*ChunkPresignedURL, error) {
	logger.Debug("获取分片预签名URL",
		zap.Uint("sessionID", sessionID),
		zap.Int("chunkIndex", chunkIndex))

	// 1. 获取上传会话
	session, err := s.ValidateUploadSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	// 2. 检查会话状态
	if session.Status != "uploading" && session.Status != "paused" {
		return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("上传会话状态无效: %s", session.Status))
	}
	// 3. 检查会话是否过期
	if session.ExpireTime.Before(time.Now()) {
		return nil, errors.New(errors.ErrInvalidParam, "上传会话已过期")
	}
	// 4. 检查分片索引是否有效
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片索引无效: %d", chunkIndex))
	}

	// 5. 获取预签名URL，过期时间设为1小时
	expireSeconds := 3600 // 1小时
	var presignedURL string

	// 如果会话已经初始化了存储端原生分块上传，使用原生分块预签名URL
	if session.StorageUploadID != "" {
		// 使用原生分块上传接口
		presignedURL, err = s.storage.PresignedChunkUploadURL(ctx, session.StoragePath, session.StorageUploadID, chunkIndex, expireSeconds, "application/octet-stream")
		if err != nil {
			logger.Error("获取分块预签名URL失败",
				zap.Uint("sessionID", sessionID),
				zap.Int("chunkIndex", chunkIndex),
				zap.Error(err))
			return nil, err
		}
	} else {
		// 传统模式：每个分片单独上传到临时路径
		chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, chunkIndex)
		presignedURL, err = s.storage.PresignedUploadURL(ctx, chunkPath, expireSeconds, "application/octet-stream")
		if err != nil {
			logger.Error("获取预签名URL失败",
				zap.Uint("sessionID", sessionID),
				zap.Int("chunkIndex", chunkIndex),
				zap.Error(err))
			return nil, err
		}
	}

	// 6. 返回结果
	return &ChunkPresignedURL{
		SessionID:     sessionID,
		ChunkIndex:    chunkIndex,
		PresignedURL:  presignedURL,
		ExpireSeconds: expireSeconds,
	}, nil
}

// MarkChunkUploaded 标记分片已上传（前端直传模式）
// 整个流程放进单个事务 + 行锁，保证并发安全，不会覆盖
func (s *multipartUploadService) MarkChunkUploaded(ctx context.Context, sessionID uint, chunkIndex int, etag string, userID uint) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查询会话并加行锁（SELECT FOR UPDATE），保证并发更新排队有序，每个人都能拿到最新数据
		var lockedSession entity.CloudUploadSession
		if err := tx.WithContext(ctx).
			Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
			First(&lockedSession).
			Clauses(clause.Locking{Strength: "UPDATE"}). // 加行锁，并发更新会阻塞等待，不会覆盖
			Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New(errors.ErrResourceNotFound, "上传会话不存在")
			}
			return errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
		}

		// 2. 验证分片索引
		if chunkIndex < 0 || chunkIndex >= lockedSession.TotalChunks {
			return errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片索引无效: %d", chunkIndex))
		}

		// 3. 构造分片存储路径
		var chunkPath string
		if lockedSession.StorageUploadID == "" {
			chunkPath = fmt.Sprintf("%s/chunk_%d", lockedSession.StoragePath, chunkIndex)
		} else {
			// 原生分块上传模式下，分片直接上传到COS，不需要临时路径
			chunkPath = ""
		}

		// 4. 查询分片是否已存在
		var chunkRecord entity.CloudChunkRecord
		err := tx.WithContext(ctx).
			Where("SESSION_ID = ? AND CHUNK_INDEX = ?", sessionID, chunkIndex).
			First(&chunkRecord).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新记录
			now := time.Now()
			chunkRecord = entity.CloudChunkRecord{
				SessionID:   sessionID,
				ChunkIndex:  chunkIndex,
				ChunkSize:   int(lockedSession.ChunkSize),
				ChunkMD5:    "", // 直传模式下，MD5由前端计算，但我们不需要存储
				ETag:        etag,
				ChunkPath:   chunkPath,
				Uploaded:    true,
				UploadTime:  &now,
				RetryCount:  0,
			}

			if err := tx.WithContext(ctx).Create(&chunkRecord).Error; err != nil {
				if err.Error() == "context canceled" {
					logger.Debug("创建分片记录被取消", zap.Uint("sessionId", sessionID), zap.Int("chunkIndex", chunkIndex))
				} else {
					logger.Error("创建分片记录失败", zap.Uint("sessionId", sessionID), zap.Int("chunkIndex", chunkIndex), zap.Error(err))
				}
				return errors.Wrap(errors.ErrDatabase, "创建分片记录失败", err)
			}
		} else if err == nil && chunkRecord.Uploaded {
			// 已上传，更新ETag（如果有变化）
			if chunkRecord.ETag != etag && etag != "" {
				if err := tx.WithContext(ctx).Model(&chunkRecord).Update("ETAG", etag).Error; err != nil {
					logger.Warn("更新分片ETag失败", zap.Uint("sessionId", sessionID), zap.Int("chunkIndex", chunkIndex))
				}
			}
			// 已上传，直接返回（不重复添加）
		} else {
			// 存在但未上传，更新为已上传
			updates := map[string]interface{}{
				"UPLOADED":    true,
				"UPLOAD_TIME": time.Now(),
				"RETRY_COUNT": gorm.Expr("RETRY_COUNT + 1"),
			}
			if etag != "" {
				updates["ETAG"] = etag
			}
			if err := tx.WithContext(ctx).Model(&chunkRecord).Updates(updates).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "更新分片记录失败", err)
			}
		}

		// 5. 更新会话的已上传分片列表
		// 因为加了行锁，这里肯定拿到的是最新数据
		var uploadedChunks []int
		if lockedSession.UploadedChunks != "" {
			json.Unmarshal([]byte(lockedSession.UploadedChunks), &uploadedChunks)
		}

		// 检查是否已存在
		exists := false
		for _, idx := range uploadedChunks {
			if idx == chunkIndex {
				exists = true
				break
			}
		}

		if (!exists) {
			// 追加分片
			uploadedChunks = append(uploadedChunks, chunkIndex)
			sort.Ints(uploadedChunks)
			uploadedChunksJSON, _ := json.Marshal(uploadedChunks)

			// 更新数据库，在事务中肯定不会覆盖
			if err := tx.Model(&lockedSession).
				Update("UPLOADED_CHUNKS", string(uploadedChunksJSON)).Error; err != nil {
				logger.Warn("更新会话已上传分片列表失败", zap.Uint("sessionId", sessionID), zap.Int("chunkIndex", chunkIndex), zap.Error(err))
				return err // 回滚事务
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	logger.Debug("分片标记为已上传（直传模式）", zap.Uint("sessionId", sessionID), zap.Int("chunkIndex", chunkIndex))
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
