package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
)

// TencentCOS 腾讯云COS存储实现
type TencentCOS struct {
	bucketURL  string // Bucket URL
	secretID   string // Secret ID
	secretKey  string // Secret Key
	bucketName string // Bucket名称
	region     string // 区域
	cdnDomain  string // CDN加速域名
	httpClient *http.Client
}

// TencentCOSConfig 腾讯云COS配置
type TencentCOSConfig struct {
	BucketURL  string // Bucket URL（如: https://your-bucket-name.cos.ap-guangzhou.myqcloud.com）
	SecretID   string // Secret ID
	SecretKey  string // Secret Key
	BucketName string // Bucket名称
	Region     string // 区域（如: ap-guangzhou）
	CDNDomain  string // CDN加速域名（可选）
}

// NewTencentCOS 创建腾讯云COS存储
func NewTencentCOS(cfg *TencentCOSConfig) (Storage, error) {
	// 验证配置
	if cfg.BucketURL == "" {
		return nil, fmt.Errorf("BucketURL不能为空")
	}
	if cfg.SecretID == "" {
		return nil, fmt.Errorf("SecretID不能为空")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("SecretKey不能为空")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("BucketName不能为空")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("Region不能为空")
	}

	// 增加超时时间以支持更大分片上传（50MB分片）
	timeout := 5 * time.Minute // 5分钟超时，足够支持大文件上传

	return &TencentCOS{
		bucketURL:  cfg.BucketURL,
		secretID:   cfg.SecretID,
		secretKey:  cfg.SecretKey,
		bucketName: cfg.BucketName,
		region:     cfg.Region,
		cdnDomain:  cfg.CDNDomain,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Upload 上传文件
func (s *TencentCOS) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	// 构建请求URL
	uploadURL := fmt.Sprintf("%s/%s", s.bucketURL, path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, reader)
	if err != nil {
		logger.Error("Failed to create upload request",
			zap.String("path", path),
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return "", errors.Wrap(errors.ErrInternal, "创建上传请求失败", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", contentType)
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to upload file to COS",
			zap.String("path", path),
			zap.String("contentType", contentType),
			zap.Error(err),
		)
		return "", errors.Wrap(errors.ErrInternal, "上传文件到COS失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to upload file",
			zap.String("path", path),
			zap.String("contentType", contentType),
			zap.Int("statusCode", resp.StatusCode),
		)
		return "", errors.Wrap(errors.ErrInternal, fmt.Sprintf("上传文件失败，状态码: %d", resp.StatusCode), nil)
	}

	// 返回访问URL
	url, err := s.GetURL(ctx, path, 0)
	if err != nil {
		logger.Error("Failed to get URL after upload",
			zap.String("path", path),
			zap.Error(err),
		)
		return "", err
	}

	return url, nil
}

// Download 下载文件
func (s *TencentCOS) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	// 构建请求URL
	downloadURL := fmt.Sprintf("%s/%s", s.bucketURL, path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		logger.Error("Failed to create download request",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrInternal, "创建下载请求失败", err)
	}

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to download file",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrResourceNotFound, "下载文件失败", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to download file",
			zap.String("path", path),
			zap.Int("statusCode", resp.StatusCode),
		)
		resp.Body.Close()
		return nil, errors.Wrap(errors.ErrResourceNotFound, fmt.Sprintf("下载文件失败，状态码: %d", resp.StatusCode), nil)
	}

	return resp.Body, nil
}

// Delete 删除文件
func (s *TencentCOS) Delete(ctx context.Context, path string) error {
	// 构建请求URL
	deleteURL := fmt.Sprintf("%s/%s", s.bucketURL, path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		logger.Error("Failed to create delete request",
			zap.String("path", path),
			zap.Error(err),
		)
		return errors.Wrap(errors.ErrInternal, "创建删除请求失败", err)
	}

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to delete file",
			zap.String("path", path),
			zap.Error(err),
		)
		return errors.Wrap(errors.ErrInternal, "删除文件失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusNoContent {
		logger.Error("Failed to delete file",
			zap.String("path", path),
			zap.Int("statusCode", resp.StatusCode),
		)
		return errors.Wrap(errors.ErrInternal, fmt.Sprintf("删除文件失败，状态码: %d", resp.StatusCode), nil)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *TencentCOS) Exists(ctx context.Context, path string) (bool, error) {
	// 构建请求URL
	existsURL := fmt.Sprintf("%s/%s", s.bucketURL, path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "HEAD", existsURL, nil)
	if err != nil {
		logger.Error("Failed to create exists check request",
			zap.String("path", path),
			zap.Error(err),
		)
		return false, errors.Wrap(errors.ErrInternal, "创建检查请求失败", err)
	}

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to check if file exists",
			zap.String("path", path),
			zap.Error(err),
		)
		return false, errors.Wrap(errors.ErrInternal, "检查文件是否存在失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		logger.Error("Failed to check if file exists",
			zap.String("path", path),
			zap.Int("statusCode", resp.StatusCode),
		)
		return false, errors.Wrap(errors.ErrInternal, fmt.Sprintf("检查文件是否存在失败，状态码: %d", resp.StatusCode), nil)
	}
}

// GetURL 获取文件访问URL
func (s *TencentCOS) GetURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	// 如果配置了CDN域名，使用CDN域名
	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, path), nil
	}

	// 公共读URL
	return fmt.Sprintf("%s/%s", s.bucketURL, path), nil
}

// ListObjects 列出对象
func (s *TencentCOS) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error) {
	// 构建请求URL
	listURL := fmt.Sprintf("%s/?prefix=%s&max-keys=%d", s.bucketURL, prefix, maxKeys)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		logger.Error("Failed to create list objects request",
			zap.String("prefix", prefix),
			zap.Int("maxKeys", maxKeys),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrInternal, "创建列举请求失败", err)
	}

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to list objects",
			zap.String("prefix", prefix),
			zap.Int("maxKeys", maxKeys),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrInternal, "列举对象失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to list objects",
			zap.String("prefix", prefix),
			zap.Int("maxKeys", maxKeys),
			zap.Int("statusCode", resp.StatusCode),
		)
		return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("列举对象失败，状态码: %d", resp.StatusCode), nil)
	}

	// TODO: 解析XML响应
	// 这里需要实现XML响应解析，返回Object列表
	// 由于复杂度较高，这里暂时返回空列表
	return []Object{}, nil
}

// CopyObject 复制对象
func (s *TencentCOS) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	// 构建请求URL
	copyURL := fmt.Sprintf("%s/%s", s.bucketURL, dstPath)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "PUT", copyURL, nil)
	if err != nil {
		logger.Error("Failed to create copy request",
			zap.String("srcPath", srcPath),
			zap.String("dstPath", dstPath),
			zap.Error(err),
		)
		return errors.Wrap(errors.ErrInternal, "创建复制请求失败", err)
	}

	// 设置复制源
	srcObject := fmt.Sprintf("%s/%s", s.bucketName, srcPath)
	req.Header.Set("x-cos-copy-source", srcObject)

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to copy object",
			zap.String("srcPath", srcPath),
			zap.String("dstPath", dstPath),
			zap.Error(err),
		)
		return errors.Wrap(errors.ErrInternal, "复制对象失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to copy object",
			zap.String("srcPath", srcPath),
			zap.String("dstPath", dstPath),
			zap.Int("statusCode", resp.StatusCode),
		)
		return errors.Wrap(errors.ErrInternal, fmt.Sprintf("复制对象失败，状态码: %d", resp.StatusCode), nil)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *TencentCOS) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	// 构建请求URL
	infoURL := fmt.Sprintf("%s/%s", s.bucketURL, path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "HEAD", infoURL, nil)
	if err != nil {
		logger.Error("Failed to create get object info request",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrInternal, "创建获取对象信息请求失败", err)
	}

	// 签名请求
	s.signRequest(req)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to get object info",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.ErrResourceNotFound, "获取对象信息失败", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to get object info",
			zap.String("path", path),
			zap.Int("statusCode", resp.StatusCode),
		)
		return nil, errors.Wrap(errors.ErrResourceNotFound, fmt.Sprintf("获取对象信息失败，状态码: %d", resp.StatusCode), nil)
	}

	// 构建对象信息
	size := int64(0)
	if sizeStr := resp.Header.Get("Content-Length"); sizeStr != "" {
		fmt.Sscanf(sizeStr, "%d", &size)
	}

	lastModified := ""
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		lastModified = lm
	}

	etag := ""
	if et := resp.Header.Get("ETag"); et != "" {
		etag = et
	}

	contentType := ""
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		contentType = ct
	}

	// 提取元数据
	metadata := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			metadata[key] = values[0]
		}
	}

	return &ObjectInfo{
		Key:          path,
		Size:         size,
		ContentType:  contentType,
		LastModified: lastModified,
		ETag:         etag,
		Metadata:     metadata,
	}, nil
}

// signRequest 签名请求
func (s *TencentCOS) signRequest(req *http.Request) {
	// 使用腾讯云COS官方SDK提供的签名方法
	// 参考文档: https://cloud.tencent.com/document/product/436/7778
	// 使用 AddAuthorizationHeader 函数进行签名

	// 添加授权头
	authTime := cos.NewAuthTime(5 * time.Minute)
	cos.AddAuthorizationHeader(s.secretID, s.secretKey, "", req, authTime)
}
