package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

var validMeters = map[string]bool{entity.MeterLiveDuration: true, entity.MeterPlayFlux: true, entity.MeterStorage: true}

// priceOf 某公司某计费点在某天适用的单价：先找该公司的专属价，再找默认价；都没有时为 0
func (m *Meter) priceOf(ctx context.Context, companyID uint, meter, day string) int64 {
	return priceIn(m.db.WithContext(ctx), companyID, meter, day)
}

// priceIn 在指定连接（事务）中查找单价
func priceIn(db *gorm.DB, companyID uint, meter, day string) int64 {
	for _, c := range []uint{companyID, 0} {
		var p entity.BillPrice
		err := db.Where("SYS_COMPANY_ID = ? AND METER = ? AND EFFECTIVE_DATE <= ?", c, meter, day).
			Order("EFFECTIVE_DATE DESC, ID DESC").First(&p).Error
		if err == nil {
			return p.UnitPrice
		}
	}
	return 0
}

// lastChargedDay 已实际扣费的最后一个用量日期（没有时为空）：单价只能改在这之后，避免改动已扣费的历史
func (m *Meter) lastChargedDay(ctx context.Context) string {
	var day string
	m.db.WithContext(ctx).Model(&entity.BillCharge{}).Where("MODE = 'charge'").Select("COALESCE(MAX(USAGE_DATE), '')").Scan(&day)
	return day
}

// ListPrices 全部单价（默认价与公司专属价），按计费点、公司、生效日期排列
func (m *Meter) ListPrices(ctx context.Context) ([]entity.BillPrice, error) {
	var out []entity.BillPrice
	err := m.db.WithContext(ctx).Order("METER, SYS_COMPANY_ID, EFFECTIVE_DATE DESC, ID DESC").Find(&out).Error
	return out, err
}

// CreatePrice 新增一个单价版本；公司为 0 时是默认价。生效日期须晚于已扣费的最后一天：
// 试算期间可设置过去的日期（上线前为历史用量试算），开始扣费后已扣费的日期不能改价
func (m *Meter) CreatePrice(ctx context.Context, op Operator, p *entity.BillPrice) error {
	if !validMeters[p.Meter] {
		return fmt.Errorf("不支持的计费点：%s", p.Meter)
	}
	if p.UnitPrice < 0 {
		return errors.New("单价不能为负")
	}
	if _, _, err := dayRange(p.EffectiveDate); err != nil {
		return err
	}
	if last := m.lastChargedDay(ctx); p.EffectiveDate <= last {
		return fmt.Errorf("生效日期须晚于已扣费的最后一天（%s），已扣费的用量不能改价", last)
	}
	if p.SysCompanyID != 0 {
		if err := m.ensureCompany(ctx, p.SysCompanyID); err != nil {
			return err
		}
	}
	p.ID, p.CreateBy = 0, op.Name
	return m.db.WithContext(ctx).Create(p).Error
}

// DeletePrice 删除尚未用于扣费的单价版本
func (m *Meter) DeletePrice(ctx context.Context, id uint) error {
	var p entity.BillPrice
	if err := m.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return errors.New("单价不存在")
	}
	if last := m.lastChargedDay(ctx); p.EffectiveDate <= last {
		return errors.New("该单价已用于扣费，不能删除；可新增一个版本替代")
	}
	return m.db.WithContext(ctx).Delete(&p).Error
}

// ListPlans 套餐；onSale 为 true 时只返回在售的
func (m *Meter) ListPlans(ctx context.Context, onSale bool) ([]entity.BillPlan, error) {
	var out []entity.BillPlan
	q := m.db.WithContext(ctx).Order("SORT_NO, MONTHLY_FEE, ID")
	if onSale {
		q = q.Where("STATUS = 'on'")
	}
	err := q.Find(&out).Error
	return out, err
}

// SavePlan 新增或修改套餐。修改只影响之后开通、续费、升级的周期
func (m *Meter) SavePlan(ctx context.Context, op Operator, p *entity.BillPlan) error {
	p.Name = strings.TrimSpace(p.Name)
	switch {
	case p.Name == "":
		return errors.New("请填写套餐名称")
	case p.MonthlyFee < 0 || p.QuotaMinutes < 0 || p.QuotaFluxGB < 0 || p.QuotaStorageGB < 0:
		return errors.New("月费和额度不能为负")
	case p.Status != "on" && p.Status != "off":
		return errors.New("状态应为在售或下架")
	}
	p.UpdateBy = op.Name
	if p.ID == 0 {
		return m.db.WithContext(ctx).Create(p).Error
	}
	var old entity.BillPlan
	if err := m.db.WithContext(ctx).First(&old, p.ID).Error; err != nil {
		return errors.New("套餐不存在")
	}
	p.CreateTime = old.CreateTime
	return m.db.WithContext(ctx).Save(p).Error
}

// ListPackDefs 资源包定义；onSale 为 true 时只返回在售的
func (m *Meter) ListPackDefs(ctx context.Context, onSale bool) ([]entity.BillPackDef, error) {
	var out []entity.BillPackDef
	q := m.db.WithContext(ctx).Order("SORT_NO, METER, PRICE, ID")
	if onSale {
		q = q.Where("STATUS = 'on'")
	}
	err := q.Find(&out).Error
	return out, err
}

// SavePackDef 新增或修改资源包定义（只支持直播时长、播放流量）
func (m *Meter) SavePackDef(ctx context.Context, op Operator, p *entity.BillPackDef) error {
	p.Name = strings.TrimSpace(p.Name)
	switch {
	case p.Name == "":
		return errors.New("请填写资源包名称")
	case p.Meter != entity.MeterLiveDuration && p.Meter != entity.MeterPlayFlux:
		return errors.New("资源包只支持直播时长和播放流量")
	case p.Quantity <= 0:
		return errors.New("额度必须大于 0")
	case p.ValidMonths <= 0 || p.ValidMonths > 36:
		return errors.New("有效期应为 1–36 个月")
	case p.Price < 0:
		return errors.New("售价不能为负")
	case p.Status != "on" && p.Status != "off":
		return errors.New("状态应为在售或下架")
	}
	p.UpdateBy = op.Name
	if p.ID == 0 {
		return m.db.WithContext(ctx).Create(p).Error
	}
	var old entity.BillPackDef
	if err := m.db.WithContext(ctx).First(&old, p.ID).Error; err != nil {
		return errors.New("资源包不存在")
	}
	p.CreateTime = old.CreateTime
	return m.db.WithContext(ctx).Save(p).Error
}
