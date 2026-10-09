package storage

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// IsExternalURL 存储路径是否为完整的外部地址（如腾讯云直播录制、高光切片回调的点播地址）。
// 这类云盘记录只保存外部地址，文件本身不在公司存储中。
func IsExternalURL(path string) bool {
	p := strings.ToLower(strings.TrimSpace(path))
	return strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://")
}

// ExternalURLStorage 外部地址"存储"：按地址直接读取，不支持写入；
// 删除为空操作——删除云盘记录时绝不删除外部的源文件（如点播中的录制文件）。
type ExternalURLStorage struct {
	client *http.Client
}

// NewExternalURLStorage 创建外部地址存储
func NewExternalURLStorage() *ExternalURLStorage {
	return &ExternalURLStorage{client: &http.Client{
		// 不设整体超时（大文件下载耗时长），只限制建连与等待响应头
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}}
}

var errExternalReadOnly = errors.New(errors.ErrInvalidParam, "外部地址文件不支持该操作")

func (s *ExternalURLStorage) get(ctx context.Context, method, url string) (*http.Response, error) {
	if !IsExternalURL(url) {
		return nil, errors.New(errors.ErrInvalidParam, "不是外部地址")
	}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建请求失败", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(errors.ErrStorage, "读取外部文件失败", err)
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, errors.New(errors.ErrStorage, fmt.Sprintf("读取外部文件失败：HTTP %d", resp.StatusCode))
	}
	return resp, nil
}

// Download 读取外部地址的内容
func (s *ExternalURLStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	resp, err := s.get(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PresignedDownloadURL 外部地址本身可直接读取
func (s *ExternalURLStorage) PresignedDownloadURL(ctx context.Context, path string, expire time.Duration) (string, error) {
	return path, nil
}

// GetURL 返回外部地址
func (s *ExternalURLStorage) GetURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	return path, nil
}

// Exists 外部地址是否可访问
func (s *ExternalURLStorage) Exists(ctx context.Context, path string) (bool, error) {
	resp, err := s.get(ctx, http.MethodHead, path)
	if err != nil {
		return false, nil
	}
	resp.Body.Close()
	return true, nil
}

// GetObjectInfo 外部文件信息（大小、类型）
func (s *ExternalURLStorage) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	resp, err := s.get(ctx, http.MethodHead, path)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return &ObjectInfo{Key: path, Size: resp.ContentLength, ContentType: resp.Header.Get("Content-Type"), ETag: resp.Header.Get("ETag")}, nil
}

// Delete 空操作：删除云盘记录不删除外部源文件
func (s *ExternalURLStorage) Delete(ctx context.Context, path string) error {
	return nil
}

// Upload 不支持
func (s *ExternalURLStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	return "", errExternalReadOnly
}

// ListObjects 不支持
func (s *ExternalURLStorage) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error) {
	return nil, errExternalReadOnly
}

// CopyObject 不支持
func (s *ExternalURLStorage) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	return errExternalReadOnly
}

// PresignedUploadURL 不支持
func (s *ExternalURLStorage) PresignedUploadURL(ctx context.Context, path string, expireSeconds int, contentType string) (string, error) {
	return "", errExternalReadOnly
}

// InitiateMultipartUpload 不支持
func (s *ExternalURLStorage) InitiateMultipartUpload(ctx context.Context, path string) (string, error) {
	return "", errExternalReadOnly
}

// PresignedChunkUploadURL 不支持
func (s *ExternalURLStorage) PresignedChunkUploadURL(ctx context.Context, path string, uploadID string, chunkNumber int, expireSeconds int, contentType string) (string, error) {
	return "", errExternalReadOnly
}

// CompleteMultipartUpload 不支持
func (s *ExternalURLStorage) CompleteMultipartUpload(ctx context.Context, path string, uploadID string, partETags []string) (string, error) {
	return "", errExternalReadOnly
}

var _ Storage = (*ExternalURLStorage)(nil)
