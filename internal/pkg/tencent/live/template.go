package live

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	live "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
	"go.uber.org/zap"
)

// TemplateAPI 回调、录制模板与规则用到的腾讯云直播接口（*live.Client 实现；测试中可替换）
type TemplateAPI interface {
	CreateLiveCallbackTemplate(*live.CreateLiveCallbackTemplateRequest) (*live.CreateLiveCallbackTemplateResponse, error)
	ModifyLiveCallbackTemplate(*live.ModifyLiveCallbackTemplateRequest) (*live.ModifyLiveCallbackTemplateResponse, error)
	DeleteLiveCallbackTemplate(*live.DeleteLiveCallbackTemplateRequest) (*live.DeleteLiveCallbackTemplateResponse, error)
	DescribeLiveCallbackTemplates(*live.DescribeLiveCallbackTemplatesRequest) (*live.DescribeLiveCallbackTemplatesResponse, error)
	DescribeLiveCallbackRules(*live.DescribeLiveCallbackRulesRequest) (*live.DescribeLiveCallbackRulesResponse, error)
	CreateLiveCallbackRule(*live.CreateLiveCallbackRuleRequest) (*live.CreateLiveCallbackRuleResponse, error)
	DeleteLiveCallbackRule(*live.DeleteLiveCallbackRuleRequest) (*live.DeleteLiveCallbackRuleResponse, error)

	CreateLiveRecordTemplate(*live.CreateLiveRecordTemplateRequest) (*live.CreateLiveRecordTemplateResponse, error)
	ModifyLiveRecordTemplate(*live.ModifyLiveRecordTemplateRequest) (*live.ModifyLiveRecordTemplateResponse, error)
	DeleteLiveRecordTemplate(*live.DeleteLiveRecordTemplateRequest) (*live.DeleteLiveRecordTemplateResponse, error)
	DescribeLiveRecordTemplates(*live.DescribeLiveRecordTemplatesRequest) (*live.DescribeLiveRecordTemplatesResponse, error)
	DescribeLiveRecordRules(*live.DescribeLiveRecordRulesRequest) (*live.DescribeLiveRecordRulesResponse, error)
	CreateLiveRecordRule(*live.CreateLiveRecordRuleRequest) (*live.CreateLiveRecordRuleResponse, error)
	DeleteLiveRecordRule(*live.DeleteLiveRecordRuleRequest) (*live.DeleteLiveRecordRuleResponse, error)
}

var (
	_ TemplateAPI    = (*live.Client)(nil)
	_ TemplateClient = (*TemplateManager)(nil)
)

// TemplateClient 模板与规则操作（服务层依赖此接口，便于测试）
type TemplateClient interface {
	ListCallbackTemplates() ([]*CallbackTemplate, error)
	CreateCallbackTemplate(t *CallbackTemplate) (int64, error)
	ModifyCallbackTemplate(t *CallbackTemplate) error
	DeleteCallbackTemplate(templateID int64) error
	ListCallbackRules() ([]*TemplateRule, error)
	CreateCallbackRule(domain, appName string, templateID int64) error
	DeleteCallbackRule(domain, appName string) error

	ListRecordTemplates() ([]*RecordTemplate, error)
	CreateRecordTemplate(t *RecordTemplate) (int64, error)
	ModifyRecordTemplate(t *RecordTemplate) error
	DeleteRecordTemplate(templateID int64) error
	ListRecordRules() ([]*TemplateRule, error)
	CreateRecordRule(domain, appName, streamName string, templateID int64) error
	DeleteRecordRule(domain, appName, streamName string) error
}

// CallbackTemplate 回调模板（只包含本系统管理的回调；截图、鉴黄、音频审核、高光切片等未传的项保持云端原值）
type CallbackTemplate struct {
	TemplateID            int64
	Name                  string
	Description           string
	StreamBeginURL        string
	StreamEndURL          string
	RecordURL             string
	RecordStatusURL       string // 腾讯云查询接口不返回此项，导入时为空
	RecordExceptionURL    string
	RecordExceptionLevels []string
	PushExceptionURL      string
	CallbackKey           string
}

// RecordTemplate 录制模板（本系统只管理 MP4 录制；其他格式仅在导入时记录）
type RecordTemplate struct {
	TemplateID      int64
	Name            string
	Description     string
	CosStore        bool   // 是否存储至 COS（否则为云点播 VOD）
	CosBucket       string // 桶名，不含 -appid
	CosRegion       string
	CosPath         string // 以 / 开头
	RecordInterval  int64  // MP4 分段时长（秒）
	RemoveWatermark bool   // true：录制原始流
	Formats         []string
}

// TemplateRule 模板绑定规则（域名 + AppName (+ StreamName) 唯一）
type TemplateRule struct {
	DomainName string
	AppName    string
	StreamName string
	TemplateID int64
}

// TemplateManager 腾讯云模板与规则管理
type TemplateManager struct {
	api      TemplateAPI
	provider RegionClientProvider
}

// NewTemplateManager 创建模板管理器；provider 用于公司地域不支持直播接口时改用默认直播地域重试
func NewTemplateManager(api TemplateAPI, provider RegionClientProvider) *TemplateManager {
	return &TemplateManager{api: api, provider: provider}
}

// callTemplate 调用模板接口；返回 UnsupportedRegion 时改用 fallbackLiveRegion 重试
func callTemplate[T any](m *TemplateManager, call func(api TemplateAPI) (T, error)) (T, error) {
	resp, err := call(m.api)
	if err == nil || m.provider == nil || !isUnsupportedRegion(err) {
		return resp, err
	}
	client, cerr := m.provider.GetLiveClientForRegion(fallbackLiveRegion)
	if cerr != nil {
		return resp, err
	}
	zap.L().Warn("公司配置的地域不支持直播接口，改用默认直播地域重试", zap.String("region", fallbackLiveRegion), zap.Error(err))
	return call(client)
}

// TencentErrorCode 取腾讯云错误码（非腾讯云错误返回空）
func TencentErrorCode(err error) string {
	var sdkErr *tcerr.TencentCloudSDKError
	if errors.As(err, &sdkErr) {
		return sdkErr.Code
	}
	return ""
}

// IsTencentNotFound 记录或规则在云端已不存在（删除时可视为成功）
func IsTencentNotFound(err error) bool {
	switch TencentErrorCode(err) {
	case "FailedOperation.NotFound", "InternalError.NotFound", "InternalError.RuleNotFound",
		"InternalError.ConfNotFound", "InvalidParameter.ConfNotFound", "ResourceNotFound":
		return true
	}
	return false
}

// TemplateErrorMessage 把腾讯云错误转为中文提示
func TemplateErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msgs := map[string]string{
		"InternalError.ConfOutLimit":          "腾讯云模板数量已达上限（每个账号最多 50 个）",
		"InternalError.RuleAlreadyExist":      "该域名已存在相同规则",
		"FailedOperation.RuleAlreadyExist":    "该域名已存在相同规则",
		"InternalError.ConfInUsed":            "模板正在被域名使用，请先解绑",
		"FailedOperation.ConfInUsed":          "模板正在被域名使用，请先解绑",
		"InternalError.ConfNotFound":          "腾讯云上的模板不存在",
		"InvalidParameter.ConfNotFound":       "腾讯云上的模板不存在",
		"InvalidParameter.UrlNotSafe":         "回调地址无法解析，请检查域名",
		"InvalidParameter.ArgsNotMatch":       "模板名称不符合要求",
		"InvalidParameter.DomainFormatError":  "域名格式错误",
		"ResourceNotFound.StopService":        "腾讯云账号已停服（可能欠费），请先恢复服务",
		"ResourceNotFound.FreezeService":      "腾讯云直播服务已被冻结",
		"ResourceNotFound.ForbidService":      "腾讯云账号已被禁用",
		"ResourceNotFound.UserDisableService": "腾讯云直播服务已被主动停用",
		"AuthFailure.SecretIdNotFound":        "腾讯云密钥不存在，请检查公司配置",
		"AuthFailure.SignatureFailure":        "腾讯云密钥错误，请检查公司配置",
		"AuthFailure.UnauthorizedOperation":   "腾讯云子账号无直播模板权限",
	}
	code := TencentErrorCode(err)
	if msg, ok := msgs[code]; ok {
		return msg + "（" + code + "）"
	}
	var sdkErr *tcerr.TencentCloudSDKError
	if errors.As(err, &sdkErr) {
		return fmt.Sprintf("腾讯云返回错误：%s（%s）", sdkErr.Message, sdkErr.Code)
	}
	return err.Error()
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func i64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// ==================== 回调模板 ====================

func callbackFromInfo(t *live.CallBackTemplateInfo) *CallbackTemplate {
	levels := make([]string, 0, len(t.RecordExceptionLevels))
	for _, l := range t.RecordExceptionLevels {
		levels = append(levels, str(l))
	}
	return &CallbackTemplate{
		TemplateID:            i64(t.TemplateId),
		Name:                  str(t.TemplateName),
		Description:           str(t.Description),
		StreamBeginURL:        str(t.StreamBeginNotifyUrl),
		StreamEndURL:          str(t.StreamEndNotifyUrl),
		RecordURL:             str(t.RecordNotifyUrl),
		RecordExceptionURL:    str(t.RecordExceptionNotifyUrl),
		RecordExceptionLevels: levels,
		PushExceptionURL:      str(t.PushExceptionNotifyUrl),
		CallbackKey:           str(t.CallbackKey),
	}
}

// ListCallbackTemplates 账号下全部回调模板
func (m *TemplateManager) ListCallbackTemplates() ([]*CallbackTemplate, error) {
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.DescribeLiveCallbackTemplatesResponse, error) {
		return api.DescribeLiveCallbackTemplates(live.NewDescribeLiveCallbackTemplatesRequest())
	})
	if err != nil {
		return nil, err
	}
	out := make([]*CallbackTemplate, 0, len(resp.Response.Templates))
	for _, t := range resp.Response.Templates {
		out = append(out, callbackFromInfo(t))
	}
	return out, nil
}

func levelPtrs(levels []string) []*string {
	out := make([]*string, 0, len(levels))
	for _, l := range levels {
		out = append(out, common.StringPtr(l))
	}
	return out
}

// CreateCallbackTemplate 创建回调模板，返回云端模板 ID
func (m *TemplateManager) CreateCallbackTemplate(t *CallbackTemplate) (int64, error) {
	req := live.NewCreateLiveCallbackTemplateRequest()
	req.TemplateName = common.StringPtr(t.Name)
	req.Description = common.StringPtr(t.Description)
	req.StreamBeginNotifyUrl = common.StringPtr(t.StreamBeginURL)
	req.StreamEndNotifyUrl = common.StringPtr(t.StreamEndURL)
	req.RecordNotifyUrl = common.StringPtr(t.RecordURL)
	req.RecordStatusNotifyUrl = common.StringPtr(t.RecordStatusURL)
	req.RecordExceptionNotifyUrl = common.StringPtr(t.RecordExceptionURL)
	req.PushExceptionNotifyUrl = common.StringPtr(t.PushExceptionURL)
	req.CallbackKey = common.StringPtr(t.CallbackKey)
	if t.RecordExceptionURL != "" && len(t.RecordExceptionLevels) > 0 {
		req.RecordExceptionLevels = levelPtrs(t.RecordExceptionLevels)
	}
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.CreateLiveCallbackTemplateResponse, error) {
		return api.CreateLiveCallbackTemplate(req)
	})
	if err != nil {
		return 0, err
	}
	return i64(resp.Response.TemplateId), nil
}

// ModifyCallbackTemplate 修改回调模板：只传本系统管理的回调项，其余项（如控制台配置的高光切片回调）不传，保持原值
func (m *TemplateManager) ModifyCallbackTemplate(t *CallbackTemplate) error {
	req := live.NewModifyLiveCallbackTemplateRequest()
	req.TemplateId = common.Int64Ptr(t.TemplateID)
	req.TemplateName = common.StringPtr(t.Name)
	req.Description = common.StringPtr(t.Description)
	req.StreamBeginNotifyUrl = common.StringPtr(t.StreamBeginURL)
	req.StreamEndNotifyUrl = common.StringPtr(t.StreamEndURL)
	req.RecordNotifyUrl = common.StringPtr(t.RecordURL)
	req.RecordStatusNotifyUrl = common.StringPtr(t.RecordStatusURL)
	req.RecordExceptionNotifyUrl = common.StringPtr(t.RecordExceptionURL)
	req.PushExceptionNotifyUrl = common.StringPtr(t.PushExceptionURL)
	req.CallbackKey = common.StringPtr(t.CallbackKey)
	if t.RecordExceptionURL != "" && len(t.RecordExceptionLevels) > 0 {
		req.RecordExceptionLevels = levelPtrs(t.RecordExceptionLevels)
	}
	_, err := callTemplate(m, func(api TemplateAPI) (*live.ModifyLiveCallbackTemplateResponse, error) {
		return api.ModifyLiveCallbackTemplate(req)
	})
	return err
}

// DeleteCallbackTemplate 删除回调模板
func (m *TemplateManager) DeleteCallbackTemplate(templateID int64) error {
	req := live.NewDeleteLiveCallbackTemplateRequest()
	req.TemplateId = common.Int64Ptr(templateID)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.DeleteLiveCallbackTemplateResponse, error) {
		return api.DeleteLiveCallbackTemplate(req)
	})
	return err
}

// ListCallbackRules 账号下全部回调规则（接口不支持按域名过滤）
func (m *TemplateManager) ListCallbackRules() ([]*TemplateRule, error) {
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.DescribeLiveCallbackRulesResponse, error) {
		return api.DescribeLiveCallbackRules(live.NewDescribeLiveCallbackRulesRequest())
	})
	if err != nil {
		return nil, err
	}
	out := make([]*TemplateRule, 0, len(resp.Response.Rules))
	for _, r := range resp.Response.Rules {
		out = append(out, &TemplateRule{DomainName: str(r.DomainName), AppName: str(r.AppName), TemplateID: i64(r.TemplateId)})
	}
	return out, nil
}

// CreateCallbackRule 绑定回调模板到 域名 + AppName
func (m *TemplateManager) CreateCallbackRule(domain, appName string, templateID int64) error {
	req := live.NewCreateLiveCallbackRuleRequest()
	req.DomainName = common.StringPtr(domain)
	req.AppName = common.StringPtr(appName)
	req.TemplateId = common.Int64Ptr(templateID)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.CreateLiveCallbackRuleResponse, error) {
		return api.CreateLiveCallbackRule(req)
	})
	return err
}

// DeleteCallbackRule 解绑回调规则：AppName 须与规则完全一致（为空也要传空字符串）
func (m *TemplateManager) DeleteCallbackRule(domain, appName string) error {
	req := live.NewDeleteLiveCallbackRuleRequest()
	req.DomainName = common.StringPtr(domain)
	req.AppName = common.StringPtr(appName)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.DeleteLiveCallbackRuleResponse, error) {
		return api.DeleteLiveCallbackRule(req)
	})
	return err
}

// ==================== 录制模板 ====================

func recordFromInfo(t *live.RecordTemplateInfo) *RecordTemplate {
	out := &RecordTemplate{
		TemplateID:  i64(t.TemplateId),
		Name:        str(t.TemplateName),
		Description: str(t.Description),
		CosStore:    i64(t.CosStore) == 1,
	}
	if t.RemoveWatermark != nil {
		out.RemoveWatermark = *t.RemoveWatermark
	}
	params := []struct {
		name string
		p    *live.RecordParam
	}{{"mp4", t.Mp4Param}, {"flv", t.FlvParam}, {"hls", t.HlsParam}, {"aac", t.AacParam}, {"mp3", t.Mp3Param}}
	for _, f := range params {
		if f.p == nil || i64(f.p.Enable) != 1 {
			continue
		}
		out.Formats = append(out.Formats, f.name)
		// 存储位置与分段取第一个开启的格式（本系统创建的模板只开 MP4）
		if out.RecordInterval == 0 {
			out.RecordInterval = i64(f.p.RecordInterval)
			out.CosBucket = str(f.p.CosBucketName)
			out.CosRegion = str(f.p.CosBucketRegion)
			out.CosPath = str(f.p.CosBucketPath)
		}
	}
	return out
}

// ListRecordTemplates 账号下全部录制模板
func (m *TemplateManager) ListRecordTemplates() ([]*RecordTemplate, error) {
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.DescribeLiveRecordTemplatesResponse, error) {
		return api.DescribeLiveRecordTemplates(live.NewDescribeLiveRecordTemplatesRequest())
	})
	if err != nil {
		return nil, err
	}
	out := make([]*RecordTemplate, 0, len(resp.Response.Templates))
	for _, t := range resp.Response.Templates {
		out = append(out, recordFromInfo(t))
	}
	return out, nil
}

// mp4Param 本系统录制模板：只开 MP4，存储至 COS
func mp4Param(t *RecordTemplate) *live.RecordParam {
	return &live.RecordParam{
		Enable:          common.Int64Ptr(1),
		RecordInterval:  common.Int64Ptr(t.RecordInterval),
		CosBucketName:   common.StringPtr(t.CosBucket),
		CosBucketRegion: common.StringPtr(t.CosRegion),
		CosBucketPath:   common.StringPtr(t.CosPath),
	}
}

// CreateRecordTemplate 创建录制模板（存储至 COS 的 MP4 录制），返回云端模板 ID
func (m *TemplateManager) CreateRecordTemplate(t *RecordTemplate) (int64, error) {
	req := live.NewCreateLiveRecordTemplateRequest()
	req.TemplateName = common.StringPtr(t.Name)
	req.Description = common.StringPtr(t.Description)
	req.CosStore = common.Int64Ptr(1)
	req.Mp4Param = mp4Param(t)
	req.RemoveWatermark = common.BoolPtr(t.RemoveWatermark)
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.CreateLiveRecordTemplateResponse, error) {
		return api.CreateLiveRecordTemplate(req)
	})
	if err != nil {
		return 0, err
	}
	return i64(resp.Response.TemplateId), nil
}

// ModifyRecordTemplate 修改录制模板
func (m *TemplateManager) ModifyRecordTemplate(t *RecordTemplate) error {
	req := live.NewModifyLiveRecordTemplateRequest()
	req.TemplateId = common.Int64Ptr(t.TemplateID)
	req.TemplateName = common.StringPtr(t.Name)
	req.Description = common.StringPtr(t.Description)
	// 存储方式（COS / VOD）创建后不能修改，修改接口没有 CosStore
	req.Mp4Param = mp4Param(t)
	req.RemoveWatermark = common.BoolPtr(t.RemoveWatermark)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.ModifyLiveRecordTemplateResponse, error) {
		return api.ModifyLiveRecordTemplate(req)
	})
	return err
}

// DeleteRecordTemplate 删除录制模板
func (m *TemplateManager) DeleteRecordTemplate(templateID int64) error {
	req := live.NewDeleteLiveRecordTemplateRequest()
	req.TemplateId = common.Int64Ptr(templateID)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.DeleteLiveRecordTemplateResponse, error) {
		return api.DeleteLiveRecordTemplate(req)
	})
	return err
}

// ListRecordRules 账号下全部录制规则（接口不支持按域名过滤）
func (m *TemplateManager) ListRecordRules() ([]*TemplateRule, error) {
	resp, err := callTemplate(m, func(api TemplateAPI) (*live.DescribeLiveRecordRulesResponse, error) {
		return api.DescribeLiveRecordRules(live.NewDescribeLiveRecordRulesRequest())
	})
	if err != nil {
		return nil, err
	}
	out := make([]*TemplateRule, 0, len(resp.Response.Rules))
	for _, r := range resp.Response.Rules {
		out = append(out, &TemplateRule{DomainName: str(r.DomainName), AppName: str(r.AppName), StreamName: str(r.StreamName), TemplateID: i64(r.TemplateId)})
	}
	return out, nil
}

// CreateRecordRule 绑定录制模板到 域名 + AppName (+ StreamName)
func (m *TemplateManager) CreateRecordRule(domain, appName, streamName string, templateID int64) error {
	req := live.NewCreateLiveRecordRuleRequest()
	req.DomainName = common.StringPtr(domain)
	req.AppName = common.StringPtr(appName)
	req.StreamName = common.StringPtr(streamName)
	req.TemplateId = common.Int64Ptr(templateID)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.CreateLiveRecordRuleResponse, error) {
		return api.CreateLiveRecordRule(req)
	})
	return err
}

// DeleteRecordRule 解绑录制规则：域名 + AppName + StreamName 须与规则完全一致（为空也要传空字符串）
func (m *TemplateManager) DeleteRecordRule(domain, appName, streamName string) error {
	req := live.NewDeleteLiveRecordRuleRequest()
	req.DomainName = common.StringPtr(domain)
	req.AppName = common.StringPtr(appName)
	req.StreamName = common.StringPtr(streamName)
	_, err := callTemplate(m, func(api TemplateAPI) (*live.DeleteLiveRecordRuleResponse, error) {
		return api.DeleteLiveRecordRule(req)
	})
	return err
}

// StripBucketAppID 去掉桶名末尾的 -appid（腾讯云录制模板要求桶名不含 appid）
func StripBucketAppID(bucket string) string {
	b := strings.TrimSpace(bucket)
	i := strings.LastIndex(b, "-")
	if i <= 0 {
		return b
	}
	suffix := b[i+1:]
	if suffix == "" {
		return b
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return b
		}
	}
	return b[:i]
}
