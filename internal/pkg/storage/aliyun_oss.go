package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// AliyunOSS 阿里云OSS存储实现
type AliyunOSS struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	endpoint   string
	cdnDomain  string // CDN加速域名（可选）
}

// AliyunOSSConfig 阿里云OSS配置
type AliyunOSSConfig struct {
	Endpoint        string // OSS endpoint（如: oss-cn-hangzhou.aliyuncs.com）
	AccessKeyID     string // AccessKey ID
	AccessKeySecret string // AccessKey Secret
	BucketName      string // Bucket名称
	CDNDomain       string // CDN加速域名（可选，如: https://cdn.example.com）
}

// IsAliyunOSS 判断是否是阿里云OSS存储（用于类型断言）
func (s *AliyunOSS) IsAliyunOSS() bool {
	return true
}

// NewAliyunOSS 创建阿里云OSS存储
func NewAliyunOSS(cfg *AliyunOSSConfig) (Storage, error) {
	// 创建OSSClient实例
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %w", err)
	}

	return &AliyunOSS{
		client:     client,
		bucket:     bucket,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
		cdnDomain:  cfg.CDNDomain,
	}, nil
}

// Upload 上传文件
func (s *AliyunOSS) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	// 设置上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
	}

	// 上传文件
	err := s.bucket.PutObject(path, reader, options...)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "上传文件到OSS失败", err)
	}

	// 返回访问URL
	url, err := s.GetURL(ctx, path, 0)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Download 下载文件
func (s *AliyunOSS) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	// 获取文件
	body, err := s.bucket.GetObject(path)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "下载文件失败", err)
	}

	return body, nil
}

// Delete 删除文件
func (s *AliyunOSS) Delete(ctx context.Context, path string) error {
	err := s.bucket.DeleteObject(path)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "删除文件失败", err)
	}
	return nil
}

// Exists 检查文件是否存在
func (s *AliyunOSS) Exists(ctx context.Context, path string) (bool, error) {
	exists, err := s.bucket.IsObjectExist(path)
	if err != nil {
		return false, errors.Wrap(errors.ErrInternal, "检查文件是否存在失败", err)
	}
	return exists, nil
}

// GetURL 获取文件访问URL
func (s *AliyunOSS) GetURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	// 如果配置了CDN域名，使用CDN域名
	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, path), nil
	}

	// 如果需要签名URL（私有读）
	if expireSeconds > 0 {
		signedURL, err := s.bucket.SignURL(path, oss.HTTPGet, int64(expireSeconds))
		if err != nil {
			return "", errors.Wrap(errors.ErrInternal, "生成签名URL失败", err)
		}
		return signedURL, nil
	}

	// 公共读URL
	return fmt.Sprintf("https://%s.%s/%s", s.bucketName, s.endpoint, path), nil
}

// ListObjects 列出对象
func (s *AliyunOSS) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error) {
	// 设置列举选项
	options := []oss.Option{
		oss.Prefix(prefix),
	}
	if maxKeys > 0 {
		options = append(options, oss.MaxKeys(maxKeys))
	}

	// 列举对象
	lsRes, err := s.bucket.ListObjects(options...)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "列举对象失败", err)
	}

	// 转换结果
	objects := make([]Object, 0, len(lsRes.Objects))
	for _, obj := range lsRes.Objects {
		objects = append(objects, Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified.Format(time.RFC3339),
			ETag:         obj.ETag,
		})
	}

	return objects, nil
}

// CopyObject 复制对象
func (s *AliyunOSS) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	// 源对象
	srcObject := fmt.Sprintf("%s/%s", s.bucketName, srcPath)

	// 复制对象
	_, err := s.bucket.CopyObject(srcObject, dstPath)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "复制对象失败", err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *AliyunOSS) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	// 获取对象元数据
	meta, err := s.bucket.GetObjectDetailedMeta(path)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "获取对象信息失败", err)
	}

	// 解析最后修改时间
	lastModified := meta.Get("Last-Modified")

	// 构建对象信息
	info := &ObjectInfo{
		Key:          path,
		Size:         0, // 需要从meta中解析
		ContentType:  meta.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         meta.Get("ETag"),
		Metadata:     make(map[string]string),
	}

	// 提取自定义元数据
	for key, values := range meta {
		if len(values) > 0 {
			info.Metadata[key] = values[0]
		}
	}

	return info, nil
}

// BatchDelete 批量删除对象
func (s *AliyunOSS) BatchDelete(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// 批量删除
	_, err := s.bucket.DeleteObjects(paths)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "批量删除对象失败", err)
	}

	return nil
}

// GetUploadURL 获取上传签名URL（用于前端直传）
func (s *AliyunOSS) GetUploadURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	signedURL, err := s.bucket.SignURL(path, oss.HTTPPut, int64(expireSeconds))
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "生成上传签名URL失败", err)
	}
	return signedURL, nil
}

// GetPostPolicy 获取POST策略（用于表单上传）
func (s *AliyunOSS) GetPostPolicy(ctx context.Context, dir string, expireSeconds int64) (map[string]string, error) {
	// 设置过期时间
	expireTime := time.Now().Unix() + expireSeconds

	// 构建POST策略
	conditions := []interface{}{
		[]interface{}{"content-length-range", 0, 1024 * 1024 * 100}, // 最大100MB
		[]interface{}{"starts-with", "$key", dir},                   // 限制上传目录
	}

	policy := map[string]interface{}{
		"expiration": time.Unix(expireTime, 0).UTC().Format("2006-01-02T15:04:05Z"),
		"conditions": conditions,
	}
	fmt.Println(policy)

	// TODO: 生成签名
	// 这里需要实现POST策略的签名逻辑

	return map[string]string{
		"OSSAccessKeyId": s.client.Config.AccessKeyID,
		"policy":         "", // Base64编码的policy
		"signature":      "", // 签名
		"expire":         fmt.Sprintf("%d", expireTime),
		"dir":            dir,
		"host":           fmt.Sprintf("https://%s.%s", s.bucketName, s.endpoint),
	}, nil
}

// PresignedUploadURL 获取预签名上传URL（用于前端直传）
func (s *AliyunOSS) PresignedUploadURL(ctx context.Context, path string, expireSeconds int, contentType string) (string, error) {
	signedURL, err := s.bucket.SignURL(path, oss.HTTPPut, int64(expireSeconds))
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "生成预签名上传URL失败", err)
	}
	return signedURL, nil
}

// InitiateMultipartUpload 初始化分块上传
func (s *AliyunOSS) InitiateMultipartUpload(ctx context.Context, path string) (string, error) {
	// 设置初始化分块上传选项
	options := []oss.Option{
		oss.ContentType("application/octet-stream"),
	}

	// 初始化分块上传
	imur, err := s.bucket.InitiateMultipartUpload(path, options...)
	if err != nil {
		logger.Error("Failed to initiate multipart upload",
			zap.String("path", path),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "初始化分块上传失败", err)
	}

	return imur.UploadID, nil
}

// PresignedChunkUploadURL 获取分块预签名上传URL
func (s *AliyunOSS) PresignedChunkUploadURL(ctx context.Context, path string, uploadID string, chunkNumber int, expireSeconds int, contentType string) (string, error) {
	// 简化实现：直接生成预签名URL，不添加额外参数
	// OSS分片编号从1开始，我们这里前端从0开始，需要加1
	ossPartNumber := chunkNumber + 1

	// 生成预签名URL - 使用基本签名
	signedURL, err := s.bucket.SignURL(path, oss.HTTPPut, int64(expireSeconds))
	if err != nil {
		logger.Error("Failed to generate presigned chunk upload URL",
			zap.String("path", path),
			zap.String("uploadID", uploadID),
			zap.Int("chunkNumber", chunkNumber),
			zap.Int("ossPartNumber", ossPartNumber),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "生成分块预签名URL失败", err)
	}

	return signedURL, nil
}

// CompleteMultipartUpload 完成分块上传，让服务端合并所有分块
func (s *AliyunOSS) CompleteMultipartUpload(ctx context.Context, path string, uploadID string, partETags []string) (string, error) {
	// 构造parts
	var parts []oss.UploadPart
	for i, etag := range partETags {
		// OSS分片编号从1开始
		parts = append(parts, oss.UploadPart{
			PartNumber: i + 1,
			ETag:       etag,
		})
	}

	// 完成分块上传
	cmur, err := s.bucket.CompleteMultipartUpload(
		oss.InitiateMultipartUploadResult{UploadID: uploadID, Key: path},
		parts)
	if err != nil {
		logger.Error("Failed to complete multipart upload",
			zap.String("path", path),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "完成分块上传失败", err)
	}

	// 返回最终文件的ETag
	return cmur.ETag, nil
}
