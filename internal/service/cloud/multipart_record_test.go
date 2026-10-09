package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// 并发上报分片时，会话的已上传列表包含全部分片且不重复
func TestRecordChunkUploaded_ConcurrentChunksNotLost(t *testing.T) {
	// SQLite 不支持行锁，限制为单连接让写事务串行执行；这里验证记录与去重逻辑，
	// 真实并发（MySQL 行锁）由分片上传端到端测试覆盖
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(&entity.CloudUploadSession{}, &entity.CloudChunkRecord{}); err != nil {
		t.Fatal(err)
	}
	expire := time.Now().Add(time.Hour)
	session := entity.CloudUploadSession{UserID: 1, FileName: "a.mp4", TotalChunks: 30, UploadedChunks: "[]", Status: "uploading", ExpireTime: &expire}
	if err := db.Create(&session).Error; err != nil {
		t.Fatal(err)
	}

	s := &multipartUploadService{db: db}
	var wg sync.WaitGroup
	errs := make(chan error, 30)
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs <- s.recordChunkUploaded(context.Background(), session.ID, idx, 1024, fmt.Sprintf("md5-%d", idx), "", "")
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	var got entity.CloudUploadSession
	db.First(&got, session.ID)
	var uploaded []int
	_ = json.Unmarshal([]byte(got.UploadedChunks), &uploaded)
	sort.Ints(uploaded)
	if len(uploaded) != 30 {
		t.Fatalf("应记录 30 个分片，实际 %d 个: %v", len(uploaded), uploaded)
	}
	var records int64
	db.Model(&entity.CloudChunkRecord{}).Where("SESSION_ID = ? AND UPLOADED = ?", session.ID, true).Count(&records)
	if records != 30 {
		t.Errorf("分片记录应为 30 条，实际 %d 条", records)
	}

	// 同一分片重复上报（客户端重试）不重复计数
	if err := s.recordChunkUploaded(context.Background(), session.ID, 5, 1024, "md5-5", "", "etag-5"); err != nil {
		t.Fatal(err)
	}
	db.First(&got, session.ID)
	_ = json.Unmarshal([]byte(got.UploadedChunks), &uploaded)
	if len(uploaded) != 30 {
		t.Errorf("重复上报后仍应为 30 个，实际 %d 个", len(uploaded))
	}
}
