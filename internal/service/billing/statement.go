package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StatementMeter 账单中一个计费点的用量与抵扣
type StatementMeter struct {
	Meter      string  `json:"meter"`
	Usage      float64 `json:"usage"`
	PlanOffset float64 `json:"planOffset"`
	PackOffset float64 `json:"packOffset"`
	Billable   float64 `json:"billable"`
	Amount     int64   `json:"amount"` // 本月用量对应的费用（已扣费的日期为实扣，试算日期为试算）
	TrialDays  int     `json:"trialDays"`
}

// StatementDetail 账单明细
type StatementDetail struct {
	Meters        []StatementMeter          `json:"meters"`
	Subscriptions []entity.BillSubscription `json:"subscriptions"`
	Rooms         []RoomUsage               `json:"rooms"`
}

// StatementView 账单与明细
type StatementView struct {
	entity.BillStatement
	Detail StatementDetail `json:"detail"`
}

func prevMonth(month string) string {
	t, _ := time.ParseInLocation(monthLayoutBill, month, time.Local)
	return t.AddDate(0, -1, 0).Format(monthLayoutBill)
}

// GenerateStatements 生成某月全部公司的账单。出账前核对每个公司的流水合计等于账户余额，不一致的公司不出账并返回错误；
// 出账后该月关闭，之后发生的该月补扣计入下一个月
func (m *Meter) GenerateStatements(ctx context.Context, month string) error {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return err
	}
	if !d1.After(d0) || month >= m.now().Format(monthLayoutBill) {
		return errors.New("只能为已结束的月份出账")
	}
	var companies []uint
	if err := m.db.WithContext(ctx).Raw(
		"SELECT SYS_COMPANY_ID FROM bill_account UNION SELECT DISTINCT SYS_COMPANY_ID FROM bill_charge WHERE USAGE_DATE >= ? AND USAGE_DATE < ?",
		d0.Format(dayLayout), d1.Format(dayLayout)).Scan(&companies).Error; err != nil {
		return err
	}
	var errs []error
	for _, c := range companies {
		if c == 0 {
			continue
		}
		if err := m.generateStatement(ctx, c, month); err != nil {
			zap.L().Error("计费：出账失败", zap.Uint("company", c), zap.String("month", month), zap.Error(err))
			errs = append(errs, fmt.Errorf("公司 %d：%w", c, err))
		}
	}
	if len(errs) == 0 && month > m.getState(ctx, statementClosedKey) {
		if err := m.setState(ctx, statementClosedKey, month); err != nil {
			return err
		}
	}
	return errors.Join(errs...)
}

func (m *Meter) generateStatement(ctx context.Context, companyID uint, month string) error {
	d0, d1, _ := m.monthRange(month)
	// 直播间用量只读用量表，放在事务外查询
	rooms, err := m.Rooms(ctx, companyID, month)
	if err != nil {
		return err
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := lockAccount(tx, companyID)
		if err != nil {
			return err
		}
		var total int64
		tx.Model(&entity.BillTransaction{}).Where("SYS_COMPANY_ID = ?", companyID).Select("COALESCE(SUM(AMOUNT), 0)").Scan(&total)
		if total != acc.Balance {
			return fmt.Errorf("流水合计 %s 元与账户余额 %s 元不一致，请核对", yuan(total), yuan(acc.Balance))
		}
		st := entity.BillStatement{SysCompanyID: companyID, Month: month}
		var prev entity.BillStatement
		if tx.Where("SYS_COMPANY_ID = ? AND MONTH = ?", companyID, prevMonth(month)).First(&prev).Error == nil {
			st.Opening = prev.Closing
		} else {
			tx.Model(&entity.BillTransaction{}).Where("SYS_COMPANY_ID = ? AND PERIOD < ?", companyID, month).
				Select("COALESCE(SUM(AMOUNT), 0)").Scan(&st.Opening)
		}
		var txns []entity.BillTransaction
		if err := tx.Where("SYS_COMPANY_ID = ? AND PERIOD = ?", companyID, month).Find(&txns).Error; err != nil {
			return err
		}
		sum := int64(0)
		for _, t := range txns {
			sum += t.Amount
			switch t.Type {
			case entity.TxnUsage, entity.TxnUsageAdjust:
				if t.UsageDate < d0.Format(dayLayout) {
					st.UsageAdjust -= t.Amount
					continue
				}
				switch t.Meter {
				case entity.MeterLiveDuration:
					st.UsageDuration -= t.Amount
				case entity.MeterPlayFlux:
					st.UsageFlux -= t.Amount
				default:
					st.UsageStorage -= t.Amount
				}
			case entity.TxnPlanFee:
				st.PlanFee -= t.Amount
			case entity.TxnPackFee:
				st.PackFee -= t.Amount
			case entity.TxnRecharge:
				st.Recharge += t.Amount
			case entity.TxnGift:
				st.Gift += t.Amount
			default:
				st.Adjust += t.Amount
			}
		}
		st.Closing = st.Opening + sum
		if st.Closing != st.Opening+st.Recharge+st.Gift+st.Adjust-st.UsageDuration-st.UsageFlux-st.UsageStorage-st.UsageAdjust-st.PlanFee-st.PackFee {
			return errors.New("账单收支不平")
		}
		detail, err := statementDetail(tx, companyID, d0, d1)
		if err != nil {
			return err
		}
		detail.Rooms = rooms
		b, _ := json.Marshal(detail)
		st.Detail = string(b)
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "SYS_COMPANY_ID"}, {Name: "MONTH"}},
			DoUpdates: clause.AssignmentColumns([]string{"OPENING", "USAGE_DURATION", "USAGE_FLUX", "USAGE_STORAGE", "USAGE_ADJUST", "PLAN_FEE", "PACK_FEE", "RECHARGE", "GIFT", "ADJUST", "CLOSING", "DETAIL"}),
		}).Create(&st).Error
	})
}

func statementDetail(tx *gorm.DB, companyID uint, d0, d1 time.Time) (StatementDetail, error) {
	out := StatementDetail{Meters: []StatementMeter{}, Subscriptions: []entity.BillSubscription{}, Rooms: []RoomUsage{}}
	var charges []entity.BillCharge
	if err := tx.Where("SYS_COMPANY_ID = ? AND USAGE_DATE >= ? AND USAGE_DATE < ?", companyID, d0.Format(dayLayout), d1.Format(dayLayout)).
		Find(&charges).Error; err != nil {
		return out, err
	}
	by := map[string]*StatementMeter{}
	for _, meter := range meterOrder {
		by[meter] = &StatementMeter{Meter: meter}
	}
	for _, c := range charges {
		s := by[c.Meter]
		if s == nil {
			continue
		}
		s.Usage += c.Usage
		s.PlanOffset += c.PlanOffset
		s.PackOffset += c.PackOffset
		s.Billable += c.Billable
		s.Amount += c.Amount
		if c.Mode != "charge" {
			s.TrialDays++
		}
	}
	for _, meter := range meterOrder {
		s := by[meter]
		s.Usage, s.PlanOffset, s.PackOffset, s.Billable = round6(s.Usage), round6(s.PlanOffset), round6(s.PackOffset), round6(s.Billable)
		out.Meters = append(out.Meters, *s)
	}
	tx.Where("SYS_COMPANY_ID = ? AND START_DATE < ? AND END_DATE >= ?", companyID, d1.Format(dayLayout), d0.Format(dayLayout)).
		Order("START_DATE").Find(&out.Subscriptions)
	return out, nil
}

// Statements 公司的月账单列表（新的在前）
func (m *Meter) Statements(ctx context.Context, companyID uint) ([]entity.BillStatement, error) {
	var out []entity.BillStatement
	err := m.db.WithContext(ctx).Omit("DETAIL").Where("SYS_COMPANY_ID = ?", companyID).Order("MONTH DESC").Find(&out).Error
	return out, err
}

// Statement 某月账单与明细
func (m *Meter) Statement(ctx context.Context, companyID uint, month string) (*StatementView, error) {
	var st entity.BillStatement
	if err := m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND MONTH = ?", companyID, month).First(&st).Error; err != nil {
		return nil, errors.New("该月账单尚未生成")
	}
	v := &StatementView{BillStatement: st}
	_ = json.Unmarshal([]byte(st.Detail), &v.Detail)
	return v, nil
}

var meterNames = map[string]struct{ name, unit string }{
	entity.MeterLiveDuration: {"直播时长", "分钟"},
	entity.MeterPlayFlux:     {"播放流量", "GB"},
	entity.MeterStorage:      {"存储空间", "GB·天"},
}

var txnNames = map[string]string{
	entity.TxnRecharge: "充值", entity.TxnGift: "赠送", entity.TxnUsage: "用量扣费", entity.TxnUsageAdjust: "补扣 / 退还",
	entity.TxnPlanFee: "套餐费", entity.TxnPackFee: "资源包", entity.TxnAdjust: "调账",
}

// ExportStatement 导出某月账单（Excel）：汇总、每日费用、直播间用量、余额流水
func (m *Meter) ExportStatement(ctx context.Context, companyID uint, month string) ([]byte, error) {
	v, err := m.Statement(ctx, companyID, month)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	defer f.Close()
	money := func(li int64) float64 { return float64(li) / 1000 }

	sh := "账单汇总"
	f.SetSheetName("Sheet1", sh)
	rows := [][]any{
		{"账单月份", month},
		{"期初余额（元）", money(v.Opening)},
		{"充值（元）", money(v.Recharge)},
		{"赠送（元）", money(v.Gift)},
		{"调账（元）", money(v.Adjust)},
		{"直播时长费用（元）", -money(v.UsageDuration)},
		{"播放流量费用（元）", -money(v.UsageFlux)},
		{"存储空间费用（元）", -money(v.UsageStorage)},
		{"以往月份补扣 / 退还（元）", -money(v.UsageAdjust)},
		{"套餐费（元）", -money(v.PlanFee)},
		{"资源包（元）", -money(v.PackFee)},
		{"期末余额（元）", money(v.Closing)},
		{},
		{"计费点", "用量", "单位", "套餐抵扣", "资源包抵扣", "计费量", "费用（元）", "试算天数"},
	}
	for _, s := range v.Detail.Meters {
		n := meterNames[s.Meter]
		rows = append(rows, []any{n.name, s.Usage, n.unit, s.PlanOffset, s.PackOffset, s.Billable, money(s.Amount), s.TrialDays})
	}
	writeRows(f, sh, rows)

	sh = "每日费用"
	f.NewSheet(sh)
	d0, d1, _ := m.monthRange(month)
	var charges []entity.BillCharge
	m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND USAGE_DATE >= ? AND USAGE_DATE < ?", companyID, d0.Format(dayLayout), d1.Format(dayLayout)).
		Order("USAGE_DATE, METER").Find(&charges)
	rows = [][]any{{"日期", "计费点", "用量", "套餐抵扣", "资源包抵扣", "计费量", "单价（元）", "费用（元）", "状态"}}
	for _, c := range charges {
		mode := "已扣费"
		if c.Mode != "charge" {
			mode = "试算"
		}
		rows = append(rows, []any{c.UsageDate, meterNames[c.Meter].name, c.Usage, c.PlanOffset, c.PackOffset, c.Billable, float64(c.UnitPrice) / 10000, money(c.Amount), mode})
	}
	writeRows(f, sh, rows)

	sh = "直播间用量"
	f.NewSheet(sh)
	rows = [][]any{{"直播间 ID", "直播间", "直播时长（分钟）", "推流次数", "播放流量（GB）"}}
	for _, r := range v.Detail.Rooms {
		name := r.RoomName
		if r.Deleted {
			name += "（已删除）"
		}
		rows = append(rows, []any{r.RoomID, name, r.LiveMinutes, r.Sessions, r.PlayGB})
	}
	writeRows(f, sh, rows)

	sh = "余额流水"
	f.NewSheet(sh)
	var txns []entity.BillTransaction
	m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND PERIOD = ?", companyID, month).Order("ID").Find(&txns)
	rows = [][]any{{"时间", "类型", "金额（元）", "变动后余额（元）", "说明", "操作人"}}
	for _, t := range txns {
		rows = append(rows, []any{t.CreateTime.Format("2006-01-02 15:04:05"), txnNames[t.Type], money(t.Amount), money(t.BalanceAfter), t.Remark, t.Operator})
	}
	writeRows(f, sh, rows)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeRows(f *excelize.File, sheet string, rows [][]any) {
	for i, r := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		_ = f.SetSheetRow(sheet, cell, &r)
	}
	_ = f.SetColWidth(sheet, "A", "I", 16)
}
