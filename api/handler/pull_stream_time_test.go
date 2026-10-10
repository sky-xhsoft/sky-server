package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// fakePullLive 记录发给腾讯云的创建请求（不真正调用腾讯云）
type fakePullLive struct {
	liveService.Service
	created []*live.CreatePullStreamTaskRequest
}

func (f *fakePullLive) CreatePullStreamTask(ctx context.Context, req *live.CreatePullStreamTaskRequest) (string, error) {
	f.created = append(f.created, req)
	return "task-1", nil
}

type fakePullStore struct {
	liveService.PullStreamTaskService
	saved []*entity.PullStreamTask
}

func (f *fakePullStore) CreatePullStreamTask(ctx context.Context, task *entity.PullStreamTask) error {
	f.saved = append(f.saved, task)
	return nil
}

func withShanghai(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skip("缺少时区数据")
	}
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
}

func TestNormalizeTaskTime(t *testing.T) {
	withShanghai(t)
	cases := map[string]string{
		"2026-10-10 14:00:00":       "2026-10-10 14:00:00",
		"2026-10-10T14:00:00+08:00": "2026-10-10 14:00:00", // 直播间时间
		"2026-10-10T06:00:00Z":      "2026-10-10 14:00:00", // UTC 转北京时间
		"2026-10-10T14:00:00":       "2026-10-10 14:00:00",
	}
	for in, want := range cases {
		if got, err := normalizeTaskTime(in); err != nil || got != want {
			t.Errorf("%q → %q, %v；期望 %q", in, got, err, want)
		}
	}
	if _, err := normalizeTaskTime("2026/10/10"); err == nil {
		t.Error("非法格式应报错")
	}
	s, e := "2026-10-10 15:00:00", "2026-10-10T14:00:00+08:00"
	if err := normalizeTaskTimeRange(&s, &e); err == nil {
		t.Error("结束时间早于开始时间应报错")
	}
}

func postCreate(h *PullStreamHandler, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/tasks", h.CreatePullStreamTask)
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const createBodyTpl = `{"sourceType":"PullLivePushLive","sourceUrls":["rtmp://src/live/a"],"toUrl":"rtmps://dst/app/key",
"startTime":"%START%","endTime":"%END%","operator":"admin","comment":"t"}`

func createBody(start, end string) string {
	return strings.NewReplacer("%START%", start, "%END%", end).Replace(createBodyTpl)
}

func TestCreatePullStreamTask_RoomTimeFormat(t *testing.T) {
	withShanghai(t)
	lv, store := &fakePullLive{}, &fakePullStore{}
	w := postCreate(NewPullStreamHandler(lv, store), createBody("2026-10-10T14:00:00+08:00", "2026-10-10T16:30:00+08:00"))
	if w.Code != http.StatusOK {
		t.Fatalf("使用直播间时间格式应创建成功，got %d %s", w.Code, w.Body.String())
	}
	if len(lv.created) != 1 || lv.created[0].StartTime != "2026-10-10 14:00:00" || lv.created[0].EndTime != "2026-10-10 16:30:00" {
		t.Fatalf("发给腾讯云的时间应统一为北京时间格式，got %+v", lv.created)
	}
	if len(store.saved) != 1 || store.saved[0].StartTime.Format(taskTimeLayout) != "2026-10-10 14:00:00" {
		t.Fatalf("本地应保存任务，got %+v", store.saved)
	}
}

func TestCreatePullStreamTask_BadTimeDoesNotCallTencent(t *testing.T) {
	lv, store := &fakePullLive{}, &fakePullStore{}
	w := postCreate(NewPullStreamHandler(lv, store), createBody("2026/10/10 14:00", "2026-10-10 16:00:00"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("时间格式错误应返回 400，got %d", w.Code)
	}
	if len(lv.created) != 0 || len(store.saved) != 0 {
		t.Fatal("时间格式错误时不应调用腾讯云，也不应保存")
	}
}
