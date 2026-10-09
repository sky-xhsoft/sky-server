package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExternalURLStorage(t *testing.T) {
	var deleted bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = true
		}
		if r.URL.Path == "/missing.mp4" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("video-bytes"))
	}))
	defer srv.Close()

	if !IsExternalURL(srv.URL+"/a.mp4") || IsExternalURL("cloud/1/a.mp4") {
		t.Fatal("IsExternalURL 判断错误")
	}
	st := NewExternalURLStorage()
	ctx := context.Background()

	rc, err := st.Download(ctx, srv.URL+"/a.mp4")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "video-bytes" {
		t.Fatalf("下载内容 = %q", data)
	}
	if _, err := st.Download(ctx, srv.URL+"/missing.mp4"); err == nil {
		t.Error("404 应返回错误")
	}
	// 删除云盘记录时不得删除外部源文件
	if err := st.Delete(ctx, srv.URL+"/a.mp4"); err != nil || deleted {
		t.Errorf("Delete 应为空操作：err=%v deleted=%v", err, deleted)
	}
	if _, err := st.Upload(ctx, "x", nil, ""); err == nil {
		t.Error("外部地址不应支持上传")
	}
}
