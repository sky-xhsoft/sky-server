package cloud

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"go.uber.org/zap"
)

// isAssembled 所有分片是否都已直接写入组装文件（本地存储），是则完成时只需改名
func (s *multipartUploadService) isAssembled(ctx context.Context, session *entity.CloudUploadSession) bool {
	if session.StorageUploadID != "" {
		return false
	}
	var n int64
	s.db.WithContext(ctx).Model(&entity.CloudChunkRecord{}).
		Where("SESSION_ID = ? AND UPLOADED = ? AND CHUNK_PATH = ?", session.ID, true, assembledPartPath(session.StoragePath)).
		Count(&n)
	return int(n) == session.TotalChunks
}

// completeAssembled 完成直接组装的上传：校验大小后把组装文件改名为最终文件，不再读写文件内容。
// 每个分片写入前已校验 MD5 与大小，并按偏移写入，文件内容即完整文件。
func (s *multipartUploadService) completeAssembled(ctx context.Context, session *entity.CloudUploadSession, userID uint, storageInstance storage.Storage) (*entity.CloudItem, error) {
	renamer, ok := storageInstance.(interface {
		Rename(ctx context.Context, srcPath, dstPath string) (string, error)
	})
	if !ok {
		return nil, errors.New(errors.ErrStorage, "存储不支持移动文件")
	}

	partPath := assembledPartPath(session.StoragePath)
	info, err := storageInstance.GetObjectInfo(ctx, partPath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "读取上传文件失败", err)
	}
	if info.Size != session.FileSize {
		return nil, errors.New(errors.ErrInvalidParam,
			fmt.Sprintf("文件大小不一致：期望 %d，实际 %d", session.FileSize, info.Size))
	}

	// 已有相同文件：秒传引用已有文件，删除本次上传的数据
	if existing, ok := s.findInstantSource(ctx, userID, session.FileID, session.FileSize); ok {
		item, err := s.completeInstant(ctx, session, userID, existing)
		if err == nil {
			go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)
		}
		return item, err
	}

	start := time.Now()
	finalPath := fmt.Sprintf("cloud/%d/%s/%s%s", userID, time.Now().Format("2006/01/02"), uuid.New().String(), filepath.Ext(session.FileName))
	accessURL, err := renamer.Rename(ctx, partPath, finalPath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "保存文件失败", err)
	}

	username := s.getUsernameByID(ctx, userID)
	storageType := session.StorageType
	fileSize := session.FileSize
	fileType := session.FileType
	fileMD5 := session.FileID
	fileExt := filepath.Ext(session.FileName)
	file := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy:     username,
			UpdateBy:     username,
			IsActive:     "Y",
			SysCompanyID: s.cloudService.GetUserCompanyID(ctx, userID),
		},
		ItemType:    "file",
		Name:        session.FileName,
		ParentID:    session.FolderID,
		Path:        s.childPath(ctx, session.FolderID, session.FileName),
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
		// 移回原处，保留已上传数据以便重试
		if _, rerr := renamer.Rename(ctx, finalPath, partPath); rerr != nil {
			logger.Error("回滚上传文件失败", zap.String("path", finalPath), zap.Error(rerr))
		}
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败", err)
	}

	s.cloudService.UpdateQuota(ctx, userID, session.FileSize, 1, 0)
	s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", session.ID).
		Update("STATUS", "completed")
	go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

	logger.Info("文件上传完成（分片直接组装）",
		zap.Uint("fileID", file.ID),
		zap.String("fileName", file.Name),
		zap.Int64("fileSize", fileSize),
		zap.Duration("elapsed", time.Since(start)))
	return file, nil
}
