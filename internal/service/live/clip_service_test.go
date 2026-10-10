package live

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"gorm.io/gorm"
)

// 录制列表返回直播间名称，关键词可按直播间名称搜索
func TestListRecordings_RoomName(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, stmt := range []string{
		`CREATE TABLE live_room (ID INTEGER PRIMARY KEY, ROOM_NAME TEXT)`,
		`INSERT INTO live_room VALUES (92, '测试'), (93, '科技日发布会')`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AutoMigrate(&entity.LiveRecode{}); err != nil {
		t.Fatal(err)
	}
	r92, r93 := int64(92), int64(93)
	for _, rec := range []*entity.LiveRecode{
		{RoomID: &r92, StreamID: "92", RecordURL: "http://a/1.mp4", SysCompanyID: 50, IsActive: "Y"},
		{RoomID: &r93, StreamID: "93", RecordURL: "http://a/2.mp4", SysCompanyID: 50, IsActive: "Y"},
		{StreamID: "old", RecordURL: "http://a/3.mp4", SysCompanyID: 50, IsActive: "Y"},
	} {
		if err := db.Create(rec).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &clipService{db: db}
	ctx := context.WithValue(context.Background(), contextutil.CompanyIDKey, uint(50))

	list, total, err := s.ListRecordings(ctx, &RecordingFilter{})
	if err != nil || total != 3 {
		t.Fatalf("total=%d err=%v", total, err)
	}
	names := map[string]string{}
	for _, r := range list {
		names[r.StreamID] = r.RoomName
	}
	if names["92"] != "测试" || names["93"] != "科技日发布会" || names["old"] != "" {
		t.Fatalf("直播间名称填充不正确: %v", names)
	}

	list, total, _ = s.ListRecordings(ctx, &RecordingFilter{Keyword: "科技日"})
	if total != 1 || list[0].StreamID != "93" {
		t.Fatalf("按直播间名称搜索应只返回 93，got total=%d", total)
	}
}
