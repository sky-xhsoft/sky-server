package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// Service Dashboard 聚合服务接口
type Service interface {
	// GetStats 获取 Dashboard 统计卡
	GetStats(ctx context.Context, companyID uint) (*Stats, error)

	// GetTrend 获取近 N 日趋势（默认 7 天）
	GetTrend(ctx context.Context, companyID uint, days int) (*Trend, error)

	// GetRecentRooms 获取最近直播
	GetRecentRooms(ctx context.Context, companyID uint, limit int) ([]*RecentRoom, error)

	// GetActivities 获取最新动态
	GetActivities(ctx context.Context, companyID uint, limit int) ([]*Activity, error)
}

// Stats 统计卡
type Stats struct {
	TotalRooms        int64 `json:"totalRooms"`        // 直播间总数
	TodayLiveSessions int64 `json:"todayLiveSessions"` // 今日直播场次
	RecordingStorage  int64 `json:"recordingStorage"`  // 直播录制使用空间（字节：live_recode + live_cut）
	UsedStorage       int64 `json:"usedStorage"`       // 云盘已用空间（字节）兼容保留
	GroupCount        int64 `json:"groupCount"`        // 权限组数 兼容保留
}

// Trend 趋势图数据
type Trend struct {
	Dates        []string `json:"dates"`        // 日期列表（YYYY-MM-DD）
	LiveSessions []int64  `json:"liveSessions"` // 每日直播场次
	Recordings   []int64  `json:"recordings"`   // 每日新增录制数
}

// RecentRoom 最近直播
type RecentRoom struct {
	ID          uint       `json:"id"`
	RoomName    string     `json:"roomName"`
	Status      string     `json:"status"`
	StartTime   *time.Time `json:"startTime"`
	ViewerCount int        `json:"viewerCount"`
}

// Activity 最新动态
type Activity struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	ResourceName string    `json:"resourceName"`
	CreatedAt    time.Time `json:"createdAt"`
}

// service Dashboard 服务实现
type service struct {
	db *gorm.DB
}

// NewService 创建 Dashboard 服务
func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

// GetStats 获取 Dashboard 统计卡
func (s *service) GetStats(ctx context.Context, companyID uint) (*Stats, error) {
	stats := &Stats{}

	// 直播间总数
	if err := s.db.WithContext(ctx).Model(&entity.LiveRoom{}).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y").
		Count(&stats.TotalRooms).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计直播间失败", err)
	}

	// 今日直播场次（START_TIME 落在今天）
	todayStart, todayEnd := dayRange(time.Now())
	if err := s.db.WithContext(ctx).Model(&entity.LiveRoom{}).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ? AND STATUS IN (?, ?, ?)", companyID, "Y", "live", "ended", "archived").
		Where("START_TIME >= ? AND START_TIME < ?", todayStart, todayEnd).
		Count(&stats.TodayLiveSessions).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计今日直播场次失败", err)
	}

	// 云盘已用空间：汇总公司下所有用户的配额
	if err := s.db.WithContext(ctx).Model(&entity.CloudQuota{}).
		Select("COALESCE(SUM(USED_SPACE), 0)").
		Joins("INNER JOIN sys_user ON sys_user.ID = cloud_quota.USER_ID").
		Where("sys_user.SYS_COMPANY_ID = ? AND sys_user.IS_ACTIVE = ?", companyID, "Y").
		Scan(&stats.UsedStorage).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计云盘用量失败", err)
	}

	// 权限组数
	if err := s.db.WithContext(ctx).Model(&entity.SysGroups{}).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y").
		Count(&stats.GroupCount).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计权限组失败", err)
	}

	// 直播录制使用空间：汇总录制(live_recode) + 剪辑(live_cut) 的文件大小
	for _, table := range []string{"live_recode", "live_cut"} {
		var size int64
		if err := s.db.WithContext(ctx).Table(table).
			Select("COALESCE(SUM(FILE_SIZE), 0)").
			Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ? AND FILE_SIZE > 0", companyID, "Y").
			Scan(&size).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "统计直播录制空间失败", err)
		}
		stats.RecordingStorage += size
	}

	return stats, nil
}

// GetTrend 获取近 N 日趋势
func (s *service) GetTrend(ctx context.Context, companyID uint, days int) (*Trend, error) {
	if days <= 0 {
		days = 7
	}

	dates := make([]string, days)
	liveSessions := make([]int64, days)
	recordings := make([]int64, days)

	now := time.Now()
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -(days - 1 - i))
		dates[i] = d.Format("2006-01-02")
	}

	startDate, _ := dayRange(now.AddDate(0, 0, -(days - 1)))
	endDate, _ := dayRange(now.AddDate(0, 0, 1))

	// 每日直播场次
	liveMap, err := s.countByDate(ctx, "live_room", "START_TIME", companyID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 每日录制事件
	recordingMap, err := s.countByDate(ctx, "live_callback_event", "CREATE_TIME", int64(companyID), startDate, endDate)
	if err != nil {
		return nil, err
	}

	for i, d := range dates {
		liveSessions[i] = liveMap[d]
		recordings[i] = recordingMap[d]
	}

	return &Trend{
		Dates:        dates,
		LiveSessions: liveSessions,
		Recordings:   recordings,
	}, nil
}

// countByDate 按日期统计数量（MySQL DATE_FORMAT）
func (s *service) countByDate(ctx context.Context, table, timeColumn string, companyID interface{}, start, end time.Time) (map[string]int64, error) {
	result := make(map[string]int64)

	var rows []struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}

	companyColumn := "SYS_COMPANY_ID"
	if table == "live_callback_event" {
		companyColumn = "SYS_COMPANY_ID"
	}

	query := fmt.Sprintf(
		"SELECT DATE_FORMAT(%s, '%%Y-%%m-%%d') AS date, COUNT(*) AS count FROM %s WHERE (%s = ? OR %s IN (0, -1)) AND %s >= ? AND %s < ?",
		timeColumn, table, companyColumn, companyColumn, timeColumn, timeColumn,
	)
	if table == "live_callback_event" {
		query += " AND EVENT_TYPE = 'recording_file'"
	}
	query += " GROUP BY date"

	if err := s.db.WithContext(ctx).Raw(query, companyID, start, end).Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计趋势失败", err)
	}

	for _, row := range rows {
		result[row.Date] = row.Count
	}

	return result, nil
}

// GetRecentRooms 获取最近直播
func (s *service) GetRecentRooms(ctx context.Context, companyID uint, limit int) ([]*RecentRoom, error) {
	if limit <= 0 {
		limit = 5
	}

	var rooms []*entity.LiveRoom
	if err := s.db.WithContext(ctx).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y").
		Order("CREATE_TIME DESC").
		Limit(limit).
		Find(&rooms).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最近直播失败", err)
	}

	result := make([]*RecentRoom, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, &RecentRoom{
			ID:          room.ID,
			RoomName:    room.RoomName,
			Status:      room.Status,
			StartTime:   room.StartTime,
			ViewerCount: room.ViewerCount,
		})
	}

	return result, nil
}

// GetActivities 获取最新动态
func (s *service) GetActivities(ctx context.Context, companyID uint, limit int) ([]*Activity, error) {
	if limit <= 0 {
		limit = 10
	}

	var logs []*entity.AuditLog
	if err := s.db.WithContext(ctx).
		Where("SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)", companyID).
		Order("CREATED_AT DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最新动态失败", err)
	}

	result := make([]*Activity, 0, len(logs))
	for _, log := range logs {
		result = append(result, &Activity{
			ID:           log.ID,
			Username:     log.Username,
			Action:       log.Action,
			Resource:     log.Resource,
			ResourceName: log.ResourceName,
			CreatedAt:    log.CreatedAt,
		})
	}

	return result, nil
}

// dayRange 返回某一天的开始和结束时间
func dayRange(t time.Time) (start, end time.Time) {
	start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end = start.AddDate(0, 0, 1)
	return
}
