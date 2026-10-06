package workflow

import "testing"

func TestEvalCondition(t *testing.T) {
	vars := map[string]interface{}{
		"amount":   float64(12000), // JSON 反序列化后的数字
		"days":     3,
		"dept":     "sales",
		"urgent":   true,
		"strScore": "88",
	}
	cases := []struct {
		cond string
		want bool
	}{
		{"", true},
		{"amount > 10000", true},
		{"amount >= 12000", true},
		{"amount < 10000", false},
		{"days <= 3", true},
		{"days != 3", false},
		{`dept == "sales"`, true},
		{"dept == 'hr'", false},
		{"dept != hr", true},
		{"urgent == true", true},
		{"strScore > 60", true},
		{`amount > 10000 && dept == "sales"`, true},
		{`amount > 10000 && dept == "hr"`, false},
		{`amount < 100 || urgent == true`, true},
		{"missing == 1", false},
	}
	for _, tc := range cases {
		got, err := evalCondition(tc.cond, vars)
		if err != nil {
			t.Errorf("%q: unexpected error %v", tc.cond, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%q = %v, want %v", tc.cond, got, tc.want)
		}
	}
}

func TestEvalCondition_Invalid(t *testing.T) {
	for _, cond := range []string{"amount", "== 3", "dept > abc"} {
		if _, err := evalCondition(cond, map[string]interface{}{"amount": 1, "dept": "x"}); err == nil {
			t.Errorf("%q 应报错", cond)
		}
	}
	// 服务层对非法条件按“不满足”处理
	if (&service{}).evaluateCondition("amount", map[string]interface{}{"amount": 1}) {
		t.Error("非法条件不应视为满足")
	}
}
