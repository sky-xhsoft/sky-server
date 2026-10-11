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
	old, oldNow := time.Local, taskNow
	time.Local = loc
	// 固定「当前时间」为 2026-10-10 08:00（北京时间），用例中的任务时间都在其之后
	taskNow = func() time.Time { return time.Date(2026, 10, 10, 8, 0, 0, 0, loc) }
	t.Cleanup(func() { time.Local, taskNow = old, oldNow })
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
	// 按北京时间保存：14:00（北京）即 06:00 UTC，不能当成 14:00 UTC（入库后会多出 8 小时）
	if got := store.saved[0].StartTime.UTC().Format(taskTimeLayout); got != "2026-10-10 06:00:00" {
		t.Fatalf("保存的开始时间应为北京时间，UTC 应为 06:00，got %s", got)
	}
}

func TestNormalizeTaskTimeRange_TencentLimits(t *testing.T) {
	withShanghai(t)
	cases := []struct {
		start, end string
		wantErr    string
	}{
		{"2026-10-10 13:45:11", "2026-11-11 13:45:11", "不能超过 30 天"}, // 32 天：腾讯云返回 InvalidTaskTime
		{"2026-10-10 09:00:00", "2026-11-09 09:00:00", "不能超过 30 天"}, // 恰好 30 天也不行（须小于 30 天）
		{"2026-10-09 09:00:00", "2026-10-10 07:00:00", "晚于当前时间"},
		{"2026-10-10 09:00:00", "2026-10-24 09:00:00", ""},
		{"2026-10-09 13:49:55", "2026-10-23 13:49:55", ""}, // 开始时间已过、结束时间在将来：腾讯云允许，立即开始
	}
	for _, c := range cases {
		s, e := c.start, c.end
		err := normalizeTaskTimeRange(&s, &e)
		if c.wantErr == "" {
			if err != nil {
				t.Errorf("%s ~ %s 不应报错：%v", c.start, c.end, err)
			}
		} else if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s ~ %s 应报错「%s」，got %v", c.start, c.end, c.wantErr, err)
		}
	}
}

func TestCreatePullStreamTask_Over30DaysDoesNotCallTencent(t *testing.T) {
	withShanghai(t)
	lv, store := &fakePullLive{}, &fakePullStore{}
	w := postCreate(NewPullStreamHandler(lv, store), createBody("2026-10-10 13:45:11", "2026-11-11 13:45:11"))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "30 天") {
		t.Fatalf("超过 30 天应返回 400 并提示，got %d %s", w.Code, w.Body.String())
	}
	if len(lv.created) != 0 {
		t.Fatal("超过 30 天时不应调用腾讯云")
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
