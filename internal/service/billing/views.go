package billing

import (
	"context"
	"math"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// MeterCost 本月某计费点的费用
type MeterCost struct {
	Meter  string `json:"meter"`
	Amount int64  `json:"amount"` // 厘
}

// AccountOverview 费用中心概览
type AccountOverview struct {
	Account      entity.BillAccount        `json:"account"`
	Status       string                    `json:"status"` // normal 正常 / arrears 欠费（余额加授信小于 0）
	Settings     Settings                  `json:"settings"`
	Month        string                    `json:"month"`
	RatedDay     string                    `json:"ratedDay"`     // 本月已计价到的日期
	UsageCost    int64                     `json:"usageCost"`    // 本月用量费用（试算期间为试算金额）
	MeterCosts   []MeterCost               `json:"meterCosts"`   // 本月各计费点费用
	FeeCost      int64                     `json:"feeCost"`      // 本月套餐费、资源包
	Estimate     int64                     `json:"estimate"`     // 按已计价天数的日均推算到月底的总费用
	Subscription *entity.BillSubscription  `json:"subscription"` // 当前套餐周期
	Packs        []entity.BillResourcePack `json:"packs"`        // 有效的资源包
	PendingCount int64                     `json:"pendingCount"` // 平台视图：待复核申请数
}

// Overview 公司的费用概览
func (m *Meter) Overview(ctx context.Context, companyID uint) (*AccountOverview, error) {
	now := m.now()
	today := now.Format(dayLayout)
	month := now.Format(monthLayoutBill)
	out := &AccountOverview{Account: m.account(ctx, companyID), Settings: m.GetSettings(ctx), Month: month, MeterCosts: []MeterCost{}, Packs: []entity.BillResourcePack{}}
	out.Status = "normal"
	if out.Account.Balance+out.Account.CreditLimit < 0 {
		out.Status = "arrears"
	}
	d0, d1, _ := m.monthRange(month)
	type row struct {
		Meter  string
		Amount int64
		Last   string
	}
	var rows []row
	m.db.WithContext(ctx).Model(&entity.BillCharge{}).Select("METER AS meter, SUM(AMOUNT) AS amount, MAX(USAGE_DATE) AS last").
		Where("SYS_COMPANY_ID = ? AND USAGE_DATE >= ? AND USAGE_DATE < ?", companyID, d0.Format(dayLayout), d1.Format(dayLayout)).
		Group("METER").Scan(&rows)
	byMeter := map[string]int64{}
	for _, r := range rows {
		byMeter[r.Meter] = r.Amount
		out.UsageCost += r.Amount
		if r.Last > out.RatedDay {
			out.RatedDay = r.Last
		}
	}
	for _, meter := range meterOrder {
		out.MeterCosts = append(out.MeterCosts, MeterCost{Meter: meter, Amount: byMeter[meter]})
	}
	m.db.WithContext(ctx).Model(&entity.BillTransaction{}).Select("COALESCE(-SUM(AMOUNT), 0)").
		Where("SYS_COMPANY_ID = ? AND PERIOD = ? AND TYPE IN ?", companyID, month, []string{entity.TxnPlanFee, entity.TxnPackFee}).Scan(&out.FeeCost)
	out.Estimate = out.UsageCost + out.FeeCost
	if out.RatedDay != "" {
		rated := daysBetween(d0.Format(dayLayout), out.RatedDay)
		out.Estimate = int64(math.Round(float64(out.UsageCost)/float64(rated)*float64(daysInMonth(out.RatedDay)))) + out.FeeCost
	}
	if sub, err := activeSub(m.db.WithContext(ctx), companyID, today); err == nil {
		out.Subscription = sub
	}
	m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND VALID_TO >= ? AND REMAINING > 0", companyID, today).Order("VALID_TO, ID").Find(&out.Packs)
	return out, nil
}

// Charges 公司某月的费用明细（日期倒序）
func (m *Meter) Charges(ctx context.Context, companyID uint, month string) ([]entity.BillCharge, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	out := []entity.BillCharge{}
	err = m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ? AND USAGE_DATE >= ? AND USAGE_DATE < ?", companyID, d0.Format(dayLayout), d1.Format(dayLayout)).
		Order("USAGE_DATE DESC, METER").Find(&out).Error
	return out, err
}

// TxnPage 余额流水分页
type TxnPage struct {
	Items []entity.BillTransaction `json:"items"`
	Total int64                    `json:"total"`
}

// Transactions 公司的余额流水；period 为空时不限账期，typ 为空时不限类型
func (m *Meter) Transactions(ctx context.Context, companyID uint, period, typ string, page, pageSize int) (*TxnPage, error) {
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	q := m.db.WithContext(ctx).Model(&entity.BillTransaction{}).Where("SYS_COMPANY_ID = ?", companyID)
	if period != "" {
		q = q.Where("PERIOD = ?", period)
	}
	if typ != "" {
		q = q.Where("TYPE = ?", typ)
	}
	out := &TxnPage{Items: []entity.BillTransaction{}}
	if err := q.Count(&out.Total).Error; err != nil {
		return nil, err
	}
	err := q.Order("ID DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&out.Items).Error
	return out, err
}

// AccountRow 平台：公司账户列表一行
type AccountRow struct {
	CompanyID   uint   `json:"companyId"`
	CompanyName string `json:"companyName"`
	Balance     int64  `json:"balance"`
	GiftBalance int64  `json:"giftBalance"`
	CreditLimit int64  `json:"creditLimit"`
	Status      string `json:"status"`
	PlanName    string `json:"planName"`
	PlanEndDate string `json:"planEndDate"`
	MonthCost   int64  `json:"monthCost"` // 本月用量费用
	LastTxnTime string `json:"lastTxnTime"`
}

// AccountList 平台：全部有效公司的账户（没有账户的公司余额为 0）
func (m *Meter) AccountList(ctx context.Context) ([]AccountRow, error) {
	type company struct {
		ID   uint
		Name string
	}
	var companies []company
	if err := m.db.WithContext(ctx).Table("sys_company").Select("ID AS id, NAME AS name").Where("IS_ACTIVE = 'Y'").Order("ID").Scan(&companies).Error; err != nil {
		return nil, err
	}
	var accs []entity.BillAccount
	m.db.WithContext(ctx).Find(&accs)
	accBy := map[uint]entity.BillAccount{}
	for _, a := range accs {
		accBy[a.SysCompanyID] = a
	}
	now := m.now()
	today, month := now.Format(dayLayout), now.Format(monthLayoutBill)
	var subs []entity.BillSubscription
	m.db.WithContext(ctx).Where("STATUS = 'active' AND START_DATE <= ? AND END_DATE >= ?", today, today).Find(&subs)
	subBy := map[uint]entity.BillSubscription{}
	for _, s := range subs {
		subBy[s.SysCompanyID] = s
	}
	d0, d1, _ := m.monthRange(month)
	type cost struct {
		Company uint
		Amount  int64
	}
	var costs []cost
	m.db.WithContext(ctx).Model(&entity.BillCharge{}).Select("SYS_COMPANY_ID AS company, SUM(AMOUNT) AS amount").
		Where("USAGE_DATE >= ? AND USAGE_DATE < ?", d0.Format(dayLayout), d1.Format(dayLayout)).Group("SYS_COMPANY_ID").Scan(&costs)
	costBy := map[uint]int64{}
	for _, c := range costs {
		costBy[c.Company] = c.Amount
	}
	type last struct {
		Company uint
		T       time.Time
	}
	var lasts []last
	m.db.WithContext(ctx).Model(&entity.BillTransaction{}).Select("SYS_COMPANY_ID AS company, MAX(CREATE_TIME) AS t").Group("SYS_COMPANY_ID").Scan(&lasts)
	lastBy := map[uint]string{}
	for _, l := range lasts {
		lastBy[l.Company] = l.T.Format("2006-01-02 15:04")
	}
	out := make([]AccountRow, 0, len(companies))
	for _, c := range companies {
		a := accBy[c.ID]
		r := AccountRow{CompanyID: c.ID, CompanyName: c.Name, Balance: a.Balance, GiftBalance: a.GiftBalance, CreditLimit: a.CreditLimit,
			Status: "normal", MonthCost: costBy[c.ID], LastTxnTime: lastBy[c.ID]}
		if a.Balance+a.CreditLimit < 0 {
			r.Status = "arrears"
		}
		if s, ok := subBy[c.ID]; ok {
			r.PlanName, r.PlanEndDate = s.PlanName, s.EndDate
		}
		out = append(out, r)
	}
	return out, nil
}

// AdjustRequests 平台：赠送 / 调账申请；status 为空时返回全部
func (m *Meter) AdjustRequests(ctx context.Context, status string) ([]entity.BillAdjustRequest, error) {
	out := []entity.BillAdjustRequest{}
	q := m.db.WithContext(ctx).Order("ID DESC").Limit(200)
	if status != "" {
		q = q.Where("STATUS = ?", status)
	}
	err := q.Find(&out).Error
	return out, err
}

// PendingRequestCount 待复核申请数
func (m *Meter) PendingRequestCount(ctx context.Context) int64 {
	var n int64
	m.db.WithContext(ctx).Model(&entity.BillAdjustRequest{}).Where("STATUS = 'pending'").Count(&n)
	return n
}
