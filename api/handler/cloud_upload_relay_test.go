package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
)

// relayCheckService 云存储支持直传：超过 50MB 拒绝经服务器上传
type relayCheckService struct {
	cloud.Service
	checked []int64
}

func (f *relayCheckService) CheckRelayUpload(ctx context.Context, userID uint, size int64) error {
	f.checked = append(f.checked, size)
	if size > cloud.MaxRelayUploadSize {
		return errors.New(errors.ErrInvalidParam, "超过 50MB 的文件需直传到云存储")
	}
	return nil
}

// countingBody 记录服务端从请求体读取了多少字节
type countingBody struct {
	r    io.Reader
	read int64
}

func (b *countingBody) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	b.read += int64(n)
	return n, err
}

func (b *countingBody) Close() error { return nil }

func TestUploadFile_RejectsLargeRelayBeforeReadingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &relayCheckService{}
	h := NewCloudHandler(svc)
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) { c.Set("userID", uint(7)); h.UploadFile(c) })

	size := cloud.MaxRelayUploadSize + 10<<20
	body := &countingBody{r: io.LimitReader(strings.NewReader(strings.Repeat("x", 1<<20)), size)}
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Body = body
	req.ContentLength = size
	req.Header.Set("Content-Type", "multipart/form-data; boundary=b")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code < 400 {
		t.Fatalf("超过 50MB 应拒绝，got %d", w.Code)
	}
	if body.read != 0 {
		t.Errorf("应在读取请求体之前拒绝，实际已读取 %d 字节", body.read)
	}
	if len(svc.checked) != 1 || svc.checked[0] <= cloud.MaxRelayUploadSize {
		t.Errorf("应按 Content-Length 检查一次，got %v", svc.checked)
	}
}
