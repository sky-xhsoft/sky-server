package storage

import (
	"context"
	"io"
)

// Storage 存储接口（支持本地和云存储）
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error)

	// Download 下载文件
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, path string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetURL 获取文件访问URL
	GetURL(ctx context.Context, path string, expireSeconds int) (string, error)

	// ListObjects 列出对象
	ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error)

	// CopyObject 复制对象
	CopyObject(ctx context.Context, srcPath, dstPath string) error

	// GetObjectInfo 获取对象信息
	GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error)
}

// Object 存储对象
type Object struct {
	Key          string `json:"key"`          // 对象键
	Size         int64  `json:"size"`         // 大小（字节）
	LastModified string `json:"lastModified"` // 最后修改时间
	ETag         string `json:"etag"`         // ETag
}

// ObjectInfo 对象详细信息
type ObjectInfo struct {
	Key          string            `json:"key"`          // 对象键
	Size         int64             `json:"size"`         // 大小（字节）
	ContentType  string            `json:"contentType"`  // 内容类型
	LastModified string            `json:"lastModified"` // 最后修改时间
	ETag         string            `json:"etag"`         // ETag
	Metadata     map[string]string `json:"metadata"`     // 元数据
}

// UploadResult 上传结果
type UploadResult struct {
	URL      string `json:"url"`      // 访问URL
	Key      string `json:"key"`      // 对象键
	ETag     string `json:"etag"`     // ETag
	Size     int64  `json:"size"`     // 大小
	Bucket   string `json:"bucket"`   // 存储桶
	Location string `json:"location"` // 位置
}
