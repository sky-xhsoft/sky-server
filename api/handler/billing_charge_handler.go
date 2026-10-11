package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/billing"
)

// 计费二期接口：本公司的费用、账单、流水、套餐与资源包；平台的价格、套餐、账户与复核。
// 金额以「厘」（0.001 元）整数传递，单价以 0.0001 元整数传递

func operator(c *gin.Context) billing.Operator {
	op := billing.Operator{Name: c.GetString("username")}
	if v, ok := c.Get("userID"); ok {
		op.UserID, _ = v.(uint)
	}
	return op
}

func paramUint(c *gin.Context, name string) uint {
	v, _ := strconv.ParseUint(c.Param(name), 10, 32)
	return uint(v)
}

func (h *BillingHandler) fail(c *gin.Context, err error) {
	utils.BadRequest(c, err.Error())
}

// ---------- 本公司 ----------

// Overview 费用概览：GET /billing/account
func (h *BillingHandler) Overview(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Overview(c.Request.Context(), id)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// Charges 费用明细：GET /billing/charges?month=
func (h *BillingHandler) Charges(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Charges(c.Request.Context(), id, c.Query("month"))
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

func (h *BillingHandler) transactions(c *gin.Context, companyID uint) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("pageSize"))
	out, err := h.meter.Transactions(c.Request.Context(), companyID, c.Query("period"), c.Query("type"), page, size)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// Transactions 余额流水：GET /billing/transactions?period=&type=&page=&pageSize=
func (h *BillingHandler) Transactions(c *gin.Context) {
	if id, ok := h.company(c); ok {
		h.transactions(c, id)
	}
}

// Statements 月账单列表：GET /billing/statements
func (h *BillingHandler) Statements(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Statements(c.Request.Context(), id)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// Statement 月账单详情：GET /billing/statements/:month
func (h *BillingHandler) Statement(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	out, err := h.meter.Statement(c.Request.Context(), id, c.Param("month"))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, out)
}

func (h *BillingHandler) exportStatement(c *gin.Context, companyID uint) {
	month := c.Param("month")
	b, err := h.meter.ExportStatement(c.Request.Context(), companyID, month)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	name := fmt.Sprintf("账单-%s.xlsx", month)
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b)
}

// ExportStatement 导出月账单：GET /billing/statements/:month/export
func (h *BillingHandler) ExportStatement(c *gin.Context) {
	if id, ok := h.company(c); ok {
		h.exportStatement(c, id)
	}
}

// OnSale 在售的套餐与资源包：GET /billing/products
func (h *BillingHandler) OnSale(c *gin.Context) {
	plans, err := h.meter.ListPlans(c.Request.Context(), true)
	if err != nil {
		h.fail(c, err)
		return
	}
	packs, err := h.meter.ListPackDefs(c.Request.Context(), true)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, gin.H{"plans": plans, "packs": packs})
}

type idRequest struct {
	ID uint `json:"id" binding:"required"`
}

// Subscribe 用余额开通或升级套餐：POST /billing/subscribe {id}
func (h *BillingHandler) Subscribe(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择套餐")
		return
	}
	out, err := h.meter.Subscribe(c.Request.Context(), operator(c), id, req.ID, billing.SubscribeOptions{})
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// AutoRenew 设置自动续费：POST /billing/auto-renew {autoRenew}
func (h *BillingHandler) AutoRenew(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	var req struct {
		AutoRenew bool `json:"autoRenew"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.meter.SetAutoRenew(c.Request.Context(), id, req.AutoRenew); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, nil)
}

// BuyPack 用余额购买资源包：POST /billing/packs/buy {id}
func (h *BillingHandler) BuyPack(c *gin.Context) {
	id, ok := h.company(c)
	if !ok {
		return
	}
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择资源包")
		return
	}
	out, err := h.meter.BuyPack(c.Request.Context(), operator(c), id, req.ID, false)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// ---------- 平台 ----------

// GetSettings 计费设置：GET /billing/admin/settings
func (h *BillingHandler) GetSettings(c *gin.Context) {
	if h.platformOnly(c) {
		s := h.meter.GetSettings(c.Request.Context())
		utils.Success(c, gin.H{"settings": s, "pendingCount": h.meter.PendingRequestCount(c.Request.Context())})
	}
}

// UpdateSettings 修改扣费模式与复核金额：PUT /billing/admin/settings
func (h *BillingHandler) UpdateSettings(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req billing.Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.meter.UpdateSettings(c.Request.Context(), req.ChargeMode, req.ChargeFrom, req.ReviewThreshold); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, h.meter.GetSettings(c.Request.Context()))
}

// ListPrices 单价：GET /billing/admin/prices
func (h *BillingHandler) ListPrices(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.ListPrices(c.Request.Context())
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// CreatePrice 新增单价版本：POST /billing/admin/prices
func (h *BillingHandler) CreatePrice(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var p entity.BillPrice
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	if err := h.meter.CreatePrice(c.Request.Context(), operator(c), &p); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, p)
}

// DeletePrice 删除未生效的单价：DELETE /billing/admin/prices/:id
func (h *BillingHandler) DeletePrice(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	if err := h.meter.DeletePrice(c.Request.Context(), paramUint(c, "id")); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, nil)
}

// ListPlans 全部套餐：GET /billing/admin/plans
func (h *BillingHandler) ListPlans(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.ListPlans(c.Request.Context(), false)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// SavePlan 新增或修改套餐：POST /billing/admin/plans、PUT /billing/admin/plans/:id
func (h *BillingHandler) SavePlan(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var p entity.BillPlan
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	p.ID = paramUint(c, "id")
	if err := h.meter.SavePlan(c.Request.Context(), operator(c), &p); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, p)
}

// ListPackDefs 全部资源包：GET /billing/admin/pack-defs
func (h *BillingHandler) ListPackDefs(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.ListPackDefs(c.Request.Context(), false)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// SavePackDef 新增或修改资源包：POST /billing/admin/pack-defs、PUT /billing/admin/pack-defs/:id
func (h *BillingHandler) SavePackDef(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var p entity.BillPackDef
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}
	p.ID = paramUint(c, "id")
	if err := h.meter.SavePackDef(c.Request.Context(), operator(c), &p); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, p)
}

// Accounts 全部公司账户：GET /billing/admin/accounts
func (h *BillingHandler) Accounts(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.AccountList(c.Request.Context())
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// AccountOverview 某公司的费用概览：GET /billing/admin/accounts/:companyId
func (h *BillingHandler) AccountOverview(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.Overview(c.Request.Context(), paramUint(c, "companyId"))
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// AccountTransactions 某公司的余额流水：GET /billing/admin/accounts/:companyId/transactions
func (h *BillingHandler) AccountTransactions(c *gin.Context) {
	if h.platformOnly(c) {
		h.transactions(c, paramUint(c, "companyId"))
	}
}

// AccountExport 导出某公司的月账单：GET /billing/admin/accounts/:companyId/statements/:month/export
func (h *BillingHandler) AccountExport(c *gin.Context) {
	if h.platformOnly(c) {
		h.exportStatement(c, paramUint(c, "companyId"))
	}
}

type moneyRequest struct {
	Amount  int64  `json:"amount"` // 厘
	Voucher string `json:"voucher"`
	Reason  string `json:"reason"`
}

// Recharge 录入线下充值：POST /billing/admin/accounts/:companyId/recharge
func (h *BillingHandler) Recharge(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req moneyRequest
	_ = c.ShouldBindJSON(&req)
	t, err := h.meter.Recharge(c.Request.Context(), operator(c), paramUint(c, "companyId"), req.Amount, req.Voucher, req.Reason)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, t)
}

func (h *BillingHandler) giftOrAdjust(c *gin.Context, typ string) {
	if !h.platformOnly(c) {
		return
	}
	var req moneyRequest
	_ = c.ShouldBindJSON(&req)
	t, pending, err := h.meter.GiftOrAdjust(c.Request.Context(), operator(c), paramUint(c, "companyId"), typ, req.Amount, req.Reason)
	switch {
	case errors.Is(err, billing.ErrNeedReview):
		utils.Success(c, gin.H{"pending": pending})
	case err != nil:
		h.fail(c, err)
	default:
		utils.Success(c, gin.H{"transaction": t})
	}
}

// Gift 赠送：POST /billing/admin/accounts/:companyId/gift
func (h *BillingHandler) Gift(c *gin.Context) { h.giftOrAdjust(c, entity.TxnGift) }

// Adjust 调账：POST /billing/admin/accounts/:companyId/adjust
func (h *BillingHandler) Adjust(c *gin.Context) { h.giftOrAdjust(c, entity.TxnAdjust) }

// SetCredit 设置授信额度：PUT /billing/admin/accounts/:companyId/credit {amount}
func (h *BillingHandler) SetCredit(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req moneyRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.meter.SetCredit(c.Request.Context(), paramUint(c, "companyId"), req.Amount); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, nil)
}

// AdminSubscribe 为公司开通或升级套餐：POST /billing/admin/accounts/:companyId/subscribe {id, free}
func (h *BillingHandler) AdminSubscribe(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req struct {
		ID   uint `json:"id" binding:"required"`
		Free bool `json:"free"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择套餐")
		return
	}
	out, err := h.meter.Subscribe(c.Request.Context(), operator(c), paramUint(c, "companyId"), req.ID, billing.SubscribeOptions{ByPlatform: true, Free: req.Free})
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// GrantPack 发放资源包（免费）：POST /billing/admin/accounts/:companyId/packs {id}
func (h *BillingHandler) GrantPack(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择资源包")
		return
	}
	out, err := h.meter.BuyPack(c.Request.Context(), operator(c), paramUint(c, "companyId"), req.ID, true)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// Requests 赠送 / 调账申请：GET /billing/admin/requests?status=
func (h *BillingHandler) Requests(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	out, err := h.meter.AdjustRequests(c.Request.Context(), c.Query("status"))
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// ReviewRequest 复核申请：POST /billing/admin/requests/:id/review {approve, remark}
func (h *BillingHandler) ReviewRequest(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req struct {
		Approve bool   `json:"approve"`
		Remark  string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&req)
	out, err := h.meter.ReviewRequest(c.Request.Context(), operator(c), paramUint(c, "id"), req.Approve, req.Remark)
	if err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, out)
}

// GenerateStatements 为某月出账（已出账的月份重新生成）：POST /billing/admin/statements {month}
func (h *BillingHandler) GenerateStatements(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req struct {
		Month string `json:"month"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.meter.GenerateStatements(c.Request.Context(), req.Month); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, nil)
}

// Rerate 重新计价一段日期（不重新计量）：POST /billing/admin/rerate {from, to}
func (h *BillingHandler) Rerate(c *gin.Context) {
	if !h.platformOnly(c) {
		return
	}
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.meter.Rerate(c.Request.Context(), req.From, req.To); err != nil {
		h.fail(c, err)
		return
	}
	utils.Success(c, nil)
}
