package live

import (
	"testing"
	"time"

	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

func TestDomainCacheTTL(t *testing.T) {
	ok := &tencentLive.DomainInfo{Name: "push.example.com", BCName: 1}
	pending := &tencentLive.DomainInfo{Name: "play.example.com", BCName: 0}

	cases := []struct {
		name    string
		domains []*tencentLive.DomainInfo
		want    time.Duration
	}{
		{"全部已配置 CNAME", []*tencentLive.DomainInfo{ok, ok}, 24 * time.Hour},
		{"有一个未配置 CNAME", []*tencentLive.DomainInfo{ok, pending}, 5 * time.Minute},
		{"单个域名未配置", []*tencentLive.DomainInfo{pending}, 5 * time.Minute},
		{"没有域名", nil, 24 * time.Hour},
		{"忽略空元素", []*tencentLive.DomainInfo{nil, ok}, 24 * time.Hour},
	}
	for _, c := range cases {
		if got := domainCacheTTL(c.domains...); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
