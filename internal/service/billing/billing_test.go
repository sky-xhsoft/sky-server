package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// fakeLive 一个腾讯云账号
type fakeLive struct {
	online  map[string]bool
	domains []string
	flux    map[string]map[string]float64 // 域名 → 流 → MB
	billing map[string]float64            // 域名 → MB
	err     error
}

func (f *fakeLive) OnlineStreams(ctx context.Context) (map[string]bool, error) {
	return f.online, f.err
}
func (f *fakeLive) PlayDomains(ctx context.Context) ([]string, error) { return f.domains, f.err }
func (f *fakeLive) DayStreamFlux(ctx context.Context, d, day string) (map[string]float64, error) {
	return f.flux[d], f.err
}
func (f *fakeLive) DomainBillingFlux(ctx context.Context, d, day string) (float64, error) {
	return f.billing[d], f.err
}

type fakeAccounts struct {
	own  map[uint]bool
	apis map[string]*fakeLive
}

func (a *fakeAccounts) Of(ctx context.Context, companyID uint) string {
	if a.own[companyID] {
		return "company:" + string(rune('0'+companyID%10))
	}
	return GlobalAccount
}
func (a *fakeAccounts) API(ctx context.Context, account string) (LiveAPI, error) {
	if f, ok := a.apis[account]; ok {
		return f, nil
	}
	return nil, errors.New("no api")
}

var ctx = context.Background()

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func setup(t *testing.T) (*Meter, *fakeAccounts, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.LiveCallbackEvent{},
		&entity.BillLiveSession{}, &entity.BillUsageDaily{}, &entity.BillFluxReconcile{}, &entity.BillMeterState{}); err != nil {
		t.Fatal(err)
	}
	// 计量只用到直播间的 ID、公司、名称
	mustExec(t, db, "CREATE TABLE live_room (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, ROOM_NAME TEXT, STATUS TEXT, IS_ACTIVE TEXT)")
	// 存储快照只用到公司、大小、有效标记和创建时间
	for _, tbl := range []string{"live_recode", "live_cut"} {
		mustExec(t, db, "CREATE TABLE "+tbl+" (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, FILE_SIZE INTEGER, IS_ACTIVE TEXT, CREATE_TIME DATETIME)")
	}
	mustExec(t, db, "CREATE TABLE cloud_item (ID INTEGER PRIMARY KEY, SYS_COMPANY_ID INTEGER, ITEM_TYPE TEXT, NAME TEXT, OWNER_ID INTEGER, FILE_SIZE INTEGER, IS_ACTIVE TEXT, CREATE_TIME DATETIME)")
	// 直播间 101、102 属于公司 50（平台账号），201 属于公司 1（自有账号 company:1）
	for _, r := range []struct{ id, company uint }{{101, 50}, {102, 50}, {201, 1}} {
		if err := db.Exec("INSERT INTO live_room (ID, SYS_COMPANY_ID, ROOM_NAME, STATUS, IS_ACTIVE) VALUES (?, ?, ?, 'draft', 'Y')", r.id, r.company, "room").Error; err != nil {
			t.Fatal(err)
		}
	}
	acc := &fakeAccounts{own: map[uint]bool{1: true}, apis: map[string]*fakeLive{}}
	m := NewMeter(db, acc)
	return m, acc, db
}

func addEvent(t *testing.T, db *gorm.DB, id int64, typ, stream string, company int64, when string, data string) {
	t.Helper()
	e := entity.LiveCallbackEvent{ID: id, EventType: typ, StreamName: stream, StreamID: stream, SysCompanyID: company, EventTime: at(when).Unix(), EventData: data, IsActive: "Y"}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatal(err)
	}
}

func sessions(t *testing.T, db *gorm.DB) []entity.BillLiveSession {
	var out []entity.BillLiveSession
	db.Order("START_TIME").Find(&out)
	return out
}

func TestProcessEvents(t *testing.T) {
	m, _, db := setup(t)
	addEvent(t, db, 1, "push_stream", "101", 50, "2026-10-10 20:00:00", "")
	addEvent(t, db, 2, "disconnect_stream", "101", 50, "2026-10-10 20:30:10", "")
	// 重连：没收到断流又推流，上一段在新推流时结束
	addEvent(t, db, 3, "push_stream", "102", 50, "2026-10-10 21:00:00", "")
	addEvent(t, db, 4, "push_stream", "102", 50, "2026-10-10 21:10:00", "")
	addEvent(t, db, 5, "disconnect_stream", "102", 50, "2026-10-10 21:20:00", "")
	// 漏了推流回调：用断流回调里的时长补一段
	addEvent(t, db, 6, "disconnect_stream", "201", 1, "2026-10-10 22:00:00", `{"duration":600}`)
	// 不是直播间的流、没有公司的事件忽略
	addEvent(t, db, 7, "push_stream", "test_abc", 50, "2026-10-10 22:00:00", "")

	n, err := m.ProcessEvents(ctx)
	if err != nil || n != 7 {
		t.Fatalf("处理事件 %d %v", n, err)
	}
	got := sessions(t, db)
	if len(got) != 4 {
		t.Fatalf("应有 4 段推流，got %d", len(got))
	}
	want := []int64{1810, 600, 600, 600}
	for i, s := range got {
		if s.EndTime == nil || s.Seconds != want[i] {
			t.Fatalf("第 %d 段 %+v，期望 %d 秒", i, s, want[i])
		}
	}
	if got[3].RoomID != 201 || !got[3].StartTime.Equal(at("2026-10-10 21:50:00")) {
		t.Fatalf("断流补出的段落：%+v", got[3])
	}
	// 再次处理不会重复
	if n, _ := m.ProcessEvents(ctx); n != 0 || len(sessions(t, db)) != 4 {
		t.Fatal("游标应避免重复处理")
	}
}

func TestPatrol(t *testing.T) {
	m, acc, db := setup(t)
	acc.apis[GlobalAccount] = &fakeLive{online: map[string]bool{"101": true}}
	acc.apis["company:1"] = &fakeLive{err: errors.New("密钥无效")}
	addEvent(t, db, 1, "push_stream", "101", 50, "2026-10-10 20:00:00", "")
	addEvent(t, db, 2, "push_stream", "102", 50, "2026-10-10 20:00:00", "")
	addEvent(t, db, 3, "push_stream", "201", 1, "2026-10-09 10:00:00", "")
	m.ProcessEvents(ctx)

	m.now = func() time.Time { return at("2026-10-10 20:05:00") }
	m.Patrol(ctx)
	got := sessions(t, db)
	if got[1].EndTime != nil || got[2].EndTime != nil {
		t.Fatal("不在线未满 10 分钟不结束")
	}
	m.now = func() time.Time { return at("2026-10-10 20:20:00") }
	m.Patrol(ctx)
	got = sessions(t, db)
	// got 按开始时间：201（10-09）、101、102
	if got[1].EndTime != nil || !got[1].LastSeenTime.Equal(at("2026-10-10 20:20:00")) {
		t.Fatalf("在线的流更新最后在线时间：%+v", got[1])
	}
	if got[2].EndTime == nil || got[2].EndSource != "patrol" || got[2].Seconds != 0 {
		t.Fatalf("不在线超过 10 分钟按最后在线时间结束：%+v", got[2])
	}
	if got[0].EndTime == nil || got[0].EndSource != "cap" || got[0].Seconds != 86400 {
		t.Fatalf("无法巡检时最长 24 小时：%+v", got[0])
	}
}

func usage(t *testing.T, db *gorm.DB, meter, day string) map[uint]entity.BillUsageDaily {
	var rows []entity.BillUsageDaily
	db.Where("METER = ? AND USAGE_DATE = ?", meter, day).Find(&rows)
	out := map[uint]entity.BillUsageDaily{}
	for _, r := range rows {
		out[r.RoomID] = r
	}
	return out
}

func TestComputeDuration(t *testing.T) {
	m, _, db := setup(t)
	end := func(s string) *time.Time { v := at(s); return &v }
	for i, s := range []entity.BillLiveSession{
		// 跨零点：10-10 占 10 分钟，10-11 占 20 分钟 30 秒 → 21 分钟
		{SysCompanyID: 50, RoomID: 101, StreamName: "101", StartTime: at("2026-10-10 23:50:00"), EndTime: end("2026-10-11 00:20:30"), Seconds: 1830},
		// 同一天第二段 90 秒 → 2 分钟
		{SysCompanyID: 50, RoomID: 101, StreamName: "101", StartTime: at("2026-10-11 08:00:00"), EndTime: end("2026-10-11 08:01:30"), Seconds: 90},
		// 不足 60 秒的测试推流不计
		{SysCompanyID: 50, RoomID: 102, StreamName: "102", StartTime: at("2026-10-11 09:00:00"), EndTime: end("2026-10-11 09:00:40"), Seconds: 40},
		// 未结束：算到最后在线时间
		{SysCompanyID: 1, RoomID: 201, StreamName: "201", StartTime: at("2026-10-11 10:00:00"), LastSeenTime: at("2026-10-11 10:05:00")},
	} {
		s.PushEventID = int64(i + 1)
		if s.LastSeenTime.IsZero() {
			s.LastSeenTime = s.StartTime
		}
		db.Create(&s)
	}
	if err := m.ComputeDuration(ctx, "2026-10-11"); err != nil {
		t.Fatal(err)
	}
	got := usage(t, db, entity.MeterLiveDuration, "2026-10-11")
	if got[101].Quantity != 23 || got[101].RawQuantity != 1320 {
		t.Fatalf("直播间 101：%+v", got[101])
	}
	if _, ok := got[102]; ok {
		t.Fatal("不足 60 秒不计")
	}
	page, _ := m.Sessions(ctx, 50, 102, "2026-10", 1, 20)
	if len(page.Items) != 1 || page.Items[0].Counted {
		t.Fatalf("不足 60 秒的段落标记为未计入：%+v", page.Items)
	}
	page, _ = m.Sessions(ctx, 1, 201, "2026-10", 1, 20)
	if len(page.Items) != 1 || !page.Items[0].Counted {
		t.Fatalf("进行中的段落计入：%+v", page.Items)
	}
	if got[201].Quantity != 5 || got[201].SysCompanyID != 1 {
		t.Fatalf("未结束的段落：%+v", got[201])
	}
	m.ComputeDuration(ctx, "2026-10-10")
	if q := usage(t, db, entity.MeterLiveDuration, "2026-10-10")[101].Quantity; q != 10 {
		t.Fatalf("10-10 应为 10 分钟，got %v", q)
	}
	// 重算替换旧值
	db.Where("ROOM_ID = ?", 201).Delete(&entity.BillLiveSession{})
	m.ComputeDuration(ctx, "2026-10-11")
	if _, ok := usage(t, db, entity.MeterLiveDuration, "2026-10-11")[201]; ok {
		t.Fatal("重算应删除旧值")
	}
}

func TestComputeFlux(t *testing.T) {
	m, acc, db := setup(t)
	acc.apis[GlobalAccount] = &fakeLive{
		domains: []string{"play.a.com"},
		// 监控流量 1000 MB（含转码流和找不到直播间的流），计费流量 1100 MB → 按 1.1 分摊
		flux:    map[string]map[string]float64{"play.a.com": {"101": 600, "101_hd": 200, "102": 100, "999": 50, "201": 50}},
		billing: map[string]float64{"play.a.com": 1100},
	}
	acc.apis["company:1"] = &fakeLive{
		domains: []string{"play.own.com"},
		flux:    map[string]map[string]float64{"play.own.com": {"201": 2048}},
		billing: map[string]float64{"play.own.com": 2048},
	}
	if err := m.ComputeFlux(ctx, "2026-10-10"); err != nil {
		t.Fatal(err)
	}
	got := usage(t, db, entity.MeterPlayFlux, "2026-10-10")
	near := func(a, b float64) bool { return a-b < 1e-6 && b-a < 1e-6 }
	if !near(got[101].Quantity, 880.0/1024) || !near(got[101].RawQuantity, 800.0/1024) {
		t.Fatalf("101（含转码流）按比例分摊：%+v", got[101])
	}
	if !near(got[102].Quantity, 110.0/1024) {
		t.Fatalf("102：%+v", got[102])
	}
	// 201 属于自有账号的公司，平台账号上的 201 流不归它
	if got[201].Quantity != 2 || got[201].SysCompanyID != 1 {
		t.Fatalf("自有账号的直播间：%+v", got[201])
	}
	var rec entity.BillFluxReconcile
	db.Where("ACCOUNT = ? AND PLAY_DOMAIN = ?", GlobalAccount, "play.a.com").First(&rec)
	if rec.Status != "diff" || rec.MonitorMB != 1000 || rec.BillingMB != 1100 || rec.UnattributedMB != 110 || rec.StreamCount != 5 {
		t.Fatalf("对账：%+v", rec)
	}

	// 零星流量（低于 1 MB）没有按流数据也不标记差异；达到 1 MB 才标记
	acc.apis[GlobalAccount].domains = []string{"play.a.com", "play.tiny.com", "play.nostream.com"}
	acc.apis[GlobalAccount].billing["play.tiny.com"] = 0.4
	acc.apis[GlobalAccount].billing["play.nostream.com"] = 5
	m.ComputeFlux(ctx, "2026-10-10")
	var tiny, nostream entity.BillFluxReconcile
	db.Where("PLAY_DOMAIN = ?", "play.tiny.com").First(&tiny)
	db.Where("PLAY_DOMAIN = ?", "play.nostream.com").First(&nostream)
	if tiny.Status != "ok" || nostream.Status != "diff" {
		t.Fatalf("零星流量不标记：%s；有计费流量无按流数据标记：%s", tiny.Status, nostream.Status)
	}
	acc.apis[GlobalAccount].domains = []string{"play.a.com"}

	// 自有账号拉取失败：该公司保留原值，平台账号照常更新
	acc.apis["company:1"].err = errors.New("AuthFailure")
	acc.apis[GlobalAccount].billing["play.a.com"] = 2000
	if err := m.ComputeFlux(ctx, "2026-10-10"); err == nil {
		t.Fatal("应返回失败账号的错误")
	}
	got = usage(t, db, entity.MeterPlayFlux, "2026-10-10")
	if got[201].Quantity != 2 {
		t.Fatalf("失败账号的公司保留原值：%+v", got[201])
	}
	if !near(got[101].Quantity, 1600.0/1024) {
		t.Fatalf("平台账号应更新：%+v", got[101])
	}
}

func TestComputeStorageAndSummary(t *testing.T) {
	m, _, db := setup(t)
	gb := int64(1 << 30)
	mustExec(t, db, "INSERT INTO live_recode (SYS_COMPANY_ID, FILE_SIZE, IS_ACTIVE, CREATE_TIME) VALUES (50, ?, 'Y', ?), (50, ?, 'N', ?), (50, ?, 'Y', ?)",
		2*gb, at("2026-10-01 10:00:00"), 5*gb, at("2026-10-01 10:00:00"), 7*gb, at("2026-10-11 10:00:00"))
	mustExec(t, db, "INSERT INTO live_cut (SYS_COMPANY_ID, FILE_SIZE, IS_ACTIVE, CREATE_TIME) VALUES (50, ?, 'Y', ?)", gb/2, at("2026-10-02 10:00:00"))
	mustExec(t, db, "INSERT INTO cloud_item (SYS_COMPANY_ID, ITEM_TYPE, NAME, OWNER_ID, FILE_SIZE, IS_ACTIVE, CREATE_TIME) VALUES (50, 'file', 'a.mp4', 1, ?, 'Y', ?), (50, 'folder', 'f', 1, NULL, 'Y', ?)",
		gb, at("2026-10-03 10:00:00"), at("2026-10-03 10:00:00"))
	for _, day := range []string{"2026-10-09", "2026-10-10"} {
		if err := m.ComputeStorage(ctx, day); err != nil {
			t.Fatal(err)
		}
	}
	var rows []entity.BillUsageDaily
	db.Where("METER = ? AND USAGE_DATE = ?", entity.MeterStorage, "2026-10-10").Find(&rows)
	items := map[string]float64{}
	for _, r := range rows {
		items[r.Item] = r.Quantity
	}
	if items["record"] != 2 || items["clip"] != 0.5 || items["disk"] != 1 {
		t.Fatalf("存储快照（不含已删除、当天之后创建的文件）：%v", items)
	}

	db.Create(&entity.BillUsageDaily{SysCompanyID: 50, UsageDate: "2026-10-10", Meter: entity.MeterLiveDuration, RoomID: 101, Quantity: 30, Unit: "分钟", SourceKey: "k1"})
	db.Create(&entity.BillUsageDaily{SysCompanyID: 50, UsageDate: "2026-10-10", Meter: entity.MeterPlayFlux, RoomID: 101, Quantity: 1.5, Unit: "GB", SourceKey: "k2"})
	m.now = func() time.Time { return at("2026-10-11 12:00:00") }
	sum, err := m.Summary(ctx, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Month != "2026-10" || sum.LiveMinutes != 30 || sum.PlayGB != 1.5 || sum.StorageAvgGB != 3.5 || sum.StorageLatestDay != "2026-10-10" || sum.FluxLatestDay != "2026-10-10" || sum.LiveRooms != 1 {
		t.Fatalf("汇总：%+v", sum)
	}
	days, _ := m.Daily(ctx, 50, "2026-10")
	if len(days) != 31 || days[9].LiveMinutes != 30 || days[9].StorageGB != 3.5 {
		t.Fatalf("每日：%d %+v", len(days), days[9])
	}
	db.Create(&entity.BillFluxReconcile{UsageDate: "2026-10-11", Account: GlobalAccount, PlayDomain: "p", Status: "ok"})
	db.Create(&entity.BillFluxReconcile{UsageDate: "2026-10-09", Account: GlobalAccount, PlayDomain: "p", Status: "error"})
	if sum, _ = m.Summary(ctx, 50, "2026-10"); sum.FluxDays != 1 || sum.FluxLatestDay != "2026-10-11" {
		t.Fatalf("已计量流量的天数以对账记录为准（不含失败）：%+v", sum)
	}
	if apr, _ := m.Summary(ctx, 50, "2026-04"); apr.FluxDays != 0 || apr.FluxLatestDay != "" {
		t.Fatalf("没有计量过的月份：%+v", apr)
	}
	if other, _ := m.Summary(ctx, 51, "2026-10"); other.LiveMinutes != 0 || other.StorageAvgGB != 0 {
		t.Fatal("其他公司看不到")
	}
}

func TestRecomputeValidation(t *testing.T) {
	m, _, _ := setup(t)
	m.now = func() time.Time { return at("2026-10-11 12:00:00") }
	for _, c := range [][2]string{{"2026-10-05", "2026-10-01"}, {"2026-08-01", "2026-10-01"}, {"2026-10-01", "2026-10-20"}, {"bad", "2026-10-01"}} {
		if err := m.Recompute(ctx, c[0], c[1], []string{entity.MeterLiveDuration}); err == nil {
			t.Fatalf("应拒绝 %v", c)
		}
	}
	if err := m.Recompute(ctx, "2026-10-01", "2026-10-02", []string{"flux"}); err == nil {
		t.Fatal("应拒绝不支持的计费点")
	}
	if err := m.Recompute(ctx, "2026-10-01", "2026-10-10", []string{entity.MeterLiveDuration, entity.MeterStorage}); err != nil {
		t.Fatal(err)
	}
}

func TestRoomIDOfStream(t *testing.T) {
	for s, want := range map[string]uint{"101": 101, "101_hd": 101, "abc": 0, "": 0, "12a": 0, "99999999999": 0} {
		if got := roomIDOfStream(s); got != want {
			t.Fatalf("%q → %d，期望 %d", s, got, want)
		}
	}
}

func TestDisconnectAfterPatrolAndDuplicate(t *testing.T) {
	m, acc, db := setup(t)
	acc.apis[GlobalAccount] = &fakeLive{online: map[string]bool{}}
	addEvent(t, db, 1, "push_stream", "101", 50, "2026-10-10 20:00:00", "")
	m.ProcessEvents(ctx)
	// 回调迟到：巡检先按最后在线时间结束
	m.now = func() time.Time { return at("2026-10-10 20:15:00") }
	m.Patrol(ctx)
	if s := sessions(t, db)[0]; s.EndSource != "patrol" {
		t.Fatalf("应先由巡检结束：%+v", s)
	}
	// 迟到的断流回调：校正结束时间，不另补一段
	addEvent(t, db, 2, "disconnect_stream", "101", 50, "2026-10-10 20:12:00", `{"duration":720}`)
	// 重复的断流回调：忽略
	addEvent(t, db, 3, "disconnect_stream", "101", 50, "2026-10-10 20:12:00", `{"duration":720}`)
	m.ProcessEvents(ctx)
	got := sessions(t, db)
	if len(got) != 1 || got[0].EndSource != "callback" || got[0].Seconds != 720 {
		t.Fatalf("迟到与重复的断流不应重复计时：%d %+v", len(got), got)
	}
	// 之后真的漏了推流回调的另一段，照常补
	addEvent(t, db, 4, "disconnect_stream", "101", 50, "2026-10-10 23:00:00", `{"duration":600}`)
	m.ProcessEvents(ctx)
	if got := sessions(t, db); len(got) != 2 || got[1].Seconds != 600 {
		t.Fatalf("不重叠的断流照常补一段：%+v", got)
	}
}

func TestTickDaily(t *testing.T) {
	m, acc, db := setup(t)
	acc.apis[GlobalAccount] = &fakeLive{domains: []string{}}
	acc.apis["company:1"] = &fakeLive{err: errors.New("AuthFailure")}
	cfg := SchedulerConfig{DailyHour: 4}
	// 未到 4 点不跑每日计量
	m.now = func() time.Time { return at("2026-10-11 03:30:00") }
	m.Tick(ctx, cfg)
	if m.getState(ctx, dailyDoneKey) != "" {
		t.Fatal("4 点前不应执行每日计量")
	}
	// 自有账号拉取失败：重试，达到次数上限后不再重试
	m.now = func() time.Time { return at("2026-10-11 04:05:00") }
	for i := 1; i < dailyMaxAttempts; i++ {
		m.Tick(ctx, cfg)
		if m.getState(ctx, dailyDoneKey) != "" {
			t.Fatalf("第 %d 次失败后应继续重试", i)
		}
	}
	m.Tick(ctx, cfg)
	if m.getState(ctx, dailyDoneKey) != "2026-10-10" {
		t.Fatal("达到重试上限后应标记完成，不再重试")
	}
	var rec entity.BillFluxReconcile
	db.Where("ACCOUNT = ?", "company:1").First(&rec)
	if rec.ID != 0 {
		t.Fatalf("没有播放域名时不应有对账记录：%+v", rec)
	}
	// 失败原因能在对账中看到：域名拉取失败记录为 error
	acc.apis["company:1"] = &fakeLive{domains: []string{"play.own.com"}, err: nil}
	acc.apis["company:1"].flux = map[string]map[string]float64{}
	if err := m.ComputeFlux(ctx, "2026-10-10"); err != nil {
		t.Fatal(err)
	}
}

func TestRoomsDeletedRoomName(t *testing.T) {
	m, _, db := setup(t)
	m.now = func() time.Time { return at("2026-10-11 12:00:00") }
	// 直播间 301 已删除，名称取推流回调记录
	addEvent(t, db, 1, "push_stream", "301", 50, "2026-10-10 20:00:00", "")
	db.Model(&entity.LiveCallbackEvent{}).Where("ID = 1").Update("ROOM_NAME", "旧发布会")
	db.Create(&entity.BillUsageDaily{SysCompanyID: 50, UsageDate: "2026-10-10", Meter: entity.MeterLiveDuration, RoomID: 301, Quantity: 5, Unit: "分钟", SourceKey: "d1"})
	db.Create(&entity.BillUsageDaily{SysCompanyID: 50, UsageDate: "2026-10-10", Meter: entity.MeterLiveDuration, RoomID: 101, Quantity: 9, Unit: "分钟", SourceKey: "d2"})
	rooms, err := m.Rooms(ctx, 50, "2026-10")
	if err != nil || len(rooms) != 2 {
		t.Fatalf("%v %+v", err, rooms)
	}
	if rooms[0].RoomID != 101 || rooms[0].Deleted || rooms[0].RoomName != "room" {
		t.Fatalf("现有直播间：%+v", rooms[0])
	}
	if rooms[1].RoomID != 301 || !rooms[1].Deleted || rooms[1].RoomName != "旧发布会" {
		t.Fatalf("已删除的直播间：%+v", rooms[1])
	}
}
