package billing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// periodEnd 从 start 起 months 个月的最后一天（含）：10-15 起一个月到 11-14；
// 1-31 起一个月到 2 月最后一天（下月没有对应日期时取月末）
func periodEnd(start time.Time, months int) time.Time {
	y, mo, d := start.Date()
	first := time.Date(y, mo+time.Month(months), 1, 0, 0, 0, 0, start.Location())
	last := first.AddDate(0, 1, -1).Day()
	if d > last {
		return time.Date(first.Year(), first.Month(), last, 0, 0, 0, 0, start.Location())
	}
	return time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, start.Location()).AddDate(0, 0, -1)
}

func daysBetween(from, to string) int {
	a, _ := time.ParseInLocation(dayLayout, from, time.Local)
	b, _ := time.ParseInLocation(dayLayout, to, time.Local)
	return int(math.Round(b.Sub(a).Hours()/24)) + 1
}

// activeSub 某天所在的套餐周期。不看状态：到期处理会先把周期标为已结束，之后计价最后一天时仍要用它的额度
func activeSub(tx *gorm.DB, companyID uint, day string) (*entity.BillSubscription, error) {
	var s entity.BillSubscription
	err := tx.Where("SYS_COMPANY_ID = ? AND START_DATE <= ? AND END_DATE >= ?", companyID, day, day).
		Order("START_DATE DESC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

// checkAffordable 公司自助购买：余额加授信须足够
func checkAffordable(acc *entity.BillAccount, fee int64) error {
	if acc.Balance+acc.CreditLimit < fee {
		return fmt.Errorf("余额不足：需要 %s 元，可用 %s 元", yuan(fee), yuan(acc.Balance+acc.CreditLimit))
	}
	return nil
}

func yuan(li int64) string { return fmt.Sprintf("%.2f", float64(li)/1000) }

// SubscribeOptions 开通方式
type SubscribeOptions struct {
	ByPlatform bool // 平台开通：可开通下架套餐，不检查余额
	Free       bool // 平台免费开通：不扣套餐费
}

// Subscribe 开通或升级套餐。当前没有有效周期时开通（今天起一个月）；有周期时只能升级到月费更高的套餐，
// 按剩余天数补差价、按比例补发额度（存储额度直接按新套餐）
func (m *Meter) Subscribe(ctx context.Context, op Operator, companyID, planID uint, opt SubscribeOptions) (*entity.BillSubscription, error) {
	var out *entity.BillSubscription
	today := m.now().Format(dayLayout)
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var plan entity.BillPlan
		if err := tx.First(&plan, planID).Error; err != nil {
			return errors.New("套餐不存在")
		}
		if plan.Status != "on" && !opt.ByPlatform {
			return errors.New("套餐已下架")
		}
		acc, err := lockAccount(tx, companyID)
		if err != nil {
			return err
		}
		cur, err := activeSub(tx, companyID, today)
		if err != nil {
			return err
		}
		fee := plan.MonthlyFee
		if cur == nil {
			start, _ := time.ParseInLocation(dayLayout, today, time.Local)
			out = &entity.BillSubscription{SysCompanyID: companyID, PlanID: plan.ID, PlanName: plan.Name, StartDate: today,
				EndDate: periodEnd(start, 1).Format(dayLayout), QuotaMinutes: plan.QuotaMinutes, QuotaFluxGB: plan.QuotaFluxGB,
				QuotaStorageGB: plan.QuotaStorageGB, AutoRenew: true, Status: "active", CreateBy: op.Name}
		} else {
			if cur.PlanID == plan.ID {
				return errors.New("已开通该套餐")
			}
			var old entity.BillPlan
			tx.First(&old, cur.PlanID)
			if plan.MonthlyFee <= old.MonthlyFee {
				return errors.New("周期内只能升级到月费更高的套餐；如需降级，请关闭自动续费，到期后再开通")
			}
			ratio := float64(daysBetween(today, cur.EndDate)) / float64(daysBetween(cur.StartDate, cur.EndDate))
			fee = int64(math.Round(float64(plan.MonthlyFee-old.MonthlyFee) * ratio))
			cur.QuotaMinutes += math.Max(0, plan.QuotaMinutes-old.QuotaMinutes) * ratio
			cur.QuotaFluxGB += math.Max(0, plan.QuotaFluxGB-old.QuotaFluxGB) * ratio
			cur.QuotaStorageGB = plan.QuotaStorageGB
			cur.PlanID, cur.PlanName = plan.ID, plan.Name
			out = cur
		}
		if opt.Free {
			fee = 0
		} else if !opt.ByPlatform {
			if err := checkAffordable(acc, fee); err != nil {
				return err
			}
		}
		out.Fee += fee
		if err := tx.Save(out).Error; err != nil {
			return err
		}
		remark := fmt.Sprintf("开通「%s」%s 至 %s", plan.Name, out.StartDate, out.EndDate)
		if cur != nil {
			remark = fmt.Sprintf("升级为「%s」，补 %s 至 %s 的差价", plan.Name, today, out.EndDate)
		}
		return m.postTxn(ctx, tx, &entity.BillTransaction{SysCompanyID: companyID, Type: entity.TxnPlanFee, Amount: -fee,
			RefType: "subscription", RefID: out.ID, Remark: remark, Operator: op.Name}, "")
	})
	return out, err
}

// SetAutoRenew 设置当前套餐是否自动续费
func (m *Meter) SetAutoRenew(ctx context.Context, companyID uint, auto bool) error {
	cur, err := activeSub(m.db.WithContext(ctx), companyID, m.now().Format(dayLayout))
	if err != nil {
		return err
	}
	if cur == nil {
		return errors.New("当前没有有效的套餐")
	}
	return m.db.WithContext(ctx).Model(cur).Update("AUTO_RENEW", auto).Error
}

// RenewDue 结束已到期的套餐周期；昨天到期且开启自动续费的，余额（含授信）足够时从今天起续一个周期
func (m *Meter) RenewDue(ctx context.Context) error {
	today := m.now().Format(dayLayout)
	yesterday := m.now().AddDate(0, 0, -1).Format(dayLayout)
	var due []entity.BillSubscription
	if err := m.db.WithContext(ctx).Where("STATUS = 'active' AND END_DATE < ?", today).Find(&due).Error; err != nil {
		return err
	}
	for _, s := range due {
		err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&entity.BillSubscription{}).Where("ID = ? AND STATUS = 'active'", s.ID).Update("STATUS", "ended").Error; err != nil {
				return err
			}
			if !s.AutoRenew || s.EndDate != yesterday {
				return nil
			}
			var plan entity.BillPlan
			if err := tx.First(&plan, s.PlanID).Error; err != nil || plan.Status != "on" {
				zap.L().Info("计费：套餐已下架，不再自动续费", zap.Uint("company", s.SysCompanyID), zap.Uint("plan", s.PlanID))
				return nil
			}
			acc, err := lockAccount(tx, s.SysCompanyID)
			if err != nil {
				return err
			}
			if checkAffordable(acc, plan.MonthlyFee) != nil {
				zap.L().Info("计费：余额不足，未自动续费", zap.Uint("company", s.SysCompanyID), zap.String("plan", plan.Name))
				return nil
			}
			start, _ := time.ParseInLocation(dayLayout, today, time.Local)
			next := &entity.BillSubscription{SysCompanyID: s.SysCompanyID, PlanID: plan.ID, PlanName: plan.Name, StartDate: today,
				EndDate: periodEnd(start, 1).Format(dayLayout), QuotaMinutes: plan.QuotaMinutes, QuotaFluxGB: plan.QuotaFluxGB,
				QuotaStorageGB: plan.QuotaStorageGB, AutoRenew: true, Status: "active", Fee: plan.MonthlyFee, CreateBy: "自动续费"}
			if err := tx.Create(next).Error; err != nil {
				return err
			}
			return m.postTxn(ctx, tx, &entity.BillTransaction{SysCompanyID: s.SysCompanyID, Type: entity.TxnPlanFee, Amount: -plan.MonthlyFee,
				RefType: "subscription", RefID: next.ID, Remark: fmt.Sprintf("自动续费「%s」%s 至 %s", plan.Name, next.StartDate, next.EndDate), Operator: "自动续费"}, "")
		})
		if err != nil {
			zap.L().Error("计费：处理到期套餐失败", zap.Uint("subscription", s.ID), zap.Error(err))
		}
	}
	return nil
}

// BuyPack 购买（公司用余额）或发放（平台，免费）资源包，从今天起生效
func (m *Meter) BuyPack(ctx context.Context, op Operator, companyID, defID uint, grant bool) (*entity.BillResourcePack, error) {
	var out *entity.BillResourcePack
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var def entity.BillPackDef
		if err := tx.First(&def, defID).Error; err != nil {
			return errors.New("资源包不存在")
		}
		if def.Status != "on" && !grant {
			return errors.New("资源包已下架")
		}
		acc, err := lockAccount(tx, companyID)
		if err != nil {
			return err
		}
		price := def.Price
		if grant {
			price = 0
		} else if err := checkAffordable(acc, price); err != nil {
			return err
		}
		start := m.now()
		source := "purchase"
		if grant {
			source = "grant"
		}
		out = &entity.BillResourcePack{SysCompanyID: companyID, PackDefID: def.ID, Name: def.Name, Meter: def.Meter, Total: def.Quantity,
			Remaining: def.Quantity, ValidFrom: start.Format(dayLayout), ValidTo: periodEnd(start, def.ValidMonths).Format(dayLayout),
			Source: source, Price: price, CreateBy: op.Name}
		if err := tx.Create(out).Error; err != nil {
			return err
		}
		return m.postTxn(ctx, tx, &entity.BillTransaction{SysCompanyID: companyID, Type: entity.TxnPackFee, Amount: -price,
			RefType: "resource_pack", RefID: out.ID, Remark: fmt.Sprintf("购买「%s」，有效期至 %s", def.Name, out.ValidTo), Operator: op.Name}, "")
	})
	return out, err
}

// lockSub 锁定套餐周期行
func lockSub(tx *gorm.DB, id uint) (*entity.BillSubscription, error) {
	var s entity.BillSubscription
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
