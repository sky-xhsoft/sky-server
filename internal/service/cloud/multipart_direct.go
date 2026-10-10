package cloud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"go.uber.org/zap"
)

// 分片上传方式
const (
	// UploadModeDirect 浏览器用预签名地址直接把分块传到云存储（不经过应用服务器）
	UploadModeDirect = "direct"
	// UploadModeRelay 分块经应用服务器中转（本地存储，或云存储不支持直传时）
	UploadModeRelay = "relay"
)

// MaxRelayUploadSize 云存储支持直传时，经应用服务器上传的单个文件上限（50MB）；
// 更大的文件由前端分片直传到云存储，不占用应用服务器带宽
const MaxRelayUploadSize int64 = 50 << 20

// directUploadExpireSeconds 直传分块地址的有效期
const directUploadExpireSeconds = 3600

// maxPresignBatch 一次最多签发的分块地址数
const maxPresignBatch = 200

// directPartStorage 支持浏览器直传原生分块的存储（腾讯云 COS）：
// 能签发分块上传地址，完成时能查询存储端实际收到的分块，用于校验
type directPartStorage interface {
	PresignedChunkUploadURL(ctx context.Context, path string, uploadID string, chunkNumber int, expireSeconds int, contentType string) (string, error)
	ListUploadedParts(ctx context.Context, path string, uploadID string) ([]storage.UploadedPart, error)
}

// uploadModeOf 会话的上传方式：已初始化原生分块且存储支持直传时为 direct，否则 relay
func uploadModeOf(st storage.Storage, session *entity.CloudUploadSession) string {
	if session.StorageUploadID == "" {
		return UploadModeRelay
	}
	if _, ok := st.(directPartStorage); ok {
		return UploadModeDirect
	}
	return UploadModeRelay
}

// sessionUploadMode 查询会话所用存储后判断上传方式；存储不可用时按中转处理
func (s *multipartUploadService) sessionUploadMode(ctx context.Context, session *entity.CloudUploadSession) string {
	st, err := s.getSessionStorage(ctx, session)
	if err != nil {
		return UploadModeRelay
	}
	return uploadModeOf(st, session)
}

// GetChunkPresignedURLs 批量签发分块直传地址（减少请求次数）
func (s *multipartUploadService) GetChunkPresignedURLs(ctx context.Context, sessionID uint, chunkIndexes []int, userID uint) ([]*ChunkPresignedURL, error) {
	if len(chunkIndexes) == 0 {
		return []*ChunkPresignedURL{}, nil
	}
	if len(chunkIndexes) > maxPresignBatch {
		return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("一次最多签发 %d 个分块地址", maxPresignBatch))
	}
	session, err := s.ValidateUploadSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	st, err := s.getSessionStorage(ctx, session)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "获取存储实例失败", err)
	}
	direct, ok := st.(directPartStorage)
	if !ok || session.StorageUploadID == "" {
		return nil, errors.New(errors.ErrInvalidParam, "该存储不支持直传，请经服务器上传")
	}

	result := make([]*ChunkPresignedURL, 0, len(chunkIndexes))
	for _, index := range chunkIndexes {
		if err := s.ValidateChunkIndex(index, session); err != nil {
			return nil, err
		}
		u, err := direct.PresignedChunkUploadURL(ctx, session.StoragePath, session.StorageUploadID, index, directUploadExpireSeconds, "")
		if err != nil {
			return nil, err
		}
		result = append(result, &ChunkPresignedURL{SessionID: sessionID, ChunkIndex: index, PresignedURL: u, ExpireSeconds: directUploadExpireSeconds})
	}
	return result, nil
}

// expectedChunkSize 第 index 个分片应有的大小（最后一片可能较小）
func expectedChunkSize(session *entity.CloudUploadSession, index int) int64 {
	size := int64(session.ChunkSize)
	if rest := session.FileSize - int64(index)*size; rest < size {
		return rest
	}
	return size
}

// verifiedPartETags 以存储端实际收到的分块为准，校验分块齐全且大小与会话一致，返回按编号排序的 ETag。
// 直传时分块内容不经过服务器，这里是唯一的校验：防止缺块，也防止上传的数据超过申报的文件大小（绕过配额）
func verifiedPartETags(ctx context.Context, lister directPartStorage, session *entity.CloudUploadSession) ([]string, error) {
	parts, err := lister.ListUploadedParts(ctx, session.StoragePath, session.StorageUploadID)
	if err != nil {
		return nil, err
	}
	byNumber := make(map[int]storage.UploadedPart, len(parts))
	for _, p := range parts {
		byNumber[p.PartNumber] = p
	}
	etags := make([]string, session.TotalChunks)
	for i := 0; i < session.TotalChunks; i++ {
		p, ok := byNumber[i+1]
		if !ok {
			return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片 %d 未上传到存储，请重新上传", i))
		}
		if want := expectedChunkSize(session, i); p.Size != want {
			logger.Warn("分块大小与会话不一致", zap.Uint("sessionID", session.ID), zap.Int("chunkIndex", i),
				zap.Int64("want", want), zap.Int64("got", p.Size))
			return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("分片 %d 大小不正确，请重新上传", i))
		}
		etags[i] = p.ETag
	}
	if len(byNumber) != session.TotalChunks {
		return nil, errors.New(errors.ErrInvalidParam, "存储端分片数量与文件不一致，请重新上传")
	}
	return etags, nil
}

// CheckRelayUpload 经服务器上传（非分片）前检查：公司云存储支持直传时，超过 50MB 的文件拒绝中转
func (s *service) CheckRelayUpload(ctx context.Context, userID uint, fileSize int64) error {
	if fileSize <= MaxRelayUploadSize {
		return nil
	}
	st, err := s.getCompanyStorage(ctx, userID)
	if err != nil {
		return nil // 存储不可用时由后续上传流程报错
	}
	if _, ok := st.(directPartStorage); ok {
		return errors.New(errors.ErrInvalidParam, "超过 50MB 的文件需直传到云存储，请刷新页面后重新上传")
	}
	return nil
}
