package sqlutils

import "testing"

func TestSanitizeIdentifier_Valid(t *testing.T) {
	valid := []string{
		"users",
		"sys_user",
		"_private",
		"Table1",
		"SYS_TABLE",
		"a",
		"a1b2c3",
		"my_table_v2",
	}

	for _, name := range valid {
		got, err := SanitizeIdentifier(name)
		if err != nil {
			t.Errorf("SanitizeIdentifier(%q) should be valid, got error: %v", name, err)
		}
		if got != name {
			t.Errorf("SanitizeIdentifier(%q): expected %q, got %q", name, name, got)
		}
	}
}

func TestSanitizeIdentifier_Invalid(t *testing.T) {
	invalid := []string{
		"",                  // 空
		"1table",            // 数字开头
		"my-table",          // 含横线
		"my table",          // 含空格
		"my;table",          // SQL 注入字符
		"table'name",        // 引号
		"DROP TABLE users",  // SQL 注入
		"table/*comment*/",  // 注释字符
		"my.table",          // 含点
		"my(table)",         // 含括号
	}

	for _, name := range invalid {
		_, err := SanitizeIdentifier(name)
		if err == nil {
			t.Errorf("SanitizeIdentifier(%q) should reject", name)
		}
	}
}
