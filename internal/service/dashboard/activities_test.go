package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// 最新动态只返回本公司、成功、可识别的操作
func TestGetActivities_CompanyScopedAndRecognized(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	logs := []*entity.AuditLog{
		{SysCompanyID: 50, Username: "CLX", Action: "upload", Resource: "cloud_file", Status: "success", CreatedAt: now},
		{SysCompanyID: 50, Username: "CLX", Action: "create", Resource: "unknown", Status: "success", CreatedAt: now},    // 未识别
		{SysCompanyID: 50, Username: "CLX", Action: "delete", Resource: "cloud_file", Status: "failure", CreatedAt: now}, // 失败
		{SysCompanyID: 0, Username: "CLX", Action: "create", Resource: "live_room", Status: "success", CreatedAt: now},   // 历史未记公司
		{SysCompanyID: 1, Username: "admin", Action: "create", Resource: "live_room", Status: "success", CreatedAt: now}, // 其他公司
		{SysCompanyID: 50, Username: "", Action: "login", Resource: "user", Status: "success", CreatedAt: now},           // 无操作人
	}
	for _, l := range logs {
		if err := db.Create(l).Error; err != nil {
			t.Fatal(err)
		}
	}
	got, err := (&service{db: db}).GetActivities(context.Background(), 50, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Action != "upload" || got[0].Resource != "cloud_file" {
		t.Fatalf("应只返回公司 50 的一条上传记录，got %+v", got)
	}
}
