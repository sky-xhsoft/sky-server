package cloud

import (
	"context"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// ShareDownloadInfo 分享下载地址：URL 为空表示文件没有直链，需走后端流式下载接口
type ShareDownloadInfo struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// ClaimShareDownload 计入一次分享下载：未设上限或未达上限时下载次数 +1，已达上限返回错误。
// 检查与计数在同一条 UPDATE 中完成，并发下载不会超出上限
func (s *service) ClaimShareDownload(ctx context.Context, shareID uint) error {
	res := s.db.WithContext(ctx).Model(&entity.CloudShare{}).
		Where("ID = ? AND (MAX_DOWNLOADS IS NULL OR MAX_DOWNLOADS <= 0 OR DOWNLOAD_COUNT < MAX_DOWNLOADS)", shareID).
		Update("DOWNLOAD_COUNT", gorm.Expr("DOWNLOAD_COUNT + 1"))
	if res.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "更新下载次数失败", res.Error)
	}
	if res.RowsAffected == 0 {
		return errors.New(errors.ErrInvalidParam, "分享下载次数已达上限")
	}
	return nil
}

// PrepareShareDownload 点击下载时调用：文件有直链则计入一次下载并返回直链；
// 没有直链时不计数，由调用方改走流式下载接口（该接口自己计数）。
// 调用方须已校验访问者与访问凭证，文件夹分享还须已校验 fileID 位于分享范围内
func (s *service) PrepareShareDownload(ctx context.Context, share *entity.CloudShare, fileID uint) (*ShareDownloadInfo, error) {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	info := &ShareDownloadInfo{Name: file.Name}
	if file.AccessURL == nil || !(strings.HasPrefix(*file.AccessURL, "http://") || strings.HasPrefix(*file.AccessURL, "https://")) {
		return info, nil
	}
	if err := s.ClaimShareDownload(ctx, share.ID); err != nil {
		return nil, err
	}
	info.URL = *file.AccessURL
	return info, nil
}
