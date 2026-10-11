package live

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"gorm.io/gorm"
)

// 本系统的回调接收路径（拼在系统回调地址 live.callbackURL 之后）
const (
	callbackPathPush            = "/api/v1/live/callback/push"
	callbackPathDisconnect      = "/api/v1/live/callback/disconnect"
	callbackPathRecordFile      = "/api/v1/live/callback/recording-file"
	callbackPathRecordStatus    = "/api/v1/live/callback/recording-status"
	callbackPathRecordException = "/api/v1/live/callback/record-exception"
	callbackPathPushException   = "/api/v1/live/callback/push-exception"
	callbackPathHighlight       = "/api/v1/live/callback/highlight"
)

// 标准配置
const (
	StandardAppName        = "live" // 推流地址统一为 /live/
	standardRecordInterval = 1800   // 腾讯云默认分段 30 分钟
	standardRecordPath     = "/video/live/{StreamID}/{StartYear}{StartMonth}{StartDay}/"
	// 标准回调方式：补全现有模板（保留控制台配置的高光切片回调）或新建并替换
	StandardCallbackComplete = "complete"
	StandardCallbackReplace  = "replace"
)

var (
	// ErrTemplateNotFound 模板不存在或不属于本公司
	ErrTemplateNotFound = errors.New("模板不存在")
	// ErrNoCallbackBase 未配置系统回调地址
	ErrNoCallbackBase = errors.New("未配置系统回调地址（tencentCloud.live.callbackURL），无法生成回调地址")

	templateNameRe = regexp.MustCompile(`^[\p{Han}A-Za-z0-9_-]{1,255}$`)
)

// TemplateDefaults 新建模板时的默认值
type TemplateDefaults struct {
	CallbackBase       string `json:"callbackBase"`
	StreamBeginURL     string `json:"streamBeginUrl"`
	StreamEndURL       string `json:"streamEndUrl"`
	RecordURL          string `json:"recordUrl"`
	RecordStatusURL    string `json:"recordStatusUrl"`
	RecordExceptionURL string `json:"recordExceptionUrl"`
	PushExceptionURL   string `json:"pushExceptionUrl"`
	HighlightURL       string `json:"highlightUrl"`
	CosBucket          string `json:"cosBucket"`
	CosRegion          string `json:"cosRegion"`
	CosPath            string `json:"cosPath"`
	RecordInterval     int64  `json:"recordInterval"`
	TencentConfigured  bool   `json:"tencentConfigured"`
}

// TemplateRuleRef 模板绑定的域名
type TemplateRuleRef struct {
	DomainName string `json:"domainName"`
	AppName    string `json:"appName"`
	StreamName string `json:"streamName"`
}

// CallbackTemplateView 回调模板列表项
type CallbackTemplateView struct {
	*entity.LiveCallbackTemplate
	Rules []TemplateRuleRef `json:"rules"`
}

// RecordTemplateView 录制模板列表项
type RecordTemplateView struct {
	*entity.LiveRecordTemplate
	Rules []TemplateRuleRef `json:"rules"`
}

// TemplateListResult 模板列表（CloudError 非空表示未能与腾讯云对账）
type TemplateListResult[T any] struct {
	Items      []T    `json:"items"`
	CloudError string `json:"cloudError,omitempty"`
}

// ImportResult 从腾讯云导入的结果
type ImportResult struct {
	CallbackTemplates int `json:"callbackTemplates"`
	RecordTemplates   int `json:"recordTemplates"`
	Rules             int `json:"rules"`
}

// BindTemplateRequest 绑定模板到域名
type BindTemplateRequest struct {
	Kind       string `json:"kind"`       // callback / record
	DomainName string `json:"domainName"` // 推流域名
	TemplateID uint   `json:"templateId"` // 本地模板 ID
	AppName    string `json:"appName"`    // 默认 live
	DomainWide bool   `json:"domainWide"` // AppName 为空的规则（对整个域名生效），用于替换已有的空 AppName 规则
	StreamName string `json:"streamName"` // 仅录制规则，空为全部流
}

// TemplateIssue 配置体检问题
type TemplateIssue struct {
	Level   string `json:"level"` // error / warning / info
	Message string `json:"message"`
}

// DomainRuleView 域名上的一条规则及其模板
type DomainRuleView struct {
	AppName         string                       `json:"appName"`
	StreamName      string                       `json:"streamName"`
	CloudTemplateID int64                        `json:"cloudTemplateId"`
	Managed         bool                         `json:"managed"` // 模板已纳入本系统管理
	SyncStatus      string                       `json:"syncStatus"`
	SyncError       string                       `json:"syncError,omitempty"`
	SyncTime        *time.Time                   `json:"syncTime,omitempty"`
	Callback        *entity.LiveCallbackTemplate `json:"callback,omitempty"`
	Record          *entity.LiveRecordTemplate   `json:"record,omitempty"`
}

// HighlightStats 高光切片运行情况（腾讯云未开放配置接口，只能根据收到的回调判断）
type HighlightStats struct {
	Count7d     int64      `json:"count7d"`
	LastTime    *time.Time `json:"lastTime,omitempty"`
	CallbackURL string     `json:"callbackUrl"`
}

// DomainTemplateOverview 推流域名的模板配置概览
type DomainTemplateOverview struct {
	DomainName    string           `json:"domainName"`
	CallbackBase  string           `json:"callbackBase"`
	CallbackRules []DomainRuleView `json:"callbackRules"`
	RecordRules   []DomainRuleView `json:"recordRules"`
	Issues        []TemplateIssue  `json:"issues"`
	Highlight     HighlightStats   `json:"highlight"`
}

// StandardPlan 应用标准配置的预览
type StandardPlan struct {
	DomainName   string `json:"domainName"`
	CallbackBase string `json:"callbackBase"`
	Callback     struct {
		Exists          bool     `json:"exists"`          // 已绑定回调模板
		AppName         string   `json:"appName"`         // 现有规则的 AppName（为空表示对整个域名生效）
		TemplateName    string   `json:"templateName"`    // 当前模板
		CloudTemplateID int64    `json:"cloudTemplateId"` // 当前模板云端 ID
		Missing         []string `json:"missing"`         // 当前模板缺少的回调
		StandardName    string   `json:"standardName"`    // 新建时的模板名
	} `json:"callback"`
	Record struct {
		Exists           bool   `json:"exists"`
		AppName          string `json:"appName"` // 现有规则的 AppName；替换时沿用，避免同时存在两条规则导致重复录制
		TemplateName     string `json:"templateName"`
		CloudTemplateID  int64  `json:"cloudTemplateId"`
		CurrentInterval  int64  `json:"currentInterval"`
		StoresCompanyCos bool   `json:"storesCompanyCos"` // 当前模板已存储至本公司 COS
		ReplaceDefault   bool   `json:"replaceDefault"`   // 默认是否替换
		StandardName     string `json:"standardName"`
		CosBucket        string `json:"cosBucket"`
		CosRegion        string `json:"cosRegion"`
		CosPath          string `json:"cosPath"`
		RecordInterval   int64  `json:"recordInterval"`
		Error            string `json:"error,omitempty"` // 无法创建标准录制模板的原因
	} `json:"record"`
}

// StandardOptions 应用标准配置的选项
type StandardOptions struct {
	CallbackMode  string `json:"callbackMode"`  // complete（默认）/ replace
	ReplaceRecord *bool  `json:"replaceRecord"` // 空：按预览默认值
}

// StandardStepResult 标准配置每一步的执行结果
type StandardStepResult struct {
	Step    string `json:"step"`
	OK      bool   `json:"ok"`
	Skipped bool   `json:"skipped,omitempty"`
	Message string `json:"message"`
}

// TemplateService 直播回调、录制模板管理：本地为准，保存后同步到公司自己的腾讯云账号
type TemplateService interface {
	Defaults(ctx context.Context) (*TemplateDefaults, error)
	ListCallbackTemplates(ctx context.Context) (*TemplateListResult[*CallbackTemplateView], error)
	ListRecordTemplates(ctx context.Context) (*TemplateListResult[*RecordTemplateView], error)
	SaveCallbackTemplate(ctx context.Context, id uint, in *entity.LiveCallbackTemplate, operator string) (*entity.LiveCallbackTemplate, error)
	SaveRecordTemplate(ctx context.Context, id uint, in *entity.LiveRecordTemplate, operator string) (*entity.LiveRecordTemplate, error)
	DeleteTemplate(ctx context.Context, kind string, id uint) error
	SyncTemplate(ctx context.Context, kind string, id uint) error
	PullTemplate(ctx context.Context, kind string, id uint) error
	Import(ctx context.Context, operator string) (*ImportResult, error)
	DomainOverview(ctx context.Context, domain string) (*DomainTemplateOverview, error)
	Bind(ctx context.Context, req *BindTemplateRequest, operator string) (*entity.LiveTemplateRule, error)
	Unbind(ctx context.Context, kind, domain, appName, streamName string) error
	StandardPlan(ctx context.Context, domain string) (*StandardPlan, error)
	ApplyStandard(ctx context.Context, domain string, opt StandardOptions, operator string) ([]StandardStepResult, error)
}

type templateService struct {
	db           *gorm.DB
	live         Service
	callbackBase string
	now          func() time.Time
}

// NewTemplateService 创建模板服务；callbackBase 为本系统对外的回调地址（如 http://web.zigebo.com）
func NewTemplateService(db *gorm.DB, live Service, callbackBase string) TemplateService {
	return &templateService{db: db, live: live, callbackBase: strings.TrimRight(strings.TrimSpace(callbackBase), "/"), now: time.Now}
}

func companyOf(ctx context.Context) uint {
	id, _ := ctx.Value(contextutil.CompanyIDKey).(uint)
	return id
}

func (s *templateService) url(path string) string {
	if s.callbackBase == "" {
		return ""
	}
	return s.callbackBase + path
}

// ==================== 默认值 ====================

func (s *templateService) companyCos(ctx context.Context) (bucket, region string) {
	var conf entity.SysCompanyConf
	if err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ?", companyOf(ctx)).First(&conf).Error; err != nil {
		return "", ""
	}
	return tencentLive.StripBucketAppID(conf.TencentCOSBucketName), strings.TrimSpace(conf.TencentCOSRegion)
}

func (s *templateService) Defaults(ctx context.Context) (*TemplateDefaults, error) {
	bucket, region := s.companyCos(ctx)
	_, err := s.live.TemplateClient(ctx)
	return &TemplateDefaults{
		CallbackBase:       s.callbackBase,
		StreamBeginURL:     s.url(callbackPathPush),
		StreamEndURL:       s.url(callbackPathDisconnect),
		RecordURL:          s.url(callbackPathRecordFile),
		RecordStatusURL:    s.url(callbackPathRecordStatus),
		RecordExceptionURL: s.url(callbackPathRecordException),
		PushExceptionURL:   s.url(callbackPathPushException),
		HighlightURL:       s.url(callbackPathHighlight),
		CosBucket:          bucket,
		CosRegion:          region,
		CosPath:            standardRecordPath,
		RecordInterval:     standardRecordInterval,
		TencentConfigured:  err == nil,
	}, nil
}

// ==================== 转换与比较 ====================

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toCloudCallback(t *entity.LiveCallbackTemplate) *tencentLive.CallbackTemplate {
	return &tencentLive.CallbackTemplate{
		TemplateID:            t.CloudTemplateID,
		Name:                  t.Name,
		Description:           t.Description,
		StreamBeginURL:        t.StreamBeginURL,
		StreamEndURL:          t.StreamEndURL,
		RecordURL:             t.RecordURL,
		RecordStatusURL:       t.RecordStatusURL,
		RecordExceptionURL:    t.RecordExceptionURL,
		RecordExceptionLevels: splitList(t.RecordExceptionLevels),
		PushExceptionURL:      t.PushExceptionURL,
		CallbackKey:           t.CallbackKey,
	}
}

// applyCloudCallback 用云端模板覆盖本地字段（录制状态回调云端不返回，保留本地值）
func applyCloudCallback(t *entity.LiveCallbackTemplate, c *tencentLive.CallbackTemplate) {
	t.Name = c.Name
	t.Description = c.Description
	t.StreamBeginURL = c.StreamBeginURL
	t.StreamEndURL = c.StreamEndURL
	t.RecordURL = c.RecordURL
	t.RecordExceptionURL = c.RecordExceptionURL
	t.RecordExceptionLevels = strings.Join(c.RecordExceptionLevels, ",")
	t.PushExceptionURL = c.PushExceptionURL
	t.CallbackKey = c.CallbackKey
	t.CloudTemplateID = c.TemplateID
}

// diffCallback 本地与云端不一致的项
func diffCallback(t *entity.LiveCallbackTemplate, c *tencentLive.CallbackTemplate) []string {
	var d []string
	check := func(name, a, b string) {
		if strings.TrimSpace(a) != strings.TrimSpace(b) {
			d = append(d, name)
		}
	}
	check("模板名称", t.Name, c.Name)
	check("开播回调", t.StreamBeginURL, c.StreamBeginURL)
	check("断流回调", t.StreamEndURL, c.StreamEndURL)
	check("录制文件回调", t.RecordURL, c.RecordURL)
	check("录制异常回调", t.RecordExceptionURL, c.RecordExceptionURL)
	check("推流异常回调", t.PushExceptionURL, c.PushExceptionURL)
	check("回调密钥", t.CallbackKey, c.CallbackKey)
	return d
}

func yn(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func toCloudRecord(t *entity.LiveRecordTemplate) *tencentLive.RecordTemplate {
	return &tencentLive.RecordTemplate{
		TemplateID:      t.CloudTemplateID,
		Name:            t.Name,
		Description:     t.Description,
		CosStore:        t.CosStore == "Y",
		CosBucket:       t.CosBucket,
		CosRegion:       t.CosRegion,
		CosPath:         t.CosPath,
		RecordInterval:  t.RecordInterval,
		RemoveWatermark: t.RemoveWatermark == "Y",
	}
}

func applyCloudRecord(t *entity.LiveRecordTemplate, c *tencentLive.RecordTemplate) {
	t.Name = c.Name
	t.Description = c.Description
	t.CosStore = yn(c.CosStore)
	t.CosBucket = c.CosBucket
	t.CosRegion = c.CosRegion
	t.CosPath = c.CosPath
	t.RecordInterval = c.RecordInterval
	t.RemoveWatermark = yn(c.RemoveWatermark)
	t.Formats = strings.Join(c.Formats, ",")
	t.CloudTemplateID = c.TemplateID
}

func diffRecord(t *entity.LiveRecordTemplate, c *tencentLive.RecordTemplate) []string {
	var d []string
	if t.Name != c.Name {
		d = append(d, "模板名称")
	}
	if (t.CosStore == "Y") != c.CosStore {
		d = append(d, "存储方式")
	}
	if t.CosBucket != c.CosBucket || t.CosRegion != c.CosRegion {
		d = append(d, "存储桶")
	}
	if t.CosPath != c.CosPath {
		d = append(d, "存储路径")
	}
	if t.RecordInterval != c.RecordInterval {
		d = append(d, "分段时长")
	}
	if (t.RemoveWatermark == "Y") != c.RemoveWatermark {
		d = append(d, "录制原始流")
	}
	return d
}

func (s *templateService) markSynced(st *entity.TemplateSync) {
	now := s.now()
	st.SyncStatus, st.SyncError, st.SyncTime = entity.TemplateSyncSynced, "", &now
}

func (s *templateService) markFailed(st *entity.TemplateSync, err error) {
	now := s.now()
	msg := tencentLive.TemplateErrorMessage(err)
	if errors.Is(err, ErrNoCompanyTencent) {
		msg = err.Error()
	}
	if len([]rune(msg)) > 480 {
		msg = string([]rune(msg)[:480])
	}
	st.SyncStatus, st.SyncError, st.SyncTime = entity.TemplateSyncFailed, msg, &now
}

func driftMessage(fields []string) string {
	return "腾讯云上的配置与本地不一致：" + strings.Join(fields, "、")
}

// ==================== 校验 ====================

func validTemplateName(name string) error {
	if !templateNameRe.MatchString(name) {
		return errors.New("模板名称只能包含中文、英文、数字、_ 和 -")
	}
	return nil
}

func validURL(name, u string) error {
	if u == "" {
		return nil
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return fmt.Errorf("%s须以 http:// 或 https:// 开头", name)
	}
	return nil
}

func validateCallback(t *entity.LiveCallbackTemplate) error {
	if err := validTemplateName(t.Name); err != nil {
		return err
	}
	urls := []struct{ name, v string }{
		{"开播回调", t.StreamBeginURL}, {"断流回调", t.StreamEndURL}, {"录制文件回调", t.RecordURL},
		{"录制状态回调", t.RecordStatusURL}, {"录制异常回调", t.RecordExceptionURL}, {"推流异常回调", t.PushExceptionURL},
	}
	any := false
	for _, u := range urls {
		if err := validURL(u.name, u.v); err != nil {
			return err
		}
		any = any || u.v != ""
	}
	if !any {
		return errors.New("至少填写一个回调地址")
	}
	for _, l := range splitList(t.RecordExceptionLevels) {
		if l != "error" && l != "alarm" {
			return fmt.Errorf("录制异常级别只能是 error 或 alarm")
		}
	}
	return nil
}

func hasPathPlaceholder(p string) bool {
	for _, k := range []string{"{Start", "{End", "{RandomID}"} {
		if strings.Contains(p, k) {
			return true
		}
	}
	return false
}

func validateRecord(t *entity.LiveRecordTemplate) error {
	if err := validTemplateName(t.Name); err != nil {
		return err
	}
	if t.CosBucket == "" || t.CosRegion == "" {
		return errors.New("请填写 COS 存储桶与地域")
	}
	if !strings.HasPrefix(t.CosPath, "/") {
		return errors.New("存储路径须以 / 开头")
	}
	if !hasPathPlaceholder(t.CosPath) {
		return errors.New("存储路径须包含时间占位符（如 {StartYear}）或 {RandomID}，否则录制文件会重名覆盖")
	}
	if t.RecordInterval < 60 || t.RecordInterval > 43200 {
		return errors.New("分段时长须在 60 到 43200 秒之间")
	}
	return nil
}

// ==================== 列表与对账 ====================

func (s *templateService) rulesByTemplate(ctx context.Context, kind string) map[uint][]TemplateRuleRef {
	var rules []entity.LiveTemplateRule
	s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND KIND = ?", companyOf(ctx), kind).Order("DOMAIN_NAME").Find(&rules)
	out := map[uint][]TemplateRuleRef{}
	for _, r := range rules {
		out[r.TemplateID] = append(out[r.TemplateID], TemplateRuleRef{DomainName: r.DomainName, AppName: r.AppName, StreamName: r.StreamName})
	}
	return out
}

// reconcileStatus 已同步过的模板与云端比对，更新为「已同步」或「云端不一致」
func (s *templateService) reconcileStatus(st *entity.TemplateSync, cloudID int64, found bool, diffs []string) bool {
	if cloudID == 0 || st.SyncStatus == entity.TemplateSyncPending || st.SyncStatus == entity.TemplateSyncFailed {
		return false
	}
	switch {
	case !found:
		if st.SyncStatus == entity.TemplateSyncDrift && st.SyncError == "腾讯云上已不存在该模板" {
			return false
		}
		st.SyncStatus, st.SyncError = entity.TemplateSyncDrift, "腾讯云上已不存在该模板"
	case len(diffs) > 0:
		msg := driftMessage(diffs)
		if st.SyncStatus == entity.TemplateSyncDrift && st.SyncError == msg {
			return false
		}
		st.SyncStatus, st.SyncError = entity.TemplateSyncDrift, msg
	default:
		if st.SyncStatus == entity.TemplateSyncSynced {
			return false
		}
		st.SyncStatus, st.SyncError = entity.TemplateSyncSynced, ""
	}
	return true
}

func (s *templateService) ListCallbackTemplates(ctx context.Context) (*TemplateListResult[*CallbackTemplateView], error) {
	var list []*entity.LiveCallbackTemplate
	if err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx)).Order("ID").Find(&list).Error; err != nil {
		return nil, err
	}
	res := &TemplateListResult[*CallbackTemplateView]{}
	if client, err := s.live.TemplateClient(ctx); err != nil {
		res.CloudError = err.Error()
	} else if cloud, err := client.ListCallbackTemplates(); err != nil {
		res.CloudError = tencentLive.TemplateErrorMessage(err)
	} else {
		byID := map[int64]*tencentLive.CallbackTemplate{}
		for _, c := range cloud {
			byID[c.TemplateID] = c
		}
		for _, t := range list {
			c, found := byID[t.CloudTemplateID]
			var diffs []string
			if found {
				diffs = diffCallback(t, c)
			}
			if s.reconcileStatus(&t.TemplateSync, t.CloudTemplateID, found, diffs) {
				s.db.WithContext(ctx).Model(t).Select("SYNC_STATUS", "SYNC_ERROR").Updates(t)
			}
		}
	}
	rules := s.rulesByTemplate(ctx, entity.TemplateKindCallback)
	for _, t := range list {
		res.Items = append(res.Items, &CallbackTemplateView{LiveCallbackTemplate: t, Rules: rules[t.ID]})
	}
	return res, nil
}

func (s *templateService) ListRecordTemplates(ctx context.Context) (*TemplateListResult[*RecordTemplateView], error) {
	var list []*entity.LiveRecordTemplate
	if err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx)).Order("ID").Find(&list).Error; err != nil {
		return nil, err
	}
	res := &TemplateListResult[*RecordTemplateView]{}
	if client, err := s.live.TemplateClient(ctx); err != nil {
		res.CloudError = err.Error()
	} else if cloud, err := client.ListRecordTemplates(); err != nil {
		res.CloudError = tencentLive.TemplateErrorMessage(err)
	} else {
		byID := map[int64]*tencentLive.RecordTemplate{}
		for _, c := range cloud {
			byID[c.TemplateID] = c
		}
		for _, t := range list {
			c, found := byID[t.CloudTemplateID]
			var diffs []string
			if found {
				diffs = diffRecord(t, c)
			}
			if s.reconcileStatus(&t.TemplateSync, t.CloudTemplateID, found, diffs) {
				s.db.WithContext(ctx).Model(t).Select("SYNC_STATUS", "SYNC_ERROR").Updates(t)
			}
		}
	}
	rules := s.rulesByTemplate(ctx, entity.TemplateKindRecord)
	for _, t := range list {
		res.Items = append(res.Items, &RecordTemplateView{LiveRecordTemplate: t, Rules: rules[t.ID]})
	}
	return res, nil
}

// ==================== 保存与同步 ====================

func (s *templateService) loadCallback(ctx context.Context, id uint) (*entity.LiveCallbackTemplate, error) {
	var t entity.LiveCallbackTemplate
	if err := s.db.WithContext(ctx).Where("ID = ? AND SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", id, companyOf(ctx)).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *templateService) loadRecord(ctx context.Context, id uint) (*entity.LiveRecordTemplate, error) {
	var t entity.LiveRecordTemplate
	if err := s.db.WithContext(ctx).Where("ID = ? AND SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", id, companyOf(ctx)).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *templateService) nameTaken(ctx context.Context, model any, name string, exceptID uint) bool {
	var n int64
	s.db.WithContext(ctx).Model(model).Where("SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y' AND NAME = ? AND ID <> ?", companyOf(ctx), name, exceptID).Count(&n)
	return n > 0
}

// pushCallback 把本地回调模板写到云端（无云端 ID 则创建，否则修改），并更新同步状态
func (s *templateService) pushCallback(ctx context.Context, t *entity.LiveCallbackTemplate) error {
	client, err := s.live.TemplateClient(ctx)
	if err == nil {
		if t.CloudTemplateID == 0 {
			var id int64
			if id, err = client.CreateCallbackTemplate(toCloudCallback(t)); err == nil {
				t.CloudTemplateID = id
			}
		} else {
			err = client.ModifyCallbackTemplate(toCloudCallback(t))
		}
	}
	if err != nil {
		s.markFailed(&t.TemplateSync, err)
	} else {
		s.markSynced(&t.TemplateSync)
	}
	if serr := s.db.WithContext(ctx).Save(t).Error; serr != nil {
		return serr
	}
	return err
}

func (s *templateService) pushRecord(ctx context.Context, t *entity.LiveRecordTemplate) error {
	client, err := s.live.TemplateClient(ctx)
	if err == nil {
		if t.CloudTemplateID == 0 {
			var id int64
			if id, err = client.CreateRecordTemplate(toCloudRecord(t)); err == nil {
				t.CloudTemplateID = id
			}
		} else {
			err = client.ModifyRecordTemplate(toCloudRecord(t))
		}
	}
	if err != nil {
		s.markFailed(&t.TemplateSync, err)
	} else {
		s.markSynced(&t.TemplateSync)
	}
	if serr := s.db.WithContext(ctx).Save(t).Error; serr != nil {
		return serr
	}
	return err
}

// SaveCallbackTemplate 保存回调模板并同步；同步失败不算保存失败，结果在 SyncStatus / SyncError 中
func (s *templateService) SaveCallbackTemplate(ctx context.Context, id uint, in *entity.LiveCallbackTemplate, operator string) (*entity.LiveCallbackTemplate, error) {
	in.Name = strings.TrimSpace(in.Name)
	if err := validateCallback(in); err != nil {
		return nil, err
	}
	if s.nameTaken(ctx, &entity.LiveCallbackTemplate{}, in.Name, id) {
		return nil, fmt.Errorf("模板名称「%s」已存在", in.Name)
	}
	t := &entity.LiveCallbackTemplate{SysCompanyID: companyOf(ctx), Source: entity.TemplateSourceLocal, IsActive: "Y", CreateBy: operator}
	t.SyncStatus = entity.TemplateSyncPending
	if id > 0 {
		var err error
		if t, err = s.loadCallback(ctx, id); err != nil {
			return nil, err
		}
	}
	t.Name, t.Description = in.Name, in.Description
	t.StreamBeginURL, t.StreamEndURL, t.RecordURL = in.StreamBeginURL, in.StreamEndURL, in.RecordURL
	t.RecordStatusURL, t.RecordExceptionURL, t.PushExceptionURL = in.RecordStatusURL, in.RecordExceptionURL, in.PushExceptionURL
	t.RecordExceptionLevels, t.CallbackKey = in.RecordExceptionLevels, in.CallbackKey
	t.UpdateBy = operator
	if err := s.db.WithContext(ctx).Save(t).Error; err != nil {
		return nil, err
	}
	_ = s.pushCallback(ctx, t)
	return t, nil
}

// SaveRecordTemplate 保存录制模板（存储至 COS 的 MP4 录制）并同步
func (s *templateService) SaveRecordTemplate(ctx context.Context, id uint, in *entity.LiveRecordTemplate, operator string) (*entity.LiveRecordTemplate, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.CosBucket = tencentLive.StripBucketAppID(in.CosBucket)
	in.CosRegion = strings.TrimSpace(in.CosRegion)
	in.CosPath = strings.TrimSpace(in.CosPath)
	if err := validateRecord(in); err != nil {
		return nil, err
	}
	if s.nameTaken(ctx, &entity.LiveRecordTemplate{}, in.Name, id) {
		return nil, fmt.Errorf("模板名称「%s」已存在", in.Name)
	}
	t := &entity.LiveRecordTemplate{SysCompanyID: companyOf(ctx), Source: entity.TemplateSourceLocal, CosStore: "Y", Formats: "mp4", IsActive: "Y", CreateBy: operator}
	t.SyncStatus = entity.TemplateSyncPending
	if id > 0 {
		var err error
		if t, err = s.loadRecord(ctx, id); err != nil {
			return nil, err
		}
		if t.CosStore != "Y" {
			return nil, errors.New("存储至云点播 VOD 的录制模板仅支持绑定，不支持在本系统修改")
		}
	}
	t.Name, t.Description = in.Name, in.Description
	t.CosBucket, t.CosRegion, t.CosPath = in.CosBucket, in.CosRegion, in.CosPath
	t.RecordInterval = in.RecordInterval
	t.RemoveWatermark = yn(in.RemoveWatermark != "N")
	t.UpdateBy = operator
	if err := s.db.WithContext(ctx).Save(t).Error; err != nil {
		return nil, err
	}
	_ = s.pushRecord(ctx, t)
	return t, nil
}

func (s *templateService) countRules(ctx context.Context, kind string, templateID uint) int64 {
	var n int64
	s.db.WithContext(ctx).Model(&entity.LiveTemplateRule{}).Where("SYS_COMPANY_ID = ? AND KIND = ? AND TEMPLATE_ID = ?", companyOf(ctx), kind, templateID).Count(&n)
	return n
}

// DeleteTemplate 删除模板：已绑定域名的须先解绑；先删云端，再删本地
func (s *templateService) DeleteTemplate(ctx context.Context, kind string, id uint) error {
	if n := s.countRules(ctx, kind, id); n > 0 {
		return fmt.Errorf("模板已绑定 %d 条域名规则，请先解绑", n)
	}
	var cloudID int64
	var model any
	switch kind {
	case entity.TemplateKindCallback:
		t, err := s.loadCallback(ctx, id)
		if err != nil {
			return err
		}
		cloudID, model = t.CloudTemplateID, t
	case entity.TemplateKindRecord:
		t, err := s.loadRecord(ctx, id)
		if err != nil {
			return err
		}
		cloudID, model = t.CloudTemplateID, t
	default:
		return errors.New("未知的模板类型")
	}
	if cloudID > 0 {
		client, err := s.live.TemplateClient(ctx)
		if err != nil {
			return err
		}
		if kind == entity.TemplateKindCallback {
			err = client.DeleteCallbackTemplate(cloudID)
		} else {
			err = client.DeleteRecordTemplate(cloudID)
		}
		if err != nil && !tencentLive.IsTencentNotFound(err) {
			return errors.New(tencentLive.TemplateErrorMessage(err))
		}
	}
	return s.db.WithContext(ctx).Delete(model).Error
}

// SyncTemplate 以本地为准写到云端（重新同步、以本地覆盖云端）
func (s *templateService) SyncTemplate(ctx context.Context, kind string, id uint) error {
	var err error
	switch kind {
	case entity.TemplateKindCallback:
		var t *entity.LiveCallbackTemplate
		if t, err = s.loadCallback(ctx, id); err == nil {
			err = s.pushCallback(ctx, t)
		}
	case entity.TemplateKindRecord:
		var t *entity.LiveRecordTemplate
		if t, err = s.loadRecord(ctx, id); err == nil {
			if t.CosStore != "Y" {
				return errors.New("存储至云点播 VOD 的录制模板不支持从本系统同步")
			}
			err = s.pushRecord(ctx, t)
		}
	default:
		return errors.New("未知的模板类型")
	}
	if err != nil && !errors.Is(err, ErrTemplateNotFound) && !errors.Is(err, ErrNoCompanyTencent) {
		return errors.New(tencentLive.TemplateErrorMessage(err))
	}
	return err
}

// PullTemplate 用云端配置更新本地
func (s *templateService) PullTemplate(ctx context.Context, kind string, id uint) error {
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return err
	}
	switch kind {
	case entity.TemplateKindCallback:
		t, err := s.loadCallback(ctx, id)
		if err != nil {
			return err
		}
		cloud, err := client.ListCallbackTemplates()
		if err != nil {
			return errors.New(tencentLive.TemplateErrorMessage(err))
		}
		for _, c := range cloud {
			if c.TemplateID == t.CloudTemplateID && t.CloudTemplateID > 0 {
				applyCloudCallback(t, c)
				s.markSynced(&t.TemplateSync)
				return s.db.WithContext(ctx).Save(t).Error
			}
		}
	case entity.TemplateKindRecord:
		t, err := s.loadRecord(ctx, id)
		if err != nil {
			return err
		}
		cloud, err := client.ListRecordTemplates()
		if err != nil {
			return errors.New(tencentLive.TemplateErrorMessage(err))
		}
		for _, c := range cloud {
			if c.TemplateID == t.CloudTemplateID && t.CloudTemplateID > 0 {
				applyCloudRecord(t, c)
				s.markSynced(&t.TemplateSync)
				return s.db.WithContext(ctx).Save(t).Error
			}
		}
	default:
		return errors.New("未知的模板类型")
	}
	return errors.New("腾讯云上已不存在该模板，无法用云端更新")
}

// ==================== 导入 ====================

// importCallback 云端回调模板纳入本地管理（已有则返回本地记录）
func (s *templateService) importCallback(ctx context.Context, c *tencentLive.CallbackTemplate, operator string) (*entity.LiveCallbackTemplate, bool, error) {
	var t entity.LiveCallbackTemplate
	err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND CLOUD_TEMPLATE_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), c.TemplateID).First(&t).Error
	if err == nil {
		return &t, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	t = entity.LiveCallbackTemplate{SysCompanyID: companyOf(ctx), Source: entity.TemplateSourceImport, IsActive: "Y", CreateBy: operator, UpdateBy: operator}
	applyCloudCallback(&t, c)
	s.markSynced(&t.TemplateSync)
	return &t, true, s.db.WithContext(ctx).Create(&t).Error
}

func (s *templateService) importRecord(ctx context.Context, c *tencentLive.RecordTemplate, operator string) (*entity.LiveRecordTemplate, bool, error) {
	var t entity.LiveRecordTemplate
	err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND CLOUD_TEMPLATE_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), c.TemplateID).First(&t).Error
	if err == nil {
		return &t, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	t = entity.LiveRecordTemplate{SysCompanyID: companyOf(ctx), Source: entity.TemplateSourceImport, IsActive: "Y", CreateBy: operator, UpdateBy: operator}
	applyCloudRecord(&t, c)
	s.markSynced(&t.TemplateSync)
	return &t, true, s.db.WithContext(ctx).Create(&t).Error
}

// upsertRule 写入本地规则（按 公司 + 类型 + 域名 + AppName + StreamName 唯一）
func (s *templateService) upsertRule(ctx context.Context, r *entity.LiveTemplateRule) error {
	var existing entity.LiveTemplateRule
	err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND KIND = ? AND DOMAIN_NAME = ? AND APP_NAME = ? AND STREAM_NAME = ?",
		r.SysCompanyID, r.Kind, r.DomainName, r.AppName, r.StreamName).First(&existing).Error
	if err == nil {
		r.ID, r.CreateBy, r.CreateTime = existing.ID, existing.CreateBy, existing.CreateTime
		return s.db.WithContext(ctx).Save(r).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.db.WithContext(ctx).Create(r).Error
}

// Import 把公司腾讯云账号中已有的模板与规则纳入本地管理
func (s *templateService) Import(ctx context.Context, operator string) (*ImportResult, error) {
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return nil, err
	}
	res := &ImportResult{}
	cbs, err := client.ListCallbackTemplates()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	recs, err := client.ListRecordTemplates()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	cbRules, err := client.ListCallbackRules()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	recRules, err := client.ListRecordRules()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}

	cbLocal := map[int64]uint{}
	for _, c := range cbs {
		t, created, err := s.importCallback(ctx, c, operator)
		if err != nil {
			return nil, err
		}
		cbLocal[c.TemplateID] = t.ID
		if created {
			res.CallbackTemplates++
		}
	}
	recLocal := map[int64]uint{}
	for _, c := range recs {
		t, created, err := s.importRecord(ctx, c, operator)
		if err != nil {
			return nil, err
		}
		recLocal[c.TemplateID] = t.ID
		if created {
			res.RecordTemplates++
		}
	}
	addRules := func(kind string, rules []*tencentLive.TemplateRule, local map[int64]uint) error {
		for _, r := range rules {
			id, ok := local[r.TemplateID]
			if !ok {
				continue
			}
			rule := &entity.LiveTemplateRule{SysCompanyID: companyOf(ctx), Kind: kind, DomainName: r.DomainName, AppName: r.AppName,
				StreamName: r.StreamName, TemplateID: id, CreateBy: operator, UpdateBy: operator}
			s.markSynced(&rule.TemplateSync)
			if err := s.upsertRule(ctx, rule); err != nil {
				return err
			}
			res.Rules++
		}
		return nil
	}
	if err := addRules(entity.TemplateKindCallback, cbRules, cbLocal); err != nil {
		return nil, err
	}
	if err := addRules(entity.TemplateKindRecord, recRules, recLocal); err != nil {
		return nil, err
	}
	return res, nil
}

// ==================== 域名概览与体检 ====================

func (s *templateService) highlightStats(ctx context.Context, domain string) HighlightStats {
	st := HighlightStats{CallbackURL: s.url(callbackPathHighlight)}
	var row struct {
		N    int64
		Last *time.Time
	}
	s.db.WithContext(ctx).Model(&entity.LiveCallbackEvent{}).
		Select("COUNT(*) AS n, MAX(CREATE_TIME) AS last").
		Where("SYS_COMPANY_ID = ? AND DOMAIN_NAME = ? AND EVENT_TYPE = ? AND CREATE_TIME >= ?", companyOf(ctx), domain, "highlight", s.now().AddDate(0, 0, -7)).
		Scan(&row)
	st.Count7d, st.LastTime = row.N, row.Last
	return st
}

// pointsHere 回调地址指向本系统
func (s *templateService) pointsHere(u string) bool {
	return s.callbackBase != "" && strings.HasPrefix(u, s.callbackBase+"/")
}

func (s *templateService) DomainOverview(ctx context.Context, domain string) (*DomainTemplateOverview, error) {
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.checkPushDomain(ctx, domain); err != nil {
		return nil, err
	}
	cbRules, err := client.ListCallbackRules()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	recRules, err := client.ListRecordRules()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	cbs, err := client.ListCallbackTemplates()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	recs, err := client.ListRecordTemplates()
	if err != nil {
		return nil, errors.New(tencentLive.TemplateErrorMessage(err))
	}
	cbCloud := map[int64]*tencentLive.CallbackTemplate{}
	for _, c := range cbs {
		cbCloud[c.TemplateID] = c
	}
	recCloud := map[int64]*tencentLive.RecordTemplate{}
	for _, c := range recs {
		recCloud[c.TemplateID] = c
	}

	ov := &DomainTemplateOverview{DomainName: domain, CallbackBase: s.callbackBase, Highlight: s.highlightStats(ctx, domain)}
	var localRules []entity.LiveTemplateRule
	s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND DOMAIN_NAME = ?", companyOf(ctx), domain).Find(&localRules)
	localRule := func(kind, app, stream string) *entity.LiveTemplateRule {
		for i := range localRules {
			r := &localRules[i]
			if r.Kind == kind && r.AppName == app && r.StreamName == stream {
				return r
			}
		}
		return nil
	}
	seen := map[uint]bool{}

	// 回调规则
	for _, r := range cbRules {
		if r.DomainName != domain {
			continue
		}
		v := DomainRuleView{AppName: r.AppName, CloudTemplateID: r.TemplateID, SyncStatus: entity.TemplateSyncSynced}
		var local entity.LiveCallbackTemplate
		if s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND CLOUD_TEMPLATE_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), r.TemplateID).First(&local).Error == nil {
			v.Managed, v.Callback = true, &local
		} else if c, ok := cbCloud[r.TemplateID]; ok {
			t := &entity.LiveCallbackTemplate{}
			applyCloudCallback(t, c)
			v.Callback = t
		}
		if lr := localRule(entity.TemplateKindCallback, r.AppName, ""); lr != nil {
			seen[lr.ID] = true
			if v.Managed && lr.TemplateID != v.Callback.ID {
				v.SyncStatus, v.SyncError = entity.TemplateSyncDrift, "腾讯云上绑定的模板与本地不一致"
			}
			v.SyncTime = lr.SyncTime
		}
		ov.CallbackRules = append(ov.CallbackRules, v)
	}
	// 录制规则
	for _, r := range recRules {
		if r.DomainName != domain {
			continue
		}
		v := DomainRuleView{AppName: r.AppName, StreamName: r.StreamName, CloudTemplateID: r.TemplateID, SyncStatus: entity.TemplateSyncSynced}
		var local entity.LiveRecordTemplate
		if s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND CLOUD_TEMPLATE_ID = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), r.TemplateID).First(&local).Error == nil {
			v.Managed, v.Record = true, &local
		} else if c, ok := recCloud[r.TemplateID]; ok {
			t := &entity.LiveRecordTemplate{}
			applyCloudRecord(t, c)
			v.Record = t
		}
		if lr := localRule(entity.TemplateKindRecord, r.AppName, r.StreamName); lr != nil {
			seen[lr.ID] = true
			if v.Managed && lr.TemplateID != v.Record.ID {
				v.SyncStatus, v.SyncError = entity.TemplateSyncDrift, "腾讯云上绑定的模板与本地不一致"
			}
			v.SyncTime = lr.SyncTime
		}
		ov.RecordRules = append(ov.RecordRules, v)
	}
	// 本地有、云端没有的规则（同步失败，或在控制台被解绑）
	for _, lr := range localRules {
		if seen[lr.ID] {
			continue
		}
		v := DomainRuleView{AppName: lr.AppName, StreamName: lr.StreamName, Managed: true, SyncStatus: lr.SyncStatus, SyncError: lr.SyncError, SyncTime: lr.SyncTime}
		if v.SyncStatus == entity.TemplateSyncSynced {
			v.SyncStatus, v.SyncError = entity.TemplateSyncDrift, "腾讯云上已没有这条规则"
		}
		if lr.Kind == entity.TemplateKindCallback {
			if t, err := s.loadCallback(ctx, lr.TemplateID); err == nil {
				v.Callback, v.CloudTemplateID = t, t.CloudTemplateID
			}
			ov.CallbackRules = append(ov.CallbackRules, v)
		} else {
			if t, err := s.loadRecord(ctx, lr.TemplateID); err == nil {
				v.Record, v.CloudTemplateID = t, t.CloudTemplateID
			}
			ov.RecordRules = append(ov.RecordRules, v)
		}
	}
	sortRules := func(rs []DomainRuleView) {
		sort.SliceStable(rs, func(i, j int) bool { return rs[i].AppName == StandardAppName && rs[j].AppName != StandardAppName })
	}
	sortRules(ov.CallbackRules)
	sortRules(ov.RecordRules)
	ov.Issues = s.checkIssues(ctx, ov)
	return ov, nil
}

// checkIssues 配置体检
func (s *templateService) checkIssues(ctx context.Context, ov *DomainTemplateOverview) []TemplateIssue {
	issues := []TemplateIssue{}
	add := func(level, format string, args ...any) {
		issues = append(issues, TemplateIssue{Level: level, Message: fmt.Sprintf(format, args...)})
	}
	if s.callbackBase == "" {
		add("warning", "未配置系统回调地址（tencentCloud.live.callbackURL），无法判断回调是否指向本系统，也不能应用标准配置")
	}
	var cb *DomainRuleView
	for i := range ov.CallbackRules {
		if ov.CallbackRules[i].AppName == StandardAppName || ov.CallbackRules[i].AppName == "" {
			cb = &ov.CallbackRules[i]
			break
		}
	}
	if cb == nil || cb.Callback == nil {
		add("error", "未绑定回调模板（AppName = live）：直播间状态不会自动更新，录制文件也不会转存到云盘")
	} else {
		t := cb.Callback
		var missing []string
		if t.StreamBeginURL == "" {
			missing = append(missing, "开播")
		}
		if t.StreamEndURL == "" {
			missing = append(missing, "断流")
		}
		if len(missing) > 0 {
			add("warning", "回调模板「%s」未配置%s回调 —— 直播间状态不会自动更新", t.Name, strings.Join(missing, "、"))
		}
		if t.RecordURL == "" {
			add("warning", "回调模板「%s」未配置录制文件回调 —— 录制文件不会转存到云盘", t.Name)
		}
		if t.RecordExceptionURL == "" {
			add("info", "回调模板「%s」未配置录制异常回调 —— 录制失败时系统无法提示", t.Name)
		}
		if s.callbackBase != "" {
			for _, u := range []string{t.StreamBeginURL, t.StreamEndURL, t.RecordURL} {
				if u != "" && !s.pointsHere(u) {
					add("warning", "回调模板「%s」的回调地址未指向本系统（%s），回调不会到达本系统", t.Name, s.callbackBase)
					break
				}
			}
		}
		if cb.SyncStatus == entity.TemplateSyncDrift {
			add("warning", "回调规则：%s", cb.SyncError)
		}
	}
	bucket, region := s.companyCos(ctx)
	var rec *DomainRuleView
	for i := range ov.RecordRules {
		if ov.RecordRules[i].AppName == StandardAppName || ov.RecordRules[i].AppName == "" {
			rec = &ov.RecordRules[i]
			break
		}
	}
	if rec == nil || rec.Record == nil {
		add("info", "未绑定录制模板：推流不会自动录制")
	} else {
		t := rec.Record
		if t.CosStore != "Y" {
			add("warning", "录制模板「%s」存储至云点播 VOD，录制文件不会进入本公司 COS", t.Name)
		} else if bucket == "" || region == "" {
			add("info", "公司配置中的 COS 存储桶或地域未填写，无法判断录制是否存储至本公司 COS，也不能创建标准录制模板")
		} else if t.CosBucket != bucket || t.CosRegion != region {
			add("warning", "录制模板「%s」存储到 %s（%s），不是本公司配置的 COS 存储桶 %s（%s）", t.Name, t.CosBucket, t.CosRegion, bucket, region)
		}
		if rec.SyncStatus == entity.TemplateSyncDrift {
			add("warning", "录制规则：%s", rec.SyncError)
		}
	}
	for _, r := range append(append([]DomainRuleView{}, ov.CallbackRules...), ov.RecordRules...) {
		if r.AppName != StandardAppName && r.AppName != "" {
			add("info", "存在 AppName = %s 的规则，本系统推流地址使用 live，该规则对本系统直播间不生效", r.AppName)
			break
		}
	}
	if ov.Highlight.Count7d == 0 {
		add("info", "近 7 天未收到高光切片，如需使用请在腾讯云控制台开启（高光切片仅支持在控制台配置）")
	}
	return issues
}

// ==================== 绑定与解绑 ====================

// checkPushDomain 域名须为本公司腾讯云账号下的推流域名
func (s *templateService) checkPushDomain(ctx context.Context, domain string) error {
	info, err := s.live.DescribeDomain(ctx, domain)
	if err != nil || info == nil {
		return fmt.Errorf("域名 %s 不存在或不属于本公司腾讯云账号", domain)
	}
	if info.Type != 0 {
		return fmt.Errorf("%s 不是推流域名，模板只能绑定到推流域名", domain)
	}
	return nil
}

// ensureCloudTemplate 模板须已同步到云端，返回云端模板 ID
func (s *templateService) ensureCloudTemplate(ctx context.Context, kind string, id uint) (int64, error) {
	switch kind {
	case entity.TemplateKindCallback:
		t, err := s.loadCallback(ctx, id)
		if err != nil {
			return 0, err
		}
		if t.CloudTemplateID == 0 || t.SyncStatus == entity.TemplateSyncFailed {
			if err := s.pushCallback(ctx, t); err != nil {
				return 0, fmt.Errorf("模板同步失败：%s", t.SyncError)
			}
		}
		return t.CloudTemplateID, nil
	case entity.TemplateKindRecord:
		t, err := s.loadRecord(ctx, id)
		if err != nil {
			return 0, err
		}
		if t.CloudTemplateID == 0 || (t.SyncStatus == entity.TemplateSyncFailed && t.CosStore == "Y") {
			if err := s.pushRecord(ctx, t); err != nil {
				return 0, fmt.Errorf("模板同步失败：%s", t.SyncError)
			}
		}
		return t.CloudTemplateID, nil
	}
	return 0, errors.New("未知的模板类型")
}

// Bind 绑定模板到推流域名：腾讯云没有修改规则的接口，已有不同模板的规则时先删后建
func (s *templateService) Bind(ctx context.Context, req *BindTemplateRequest, operator string) (*entity.LiveTemplateRule, error) {
	req.AppName = strings.TrimSpace(req.AppName)
	if req.AppName == "" && !req.DomainWide {
		req.AppName = StandardAppName
	}
	return s.bindRule(ctx, req, operator)
}

// ruleApp 替换已有规则时沿用其 AppName（可能为空），新建时用 live
func ruleApp(exists bool, app string) string {
	if exists {
		return app
	}
	return StandardAppName
}

// bindRule 按请求中的 AppName 原值绑定（可为空：对整个域名生效）
func (s *templateService) bindRule(ctx context.Context, req *BindTemplateRequest, operator string) (*entity.LiveTemplateRule, error) {
	req.DomainName = strings.TrimSpace(req.DomainName)
	if req.Kind == entity.TemplateKindCallback {
		req.StreamName = ""
	}
	if req.Kind != entity.TemplateKindCallback && req.Kind != entity.TemplateKindRecord {
		return nil, errors.New("未知的模板类型")
	}
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.checkPushDomain(ctx, req.DomainName); err != nil {
		return nil, err
	}
	cloudID, err := s.ensureCloudTemplate(ctx, req.Kind, req.TemplateID)
	if err != nil {
		return nil, err
	}

	rule := &entity.LiveTemplateRule{SysCompanyID: companyOf(ctx), Kind: req.Kind, DomainName: req.DomainName, AppName: req.AppName,
		StreamName: req.StreamName, TemplateID: req.TemplateID, CreateBy: operator, UpdateBy: operator}
	err = s.bindCloud(client, req.Kind, req.DomainName, req.AppName, req.StreamName, cloudID)
	if err != nil {
		s.markFailed(&rule.TemplateSync, err)
	} else {
		s.markSynced(&rule.TemplateSync)
	}
	if serr := s.upsertRule(ctx, rule); serr != nil {
		return nil, serr
	}
	if err != nil {
		return rule, errors.New(rule.SyncError)
	}
	return rule, nil
}

// bindCloud 在云端绑定：相同模板已绑定则不动，不同模板先解绑再绑定
func (s *templateService) bindCloud(client tencentLive.TemplateClient, kind, domain, app, stream string, cloudID int64) error {
	var rules []*tencentLive.TemplateRule
	var err error
	if kind == entity.TemplateKindCallback {
		rules, err = client.ListCallbackRules()
	} else {
		rules, err = client.ListRecordRules()
	}
	if err != nil {
		return err
	}
	for _, r := range rules {
		if r.DomainName != domain || r.AppName != app || r.StreamName != stream {
			continue
		}
		if r.TemplateID == cloudID {
			return nil
		}
		if kind == entity.TemplateKindCallback {
			err = client.DeleteCallbackRule(domain, app)
		} else {
			err = client.DeleteRecordRule(domain, app, stream)
		}
		if err != nil && !tencentLive.IsTencentNotFound(err) {
			return err
		}
	}
	if kind == entity.TemplateKindCallback {
		return client.CreateCallbackRule(domain, app, cloudID)
	}
	return client.CreateRecordRule(domain, app, stream, cloudID)
}

// Unbind 解绑：云端按规则原值精确删除，再删本地
func (s *templateService) Unbind(ctx context.Context, kind, domain, appName, streamName string) error {
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return err
	}
	switch kind {
	case entity.TemplateKindCallback:
		err = client.DeleteCallbackRule(domain, appName)
		streamName = ""
	case entity.TemplateKindRecord:
		err = client.DeleteRecordRule(domain, appName, streamName)
	default:
		return errors.New("未知的模板类型")
	}
	if err != nil && !tencentLive.IsTencentNotFound(err) {
		return errors.New(tencentLive.TemplateErrorMessage(err))
	}
	return s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND KIND = ? AND DOMAIN_NAME = ? AND APP_NAME = ? AND STREAM_NAME = ?",
		companyOf(ctx), kind, domain, appName, streamName).Delete(&entity.LiveTemplateRule{}).Error
}

// ==================== 标准配置 ====================

func (s *templateService) standardCallbackName(ctx context.Context) string {
	return fmt.Sprintf("zigebo-callback-%d", companyOf(ctx))
}

func (s *templateService) standardRecordName(ctx context.Context) string {
	return fmt.Sprintf("zigebo-record-%d", companyOf(ctx))
}

// standardCallbackFields 标准回调地址：只补全为空的项，已有的保留
func (s *templateService) fillStandardCallback(t *entity.LiveCallbackTemplate) []string {
	var filled []string
	fill := func(name string, field *string, path string) {
		if strings.TrimSpace(*field) == "" {
			*field = s.url(path)
			filled = append(filled, name)
		}
	}
	fill("开播", &t.StreamBeginURL, callbackPathPush)
	fill("断流", &t.StreamEndURL, callbackPathDisconnect)
	fill("录制文件", &t.RecordURL, callbackPathRecordFile)
	fill("录制状态", &t.RecordStatusURL, callbackPathRecordStatus)
	fill("录制异常", &t.RecordExceptionURL, callbackPathRecordException)
	fill("推流异常", &t.PushExceptionURL, callbackPathPushException)
	if t.RecordExceptionURL != "" && t.RecordExceptionLevels == "" {
		t.RecordExceptionLevels = "error"
	}
	return filled
}

// findRule 对本系统推流（AppName = live）生效的规则：优先 live，其次 AppName 为空（对整个域名生效）的规则
func findRule(rules []DomainRuleView) *DomainRuleView {
	for _, app := range []string{StandardAppName, ""} {
		for i := range rules {
			if rules[i].AppName == app && rules[i].StreamName == "" {
				return &rules[i]
			}
		}
	}
	return nil
}

func (s *templateService) StandardPlan(ctx context.Context, domain string) (*StandardPlan, error) {
	if s.callbackBase == "" {
		return nil, ErrNoCallbackBase
	}
	ov, err := s.DomainOverview(ctx, domain)
	if err != nil {
		return nil, err
	}
	p := &StandardPlan{DomainName: domain, CallbackBase: s.callbackBase}
	p.Callback.StandardName = s.standardCallbackName(ctx)
	if r := findRule(ov.CallbackRules); r != nil && r.Callback != nil {
		p.Callback.Exists, p.Callback.AppName = true, r.AppName
		p.Callback.TemplateName, p.Callback.CloudTemplateID = r.Callback.Name, r.CloudTemplateID
		probe := *r.Callback
		p.Callback.Missing = s.fillStandardCallback(&probe)
	}
	bucket, region := s.companyCos(ctx)
	p.Record.StandardName = s.standardRecordName(ctx)
	p.Record.CosBucket, p.Record.CosRegion, p.Record.CosPath, p.Record.RecordInterval = bucket, region, standardRecordPath, standardRecordInterval
	if bucket == "" || region == "" {
		p.Record.Error = "公司未配置 COS 存储桶与地域，无法创建录制模板（请先在公司配置中填写）"
	}
	if r := findRule(ov.RecordRules); r != nil && r.Record != nil {
		p.Record.Exists, p.Record.AppName = true, r.AppName
		p.Record.TemplateName, p.Record.CloudTemplateID, p.Record.CurrentInterval = r.Record.Name, r.CloudTemplateID, r.Record.RecordInterval
		p.Record.StoresCompanyCos = r.Record.CosStore == "Y" && bucket != "" && r.Record.CosBucket == bucket && r.Record.CosRegion == region
		p.Record.ReplaceDefault = !p.Record.StoresCompanyCos
	} else {
		p.Record.ReplaceDefault = true
	}
	return p, nil
}

// ApplyStandard 应用标准配置：回调默认补全现有模板（保留控制台配置的高光切片回调），录制按选项替换为本公司 COS 的标准模板
func (s *templateService) ApplyStandard(ctx context.Context, domain string, opt StandardOptions, operator string) ([]StandardStepResult, error) {
	plan, err := s.StandardPlan(ctx, domain)
	if err != nil {
		return nil, err
	}
	if err := s.checkPushDomain(ctx, domain); err != nil {
		return nil, err
	}
	client, err := s.live.TemplateClient(ctx)
	if err != nil {
		return nil, err
	}
	var steps []StandardStepResult
	result := func(step string, err error, okMsg string) {
		if err != nil {
			steps = append(steps, StandardStepResult{Step: step, Message: tencentLive.TemplateErrorMessage(err)})
			return
		}
		steps = append(steps, StandardStepResult{Step: step, OK: true, Message: okMsg})
	}

	// 1. 回调
	if plan.Callback.Exists && opt.CallbackMode != StandardCallbackReplace {
		var c *tencentLive.CallbackTemplate
		if cloud, err := client.ListCallbackTemplates(); err == nil {
			for _, x := range cloud {
				if x.TemplateID == plan.Callback.CloudTemplateID {
					c = x
				}
			}
		}
		if c == nil {
			result("补全回调模板", errors.New("腾讯云上找不到当前绑定的回调模板"), "")
		} else if t, _, err := s.importCallback(ctx, c, operator); err != nil {
			result("补全回调模板", err, "")
		} else {
			filled := s.fillStandardCallback(t)
			t.UpdateBy = operator
			if len(filled) == 0 {
				steps = append(steps, StandardStepResult{Step: "补全回调模板", OK: true, Skipped: true, Message: fmt.Sprintf("「%s」已包含全部标准回调，无需修改", t.Name)})
			} else {
				err := s.pushCallback(ctx, t)
				if err != nil {
					err = errors.New(t.SyncError)
				}
				result("补全回调模板", err, fmt.Sprintf("「%s」已补全%s回调", t.Name, strings.Join(filled, "、")))
			}
			rule := &entity.LiveTemplateRule{SysCompanyID: companyOf(ctx), Kind: entity.TemplateKindCallback, DomainName: domain, AppName: plan.Callback.AppName, TemplateID: t.ID, CreateBy: operator, UpdateBy: operator}
			s.markSynced(&rule.TemplateSync)
			_ = s.upsertRule(ctx, rule)
		}
	} else {
		t, err := s.ensureStandardCallback(ctx, operator)
		if err == nil {
			_, err = s.bindRule(ctx, &BindTemplateRequest{Kind: entity.TemplateKindCallback, DomainName: domain, TemplateID: t.ID, AppName: ruleApp(plan.Callback.Exists, plan.Callback.AppName)}, operator)
		}
		result("绑定标准回调模板", err, fmt.Sprintf("已绑定「%s」", s.standardCallbackName(ctx)))
	}

	// 2. 录制
	replace := plan.Record.ReplaceDefault
	if opt.ReplaceRecord != nil {
		replace = *opt.ReplaceRecord
	}
	switch {
	case !replace:
		steps = append(steps, StandardStepResult{Step: "录制模板", OK: true, Skipped: true, Message: "保留当前录制模板"})
	case plan.Record.Error != "":
		steps = append(steps, StandardStepResult{Step: "录制模板", Message: plan.Record.Error})
	default:
		t, err := s.ensureStandardRecord(ctx, plan, operator)
		if err == nil {
			_, err = s.bindRule(ctx, &BindTemplateRequest{Kind: entity.TemplateKindRecord, DomainName: domain, TemplateID: t.ID, AppName: ruleApp(plan.Record.Exists, plan.Record.AppName)}, operator)
		}
		result("绑定标准录制模板", err, fmt.Sprintf("已绑定「%s」（存储至本公司 COS，每 %d 分钟一段）", s.standardRecordName(ctx), plan.Record.RecordInterval/60))
	}
	return steps, nil
}

// ensureStandardCallback 本公司标准回调模板（没有则创建）
func (s *templateService) ensureStandardCallback(ctx context.Context, operator string) (*entity.LiveCallbackTemplate, error) {
	name := s.standardCallbackName(ctx)
	var t entity.LiveCallbackTemplate
	err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND NAME = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), name).First(&t).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		t = entity.LiveCallbackTemplate{SysCompanyID: companyOf(ctx), Name: name, Description: "本系统标准回调", Source: entity.TemplateSourceStandard, IsActive: "Y", CreateBy: operator}
		t.SyncStatus = entity.TemplateSyncPending
	}
	s.fillStandardCallback(&t)
	t.UpdateBy = operator
	if err := s.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, err
	}
	if err := s.pushCallback(ctx, &t); err != nil {
		return nil, errors.New(t.SyncError)
	}
	return &t, nil
}

// ensureStandardRecord 本公司标准录制模板（存储至公司 COS；没有则创建）
func (s *templateService) ensureStandardRecord(ctx context.Context, plan *StandardPlan, operator string) (*entity.LiveRecordTemplate, error) {
	name := s.standardRecordName(ctx)
	var t entity.LiveRecordTemplate
	err := s.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND NAME = ? AND IS_ACTIVE = 'Y'", companyOf(ctx), name).First(&t).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil && t.CloudTemplateID > 0 && t.SyncStatus == entity.TemplateSyncSynced {
		return &t, nil
	}
	if err != nil {
		t = entity.LiveRecordTemplate{SysCompanyID: companyOf(ctx), Name: name, Description: "本系统标准录制（MP4，存储至公司 COS）",
			CosStore: "Y", Formats: "mp4", RemoveWatermark: "Y", Source: entity.TemplateSourceStandard, IsActive: "Y", CreateBy: operator}
		t.SyncStatus = entity.TemplateSyncPending
	}
	t.CosBucket, t.CosRegion, t.CosPath, t.RecordInterval = plan.Record.CosBucket, plan.Record.CosRegion, plan.Record.CosPath, plan.Record.RecordInterval
	t.UpdateBy = operator
	if err := s.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, err
	}
	if err := s.pushRecord(ctx, &t); err != nil {
		return nil, errors.New(t.SyncError)
	}
	return &t, nil
}
