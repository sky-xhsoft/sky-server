package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var meterOrder = []string{entity.MeterLiveDuration, entity.MeterPlayFlux, entity.MeterStorage}

func daysInMonth(day string) int {
	t, _ := time.ParseInLocation(dayLayout, day, time.Local)
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// chargeAmount 计费量 × 单价（0.0001 元）→ 厘；存储单价为 GB·月，按当月天数折算到一天
func chargeAmount(meter, day string, billable float64, unitPrice int64) int64 {
	v := billable * float64(unitPrice) / 10
	if meter == entity.MeterStorage {
		v /= float64(daysInMonth(day))
	}
	return int64(math.Round(v))
}

// RateDay 计算某天全部公司的费用。重复执行时只处理差额：先退回本行上次占用的套餐额度和资源包，再重新抵扣
func (m *Meter) RateDay(ctx context.Context, day string) error {
	if _, _, err := dayRange(day); err != nil {
		return err
	}
	var companies []uint
	if err := m.db.WithContext(ctx).Raw(
		"SELECT DISTINCT SYS_COMPANY_ID FROM bill_usage_daily WHERE USAGE_DATE = ? UNION SELECT DISTINCT SYS_COMPANY_ID FROM bill_charge WHERE USAGE_DATE = ?",
		day, day).Scan(&companies).Error; err != nil {
		return err
	}
	set := m.GetSettings(ctx)
	charging := set.ChargeMode == "charge" && set.ChargeFrom != "" && day >= set.ChargeFrom
	var errs []error
	for _, c := range companies {
		if c == 0 {
			continue
		}
		if err := m.rateCompanyDay(ctx, c, day, charging); err != nil {
			zap.L().Error("计费：计价失败", zap.Uint("company", c), zap.String("day", day), zap.Error(err))
			errs = append(errs, fmt.Errorf("公司 %d：%w", c, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Meter) rateCompanyDay(ctx context.Context, companyID uint, day string, charging bool) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先锁账户，保证同一公司的计价、扣费、购买串行执行
		if _, err := lockAccount(tx, companyID); err != nil {
			return err
		}
		type row struct {
			Meter string
			Qty   float64
		}
		var rows []row
		if err := tx.Model(&entity.BillUsageDaily{}).Select("METER AS meter, SUM(QUANTITY) AS qty").
			Where("SYS_COMPANY_ID = ? AND USAGE_DATE = ?", companyID, day).Group("METER").Scan(&rows).Error; err != nil {
			return err
		}
		usage := map[string]float64{}
		for _, r := range rows {
			usage[r.Meter] = r.Qty
		}
		sub, err := activeSub(tx, companyID, day)
		if err != nil {
			return err
		}
		if sub != nil {
			if sub, err = lockSub(tx, sub.ID); err != nil {
				return err
			}
		}
		for _, meter := range meterOrder {
			if err := m.rateMeter(ctx, tx, companyID, day, meter, round6(usage[meter]), sub, charging); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Meter) rateMeter(ctx context.Context, tx *gorm.DB, companyID uint, day, meter string, usage float64, sub *entity.BillSubscription, charging bool) error {
	var old entity.BillCharge
	exists := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("SYS_COMPANY_ID = ? AND USAGE_DATE = ? AND METER = ?", companyID, day, meter).First(&old).Error == nil
	if !exists && usage == 0 {
		return nil
	}
	// 退回上次占用的额度
	if exists {
		if err := releaseOffsets(tx, &old, sub); err != nil {
			return err
		}
	}
	c := entity.BillCharge{ID: old.ID, SysCompanyID: companyID, UsageDate: day, Meter: meter, Usage: usage, Charged: old.Charged}
	rest := usage
	if sub != nil {
		c.SubscriptionID = sub.ID
		switch meter {
		case entity.MeterStorage:
			// 存储额度是日均水平，不累计消耗
			c.PlanOffset = math.Min(rest, sub.QuotaStorageGB)
		case entity.MeterLiveDuration:
			c.PlanOffset = math.Min(rest, math.Max(0, sub.QuotaMinutes-sub.UsedMinutes))
			sub.UsedMinutes = round6(sub.UsedMinutes + c.PlanOffset)
		case entity.MeterPlayFlux:
			c.PlanOffset = math.Min(rest, math.Max(0, sub.QuotaFluxGB-sub.UsedFluxGB))
			sub.UsedFluxGB = round6(sub.UsedFluxGB + c.PlanOffset)
		}
		c.PlanOffset = round6(c.PlanOffset)
		rest = round6(rest - c.PlanOffset)
		if err := tx.Model(&entity.BillSubscription{}).Where("ID = ?", sub.ID).
			Updates(map[string]any{"USED_MINUTES": sub.UsedMinutes, "USED_FLUX_GB": sub.UsedFluxGB}).Error; err != nil {
			return err
		}
	}
	if rest > 0 && meter != entity.MeterStorage {
		var packs []entity.BillResourcePack
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("SYS_COMPANY_ID = ? AND METER = ? AND VALID_FROM <= ? AND VALID_TO >= ? AND REMAINING > 0", companyID, meter, day, day).
			Order("VALID_TO, ID").Find(&packs).Error; err != nil {
			return err
		}
		detail := map[string]float64{}
		for _, p := range packs {
			if rest <= 0 {
				break
			}
			take := round6(math.Min(rest, p.Remaining))
			if err := tx.Model(&entity.BillResourcePack{}).Where("ID = ?", p.ID).Update("REMAINING", round6(p.Remaining-take)).Error; err != nil {
				return err
			}
			detail[fmt.Sprint(p.ID)] = take
			c.PackOffset = round6(c.PackOffset + take)
			rest = round6(rest - take)
		}
		if len(detail) > 0 {
			b, _ := json.Marshal(detail)
			c.PackDetail = string(b)
		}
	}
	c.Billable = math.Max(0, rest)
	c.UnitPrice = priceIn(tx, companyID, meter, day)
	c.Amount = chargeAmount(meter, day, c.Billable, c.UnitPrice)
	// 已经扣过费的日期继续按差额调整（即使之后切回试算）
	c.Mode = "trial"
	delta := int64(0)
	if charging || old.Charged != 0 {
		c.Mode = "charge"
		delta = c.Amount - old.Charged
		c.Charged = c.Amount
	}
	if err := tx.Save(&c).Error; err != nil {
		return err
	}
	if delta == 0 {
		return nil
	}
	typ, remark := entity.TxnUsage, fmt.Sprintf("%s 用量扣费", day)
	if old.Mode == "charge" {
		typ = entity.TxnUsageAdjust
		if delta > 0 {
			remark = fmt.Sprintf("%s 重新计量后补扣", day)
		} else {
			remark = fmt.Sprintf("%s 重新计量后退还", day)
		}
	}
	return m.postTxn(ctx, tx, &entity.BillTransaction{SysCompanyID: companyID, Type: typ, Meter: meter, Amount: -delta,
		UsageDate: day, RefType: "charge", RefID: c.ID, Remark: remark, Operator: "日结"}, day[:7])
}

// releaseOffsets 退回费用明细上次占用的套餐额度与资源包
func releaseOffsets(tx *gorm.DB, old *entity.BillCharge, cur *entity.BillSubscription) error {
	if old.PlanOffset > 0 && old.SubscriptionID != 0 && old.Meter != entity.MeterStorage {
		col := "USED_MINUTES"
		if old.Meter == entity.MeterPlayFlux {
			col = "USED_FLUX_GB"
		}
		if cur != nil && cur.ID == old.SubscriptionID {
			if old.Meter == entity.MeterPlayFlux {
				cur.UsedFluxGB = round6(math.Max(0, cur.UsedFluxGB-old.PlanOffset))
			} else {
				cur.UsedMinutes = round6(math.Max(0, cur.UsedMinutes-old.PlanOffset))
			}
		} else if err := tx.Model(&entity.BillSubscription{}).Where("ID = ?", old.SubscriptionID).
			Update(col, gorm.Expr("CASE WHEN "+col+" > ? THEN "+col+" - ? ELSE 0 END", old.PlanOffset, old.PlanOffset)).Error; err != nil {
			return err
		}
	}
	if old.PackDetail != "" {
		detail := map[string]float64{}
		_ = json.Unmarshal([]byte(old.PackDetail), &detail)
		ids := make([]string, 0, len(detail))
		for id := range detail {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			if err := tx.Model(&entity.BillResourcePack{}).Where("ID = ?", id).
				Update("REMAINING", gorm.Expr("REMAINING + ?", detail[id])).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// RateRange 重新计价一段日期（平台重新计量后调用）
func (m *Meter) RateRange(ctx context.Context, from, to string) error {
	d0, _, err := dayRange(from)
	if err != nil {
		return err
	}
	d1, _, err := dayRange(to)
	if err != nil {
		return err
	}
	var errs []error
	for d := d0; !d.After(d1); d = d.AddDate(0, 0, 1) {
		errs = append(errs, m.RateDay(ctx, d.Format(dayLayout)))
	}
	return errors.Join(errs...)
}

// maxRerateDays 平台一次重新计价的最多天数
const maxRerateDays = 92

// Rerate 平台重新计价一段日期（不重新计量）：调整单价、套餐后用于试算或按差额补扣、退还。
// 从开始日期起把之后已计价的日期按顺序重算，让套餐额度和资源包重新分配
func (m *Meter) Rerate(ctx context.Context, from, to string) error {
	d0, _, err := dayRange(from)
	if err != nil {
		return err
	}
	d1, _, err := dayRange(to)
	if err != nil {
		return err
	}
	yesterday := m.now().AddDate(0, 0, -1).Format(dayLayout)
	switch {
	case d1.Before(d0):
		return errors.New("结束日期不能早于开始日期")
	case to > yesterday:
		return errors.New("只能重新计价昨天及以前的日期")
	case d1.Sub(d0) > maxRerateDays*24*time.Hour:
		return fmt.Errorf("一次最多重新计价 %d 天", maxRerateDays)
	}
	end := to
	if rated := m.getState(ctx, ratedDoneKey); rated > end {
		end = rated
	}
	if err := m.RateRange(ctx, from, end); err != nil {
		return err
	}
	if end > m.getState(ctx, ratedDoneKey) {
		return m.setState(ctx, ratedDoneKey, end)
	}
	return nil
}
