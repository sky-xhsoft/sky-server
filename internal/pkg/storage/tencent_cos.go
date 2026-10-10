package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// IsTencentCOS 判断是否是腾讯云COS存储（用于类型断言）
func (s *TencentCOS) IsTencentCOS() bool {
	return true
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

	// 增加超时时间以支持大文件上传
	// 对于7GB+文件，需要更长超时时间，按10MB/s估算需要12分钟，留15分钟余量
	timeout := 15 * time.Minute // 15分钟超时，足够支持10GB+大文件上传

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

// PresignedUploadURL 获取预签名上传URL（用于前端直传）。签名放在 URL 查询参数中，浏览器直接 PUT 即可
func (s *TencentCOS) PresignedUploadURL(ctx context.Context, path string, expireSeconds int, contentType string) (string, error) {
	return s.presignedPutURL(ctx, path, nil, expireSeconds)
}

// presignedPutURL 生成带签名的 PUT 地址；query 中的参数（如 uploadId、partNumber）一并签入
func (s *TencentCOS) presignedPutURL(ctx context.Context, path string, query *url.Values, expireSeconds int) (string, error) {
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, s.httpClient)
	expire := time.Duration(expireSeconds) * time.Second
	presigned, err := client.Object.GetPresignedURL(ctx, http.MethodPut, path, s.secretID, s.secretKey, expire,
		&cos.PresignedURLOptions{Query: query})
	if err != nil {
		logger.Error("Failed to create presigned upload url", zap.String("path", path), zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "生成上传地址失败", err)
	}
	return presigned.String(), nil
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

// InitiateMultipartUpload 初始化分块上传
// 返回 uploadID
func (s *TencentCOS) InitiateMultipartUpload(ctx context.Context, path string) (string, error) {
	// 解析 bucket URL
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		logger.Error("Failed to parse bucket URL", zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}

	// 创建 cos baseURL
	base := &cos.BaseURL{
		BucketURL: bucketURL,
	}

	// 创建 cos client
	client := cos.NewClient(base, s.httpClient)

	// 初始化分块上传
	result, _, err := client.Object.InitiateMultipartUpload(ctx, path, nil)
	if err != nil {
		logger.Error("Failed to initiate multipart upload",
			zap.String("path", path),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "初始化分块上传失败", err)
	}

	return result.UploadID, nil
}

// PresignedChunkUploadURL 获取原生分块的预签名上传地址（chunkNumber 从 0 开始，COS 分块编号从 1 开始）。
// 浏览器用该地址直接 PUT 分块到 COS，不经过应用服务器
func (s *TencentCOS) PresignedChunkUploadURL(ctx context.Context, path string, uploadID string, chunkNumber int, expireSeconds int, contentType string) (string, error) {
	query := url.Values{}
	query.Set("partNumber", strconv.Itoa(chunkNumber+1))
	query.Set("uploadId", uploadID)
	return s.presignedPutURL(ctx, path, &query, expireSeconds)
}

// UploadedPart COS 端已上传的分块
type UploadedPart struct {
	PartNumber int // 从 1 开始
	ETag       string
	Size       int64
}

// ListUploadedParts 列出原生分块上传中 COS 端实际已收到的全部分块（自动翻页）
func (s *TencentCOS) ListUploadedParts(ctx context.Context, path string, uploadID string) ([]UploadedPart, error) {
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, s.httpClient)
	var parts []UploadedPart
	opt := &cos.ObjectListPartsOptions{MaxParts: "1000"}
	for {
		result, _, err := client.Object.ListParts(ctx, path, uploadID, opt)
		if err != nil {
			return nil, errors.Wrap(errors.ErrStorage, "查询已上传分块失败", err)
		}
		for _, p := range result.Parts {
			parts = append(parts, UploadedPart{PartNumber: p.PartNumber, ETag: strings.Trim(p.ETag, `"`), Size: p.Size})
		}
		if !result.IsTruncated || result.NextPartNumberMarker == "" {
			return parts, nil
		}
		opt.PartNumberMarker = result.NextPartNumberMarker
	}
}

// UploadPart 上传一个原生分块（chunkIndex 从 0 开始），返回 COS 的 ETag。
// 服务端中转上传时直接写入原生分块，完成时由 COS 合并，无需再下载分片重新上传
func (s *TencentCOS) UploadPart(ctx context.Context, path string, uploadID string, chunkIndex int, data []byte) (string, error) {
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, s.httpClient)

	sum := md5.Sum(data)
	resp, err := client.Object.UploadPart(ctx, path, uploadID, chunkIndex+1, bytes.NewReader(data), &cos.ObjectUploadPartOptions{
		ContentLength: int64(len(data)),
		ContentMD5:    base64.StdEncoding.EncodeToString(sum[:]),
	})
	if err != nil {
		logger.Error("Failed to upload part",
			zap.String("path", path),
			zap.String("uploadID", uploadID),
			zap.Int("chunkIndex", chunkIndex),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "上传分块失败", err)
	}
	return strings.Trim(resp.Header.Get("ETag"), `"`), nil
}

// PresignedDownloadURL 生成带签名的下载地址（私有桶可读），支持 Range 读取部分内容
func (s *TencentCOS) PresignedDownloadURL(ctx context.Context, path string, expire time.Duration) (string, error) {
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, s.httpClient)
	presigned, err := client.Object.GetPresignedURL(ctx, http.MethodGet, path, s.secretID, s.secretKey, expire, nil)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "生成下载地址失败", err)
	}
	return presigned.String(), nil
}

// AbortMultipartUpload 终止原生分块上传，释放已上传但未合并的分块
func (s *TencentCOS) AbortMultipartUpload(ctx context.Context, path string, uploadID string) error {
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, s.httpClient)
	if _, err := client.Object.AbortMultipartUpload(ctx, path, uploadID); err != nil {
		return errors.Wrap(errors.ErrInternal, "终止分块上传失败", err)
	}
	return nil
}

// CompleteMultipartUpload 完成分块上传，让服务端合并所有分块
func (s *TencentCOS) CompleteMultipartUpload(ctx context.Context, path string, uploadID string, partETags []string) (string, error) {
	// 解析 bucket URL
	bucketURL, err := url.Parse(s.bucketURL)
	if err != nil {
		logger.Error("Failed to parse bucket URL", zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "解析Bucket URL失败", err)
	}

	// 创建 cos baseURL
	base := &cos.BaseURL{
		BucketURL: bucketURL,
	}

	// 创建 cos client
	client := cos.NewClient(base, s.httpClient)

	// 构造 parts
	var parts []cos.Object
	for i, etag := range partETags {
		parts = append(parts, cos.Object{
			PartNumber: i + 1, // COS 分片编号从 1 开始
			ETag:       etag,
		})
	}

	// 构造 options
	opt := &cos.CompleteMultipartUploadOptions{
		Parts: parts,
	}

	// 完成分块上传
	result, _, err := client.Object.CompleteMultipartUpload(ctx, path, uploadID, opt)
	if err != nil {
		logger.Error("Failed to complete multipart upload",
			zap.String("path", path),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return "", errors.Wrap(errors.ErrInternal, "完成分块上传失败", err)
	}

	// 返回最终文件的ETag
	return result.ETag, nil
}
