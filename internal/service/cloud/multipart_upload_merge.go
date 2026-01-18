package cloud

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"crypto/md5"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompleteUpload 完成上传（合并分片）
// 此方法会根据文件大小和分片数量自动选择最优的合并策略：
// - 小文件或少量分片：使用临时文件合并（稳定可靠）
// - 大文件且多分片：使用流式合并（高性能、低内存）
func (s *multipartUploadService) CompleteUpload(ctx context.Context, sessionID uint, userID uint) (*entity.CloudItem, error) {
	logger.Info("完成分片上传", zap.Uint("sessionID", sessionID), zap.Uint("userID", userID))

	// 1. 获取上传会话
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	// 2. 检查所有分片是否已上传
	var uploadedChunks []int
	if session.UploadedChunks != "" {
		json.Unmarshal([]byte(session.UploadedChunks), &uploadedChunks)
	}

	if len(uploadedChunks) != session.TotalChunks {
		return nil, errors.New(errors.ErrInvalidParam,
			fmt.Sprintf("分片未完全上传：已上传 %d/%d", len(uploadedChunks), session.TotalChunks))
	}

	// 2.5 自动选择合并策略：大文件且多分片使用流式合并
	const (
		streamingChunkThreshold    = 10               // 分片数量阈值
		streamingFileSizeThreshold = 50 * 1024 * 1024 // 50MB 文件大小阈值
	)

	// 判断是否使用流式合并优化
	if session.TotalChunks > streamingChunkThreshold && session.FileSize > streamingFileSizeThreshold {
		logger.Info("使用流式合并优化",
			zap.Int("totalChunks", session.TotalChunks),
			zap.Int64("fileSize", session.FileSize))
		return s.completeUploadOptimized(ctx, &session, userID)
	}

	// 3. 检查是否已存在相同MD5的文件（秒传）
	var existingFile entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("MD5 = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", session.FileID, userID, "Y", "file").
		First(&existingFile).Error

	if err == nil {
		// 文件已存在，直接创建新记录（秒传）
		logger.Info("检测到相同文件，使用秒传", zap.String("md5", session.FileID))

		storageType := session.StorageType
		storagePath := existingFile.StoragePath
		fileSize := session.FileSize
		fileType := session.FileType
		fileMD5 := session.FileID
		accessURL := existingFile.AccessURL
		fileExt := filepath.Ext(session.FileName)

		newFile := &entity.CloudItem{
			BaseModel: entity.BaseModel{
				CreateBy: s.getUsernameByID(ctx, userID),
				UpdateBy: s.getUsernameByID(ctx, userID),
				IsActive: "Y",
			},
			ItemType:    "file",
			Name:        session.FileName,
			ParentID:    session.FolderID,
			OwnerID:     userID,
			StorageType: &storageType,
			StoragePath: storagePath,
			FileSize:    &fileSize,
			FileType:    &fileType,
			FileExt:     &fileExt,
			MD5:         &fileMD5,
			AccessURL:   accessURL,
		}

		if err := s.db.WithContext(ctx).Create(newFile).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败（秒传）", err)
		}

		// 更新会话状态
		s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
			Where("ID = ?", session.ID).
			Update("STATUS", "completed")

		// 异步清理临时文件
		go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

		// 更新配额（只增加文件计数，不增加空间使用，因为文件已存在）
		s.cloudService.UpdateQuota(ctx, userID, 0, 1, 0)

		return newFile, nil
	}

	// 4. 合并分片到最终文件
	finalPath := fmt.Sprintf("cloud/%d/%s/%s%s",
		userID,
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		filepath.Ext(session.FileName))

	// 使用临时文件合并
	tempMergedFile := filepath.Join(os.TempDir(), fmt.Sprintf("merge_%d.tmp", sessionID))
	mergedFile, err := os.Create(tempMergedFile)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建合并文件失败", err)
	}
	defer func() {
		mergedFile.Close()
		os.Remove(tempMergedFile)
	}()

	// 按顺序合并所有分片
	logger.Info("开始合并分片", zap.Int("totalChunks", session.TotalChunks))

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, i)

		// 读取分片
		chunkReader, err := s.storage.Download(ctx, chunkPath)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("读取分片 %d 失败", i), err)
		}

		// 写入合并文件
		written, err := io.Copy(mergedFile, chunkReader)
		chunkReader.Close()

		if err != nil {
			return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("合并分片 %d 失败", i), err)
		}

		logger.Debug("合并分片", zap.Int("chunkIndex", i), zap.Int64("written", written))
	}

	// 5. 验证文件完整性（MD5）
	mergedFile.Seek(0, io.SeekStart)
	actualMD5, err := calculateFileMD5(mergedFile)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "计算文件MD5失败", err)
	}

	if actualMD5 != session.FileID {
		logger.Error("文件MD5校验失败",
			zap.String("expected", session.FileID),
			zap.String("actual", actualMD5))
		return nil, errors.New(errors.ErrInvalidParam, "文件MD5校验失败")
	}

	logger.Info("文件MD5校验成功", zap.String("md5", actualMD5))

	// 6. 上传最终文件到存储
	mergedFile.Seek(0, io.SeekStart)
	accessURL, err := s.storage.Upload(ctx, finalPath, mergedFile, session.FileType)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "上传最终文件失败", err)
	}

	storageType := session.StorageType
	fileSize := session.FileSize
	fileType := session.FileType
	fileMD5 := session.FileID
	fileExt := filepath.Ext(session.FileName)
	// 7. 创建文件记录
	file := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy: s.getUsernameByID(ctx, userID),
			UpdateBy: s.getUsernameByID(ctx, userID),
			IsActive: "Y",
		},
		ItemType:    "file",
		Name:        session.FileName,
		ParentID:    session.FolderID,
		OwnerID:     userID,
		StoragePath: &finalPath,
		FileSize:    &fileSize,
		FileType:    &fileType,
		FileExt:     &fileExt,
		MD5:         &fileMD5,
		StorageType: &storageType,
		AccessURL:   &accessURL,
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		// 删除已上传的文件
		s.storage.Delete(ctx, finalPath)
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	// 8. 更新配额：文件数量+1，空间使用+文件大小
	s.cloudService.UpdateQuota(ctx, userID, session.FileSize, 1, 0)

	// 9. 更新会话状态为已完成
	s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", session.ID).
		Update("STATUS", "completed")

	// 10. 异步清理临时文件
	go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

	logger.Info("文件上传完成",
		zap.Uint("fileID", file.ID),
		zap.String("fileName", file.Name),
		zap.Int64("fileSize", *file.FileSize))

	return file, nil
}

// completeUploadOptimized 完成上传（流式合并优化版本）
// 优化说明：
// 1. 使用 io.Pipe() 创建管道，避免创建本地临时文件
// 2. 使用 goroutine 并发处理：一边读取分片、计算MD5，一边上传到存储
// 3. 减少磁盘 I/O，降低内存占用
// 4. 适用于大文件和多分片场景
func (s *multipartUploadService) completeUploadOptimized(ctx context.Context, session *entity.CloudUploadSession, userID uint) (*entity.CloudItem, error) {
	logger.Info("使用流式合并优化",
		zap.Uint("sessionID", session.ID),
		zap.Int("totalChunks", session.TotalChunks),
		zap.Int64("fileSize", session.FileSize))

	// 1. 检查是否已存在相同MD5的文件（秒传）
	var existingFile entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("MD5 = ? AND OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", session.FileID, userID, "Y", "file").
		First(&existingFile).Error

	if err == nil {
		// 文件已存在，直接创建新记录（秒传）
		logger.Info("检测到相同文件，使用秒传", zap.String("md5", session.FileID))

		storageType := session.StorageType
		storagePath := existingFile.StoragePath
		fileSize := session.FileSize
		fileType := session.FileType
		fileMD5 := session.FileID
		accessURL := existingFile.AccessURL
		fileExt := filepath.Ext(session.FileName)

		newFile := &entity.CloudItem{
			BaseModel: entity.BaseModel{
				CreateBy: s.getUsernameByID(ctx, userID),
				UpdateBy: s.getUsernameByID(ctx, userID),
				IsActive: "Y",
			},
			ItemType:    "file",
			Name:        session.FileName,
			ParentID:    session.FolderID,
			OwnerID:     userID,
			StorageType: &storageType,
			StoragePath: storagePath,
			FileSize:    &fileSize,
			FileType:    &fileType,
			FileExt:     &fileExt,
			MD5:         &fileMD5,
			AccessURL:   accessURL,
		}

		if err := s.db.WithContext(ctx).Create(newFile).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败（秒传）", err)
		}

		// 更新会话状态
		s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
			Where("ID = ?", session.ID).
			Update("STATUS", "completed")

		// 异步清理临时文件
		go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

		// 更新配额（只增加文件计数，不增加空间使用，因为文件已存在）
		s.cloudService.UpdateQuota(ctx, userID, 0, 1, 0)

		return newFile, nil
	}

	// 2. 构造最终文件路径
	finalPath := fmt.Sprintf("cloud/%d/%s/%s%s",
		userID,
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		filepath.Ext(session.FileName))

	// 3. 创建管道 - 核心优化点
	// Reader 端用于存储上传，Writer 端用于写入合并后的数据
	pipeReader, pipeWriter := io.Pipe()

	// 用于收集错误和结果
	type uploadResult struct {
		accessURL string
		err       error
	}
	uploadResultChan := make(chan uploadResult, 1)

	type mergeResult struct {
		actualMD5 string
		err       error
	}
	mergeResultChan := make(chan mergeResult, 1)

	// 4. Goroutine 1: 上传数据到存储（从管道读取）
	go func() {
		defer pipeReader.Close()

		// 直接从管道读取并上传，无需临时文件
		accessURL, err := s.storage.Upload(ctx, finalPath, pipeReader, session.FileType)
		uploadResultChan <- uploadResult{accessURL: accessURL, err: err}
	}()

	// 5. Goroutine 2: 合并分片并写入管道 + 计算MD5
	go func() {
		defer pipeWriter.Close()

		// 创建 MD5 计算器
		hash := md5.New()
		// 使用 MultiWriter 同时写入管道和 MD5 计算器
		multiWriter := io.MultiWriter(pipeWriter, hash)

		// 按顺序读取并合并所有分片
		for i := 0; i < session.TotalChunks; i++ {
			chunkPath := fmt.Sprintf("%s/chunk_%d", session.StoragePath, i)

			// 读取分片
			chunkReader, err := s.storage.Download(ctx, chunkPath)
			if err != nil {
				pipeWriter.CloseWithError(errors.Wrap(errors.ErrInternal, fmt.Sprintf("读取分片 %d 失败", i), err))
				mergeResultChan <- mergeResult{err: err}
				return
			}

			// 流式写入：同时写入管道（用于上传）和 MD5 计算器
			written, err := io.Copy(multiWriter, chunkReader)
			chunkReader.Close()

			if err != nil {
				pipeWriter.CloseWithError(errors.Wrap(errors.ErrInternal, fmt.Sprintf("合并分片 %d 失败", i), err))
				mergeResultChan <- mergeResult{err: err}
				return
			}

			logger.Debug("流式合并分片",
				zap.Int("chunkIndex", i),
				zap.Int64("written", written))
		}

		// 计算最终的 MD5
		actualMD5 := hex.EncodeToString(hash.Sum(nil))
		mergeResultChan <- mergeResult{actualMD5: actualMD5, err: nil}
	}()

	// 6. 等待两个 goroutine 完成
	uploadRes := <-uploadResultChan
	mergeRes := <-mergeResultChan

	// 检查合并错误
	if mergeRes.err != nil {
		logger.Error("流式合并失败", zap.Error(mergeRes.err))
		// 尝试删除可能已部分上传的文件
		s.storage.Delete(ctx, finalPath)
		return nil, mergeRes.err
	}

	// 检查上传错误
	if uploadRes.err != nil {
		logger.Error("流式上传失败", zap.Error(uploadRes.err))
		return nil, errors.Wrap(errors.ErrInternal, "上传最终文件失败", uploadRes.err)
	}

	// 7. 验证文件完整性（MD5）
	if mergeRes.actualMD5 != session.FileID {
		logger.Error("文件MD5校验失败",
			zap.String("expected", session.FileID),
			zap.String("actual", mergeRes.actualMD5))
		// 删除已上传的错误文件
		s.storage.Delete(ctx, finalPath)
		return nil, errors.New(errors.ErrInvalidParam, "文件MD5校验失败")
	}

	logger.Info("文件MD5校验成功（流式合并）", zap.String("md5", mergeRes.actualMD5))

	storageType := session.StorageType
	fileSize := session.FileSize
	fileType := session.FileType
	fileMD5 := session.FileID
	fileExt := filepath.Ext(session.FileName)

	// 8. 创建文件记录
	file := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy: s.getUsernameByID(ctx, userID),
			UpdateBy: s.getUsernameByID(ctx, userID),
			IsActive: "Y",
		},
		ItemType:    "file",
		Name:        session.FileName,
		ParentID:    session.FolderID,
		OwnerID:     userID,
		StoragePath: &finalPath,
		FileSize:    &fileSize,
		FileType:    &fileType,
		FileExt:     &fileExt,
		MD5:         &fileMD5,
		StorageType: &storageType,
		AccessURL:   &uploadRes.accessURL,
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		// 删除已上传的文件
		s.storage.Delete(ctx, finalPath)
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	// 9. 更新配额：文件数量+1，空间使用+文件大小（流式优化版本）
	s.cloudService.UpdateQuota(ctx, userID, session.FileSize, 1, 0)

	// 10. 更新会话状态为已完成
	s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", session.ID).
		Update("STATUS", "completed")

	// 11. 异步清理临时文件
	go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

	logger.Info("文件上传完成（流式合并）",
		zap.Uint("fileID", file.ID),
		zap.String("fileName", file.Name),
		zap.Int64("fileSize", *file.FileSize))

	return file, nil
}
