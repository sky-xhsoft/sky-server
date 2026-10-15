package config

import "testing"

func TestAppConfig_SubdomainSuffix(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"  ":                 "",
		"zhibo.skyzhou.cn":   ".zhibo.skyzhou.cn",
		".zhibo.skyzhou.cn":  ".zhibo.skyzhou.cn",
		"zhibo.skyzhou.cn.":  ".zhibo.skyzhou.cn",
		" zhibo.skyzhou.cn ": ".zhibo.skyzhou.cn",
	}
	for in, want := range cases {
		if got := (AppConfig{SubdomainBase: in}).SubdomainSuffix(); got != want {
			t.Errorf("SubdomainBase=%q got %q want %q", in, got, want)
		}
	}
}
