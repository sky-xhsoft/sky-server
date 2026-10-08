package cloud

import (
	"strings"
	"testing"
)

func TestValidateItemName(t *testing.T) {
	ok := map[string]string{
		"报告.txt":      "报告.txt",
		"  项目资料  ":    "项目资料",
		"a.b.c":       "a.b.c",
		"0427-ICAR 峰会": "0427-ICAR 峰会",
	}
	for in, want := range ok {
		got, err := ValidateItemName(in)
		if err != nil || got != want {
			t.Errorf("ValidateItemName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "   ", "a/b", `a\b`, ".", "..", strings.Repeat("长", 256)} {
		if _, err := ValidateItemName(bad); err == nil {
			t.Errorf("ValidateItemName(%q) 应被拒绝", bad)
		}
	}
}
