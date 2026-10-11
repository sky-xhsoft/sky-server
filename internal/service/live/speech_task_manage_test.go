package live

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
	"gorm.io/gorm"
)

// fakeSpeechAPI 内存中的火山同传接口（共用账号，含其他公司的任务）
type fakeSpeechAPI struct {
	status   map[string]string
	deletes  []volcengine.DeleteSpeechTaskRequest
	restarts []string
	searches int
}

func (f *fakeSpeechAPI) CreateSpeechTask(ctx context.Context, req *volcengine.CreateSpeechTaskRequest) (string, error) {
	return "new", nil
}
func (f *fakeSpeechAPI) DeleteSpeechTask(ctx context.Context, req *volcengine.DeleteSpeechTaskRequest) error {
	f.deletes = append(f.deletes, *req)
	return nil
}
func (f *fakeSpeechAPI) GetSpeechTask(ctx context.Context, id string) (*volcengine.SpeechTaskInfo, error) {
	return &volcengine.SpeechTaskInfo{TaskID: id, Status: f.status[id]}, nil
}
func (f *fakeSpeechAPI) SearchSpeechTask(ctx context.Context, req *volcengine.SearchSpeechTaskRequest) (int, []volcengine.SpeechTaskInfo, error) {
	f.searches++
	var out []volcengine.SpeechTaskInfo
	for id, st := range f.status {
		out = append(out, volcengine.SpeechTaskInfo{TaskID: id, Status: st})
	}
	return len(out), out, nil
}
func (f *fakeSpeechAPI) RestartSpeechTask(ctx context.Context, id string, reload bool) error {
	f.restarts = append(f.restarts, id)
	return nil
}

func setupSpeechManage(t *testing.T) (*speechTaskService, *fakeSpeechAPI) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&SpeechTaskRecord{}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []SpeechTaskRecord{
		{ID: 1, CompanyID: 50, TaskID: "t-run", Name: "发布会同传", RoomName: "主题大会", Status: "created", CreateTime: "2026-10-11 16:00:00", RuleJSON: `{"Pipe":"si"}`},
		{ID: 2, CompanyID: 50, TaskID: "t-done", Name: "昨天的任务", RoomName: "音乐节", Status: "running", CreateTime: "2026-10-10 10:00:00"},
		{ID: 3, CompanyID: 50, TaskID: "t-gone", Name: "已删除", RoomName: "音乐节", Status: "running", CreateTime: "2026-10-10 09:00:00"},
		{ID: 4, CompanyID: 51, TaskID: "t-other", Name: "其他公司", Status: "running", CreateTime: "2026-10-11 16:00:00"},
	} {
		db.Create(&r)
	}
	api := &fakeSpeechAPI{status: map[string]string{"t-run": "running", "t-done": "done", "t-other": "running"}}
	return &speechTaskService{db: db, api: api}, api
}

func TestListSpeechTasksPaged_SyncsStatus(t *testing.T) {
	svc, api := setupSpeechManage(t)
	page, err := svc.ListSpeechTasksPaged(context.Background(), 50, SpeechTaskQuery{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || api.searches != 1 {
		t.Fatalf("只返回本公司任务，并只搜索一次火山：total=%d searches=%d", page.Total, api.searches)
	}
	got := map[string]string{}
	for _, it := range page.Items {
		got[it.TaskID] = it.Status
	}
	if got["t-run"] != "running" || got["t-done"] != "done" || got["t-gone"] != "done" {
		t.Fatalf("状态应同步自火山（搜不到的视为已结束）：%v", got)
	}
	// 写回本地，已结束的任务下次不再同步
	page, _ = svc.ListSpeechTasksPaged(context.Background(), 50, SpeechTaskQuery{Status: "active", Page: 1, PageSize: 20})
	if page.Total != 1 || page.Items[0].TaskID != "t-run" {
		t.Fatalf("active 只返回未结束的任务：%+v", page.Items)
	}
	page, _ = svc.ListSpeechTasksPaged(context.Background(), 50, SpeechTaskQuery{Keyword: "音乐节", Page: 1, PageSize: 1})
	if page.Total != 2 || len(page.Items) != 1 {
		t.Fatalf("按直播间名称搜索并分页：total=%d items=%d", page.Total, len(page.Items))
	}
}

func TestStopRemoveRestart(t *testing.T) {
	svc, api := setupSpeechManage(t)
	ctx := context.Background()

	if err := svc.StopSpeechTask(ctx, 50, "t-other"); !errors.Is(err, ErrSpeechTaskNotFound) {
		t.Fatalf("不能停止其他公司的任务，got %v", err)
	}
	if err := svc.StopSpeechTask(ctx, 50, "t-run"); err != nil {
		t.Fatal(err)
	}
	if len(api.deletes) != 1 || api.deletes[0].WithRecord {
		t.Fatalf("停止应保留火山历史（WithRecord=false）：%+v", api.deletes)
	}
	var r SpeechTaskRecord
	svc.db.Where("TASK_ID = ?", "t-run").First(&r)
	if r.Status != SpeechStatusPrestop {
		t.Fatalf("停止后本地记录保留并标记关闭中，got %s", r.Status)
	}

	if err := svc.RestartSpeechTask(ctx, 50, "t-done", false); err == nil {
		t.Fatal("已结束的任务不能重启")
	}
	api.status["t-gone"] = "running"
	if err := svc.RestartSpeechTask(ctx, 50, "t-gone", true); err != nil || len(api.restarts) != 1 {
		t.Fatalf("运行中的任务应可重启：%v %v", err, api.restarts)
	}

	if err := svc.RemoveSpeechTask(ctx, 50, "t-done"); err != nil {
		t.Fatal(err)
	}
	if !api.deletes[len(api.deletes)-1].WithRecord {
		t.Fatal("彻底删除应连火山历史一起删除（WithRecord=true）")
	}
	var n int64
	svc.db.Model(&SpeechTaskRecord{}).Where("TASK_ID = ?", "t-done").Count(&n)
	if n != 0 {
		t.Fatal("彻底删除应删除本地记录")
	}

	rule, err := svc.GetSpeechTaskRule(ctx, 50, "t-run")
	if err != nil || string(rule) != `{"Pipe":"si"}` {
		t.Fatalf("复制新建应返回创建参数：%s %v", rule, err)
	}
	if _, err := svc.GetSpeechTaskRule(ctx, 50, "t-other"); err == nil {
		t.Fatal("不能读取其他公司的任务参数")
	}
}
