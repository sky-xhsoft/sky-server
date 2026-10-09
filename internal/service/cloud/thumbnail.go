package cloud

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	// 注册图片解码器
	_ "image/gif"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"golang.org/x/sync/singleflight"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	// thumbnailWidth 缩略图最大宽度（卡片 2 倍图）
	thumbnailWidth = 360
	// thumbnailMaxImageSize 超过该大小的图片不生成缩略图（避免解码超大图占用内存）
	thumbnailMaxImageSize = 40 << 20
	// thumbnailVersion 生成规则变化时递增，使旧缓存失效
	thumbnailVersion = "v1"
)

// ThumbnailCacheDir 缩略图缓存目录（服务端本地磁盘，不写入用户存储桶）
var ThumbnailCacheDir = filepath.Join("data", "thumbnails")

// signedQuery 日志中隐去签名地址的查询参数（含临时签名）
var signedQuery = regexp.MustCompile(`\?[^\s"']+`)

var (
	thumbnailGroup singleflight.Group
	// 同时最多生成 2 个缩略图（ffmpeg 解码占 CPU 与内存）
	thumbnailSlots = make(chan struct{}, 2)
	ffmpegOnce     sync.Once
	ffmpegPath     string
)

var (
	thumbImageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true}
	thumbVideoExts = map[string]bool{".mp4": true, ".mov": true, ".m4v": true, ".avi": true, ".wmv": true, ".flv": true, ".mkv": true, ".webm": true, ".mpg": true, ".mpeg": true, ".ts": true, ".3gp": true}
)

// thumbnailKind 返回 image / video，不支持缩略图时返回空
func thumbnailKind(item *entity.CloudItem) string {
	ext := strings.ToLower(filepath.Ext(item.Name))
	if item.FileExt != nil && *item.FileExt != "" {
		ext = strings.ToLower(*item.FileExt)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
	}
	mime := ""
	if item.FileType != nil {
		mime = strings.ToLower(*item.FileType)
	}
	switch {
	case thumbImageExts[ext] || (ext == "" && strings.HasPrefix(mime, "image/")):
		return "image"
	case thumbVideoExts[ext] || (ext == "" && strings.HasPrefix(mime, "video/")):
		return "video"
	}
	return ""
}

// GetThumbnail 返回文件缩略图（JPEG）：图片为缩小图，视频为第一帧。生成后缓存在服务端磁盘
func (s *service) GetThumbnail(ctx context.Context, fileID uint, userID uint) ([]byte, error) {
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
	kind := thumbnailKind(file)
	if kind == "" {
		return nil, errors.New(errors.ErrResourceNotFound, "该文件类型没有缩略图")
	}
	if kind == "image" && file.FileSize != nil && *file.FileSize > thumbnailMaxImageSize {
		return nil, errors.New(errors.ErrResourceNotFound, "图片过大，不生成缩略图")
	}

	cachePath := thumbnailCachePath(file)
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	result, err, _ := thumbnailGroup.Do(cachePath, func() (interface{}, error) {
		if data, err := os.ReadFile(cachePath); err == nil {
			return data, nil
		}
		thumbnailSlots <- struct{}{}
		defer func() { <-thumbnailSlots }()

		// 与请求解耦：浏览器取消请求不打断生成，生成结果供后续请求复用
		genCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var data []byte
		var genErr error
		if kind == "image" {
			data, genErr = s.imageThumbnail(genCtx, file)
		} else {
			data, genErr = s.videoThumbnail(genCtx, file)
		}
		if genErr != nil {
			return nil, genErr
		}
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err == nil {
			tmp := cachePath + ".tmp"
			if os.WriteFile(tmp, data, 0644) == nil {
				_ = os.Rename(tmp, cachePath)
			}
		}
		return data, nil
	})
	if err != nil {
		logger.Warn("生成缩略图失败", zap.Uint("fileID", fileID), zap.String("error", signedQuery.ReplaceAllString(err.Error(), "?***")))
		return nil, errors.Wrap(errors.ErrResourceNotFound, "缩略图不可用", err)
	}
	return result.([]byte), nil
}

// thumbnailCachePath 按文件内容（MD5）缓存：相同内容只生成一次；无 MD5 时按存储路径
func thumbnailCachePath(file *entity.CloudItem) string {
	key := ""
	if file.MD5 != nil && *file.MD5 != "" {
		key = *file.MD5
	} else {
		sum := sha1.Sum([]byte(*file.StoragePath))
		key = hex.EncodeToString(sum[:])
	}
	key = thumbnailVersion + "_" + key
	return filepath.Join(ThumbnailCacheDir, key[len(thumbnailVersion)+1:len(thumbnailVersion)+3], key+".jpg")
}

// imageThumbnail 解码图片并缩放为 JPEG
func (s *service) imageThumbnail(ctx context.Context, file *entity.CloudItem) ([]byte, error) {
	st, err := s.getItemStorage(ctx, file)
	if err != nil {
		return nil, err
	}
	reader, err := st.Download(ctx, *file.StoragePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	src, _, err := image.Decode(io.LimitReader(reader, thumbnailMaxImageSize))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}
	return encodeThumbnail(src)
}

// encodeThumbnail 等比缩放到最大宽度并编码为 JPEG（透明背景铺白）
func encodeThumbnail(src image.Image) ([]byte, error) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("图片尺寸无效")
	}
	if w > thumbnailWidth {
		h = h * thumbnailWidth / w
		w = thumbnailWidth
		if h < 1 {
			h = 1
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// videoThumbnail 用 ffmpeg 截取视频第一帧。本地文件直接读磁盘，云存储通过签名地址按需读取（不下载整个文件）
func (s *service) videoThumbnail(ctx context.Context, file *entity.CloudItem) ([]byte, error) {
	st, err := s.getItemStorage(ctx, file)
	if err != nil {
		return nil, err
	}
	return s.videoThumbnailFrom(ctx, st, *file.StoragePath)
}

// videoThumbnailFrom 从存储中的视频截取第一帧
func (s *service) videoThumbnailFrom(ctx context.Context, st interface{}, path string) ([]byte, error) {
	ffmpegOnce.Do(func() {
		ffmpegPath, _ = exec.LookPath("ffmpeg")
	})
	if ffmpegPath == "" {
		return nil, fmt.Errorf("服务器未安装 ffmpeg")
	}
	var input string
	var err error
	switch v := st.(type) {
	case interface{ LocalPath(string) (string, error) }:
		input, err = v.LocalPath(path)
	case interface {
		PresignedDownloadURL(context.Context, string, time.Duration) (string, error)
	}:
		input, err = v.PresignedDownloadURL(ctx, path, 10*time.Minute)
	default:
		err = fmt.Errorf("存储不支持读取视频")
	}
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-i", input,
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':-2", thumbnailWidth),
		"-f", "image2", "-c:v", "mjpeg", "-q:v", "5",
		"pipe:1",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("截取视频第一帧失败: %v %s", err, strings.TrimSpace(stderr.String()))
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("截取视频第一帧失败: 无输出")
	}
	return stdout.Bytes(), nil
}
