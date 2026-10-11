package billing

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var admin = Operator{UserID: 1, Name: "admin"}

// 单价：时长 0.05 元 / 分钟，流量 0.25 元 / GB，存储 0.31 元 / GB·月（10 月 31 天，每 GB 每天 0.01 元）
func setupPricing(t *testing.T, m *Meter, db *gorm.DB) {
	t.Helper()
	for _, p := range []entity.BillPrice{
		{Meter: entity.MeterLiveDuration, UnitPrice: 500, EffectiveDate: "2026-01-01"},
		{Meter: entity.MeterPlayFlux, UnitPrice: 2500, EffectiveDate: "2026-01-01"},
		{Meter: entity.MeterStorage, UnitPrice: 3100, EffectiveDate: "2026-01-01"},
	} {
		db.Create(&p)
	}
}

func addUsage(t *testing.T, db *gorm.DB, company uint, day, meter string, qty float64) {
	t.Helper()
	db.Where("SYS_COMPANY_ID = ? AND USAGE_DATE = ? AND METER = ?", company, day, meter).Delete(&entity.BillUsageDaily{})
	if qty > 0 {
		db.Create(&entity.BillUsageDaily{SysCompanyID: company, UsageDate: day, Meter: meter, Quantity: qty, Unit: "x", SourceKey: meter + day + string(rune(company))})
	}
}

func charge(db *gorm.DB, company uint, day, meter string) entity.BillCharge {
	var c entity.BillCharge
	db.Where("SYS_COMPANY_ID = ? AND USAGE_DATE = ? AND METER = ?", company, day, meter).First(&c)
	return c
}

func balance(m *Meter, company uint) entity.BillAccount { return m.account(ctx, company) }

func txnSum(db *gorm.DB, company uint) int64 {
	var s int64
	db.Model(&entity.BillTransaction{}).Where("SYS_COMPANY_ID = ?", company).Select("COALESCE(SUM(AMOUNT),0)").Scan(&s)
	return s
}

func TestPriceLookup(t *testing.T) {
	m, _, db := setup(t)
	setupPricing(t, m, db)
	db.Create(&entity.BillPrice{Meter: entity.MeterLiveDuration, UnitPrice: 400, EffectiveDate: "2026-10-10"})
	db.Create(&entity.BillPrice{SysCompanyID: 50, Meter: entity.MeterLiveDuration, UnitPrice: 300, EffectiveDate: "2026-10-05"})
	cases := []struct {
		company uint
		day     string
		want    int64
	}{{51, "2026-10-09", 500}, {51, "2026-10-10", 400}, {50, "2026-10-04", 500}, {50, "2026-10-05", 300}, {50, "2026-10-20", 300}, {51, "2025-12-31", 0}}
	for _, c := range cases {
		if got := m.priceOf(ctx, c.company, entity.MeterLiveDuration, c.day); got != c.want {
			t.Fatalf("公司 %d %s：%d，期望 %d", c.company, c.day, got, c.want)
		}
	}
	m.now = func() time.Time { return at("2026-10-11 10:00:00") }
	// 试算期间（没有扣过费）可以设置过去的日期
	past := &entity.BillPrice{Meter: entity.MeterStorage, UnitPrice: 1, EffectiveDate: "2026-09-01"}
	if err := m.CreatePrice(ctx, admin, past); err != nil {
		t.Fatal("试算期间可设置过去的生效日期", err)
	}
	// 扣过费后，生效日期须晚于已扣费的最后一天
	db.Create(&entity.BillCharge{SysCompanyID: 50, UsageDate: "2026-10-09", Meter: entity.MeterLiveDuration, Mode: "charge"})
	if err := m.CreatePrice(ctx, admin, &entity.BillPrice{Meter: entity.MeterLiveDuration, UnitPrice: 1, EffectiveDate: "2026-10-09"}); err == nil {
		t.Fatal("已扣费的日期不能改价")
	}
	if err := m.DeletePrice(ctx, past.ID); err == nil {
		t.Fatal("已用于扣费的单价不能删除")
	}
	if err := m.CreatePrice(ctx, admin, &entity.BillPrice{Meter: "x", UnitPrice: 1, EffectiveDate: "2026-10-12"}); err == nil {
		t.Fatal("不支持的计费点")
	}
	p := &entity.BillPrice{SysCompanyID: 51, Meter: entity.MeterPlayFlux, UnitPrice: 2000, EffectiveDate: "2026-10-12"}
	if err := m.CreatePrice(ctx, admin, p); err != nil {
		t.Fatal(err)
	}
	if err := m.DeletePrice(ctx, p.ID); err != nil {
		t.Fatal("未生效的单价可以删除", err)
	}
	if err := m.CreatePrice(ctx, admin, &entity.BillPrice{Meter: entity.MeterLiveDuration, UnitPrice: 1, EffectiveDate: "2026-10-10"}); err != nil {
		t.Fatal("已扣费最后一天之后的日期可以设置", err)
	}
}

func TestRateTrialAndCharge(t *testing.T) {
	m, _, db := setup(t)
	setupPricing(t, m, db)
	addUsage(t, db, 50, "2026-10-09", entity.MeterLiveDuration, 100)
	addUsage(t, db, 50, "2026-10-09", entity.MeterPlayFlux, 4)
	addUsage(t, db, 50, "2026-10-09", entity.MeterStorage, 10)
	// 默认试算：只算金额，不动余额
	if err := m.RateDay(ctx, "2026-10-09"); err != nil {
		t.Fatal(err)
	}
	if c := charge(db, 50, "2026-10-09", entity.MeterLiveDuration); c.Amount != 5000 || c.Mode != "trial" || c.Charged != 0 {
		t.Fatalf("试算时长：%+v", c)
	}
	if c := charge(db, 50, "2026-10-09", entity.MeterPlayFlux); c.Amount != 1000 {
		t.Fatalf("流量 4 GB × 0.25 = 1 元：%+v", c)
	}
	if c := charge(db, 50, "2026-10-09", entity.MeterStorage); c.Amount != 100 {
		t.Fatalf("存储 10 GB × 0.31 ÷ 31 = 0.1 元：%+v", c)
	}
	if balance(m, 50).Balance != 0 || txnSum(db, 50) != 0 {
		t.Fatal("试算不应扣费")
	}

	// 从 10-10 起扣费：10-09 仍是试算
	if err := m.UpdateSettings(ctx, "charge", "2026-10-10", defaultReviewThreshold); err != nil {
		t.Fatal(err)
	}
	m.RateDay(ctx, "2026-10-09")
	if balance(m, 50).Balance != 0 {
		t.Fatal("开始扣费日期之前的用量不补扣")
	}
	addUsage(t, db, 50, "2026-10-10", entity.MeterLiveDuration, 60)
	m.RateDay(ctx, "2026-10-10")
	if b := balance(m, 50).Balance; b != -3000 {
		t.Fatalf("扣 60 分钟 × 0.05 = 3 元，余额 %d", b)
	}
	// 重复计价不重复扣
	m.RateDay(ctx, "2026-10-10")
	if b := balance(m, 50).Balance; b != -3000 {
		t.Fatalf("重复计价不应重复扣费，余额 %d", b)
	}
	// 重新计量后用量变少：退还差额
	addUsage(t, db, 50, "2026-10-10", entity.MeterLiveDuration, 40)
	m.RateDay(ctx, "2026-10-10")
	if b := balance(m, 50).Balance; b != -2000 {
		t.Fatalf("退还 1 元后余额 %d", b)
	}
	var adj entity.BillTransaction
	db.Where("TYPE = ?", entity.TxnUsageAdjust).First(&adj)
	if adj.Amount != 1000 || adj.Period != "2026-10" || adj.UsageDate != "2026-10-10" {
		t.Fatalf("退还流水：%+v", adj)
	}
	// 切回试算后，已扣费的日期仍按差额调整
	m.UpdateSettings(ctx, "trial", "", defaultReviewThreshold)
	addUsage(t, db, 50, "2026-10-10", entity.MeterLiveDuration, 0)
	m.RateDay(ctx, "2026-10-10")
	if b := balance(m, 50).Balance; b != 0 || txnSum(db, 50) != 0 {
		t.Fatalf("用量清零后全部退还，余额 %d", b)
	}
}

func TestRatePlanAndPacks(t *testing.T) {
	m, _, db := setup(t)
	setupPricing(t, m, db)
	m.UpdateSettings(ctx, "charge", "2026-10-01", defaultReviewThreshold)
	m.now = func() time.Time { return at("2026-10-01 09:00:00") }
	plan := &entity.BillPlan{Name: "基础版", MonthlyFee: 299000, QuotaMinutes: 100, QuotaFluxGB: 10, QuotaStorageGB: 5, Status: "on"}
	m.SavePlan(ctx, admin, plan)
	pack := &entity.BillPackDef{Name: "时长包 50 分钟", Meter: entity.MeterLiveDuration, Quantity: 50, ValidMonths: 1, Price: 2000, Status: "on"}
	m.SavePackDef(ctx, admin, pack)
	m.Recharge(ctx, admin, 50, 1000000, "TX001", "")
	if _, err := m.Subscribe(ctx, Operator{UserID: 9, Name: "u"}, 50, plan.ID, SubscribeOptions{}); err != nil {
		t.Fatal(err)
	}
	p1, _ := m.BuyPack(ctx, Operator{UserID: 9, Name: "u"}, 50, pack.ID, false)
	p2, _ := m.BuyPack(ctx, admin, 50, pack.ID, true) // 平台发放，免费
	db.Model(p2).Update("VALID_TO", "2026-10-20")     // 先到期的先用
	if b := balance(m, 50).Balance; b != 1000000-299000-2000 {
		t.Fatalf("充值 1000 − 套餐 299 − 资源包 2 = 699，余额 %d", b)
	}

	// 10-01 用 80 分钟、8 GB、8 GB 存储：全部在套餐内，存储超出 3 GB
	addUsage(t, db, 50, "2026-10-01", entity.MeterLiveDuration, 80)
	addUsage(t, db, 50, "2026-10-01", entity.MeterPlayFlux, 8)
	addUsage(t, db, 50, "2026-10-01", entity.MeterStorage, 8)
	m.RateDay(ctx, "2026-10-01")
	if c := charge(db, 50, "2026-10-01", entity.MeterLiveDuration); c.PlanOffset != 80 || c.Amount != 0 {
		t.Fatalf("时长全部由套餐抵扣：%+v", c)
	}
	if c := charge(db, 50, "2026-10-01", entity.MeterStorage); c.PlanOffset != 5 || c.Billable != 3 || c.Amount != 30 {
		t.Fatalf("存储超出 3 GB 按天计费 0.03 元：%+v", c)
	}
	// 10-02 用 90 分钟：套餐剩 20，资源包（先到期的 p2）抵 50，p1 抵 20，计费 0
	addUsage(t, db, 50, "2026-10-02", entity.MeterLiveDuration, 90)
	m.RateDay(ctx, "2026-10-02")
	c := charge(db, 50, "2026-10-02", entity.MeterLiveDuration)
	if c.PlanOffset != 20 || c.PackOffset != 70 || c.Billable != 0 {
		t.Fatalf("套餐 20 + 资源包 70：%+v", c)
	}
	var r1, r2 entity.BillResourcePack
	db.First(&r1, p1.ID)
	db.First(&r2, p2.ID)
	if r2.Remaining != 0 || r1.Remaining != 30 {
		t.Fatalf("先用先到期的资源包：p2 %v p1 %v", r2.Remaining, r1.Remaining)
	}
	// 10-03 用 100 分钟：p1 剩 30，计费 70 分钟 = 3.5 元
	addUsage(t, db, 50, "2026-10-03", entity.MeterLiveDuration, 100)
	m.RateDay(ctx, "2026-10-03")
	if c := charge(db, 50, "2026-10-03", entity.MeterLiveDuration); c.PackOffset != 30 || c.Billable != 70 || c.Amount != 3500 {
		t.Fatalf("超出部分按量：%+v", c)
	}
	// 重新计价 10-02（用量改为 10 分钟）：退回额度后重新抵扣，不重复占用
	addUsage(t, db, 50, "2026-10-02", entity.MeterLiveDuration, 10)
	m.RateDay(ctx, "2026-10-02")
	var sub entity.BillSubscription
	db.First(&sub)
	db.First(&r1, p1.ID)
	db.First(&r2, p2.ID)
	// 10-02 退回套餐 20、p2 50、p1 20，重新用套餐 10 分钟；10-03 已占用的 p1 30 不变
	if sub.UsedMinutes != 90 || r2.Remaining != 50 || r1.Remaining != 20 {
		t.Fatalf("重新计价后：套餐已用 %v，p2 剩 %v，p1 剩 %v", sub.UsedMinutes, r2.Remaining, r1.Remaining)
	}
	// 按顺序重算 10-02 之后的日期：10-02 退回的额度由 10-03 使用，10-03 原先扣的 3.5 元退还
	before := balance(m, 50).Balance
	if err := m.RateRange(ctx, "2026-10-02", "2026-10-03"); err != nil {
		t.Fatal(err)
	}
	if c := charge(db, 50, "2026-10-03", entity.MeterLiveDuration); c.PlanOffset != 10 || c.PackOffset != 90 || c.Billable != 0 {
		t.Fatalf("10-03 重新分配额度：%+v", c)
	}
	if b := balance(m, 50).Balance; b != before+3500 {
		t.Fatalf("退还 3.5 元：%d → %d", before, b)
	}
	if acc := balance(m, 50); acc.Balance != txnSum(db, 50) {
		t.Fatalf("余额 %d 应等于流水合计 %d", acc.Balance, txnSum(db, 50))
	}
}

func TestSubscribeUpgradeRenew(t *testing.T) {
	m, _, db := setup(t)
	m.now = func() time.Time { return at("2026-10-16 09:00:00") }
	basic := &entity.BillPlan{Name: "基础版", MonthlyFee: 310000, QuotaMinutes: 3100, QuotaFluxGB: 310, QuotaStorageGB: 100, Status: "on"}
	pro := &entity.BillPlan{Name: "专业版", MonthlyFee: 620000, QuotaMinutes: 6200, QuotaFluxGB: 620, QuotaStorageGB: 500, Status: "on"}
	m.SavePlan(ctx, admin, basic)
	m.SavePlan(ctx, admin, pro)
	user := Operator{UserID: 9, Name: "u"}
	if _, err := m.Subscribe(ctx, user, 50, basic.ID, SubscribeOptions{}); err == nil {
		t.Fatal("余额不足不能购买")
	}
	m.Recharge(ctx, admin, 50, 400000, "TX1", "")
	s, err := m.Subscribe(ctx, user, 50, basic.ID, SubscribeOptions{})
	if err != nil || s.StartDate != "2026-10-16" || s.EndDate != "2026-11-15" {
		t.Fatalf("周期从开通日起一个月：%+v %v", s, err)
	}
	if _, err := m.Subscribe(ctx, user, 50, basic.ID, SubscribeOptions{}); err == nil {
		t.Fatal("已开通同一套餐")
	}
	// 10-26 升级：周期 31 天剩 21 天，补差价 310 × 21/31 = 210 元；余额只剩 90，不够
	m.now = func() time.Time { return at("2026-10-26 09:00:00") }
	if _, err := m.Subscribe(ctx, user, 50, pro.ID, SubscribeOptions{}); err == nil {
		t.Fatal("余额不足不能升级")
	}
	m.Recharge(ctx, admin, 50, 200000, "TX2", "")
	s, err = m.Subscribe(ctx, user, 50, pro.ID, SubscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if s.PlanName != "专业版" || s.EndDate != "2026-11-15" || s.QuotaMinutes != 3100+3100.0*21/31 || s.QuotaStorageGB != 500 {
		t.Fatalf("升级：%+v", s)
	}
	if b := balance(m, 50).Balance; b != 600000-310000-210000 {
		t.Fatalf("升级补差价后余额 %d", b)
	}
	if _, err := m.Subscribe(ctx, user, 50, basic.ID, SubscribeOptions{}); err == nil {
		t.Fatal("周期内不能降级")
	}

	// 11-16：昨天到期，余额 80 不够续费 → 结束
	m.now = func() time.Time { return at("2026-11-16 01:00:00") }
	m.RenewDue(ctx)
	if cur, _ := activeSub(db, 50, "2026-11-16"); cur != nil {
		t.Fatalf("余额不足不续费：%+v", cur)
	}
	// 公司 51：平台免费开通，到期自动续费从余额扣
	m.now = func() time.Time { return at("2026-10-16 09:00:00") }
	if _, err := m.Subscribe(ctx, admin, 51, basic.ID, SubscribeOptions{ByPlatform: true, Free: true}); err != nil {
		t.Fatal(err)
	}
	m.Recharge(ctx, admin, 51, 500000, "TX3", "")
	m.now = func() time.Time { return at("2026-11-16 01:00:00") }
	m.RenewDue(ctx)
	next, _ := activeSub(db, 51, "2026-11-16")
	if next == nil || next.StartDate != "2026-11-16" || next.EndDate != "2026-12-15" || balance(m, 51).Balance != 190000 {
		t.Fatalf("自动续费：%+v，余额 %d", next, balance(m, 51).Balance)
	}
	// 到期当天（11-15）计价仍使用已结束周期的额度
	if prev, _ := activeSub(db, 51, "2026-11-15"); prev == nil || prev.Status != "ended" {
		t.Fatalf("历史日期仍能找到所在周期：%+v", prev)
	}
}

func TestGiftAndReview(t *testing.T) {
	m, _, db := setup(t)
	setupPricing(t, m, db)
	m.UpdateSettings(ctx, "charge", "2026-10-01", 500000) // 复核阈值 500 元
	m.Recharge(ctx, admin, 50, 100000, "TX1", "")
	if _, _, err := m.GiftOrAdjust(ctx, admin, 50, entity.TxnGift, 50000, "试用"); err != nil {
		t.Fatal(err)
	}
	addUsage(t, db, 50, "2026-10-05", entity.MeterLiveDuration, 1200) // 60 元
	m.RateDay(ctx, "2026-10-05")
	if acc := balance(m, 50); acc.Balance != 90000 || acc.GiftBalance != 0 {
		t.Fatalf("先扣赠送 50 元再扣现金 10 元：%+v", acc)
	}
	// 超过阈值：生成申请，申请人不能自己复核
	_, req, err := m.GiftOrAdjust(ctx, admin, 50, entity.TxnAdjust, -600000, "扣回")
	if !errors.Is(err, ErrNeedReview) || req.Status != "pending" {
		t.Fatalf("超过阈值应提交复核：%v %+v", err, req)
	}
	if balance(m, 50).Balance != 90000 {
		t.Fatal("复核前不入账")
	}
	if _, err := m.ReviewRequest(ctx, admin, req.ID, true, ""); err == nil {
		t.Fatal("申请人不能复核")
	}
	if r, err := m.ReviewRequest(ctx, Operator{UserID: 2, Name: "admin2"}, req.ID, true, "同意"); err != nil || r.Status != "approved" || r.TxnID == 0 {
		t.Fatalf("另一位管理员复核：%v %+v", err, r)
	}
	if b := balance(m, 50).Balance; b != -510000 {
		t.Fatalf("复核后入账，余额 %d", b)
	}
	if _, err := m.ReviewRequest(ctx, Operator{UserID: 3, Name: "x"}, req.ID, true, ""); err == nil {
		t.Fatal("已处理的申请不能重复复核")
	}
	for _, c := range []struct {
		typ string
		amt int64
	}{{entity.TxnGift, -1}, {entity.TxnAdjust, 0}, {"other", 1}} {
		if _, _, err := m.GiftOrAdjust(ctx, admin, 50, c.typ, c.amt, "x"); err == nil {
			t.Fatalf("应拒绝 %v", c)
		}
	}
	if _, err := m.Recharge(ctx, admin, 50, 100, "", ""); err == nil {
		t.Fatal("线下充值须填写凭证号")
	}
	if _, err := m.Recharge(ctx, admin, 99, 100, "x", ""); err == nil {
		t.Fatal("公司不存在")
	}
}

func TestStatementAndLateAdjust(t *testing.T) {
	m, _, db := setup(t)
	setupPricing(t, m, db)
	m.UpdateSettings(ctx, "charge", "2026-09-01", defaultReviewThreshold)
	m.now = func() time.Time { return at("2026-09-10 10:00:00") }
	m.Recharge(ctx, admin, 50, 100000, "TX1", "")
	addUsage(t, db, 50, "2026-09-30", entity.MeterLiveDuration, 200) // 10 元
	addUsage(t, db, 50, "2026-09-30", entity.MeterPlayFlux, 8)       // 2 元
	// 9-30 的用量在 10-01 计价，仍计入 9 月账期
	m.now = func() time.Time { return at("2026-10-01 04:30:00") }
	m.RateDay(ctx, "2026-09-30")
	m.Recharge(ctx, admin, 50, 50000, "TX2", "") // 10 月的充值
	if err := m.GenerateStatements(ctx, "2026-09"); err != nil {
		t.Fatal(err)
	}
	st, err := m.Statement(ctx, 50, "2026-09")
	if err != nil {
		t.Fatal(err)
	}
	if st.Opening != 0 || st.Recharge != 100000 || st.UsageDuration != 10000 || st.UsageFlux != 2000 || st.Closing != 88000 {
		t.Fatalf("9 月账单：%+v", st.BillStatement)
	}
	if len(st.Detail.Meters) != 3 || st.Detail.Meters[0].Usage != 200 {
		t.Fatalf("账单明细：%+v", st.Detail)
	}
	if m.getState(ctx, statementClosedKey) != "2026-09" {
		t.Fatal("出账后关闭 9 月")
	}
	// 9 月出账后重新计量 9-30：补扣计入 10 月
	addUsage(t, db, 50, "2026-09-30", entity.MeterLiveDuration, 300)
	m.RateDay(ctx, "2026-09-30")
	var adj entity.BillTransaction
	db.Where("TYPE = ?", entity.TxnUsageAdjust).First(&adj)
	if adj.Amount != -5000 || adj.Period != "2026-10" {
		t.Fatalf("补扣计入下个账期：%+v", adj)
	}
	m.now = func() time.Time { return at("2026-11-01 05:00:00") }
	if err := m.GenerateStatements(ctx, "2026-10"); err != nil {
		t.Fatal(err)
	}
	oct, _ := m.Statement(ctx, 50, "2026-10")
	if oct.Opening != 88000 || oct.Recharge != 50000 || oct.UsageAdjust != 5000 || oct.Closing != 133000 || oct.Closing != balance(m, 50).Balance {
		t.Fatalf("10 月账单：%+v", oct.BillStatement)
	}
	// 重复出账结果相同
	if err := m.GenerateStatements(ctx, "2026-10"); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&entity.BillStatement{}).Where("SYS_COMPANY_ID = 50").Count(&n)
	if n != 2 {
		t.Fatalf("重复出账不新增：%d", n)
	}
	// 余额与流水不一致时不出账
	db.Model(&entity.BillAccount{}).Where("SYS_COMPANY_ID = 50").Update("BALANCE", 1)
	if err := m.GenerateStatements(ctx, "2026-10"); err == nil {
		t.Fatal("余额与流水不一致时应报错")
	}
	if err := m.GenerateStatements(ctx, "2026-11"); err == nil {
		t.Fatal("未结束的月份不能出账")
	}
	// 导出 Excel
	db.Model(&entity.BillAccount{}).Where("SYS_COMPANY_ID = 50").Update("BALANCE", 133000)
	b, err := m.ExportStatement(ctx, 50, "2026-10")
	if err != nil || len(b) == 0 {
		t.Fatalf("导出：%v", err)
	}
	f, err := excelize.OpenReader(bytesReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if sheets := f.GetSheetList(); len(sheets) != 4 || sheets[0] != "账单汇总" {
		t.Fatalf("工作表：%v", sheets)
	}
	v, _ := f.GetCellValue("账单汇总", "B12")
	if v != "133" {
		t.Fatalf("期末余额单元格：%q", v)
	}
}

func TestTickRatesAndCloses(t *testing.T) {
	m, acc, db := setup(t)
	setupPricing(t, m, db)
	acc.apis[GlobalAccount] = &fakeLive{domains: []string{}}
	acc.apis["company:1"] = &fakeLive{domains: []string{}}
	m.UpdateSettings(ctx, "charge", "2026-09-01", defaultReviewThreshold)
	// 9-30 推流 60 分钟：每日计量算出时长后计价，扣 3 元
	end := at("2026-09-30 11:00:00")
	db.Create(&entity.BillLiveSession{SysCompanyID: 50, RoomID: 101, StreamName: "101", StartTime: at("2026-09-30 10:00:00"), EndTime: &end, LastSeenTime: end, Seconds: 3600, EndSource: "callback", PushEventID: 1})
	db.Create(&entity.BillMeterState{StateKey: ratedDoneKey, StateValue: "2026-09-29"})
	m.now = func() time.Time { return at("2026-10-01 04:10:00") }
	m.Tick(ctx, SchedulerConfig{DailyHour: 4})
	if m.getState(ctx, ratedDoneKey) != "2026-09-30" || balance(m, 50).Balance != -3000 {
		t.Fatalf("计量后计价昨天：%s 余额 %d", m.getState(ctx, ratedDoneKey), balance(m, 50).Balance)
	}
	if m.getState(ctx, statementClosedKey) != "2026-09" {
		t.Fatal("上月最后一天计价后出账")
	}
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }

func TestRerate(t *testing.T) {
	m, _, db := setup(t)
	m.now = func() time.Time { return at("2026-10-11 10:00:00") }
	addUsage(t, db, 50, "2026-10-05", entity.MeterLiveDuration, 100)
	addUsage(t, db, 50, "2026-10-06", entity.MeterLiveDuration, 100)
	// 试算期间补设默认单价后重新计价历史日期
	if err := m.CreatePrice(ctx, admin, &entity.BillPrice{Meter: entity.MeterLiveDuration, UnitPrice: 500, EffectiveDate: "2026-10-01"}); err != nil {
		t.Fatal(err)
	}
	if err := m.Rerate(ctx, "2026-10-01", "2026-10-10"); err != nil {
		t.Fatal(err)
	}
	if c := charge(db, 50, "2026-10-06", entity.MeterLiveDuration); c.Amount != 5000 || c.Mode != "trial" {
		t.Fatalf("试算按新单价：%+v", c)
	}
	if m.getState(ctx, ratedDoneKey) != "2026-10-10" {
		t.Fatal("重新计价后记录已计价到的日期")
	}
	for _, r := range [][2]string{{"2026-10-05", "2026-10-01"}, {"2026-10-01", "2026-10-11"}, {"2026-06-01", "2026-10-10"}} {
		if err := m.Rerate(ctx, r[0], r[1]); err == nil {
			t.Fatalf("应拒绝 %v", r)
		}
	}
}
