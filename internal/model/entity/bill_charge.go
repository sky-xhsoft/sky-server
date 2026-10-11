package entity

import "time"

// 计费二期：价格、套餐、资源包、账户、费用明细、余额流水、月账单。
// 金额一律以「厘」（0.001 元）整数存储；单价以 0.0001 元为单位，便于表示 0.0025 元 / 分钟这类小额单价

// 余额流水类型
const (
	TxnRecharge    = "recharge"     // 充值（线下录入）
	TxnGift        = "gift"         // 赠送
	TxnUsage       = "usage"        // 用量扣费
	TxnUsageAdjust = "usage_adjust" // 重新计量后的差额补扣 / 退还
	TxnPlanFee     = "plan_fee"     // 套餐费（开通、续费、升级）
	TxnPackFee     = "pack_fee"     // 购买资源包
	TxnAdjust      = "adjust"       // 平台调账
)

// BillPrice 单价：公司为 0 时是默认价，否则是该公司的专属价；同一公司同一计费点取生效日期不晚于用量日期的最新一条
type BillPrice struct {
	ID            uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID  uint      `gorm:"column:SYS_COMPANY_ID;not null;default:0;index:IDX_BP_LOOKUP,priority:1" json:"companyId"`
	Meter         string    `gorm:"column:METER;type:varchar(32);not null;index:IDX_BP_LOOKUP,priority:2" json:"meter"`
	UnitPrice     int64     `gorm:"column:UNIT_PRICE;not null" json:"unitPrice"` // 单位：0.0001 元 / 计价单位（存储为 GB·月）
	EffectiveDate string    `gorm:"column:EFFECTIVE_DATE;type:char(10);not null;index:IDX_BP_LOOKUP,priority:3" json:"effectiveDate"`
	Remark        string    `gorm:"column:REMARK;type:varchar(255)" json:"remark"`
	CreateBy      string    `gorm:"column:CREATE_BY;type:varchar(80)" json:"createBy"`
	CreateTime    time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
}

func (BillPrice) TableName() string { return "bill_price" }

// BillPlan 套餐：月费 + 各计费点的月度额度
type BillPlan struct {
	ID             uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"column:NAME;type:varchar(64);not null" json:"name"`
	MonthlyFee     int64     `gorm:"column:MONTHLY_FEE;not null" json:"monthlyFee"` // 厘
	QuotaMinutes   float64   `gorm:"column:QUOTA_MINUTES;type:decimal(20,6);not null;default:0" json:"quotaMinutes"`
	QuotaFluxGB    float64   `gorm:"column:QUOTA_FLUX_GB;type:decimal(20,6);not null;default:0" json:"quotaFluxGb"`
	QuotaStorageGB float64   `gorm:"column:QUOTA_STORAGE_GB;type:decimal(20,6);not null;default:0" json:"quotaStorageGb"` // 日均存储额度，超出部分按量
	Status         string    `gorm:"column:STATUS;type:varchar(10);not null;default:'on'" json:"status"`                  // on 在售 / off 下架
	SortNo         int       `gorm:"column:SORT_NO;not null;default:0" json:"sortNo"`
	Description    string    `gorm:"column:DESCRIPTION;type:varchar(500)" json:"description"`
	UpdateBy       string    `gorm:"column:UPDATE_BY;type:varchar(80)" json:"updateBy"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateTime     time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillPlan) TableName() string { return "bill_plan" }

// BillPackDef 资源包定义（直播时长或播放流量）
type BillPackDef struct {
	ID          uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:NAME;type:varchar(64);not null" json:"name"`
	Meter       string    `gorm:"column:METER;type:varchar(32);not null" json:"meter"`
	Quantity    float64   `gorm:"column:QUANTITY;type:decimal(20,6);not null" json:"quantity"`
	ValidMonths int       `gorm:"column:VALID_MONTHS;not null" json:"validMonths"`
	Price       int64     `gorm:"column:PRICE;not null" json:"price"` // 厘
	Status      string    `gorm:"column:STATUS;type:varchar(10);not null;default:'on'" json:"status"`
	SortNo      int       `gorm:"column:SORT_NO;not null;default:0" json:"sortNo"`
	Description string    `gorm:"column:DESCRIPTION;type:varchar(500)" json:"description"`
	UpdateBy    string    `gorm:"column:UPDATE_BY;type:varchar(80)" json:"updateBy"`
	CreateTime  time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillPackDef) TableName() string { return "bill_pack_def" }

// BillAccount 公司计费账户
type BillAccount struct {
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;primaryKey;autoIncrement:false" json:"companyId"`
	Balance      int64     `gorm:"column:BALANCE;not null;default:0" json:"balance"`          // 厘，可为负（欠费）
	GiftBalance  int64     `gorm:"column:GIFT_BALANCE;not null;default:0" json:"giftBalance"` // 余额中来自赠送的部分（先于现金消费）
	CreditLimit  int64     `gorm:"column:CREDIT_LIMIT;not null;default:0" json:"creditLimit"` // 授信额度，厘
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillAccount) TableName() string { return "bill_account" }

// BillSubscription 套餐周期：从开通日起一个月（含首尾）；周期内升级直接修改本行
type BillSubscription struct {
	ID             uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID   uint      `gorm:"column:SYS_COMPANY_ID;not null;index:IDX_BS_COMPANY,priority:1" json:"companyId"`
	PlanID         uint      `gorm:"column:PLAN_ID;not null" json:"planId"`
	PlanName       string    `gorm:"column:PLAN_NAME;type:varchar(64);not null" json:"planName"`
	StartDate      string    `gorm:"column:START_DATE;type:char(10);not null;index:IDX_BS_COMPANY,priority:2" json:"startDate"`
	EndDate        string    `gorm:"column:END_DATE;type:char(10);not null" json:"endDate"`
	QuotaMinutes   float64   `gorm:"column:QUOTA_MINUTES;type:decimal(20,6);not null" json:"quotaMinutes"`
	QuotaFluxGB    float64   `gorm:"column:QUOTA_FLUX_GB;type:decimal(20,6);not null" json:"quotaFluxGb"`
	QuotaStorageGB float64   `gorm:"column:QUOTA_STORAGE_GB;type:decimal(20,6);not null" json:"quotaStorageGb"`
	UsedMinutes    float64   `gorm:"column:USED_MINUTES;type:decimal(20,6);not null;default:0" json:"usedMinutes"`
	UsedFluxGB     float64   `gorm:"column:USED_FLUX_GB;type:decimal(20,6);not null;default:0" json:"usedFluxGb"`
	Fee            int64     `gorm:"column:FEE;not null" json:"fee"` // 本周期已收套餐费（含升级补差），厘
	AutoRenew      bool      `gorm:"column:AUTO_RENEW;not null;default:true" json:"autoRenew"`
	Status         string    `gorm:"column:STATUS;type:varchar(10);not null;default:'active'" json:"status"` // active 有效 / ended 已结束
	CreateBy       string    `gorm:"column:CREATE_BY;type:varchar(80)" json:"createBy"`
	CreateTime     time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateTime     time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillSubscription) TableName() string { return "bill_subscription" }

// BillResourcePack 公司持有的资源包
type BillResourcePack struct {
	ID           uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;not null;index" json:"companyId"`
	PackDefID    uint      `gorm:"column:PACK_DEF_ID;not null" json:"packDefId"`
	Name         string    `gorm:"column:NAME;type:varchar(64);not null" json:"name"`
	Meter        string    `gorm:"column:METER;type:varchar(32);not null" json:"meter"`
	Total        float64   `gorm:"column:TOTAL;type:decimal(20,6);not null" json:"total"`
	Remaining    float64   `gorm:"column:REMAINING;type:decimal(20,6);not null" json:"remaining"`
	ValidFrom    string    `gorm:"column:VALID_FROM;type:char(10);not null" json:"validFrom"`
	ValidTo      string    `gorm:"column:VALID_TO;type:char(10);not null" json:"validTo"`
	Source       string    `gorm:"column:SOURCE;type:varchar(20);not null" json:"source"` // purchase 购买 / grant 平台发放
	Price        int64     `gorm:"column:PRICE;not null;default:0" json:"price"`          // 购买价，厘
	CreateBy     string    `gorm:"column:CREATE_BY;type:varchar(80)" json:"createBy"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
}

func (BillResourcePack) TableName() string { return "bill_resource_pack" }

// BillCharge 费用明细：每个公司、日期、计费点一行。重新计价时按差额补扣或退还
type BillCharge struct {
	ID             uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID   uint      `gorm:"column:SYS_COMPANY_ID;not null;uniqueIndex:UK_BC_DAY,priority:1" json:"-"`
	UsageDate      string    `gorm:"column:USAGE_DATE;type:char(10);not null;uniqueIndex:UK_BC_DAY,priority:2" json:"usageDate"`
	Meter          string    `gorm:"column:METER;type:varchar(32);not null;uniqueIndex:UK_BC_DAY,priority:3" json:"meter"`
	Usage          float64   `gorm:"column:USAGE_QTY;type:decimal(20,6);not null" json:"usage"`
	PlanOffset     float64   `gorm:"column:PLAN_OFFSET;type:decimal(20,6);not null;default:0" json:"planOffset"`
	PackOffset     float64   `gorm:"column:PACK_OFFSET;type:decimal(20,6);not null;default:0" json:"packOffset"`
	PackDetail     string    `gorm:"column:PACK_DETAIL;type:varchar(1000)" json:"-"` // 资源包抵扣明细 {"包ID": 数量}
	Billable       float64   `gorm:"column:BILLABLE;type:decimal(20,6);not null" json:"billable"`
	UnitPrice      int64     `gorm:"column:UNIT_PRICE;not null" json:"unitPrice"`       // 0.0001 元
	Amount         int64     `gorm:"column:AMOUNT;not null" json:"amount"`              // 应扣，厘
	Charged        int64     `gorm:"column:CHARGED;not null;default:0" json:"charged"`  // 已从余额扣除，厘
	Mode           string    `gorm:"column:MODE;type:varchar(10);not null" json:"mode"` // trial 试算 / charge 扣费
	SubscriptionID uint      `gorm:"column:SUBSCRIPTION_ID;not null;default:0" json:"subscriptionId"`
	UpdateTime     time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
}

func (BillCharge) TableName() string { return "bill_charge" }

// BillTransaction 余额流水。余额只通过流水变动；PERIOD 为计入的账期（月账单按账期汇总）
type BillTransaction struct {
	ID           uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;not null;index:IDX_BT_COMPANY,priority:1" json:"-"`
	Period       string    `gorm:"column:PERIOD;type:char(7);not null;index:IDX_BT_COMPANY,priority:2" json:"period"`
	Type         string    `gorm:"column:TYPE;type:varchar(20);not null" json:"type"`
	Meter        string    `gorm:"column:METER;type:varchar(32);not null;default:''" json:"meter"` // 用量扣费的计费点
	Amount       int64     `gorm:"column:AMOUNT;not null" json:"amount"`                           // 厘，正为收入
	GiftAmount   int64     `gorm:"column:GIFT_AMOUNT;not null;default:0" json:"giftAmount"`        // 其中赠送部分的变动
	BalanceAfter int64     `gorm:"column:BALANCE_AFTER;not null" json:"balanceAfter"`
	RefType      string    `gorm:"column:REF_TYPE;type:varchar(32)" json:"refType"`
	RefID        uint      `gorm:"column:REF_ID;not null;default:0" json:"refId"`
	UsageDate    string    `gorm:"column:USAGE_DATE;type:char(10);not null;default:''" json:"usageDate"` // 用量扣费对应的用量日期
	Voucher      string    `gorm:"column:VOUCHER;type:varchar(128)" json:"voucher"`                      // 线下充值凭证号
	Remark       string    `gorm:"column:REMARK;type:varchar(500)" json:"remark"`
	Operator     string    `gorm:"column:OPERATOR;type:varchar(80)" json:"operator"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
}

func (BillTransaction) TableName() string { return "bill_transaction" }

// BillStatement 月账单
type BillStatement struct {
	ID            uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID  uint      `gorm:"column:SYS_COMPANY_ID;not null;uniqueIndex:UK_BST_MONTH,priority:1" json:"companyId"`
	Month         string    `gorm:"column:MONTH;type:char(7);not null;uniqueIndex:UK_BST_MONTH,priority:2" json:"month"`
	Opening       int64     `gorm:"column:OPENING;not null" json:"opening"`
	UsageDuration int64     `gorm:"column:USAGE_DURATION;not null;default:0" json:"usageDuration"` // 以下为支出，正数表示金额
	UsageFlux     int64     `gorm:"column:USAGE_FLUX;not null;default:0" json:"usageFlux"`
	UsageStorage  int64     `gorm:"column:USAGE_STORAGE;not null;default:0" json:"usageStorage"`
	UsageAdjust   int64     `gorm:"column:USAGE_ADJUST;not null;default:0" json:"usageAdjust"` // 以往月份的补扣（正）/ 退还（负）
	PlanFee       int64     `gorm:"column:PLAN_FEE;not null;default:0" json:"planFee"`
	PackFee       int64     `gorm:"column:PACK_FEE;not null;default:0" json:"packFee"`
	Recharge      int64     `gorm:"column:RECHARGE;not null;default:0" json:"recharge"` // 以下为收入
	Gift          int64     `gorm:"column:GIFT;not null;default:0" json:"gift"`
	Adjust        int64     `gorm:"column:ADJUST;not null;default:0" json:"adjust"` // 调账（可正可负）
	Closing       int64     `gorm:"column:CLOSING;not null" json:"closing"`
	Detail        string    `gorm:"column:DETAIL;type:mediumtext" json:"-"` // 用量与抵扣明细（JSON）
	CreateTime    time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
}

func (BillStatement) TableName() string { return "bill_statement" }

// BillAdjustRequest 超过复核金额的赠送 / 调账申请，须另一位平台管理员复核后入账
type BillAdjustRequest struct {
	ID           uint       `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint       `gorm:"column:SYS_COMPANY_ID;not null;index" json:"companyId"`
	Type         string     `gorm:"column:TYPE;type:varchar(20);not null" json:"type"` // gift / adjust
	Amount       int64      `gorm:"column:AMOUNT;not null" json:"amount"`
	Reason       string     `gorm:"column:REASON;type:varchar(500);not null" json:"reason"`
	Status       string     `gorm:"column:STATUS;type:varchar(10);not null;default:'pending'" json:"status"` // pending / approved / rejected
	RequestBy    string     `gorm:"column:REQUEST_BY;type:varchar(80);not null" json:"requestBy"`
	RequestByID  uint       `gorm:"column:REQUEST_BY_ID;not null" json:"-"`
	ReviewBy     string     `gorm:"column:REVIEW_BY;type:varchar(80)" json:"reviewBy"`
	ReviewTime   *time.Time `gorm:"column:REVIEW_TIME" json:"reviewTime"`
	ReviewRemark string     `gorm:"column:REVIEW_REMARK;type:varchar(500)" json:"reviewRemark"`
	TxnID        uint       `gorm:"column:TXN_ID;not null;default:0" json:"txnId"`
	CreateTime   time.Time  `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
}

func (BillAdjustRequest) TableName() string { return "bill_adjust_request" }
