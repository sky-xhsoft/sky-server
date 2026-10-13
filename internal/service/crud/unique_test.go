package crud

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUniqueDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, USERNAME TEXT, PHONE TEXT, TRUE_NAME TEXT, IS_ACTIVE TEXT)",
		// 34、35 手机号重复（历史数据）；39 已停用
		"INSERT INTO sys_user VALUES (31, 'sajida', '13579663748', '', 'Y'), (34, '123', '123', '', 'Y'), (35, '123_55', '123', '', 'Y'), (39, 'olduser', '13700000000', '', 'N')",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestCheckUniqueColumns_SysUser(t *testing.T) {
	db := setupUniqueDB(t)
	cases := []struct {
		name    string
		id      uint
		data    map[string]interface{}
		wantErr string
	}{
		{"新增重复用户名", 0, map[string]interface{}{"USERNAME": "sajida"}, "该用户名已被使用"},
		{"新增重复用户名（首尾空格）", 0, map[string]interface{}{"USERNAME": " sajida "}, "该用户名已被使用"},
		{"新增重复手机号", 0, map[string]interface{}{"USERNAME": "new", "PHONE": "13579663748"}, "该手机号已被使用"},
		{"新增：已停用用户的用户名、手机号可用", 0, map[string]interface{}{"USERNAME": "olduser", "PHONE": "13700000000"}, ""},
		{"新增：手机号为空不检查", 0, map[string]interface{}{"USERNAME": "new", "PHONE": ""}, ""},
		{"修改：不改用户名、手机号（历史重复不拦截）", 35, map[string]interface{}{"USERNAME": "123_55", "PHONE": "123", "TRUE_NAME": "x"}, ""},
		{"修改：用户名改成别人的", 35, map[string]interface{}{"USERNAME": "sajida"}, "该用户名已被使用"},
		{"修改：手机号改成别人的", 31, map[string]interface{}{"PHONE": "123"}, "该手机号已被使用"},
		{"修改：用户名改成未被使用的", 35, map[string]interface{}{"USERNAME": "abc"}, ""},
	}
	for _, c := range cases {
		err := checkUniqueColumns(db, "sys_user", c.id, c.data)
		switch {
		case c.wantErr == "" && err != nil:
			t.Errorf("%s: 不应报错，实际 %v", c.name, err)
		case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
			t.Errorf("%s: 期望「%s」，实际 %v", c.name, c.wantErr, err)
		}
	}
	// 其他表不检查
	if err := checkUniqueColumns(db, "live_room", 0, map[string]interface{}{"USERNAME": "sajida"}); err != nil {
		t.Errorf("未配置唯一字段的表不应检查: %v", err)
	}
}
