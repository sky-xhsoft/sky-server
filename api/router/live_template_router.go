package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
)

// liveDomainTableID 直播域名的表权限（模板管理与之共用）
const liveDomainTableID = 35

// registerLiveTemplateRoutes 直播回调、录制模板管理与域名绑定
func registerLiveTemplateRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, svc liveService.TemplateService, groupService groups.Service) {
	if svc == nil {
		return
	}
	h := handler.NewLiveTemplateHandler(svc)
	auth := []gin.HandlerFunc{middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, liveDomainTableID, groups.PermRead)}
	write := middleware.TablePermissionRequired(groupService, liveDomainTableID, groups.PermUpdate)

	templates := r.Group("/live/templates", auth...)
	{
		templates.GET("", h.List)                                // 模板列表（kind=callback|record，同时与腾讯云对账）
		templates.GET("/defaults", h.Defaults)                   // 新建模板默认值
		templates.GET("/cos-buckets", h.CosBuckets)              // 公司腾讯云账号下的存储桶
		templates.POST("/cos-buckets", write, h.CreateCosBucket) // 新建存储桶（公开读写 + CORS）
		templates.POST("/import", write, h.Import)               // 从腾讯云导入
		templates.POST("/callback", write, h.SaveCallback)       // 新增回调模板（保存即同步）
		templates.PUT("/callback/:id", write, h.SaveCallback)    // 修改回调模板
		templates.POST("/record", write, h.SaveRecord)           // 新增录制模板
		templates.PUT("/record/:id", write, h.SaveRecord)        // 修改录制模板
		templates.DELETE("/:kind/:id", write, h.Delete)          // 删除模板
		templates.POST("/:kind/:id/sync", write, h.Sync)         // 以本地为准同步到腾讯云
		templates.POST("/:kind/:id/pull", write, h.Pull)         // 用腾讯云配置更新本地
	}

	domain := r.Group("/live/domains/:domainName/templates", auth...)
	{
		domain.GET("", h.DomainOverview)                 // 域名模板概览与体检
		domain.POST("/bind", write, h.Bind)              // 绑定模板
		domain.DELETE("/:kind", write, h.Unbind)         // 解绑
		domain.GET("/standard", h.StandardPlan)          // 标准配置预览
		domain.POST("/standard", write, h.ApplyStandard) // 应用标准配置
	}
}
