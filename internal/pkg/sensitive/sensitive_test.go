package sensitive

import (
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		mustHide []string
		mustKeep []string
	}{
		{"小写 password", `{"username":"admin","password":"admin123"}`, []string{"admin123"}, []string{`"username":"admin"`}},
		{"大写 PASSWORD（通用 CRUD）", `{"USERNAME":"u1","PASSWORD":"Secr3t!"}`, []string{"Secr3t!"}, []string{`"USERNAME":"u1"`}},
		{"驼峰 secretKey", `{"tencentCloudSecretKey":"mKdiAbCdEf123456","region":"ap-guangzhou"}`, []string{"mKdiAbCdEf123456"}, []string{`"region":"ap-guangzhou"`}},
		{"下划线 SECRET_KEY", `{"SECRET_KEY":"abcdefgh12345678"}`, []string{"abcdefgh12345678"}, nil},
		{"嵌套与数组", `{"items":[{"token":"t-123456"}],"conf":{"accessKeySecret":"zzz999"}}`, []string{"t-123456", "zzz999"}, nil},
		{"大整数不丢精度", `{"id":1234567890123456789,"pwd":"p@ss"}`, []string{"p@ss"}, []string{"1234567890123456789"}},
		{"表单格式", `username=admin&password=admin123&remember=1`, []string{"admin123"}, []string{"username=admin", "remember=1"}},
		{"非 JSON 也非表单", `plain text without secrets`, nil, []string{"plain text without secrets"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.in)
			for _, s := range tc.mustHide {
				if strings.Contains(got, s) {
					t.Errorf("敏感值 %q 未被遮盖: %s", s, got)
				}
			}
			for _, s := range tc.mustKeep {
				if !strings.Contains(got, s) {
					t.Errorf("普通内容 %q 丢失: %s", s, got)
				}
			}
		})
	}
}
