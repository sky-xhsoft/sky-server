package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// setupTestService 创建测试用的文件服务
func setupTestService(t *testing.T) (Service, *gorm.DB, string) {
	t.Helper()
	tmpDir := t.TempDir()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&entity.SysFile{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	svc := NewService(db, &Config{
		UploadDir:   tmpDir,
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		AllowedExts: []string{".txt", ".jpg", ".png", ".pdf"},
	})

	return svc, db, tmpDir
}

func TestFileService_GetFileByMD5_NotFound(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	// 不存在的 MD5 应返回 (nil, nil)
	file, err := svc.GetFileByMD5(ctx, "non-existent-md5")
	if err != nil {
		t.Errorf("Expected nil error for not found, got: %v", err)
	}
	if file != nil {
		t.Error("Expected nil file for not found MD5")
	}
}

func TestFileService_GetFile_NotFound(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, err := svc.GetFile(ctx, 99999)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if errors.GetCode(err) != errors.ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got code: %d", errors.GetCode(err))
	}
}

func TestFileService_GetFileByStorageName_NotFound(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	_, err := svc.GetFileByStorageName(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error")
	}
	if errors.GetCode(err) != errors.ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got code: %d", errors.GetCode(err))
	}
}

func TestFileService_ListFiles_EmptyResult(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	files, total, err := svc.ListFiles(ctx, &ListFilesRequest{
		Page:     1,
		PageSize: 10,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected 0 total, got: %d", total)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files, got: %d", len(files))
	}
}

func TestFileService_ListFiles_PaginationDefaults(t *testing.T) {
	svc, db, _ := setupTestService(t)
	ctx := context.Background()

	// 插入 5 条测试数据
	for i := 0; i < 5; i++ {
		db.Create(&entity.SysFile{
			FileName:    "test.txt",
			StorageName: "storage-" + string(rune('0'+i)),
			FilePath:    "/tmp/test",
			FileSize:    100,
		})
	}

	// 不传分页参数，应使用默认值（Page=1, PageSize=20）
	files, total, err := svc.ListFiles(ctx, &ListFilesRequest{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("Expected 5 total, got: %d", total)
	}
	if len(files) != 5 {
		t.Errorf("Expected 5 files, got: %d", len(files))
	}
}

func TestFileService_ListFiles_PageSizeClamp(t *testing.T) {
	svc, db, _ := setupTestService(t)
	ctx := context.Background()

	// 插入 1 条数据
	db.Create(&entity.SysFile{
		FileName:    "test.txt",
		StorageName: "storage-1",
		FilePath:    "/tmp/test",
		FileSize:    100,
	})

	// PageSize > 100 应被限制为 100
	_, _, err := svc.ListFiles(ctx, &ListFilesRequest{Page: 1, PageSize: 500})
	if err != nil {
		t.Errorf("Unexpected error with large pageSize: %v", err)
	}
}

func TestFileService_DeleteFile_NotFound(t *testing.T) {
	svc, _, _ := setupTestService(t)
	ctx := context.Background()

	err := svc.DeleteFile(ctx, 99999, 1)
	if err == nil {
		t.Error("Expected error when deleting non-existent file")
	}
}

func TestFileService_DownloadFile_PhysicalFileMissing(t *testing.T) {
	svc, db, _ := setupTestService(t)
	ctx := context.Background()

	// 数据库中存在记录，但物理文件不存在
	file := &entity.SysFile{
		FileName:    "test.txt",
		StorageName: "storage-1",
		FilePath:    "/non/existent/path/file.txt",
		FileSize:    100,
	}
	db.Create(file)

	_, err := svc.DownloadFile(ctx, file.ID, 1)
	if err == nil {
		t.Error("Expected error for missing physical file")
	}
	if errors.GetCode(err) != errors.ErrResourceNotFound {
		t.Errorf("Expected ErrResourceNotFound, got code: %d", errors.GetCode(err))
	}
}

func TestFileService_DeleteFile_PhysicalRemoval(t *testing.T) {
	svc, db, tmpDir := setupTestService(t)
	ctx := context.Background()

	// 创建物理文件
	physicalPath := filepath.Join(tmpDir, "test-delete.txt")
	if err := os.WriteFile(physicalPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	file := &entity.SysFile{
		FileName:    "test.txt",
		StorageName: "delete-1",
		FilePath:    physicalPath,
		FileSize:    4,
		MD5:         "unique-md5",
		BaseModel:   entity.BaseModel{IsActive: "Y"},
	}
	db.Create(file)

	if err := svc.DeleteFile(ctx, file.ID, 1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 由于是唯一 MD5，物理文件应该被删除
	if _, err := os.Stat(physicalPath); !os.IsNotExist(err) {
		t.Error("Physical file should have been removed")
	}

	// 软删除：数据库行被标记为 inactive
	var deleted entity.SysFile
	if err := db.Unscoped().First(&deleted, file.ID).Error; err != nil {
		t.Fatalf("Should still exist in DB: %v", err)
	}
	if deleted.IsActive != "N" {
		t.Errorf("Expected IS_ACTIVE='N', got: %s", deleted.IsActive)
	}
}

func TestFileService_DeleteFile_KeepPhysicalIfShared(t *testing.T) {
	svc, db, tmpDir := setupTestService(t)
	ctx := context.Background()

	physicalPath := filepath.Join(tmpDir, "shared.txt")
	if err := os.WriteFile(physicalPath, []byte("shared content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建两条共享同一 MD5 的记录
	file1 := &entity.SysFile{
		FileName:    "a.txt",
		StorageName: "share-1",
		FilePath:    physicalPath,
		MD5:         "shared-md5",
		BaseModel:   entity.BaseModel{IsActive: "Y"},
	}
	file2 := &entity.SysFile{
		FileName:    "b.txt",
		StorageName: "share-2",
		FilePath:    physicalPath,
		MD5:         "shared-md5",
		BaseModel:   entity.BaseModel{IsActive: "Y"},
	}
	db.Create(file1)
	db.Create(file2)

	// 删除其中一个，物理文件应该保留
	if err := svc.DeleteFile(ctx, file1.ID, 1); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(physicalPath); err != nil {
		t.Error("Physical file should still exist when shared")
	}
}

func TestFileService_NewService_DefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})

	svc := NewService(db, &Config{
		UploadDir: tmpDir,
	})

	if svc == nil {
		t.Error("Service should not be nil")
	}
}
