package dashboard

import (
	"context"
	"sort"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// SystemOverview 系统中心概览（按当前公司统计）
type SystemOverview struct {
	Users     int64          `json:"users"`
	Groups    int64          `json:"groups"`
	Tables    int64          `json:"tables"`
	Companies int64          `json:"companies"`
	Trend     SystemTrend    `json:"trend"`
	RecentOps []*OperationOp `json:"recentOps"`
}

// SystemTrend 近 N 日登录与操作次数
type SystemTrend struct {
	Dates  []string `json:"dates"`
	Logins []int64  `json:"logins"`
	Ops    []int64  `json:"ops"`
}

// OperationOp 最近操作（来自审计日志）
type OperationOp struct {
	ID       uint      `gorm:"column:id" json:"id"`
	Username string    `gorm:"column:username" json:"user"`
	Action   string    `gorm:"column:action" json:"action"`
	Target   string    `gorm:"column:target" json:"target"`
	Time     time.Time `gorm:"column:time" json:"time"`
}

// CloudOverview 云盘概览（按当前用户自己的网盘统计，避免暴露他人文件）
type CloudOverview struct {
	UsedStorage int64            `json:"usedStorage"`
	FileCount   int64            `json:"fileCount"`
	FolderCount int64            `json:"folderCount"`
	ShareCount  int64            `json:"shareCount"`
	Trend       CloudTrend       `json:"trend"`
	RecentFiles []*CloudFileItem `json:"recentFiles"`
	Activities  []*CloudActivity `json:"activities"`
}

// CloudTrend 近 N 日新增文件与新增条目（文件 + 文件夹）
type CloudTrend struct {
	Dates   []string `json:"dates"`
	Uploads []int64  `json:"uploads"`
	Files   []int64  `json:"files"`
}

// CloudFileItem 最近文件
type CloudFileItem struct {
	ID         uint      `gorm:"column:id" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	FileType   string    `gorm:"column:file_type" json:"fileType"`
	FileExt    string    `gorm:"column:file_ext" json:"fileExt"`
	Size       int64     `gorm:"column:size" json:"size"`
	CreateTime time.Time `gorm:"column:create_time" json:"time"`
}

// CloudActivity 最近动态：uploaded / createdFolder / createdShare
type CloudActivity struct {
	Action string    `json:"action"`
	Target string    `json:"target"`
	Time   time.Time `json:"time"`
}

// trendDates 近 days 天（含今天）的日期序列
func trendDates(now time.Time, days int) []string {
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		dates[i] = now.AddDate(0, 0, -(days - 1 - i)).Format("2006-01-02")
	}
	return dates
}

// bucketByDay 将时间点按日期归入 dates 对应的计数（在 Go 中分桶，与数据库方言无关）
func bucketByDay(dates []string, times []time.Time) []int64 {
	index := make(map[string]int, len(dates))
	for i, d := range dates {
		index[d] = i
	}
	counts := make([]int64, len(dates))
	for _, t := range times {
		if i, ok := index[t.Local().Format("2006-01-02")]; ok {
			counts[i]++
		}
	}
	return counts
}

// GetSystemOverview 系统中心概览
func (s *service) GetSystemOverview(ctx context.Context, companyID uint, days int) (*SystemOverview, error) {
	if days <= 0 {
		days = 7
	}
	db := s.db.WithContext(ctx)
	ov := &SystemOverview{RecentOps: []*OperationOp{}}

	counts := []struct {
		dst   *int64
		table string
		where string
		args  []interface{}
	}{
		{&ov.Users, "sys_user", "SYS_COMPANY_ID = ? AND IS_ACTIVE = 'Y'", []interface{}{companyID}},
		{&ov.Groups, "sys_groups", "(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = 'Y'", []interface{}{companyID}},
		{&ov.Tables, "sys_table", "IS_ACTIVE = 'Y'", nil},
		{&ov.Companies, "sys_company", "IS_ACTIVE = 'Y'", nil},
	}
	for _, c := range counts {
		if err := db.Table(c.table).Where(c.where, c.args...).Count(c.dst).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "统计 "+c.table+" 失败", err)
		}
	}

	// 审计日志只属于本公司用户（按用户所属公司过滤，避免看到其他公司的操作）
	now := time.Now()
	start, _ := dayRange(now.AddDate(0, 0, -(days - 1)))
	companyAudit := func() *gorm.DB {
		return db.Table("audit_log").
			Joins("JOIN sys_user u ON u.ID = audit_log.USER_ID").
			Where("u.SYS_COMPANY_ID = ?", companyID)
	}

	var events []struct {
		Action    string    `gorm:"column:ACTION"`
		CreatedAt time.Time `gorm:"column:CREATED_AT"`
	}
	if err := companyAudit().Select("audit_log.ACTION, audit_log.CREATED_AT").
		Where("audit_log.CREATED_AT >= ?", start).Scan(&events).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计操作趋势失败", err)
	}
	var logins, ops []time.Time
	for _, e := range events {
		switch e.Action {
		case "login":
			logins = append(logins, e.CreatedAt)
		case "logout":
		default:
			ops = append(ops, e.CreatedAt)
		}
	}
	ov.Trend.Dates = trendDates(now, days)
	ov.Trend.Logins = bucketByDay(ov.Trend.Dates, logins)
	ov.Trend.Ops = bucketByDay(ov.Trend.Dates, ops)

	if err := companyAudit().
		Select("audit_log.ID AS id, audit_log.USERNAME AS username, audit_log.ACTION AS action, audit_log.PATH AS target, audit_log.CREATED_AT AS time").
		Where("audit_log.ACTION NOT IN ('login', 'logout')").
		Order("audit_log.CREATED_AT DESC").Limit(5).Scan(&ov.RecentOps).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最近操作失败", err)
	}
	return ov, nil
}

// GetCloudOverview 云盘概览（当前用户）
func (s *service) GetCloudOverview(ctx context.Context, userID uint, days int) (*CloudOverview, error) {
	if days <= 0 {
		days = 7
	}
	db := s.db.WithContext(ctx)
	ov := &CloudOverview{RecentFiles: []*CloudFileItem{}, Activities: []*CloudActivity{}}
	mine := func() *gorm.DB {
		return db.Table("cloud_item").Where("OWNER_ID = ? AND IS_ACTIVE = 'Y'", userID)
	}

	if err := mine().Where("ITEM_TYPE = 'file'").Select("COALESCE(SUM(FILE_SIZE), 0)").Scan(&ov.UsedStorage).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计云盘用量失败", err)
	}
	if err := mine().Where("ITEM_TYPE = 'file'").Count(&ov.FileCount).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计文件数失败", err)
	}
	if err := mine().Where("ITEM_TYPE = 'folder'").Count(&ov.FolderCount).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计文件夹数失败", err)
	}
	if err := db.Table("cloud_share").Where("SHARER_ID = ? AND STATUS = 'active' AND IS_ACTIVE = 'Y'", userID).
		Count(&ov.ShareCount).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计分享数失败", err)
	}

	now := time.Now()
	start, _ := dayRange(now.AddDate(0, 0, -(days - 1)))
	var created []struct {
		ItemType   string    `gorm:"column:ITEM_TYPE"`
		CreateTime time.Time `gorm:"column:CREATE_TIME"`
	}
	if err := mine().Select("ITEM_TYPE, CREATE_TIME").Where("CREATE_TIME >= ?", start).Scan(&created).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "统计云盘趋势失败", err)
	}
	var fileTimes, itemTimes []time.Time
	for _, c := range created {
		itemTimes = append(itemTimes, c.CreateTime)
		if c.ItemType == "file" {
			fileTimes = append(fileTimes, c.CreateTime)
		}
	}
	ov.Trend.Dates = trendDates(now, days)
	ov.Trend.Uploads = bucketByDay(ov.Trend.Dates, fileTimes)
	ov.Trend.Files = bucketByDay(ov.Trend.Dates, itemTimes)

	if err := mine().Where("ITEM_TYPE = 'file'").
		Select("ID AS id, NAME AS name, IFNULL(FILE_TYPE, '') AS file_type, IFNULL(FILE_EXT, '') AS file_ext, IFNULL(FILE_SIZE, 0) AS size, CREATE_TIME AS create_time").
		Order("CREATE_TIME DESC").Limit(4).Scan(&ov.RecentFiles).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最近文件失败", err)
	}

	// 最近动态：合并最近的上传、新建文件夹、新建分享，按时间倒序取 4 条
	var items []struct {
		ItemType   string    `gorm:"column:ITEM_TYPE"`
		Name       string    `gorm:"column:NAME"`
		CreateTime time.Time `gorm:"column:CREATE_TIME"`
	}
	if err := mine().Select("ITEM_TYPE, NAME, CREATE_TIME").Order("CREATE_TIME DESC").Limit(4).Scan(&items).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最近动态失败", err)
	}
	for _, it := range items {
		action := "uploaded"
		if it.ItemType == "folder" {
			action = "createdFolder"
		}
		ov.Activities = append(ov.Activities, &CloudActivity{Action: action, Target: it.Name, Time: it.CreateTime})
	}
	var shares []struct {
		Name       string    `gorm:"column:NAME"`
		CreateTime time.Time `gorm:"column:CREATE_TIME"`
	}
	if err := db.Table("cloud_share s").
		Select("IFNULL(i.NAME, s.SHARE_CODE) AS NAME, s.CREATE_TIME AS CREATE_TIME").
		Joins("LEFT JOIN cloud_item i ON i.ID = s.RESOURCE_ID").
		Where("s.SHARER_ID = ? AND s.IS_ACTIVE = 'Y'", userID).
		Order("s.CREATE_TIME DESC").Limit(4).Scan(&shares).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询最近分享失败", err)
	}
	for _, sh := range shares {
		ov.Activities = append(ov.Activities, &CloudActivity{Action: "createdShare", Target: sh.Name, Time: sh.CreateTime})
	}
	sort.SliceStable(ov.Activities, func(i, j int) bool { return ov.Activities[i].Time.After(ov.Activities[j].Time) })
	if len(ov.Activities) > 4 {
		ov.Activities = ov.Activities[:4]
	}
	return ov, nil
}
