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

const (
	dayLayout = "2006-01-02"
	mbPerGB   = 1024.0
	gbBytes   = 1024.0 * 1024 * 1024
	// fluxDiffRate 监控流量与计费流量差异超过该比例时，对账标记为 diff
	fluxDiffRate = 0.03
	// fluxDiffMinMB 计费流量低于该值（MB）时不标记差异：几 KB 的零星流量不值得人工核对
	fluxDiffMinMB = 1.0
)

// dayRange 自然日（本地时区）的起止
func dayRange(day string) (time.Time, time.Time, error) {
	d0, err := time.ParseInLocation(dayLayout, day, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("日期格式应为 YYYY-MM-DD：%s", day)
	}
	return d0, d0.AddDate(0, 0, 1), nil
}

func round6(v float64) float64 { return math.Round(v*1e6) / 1e6 }

type usageKey struct {
	company uint
	room    uint
	item    string
}

// replaceDaily 用新结果替换某天某计费点的用量；companies 非空时只替换这些公司（其余公司本次未计量，保留原值）
func (m *Meter) replaceDaily(ctx context.Context, day, meter string, companies []uint, rows []entity.BillUsageDaily) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Where("USAGE_DATE = ? AND METER = ?", day, meter)
		if companies != nil {
			if len(companies) == 0 {
				return nil
			}
			q = q.Where("SYS_COMPANY_ID IN ?", companies)
		}
		if err := q.Delete(&entity.BillUsageDaily{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.CreateInBatches(rows, 200).Error
	})
}

func usageRow(day, meter, unit string, k usageKey, qty, raw float64) entity.BillUsageDaily {
	return entity.BillUsageDaily{
		SysCompanyID: k.company,
		UsageDate:    day,
		Meter:        meter,
		Item:         k.item,
		RoomID:       k.room,
		Quantity:     round6(qty),
		Unit:         unit,
		RawQuantity:  round6(raw),
		SourceKey:    fmt.Sprintf("%s:%d:%d:%s:%s", meter, k.company, k.room, k.item, day),
	}
}

// ComputeDuration 某天各直播间的直播时长：每段推流与当天重叠的部分向上取整到分钟；
// 未结束的段落算到最后在线时间；已结束且总时长不足 MinSessionSeconds 的不计
func (m *Meter) ComputeDuration(ctx context.Context, day string) error {
	d0, d1, err := dayRange(day)
	if err != nil {
		return err
	}
	var sessions []entity.BillLiveSession
	if err := m.db.WithContext(ctx).
		Where("START_TIME < ? AND (END_TIME IS NULL OR END_TIME > ?)", d1, d0).
		Find(&sessions).Error; err != nil {
		return err
	}
	minutes := map[usageKey]float64{}
	seconds := map[usageKey]float64{}
	for _, s := range sessions {
		end := s.LastSeenTime
		if s.EndTime != nil {
			end = *s.EndTime
			if s.Seconds < m.MinSessionSeconds {
				continue
			}
		}
		from, to := s.StartTime, end
		if from.Before(d0) {
			from = d0
		}
		if to.After(d1) {
			to = d1
		}
		piece := to.Sub(from).Seconds()
		if piece <= 0 {
			continue
		}
		k := usageKey{company: s.SysCompanyID, room: s.RoomID}
		minutes[k] += math.Ceil(piece / 60)
		seconds[k] += piece
	}
	rows := make([]entity.BillUsageDaily, 0, len(minutes))
	for k, v := range minutes {
		rows = append(rows, usageRow(day, entity.MeterLiveDuration, "分钟", k, v, seconds[k]))
	}
	return m.replaceDaily(ctx, day, entity.MeterLiveDuration, nil, rows)
}

// ComputeStorage 存储快照：当天结束前创建、目前仍有效的录制、切片、云盘文件，按公司汇总
// 删除时间未记录，补算历史日期时为近似值；每日任务在次日凌晨执行，与当天末的实际用量一致
func (m *Meter) ComputeStorage(ctx context.Context, day string) error {
	_, d1, err := dayRange(day)
	if err != nil {
		return err
	}
	type sum struct {
		Company uint
		Bytes   float64
	}
	sources := []struct{ item, table, extra string }{
		{"record", "live_recode", ""},
		{"clip", "live_cut", ""},
		{"disk", "cloud_item", " AND ITEM_TYPE = 'file'"},
	}
	var rows []entity.BillUsageDaily
	for _, src := range sources {
		var sums []sum
		if err := m.db.WithContext(ctx).Table(src.table).
			Select("SYS_COMPANY_ID AS company, COALESCE(SUM(FILE_SIZE), 0) AS bytes").
			Where("IS_ACTIVE = 'Y' AND CREATE_TIME < ? AND SYS_COMPANY_ID > 0"+src.extra, d1).
			Group("SYS_COMPANY_ID").Scan(&sums).Error; err != nil {
			return err
		}
		for _, s := range sums {
			if s.Bytes <= 0 {
				continue
			}
			k := usageKey{company: s.Company, item: src.item}
			rows = append(rows, usageRow(day, entity.MeterStorage, "GB", k, s.Bytes/gbBytes, s.Bytes))
		}
	}
	return m.replaceDaily(ctx, day, entity.MeterStorage, nil, rows)
}

// ComputeFlux 某天各直播间的播放流量。腾讯云按流数据是监控数据，不能直接计费：
// 以每条流的监控流量为比例，分摊该播放域名当天的计费流量，直播间之和与腾讯云账单一致。
// 某个账号拉取失败时，该账号下的公司保留原值，其余账号照常写入
func (m *Meter) ComputeFlux(ctx context.Context, day string) error {
	if _, _, err := dayRange(day); err != nil {
		return err
	}
	type room struct {
		ID      uint `gorm:"column:ID"`
		Company uint `gorm:"column:SYS_COMPANY_ID"`
	}
	var rooms []room
	if err := m.db.WithContext(ctx).Table("live_room").Select("ID, SYS_COMPANY_ID").Where("SYS_COMPANY_ID > 0").Scan(&rooms).Error; err != nil {
		return err
	}
	roomCompany := map[uint]uint{}
	accountOf := map[uint]string{}
	companiesOf := map[string][]uint{}
	for _, r := range rooms {
		roomCompany[r.ID] = r.Company
		if _, ok := accountOf[r.Company]; !ok {
			a := m.accounts.Of(ctx, r.Company)
			accountOf[r.Company] = a
			companiesOf[a] = append(companiesOf[a], r.Company)
		}
	}

	var errs []error
	for account, companies := range companiesOf {
		rows, err := m.fluxOfAccount(ctx, day, account, roomCompany, accountOf)
		if err != nil {
			zap.L().Warn("计量：拉取播放流量失败", zap.String("account", account), zap.String("day", day), zap.Error(err))
			errs = append(errs, fmt.Errorf("%s：%w", account, err))
			continue
		}
		if err := m.replaceDaily(ctx, day, entity.MeterPlayFlux, companies, rows); err != nil {
			return err
		}
	}
	return errors.Join(errs...)
}

func (m *Meter) fluxOfAccount(ctx context.Context, day, account string, roomCompany map[uint]uint, accountOf map[uint]string) ([]entity.BillUsageDaily, error) {
	api, err := m.accounts.API(ctx, account)
	if err != nil {
		return nil, err
	}
	domains, err := api.PlayDomains(ctx)
	if err != nil {
		return nil, err
	}
	billed := map[usageKey]float64{}
	monitor := map[usageKey]float64{}
	for _, domain := range domains {
		rec := entity.BillFluxReconcile{UsageDate: day, Account: account, PlayDomain: domain, Status: "ok"}
		streams, err := api.DayStreamFlux(ctx, domain, day)
		if err == nil {
			rec.BillingMB, err = api.DomainBillingFlux(ctx, domain, day)
		}
		if err != nil {
			rec.Status, rec.Message = "error", truncate(err.Error(), 500)
			m.saveReconcile(ctx, &rec)
			return nil, fmt.Errorf("播放域名 %s：%w", domain, err)
		}
		for _, mb := range streams {
			rec.MonitorMB += mb
		}
		rec.StreamCount = len(streams)
		factor := 0.0
		if rec.MonitorMB > 0 {
			factor = rec.BillingMB / rec.MonitorMB
		}
		for stream, mb := range streams {
			roomID := roomIDOfStream(stream)
			company := roomCompany[roomID]
			if company == 0 || accountOf[company] != account {
				rec.UnattributedMB += mb * factor
				continue
			}
			k := usageKey{company: company, room: roomID}
			billed[k] += mb * factor
			monitor[k] += mb
		}
		switch {
		case rec.BillingMB < fluxDiffMinMB:
		case rec.MonitorMB == 0:
			rec.Status, rec.Message = "diff", "有计费流量但没有按流数据，无法分摊到直播间"
		case rec.BillingMB > 0 && math.Abs(rec.BillingMB-rec.MonitorMB)/rec.BillingMB > fluxDiffRate:
			rec.Status, rec.Message = "diff", fmt.Sprintf("按流监控流量与计费流量相差 %.1f%%", math.Abs(rec.BillingMB-rec.MonitorMB)/rec.BillingMB*100)
		}
		m.saveReconcile(ctx, &rec)
	}
	rows := make([]entity.BillUsageDaily, 0, len(billed))
	for k, mb := range billed {
		rows = append(rows, usageRow(day, entity.MeterPlayFlux, "GB", k, mb/mbPerGB, monitor[k]/mbPerGB))
	}
	return rows, nil
}

func (m *Meter) saveReconcile(ctx context.Context, rec *entity.BillFluxReconcile) {
	rec.MonitorMB = math.Round(rec.MonitorMB*1000) / 1000
	rec.BillingMB = math.Round(rec.BillingMB*1000) / 1000
	rec.UnattributedMB = math.Round(rec.UnattributedMB*1000) / 1000
	if err := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "USAGE_DATE"}, {Name: "ACCOUNT"}, {Name: "PLAY_DOMAIN"}},
		DoUpdates: clause.AssignmentColumns([]string{"MONITOR_MB", "BILLING_MB", "UNATTRIBUTED_MB", "STREAM_COUNT", "STATUS", "MESSAGE", "UPDATE_TIME"}),
	}).Create(rec).Error; err != nil {
		zap.L().Warn("计量：保存流量对账失败", zap.Error(err))
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// RunDaily 计算某天的全部计费点
func (m *Meter) RunDaily(ctx context.Context, day string) error {
	return errors.Join(
		m.ComputeDuration(ctx, day),
		m.ComputeStorage(ctx, day),
		m.ComputeFlux(ctx, day),
	)
}
