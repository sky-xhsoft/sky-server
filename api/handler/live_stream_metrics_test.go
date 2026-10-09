package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// fakeMetricsLiveService 只实现 GetStreamMetrics
type fakeMetricsLiveService struct {
	liveService.Service
	err error
}

func (f *fakeMetricsLiveService) GetStreamMetrics(ctx context.Context, req *live.GetStreamMetricsRequest) (*live.StreamMetrics, error) {
	return nil, f.err
}

// 实时指标未接入不是服务端故障：返回 200 且 supported=false，不计入 5xx 错误率
func TestGetStreamMetrics_NotSupported_ReturnsSupportedFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLiveStreamHandler(&fakeMetricsLiveService{err: live.ErrStreamMetricsNotSupported})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "streamName", Value: "82"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/live/streams/82/metrics", nil)
	h.GetStreamMetrics(c)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Supported *bool `json:"supported"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Supported == nil || *resp.Data.Supported {
		t.Errorf("期望 supported=false，实际 %s", w.Body.String())
	}
}
