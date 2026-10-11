package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMediaFileSize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/clip.mp4":
			if r.Method != http.MethodHead {
				t.Errorf("应使用 HEAD 请求，实际 %s", r.Method)
			}
			w.Header().Set("Content-Length", "313488")
		case "/moved.mp4":
			http.Redirect(w, r, "/clip.mp4", http.StatusFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// 非腾讯云域名不发请求
	if got := mediaFileSize(srv.URL + "/clip.mp4"); got != 0 {
		t.Fatalf("非腾讯云地址应返回 0，实际 %d", got)
	}

	orig := mediaURLAllowed
	mediaURLAllowed = func(string) bool { return true }
	defer func() { mediaURLAllowed = orig }()

	if got := mediaFileSize(srv.URL + "/clip.mp4"); got != 313488 {
		t.Fatalf("文件大小应为 313488，实际 %d", got)
	}
	if got := mediaFileSize(srv.URL + "/missing.mp4"); got != 0 {
		t.Fatalf("404 应返回 0，实际 %d", got)
	}
	if got := mediaFileSize(srv.URL + "/moved.mp4"); got != 0 {
		t.Fatalf("不应跟随跳转，实际 %d", got)
	}
	if got := fileSizeOr(100, srv.URL+"/clip.mp4"); got != 100 {
		t.Fatalf("回调已带大小时直接使用，实际 %d", got)
	}
	if got := fileSizeOr(0, srv.URL+"/clip.mp4"); got != 313488 {
		t.Fatalf("回调未带大小时查询 COS，实际 %d", got)
	}
}
