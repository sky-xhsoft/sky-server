package cloud

import (
	"context"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func newShareDownloadDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, stmt := range []string{
		`CREATE TABLE cloud_share (ID INTEGER PRIMARY KEY, MAX_DOWNLOADS INTEGER, DOWNLOAD_COUNT INTEGER, UPDATE_TIME DATETIME)`,
		`CREATE TABLE cloud_item (ID INTEGER PRIMARY KEY, ITEM_TYPE TEXT, NAME TEXT, ACCESS_URL TEXT, IS_ACTIVE TEXT)`,
		// 分享 1：最多 2 次；分享 2：不限制
		`INSERT INTO cloud_share VALUES (1, 2, 0, NULL), (2, 0, 0, NULL)`,
		`INSERT INTO cloud_item VALUES (10, 'file', 'a.mp4', 'https://cos.example.com/a.mp4', 'Y'), (11, 'file', 'b.txt', '/files/b.txt', 'Y')`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func downloadCount(db *gorm.DB, id uint) int {
	var share entity.CloudShare
	db.Select("ID, DOWNLOAD_COUNT").First(&share, id)
	return share.DownloadCount
}

func TestPrepareShareDownload(t *testing.T) {
	db := newShareDownloadDB(t)
	s := &service{db: db}
	ctx := context.Background()
	limited := &entity.CloudShare{}
	limited.ID = 1

	for i := 0; i < 2; i++ {
		info, err := s.PrepareShareDownload(ctx, limited, 10)
		if err != nil || info.URL != "https://cos.example.com/a.mp4" {
			t.Fatalf("第 %d 次下载应返回直链, got %+v %v", i+1, info, err)
		}
	}
	if _, err := s.PrepareShareDownload(ctx, limited, 10); err == nil {
		t.Error("达到上限后应拒绝")
	}
	if got := downloadCount(db, 1); got != 2 {
		t.Errorf("下载次数应为 2, got %d", got)
	}

	// 没有直链：返回空地址且不计数，由流式接口计数
	unlimited := &entity.CloudShare{}
	unlimited.ID = 2
	info, err := s.PrepareShareDownload(ctx, unlimited, 11)
	if err != nil || info.URL != "" || info.Name != "b.txt" {
		t.Errorf("无直链应返回空地址, got %+v %v", info, err)
	}
	if got := downloadCount(db, 2); got != 0 {
		t.Errorf("无直链不应计数, got %d", got)
	}
}

func TestClaimShareDownloadConcurrent(t *testing.T) {
	db := newShareDownloadDB(t)
	s := &service{db: db}
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.ClaimShareDownload(context.Background(), 1) == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if ok != 2 || downloadCount(db, 1) != 2 {
		t.Errorf("并发 10 次只应成功 2 次, got 成功 %d 次, 计数 %d", ok, downloadCount(db, 1))
	}
}
