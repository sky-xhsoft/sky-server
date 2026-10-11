package billing

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 计费设置（存于 bill_meter_state）
const (
	chargeModeKey      = "charge_mode"      // trial 试算 / charge 扣费
	chargeFromKey      = "charge_from"      // 开始扣费的用量日期
	reviewThresholdKey = "review_threshold" // 赠送、调账超过该金额（厘）须复核
	statementClosedKey = "statement_closed" // 已出账的最后一个月份（YYYY-MM）

	defaultReviewThreshold = int64(10000 * 1000) // 1 万元
	monthLayoutBill        = "2006-01"
)

// ErrNeedReview 金额超过复核阈值，已生成待复核申请
var ErrNeedReview = errors.New("金额超过复核阈值，已提交复核")

// Operator 操作人
type Operator struct {
	UserID uint
	Name   string
}

// Settings 计费设置
type Settings struct {
	ChargeMode      string `json:"chargeMode"`      // trial / charge
	ChargeFrom      string `json:"chargeFrom"`      // 开始扣费的用量日期
	ReviewThreshold int64  `json:"reviewThreshold"` // 厘
	StatementClosed string `json:"statementClosed"` // 已出账的最后月份
}

// GetSettings 读取计费设置
func (m *Meter) GetSettings(ctx context.Context) Settings {
	s := Settings{
		ChargeMode:      m.getState(ctx, chargeModeKey),
		ChargeFrom:      m.getState(ctx, chargeFromKey),
		StatementClosed: m.getState(ctx, statementClosedKey),
		ReviewThreshold: defaultReviewThreshold,
	}
	if s.ChargeMode != "charge" {
		s.ChargeMode = "trial"
	}
	if v, err := strconv.ParseInt(m.getState(ctx, reviewThresholdKey), 10, 64); err == nil && v > 0 {
		s.ReviewThreshold = v
	}
	return s
}

// UpdateSettings 修改扣费模式与复核阈值。开启扣费时，ChargeFrom 之前的用量保持试算、不补扣
func (m *Meter) UpdateSettings(ctx context.Context, mode, chargeFrom string, threshold int64) error {
	switch mode {
	case "trial":
	case "charge":
		if _, _, err := dayRange(chargeFrom); err != nil {
			return errors.New("请填写开始扣费的日期")
		}
	default:
		return fmt.Errorf("不支持的扣费模式：%s", mode)
	}
	if threshold <= 0 {
		return errors.New("复核金额必须大于 0")
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for k, v := range map[string]string{chargeModeKey: mode, chargeFromKey: chargeFrom, reviewThresholdKey: strconv.FormatInt(threshold, 10)} {
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&entity.BillMeterState{StateKey: k, StateValue: v}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// openPeriod 流水计入的账期：某月已出账后，再发生的该月补扣计入下一个未出账的月份
func openPeriod(closed, month string) string {
	if closed == "" || month > closed {
		return month
	}
	t, err := time.ParseInLocation(monthLayoutBill, closed, time.Local)
	if err != nil {
		return month
	}
	return t.AddDate(0, 1, 0).Format(monthLayoutBill)
}

// lockAccount 锁定公司账户（不存在时创建）
func lockAccount(tx *gorm.DB, companyID uint) (*entity.BillAccount, error) {
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.BillAccount{SysCompanyID: companyID}).Error; err != nil {
		return nil, err
	}
	var acc entity.BillAccount
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("SYS_COMPANY_ID = ?", companyID).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// spends 消费类流水先用赠送余额
func spends(t string) bool {
	switch t {
	case entity.TxnUsage, entity.TxnUsageAdjust, entity.TxnPlanFee, entity.TxnPackFee:
		return true
	}
	return false
}

// postTxn 在事务中记一笔流水并更新余额（调用方负责事务）
func (m *Meter) postTxn(ctx context.Context, tx *gorm.DB, t *entity.BillTransaction, month string) error {
	if t.Amount == 0 {
		return nil
	}
	acc, err := lockAccount(tx, t.SysCompanyID)
	if err != nil {
		return err
	}
	gift := int64(0)
	switch {
	case t.Type == entity.TxnGift:
		gift = t.Amount
	case spends(t.Type) && t.Amount < 0:
		gift = -min(acc.GiftBalance, -t.Amount)
	}
	acc.Balance += t.Amount
	acc.GiftBalance += gift
	if acc.GiftBalance > max(acc.Balance, 0) {
		gift -= acc.GiftBalance - max(acc.Balance, 0)
		acc.GiftBalance = max(acc.Balance, 0)
	}
	if acc.GiftBalance < 0 {
		gift -= acc.GiftBalance
		acc.GiftBalance = 0
	}
	t.GiftAmount = gift
	t.BalanceAfter = acc.Balance
	if month == "" {
		month = m.now().Format(monthLayoutBill)
	}
	t.Period = openPeriod(m.getStateTx(tx, statementClosedKey), month)
	if err := tx.Model(&entity.BillAccount{}).Where("SYS_COMPANY_ID = ?", acc.SysCompanyID).
		Updates(map[string]any{"BALANCE": acc.Balance, "GIFT_BALANCE": acc.GiftBalance}).Error; err != nil {
		return err
	}
	return tx.Create(t).Error
}

func (m *Meter) getStateTx(tx *gorm.DB, key string) string {
	var s entity.BillMeterState
	if err := tx.Where("STATE_KEY = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.StateValue
}

// Recharge 平台录入线下充值
func (m *Meter) Recharge(ctx context.Context, op Operator, companyID uint, amount int64, voucher, remark string) (*entity.BillTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("充值金额必须大于 0")
	}
	if voucher == "" {
		return nil, errors.New("请填写转账凭证号")
	}
	if err := m.ensureCompany(ctx, companyID); err != nil {
		return nil, err
	}
	t := &entity.BillTransaction{SysCompanyID: companyID, Type: entity.TxnRecharge, Amount: amount, Voucher: voucher, Remark: remark, Operator: op.Name}
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return m.postTxn(ctx, tx, t, "") })
	return t, err
}

// GiftOrAdjust 赠送（金额为正）或调账（可正可负）；超过复核阈值时生成待复核申请并返回 ErrNeedReview
func (m *Meter) GiftOrAdjust(ctx context.Context, op Operator, companyID uint, typ string, amount int64, reason string) (*entity.BillTransaction, *entity.BillAdjustRequest, error) {
	switch {
	case typ == entity.TxnGift && amount <= 0:
		return nil, nil, errors.New("赠送金额必须大于 0")
	case typ == entity.TxnAdjust && amount == 0:
		return nil, nil, errors.New("调账金额不能为 0")
	case typ != entity.TxnGift && typ != entity.TxnAdjust:
		return nil, nil, fmt.Errorf("不支持的操作：%s", typ)
	}
	if reason == "" {
		return nil, nil, errors.New("请填写原因")
	}
	if err := m.ensureCompany(ctx, companyID); err != nil {
		return nil, nil, err
	}
	if abs(amount) > m.GetSettings(ctx).ReviewThreshold {
		req := &entity.BillAdjustRequest{SysCompanyID: companyID, Type: typ, Amount: amount, Reason: reason, Status: "pending", RequestBy: op.Name, RequestByID: op.UserID}
		if err := m.db.WithContext(ctx).Create(req).Error; err != nil {
			return nil, nil, err
		}
		return nil, req, ErrNeedReview
	}
	t := &entity.BillTransaction{SysCompanyID: companyID, Type: typ, Amount: amount, Remark: reason, Operator: op.Name}
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return m.postTxn(ctx, tx, t, "") })
	return t, nil, err
}

// ReviewRequest 复核赠送 / 调账申请：复核人不能是申请人
func (m *Meter) ReviewRequest(ctx context.Context, op Operator, id uint, approve bool, remark string) (*entity.BillAdjustRequest, error) {
	var req entity.BillAdjustRequest
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, id).Error; err != nil {
			return errors.New("申请不存在")
		}
		if req.Status != "pending" {
			return errors.New("申请已处理")
		}
		if req.RequestByID == op.UserID {
			return errors.New("复核人不能是申请人，请由另一位平台管理员复核")
		}
		now := m.now()
		req.ReviewBy, req.ReviewTime, req.ReviewRemark = op.Name, &now, remark
		req.Status = "rejected"
		if approve {
			req.Status = "approved"
			t := &entity.BillTransaction{SysCompanyID: req.SysCompanyID, Type: req.Type, Amount: req.Amount,
				Remark: fmt.Sprintf("%s（%s 申请，%s 复核）", req.Reason, req.RequestBy, op.Name), Operator: op.Name,
				RefType: "adjust_request", RefID: req.ID}
			if err := m.postTxn(ctx, tx, t, ""); err != nil {
				return err
			}
			req.TxnID = t.ID
		}
		return tx.Save(&req).Error
	})
	return &req, err
}

// SetCredit 设置授信额度
func (m *Meter) SetCredit(ctx context.Context, companyID uint, limit int64) error {
	if limit < 0 {
		return errors.New("授信额度不能为负")
	}
	if err := m.ensureCompany(ctx, companyID); err != nil {
		return err
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockAccount(tx, companyID); err != nil {
			return err
		}
		return tx.Model(&entity.BillAccount{}).Where("SYS_COMPANY_ID = ?", companyID).Update("CREDIT_LIMIT", limit).Error
	})
}

// ensureCompany 公司必须存在
func (m *Meter) ensureCompany(ctx context.Context, companyID uint) error {
	var n int64
	m.db.WithContext(ctx).Table("sys_company").Where("ID = ?", companyID).Count(&n)
	if n == 0 {
		return errors.New("公司不存在")
	}
	return nil
}

// account 读取账户（不存在时为零值）
func (m *Meter) account(ctx context.Context, companyID uint) entity.BillAccount {
	acc := entity.BillAccount{SysCompanyID: companyID}
	m.db.WithContext(ctx).Where("SYS_COMPANY_ID = ?", companyID).First(&acc)
	return acc
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
