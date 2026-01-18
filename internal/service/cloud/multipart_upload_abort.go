package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AbortUpload 取消上传
func (s *multipartUploadService) AbortUpload(ctx context.Context, sessionID uint, userID uint) error {
	logger.Info("取消上传", zap.Uint("sessionID", sessionID), zap.Uint("userID", userID))

	// 1. 获取上传会话
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	// 2. 更新会话状态
	if err := s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ?", sessionID).
		Updates(map[string]interface{}{
			"STATUS":      "failed",
			"IS_ACTIVE":   "N",
			"UPDATE_BY":   fmt.Sprintf("user_%d", userID),
			"UPDATE_TIME": time.Now(),
		}).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新会话状态失败", err)
	}

	// 3. 异步清理临时文件
	go s.cleanupChunks(context.Background(), session.ID, session.StoragePath)

	return nil
}
