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
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
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
	thumbVideoExts = map[string]bool{".mp4": true, ".mov": true, ".m4v": true, ".avi": true, ".wmv": true, ".flv": true, ".mkv": true, ".webm": true, ".mpg": true, ".mpeg": true, ".ts": true, ".3gp": true, ".m3u8": true, ".hls": true}
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
		// 公司存储（COS 等）中已有缩略图：取回并写入本机缓存（服务迁移、多实例时不必重新生成）
		fetchCtx, cancelFetch := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFetch()
		store, remoteKey := s.thumbnailStore(fetchCtx, file, cachePath)
		if store != nil {
			if data, err := readThumbnail(fetchCtx, store, remoteKey); err == nil {
				writeThumbnailCache(cachePath, data)
				return data, nil
			}
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
		writeThumbnailCache(cachePath, data)
		if store != nil {
			go saveThumbnail(store, remoteKey, data, fileID)
		}
		return data, nil
	})
	if err != nil {
		logger.Warn("生成缩略图失败", zap.Uint("fileID", fileID), zap.String("error", signedQuery.ReplaceAllString(err.Error(), "?***")))
		return nil, errors.Wrap(errors.ErrResourceNotFound, "缩略图不可用", err)
	}
	return result.([]byte), nil
}

// thumbnailStore 缩略图同时保存到的公司存储（文件所有者所在公司）。
// 本地存储与本机缓存在同一磁盘，不重复保存，返回 nil
func (s *service) thumbnailStore(ctx context.Context, file *entity.CloudItem, cachePath string) (storage.Storage, string) {
	st, err := s.getCompanyStorage(ctx, file.OwnerID)
	if err != nil || st == nil {
		return nil, ""
	}
	if local, ok := st.(interface{ IsLocalStorage() bool }); ok && local.IsLocalStorage() {
		return nil, ""
	}
	rel, err := filepath.Rel(ThumbnailCacheDir, cachePath)
	if err != nil {
		return nil, ""
	}
	return st, "thumbnails/" + filepath.ToSlash(rel)
}

func readThumbnail(ctx context.Context, st storage.Storage, key string) ([]byte, error) {
	rc, err := st.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
	if err != nil {
		return nil, err
	}
	if len(data) < 3 || data[0] != 0xFF || data[1] != 0xD8 {
		return nil, fmt.Errorf("不是有效的缩略图")
	}
	return data, nil
}

// saveThumbnail 后台保存缩略图到公司存储，失败只记日志（本机缓存仍可用）
func saveThumbnail(st storage.Storage, key string, data []byte, fileID uint) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := st.Upload(ctx, key, bytes.NewReader(data), "image/jpeg"); err != nil {
		logger.Warn("缩略图保存到公司存储失败", zap.Uint("fileID", fileID), zap.String("key", key), zap.Error(err))
	}
}

func writeThumbnailCache(cachePath string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err == nil {
		tmp := cachePath + ".tmp"
		if os.WriteFile(tmp, data, 0644) == nil {
			_ = os.Rename(tmp, cachePath)
		}
	}
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
	ffmpegOnce.Do(func() { ffmpegPath = findFFmpeg() })
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
		var signed string
		if signed, err = v.PresignedDownloadURL(ctx, path, 10*time.Minute); err == nil {
			// 经本机代理读取：静态编译的 ffmpeg 解析域名会崩溃，且签名地址不出现在进程参数中
			var stop func()
			input, stop, err = startRangeProxy(signed)
			if err == nil {
				defer stop()
			}
		}
	default:
		err = fmt.Errorf("存储不支持读取视频")
	}
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		// 只取第一帧：缩短流分析，避免为探测格式读取数秒的视频数据
		"-analyzeduration", "100000", "-probesize", "1000000",
		// 单线程解码：多线程解码要先缓冲多帧才输出第一帧，会多读取数帧数据
		"-threads", "1",
		"-i", input,
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale='min(%d,iw)':-2", thumbnailWidth),
		"-f", "image2", "-c:v", "mjpeg", "-q:v", "5",
		"pipe:1",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// ffmpeg 会走 http_proxy 环境变量，本机代理地址须直连
	cmd.Env = append(os.Environ(), "no_proxy=127.0.0.1,localhost", "NO_PROXY=127.0.0.1,localhost")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("截取视频第一帧失败: %v %s", err, strings.TrimSpace(stderr.String()))
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("截取视频第一帧失败: 无输出")
	}
	return stdout.Bytes(), nil
}

// findFFmpeg 查找 ffmpeg：环境变量 FFMPEG_PATH > PATH > 用户目录 ~/bin/ffmpeg（无管理员权限时的安装位置）
func findFFmpeg() string {
	if p := os.Getenv("FFMPEG_PATH"); p != "" {
		return p
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "bin", "ffmpeg")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// rangeProxyClient 代理请求云存储的客户端（支持 Range，单次请求最长 1 分钟）
var rangeProxyClient = &http.Client{Timeout: time.Minute}

// proxyWindow 代理单次响应的最大字节数
const proxyWindow = 2 << 20

// capRange 把客户端的 Range 限制在 proxyWindow 以内（无 Range 视为从头读取）
func capRange(header string) string {
	var start, end int64 = 0, -1
	if strings.HasPrefix(header, "bytes=") && !strings.Contains(header, ",") {
		parts := strings.SplitN(strings.TrimPrefix(header, "bytes="), "-", 2)
		if len(parts) == 2 && parts[0] != "" {
			if v, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				start = v
			}
			if v, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				end = v
			}
		} else if len(parts) == 2 && parts[0] == "" {
			return header // 后缀 Range（bytes=-N，读取末尾，如 moov 在文件尾）原样转发
		}
	}
	if limit := start + proxyWindow - 1; end < 0 || end > limit {
		end = limit
	}
	return fmt.Sprintf("bytes=%d-%d", start, end)
}

// startRangeProxy 在 127.0.0.1 启动临时代理，把请求（含 Range）转发到 target，返回本地地址与关闭函数。
// 本地地址沿用 target 的路径：HLS 播放列表中的相对分片地址解析到同一目录，由代理转发到源站同目录
func startRangeProxy(target string) (string, func(), error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return "", nil, err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upstream := *targetURL
			if r.URL.Path != targetURL.Path {
				// 分片等其他文件：同一源站下的对应路径（带其自身的查询参数）
				upstream.Path, upstream.RawPath, upstream.RawQuery = r.URL.Path, r.URL.RawPath, r.URL.RawQuery
			}
			req, err := http.NewRequestWithContext(r.Context(), r.Method, upstream.String(), nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			// 每次最多返回 proxyWindow 字节，ffmpeg 需要更多时会按偏移继续请求，
			// 避免开放式 Range（bytes=0-）让代理在 ffmpeg 退出前持续从云存储拉取数据
			req.Header.Set("Range", capRange(r.Header.Get("Range")))
			resp, err := rangeProxyClient.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()
			for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Last-Modified", "ETag"} {
				if v := resp.Header.Get(h); v != "" {
					w.Header().Set(h, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		}),
	}
	go func() { _ = srv.Serve(ln) }()
	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	localPath := targetURL.EscapedPath()
	if localPath == "" {
		localPath = "/"
	}
	return "http://" + ln.Addr().String() + localPath, stop, nil
}
