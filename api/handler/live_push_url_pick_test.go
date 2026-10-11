package handler

import (
	"testing"

	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

func TestPickRoomDomain(t *testing.T) {
	d := func(name string, status, bcname int64) *tencentLive.DomainInfo {
		return &tencentLive.DomainInfo{Name: name, Status: status, BCName: bcname}
	}
	cases := []struct {
		name    string
		domains []*tencentLive.DomainInfo
		want    string
	}{
		// 新加的域名排在最前但 CNAME 未生效，应选 CNAME 已生效的
		{"优先 CNAME 已生效", []*tencentLive.DomainInfo{d("skyttt.skyzhou.cn", 1, 0), d("skytt.skyzhou.cn", 1, 0), d("skyt.skyzhou.cn", 1, 1)}, "skyt.skyzhou.cn"},
		{"都未生效时选已启用的", []*tencentLive.DomainInfo{d("a.test", 0, 1), d("b.test", 1, 0)}, "b.test"},
		{"排除系统域名", []*tencentLive.DomainInfo{d("123.livepush.tlivecloud.com", 1, 1), d("c.test", 0, 0)}, "c.test"},
		{"同一档按列表顺序", []*tencentLive.DomainInfo{d("x.test", 1, 1), d("y.test", 1, 1)}, "x.test"},
		{"没有可用域名", []*tencentLive.DomainInfo{nil, d("", 1, 1)}, ""},
	}
	for _, c := range cases {
		if got := pickRoomDomain(c.domains); got != c.want {
			t.Errorf("%s：得到 %q，应为 %q", c.name, got, c.want)
		}
	}
}
