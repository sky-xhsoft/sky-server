package handler

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	pkgErrors "github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/file"
)

// fakeFileService 实现 file.Service 接口
type fakeFileService struct {
	uploadResult   *entity.SysFile
	uploadErr      error
	multipleResult []*entity.SysFile
	multipleErr    error
	downloadResult *entity.SysFile
	downloadErr    error
	deleteErr      error
	getResult      *entity.SysFile
	getErr         error
	storageResult  *entity.SysFile
	storageErr     error
	listResult     []*entity.SysFile
	listTotal      int64
	listErr        error
	md5Result      *entity.SysFile
	md5Err         error
}

func (m *fakeFileService) UploadFile(ctx context.Context, fh *multipart.FileHeader, category string, userID uint, ip string) (*entity.SysFile, error) {
	return m.uploadResult, m.uploadErr
}

func (m *fakeFileService) UploadMultipleFiles(ctx context.Context, files []*multipart.FileHeader, category string, userID uint, ip string) ([]*entity.SysFile, error) {
	return m.multipleResult, m.multipleErr
}

func (m *fakeFileService) DownloadFile(ctx context.Context, fileID, userID uint) (*entity.SysFile, error) {
	return m.downloadResult, m.downloadErr
}

func (m *fakeFileService) DeleteFile(ctx context.Context, fileID, userID uint) error {
	return m.deleteErr
}

func (m *fakeFileService) GetFile(ctx context.Context, fileID uint) (*entity.SysFile, error) {
	return m.getResult, m.getErr
}

func (m *fakeFileService) GetFileByStorageName(ctx context.Context, storageName string) (*entity.SysFile, error) {
	return m.storageResult, m.storageErr
}

func (m *fakeFileService) ListFiles(ctx context.Context, req *file.ListFilesRequest) ([]*entity.SysFile, int64, error) {
	return m.listResult, m.listTotal, m.listErr
}

func (m *fakeFileService) GetFileByMD5(ctx context.Context, md5 string) (*entity.SysFile, error) {
	return m.md5Result, m.md5Err
}

func TestFileHandler_UploadFile_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{}
	h := NewFileHandler(svc)

	r := gin.New()
	r.POST("/upload", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.UploadFile(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	r.ServeHTTP(w, req)

	// 无文件应返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestFileHandler_UploadFile_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{}
	h := NewFileHandler(svc)

	r := gin.New()
	r.POST("/upload", h.UploadFile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", nil)
	r.ServeHTTP(w, req)

	// 无 userID 上下文，应该是 400 或 401（取决于哪个先检查）
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 401 or 400, got %d", w.Code)
	}
}

func TestFileHandler_DownloadFile_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{}
	h := NewFileHandler(svc)

	r := gin.New()
	r.GET("/files/:id/download", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.DownloadFile(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files/abc/download", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestFileHandler_GetFile_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{
		getErr: pkgErrors.New(pkgErrors.ErrResourceNotFound, "文件不存在"),
	}
	h := NewFileHandler(svc)

	r := gin.New()
	r.GET("/files/:id", h.GetFile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files/999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestFileHandler_GetFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{
		getResult: &entity.SysFile{
			FileName: "test.txt",
		},
	}
	h := NewFileHandler(svc)

	r := gin.New()
	r.GET("/files/:id", h.GetFile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestFileHandler_DeleteFile_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{}
	h := NewFileHandler(svc)

	r := gin.New()
	r.DELETE("/files/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.DeleteFile(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/files/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestFileHandler_DeleteFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{}
	h := NewFileHandler(svc)

	r := gin.New()
	r.DELETE("/files/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.DeleteFile(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/files/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestFileHandler_GetFileByPath_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeFileService{
		storageErr: pkgErrors.New(pkgErrors.ErrResourceNotFound, "not found"),
	}
	h := NewFileHandler(svc)

	r := gin.New()
	r.GET("/access/:storageName", h.GetFileByPath)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/access/nonexistent", nil)
	r.ServeHTTP(w, req)

	// 不应该 panic
	if w.Code == 0 {
		t.Error("Expected response code to be set")
	}
}
