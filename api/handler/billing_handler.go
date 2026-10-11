package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/billing"
)

// platformCompanyID 平台公司（可使用系统中心），只有它能查看全部公司的用量与对账
const platformCompanyID = 1

// BillingHandler 计费（一期：用量）
type BillingHandler struct {
	meter *billing.Meter
}

// NewBillingHandler 计费处理器
func NewBillingHandler(meter *billing.Meter) *BillingHandler {
	return &BillingHandler{meter: meter}
}

func (h *BillingHandler) company(c *gin.Context) (uint, bool) {
	id := getCompanyID(c)
	if id == 0 {
		utils.Unauthorized(c, "未获取到公司信息")
		return 0, false
	}
	return id, true
}

// Summary 本公司某月用量汇总：GET /billing/usage/summary?month=YYYY-MM
func (h *BillingHandler) Summary(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Summary(c.Request.Context(), id, c.Query("month"))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, out)
}

// Daily 本公司某月每天的用量：GET /billing/usage/daily?month=YYYY-MM
func (h *BillingHandler) Daily(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Daily(c.Request.Context(), id, c.Query("month"))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, out)
}

// Rooms 本公司某月各直播间的用量：GET /billing/usage/rooms?month=YYYY-MM
func (h *BillingHandler) Rooms(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Rooms(c.Request.Context(), id, c.Query("month"))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, out)
}

// Sessions 本公司某月的推流记录：GET /billing/usage/sessions?month=YYYY-MM&roomId=&page=&pageSize=
func (h *BillingHandler) Sessions(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	roomID, _ := strconv.ParseUint(c.Query("roomId"), 10, 32)
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	out, err := h.meter.Sessions(c.Request.Context(), id, uint(roomID), c.Query("month"), page, pageSize)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, out)
}

// platformOnly 只有平台公司可以访问
func (h *BillingHandler) platformOnly(c *gin.Context) bool {
	if getCompanyID(c) != platformCompanyID {
		utils.Forbidden(c, "只有平台可以查看全部公司的用量")
		return false
	}
	return true
}

// Reconcile 平台：某月全部公司用量与流量对账：GET /billing/admin/reconcile?month=YYYY-MM
func (h *BillingHandler) Reconcile(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.ReconcileMonth(c.Request.Context(), c.Query("month"))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, out)
}

type recomputeRequest struct {
	From   string   `json:"from" binding:"required"`
	To     string   `json:"to" binding:"required"`
	Meters []string `json:"meters"`
}

// Recompute 平台：重新计量一段日期：POST /billing/admin/recompute
func (h *BillingHandler) Recompute(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req recomputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写开始和结束日期")
		return
	}
	if err := h.meter.Recompute(c.Request.Context(), req.From, req.To, req.Meters); err != nil {
		// 部分账号拉取流量失败时其余结果已写入，返回失败原因
		utils.BadRequest(c, "重新计量未全部完成："+strings.ReplaceAll(err.Error(), "\n", "；"))
		return
	}
	utils.Success(c, nil)
}
