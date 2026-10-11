package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/handler"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/jwt"
	"github.com/sky-xhsoft/sky-server/internal/service/billing"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
)

// 计费菜单表：费用中心（直播中心，独立权限目录，默认只有公司管理员可见）、计费管理（系统中心，仅平台公司）
const (
	billingTable      = "LIVE_BILLING"
	billingAdminTable = "BILLING_ADMIN"
)

// registerBillingRoutes 注册计费路由
func registerBillingRoutes(r *gin.RouterGroup, jwtUtil *jwt.JWT, meter *billing.Meter, groupService groups.Service, metadataService metadata.Service) {
	h := handler.NewBillingHandler(meter)
	auth := middleware.AuthRequired(jwtUtil)
	read := middleware.TablePermissionByName(groupService, metadataService, billingTable, groups.PermRead)
	write := middleware.TablePermissionByName(groupService, metadataService, billingTable, groups.PermUpdate)

	// 本公司：用量、费用、账单、流水、套餐与资源包
	g := r.Group("/billing", auth)
	{
		g.GET("/usage/summary", read, h.Summary)
		g.GET("/usage/daily", read, h.Daily)
		g.GET("/usage/rooms", read, h.Rooms)
		g.GET("/usage/sessions", read, h.Sessions)
		g.GET("/account", read, h.Overview)
		g.GET("/charges", read, h.Charges)
		g.GET("/transactions", read, h.Transactions)
		g.GET("/statements", read, h.Statements)
		g.GET("/statements/:month", read, h.Statement)
		g.GET("/statements/:month/export", read, h.ExportStatement)
		g.GET("/products", read, h.OnSale)
		g.POST("/subscribe", write, h.Subscribe)
		g.POST("/auto-renew", write, h.AutoRenew)
		g.POST("/packs/buy", write, h.BuyPack)
	}

	// 平台：计费管理（处理函数内另校验平台公司）
	adminRead := middleware.TablePermissionByName(groupService, metadataService, billingAdminTable, groups.PermRead)
	adminWrite := middleware.TablePermissionByName(groupService, metadataService, billingAdminTable, groups.PermUpdate)
	a := r.Group("/billing/admin", auth)
	{
		a.GET("/reconcile", adminRead, h.Reconcile)
		a.POST("/recompute", adminWrite, h.Recompute)
		a.GET("/settings", adminRead, h.GetSettings)
		a.PUT("/settings", adminWrite, h.UpdateSettings)
		a.GET("/prices", adminRead, h.ListPrices)
		a.POST("/prices", adminWrite, h.CreatePrice)
		a.DELETE("/prices/:id", adminWrite, h.DeletePrice)
		a.GET("/plans", adminRead, h.ListPlans)
		a.POST("/plans", adminWrite, h.SavePlan)
		a.PUT("/plans/:id", adminWrite, h.SavePlan)
		a.GET("/pack-defs", adminRead, h.ListPackDefs)
		a.POST("/pack-defs", adminWrite, h.SavePackDef)
		a.PUT("/pack-defs/:id", adminWrite, h.SavePackDef)
		a.GET("/accounts", adminRead, h.Accounts)
		a.GET("/accounts/:companyId", adminRead, h.AccountOverview)
		a.GET("/accounts/:companyId/transactions", adminRead, h.AccountTransactions)
		a.GET("/accounts/:companyId/statements/:month/export", adminRead, h.AccountExport)
		a.POST("/accounts/:companyId/recharge", adminWrite, h.Recharge)
		a.POST("/accounts/:companyId/gift", adminWrite, h.Gift)
		a.POST("/accounts/:companyId/adjust", adminWrite, h.Adjust)
		a.PUT("/accounts/:companyId/credit", adminWrite, h.SetCredit)
		a.POST("/accounts/:companyId/subscribe", adminWrite, h.AdminSubscribe)
		a.POST("/accounts/:companyId/packs", adminWrite, h.GrantPack)
		a.GET("/requests", adminRead, h.Requests)
		a.POST("/requests/:id/review", adminWrite, h.ReviewRequest)
		a.POST("/statements", adminWrite, h.GenerateStatements)
		a.POST("/rerate", adminWrite, h.Rerate)
	}
}
