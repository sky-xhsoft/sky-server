package cloud

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestCheckCompanyQuota(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, stmt := range []string{
		`CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER)`,
		`CREATE TABLE sys_company_conf (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, IS_ACTIVE TEXT, MAX_STORAGE_SIZE INTEGER, MAX_FILE_SIZE INTEGER, MAX_FILES INTEGER)`,
		`CREATE TABLE cloud_quota (ID INTEGER PRIMARY KEY, USER_ID INTEGER, USED_SPACE INTEGER, FILE_COUNT INTEGER, IS_ACTIVE TEXT)`,
		// 公司 1：总量 1000 字节、单文件 300 字节、最多 5 个文件；两个用户已用 700 字节、4 个文件
		// 公司 2：全部为 0 不限制；公司 3：没有配置
		`INSERT INTO sys_user VALUES (1, 1), (2, 1), (3, 2), (4, 3)`,
		`INSERT INTO sys_company_conf VALUES (1, 1, 'Y', 1000, 300, 5), (2, 2, 'Y', 0, 0, 0)`,
		`INSERT INTO cloud_quota VALUES (1, 1, 500, 3, 'Y'), (2, 2, 200, 1, 'Y'), (3, 3, 999999, 999, 'Y')`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &service{db: db}
	ctx := context.Background()

	cases := []struct {
		name    string
		user    uint
		size    int64
		wantErr bool
	}{
		{"未超限", 1, 300, false},
		{"超过单文件上限", 1, 301, true},
		{"超过公司总量（另一用户的用量也计入）", 2, 301, true},
		{"公司配置为 0 不限制", 3, 1 << 40, false},
		{"公司没有配置不限制", 4, 1 << 40, false},
	}
	for _, c := range cases {
		if err := s.checkCompanyQuota(ctx, c.user, c.size); (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", c.name, err, c.wantErr)
		}
	}

	// 文件数达到上限
	db.Exec(`UPDATE cloud_quota SET FILE_COUNT = 4 WHERE ID = 1`)
	if err := s.checkCompanyQuota(ctx, 1, 1); err == nil {
		t.Error("公司文件数已满 5 个时应拒绝上传")
	}
}
