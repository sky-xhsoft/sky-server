package live

import (
	"testing"

	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

func TestExcludeSystemDomains(t *testing.T) {
	in := []*tencentLive.DomainInfo{
		{Name: "push.skyzhou.cn"},
		{Name: "12345.livepush.myqcloud.com"},
		{Name: "12345.liveplay.tlivecloud.com"},
		{Name: "ABC.LivePlay.TLiveCloud.com"},
		{Name: "live.skyzhou.cn"},
		{Name: "tlivecloud.com.example.cn"},
		nil,
	}
	got := excludeSystemDomains(in)
	var names []string
	for _, d := range got {
		names = append(names, d.Name)
	}
	want := []string{"push.skyzhou.cn", "live.skyzhou.cn", "tlivecloud.com.example.cn"}
	if len(names) != len(want) {
		t.Fatalf("过滤结果 %v，期望 %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("过滤结果 %v，期望 %v", names, want)
		}
	}
}
