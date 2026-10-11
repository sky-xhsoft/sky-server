package live

import (
	"fmt"
	"testing"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

// fakeRegionProvider 记录请求的地域，返回一个占位客户端
type fakeRegionProvider struct{ regions []string }

func (f *fakeRegionProvider) GetLiveClientForRegion(region string) (*live.Client, error) {
	f.regions = append(f.regions, region)
	return &live.Client{}, nil
}

func TestCallLive_RetryOnUnsupportedRegion(t *testing.T) {
	defaultClient := &live.Client{}
	provider := &fakeRegionProvider{}
	m := NewPullStreamManagerWithProvider(defaultClient, provider)

	calls := 0
	got, err := callLive(m, func(c *live.Client) (string, error) {
		calls++
		if c == defaultClient {
			// 公司配置的地域（如 ap-nanjing）不支持直播接口
			return "", tcerr.NewTencentCloudSDKError("UnsupportedRegion", "该接口不支持此地域访问。", "req-1")
		}
		return "ok", nil
	})
	if err != nil || got != "ok" || calls != 2 {
		t.Fatalf("应改用默认直播地域重试成功，got %q %v calls=%d", got, err, calls)
	}
	if len(provider.regions) != 1 || provider.regions[0] != fallbackLiveRegion {
		t.Fatalf("应改用 %s，got %v", fallbackLiveRegion, provider.regions)
	}
}

func TestCallLive_OtherErrorNotRetried(t *testing.T) {
	provider := &fakeRegionProvider{}
	m := NewPullStreamManagerWithProvider(&live.Client{}, provider)
	calls := 0
	_, err := callLive(m, func(c *live.Client) (string, error) {
		calls++
		return "", fmt.Errorf("wrap: %w", tcerr.NewTencentCloudSDKError("InvalidParameter.InvalidTaskTime", "bad time", "req-2"))
	})
	if err == nil || calls != 1 || len(provider.regions) != 0 {
		t.Fatalf("其他错误不应重试，err=%v calls=%d regions=%v", err, calls, provider.regions)
	}
}
