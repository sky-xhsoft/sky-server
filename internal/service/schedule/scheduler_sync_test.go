package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func TestSyncNextRun(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE sys_schedule (ID INTEGER PRIMARY KEY, CRON TEXT, IS_ACTIVE TEXT, NEXT_RUN_AT DATETIME, UPDATE_TIME DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 7, 10, 3, 0, 0, time.Local)
	far := now.Add(24 * time.Hour)
	db.Exec(`INSERT INTO sys_schedule (ID, CRON, IS_ACTIVE, NEXT_RUN_AT) VALUES (1, '*/5 * * * *', 'Y', NULL), (2, '*/5 * * * *', 'Y', ?), (3, 'bad', 'Y', NULL), (4, '*/5 * * * *', 'N', NULL)`, far)

	(&Scheduler{db: db}).syncNextRun(context.Background(), now)

	get := func(id int) *time.Time {
		var sc entity.SysSchedule
		db.Select("ID, NEXT_RUN_AT").First(&sc, id)
		return sc.NextRunAt
	}
	want := time.Date(2026, 10, 7, 10, 5, 0, 0, time.Local)
	if got := get(1); got == nil || !got.Equal(want) {
		t.Errorf("新建任务应补算下次时间 %v, got %v", want, got)
	}
	if got := get(2); got == nil || !got.Equal(want) {
		t.Errorf("改为更频繁的 cron 后应提前到 %v, got %v", want, got)
	}
	if got := get(3); got != nil {
		t.Errorf("无效 cron 不应计算下次时间, got %v", got)
	}
	if got := get(4); got != nil {
		t.Errorf("停用任务不应计算下次时间, got %v", got)
	}
}
