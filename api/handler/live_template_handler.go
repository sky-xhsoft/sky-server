package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// LiveTemplateHandler 直播回调、录制模板管理（本地为准，保存后同步到公司的腾讯云账号）
type LiveTemplateHandler struct {
	svc liveService.TemplateService
}

// NewLiveTemplateHandler 创建模板处理器
func NewLiveTemplateHandler(svc liveService.TemplateService) *LiveTemplateHandler {
	return &LiveTemplateHandler{svc: svc}
}

// templateFail 业务错误均为中文提示，直接返回；模板不存在返回 404
func templateFail(c *gin.Context, err error) {
	if errors.Is(err, liveService.ErrTemplateNotFound) {
		utils.NotFound(c, err.Error())
		return
	}
	utils.BadRequest(c, err.Error())
}

func templateKind(c *gin.Context) (string, bool) {
	kind := c.Param("kind")
	if kind != entity.TemplateKindCallback && kind != entity.TemplateKindRecord {
		utils.BadRequest(c, "模板类型只能是 callback 或 record")
		return "", false
	}
	return kind, true
}

func templateID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.BadRequest(c, "模板ID错误")
		return 0, false
	}
	return uint(id), true
}

// Defaults 新建模板的默认值（系统回调地址、公司 COS）
func (h *LiveTemplateHandler) Defaults(c *gin.Context) {
	d, err := h.svc.Defaults(withCompanyID(c))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, d)
}

// CosBuckets 公司腾讯云账号下的存储桶（录制模板选择存储位置）
func (h *LiveTemplateHandler) CosBuckets(c *gin.Context) {
	list, err := h.svc.CosBuckets(withCompanyID(c))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, list)
}

// CreateCosBucket 在公司腾讯云账号下新建存储桶（公开读写 + CORS 全部允许）：POST /live/templates/cos-buckets
func (h *LiveTemplateHandler) CreateCosBucket(c *gin.Context) {
	var req struct {
		Bucket string `json:"bucket"`
		Region string `json:"region"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	b, err := h.svc.CreateCosBucket(withCompanyID(c), req.Bucket, req.Region)
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, b)
}

// List 模板列表（同时与腾讯云对账）：GET /live/templates?kind=callback|record
func (h *LiveTemplateHandler) List(c *gin.Context) {
	ctx := withCompanyID(c)
	if c.Query("kind") == entity.TemplateKindRecord {
		res, err := h.svc.ListRecordTemplates(ctx)
		if err != nil {
			templateFail(c, err)
			return
		}
		utils.Success(c, res)
		return
	}
	res, err := h.svc.ListCallbackTemplates(ctx)
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, res)
}

// SaveCallback 新增（POST）或修改（PUT /:id）回调模板
func (h *LiveTemplateHandler) SaveCallback(c *gin.Context) {
	var id uint
	if c.Param("id") != "" {
		var ok bool
		if id, ok = templateID(c); !ok {
			return
		}
	}
	var in entity.LiveCallbackTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	t, err := h.svc.SaveCallbackTemplate(withCompanyID(c), id, &in, c.GetString("username"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, t)
}

// SaveRecord 新增（POST）或修改（PUT /:id）录制模板
func (h *LiveTemplateHandler) SaveRecord(c *gin.Context) {
	var id uint
	if c.Param("id") != "" {
		var ok bool
		if id, ok = templateID(c); !ok {
			return
		}
	}
	var in entity.LiveRecordTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	t, err := h.svc.SaveRecordTemplate(withCompanyID(c), id, &in, c.GetString("username"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, t)
}

// Delete 删除模板（已绑定的须先解绑）
func (h *LiveTemplateHandler) Delete(c *gin.Context) {
	kind, ok := templateKind(c)
	if !ok {
		return
	}
	id, ok := templateID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteTemplate(withCompanyID(c), kind, id); err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, nil)
}

// Sync 以本地为准同步到腾讯云（重新同步、以本地覆盖云端）
func (h *LiveTemplateHandler) Sync(c *gin.Context) {
	kind, ok := templateKind(c)
	if !ok {
		return
	}
	id, ok := templateID(c)
	if !ok {
		return
	}
	if err := h.svc.SyncTemplate(withCompanyID(c), kind, id); err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, nil)
}

// Pull 用腾讯云上的配置更新本地
func (h *LiveTemplateHandler) Pull(c *gin.Context) {
	kind, ok := templateKind(c)
	if !ok {
		return
	}
	id, ok := templateID(c)
	if !ok {
		return
	}
	if err := h.svc.PullTemplate(withCompanyID(c), kind, id); err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, nil)
}

// Import 从腾讯云导入已有模板与规则
func (h *LiveTemplateHandler) Import(c *gin.Context) {
	res, err := h.svc.Import(withCompanyID(c), c.GetString("username"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, res)
}

// DomainOverview 推流域名的模板配置概览与体检
func (h *LiveTemplateHandler) DomainOverview(c *gin.Context) {
	ov, err := h.svc.DomainOverview(withCompanyID(c), c.Param("domainName"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, ov)
}

// Bind 绑定模板到推流域名：{kind, templateId, appName, streamName}
func (h *LiveTemplateHandler) Bind(c *gin.Context) {
	var req liveService.BindTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	req.DomainName = c.Param("domainName")
	rule, err := h.svc.Bind(withCompanyID(c), &req, c.GetString("username"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, rule)
}

// Unbind 解绑：DELETE /live/domains/:domainName/templates/:kind?appName=&streamName=
func (h *LiveTemplateHandler) Unbind(c *gin.Context) {
	kind, ok := templateKind(c)
	if !ok {
		return
	}
	// AppName 须与规则完全一致：未传时按 live，显式传空表示空 AppName 的规则
	appName, has := c.GetQuery("appName")
	if !has {
		appName = liveService.StandardAppName
	}
	if err := h.svc.Unbind(withCompanyID(c), kind, c.Param("domainName"), appName, c.Query("streamName")); err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, nil)
}

// StandardPlan 应用标准配置的预览
func (h *LiveTemplateHandler) StandardPlan(c *gin.Context) {
	p, err := h.svc.StandardPlan(withCompanyID(c), c.Param("domainName"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, p)
}

// ApplyStandard 应用标准配置：{callbackMode: complete|replace, replaceRecord}
func (h *LiveTemplateHandler) ApplyStandard(c *gin.Context) {
	var opt liveService.StandardOptions
	if err := c.ShouldBindJSON(&opt); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	steps, err := h.svc.ApplyStandard(withCompanyID(c), c.Param("domainName"), opt, c.GetString("username"))
	if err != nil {
		templateFail(c, err)
		return
	}
	utils.Success(c, gin.H{"steps": steps})
}
