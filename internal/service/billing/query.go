package billing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

const monthLayout = "2006-01"

// monthRange 月份（YYYY-MM）的首日与次月首日；空为本月
func (m *Meter) monthRange(month string) (time.Time, time.Time, error) {
	if month == "" {
		month = m.now().Format(monthLayout)
	}
	d0, err := time.ParseInLocation(monthLayout, month, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("月份格式应为 YYYY-MM：%s", month)
	}
	return d0, d0.AddDate(0, 1, 0), nil
}

func round3(v float64) float64 { return math.Round(v*1000) / 1000 }

// UsageSummary 某月用量汇总
type UsageSummary struct {
	Month            string             `json:"month"`
	LiveMinutes      float64            `json:"liveMinutes"`
	PlayGB           float64            `json:"playGb"`
	StorageAvgGB     float64            `json:"storageAvgGb"`     // 当月已计量天数的日均存储
	StorageLatestGB  float64            `json:"storageLatestGb"`  // 最近一天的存储
	StorageLatestDay string             `json:"storageLatestDay"` // 最近一次存储快照的日期
	StorageItems     map[string]float64 `json:"storageItems"`     // 最近一天的存储细项：record / clip / disk
	FluxLatestDay    string             `json:"fluxLatestDay"`    // 播放流量已计到的日期（腾讯云次日出数据）
	LiveRooms        int                `json:"liveRooms"`        // 有直播时长的直播间数
	OnlineRooms      int                `json:"onlineRooms"`      // 正在推流的直播间数
}

type usageAgg struct {
	UsageDate string
	Meter     string
	Item      string
	RoomID    uint
	Quantity  float64
	Raw       float64
}

func (m *Meter) monthRows(ctx context.Context, companyID uint, month string) ([]usageAgg, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	var rows []usageAgg
	err = m.db.WithContext(ctx).Model(&entity.BillUsageDaily{}).
		Select("USAGE_DATE AS usage_date, METER AS meter, ITEM AS item, ROOM_ID AS room_id, QUANTITY AS quantity, RAW_QUANTITY AS raw").
		Where("SYS_COMPANY_ID = ? AND USAGE_DATE >= ? AND USAGE_DATE < ?", companyID, d0.Format(dayLayout), d1.Format(dayLayout)).
		Scan(&rows).Error
	return rows, err
}

// Summary 公司某月用量汇总
func (m *Meter) Summary(ctx context.Context, companyID uint, month string) (*UsageSummary, error) {
	d0, _, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	rows, err := m.monthRows(ctx, companyID, month)
	if err != nil {
		return nil, err
	}
	out := &UsageSummary{Month: d0.Format(monthLayout), StorageItems: map[string]float64{}}
	storageByDay := map[string]float64{}
	liveRooms := map[uint]bool{}
	for _, r := range rows {
		switch r.Meter {
		case entity.MeterLiveDuration:
			out.LiveMinutes += r.Quantity
			liveRooms[r.RoomID] = true
		case entity.MeterPlayFlux:
			out.PlayGB += r.Quantity
			if r.UsageDate > out.FluxLatestDay {
				out.FluxLatestDay = r.UsageDate
			}
		case entity.MeterStorage:
			storageByDay[r.UsageDate] += r.Quantity
			if r.UsageDate > out.StorageLatestDay {
				out.StorageLatestDay = r.UsageDate
			}
		}
	}
	if len(storageByDay) > 0 {
		total := 0.0
		for _, v := range storageByDay {
			total += v
		}
		out.StorageAvgGB = round3(total / float64(len(storageByDay)))
		out.StorageLatestGB = round3(storageByDay[out.StorageLatestDay])
		for _, r := range rows {
			if r.Meter == entity.MeterStorage && r.UsageDate == out.StorageLatestDay {
				out.StorageItems[r.Item] = round3(out.StorageItems[r.Item] + r.Quantity)
			}
		}
	}
	out.LiveRooms = len(liveRooms)
	out.PlayGB = round3(out.PlayGB)
	var online int64
	m.db.WithContext(ctx).Model(&entity.BillLiveSession{}).
		Where("SYS_COMPANY_ID = ? AND END_TIME IS NULL", companyID).
		Distinct("ROOM_ID").Count(&online)
	out.OnlineRooms = int(online)
	return out, nil
}

// DailyUsage 某天的三项用量
type DailyUsage struct {
	Date        string  `json:"date"`
	LiveMinutes float64 `json:"liveMinutes"`
	PlayGB      float64 `json:"playGb"`
	StorageGB   float64 `json:"storageGb"`
}

// Daily 公司某月每天的用量（当月每一天都返回，没有用量为 0）
func (m *Meter) Daily(ctx context.Context, companyID uint, month string) ([]DailyUsage, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	rows, err := m.monthRows(ctx, companyID, month)
	if err != nil {
		return nil, err
	}
	byDay := map[string]*DailyUsage{}
	var out []DailyUsage
	for d := d0; d.Before(d1); d = d.AddDate(0, 0, 1) {
		out = append(out, DailyUsage{Date: d.Format(dayLayout)})
	}
	for i := range out {
		byDay[out[i].Date] = &out[i]
	}
	for _, r := range rows {
		u := byDay[r.UsageDate]
		if u == nil {
			continue
		}
		switch r.Meter {
		case entity.MeterLiveDuration:
			u.LiveMinutes += r.Quantity
		case entity.MeterPlayFlux:
			u.PlayGB += r.Quantity
		case entity.MeterStorage:
			u.StorageGB += r.Quantity
		}
	}
	for i := range out {
		out[i].PlayGB = round3(out[i].PlayGB)
		out[i].StorageGB = round3(out[i].StorageGB)
	}
	return out, nil
}

// RoomUsage 直播间某月用量
type RoomUsage struct {
	RoomID      uint    `json:"roomId"`
	RoomName    string  `json:"roomName"`
	LiveMinutes float64 `json:"liveMinutes"`
	LiveSeconds float64 `json:"liveSeconds"`
	PlayGB      float64 `json:"playGb"`
	Sessions    int64   `json:"sessions"`
}

// Rooms 公司某月各直播间的直播时长和播放流量，按时长倒序
func (m *Meter) Rooms(ctx context.Context, companyID uint, month string) ([]RoomUsage, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	rows, err := m.monthRows(ctx, companyID, month)
	if err != nil {
		return nil, err
	}
	byRoom := map[uint]*RoomUsage{}
	get := func(id uint) *RoomUsage {
		if byRoom[id] == nil {
			byRoom[id] = &RoomUsage{RoomID: id}
		}
		return byRoom[id]
	}
	for _, r := range rows {
		switch r.Meter {
		case entity.MeterLiveDuration:
			u := get(r.RoomID)
			u.LiveMinutes += r.Quantity
			u.LiveSeconds += r.Raw
		case entity.MeterPlayFlux:
			get(r.RoomID).PlayGB += r.Quantity
		}
	}
	if len(byRoom) == 0 {
		return []RoomUsage{}, nil
	}
	ids := make([]uint, 0, len(byRoom))
	for id := range byRoom {
		ids = append(ids, id)
	}
	type nameRow struct {
		ID   uint   `gorm:"column:ID"`
		Name string `gorm:"column:ROOM_NAME"`
	}
	var names []nameRow
	m.db.WithContext(ctx).Table("live_room").Select("ID, ROOM_NAME").Where("ID IN ?", ids).Scan(&names)
	for _, n := range names {
		if u := byRoom[n.ID]; u != nil {
			u.RoomName = n.Name
		}
	}
	type cntRow struct {
		RoomID uint  `gorm:"column:ROOM_ID"`
		N      int64 `gorm:"column:N"`
	}
	var cnts []cntRow
	m.db.WithContext(ctx).Model(&entity.BillLiveSession{}).Select("ROOM_ID, COUNT(*) AS N").
		Where("SYS_COMPANY_ID = ? AND START_TIME >= ? AND START_TIME < ?", companyID, d0, d1).
		Group("ROOM_ID").Scan(&cnts)
	for _, c := range cnts {
		if u := byRoom[c.RoomID]; u != nil {
			u.Sessions = c.N
		}
	}
	out := make([]RoomUsage, 0, len(byRoom))
	for _, u := range byRoom {
		u.PlayGB = round3(u.PlayGB)
		out = append(out, *u)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].LiveMinutes != out[j].LiveMinutes {
			return out[i].LiveMinutes > out[j].LiveMinutes
		}
		if out[i].PlayGB != out[j].PlayGB {
			return out[i].PlayGB > out[j].PlayGB
		}
		return out[i].RoomID < out[j].RoomID
	})
	return out, nil
}

// SessionPage 推流段落分页
type SessionPage struct {
	Items []entity.BillLiveSession `json:"items"`
	Total int64                    `json:"total"`
}

// Sessions 公司某月的推流段落（可按直播间筛选），开始时间倒序
func (m *Meter) Sessions(ctx context.Context, companyID, roomID uint, month string, page, pageSize int) (*SessionPage, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	q := m.db.WithContext(ctx).Model(&entity.BillLiveSession{}).
		Where("SYS_COMPANY_ID = ? AND START_TIME < ? AND (END_TIME IS NULL OR END_TIME >= ?)", companyID, d1, d0)
	if roomID != 0 {
		q = q.Where("ROOM_ID = ?", roomID)
	}
	out := &SessionPage{Items: []entity.BillLiveSession{}}
	if err := q.Count(&out.Total).Error; err != nil {
		return nil, err
	}
	if err := q.Order("START_TIME DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&out.Items).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// CompanyUsage 平台对账：各公司某月用量
type CompanyUsage struct {
	CompanyID    uint    `json:"companyId"`
	CompanyName  string  `json:"companyName"`
	Account      string  `json:"account"` // global 平台账号 / tc:<密钥 ID 摘要> 自有账号
	LiveMinutes  float64 `json:"liveMinutes"`
	PlayGB       float64 `json:"playGb"`
	StorageAvgGB float64 `json:"storageAvgGb"`
}

// Reconcile 平台对账：各公司某月用量与各账号播放域名的流量对账
type Reconcile struct {
	Companies []CompanyUsage             `json:"companies"`
	Flux      []entity.BillFluxReconcile `json:"flux"`
}

// ReconcileMonth 平台查看某月全部公司的用量与流量对账
func (m *Meter) ReconcileMonth(ctx context.Context, month string) (*Reconcile, error) {
	d0, d1, err := m.monthRange(month)
	if err != nil {
		return nil, err
	}
	from, to := d0.Format(dayLayout), d1.Format(dayLayout)
	type agg struct {
		Company uint    `gorm:"column:SYS_COMPANY_ID"`
		Meter   string  `gorm:"column:METER"`
		Qty     float64 `gorm:"column:QTY"`
		Days    int     `gorm:"column:DAYS"`
	}
	var aggs []agg
	if err := m.db.WithContext(ctx).Model(&entity.BillUsageDaily{}).
		Select("SYS_COMPANY_ID, METER, SUM(QUANTITY) AS QTY, COUNT(DISTINCT USAGE_DATE) AS DAYS").
		Where("USAGE_DATE >= ? AND USAGE_DATE < ?", from, to).
		Group("SYS_COMPANY_ID, METER").Scan(&aggs).Error; err != nil {
		return nil, err
	}
	byCompany := map[uint]*CompanyUsage{}
	for _, a := range aggs {
		u := byCompany[a.Company]
		if u == nil {
			u = &CompanyUsage{CompanyID: a.Company}
			byCompany[a.Company] = u
		}
		switch a.Meter {
		case entity.MeterLiveDuration:
			u.LiveMinutes = a.Qty
		case entity.MeterPlayFlux:
			u.PlayGB = round3(a.Qty)
		case entity.MeterStorage:
			if a.Days > 0 {
				u.StorageAvgGB = round3(a.Qty / float64(a.Days))
			}
		}
	}
	out := &Reconcile{Companies: []CompanyUsage{}, Flux: []entity.BillFluxReconcile{}}
	if len(byCompany) > 0 {
		ids := make([]uint, 0, len(byCompany))
		for id := range byCompany {
			ids = append(ids, id)
		}
		type nameRow struct {
			ID   uint   `gorm:"column:ID"`
			Name string `gorm:"column:NAME"`
		}
		var names []nameRow
		m.db.WithContext(ctx).Table("sys_company").Select("ID, NAME").Where("ID IN ?", ids).Scan(&names)
		for _, n := range names {
			byCompany[n.ID].CompanyName = n.Name
		}
		for _, u := range byCompany {
			u.Account = m.accounts.Of(ctx, u.CompanyID)
			out.Companies = append(out.Companies, *u)
		}
		sort.Slice(out.Companies, func(i, j int) bool { return out.Companies[i].CompanyID < out.Companies[j].CompanyID })
	}
	if err := m.db.WithContext(ctx).Where("USAGE_DATE >= ? AND USAGE_DATE < ?", from, to).
		Order("USAGE_DATE DESC, ACCOUNT, PLAY_DOMAIN").Find(&out.Flux).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// maxRecomputeDays 一次重新计量的最多天数
const maxRecomputeDays = 31

// Recompute 平台重新计量一段日期；meters 为空时计算全部计费点。播放流量最多支持最近 3 个月（腾讯云保留期）
func (m *Meter) Recompute(ctx context.Context, from, to string, meters []string) error {
	d0, _, err := dayRange(from)
	if err != nil {
		return err
	}
	d1, _, err := dayRange(to)
	if err != nil {
		return err
	}
	if d1.Before(d0) {
		return errors.New("结束日期不能早于开始日期")
	}
	if d1.Sub(d0) > maxRecomputeDays*24*time.Hour {
		return fmt.Errorf("一次最多重新计量 %d 天", maxRecomputeDays)
	}
	if d1.After(m.now()) {
		return errors.New("不能计量未来的日期")
	}
	want := map[string]bool{}
	for _, mt := range meters {
		want[mt] = true
	}
	all := len(want) == 0
	if _, err := m.ProcessEvents(ctx); err != nil {
		return err
	}
	var errs []error
	for d := d0; !d.After(d1); d = d.AddDate(0, 0, 1) {
		day := d.Format(dayLayout)
		if all || want[entity.MeterLiveDuration] {
			errs = append(errs, m.ComputeDuration(ctx, day))
		}
		if all || want[entity.MeterStorage] {
			errs = append(errs, m.ComputeStorage(ctx, day))
		}
		if all || want[entity.MeterPlayFlux] {
			errs = append(errs, m.ComputeFlux(ctx, day))
		}
	}
	return errors.Join(errs...)
}
