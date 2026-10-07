package schedule

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func TestClaimOnlyOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE sys_schedule (ID INTEGER PRIMARY KEY, CRON TEXT, IS_ACTIVE TEXT, NEXT_RUN_AT DATETIME, UPDATE_TIME DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 7, 10, 3, 0, 0, time.Local)
	due := now.Add(-time.Minute)
	db.Exec(`INSERT INTO sys_schedule (ID, CRON, IS_ACTIVE, NEXT_RUN_AT) VALUES (1, '*/5 * * * *', 'Y', ?)`, due)

	// 两个实例读到同一条到期任务
	var a, b entity.SysSchedule
	db.First(&a, 1)
	db.First(&b, 1)
	s := &Scheduler{db: db}
	ctx := context.Background()

	if !s.claim(ctx, &a, now) {
		t.Fatal("第一个实例应抢到任务")
	}
	if s.claim(ctx, &b, now) {
		t.Error("第二个实例不应再抢到同一次任务")
	}
	var sc entity.SysSchedule
	db.First(&sc, 1)
	if want := time.Date(2026, 10, 7, 10, 5, 0, 0, time.Local); sc.NextRunAt == nil || !sc.NextRunAt.Equal(want) {
		t.Errorf("抢到后下次时间应推到 %v, got %v", want, sc.NextRunAt)
	}
}

func TestScanConcurrentInstancesRunOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // 内存库每个连接是独立的库
	if err := db.AutoMigrate(&entity.SysSchedule{}, &entity.SysScheduleLog{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO sys_schedule (ID, SYS_COMPANY_ID, CRON, IS_ACTIVE, NEXT_RUN_AT) VALUES (1, 1, '*/5 * * * *', 'Y', ?)`, time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}

	var runs int32
	run := func(ctx context.Context, actionType, actionContent string, companyID uint) (string, error) {
		atomic.AddInt32(&runs, 1)
		return "ok", nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			NewScheduler(db, run).scan()
		}()
	}
	wg.Wait()
	if runs != 1 {
		t.Errorf("5 个实例同时扫描，任务应只执行 1 次，实际 %d 次", runs)
	}
}
