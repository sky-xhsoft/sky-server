package cloud

import (
	"context"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"go.uber.org/zap"
)

type fakePartLister struct{ parts []storage.UploadedPart }

func (f *fakePartLister) PresignedChunkUploadURL(ctx context.Context, path, uploadID string, chunkNumber, expireSeconds int, contentType string) (string, error) {
	return "", nil
}

func (f *fakePartLister) ListUploadedParts(ctx context.Context, path, uploadID string) ([]storage.UploadedPart, error) {
	return f.parts, nil
}

func TestVerifiedPartETags(t *testing.T) {
	if logger.Logger == nil {
		logger.Logger = zap.NewNop()
	}
	// 文件 25 字节、分片 10 字节：3 片，大小 10/10/5
	session := &entity.CloudUploadSession{FileSize: 25, ChunkSize: 10, TotalChunks: 3, StorageUploadID: "u"}
	ok := []storage.UploadedPart{{PartNumber: 2, ETag: "b", Size: 10}, {PartNumber: 1, ETag: "a", Size: 10}, {PartNumber: 3, ETag: "c", Size: 5}}

	etags, err := verifiedPartETags(context.Background(), &fakePartLister{ok}, session)
	if err != nil || len(etags) != 3 || etags[0] != "a" || etags[1] != "b" || etags[2] != "c" {
		t.Fatalf("分块齐全时应按编号返回 ETag，got %v %v", etags, err)
	}

	cases := map[string][]storage.UploadedPart{
		"缺少分块":          ok[:2],
		"分块大小不正确":       {ok[0], ok[1], {PartNumber: 3, ETag: "c", Size: 6}},
		"多传了分块（超过申报大小）": append(append([]storage.UploadedPart{}, ok...), storage.UploadedPart{PartNumber: 4, ETag: "d", Size: 10}),
	}
	for name, parts := range cases {
		if _, err := verifiedPartETags(context.Background(), &fakePartLister{parts}, session); err == nil {
			t.Errorf("%s：应校验失败", name)
		}
	}
}

func TestUploadModeOf(t *testing.T) {
	native := &entity.CloudUploadSession{StorageUploadID: "u"}
	if got := uploadModeOf(&storage.TencentCOS{}, native); got != UploadModeDirect {
		t.Errorf("COS 原生分块会话应直传，got %s", got)
	}
	if got := uploadModeOf(&storage.TencentCOS{}, &entity.CloudUploadSession{}); got != UploadModeRelay {
		t.Errorf("未初始化原生分块时应中转，got %s", got)
	}
	if got := uploadModeOf(&storage.LocalStorage{}, native); got != UploadModeRelay {
		t.Errorf("本地存储应中转，got %s", got)
	}
}

func TestMaxRelayUploadSize(t *testing.T) {
	if MaxRelayUploadSize != 50<<20 {
		t.Fatalf("经服务器上传上限应为 50MB（与前端 RESUMABLE_THRESHOLD 一致），got %d", MaxRelayUploadSize)
	}
}
