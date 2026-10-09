package cloud

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
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
