package cloud

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
)

func strPtr(s string) *string { return &s }

func TestThumbnailKind(t *testing.T) {
	cases := []struct {
		name, ext, mime, want string
	}{
		{"a.MP4", "", "", "video"},
		{"x", ".mov", "", "video"},
		{"x", "jpg", "", "image"},
		{"图.webp", "", "", "image"},
		{"blob", "", "video/quicktime", "video"},
		{"文档.pdf", ".pdf", "application/pdf", ""},
		{"包.zip", "", "", ""},
	}
	for _, c := range cases {
		item := &entity.CloudItem{Name: c.name}
		if c.ext != "" {
			item.FileExt = strPtr(c.ext)
		}
		if c.mime != "" {
			item.FileType = strPtr(c.mime)
		}
		if got := thumbnailKind(item); got != c.want {
			t.Errorf("%s/%s/%s: got %q want %q", c.name, c.ext, c.mime, got, c.want)
		}
	}
}

func TestEncodeThumbnail_ScalesDownKeepingRatio(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	for y := 0; y < 800; y++ {
		for x := 0; x < 1200; x++ {
			src.Set(x, y, color.RGBA{uint8(x), uint8(y), 120, 255})
		}
	}
	data, err := encodeThumbnail(src)
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != thumbnailWidth || b.Dy() != 240 {
		t.Fatalf("尺寸 = %dx%d，期望 %dx240", b.Dx(), b.Dy(), thumbnailWidth)
	}

	// 小图不放大
	small, _ := encodeThumbnail(image.NewRGBA(image.Rect(0, 0, 100, 50)))
	img, _ = jpeg.Decode(bytes.NewReader(small))
	if b := img.Bounds(); b.Dx() != 100 || b.Dy() != 50 {
		t.Fatalf("小图尺寸 = %dx%d", b.Dx(), b.Dy())
	}
}

// 本地存储的视频：ffmpeg 截取第一帧（未安装 ffmpeg 时跳过）
func TestVideoThumbnail_LocalFile(t *testing.T) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("未安装 ffmpeg")
	}
	dir := t.TempDir()
	video := filepath.Join(dir, "v.mp4")
	if out, err := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=640x360:rate=25",
		"-t", "1", "-pix_fmt", "yuv420p", video).CombinedOutput(); err != nil {
		t.Fatalf("生成测试视频失败: %v %s", err, out)
	}

	s := &service{}
	st := fakeLocalPath{path: video}
	data, err := s.videoThumbnailFrom(context.Background(), st, "v.mp4")
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("输出不是 JPEG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != thumbnailWidth || b.Dy() != 202 {
		t.Fatalf("尺寸 = %dx%d", b.Dx(), b.Dy())
	}
	_ = os.Remove(video)
}

type fakeLocalPath struct{ path string }

func (f fakeLocalPath) LocalPath(string) (string, error) { return f.path, nil }

type fakePresigner struct{ url string }

func (f fakePresigner) PresignedDownloadURL(context.Context, string, time.Duration) (string, error) {
	return f.url, nil
}

// 云存储视频：经本机代理按 Range 读取并截取第一帧，不下载整个文件（未安装 ffmpeg 时跳过）
func TestVideoThumbnail_RemoteViaProxy(t *testing.T) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("未安装 ffmpeg")
	}
	dir := t.TempDir()
	video := filepath.Join(dir, "v.mp4")
	if out, err := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=640x360:rate=25",
		"-vf", "noise=alls=100:allf=t", "-t", "10", "-b:v", "20M", "-pix_fmt", "yuv420p", "-movflags", "+faststart", video).CombinedOutput(); err != nil {
		t.Fatalf("生成测试视频失败: %v %s", err, out)
	}
	data, _ := os.ReadFile(video)

	var served int64
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &countingWriter{ResponseWriter: w, n: &served}
		http.ServeContent(cw, r, "v.mp4", time.Now(), bytes.NewReader(data))
	}))
	defer remote.Close()

	s := &service{}
	thumb, err := s.videoThumbnailFrom(context.Background(), fakePresigner{url: remote.URL + "/v.mp4?sig=secret"}, "v.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(thumb)); err != nil {
		t.Fatalf("输出不是 JPEG: %v", err)
	}
	// 只应读取开头一小部分（远小于整个文件）
	if got := atomic.LoadInt64(&served); got*4 > int64(len(data)) {
		t.Errorf("读取了 %d 字节，应只读取部分内容（文件 %d 字节）", got, len(data))
	}
}

// 云存储视频索引（moov）在文件末尾且大于代理单次窗口（如 4.8GB 录像的 moov 约 3MB）：
// ffmpeg 须按偏移分段续读，仍只读取文件的一小部分（未安装 ffmpeg 时跳过）
func TestVideoThumbnail_RemoteMoovAtEndLargerThanWindow(t *testing.T) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("未安装 ffmpeg")
	}
	old := proxyWindow
	proxyWindow = 8 << 10
	defer func() { proxyWindow = old }()

	dir := t.TempDir()
	video := filepath.Join(dir, "v.mp4")
	// 不加 faststart：moov 写在文件末尾
	if out, err := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=320x180:rate=25",
		"-t", "60", "-b:v", "2M", "-pix_fmt", "yuv420p", video).CombinedOutput(); err != nil {
		t.Fatalf("生成测试视频失败: %v %s", err, out)
	}
	data, _ := os.ReadFile(video)
	if moov := bytes.LastIndex(data, []byte("moov")); moov < len(data)/2 || int64(len(data)-moov) <= proxyWindow {
		t.Fatalf("测试视频的 moov 应在末尾且大于窗口：位置 %d，大小 %d", moov, len(data)-moov)
	}

	var served int64
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &countingWriter{ResponseWriter: w, n: &served}
		http.ServeContent(cw, r, "v.mp4", time.Now(), bytes.NewReader(data))
	}))
	defer remote.Close()

	s := &service{}
	thumb, err := s.videoThumbnailFrom(context.Background(), fakePresigner{url: remote.URL + "/v.mp4"}, "v.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(thumb)); err != nil {
		t.Fatalf("输出不是 JPEG: %v", err)
	}
	if got := atomic.LoadInt64(&served); got*4 > int64(len(data)) {
		t.Errorf("读取了 %d 字节，应只读取部分内容（文件 %d 字节）", got, len(data))
	}
}

type countingWriter struct {
	http.ResponseWriter
	n *int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.ResponseWriter.Write(p)
	atomic.AddInt64(c.n, int64(n))
	return n, err
}

func TestCapRange(t *testing.T) {
	cases := map[string]string{
		"":                 "bytes=0-2097151",
		"bytes=0-":         "bytes=0-2097151",
		"bytes=100-":       "bytes=100-2097251",
		"bytes=0-99":       "bytes=0-99",
		"bytes=0-99999999": "bytes=0-2097151",
		"bytes=-500":       "bytes=-500",
	}
	for in, want := range cases {
		if got := capRange(in); got != want {
			t.Errorf("capRange(%q) = %q, want %q", in, got, want)
		}
	}
}

// HLS 录制（m3u8 + 分片）：代理按源站同目录转发分片请求，可截取第一帧
func TestVideoThumbnail_HLSViaProxy(t *testing.T) {
	ff, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("未安装 ffmpeg")
	}
	dir := t.TempDir()
	if out, err := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc=size=640x360:rate=25",
		"-t", "6", "-pix_fmt", "yuv420p", "-f", "hls", "-hls_time", "2", "-hls_list_size", "0",
		filepath.Join(dir, "index.m3u8")).CombinedOutput(); err != nil {
		t.Fatalf("生成 HLS 失败: %v %s", err, out)
	}
	remote := httptest.NewServer(http.StripPrefix("/live/room1/", http.FileServer(http.Dir(dir))))
	defer remote.Close()

	s := &service{}
	thumb, err := s.videoThumbnailFrom(context.Background(), fakePresigner{url: remote.URL + "/live/room1/index.m3u8"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(thumb)); err != nil {
		t.Fatalf("输出不是 JPEG: %v", err)
	}
}

// memStore 内存存储：验证缩略图保存到公司存储后可取回
type memStore struct {
	storage.Storage
	objects map[string][]byte
}

func (m *memStore) Upload(_ context.Context, path string, r io.Reader, _ string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	m.objects[path] = data
	return path, nil
}

func (m *memStore) Download(_ context.Context, path string) (io.ReadCloser, error) {
	data, ok := m.objects[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func TestThumbnailSaveAndReadFromStore(t *testing.T) {
	st := &memStore{objects: map[string][]byte{}}
	img, _ := encodeThumbnail(image.NewRGBA(image.Rect(0, 0, 40, 30)))
	saveThumbnail(st, "thumbnails/ab/v1_x.jpg", img, 1)
	got, err := readThumbnail(context.Background(), st, "thumbnails/ab/v1_x.jpg")
	if err != nil || !bytes.Equal(got, img) {
		t.Fatalf("取回缩略图失败: %v", err)
	}
	st.objects["thumbnails/bad.jpg"] = []byte("not a jpeg")
	if _, err := readThumbnail(context.Background(), st, "thumbnails/bad.jpg"); err == nil {
		t.Error("非 JPEG 内容应视为无效")
	}
}

// 生成失败后短期内不再重试，过期后允许重试
func TestThumbnailFailureCache(t *testing.T) {
	old := thumbnailFailTTL
	defer func() { thumbnailFailTTL = old }()
	thumbnailFailTTL = 50 * time.Millisecond
	key := filepath.Join(ThumbnailCacheDir, "ab", "v1_failure-test.jpg")
	defer thumbnailFailures.Delete(key)

	if thumbnailRecentlyFailed(key) {
		t.Fatal("未失败过不应跳过")
	}
	markThumbnailFailed(key)
	if !thumbnailRecentlyFailed(key) {
		t.Fatal("刚失败过应跳过重试")
	}
	time.Sleep(80 * time.Millisecond)
	if thumbnailRecentlyFailed(key) {
		t.Fatal("超过有效期应允许重试")
	}
}
