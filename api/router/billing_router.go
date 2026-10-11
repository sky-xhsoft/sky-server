package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/billing"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
)

// registerBillingRoutes 注册计费路由（一期：用量）。
// 权限沿用直播间管理（表 54）的查询权限：能管理直播间的人可以查看本公司用量；全部公司的对账只对平台公司开放
func registerBillingRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, meter *billing.Meter, groupService groups.Service) {
	h := handler.NewBillingHandler(meter)
	group := r.Group("/billing")
	group.Use(middleware.AuthRequired(jwtUtil), middleware.TablePermissionRequired(groupService, 54, groups.PermRead))
	{
		group.GET("/usage/summary", h.Summary)
		group.GET("/usage/daily", h.Daily)
		group.GET("/usage/rooms", h.Rooms)
		group.GET("/usage/sessions", h.Sessions)
		group.GET("/admin/reconcile", h.Reconcile)
		group.POST("/admin/recompute", h.Recompute)
	}
}
