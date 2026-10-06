package crud

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"gorm.io/gorm"
)

// fakeMetadata 只实现外键显示用到的两个方法，其余方法由嵌入的空接口兜底（调用会 panic）
type fakeMetadata struct {
	metadata.Service
	table   *entity.SysTable
	display *entity.SysColumn
}

func (f *fakeMetadata) GetTableByID(tableID uint) (*entity.SysTable, error) { return f.table, nil }

func (f *fakeMetadata) ResolveDisplayColumn(tableID uint, columnID *uint) (*entity.SysColumn, error) {
	return f.display, nil
}

func TestProcessForeignKeys_MixedIDTypes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE sys_company (ID INTEGER PRIMARY KEY, NAME TEXT)`)
	db.Exec(`INSERT INTO sys_company VALUES (1, '先桦科技'), (2, '第二公司')`)

	refTableID := uint(4)
	svc := &service{db: db, metadataService: &fakeMetadata{
		table:   &entity.SysTable{Name: "sys_company"},
		display: &entity.SysColumn{DbName: "NAME"},
	}}
	columns := []*entity.SysColumn{{DbName: "SYS_COMPANY_ID", SetValueType: "fk", RefTableID: &refTableID}}
	// 主记录里的外键值是 uint（无符号列），关联表 ID 被扫描成 int64：类型不同但值相同
	results := []map[string]interface{}{
		{"ID": int64(10), "SYS_COMPANY_ID": uint32(1)},
		{"ID": int64(11), "SYS_COMPANY_ID": uint32(2)},
		{"ID": int64(12), "SYS_COMPANY_ID": nil},
	}

	if err := svc.processForeignKeys(context.Background(), columns, results, 1); err != nil {
		t.Fatal(err)
	}
	if results[0]["SYS_COMPANY_ID_display"] != "先桦科技" || results[1]["SYS_COMPANY_ID_display"] != "第二公司" {
		t.Errorf("类型不同的同值 ID 应匹配到显示值, got %v / %v", results[0]["SYS_COMPANY_ID_display"], results[1]["SYS_COMPANY_ID_display"])
	}
	if _, ok := results[2]["SYS_COMPANY_ID_display"]; ok {
		t.Error("空外键不应生成显示值")
	}
}
