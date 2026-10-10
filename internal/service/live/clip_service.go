package live

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/contextutil"
	"gorm.io/gorm"
)

// ClipService 直播切片服务接口
type ClipService interface {
	// ListClips 查询切片列表
	ListClips(ctx context.Context, filter *ClipFilter) ([]*entity.LiveCut, int64, error)
	// ListRecordings 查询录制列表
	ListRecordings(ctx context.Context, filter *RecordingFilter) ([]*entity.LiveRecode, int64, error)
}

// ClipFilter 切片查询过滤条件
type ClipFilter struct {
	RoomID   *int64
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

// RecordingFilter 录制查询过滤条件
type RecordingFilter struct {
	RoomID   *int64
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

type clipService struct {
	db *gorm.DB
}

// NewClipService 创建切片服务
func NewClipService(db *gorm.DB) ClipService {
	return &clipService{db: db}
}

// ListClips 查询切片列表
func (s *clipService) ListClips(ctx context.Context, filter *ClipFilter) ([]*entity.LiveCut, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	companyID, ok := ctx.Value(contextutil.CompanyIDKey).(uint)
	if !ok {
		companyID = 0
	}

	query := s.db.WithContext(ctx).Model(&entity.LiveCut{}).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y")

	if filter.RoomID != nil && *filter.RoomID > 0 {
		query = query.Where("ROOM_ID = ?", *filter.RoomID)
	}
	if filter.Status != "" {
		query = query.Where("STATUS = ?", filter.Status)
	}
	if filter.Keyword != "" {
		query = query.Where("CLIP_NAME LIKE ?", "%"+filter.Keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var cuts []*entity.LiveCut
	if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(filter.PageSize).Find(&cuts).Error; err != nil {
		return nil, 0, err
	}

	return cuts, total, nil
}

// ListRecordings 查询录制列表
func (s *clipService) ListRecordings(ctx context.Context, filter *RecordingFilter) ([]*entity.LiveRecode, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	companyID, ok := ctx.Value(contextutil.CompanyIDKey).(uint)
	if !ok {
		companyID = 0
	}

	query := s.db.WithContext(ctx).Model(&entity.LiveRecode{}).
		Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1)) AND IS_ACTIVE = ?", companyID, "Y")

	if filter.RoomID != nil && *filter.RoomID > 0 {
		query = query.Where("ROOM_ID = ?", *filter.RoomID)
	}
	if filter.Status != "" {
		query = query.Where("STATUS = ?", filter.Status)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		query = query.Where("(RECORD_URL LIKE ? OR DESCRIPTION LIKE ? OR ROOM_ID IN (SELECT ID FROM live_room WHERE ROOM_NAME LIKE ?))", kw, kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var records []*entity.LiveRecode
	if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(filter.PageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	s.fillRecordingRoomNames(ctx, records)

	return records, total, nil
}

// fillRecordingRoomNames 按 ROOM_ID 批量查出直播间名称填入列表（列表展示直播间名称而不是流名称）
func (s *clipService) fillRecordingRoomNames(ctx context.Context, records []*entity.LiveRecode) {
	ids := make([]int64, 0, len(records))
	for _, r := range records {
		if r.RoomID != nil {
			ids = append(ids, *r.RoomID)
		}
	}
	if len(ids) == 0 {
		return
	}
	var rooms []struct {
		ID       int64
		RoomName string
	}
	if err := s.db.WithContext(ctx).Table("live_room").Select("ID, ROOM_NAME AS room_name").Where("ID IN ?", ids).Scan(&rooms).Error; err != nil {
		return
	}
	names := make(map[int64]string, len(rooms))
	for _, r := range rooms {
		names[r.ID] = r.RoomName
	}
	for _, r := range records {
		if r.RoomID != nil {
			r.RoomName = names[*r.RoomID]
		}
	}
}
