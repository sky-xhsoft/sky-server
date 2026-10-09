package cloud

import (
	"context"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadStatusMerging 服务端合并中（后台执行）
const UploadStatusMerging = "merging"

// mergingStaleAfter 合并中状态超过该时长视为中断（如服务重启），允许重新发起合并
const mergingStaleAfter = time.Hour

// restoreStatus 后台合并失败后恢复的状态
func restoreStatus(status string) string {
	if status == UploadStatusMerging || status == "" {
		return "uploading"
	}
	return status
}

// StartComplete 完成分片上传。
// 原生分块（COS 端合并）与秒传很快，同步完成并返回文件；
// 需要服务端合并分片的（本地存储等），合并耗时与文件大小成正比，大文件会超过网关超时，
// 因此转为后台合并并返回 nil，调用方通过 GetUploadStatus 轮询（merging → completed / 失败时带 errorMessage）。
func (s *multipartUploadService) StartComplete(ctx context.Context, sessionID uint, userID uint) (*entity.CloudItem, error) {
	var session entity.CloudUploadSession
	if err := s.db.WithContext(ctx).
		Where("ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", sessionID, userID, "Y").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "上传会话不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询上传会话失败", err)
	}

	if session.Status == UploadStatusMerging && time.Since(session.UpdateTime) < mergingStaleAfter {
		return nil, nil
	}
	if session.Status == "completed" {
		return s.CompleteUpload(ctx, sessionID, userID)
	}

	// 原生分块全部带 ETag、本地分片已直接组装（只需改名）或本会话未上传任何分片（秒传）：同步完成
	var uploaded, withETag int64
	s.db.WithContext(ctx).Model(&entity.CloudChunkRecord{}).
		Where("SESSION_ID = ? AND UPLOADED = ?", sessionID, true).Count(&uploaded)
	s.db.WithContext(ctx).Model(&entity.CloudChunkRecord{}).
		Where("SESSION_ID = ? AND UPLOADED = ? AND ETAG <> ''", sessionID, true).Count(&withETag)
	if uploaded == 0 || (session.StorageUploadID != "" && int(withETag) == session.TotalChunks) || s.isAssembled(ctx, &session) {
		return s.CompleteUpload(ctx, sessionID, userID)
	}

	// 抢占合并：只有一个请求能把状态改为 merging，避免重复点击导致重复合并
	res := s.db.WithContext(ctx).Model(&entity.CloudUploadSession{}).
		Where("ID = ? AND (STATUS <> ? OR UPDATE_TIME < ?)", sessionID, UploadStatusMerging, time.Now().Add(-mergingStaleAfter)).
		Updates(map[string]interface{}{"STATUS": UploadStatusMerging, "ERROR_MESSAGE": "", "UPDATE_TIME": time.Now()})
	if res.Error != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "更新会话状态失败", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}

	go func(previous string) {
		start := time.Now()
		if _, err := s.CompleteUpload(context.Background(), sessionID, userID); err != nil {
			logger.Error("后台合并分片失败", zap.Uint("sessionID", sessionID), zap.Error(err))
			// 恢复为原状态，保留已上传分片，可重新发起合并
			s.db.Model(&entity.CloudUploadSession{}).Where("ID = ?", sessionID).
				Updates(map[string]interface{}{"STATUS": previous, "ERROR_MESSAGE": err.Error(), "UPDATE_TIME": time.Now()})
			return
		}
		logger.Info("后台合并分片完成", zap.Uint("sessionID", sessionID), zap.Duration("elapsed", time.Since(start)))
	}(restoreStatus(session.Status))

	return nil, nil
}
