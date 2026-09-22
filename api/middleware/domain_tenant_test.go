package middleware

import "testing"

func TestParseSubdomainCompanyID(t *testing.T) {
	cases := []struct {
		host string
		want uint
		ok   bool
	}{
		{"1.zhibo.skyzhou.cn", 1, true},
		{"46.zhibo.skyzhou.cn", 46, true},
		{"123.zhibo.skyzhou.cn", 123, true},
		{"zhibo.skyzhou.cn", 0, false},
		{"a.1.zhibo.skyzhou.cn", 0, false},
		{"foo.zhibo.skyzhou.cn", 0, false},
		{"localhost", 0, false},
		{"0.zhibo.skyzhou.cn", 0, false},
	}
	for _, c := range cases {
		got, ok := parseSubdomainCompanyID(c.host)
		if got != c.want || ok != c.ok {
			t.Errorf("host=%q got=(%d,%v) want=(%d,%v)", c.host, got, ok, c.want, c.ok)
		}
	}
}