package cloud

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

func TestNeedsPDFPreview(t *testing.T) {
	ext := func(e string) *string { return &e }
	cases := map[string]bool{".doc": true, "DOC": true, ".wps": true, ".dot": true, ".docx": false, ".pdf": false, ".xls": false}
	for e, want := range cases {
		if got := NeedsPDFPreview(&entity.CloudItem{Name: "a", FileExt: ext(e)}); got != want {
			t.Errorf("%s: got %v want %v", e, got, want)
		}
	}
	if !NeedsPDFPreview(&entity.CloudItem{Name: "委托书.doc"}) {
		t.Error("无 FILE_EXT 时应按文件名判断")
	}
}

func TestPreviewCachePath(t *testing.T) {
	md5 := "79a37d98726b59620aa7180f80e5d993"
	path := "cloud/1/a.doc"
	got := previewCachePath(&entity.CloudItem{MD5: &md5, StoragePath: &path})
	want := filepath.Join(PreviewCacheDir, "79", "v1_"+md5+".pdf")
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestFileURL(t *testing.T) {
	if got := fileURL("/tmp/x/profile"); got != "file:///tmp/x/profile" {
		t.Errorf("unix: %s", got)
	}
	if got := fileURL(`C:\tmp\profile`); got != "file:///C:/tmp/profile" && !strings.HasPrefix(got, "file:///") {
		t.Errorf("windows: %s", got)
	}
}

// 用 LibreOffice 把文档转为 PDF（未安装 LibreOffice 时跳过）
func TestRunSoffice(t *testing.T) {
	soffice := findSoffice()
	if soffice == "" {
		t.Skip("未安装 LibreOffice")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "source.rtf")
	if err := os.WriteFile(src, []byte(`{\rtf1\ansi {\b 委托授权书} \par 测试内容}`), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	data, err := runSoffice(ctx, soffice, src, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !isPDF(data) {
		t.Fatal("输出不是 PDF")
	}
}
