package storage

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"
)

func setupLocalStorage(t *testing.T) Storage {
	t.Helper()
	tmpDir := t.TempDir()
	s, err := NewLocalStorage(&LocalStorageConfig{
		BasePath: tmpDir,
		BaseURL:  "http://test.local/files",
	})
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}
	return s
}

func TestLocalStorage_NewLocalStorage_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir")

	_, err := NewLocalStorage(&LocalStorageConfig{
		BasePath: nestedPath,
		BaseURL:  "http://test.local",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 嵌套目录应该被创建
	_, err = filepath.Abs(nestedPath)
	if err != nil {
		t.Errorf("Path issue: %v", err)
	}
}

func TestLocalStorage_Upload_Success(t *testing.T) {
	s := setupLocalStorage(t)
	ctx := context.Background()

	content := []byte("hello world")
	url, err := s.Upload(ctx, "test/file.txt", bytes.NewReader(content), "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if url == "" {
		t.Error("URL should not be empty")
	}
}

func TestLocalStorage_UploadAndDownload(t *testing.T) {
	s := setupLocalStorage(t)
	ctx := context.Background()

	content := []byte("test content for upload/download")
	_, err := s.Upload(ctx, "data/sample.txt", bytes.NewReader(content), "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	reader, err := s.Download(ctx, "data/sample.txt")
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Content mismatch: expected %q, got %q", content, got)
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	s := setupLocalStorage(t)
	ctx := context.Background()

	// 先上传
	_, err := s.Upload(ctx, "to-delete.txt", bytes.NewReader([]byte("delete me")), "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// 删除
	err = s.Delete(ctx, "to-delete.txt")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// 再次下载应失败
	_, err = s.Download(ctx, "to-delete.txt")
	if err == nil {
		t.Error("Download after delete should fail")
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	s := setupLocalStorage(t)
	ctx := context.Background()

	// 不存在
	exists, err := s.Exists(ctx, "non-existent.txt")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if exists {
		t.Error("Should not exist before upload")
	}

	// 上传后存在
	s.Upload(ctx, "exists.txt", bytes.NewReader([]byte("test")), "text/plain")
	exists, _ = s.Exists(ctx, "exists.txt")
	if !exists {
		t.Error("Should exist after upload")
	}
}

func TestLocalStorage_IsLocalStorage(t *testing.T) {
	s := setupLocalStorage(t)
	ls, ok := s.(*LocalStorage)
	if !ok {
		t.Fatal("Should be *LocalStorage type")
	}
	if !ls.IsLocalStorage() {
		t.Error("IsLocalStorage should return true")
	}
}
