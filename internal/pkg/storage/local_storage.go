package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath string // 基础路径
	baseURL  string // 基础URL
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	BasePath string // 基础路径（如: ./uploads）
	BaseURL  string // 基础URL（如: http://localhost:8080/files）
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(cfg *LocalStorageConfig) (Storage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	return &LocalStorage{
		basePath: cfg.BasePath,
		baseURL:  cfg.BaseURL,
	}, nil
}

// Upload 上传文件
func (s *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) (string, error) {
	// 完整文件路径
	fullPath := filepath.Join(s.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.Wrap(errors.ErrInternal, "创建目录失败", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "创建文件失败", err)
	}
	defer file.Close()

	// 写入内容
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath) // 删除已创建的文件
		return "", errors.Wrap(errors.ErrInternal, "写入文件失败", err)
	}

	// 返回访问URL
	url := fmt.Sprintf("%s/%s", s.baseURL, path)
	return url, nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrResourceNotFound, "文件不存在")
		}
		return nil, errors.Wrap(errors.ErrInternal, "打开文件失败", err)
	}

	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为删除成功
		}
		return errors.Wrap(errors.ErrInternal, "删除文件失败", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(errors.ErrInternal, "检查文件失败", err)
	}

	return true, nil
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(ctx context.Context, path string, expireSeconds int) (string, error) {
	// 本地存储不支持过期URL，直接返回永久URL
	url := fmt.Sprintf("%s/%s", s.baseURL, path)
	return url, nil
}

// ListObjects 列出对象
func (s *LocalStorage) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]Object, error) {
	fullPrefix := filepath.Join(s.basePath, prefix)

	var objects []Object
	count := 0

	err := filepath.Walk(fullPrefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查最大数量
		if maxKeys > 0 && count >= maxKeys {
			return filepath.SkipDir
		}

		// 计算相对路径
		relPath, err := filepath.Rel(s.basePath, path)
		if err != nil {
			return err
		}

		// 添加对象
		objects = append(objects, Object{
			Key:          strings.ReplaceAll(relPath, "\\", "/"), // Windows路径转换
			Size:         info.Size(),
			LastModified: info.ModTime().Format(time.RFC3339),
			ETag:         "", // 本地存储不生成ETag
		})

		count++
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "列举对象失败", err)
	}

	return objects, nil
}

// CopyObject 复制对象
func (s *LocalStorage) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := filepath.Join(s.basePath, srcPath)
	dstFullPath := filepath.Join(s.basePath, dstPath)

	// 打开源文件
	srcFile, err := os.Open(srcFullPath)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "打开源文件失败", err)
	}
	defer srcFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrap(errors.ErrInternal, "创建目标目录失败", err)
	}

	// 创建目标文件
	dstFile, err := os.Create(dstFullPath)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "创建目标文件失败", err)
	}
	defer dstFile.Close()

	// 复制内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dstFullPath) // 删除已创建的文件
		return errors.Wrap(errors.ErrInternal, "复制文件失败", err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *LocalStorage) GetObjectInfo(ctx context.Context, path string) (*ObjectInfo, error) {
	fullPath := filepath.Join(s.basePath, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrResourceNotFound, "文件不存在")
		}
		return nil, errors.Wrap(errors.ErrInternal, "获取文件信息失败", err)
	}

	return &ObjectInfo{
		Key:          path,
		Size:         info.Size(),
		ContentType:  "", // 本地存储不存储ContentType
		LastModified: info.ModTime().Format(time.RFC3339),
		ETag:         "",
		Metadata:     make(map[string]string),
	}, nil
}
