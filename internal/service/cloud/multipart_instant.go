package cloud

import (
	"context"
	"path/filepath"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// findInstantSource 查找可用于秒传的已有文件：同一用户、相同 MD5 与大小、仍有效且有存储路径
func (s *multipartUploadService) findInstantSource(ctx context.Context, userID uint, md5 string, size int64) (*entity.CloudItem, bool) {
	if md5 == "" {
		return nil, false
	}
	var existing entity.CloudItem
	err := s.db.WithContext(ctx).
		Where("MD5 = ? AND FILE_SIZE = ? AND OWNER_ID = ? AND ITEM_TYPE = ? AND IS_ACTIVE = ? AND STORAGE_PATH IS NOT NULL",
			md5, size, userID, "file", "Y").
		Order("ID DESC").
		First(&existing).Error
	if err != nil {
		return nil, false
	}
	return &existing, true
}

// allChunkIndexes 返回 0..total-1，秒传时告诉前端所有分片都已就绪，无需上传
func allChunkIndexes(total int) []int {
	idx := make([]int, total)
	for i := range idx {
		idx[i] = i
	}
	return idx
}

// completeInstant 秒传完成：新建一条指向已有存储对象的文件记录，不上传、不合并、不清理存储
// （会话没有上传任何分片；按会话路径清理存储会按前缀删除对象，可能误删同路径的原文件）
func (s *multipartUploadService) completeInstant(ctx context.Context, session *entity.CloudUploadSession, userID uint, existing *entity.CloudItem) (*entity.CloudItem, error) {
	logger.Info("检测到相同文件，秒传完成", zap.String("md5", session.FileID), zap.Uint("sourceItem", existing.ID))

	username := s.getUsernameByID(ctx, userID)
	storageType := session.StorageType
	if existing.StorageType != nil {
		storageType = *existing.StorageType
	}
	fileSize := session.FileSize
	fileType := session.FileType
	fileMD5 := session.FileID
	fileExt := filepath.Ext(session.FileName)

	// 完整路径 = 父目录路径 + 名称（重名时由 BeforeCreate 自动改名并同步路径末段）
	itemPath := "/" + session.FileName
	if session.FolderID != nil {
		var parentPath string
		s.db.WithContext(ctx).Table("cloud_item").Select("PATH").Where("ID = ?", *session.FolderID).Scan(&parentPath)
		if parentPath != "" {
			itemPath = parentPath + "/" + session.FileName
		}
	}

	newFile := &entity.CloudItem{
		BaseModel: entity.BaseModel{
			CreateBy:     username,
			UpdateBy:     username,
			IsActive:     "Y",
			SysCompanyID: s.cloudService.GetUserCompanyID(ctx, userID),
		},
		ItemType:    "file",
		Name:        session.FileName,
		Path:        itemPath,
		ParentID:    session.FolderID,
		OwnerID:     userID,
		StorageType: &storageType,
		StoragePath: existing.StoragePath,
		FileSize:    &fileSize,
		FileType:    &fileType,
		FileExt:     &fileExt,
		MD5:         &fileMD5,
		AccessURL:   existing.AccessURL,
		Thumbnail:   existing.Thumbnail,
	}
	if err := s.cloudService.CreateItem(ctx, newFile); err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建文件记录失败（秒传）", err)
	}

	s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", session.ID).
		Update("STATUS", "completed")

	// 秒传不新增存储占用，只增加文件数
	s.cloudService.UpdateQuota(ctx, userID, 0, 1, 0)
	return newFile, nil
}
