package live

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"gorm.io/gorm"
)

// fakeCloud 内存中的腾讯云模板与规则（不调用真实接口）
type fakeCloud struct {
	callbacks map[int64]*tencentLive.CallbackTemplate
	records   map[int64]*tencentLive.RecordTemplate
	cbRules   []*tencentLive.TemplateRule
	recRules  []*tencentLive.TemplateRule
	nextID    int64
	fail      map[string]error // 按方法名注入错误
	calls     []string
}

func newFakeCloud() *fakeCloud {
	return &fakeCloud{callbacks: map[int64]*tencentLive.CallbackTemplate{}, records: map[int64]*tencentLive.RecordTemplate{}, nextID: 1000, fail: map[string]error{}}
}

func (f *fakeCloud) call(name string) error {
	f.calls = append(f.calls, name)
	return f.fail[name]
}

func (f *fakeCloud) called(name string) int {
	n := 0
	for _, c := range f.calls {
		if c == name {
			n++
		}
	}
	return n
}

func (f *fakeCloud) ListCallbackTemplates() ([]*tencentLive.CallbackTemplate, error) {
	if err := f.call("ListCallbackTemplates"); err != nil {
		return nil, err
	}
	var out []*tencentLive.CallbackTemplate
	for _, t := range f.callbacks {
		c := *t
		out = append(out, &c)
	}
	return out, nil
}
func (f *fakeCloud) CreateCallbackTemplate(t *tencentLive.CallbackTemplate) (int64, error) {
	if err := f.call("CreateCallbackTemplate"); err != nil {
		return 0, err
	}
	f.nextID++
	c := *t
	c.TemplateID = f.nextID
	f.callbacks[c.TemplateID] = &c
	return c.TemplateID, nil
}
func (f *fakeCloud) ModifyCallbackTemplate(t *tencentLive.CallbackTemplate) error {
	if err := f.call("ModifyCallbackTemplate"); err != nil {
		return err
	}
	c := *t
	f.callbacks[t.TemplateID] = &c
	return nil
}
func (f *fakeCloud) DeleteCallbackTemplate(id int64) error {
	if err := f.call("DeleteCallbackTemplate"); err != nil {
		return err
	}
	delete(f.callbacks, id)
	return nil
}
func (f *fakeCloud) ListCallbackRules() ([]*tencentLive.TemplateRule, error) {
	return f.cbRules, f.call("ListCallbackRules")
}
func (f *fakeCloud) CreateCallbackRule(domain, app string, id int64) error {
	if err := f.call("CreateCallbackRule"); err != nil {
		return err
	}
	f.cbRules = append(f.cbRules, &tencentLive.TemplateRule{DomainName: domain, AppName: app, TemplateID: id})
	return nil
}
func (f *fakeCloud) DeleteCallbackRule(domain, app string) error {
	if err := f.call("DeleteCallbackRule"); err != nil {
		return err
	}
	out := f.cbRules[:0]
	for _, r := range f.cbRules {
		if !(r.DomainName == domain && r.AppName == app) {
			out = append(out, r)
		}
	}
	f.cbRules = out
	return nil
}
func (f *fakeCloud) ListRecordTemplates() ([]*tencentLive.RecordTemplate, error) {
	if err := f.call("ListRecordTemplates"); err != nil {
		return nil, err
	}
	var out []*tencentLive.RecordTemplate
	for _, t := range f.records {
		c := *t
		out = append(out, &c)
	}
	return out, nil
}
func (f *fakeCloud) CreateRecordTemplate(t *tencentLive.RecordTemplate) (int64, error) {
	if err := f.call("CreateRecordTemplate"); err != nil {
		return 0, err
	}
	f.nextID++
	c := *t
	c.TemplateID = f.nextID
	c.CosStore = true
	f.records[c.TemplateID] = &c
	return c.TemplateID, nil
}
func (f *fakeCloud) ModifyRecordTemplate(t *tencentLive.RecordTemplate) error {
	if err := f.call("ModifyRecordTemplate"); err != nil {
		return err
	}
	c := *t
	f.records[t.TemplateID] = &c
	return nil
}
func (f *fakeCloud) DeleteRecordTemplate(id int64) error {
	if err := f.call("DeleteRecordTemplate"); err != nil {
		return err
	}
	delete(f.records, id)
	return nil
}
func (f *fakeCloud) ListRecordRules() ([]*tencentLive.TemplateRule, error) {
	return f.recRules, f.call("ListRecordRules")
}
func (f *fakeCloud) CreateRecordRule(domain, app, stream string, id int64) error {
	if err := f.call("CreateRecordRule"); err != nil {
		return err
	}
	f.recRules = append(f.recRules, &tencentLive.TemplateRule{DomainName: domain, AppName: app, StreamName: stream, TemplateID: id})
	return nil
}
func (f *fakeCloud) DeleteRecordRule(domain, app, stream string) error {
	if err := f.call("DeleteRecordRule"); err != nil {
		return err
	}
	out := f.recRules[:0]
	for _, r := range f.recRules {
		if !(r.DomainName == domain && r.AppName == app && r.StreamName == stream) {
			out = append(out, r)
		}
	}
	f.recRules = out
	return nil
}

// fakeLiveForTemplate 直播服务：只实现模板客户端与域名查询
type fakeLiveForTemplate struct {
	Service
	cloud      *fakeCloud
	noTencent  bool
	pushDomain map[string]int64 // 域名 → 类型（0 推流，1 播放）
	buckets    []tencent.CosBucket
}

func (f *fakeLiveForTemplate) ListCosBuckets(ctx context.Context) ([]tencent.CosBucket, error) {
	if f.noTencent {
		return nil, ErrNoCompanyTencent
	}
	return f.buckets, nil
}

func (f *fakeLiveForTemplate) TemplateClient(ctx context.Context) (tencentLive.TemplateClient, error) {
	if f.noTencent {
		return nil, ErrNoCompanyTencent
	}
	return f.cloud, nil
}

func (f *fakeLiveForTemplate) DescribeDomain(ctx context.Context, domain string) (*tencentLive.DomainInfo, error) {
	t, ok := f.pushDomain[domain]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &tencentLive.DomainInfo{Name: domain, Type: t}, nil
}

const testBase = "http://web.zigebo.com"

func setupTemplateTest(t *testing.T) (*templateService, *fakeLiveForTemplate, context.Context) {
	t.Helper()
	fieldcrypto.Setup("test-template-key")
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.LiveCallbackTemplate{}, &entity.LiveRecordTemplate{}, &entity.LiveTemplateRule{}, &entity.LiveCallbackEvent{}, &entity.SysCompanyConf{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&entity.SysCompanyConf{SysCompanyID: 51, TencentCOSBucketName: "zigebo-1503356301", TencentCOSRegion: "ap-shanghai"})
	lv := &fakeLiveForTemplate{cloud: newFakeCloud(), pushDomain: map[string]int64{"upload.skyzhou.cn": 0, "live.skyzhou.cn": 1}}
	svc := NewTemplateService(db, lv, testBase+"/").(*templateService)
	ctx := context.WithValue(context.Background(), contextutil.CompanyIDKey, uint(51))
	return svc, lv, ctx
}

func standardCallbackInput() *entity.LiveCallbackTemplate {
	return &entity.LiveCallbackTemplate{Name: "zigebo-callback-test", StreamBeginURL: testBase + callbackPathPush, StreamEndURL: testBase + callbackPathDisconnect}
}

func TestCosBuckets_CompanyBucketFirst(t *testing.T) {
	svc, live, ctx := setupTemplateTest(t)
	live.buckets = []tencent.CosBucket{
		{Name: "backup-1503356301", Region: "ap-guangzhou"},
		{Name: "zigebo-1503356301", Region: "ap-nanjing"},
		{Name: "zigebo-1503356301", Region: "ap-shanghai"},
	}
	list, err := svc.CosBuckets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0].Bucket != "zigebo" || list[0].Region != "ap-shanghai" || !list[0].Company {
		t.Fatalf("公司存储桶应排第一：%+v", list)
	}
	if list[1].Company || list[1].Bucket != "backup" || list[1].FullName != "backup-1503356301" {
		t.Fatalf("其他存储桶：%+v", list[1])
	}

	live.noTencent = true
	if _, err := svc.CosBuckets(ctx); err != ErrNoCompanyTencent {
		t.Fatalf("未配置密钥应提示：%v", err)
	}
}

func TestSaveCallbackTemplate_SyncsToCloud(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	got, err := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if got.SyncStatus != entity.TemplateSyncSynced || got.CloudTemplateID == 0 || got.SysCompanyID != 51 {
		t.Fatalf("应创建到云端并标记已同步，got %+v", got)
	}
	if c := lv.cloud.callbacks[got.CloudTemplateID]; c == nil || c.StreamBeginURL != testBase+callbackPathPush {
		t.Fatalf("云端模板内容不对：%+v", c)
	}
	// 再次保存走修改
	in := standardCallbackInput()
	in.RecordURL = testBase + callbackPathRecordFile
	if _, err := svc.SaveCallbackTemplate(ctx, got.ID, in, "admin"); err != nil {
		t.Fatal(err)
	}
	if lv.cloud.called("ModifyCallbackTemplate") != 1 || lv.cloud.callbacks[got.CloudTemplateID].RecordURL == "" {
		t.Fatal("已同步的模板再次保存应修改云端模板")
	}
}

func TestSaveCallbackTemplate_SyncFailureKeepsLocal(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.cloud.fail["CreateCallbackTemplate"] = tcerr.NewTencentCloudSDKError("InvalidParameter.UrlNotSafe", "bad url", "r1")
	got, err := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")
	if err != nil {
		t.Fatalf("同步失败不应导致保存失败：%v", err)
	}
	if got.ID == 0 || got.SyncStatus != entity.TemplateSyncFailed || !strings.Contains(got.SyncError, "回调地址无法解析") {
		t.Fatalf("应保存本地并记录中文失败原因，got %+v", got)
	}
	// 重新同步成功
	delete(lv.cloud.fail, "CreateCallbackTemplate")
	if err := svc.SyncTemplate(ctx, entity.TemplateKindCallback, got.ID); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := svc.loadCallback(ctx, got.ID)
	if reloaded.SyncStatus != entity.TemplateSyncSynced || reloaded.CloudTemplateID == 0 {
		t.Fatalf("重新同步后应为已同步，got %+v", reloaded)
	}
}

func TestSaveCallbackTemplate_NoCompanyTencent(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.noTencent = true
	got, err := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if got.SyncStatus != entity.TemplateSyncFailed || !strings.Contains(got.SyncError, "未配置腾讯云密钥") {
		t.Fatalf("未配置公司腾讯云时不能写入其他账号，应标记失败，got %+v", got)
	}
}

func TestTemplateValidation(t *testing.T) {
	svc, _, ctx := setupTemplateTest(t)
	bad := []*entity.LiveCallbackTemplate{
		{Name: "带 空格", StreamBeginURL: testBase},
		{Name: "ok-name"},
		{Name: "ok-name", StreamBeginURL: "ftp://x"},
		{Name: "ok-name", StreamBeginURL: testBase, RecordExceptionURL: testBase, RecordExceptionLevels: "fatal"},
	}
	for _, b := range bad {
		if _, err := svc.SaveCallbackTemplate(ctx, 0, b, "admin"); err == nil {
			t.Errorf("应校验失败：%+v", b)
		}
	}
	rec := &entity.LiveRecordTemplate{Name: "rec", CosBucket: "zigebo-1503356301", CosRegion: "ap-shanghai", CosPath: "/video/live/", RecordInterval: 1800}
	if _, err := svc.SaveRecordTemplate(ctx, 0, rec, "admin"); err == nil || !strings.Contains(err.Error(), "占位符") {
		t.Fatalf("存储路径没有占位符应报错，got %v", err)
	}
	rec.CosPath = standardRecordPath
	rec.RecordInterval = 30
	if _, err := svc.SaveRecordTemplate(ctx, 0, rec, "admin"); err == nil {
		t.Fatal("分段时长小于 60 秒应报错")
	}
	rec.RecordInterval = 1800
	got, err := svc.SaveRecordTemplate(ctx, 0, rec, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if got.CosBucket != "zigebo" {
		t.Fatalf("桶名应去掉 -appid，got %s", got.CosBucket)
	}
}

func TestBind_ReplacesDifferentTemplateRule(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.cloud.callbacks[500845] = &tencentLive.CallbackTemplate{TemplateID: 500845, Name: "ALL_Callback", RecordURL: testBase + callbackPathRecordFile}
	lv.cloud.cbRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 500845}}
	tpl, _ := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")

	rule, err := svc.Bind(ctx, &BindTemplateRequest{Kind: entity.TemplateKindCallback, DomainName: "upload.skyzhou.cn", TemplateID: tpl.ID}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if rule.AppName != "live" || rule.SyncStatus != entity.TemplateSyncSynced {
		t.Fatalf("AppName 默认 live 且已同步，got %+v", rule)
	}
	if lv.cloud.called("DeleteCallbackRule") != 1 || len(lv.cloud.cbRules) != 1 || lv.cloud.cbRules[0].TemplateID != tpl.CloudTemplateID {
		t.Fatalf("已有其他模板的规则应先删后建，rules=%+v calls=%v", lv.cloud.cbRules, lv.cloud.calls)
	}
	// 重复绑定同一模板不再调用删除、创建
	if _, err := svc.Bind(ctx, &BindTemplateRequest{Kind: entity.TemplateKindCallback, DomainName: "upload.skyzhou.cn", TemplateID: tpl.ID}, "admin"); err != nil {
		t.Fatal(err)
	}
	if lv.cloud.called("CreateCallbackRule") != 1 {
		t.Fatal("相同模板已绑定时不应重复创建规则")
	}
	// 播放域名不能绑定
	if _, err := svc.Bind(ctx, &BindTemplateRequest{Kind: entity.TemplateKindCallback, DomainName: "live.skyzhou.cn", TemplateID: tpl.ID}, "admin"); err == nil {
		t.Fatal("播放域名不能绑定模板")
	}
}

func TestDeleteTemplate_RequiresUnbind(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	tpl, _ := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")
	if _, err := svc.Bind(ctx, &BindTemplateRequest{Kind: entity.TemplateKindCallback, DomainName: "upload.skyzhou.cn", TemplateID: tpl.ID}, "admin"); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteTemplate(ctx, entity.TemplateKindCallback, tpl.ID); err == nil || !strings.Contains(err.Error(), "请先解绑") {
		t.Fatalf("已绑定的模板不能删除，got %v", err)
	}
	if err := svc.Unbind(ctx, entity.TemplateKindCallback, "upload.skyzhou.cn", "live", ""); err != nil {
		t.Fatal(err)
	}
	if len(lv.cloud.cbRules) != 0 {
		t.Fatal("解绑应删除云端规则")
	}
	if err := svc.DeleteTemplate(ctx, entity.TemplateKindCallback, tpl.ID); err != nil {
		t.Fatal(err)
	}
	if len(lv.cloud.callbacks) != 0 {
		t.Fatal("删除模板应同时删除云端模板")
	}
}

func TestListReconcile_DriftAndPull(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	tpl, _ := svc.SaveCallbackTemplate(ctx, 0, standardCallbackInput(), "admin")
	// 有人在控制台改了断流回调
	lv.cloud.callbacks[tpl.CloudTemplateID].StreamEndURL = "http://other.example.com/end"
	res, err := svc.ListCallbackTemplates(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Items[0].SyncStatus != entity.TemplateSyncDrift || !strings.Contains(res.Items[0].SyncError, "断流回调") {
		t.Fatalf("应识别云端不一致，got %+v", res.Items[0].LiveCallbackTemplate)
	}
	if err := svc.PullTemplate(ctx, entity.TemplateKindCallback, tpl.ID); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := svc.loadCallback(ctx, tpl.ID)
	if reloaded.StreamEndURL != "http://other.example.com/end" || reloaded.SyncStatus != entity.TemplateSyncSynced {
		t.Fatalf("用云端更新本地后应一致，got %+v", reloaded)
	}
}

func TestImport(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.cloud.callbacks[500845] = &tencentLive.CallbackTemplate{TemplateID: 500845, Name: "ALL_Callback", RecordURL: testBase + callbackPathRecordFile}
	lv.cloud.records[1622658] = &tencentLive.RecordTemplate{TemplateID: 1622658, Name: "直播录制", CosStore: true, CosBucket: "zigebo", CosRegion: "ap-shanghai", CosPath: "/video/live/ALL/", RecordInterval: 300, Formats: []string{"mp4"}}
	lv.cloud.cbRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 500845}}
	lv.cloud.recRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 1622658}}
	res, err := svc.Import(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if res.CallbackTemplates != 1 || res.RecordTemplates != 1 || res.Rules != 2 {
		t.Fatalf("导入数量不对：%+v", res)
	}
	// 再导入一次不重复
	res, _ = svc.Import(ctx, "admin")
	if res.CallbackTemplates != 0 || res.RecordTemplates != 0 {
		t.Fatalf("重复导入不应新增模板：%+v", res)
	}
	var n int64
	svc.db.Model(&entity.LiveTemplateRule{}).Count(&n)
	if n != 2 {
		t.Fatalf("规则不应重复，got %d", n)
	}
}

func TestApplyStandard_CompletesExistingCallbackTemplate(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	// upload.skyzhou.cn 现状：ALL_Callback 只有录制文件回调，录制模板 5 分钟一段、存在其他桶
	lv.cloud.callbacks[500845] = &tencentLive.CallbackTemplate{TemplateID: 500845, Name: "ALL_Callback", RecordURL: testBase + callbackPathRecordFile}
	lv.cloud.records[1622658] = &tencentLive.RecordTemplate{TemplateID: 1622658, Name: "直播录制", CosStore: true, CosBucket: "other", CosRegion: "ap-nanjing", CosPath: "/video/live/ALL/", RecordInterval: 300}
	lv.cloud.cbRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 500845}}
	lv.cloud.recRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 1622658}}

	plan, err := svc.StandardPlan(ctx, "upload.skyzhou.cn")
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Callback.Exists || len(plan.Callback.Missing) != 5 || !plan.Record.ReplaceDefault || plan.Record.CosBucket != "zigebo" {
		t.Fatalf("预览不对：%+v", plan)
	}
	steps, err := svc.ApplyStandard(ctx, "upload.skyzhou.cn", StandardOptions{}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range steps {
		if !s.OK {
			t.Fatalf("步骤失败：%+v", s)
		}
	}
	cb := lv.cloud.callbacks[500845]
	if cb.StreamBeginURL != testBase+callbackPathPush || cb.StreamEndURL != testBase+callbackPathDisconnect || cb.RecordURL != testBase+callbackPathRecordFile {
		t.Fatalf("应补全现有回调模板并保留原有录制回调，got %+v", cb)
	}
	if lv.cloud.called("DeleteCallbackRule") != 0 || lv.cloud.called("CreateCallbackTemplate") != 0 {
		t.Fatal("补全模式不应新建回调模板或解绑回调规则（避免丢失控制台配置的高光切片回调）")
	}
	if len(lv.cloud.recRules) != 1 || lv.cloud.recRules[0].TemplateID == 1622658 {
		t.Fatalf("录制规则应替换为标准录制模板：%+v", lv.cloud.recRules)
	}
	std := lv.cloud.records[lv.cloud.recRules[0].TemplateID]
	if std.CosBucket != "zigebo" || std.CosRegion != "ap-shanghai" || std.RecordInterval != 1800 || !strings.HasPrefix(std.CosPath, "/") {
		t.Fatalf("标准录制模板应存储至公司 COS、30 分钟一段，got %+v", std)
	}
	// 体检：已无开播、断流缺失
	ov, err := svc.DomainOverview(ctx, "upload.skyzhou.cn")
	if err != nil {
		t.Fatal(err)
	}
	for _, is := range ov.Issues {
		if strings.Contains(is.Message, "开播") || strings.Contains(is.Message, "不是本公司") {
			t.Fatalf("应用标准配置后不应再有此问题：%s", is.Message)
		}
	}
}

func TestDomainOverview_Issues(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.cloud.callbacks[500845] = &tencentLive.CallbackTemplate{TemplateID: 500845, Name: "ALL_Callback", RecordURL: "http://prod.example.com/api/v1/live/callback/recording-file"}
	lv.cloud.cbRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "live", TemplateID: 500845}}
	ov, err := svc.DomainOverview(ctx, "upload.skyzhou.cn")
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	for _, is := range ov.Issues {
		text += is.Level + ":" + is.Message + "\n"
	}
	for _, want := range []string{"未配置开播、断流回调", "未指向本系统", "未绑定录制模板", "高光切片"} {
		if !strings.Contains(text, want) {
			t.Errorf("体检应包含「%s」，got\n%s", want, text)
		}
	}
	if len(ov.CallbackRules) != 1 || ov.CallbackRules[0].Managed {
		t.Fatalf("未导入的模板应显示为未纳入管理，got %+v", ov.CallbackRules)
	}
}

// 公司 1 的现状：规则 AppName 为空（对整个域名生效）。标准配置应补全该模板、替换录制规则时沿用空 AppName，不能再加一条 live 规则（会重复录制）
func TestApplyStandard_DomainWideRules(t *testing.T) {
	svc, lv, ctx := setupTemplateTest(t)
	lv.cloud.callbacks[502814] = &tencentLive.CallbackTemplate{TemplateID: 502814, Name: "zigebo-pro", StreamBeginURL: testBase + callbackPathPush, StreamEndURL: testBase + callbackPathDisconnect, RecordURL: testBase + callbackPathRecordFile}
	lv.cloud.records[1627402] = &tencentLive.RecordTemplate{TemplateID: 1627402, Name: "zigebo-pro", CosStore: true, CosBucket: "other", CosRegion: "ap-nanjing", CosPath: "/live/{RandomID}", RecordInterval: 43200}
	lv.cloud.cbRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "", TemplateID: 502814}}
	lv.cloud.recRules = []*tencentLive.TemplateRule{{DomainName: "upload.skyzhou.cn", AppName: "", TemplateID: 1627402}}

	plan, err := svc.StandardPlan(ctx, "upload.skyzhou.cn")
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Callback.Exists || plan.Callback.AppName != "" || !plan.Record.Exists {
		t.Fatalf("空 AppName 的规则应视为已有规则：%+v", plan)
	}
	if _, err := svc.ApplyStandard(ctx, "upload.skyzhou.cn", StandardOptions{}, "admin"); err != nil {
		t.Fatal(err)
	}
	if lv.cloud.called("CreateCallbackTemplate") != 0 || lv.cloud.called("CreateCallbackRule") != 0 {
		t.Fatal("已有对整个域名生效的回调规则时，不应新建模板或规则")
	}
	if len(lv.cloud.recRules) != 1 || lv.cloud.recRules[0].AppName != "" || lv.cloud.recRules[0].TemplateID == 1627402 {
		t.Fatalf("录制规则应在原空 AppName 上替换，且只有一条：%+v", lv.cloud.recRules)
	}
	if lv.cloud.callbacks[502814].RecordExceptionURL == "" {
		t.Fatal("应补全缺少的录制异常回调")
	}
}
