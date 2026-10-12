package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/pkg/vendortext"
)

// 公司配置·云盘存储：用云盘存储的密钥（单独设置的 COS 密钥，未设置时用直播密钥）列出、新建存储桶。
// 只操作腾讯云，不修改公司配置；选用哪个存储桶由前端填入 Bucket 地址后随配置一起保存

// storageBucketView 云盘存储可选的存储桶
type storageBucketView struct {
	Bucket  string `json:"bucket"`  // 含 -APPID，云盘配置使用完整名称
	Region  string `json:"region"`  // 地域
	URL     string `json:"url"`     // Bucket 地址
	Current bool   `json:"current"` // 当前配置的存储桶
	Created string `json:"creationDate,omitempty"`
}

// 测试中替换为本地实现
var (
	listStorageBuckets  = tencent.ListCosBuckets
	createStorageBucket = tencent.CreateCosBucket
	storageAppID        = tencent.AccountAppID
)

// storageKeys 云盘存储使用的腾讯云密钥：单独设置的 COS 密钥优先，否则用直播密钥
func storageKeys(conf *entity.SysCompanyConf) (id, key string, ok bool) {
	if conf == nil {
		return "", "", false
	}
	if strings.TrimSpace(conf.TencentCOSSecretID) != "" && strings.TrimSpace(conf.TencentCOSSecretKey) != "" {
		return conf.TencentCOSSecretID, conf.TencentCOSSecretKey, true
	}
	if strings.TrimSpace(conf.TencentCloudSecretID) != "" && strings.TrimSpace(conf.TencentCloudSecretKey) != "" {
		return conf.TencentCloudSecretID, conf.TencentCloudSecretKey, true
	}
	return "", "", false
}

// storageConf 当前公司配置与云盘存储密钥；失败时已写入响应
func (h *CompanyHandler) storageConf(c *gin.Context) (*entity.SysCompanyConf, string, string, bool) {
	id, ok := c.Get("companyID")
	companyID, ok2 := id.(uint)
	if !ok || !ok2 {
		utils.Unauthorized(c, "未获取到公司信息")
		return nil, "", "", false
	}
	conf, err := h.svc.GetCompanyConfig(c.Request.Context(), companyID)
	if err != nil {
		utils.FailFromError(c, "获取公司配置失败: ", err)
		return nil, "", "", false
	}
	secretID, secretKey, has := storageKeys(conf)
	if !has {
		utils.BadRequest(c, "请先保存云账号密钥（云盘存储单独设置的存储密钥，或直播服务的密钥）")
		return nil, "", "", false
	}
	return conf, secretID, secretKey, true
}

func toStorageBucket(conf *entity.SysCompanyConf, b tencent.CosBucket) storageBucketView {
	return storageBucketView{
		Bucket:  b.Name,
		Region:  b.Region,
		URL:     tencent.CosBucketURL(b.Name, b.Region),
		Current: conf.TencentCOSBucketName == b.Name && strings.TrimSpace(conf.TencentCOSRegion) == b.Region,
		Created: b.CreationDate,
	}
}

// ListStorageBuckets 云盘存储密钥对应账号下的存储桶：GET /company/config/cos-buckets
func (h *CompanyHandler) ListStorageBuckets(c *gin.Context) {
	conf, id, key, ok := h.storageConf(c)
	if !ok {
		return
	}
	buckets, err := listStorageBuckets(c.Request.Context(), id, key)
	if err != nil {
		utils.BadRequest(c, "获取存储桶失败："+vendortext.Sanitize(err.Error()))
		return
	}
	out := make([]storageBucketView, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, toStorageBucket(conf, b))
	}
	utils.Success(c, out)
}

// CreateStorageBucket 新建存储桶（公开读写 + CORS 全部允许）：POST /company/config/cos-buckets
func (h *CompanyHandler) CreateStorageBucket(c *gin.Context) {
	var req struct {
		Bucket string `json:"bucket"`
		Region string `json:"region"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	conf, id, key, ok := h.storageConf(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	buckets, err := listStorageBuckets(ctx, id, key)
	if err != nil {
		utils.BadRequest(c, "获取存储桶失败："+vendortext.Sanitize(err.Error()))
		return
	}
	_, full, region, err := tencent.ResolveNewCosBucket(buckets, storageAppID(ctx, id, key, buckets), req.Bucket, req.Region)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if err := createStorageBucket(ctx, id, key, full, region); err != nil {
		utils.BadRequest(c, tencent.CosCreateError(err).Error())
		return
	}
	utils.Success(c, toStorageBucket(conf, tencent.CosBucket{Name: full, Region: region}))
}
