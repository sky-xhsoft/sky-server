package cloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"go.uber.org/zap"
)

// ==================== 媒体在线播放 ====================

// playURLExpire 媒体直链有效期（覆盖较长视频的观看时长）
const playURLExpire = 6 * time.Hour

// mediaPresigner 能签发带签名下载直链的存储（腾讯云 COS、外部地址）
type mediaPresigner interface {
	PresignedDownloadURL(ctx context.Context, path string, expire time.Duration) (string, error)
}

// GetPlayURL 文件直链（视频 / 音频边下边播、图片显示、文档预览读取）：云存储返回带签名的直链，
// 浏览器直接从云存储读取，不经过应用服务器；本地存储没有直链，返回空字符串（前端改为经服务器读取）
func (s *service) GetPlayURL(ctx context.Context, fileID uint, userID uint) (string, error) {
	file, err := s.readableFile(ctx, fileID, userID)
	if err != nil {
		return "", err
	}
	st, err := s.getItemStorage(ctx, file)
	if err != nil {
		return "", err
	}
	if p, ok := st.(mediaPresigner); ok {
		return p.PresignedDownloadURL(ctx, *file.StoragePath, playURLExpire)
	}
	return "", nil
}

// readableFile 查询文件并校验读权限（与下载一致）
func (s *service) readableFile(ctx context.Context, fileID uint, userID uint) (*entity.CloudItem, error) {
	file, err := s.getFileByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if !s.canAccessRead(ctx, file, userID, s.getUserCompanyID(ctx, userID)) {
		return nil, errors.New(errors.ErrPermissionDenied, "无权限查看此文件")
	}
	if file.StoragePath == nil || *file.StoragePath == "" {
		return nil, errors.New(errors.ErrResourceNotFound, "文件存储路径不存在")
	}
	return file, nil
}

// ==================== 旧版 Office 文档转 PDF 预览 ====================

const (
	// previewPDFVersion 转换规则变化时递增，使旧缓存失效
	previewPDFVersion = "v1"
	// previewMaxSourceSize 超过该大小的文档不转换
	previewMaxSourceSize = 50 << 20
	// previewMaxPDFSize 读取缓存 PDF 的上限
	previewMaxPDFSize = 200 << 20
)

// PreviewCacheDir 转换后 PDF 的本机缓存目录
var PreviewCacheDir = filepath.Join("data", "previews")

// previewPDFExts 浏览器内无法保留排版、需在服务器转换为 PDF 的旧版二进制文档
var previewPDFExts = map[string]bool{".doc": true, ".dot": true, ".wps": true}

var (
	previewGroup singleflight.Group
	// 同时只转换 1 个文档（LibreOffice 占 CPU 与内存）
	previewSlots = make(chan struct{}, 1)
	sofficeMu    sync.Mutex
	sofficePath  string
)

// currentSoffice 找到后缓存；未找到时下次重新查找（服务运行期间安装 LibreOffice 后无需重启）
func currentSoffice() string {
	sofficeMu.Lock()
	defer sofficeMu.Unlock()
	if sofficePath == "" {
		sofficePath = findSoffice()
	}
	return sofficePath
}

// NeedsPDFPreview 文件是否需要转换为 PDF 预览
func NeedsPDFPreview(item *entity.CloudItem) bool {
	return previewPDFExts[itemExt(item)]
}

// itemExt 文件扩展名（小写，带点）
func itemExt(item *entity.CloudItem) string {
	if item.FileExt != nil && *item.FileExt != "" {
		ext := strings.ToLower(*item.FileExt)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		return ext
	}
	return strings.ToLower(filepath.Ext(item.Name))
}

// GetPreviewPDF 旧版 Word（.doc / .dot / .wps）转换为 PDF 用于预览，保留排版。
// 转换结果缓存在本机磁盘并保存到公司存储（previews/ 目录），每个文件只转换一次
func (s *service) GetPreviewPDF(ctx context.Context, fileID uint, userID uint) ([]byte, error) {
	file, err := s.readableFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	if !NeedsPDFPreview(file) {
		return nil, errors.New(errors.ErrInvalidParam, "该文件不需要转换预览")
	}
	if file.FileSize != nil && *file.FileSize > previewMaxSourceSize {
		return nil, errors.New(errors.ErrResourceNotFound, "文档过大，请下载后查看")
	}

	cachePath := previewCachePath(file)
	if data, err := readPDFFile(cachePath); err == nil {
		return data, nil
	}
	result, err, _ := previewGroup.Do(cachePath, func() (interface{}, error) {
		if data, err := readPDFFile(cachePath); err == nil {
			return data, nil
		}
		fetchCtx, cancelFetch := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancelFetch()
		store, remoteKey := s.previewStore(fetchCtx, file, cachePath)
		if store != nil {
			if data, err := readPDFFromStore(fetchCtx, store, remoteKey); err == nil {
				writeThumbnailCache(cachePath, data)
				return data, nil
			}
		}

		previewSlots <- struct{}{}
		defer func() { <-previewSlots }()
		// 与请求解耦：浏览器取消请求不打断转换，结果供后续请求复用
		genCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		data, err := s.convertToPDF(genCtx, file)
		if err != nil {
			return nil, err
		}
		writeThumbnailCache(cachePath, data)
		if store != nil {
			go savePreviewPDF(store, remoteKey, data, fileID)
		}
		return data, nil
	})
	if err != nil {
		logger.Warn("文档转换预览失败", zap.Uint("fileID", fileID), zap.String("error", signedQuery.ReplaceAllString(err.Error(), "?***")))
		return nil, errors.Wrap(errors.ErrResourceNotFound, "文档预览不可用", err)
	}
	return result.([]byte), nil
}

// previewCachePath 按文件内容缓存（与缩略图相同的键）
func previewCachePath(file *entity.CloudItem) string {
	thumb := thumbnailCachePath(file)
	key := strings.TrimSuffix(filepath.Base(thumb), ".jpg")
	key = previewPDFVersion + key[len(thumbnailVersion):]
	return filepath.Join(PreviewCacheDir, filepath.Base(filepath.Dir(thumb)), key+".pdf")
}

// previewStore 转换结果同时保存到的公司存储；本地存储与本机缓存在同一磁盘，返回 nil
func (s *service) previewStore(ctx context.Context, file *entity.CloudItem, cachePath string) (storage.Storage, string) {
	st, err := s.getCompanyStorage(ctx, file.OwnerID)
	if err != nil || st == nil {
		return nil, ""
	}
	if local, ok := st.(interface{ IsLocalStorage() bool }); ok && local.IsLocalStorage() {
		return nil, ""
	}
	rel, err := filepath.Rel(PreviewCacheDir, cachePath)
	if err != nil {
		return nil, ""
	}
	return st, "previews/" + filepath.ToSlash(rel)
}

func isPDF(data []byte) bool { return bytes.HasPrefix(data, []byte("%PDF")) }

func readPDFFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !isPDF(data) {
		return nil, fmt.Errorf("缓存不是有效的 PDF")
	}
	return data, nil
}

func readPDFFromStore(ctx context.Context, st storage.Storage, key string) ([]byte, error) {
	if ok, err := st.Exists(ctx, key); err != nil || !ok {
		return nil, fmt.Errorf("公司存储中没有转换结果")
	}
	rc, err := st.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, previewMaxPDFSize))
	if err != nil {
		return nil, err
	}
	if !isPDF(data) {
		return nil, fmt.Errorf("不是有效的 PDF")
	}
	return data, nil
}

// savePreviewPDF 后台保存转换结果到公司存储，失败只记日志（本机缓存仍可用）
func savePreviewPDF(st storage.Storage, key string, data []byte, fileID uint) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := st.Upload(ctx, key, bytes.NewReader(data), "application/pdf"); err != nil {
		logger.Warn("预览 PDF 保存到公司存储失败", zap.Uint("fileID", fileID), zap.String("key", key), zap.Error(err))
	}
}

// convertToPDF 取出原文件，用 LibreOffice 无界面转换为 PDF
func (s *service) convertToPDF(ctx context.Context, file *entity.CloudItem) ([]byte, error) {
	soffice := currentSoffice()
	if soffice == "" {
		return nil, fmt.Errorf("服务器未安装 LibreOffice")
	}
	st, err := s.getItemStorage(ctx, file)
	if err != nil {
		return nil, err
	}
	rc, err := st.Download(ctx, *file.StoragePath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	dir, err := os.MkdirTemp("", "sky-preview-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	src := filepath.Join(dir, "source"+itemExt(file))
	out, err := os.Create(src)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(out, io.LimitReader(rc, previewMaxSourceSize+1))
	out.Close()
	if err != nil {
		return nil, err
	}
	if n > previewMaxSourceSize {
		return nil, fmt.Errorf("文档过大")
	}
	return runSoffice(ctx, soffice, src, dir)
}

// runSoffice 把 src 转为 PDF（输出到 dir）。每次使用独立的用户配置目录，避免并发或残留进程的锁冲突
func runSoffice(ctx context.Context, soffice, src, dir string) ([]byte, error) {
	profile := filepath.Join(dir, "profile")
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, soffice,
		"--headless", "--norestore", "--nologo", "--nolockcheck", "--nodefault",
		"-env:UserInstallation="+fileURL(profile),
		"--convert-to", "pdf", "--outdir", dir, src,
	)
	cmd.Stdout = &stderr
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "HOME="+dir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("LibreOffice 转换失败: %v %s", err, strings.TrimSpace(stderr.String()))
	}
	pdf := strings.TrimSuffix(src, filepath.Ext(src)) + ".pdf"
	data, err := readPDFFile(pdf)
	if err != nil {
		return nil, fmt.Errorf("LibreOffice 未生成 PDF: %s", strings.TrimSpace(stderr.String()))
	}
	return data, nil
}

// fileURL 本地路径转为 file:// 地址（兼容 Windows 盘符路径）
func fileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

// findSoffice 查找 LibreOffice：环境变量 SOFFICE_PATH > PATH（soffice / libreoffice）> 用户目录 ~/bin/soffice
func findSoffice() string {
	if p := os.Getenv("SOFFICE_PATH"); p != "" {
		return p
	}
	for _, name := range []string{"soffice", "libreoffice"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "bin", "soffice")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
