package schedule

import (
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestNextRunAfter(t *testing.T) {
	cases := []struct {
		name, expr, from, want string
	}{
		{"每 5 分钟", "*/5 * * * *", "2026-10-06 10:01", "2026-10-06 10:05"},
		{"每小时整点", "0 * * * *", "2026-10-06 10:01", "2026-10-06 11:00"},
		{"范围加步长", "10-30/10 * * * *", "2026-10-06 10:11", "2026-10-06 10:20"},
		{"列表", "0 8,20 * * *", "2026-10-06 09:00", "2026-10-06 20:00"},
		{"每天 2 点", "0 2 * * *", "2026-10-06 10:00", "2026-10-07 02:00"},
		{"每周一 9 点", "0 9 * * 1", "2026-10-06 10:00", "2026-10-12 09:00"}, // 2026-10-06 是周二
		{"周日写作 7", "0 9 * * 7", "2026-10-06 10:00", "2026-10-11 09:00"},
		{"每月 1 号", "0 0 1 * *", "2026-10-06 10:00", "2026-11-01 00:00"},
		// 标准 cron：日与周同时指定时任一满足即可
		{"日或周", "0 0 1 * 1", "2026-10-06 10:00", "2026-10-12 00:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NextRunAfter(tc.expr, mustTime(t, tc.from))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if want := mustTime(t, tc.want); !got.Equal(want) {
				t.Errorf("%q 从 %s 起: got %s, want %s", tc.expr, tc.from, got.Format("2006-01-02 15:04"), tc.want)
			}
		})
	}
}

func TestNextRunAfter_Invalid(t *testing.T) {
	for _, expr := range []string{
		"* * * *",       // 段数不足
		"abc * * * *",   // 非数字
		"75 * * * *",    // 分钟越界
		"* 24 * * *",    // 小时越界
		"* * 0 * *",     // 日越界
		"* * * 13 *",    // 月越界
		"5-1 * * * *",   // 范围颠倒
		"*/x * * * *",   // 步长非法
	} {
		if _, err := NextRunAfter(expr, time.Now()); err == nil {
			t.Errorf("非法表达式 %q 应报错", expr)
		}
	}
}
