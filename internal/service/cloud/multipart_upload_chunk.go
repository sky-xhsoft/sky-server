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
	// 单独保存查询结果：后面的 err 会被存储调用覆盖，不能再用它判断记录是否存在
	var chunkRecord entity.CloudChunkRecord
	findErr := s.db.WithContext(ctx).
		Where("SESSION_ID = ? AND CHUNK_INDEX = ?", session.ID, req.ChunkIndex).
		First(&chunkRecord).Error
	if findErr != nil && findErr != gorm.ErrRecordNotFound {
		return errors.Wrap(errors.ErrDatabase, "查询分片记录失败", findErr)
	}

	if findErr == nil && chunkRecord.Uploaded {
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

	// 7. 保存分片：COS 等支持原生分块的存储直接写入最终对象的分块（完成时由存储端合并），
	//    否则保存为临时分片文件，完成时由服务端合并
	storageInstance, err := s.getSessionStorage(ctx, session)
	if err != nil {
		logger.Error("获取存储实例失败", zap.Error(err))
		return errors.Wrap(errors.ErrStorage, "获取存储实例失败", err)
	}

	chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, req.ChunkIndex)
	etag := ""
	if writer, ok := storageInstance.(offsetWriter); ok && session.StorageUploadID == "" {
		// 本地存储：分片直接写入组装文件的对应偏移，完成时只需改名，无需再合并
		chunkSize := int64(session.ChunkSize)
		expected := chunkSize
		if rest := session.FileSize - int64(req.ChunkIndex)*chunkSize; rest < expected {
			expected = rest
		}
		if int64(len(req.ChunkData)) != expected {
			return errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片大小不正确：期望 %d，实际 %d", expected, len(req.ChunkData)))
		}
		chunkPath = assembledPartPath(session.StoragePath)
		if err := writer.WriteAt(ctx, chunkPath, int64(req.ChunkIndex)*chunkSize, req.ChunkData); err != nil {
			logger.Error("写入分片失败", zap.Int("chunkIndex", req.ChunkIndex), zap.Error(err))
			return errors.Wrap(errors.ErrStorage, "保存分片失败", err)
		}
	} else if partUploader, ok := storageInstance.(nativePartUploader); ok && session.StorageUploadID != "" {
		chunkPath = ""
		if etag, err = partUploader.UploadPart(ctx, session.StoragePath, session.StorageUploadID, req.ChunkIndex, req.ChunkData); err != nil {
			logger.Error("上传原生分块失败", zap.Int("chunkIndex", req.ChunkIndex), zap.Error(err))
			return errors.Wrap(errors.ErrStorage, "上传分块失败", err)
		}
	} else if _, err := storageInstance.Upload(ctx, chunkPath, &chunkReader{data: req.ChunkData}, "application/octet-stream"); err != nil {
		logger.Error("保存分片失败",
			zap.Int("chunkIndex", req.ChunkIndex),
			zap.Error(err))
		return errors.Wrap(errors.ErrInternal, "保存分片失败", err)
	}

	// 8. 记录分片并更新会话的已上传列表（会话行锁，并发上传分片时不会互相覆盖）
	if err := s.recordChunkUploaded(ctx, session.ID, req.ChunkIndex, len(req.ChunkData), req.ChunkMD5, chunkPath, etag); err != nil {
		return err
	}

	logger.Debug("分片上传成功",
		zap.Int("chunkIndex", req.ChunkIndex),
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

	// 获取公司存储实例
	storageInstance, err := s.getSessionStorage(ctx, session)
	if err != nil {
		logger.Error("获取存储实例失败", zap.Error(err))
		return nil, errors.Wrap(errors.ErrStorage, "获取存储实例失败", err)
	}

	// 如果会话已经初始化了存储端原生分块上传，使用原生分块预签名URL
	if session.StorageUploadID != "" {
		// 使用原生分块上传接口
		presignedURL, err = storageInstance.PresignedChunkUploadURL(ctx, session.StoragePath, session.StorageUploadID, chunkIndex, expireSeconds, "application/octet-stream")
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
		presignedURL, err = storageInstance.PresignedUploadURL(ctx, chunkPath, expireSeconds, "application/octet-stream")
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
			Clauses(clause.Locking{Strength: "UPDATE"}). // 加行锁（须在查询前），并发更新会阻塞等待，不会覆盖
			Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
			First(&lockedSession).Error; err != nil {
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
				SessionID:  sessionID,
				ChunkIndex: chunkIndex,
				ChunkSize:  int(lockedSession.ChunkSize),
				ChunkMD5:   "", // 直传模式下，MD5由前端计算，但我们不需要存储
				ETag:       etag,
				ChunkPath:  chunkPath,
				Uploaded:   true,
				UploadTime: &now,
				RetryCount: 0,
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

		if !exists {
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

	// 获取会话的用户ID
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).Select("USER_ID", "STORAGE_TYPE", "STORAGE_PATH", "STORAGE_UPLOAD_ID").Where("ID = ?", sessionID).First(&session).Error; err != nil {
		logger.Error("查询会话失败", zap.Uint("sessionID", sessionID), zap.Error(err))
		return
	}

	// 获取公司存储实例
	storageInstance, err := s.getSessionStorage(ctx, &session)
	if err != nil {
		logger.Error("获取存储实例失败", zap.Uint("sessionID", sessionID), zap.Error(err))
		return
	}

	// 0. 原生分块上传：终止存储端未完成的上传，释放已上传的分块（已完成合并时终止无影响）
	if aborter, ok := storageInstance.(interface {
		AbortMultipartUpload(ctx context.Context, path string, uploadID string) error
	}); ok && session.StorageUploadID != "" {
		if err := aborter.AbortMultipartUpload(ctx, session.StoragePath, session.StorageUploadID); err != nil {
			logger.Debug("终止原生分块上传（可能已完成）", zap.Uint("sessionID", sessionID), zap.Error(err))
		}
	}

	// 1. 列出所有分片文件
	objects, err := storageInstance.ListObjects(ctx, storagePath, 0)
	if err != nil {
		logger.Error("列举分片文件失败", zap.Error(err))
		return
	}

	// 2. 删除所有分片文件
	for _, obj := range objects {
		if err := storageInstance.Delete(ctx, obj.Key); err != nil {
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

// offsetWriter 支持按偏移写入的存储（本地存储）
type offsetWriter interface {
	WriteAt(ctx context.Context, path string, offset int64, data []byte) error
}

// assembledPartPath 本地存储分片直接组装的目标文件（位于会话临时目录下，随会话清理）
func assembledPartPath(storagePath string) string {
	return storagePath + "/assembled.part"
}

// nativePartUploader 支持原生分块上传的存储（如腾讯云 COS）
type nativePartUploader interface {
	UploadPart(ctx context.Context, path string, uploadID string, chunkIndex int, data []byte) (string, error)
}

// recordChunkUploaded 记录已上传的分片，并在会话行锁内把分片序号加入会话的已上传列表
func (s *multipartUploadService) recordChunkUploaded(ctx context.Context, sessionID uint, chunkIndex, size int, chunkMD5, chunkPath, etag string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session entity.CloudUploadSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("ID", "UPLOADED_CHUNKS").
			Where("ID = ?", sessionID).
			First(&session).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
		}

		now := time.Now()
		var record entity.CloudChunkRecord
		findErr := tx.Where("SESSION_ID = ? AND CHUNK_INDEX = ?", sessionID, chunkIndex).First(&record).Error
		switch {
		case findErr == gorm.ErrRecordNotFound:
			record = entity.CloudChunkRecord{
				SessionID:  sessionID,
				ChunkIndex: chunkIndex,
				ChunkSize:  size,
				ChunkMD5:   chunkMD5,
				ChunkPath:  chunkPath,
				ETag:       etag,
				Uploaded:   true,
				UploadTime: &now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "创建分片记录失败", err)
			}
		case findErr != nil:
			return errors.Wrap(errors.ErrDatabase, "查询分片记录失败", findErr)
		default:
			if err := tx.Model(&record).Updates(map[string]interface{}{
				"UPLOADED":    true,
				"UPLOAD_TIME": now,
				"CHUNK_PATH":  chunkPath,
				"ETAG":        etag,
				"RETRY_COUNT": gorm.Expr("RETRY_COUNT + 1"),
			}).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "更新分片记录失败", err)
			}
		}

		var uploaded []int
		if session.UploadedChunks != "" {
			_ = json.Unmarshal([]byte(session.UploadedChunks), &uploaded)
		}
		for _, idx := range uploaded {
			if idx == chunkIndex {
				return nil
			}
		}
		uploaded = append(uploaded, chunkIndex)
		sort.Ints(uploaded)
		data, _ := json.Marshal(uploaded)
		if err := tx.Model(&entity.CloudUploadSession{}).Where("ID = ?", sessionID).
			Update("UPLOADED_CHUNKS", string(data)).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "更新会话状态失败", err)
		}
		return nil
	})
}
