package billing

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	cursorKey     = "session_event_cursor" // 已处理到的回调事件 ID
	eventBatch    = 500
	offlineGrace  = 10 * time.Minute // 巡检连续这么久不在线、又没收到断流回调时，按最后在线时间结束
	sessionMaxAge = 24 * time.Hour   // 无法巡检（接口失败）时，单段最长按 24 小时结束
)

// Meter 计量服务
type Meter struct {
	db       *gorm.DB
	accounts Accounts
	now      func() time.Time
	// MinSessionSeconds 单段推流不足该秒数不计时长（测试推流）
	MinSessionSeconds int64
}

// NewMeter 计量服务
func NewMeter(db *gorm.DB, accounts Accounts) *Meter {
	return &Meter{db: db, accounts: accounts, now: time.Now, MinSessionSeconds: 60}
}

func (m *Meter) getState(ctx context.Context, key string) string {
	var s entity.BillMeterState
	if err := m.db.WithContext(ctx).Where("STATE_KEY = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.StateValue
}

func (m *Meter) setState(ctx context.Context, key, value string) error {
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&entity.BillMeterState{StateKey: key, StateValue: value}).Error
}

// ProcessEvents 把新的推流 / 断流回调整理成推流段落，返回处理的事件数
func (m *Meter) ProcessEvents(ctx context.Context) (int, error) {
	cursor, _ := strconv.ParseInt(m.getState(ctx, cursorKey), 10, 64)
	total := 0
	for {
		var events []entity.LiveCallbackEvent
		if err := m.db.WithContext(ctx).
			Where("ID > ? AND EVENT_TYPE IN ?", cursor, []string{"push_stream", "disconnect_stream"}).
			Order("ID").Limit(eventBatch).Find(&events).Error; err != nil {
			return total, err
		}
		if len(events) == 0 {
			return total, nil
		}
		for i := range events {
			e := &events[i]
			var err error
			if e.EventType == "push_stream" {
				err = m.onPush(ctx, e)
			} else {
				err = m.onDisconnect(ctx, e)
			}
			if err != nil {
				return total, err
			}
			cursor = e.ID
			total++
		}
		if err := m.setState(ctx, cursorKey, strconv.FormatInt(cursor, 10)); err != nil {
			return total, err
		}
	}
}

// eventStream 事件对应的流名与直播间；不是直播间的流（无法归属）返回 0
func eventStream(e *entity.LiveCallbackEvent) (string, uint) {
	stream := e.StreamName
	if stream == "" {
		stream = e.StreamID
	}
	return stream, roomIDOfStream(stream)
}

func (m *Meter) onPush(ctx context.Context, e *entity.LiveCallbackEvent) error {
	stream, roomID := eventStream(e)
	if roomID == 0 || e.SysCompanyID <= 0 || e.EventTime <= 0 {
		return nil
	}
	start := time.Unix(e.EventTime, 0)
	db := m.db.WithContext(ctx)
	// 同一条流上一段没收到断流就又推流（重连）：上一段在本次开始时结束
	var prev []entity.BillLiveSession
	if err := db.Where("SYS_COMPANY_ID = ? AND STREAM_NAME = ? AND END_TIME IS NULL AND START_TIME <= ? AND PUSH_EVENT_ID <> ?", e.SysCompanyID, stream, start, e.ID).
		Find(&prev).Error; err != nil {
		return err
	}
	for i := range prev {
		if err := closeSession(db, &prev[i], start, "callback"); err != nil {
			return err
		}
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.BillLiveSession{
		SysCompanyID: uint(e.SysCompanyID),
		RoomID:       roomID,
		StreamName:   stream,
		PushDomain:   e.DomainName,
		StartTime:    start,
		LastSeenTime: start,
		PushEventID:  e.ID,
	}).Error
}

// disconnectDuration 断流回调里的推流时长（秒）
func disconnectDuration(e *entity.LiveCallbackEvent) int64 {
	var d struct {
		Duration int64 `json:"duration"`
	}
	_ = json.Unmarshal([]byte(e.EventData), &d)
	return d.Duration
}

func (m *Meter) onDisconnect(ctx context.Context, e *entity.LiveCallbackEvent) error {
	stream, roomID := eventStream(e)
	if roomID == 0 || e.SysCompanyID <= 0 || e.EventTime <= 0 {
		return nil
	}
	end := time.Unix(e.EventTime, 0)
	db := m.db.WithContext(ctx)
	var s entity.BillLiveSession
	err := db.Where("SYS_COMPANY_ID = ? AND STREAM_NAME = ? AND END_TIME IS NULL AND START_TIME <= ?", e.SysCompanyID, stream, end).
		Order("START_TIME DESC").First(&s).Error
	if err == nil {
		return closeSession(db, &s, end, "callback")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 没有对应的推流段落（漏了推流回调）：用断流回调里的推流时长补一段
	d := disconnectDuration(e)
	if d <= 0 {
		return nil
	}
	start := end.Add(-time.Duration(d) * time.Second)
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.BillLiveSession{
		SysCompanyID: uint(e.SysCompanyID),
		RoomID:       roomID,
		StreamName:   stream,
		PushDomain:   e.DomainName,
		StartTime:    start,
		EndTime:      &end,
		LastSeenTime: end,
		Seconds:      d,
		EndSource:    "callback",
		PushEventID:  -e.ID, // 断流事件补出的段落，取负数避免与推流事件 ID 冲突
	}).Error
}

func closeSession(db *gorm.DB, s *entity.BillLiveSession, end time.Time, source string) error {
	if end.Before(s.StartTime) {
		end = s.StartTime
	}
	secs := int64(end.Sub(s.StartTime) / time.Second)
	return db.Model(&entity.BillLiveSession{}).Where("ID = ? AND END_TIME IS NULL", s.ID).
		Updates(map[string]any{"END_TIME": end, "SECONDS": secs, "END_SOURCE": source}).Error
}

// Patrol 巡检未结束的推流段落：仍在推流的更新最后在线时间；
// 已不在线且超过 offlineGrace 没收到断流回调的，按最后在线时间结束
func (m *Meter) Patrol(ctx context.Context) error {
	db := m.db.WithContext(ctx)
	var open []entity.BillLiveSession
	if err := db.Where("END_TIME IS NULL").Find(&open).Error; err != nil {
		return err
	}
	if len(open) == 0 {
		return nil
	}
	now := m.now()
	online := map[string]map[string]bool{} // 账号 → 在线流
	failed := map[string]bool{}
	for i := range open {
		s := &open[i]
		account := m.accounts.Of(ctx, s.SysCompanyID)
		if _, ok := online[account]; !ok && !failed[account] {
			set, err := m.onlineOf(ctx, account)
			if err != nil {
				zap.L().Warn("计量巡检：查询在线流失败", zap.String("account", account), zap.Error(err))
				failed[account] = true
			} else {
				online[account] = set
			}
		}
		switch {
		case failed[account]:
			if now.Sub(s.StartTime) > sessionMaxAge {
				if err := closeSession(db, s, s.StartTime.Add(sessionMaxAge), "cap"); err != nil {
					return err
				}
			}
		case online[account][s.StreamName]:
			if err := db.Model(&entity.BillLiveSession{}).Where("ID = ?", s.ID).Update("LAST_SEEN_TIME", now).Error; err != nil {
				return err
			}
		case now.Sub(s.LastSeenTime) >= offlineGrace:
			if err := closeSession(db, s, s.LastSeenTime, "patrol"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Meter) onlineOf(ctx context.Context, account string) (map[string]bool, error) {
	api, err := m.accounts.API(ctx, account)
	if err != nil {
		return nil, err
	}
	return api.OnlineStreams(ctx)
}
