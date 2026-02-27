package live

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

// LiveRoomService 直播间服务接口
type LiveRoomService interface {
	// CreateRoom 创建直播间
	CreateRoom(ctx context.Context, room *entity.LiveRoom) error
	// UpdateRoom 更新直播间
	UpdateRoom(ctx context.Context, room *entity.LiveRoom) error
	// DeleteRoom 删除直播间（软删除）
	DeleteRoom(ctx context.Context, id uint) error
	// GetRoom 获取直播间详情
	GetRoom(ctx context.Context, id uint) (*entity.LiveRoom, error)
	// GetRoomByStreamName 通过流名称获取直播间
	GetRoomByStreamName(ctx context.Context, streamName string) (*entity.LiveRoom, error)
	// ListRooms 查询直播间列表
	ListRooms(ctx context.Context, filter *RoomFilter) ([]*entity.LiveRoom, int64, error)
	// StartLive 开始直播
	StartLive(ctx context.Context, id uint) error
	// EndLive 结束直播
	EndLive(ctx context.Context, id uint) error
	// UpdateViewerCount 更新观看人数
	UpdateViewerCount(ctx context.Context, id uint, count int) error
}

// RoomFilter 直播间查询过滤器
type RoomFilter struct {
	CompanyID      uint
	RoomType       string
	Status         string
	RoomStage      string
	ViewingMethod  string
	Keyword        string // 搜索关键词（直播间名称）
	StartTimeFrom  *time.Time
	StartTimeTo    *time.Time
	Page           int
	PageSize       int
}

type liveRoomService struct {
	db *gorm.DB
}

// NewLiveRoomService 创建直播间服务实例
func NewLiveRoomService(db *gorm.DB) LiveRoomService {
	return &liveRoomService{
		db: db,
	}
}

// CreateRoom 创建直播间
func (s *liveRoomService) CreateRoom(ctx context.Context, room *entity.LiveRoom) error {
	if room.RoomName == "" {
		return errors.New("直播间名称不能为空")
	}
	if room.RoomType == "" {
		return errors.New("直播间类型不能为空")
	}
	if room.BroadcastFormat == "" {
		return errors.New("播出形式不能为空")
	}
	if room.RoomStage == "" {
		return errors.New("直播间阶段不能为空")
	}

	// 设置默认值
	if room.Status == "" {
		room.Status = "draft"
	}
	if room.ViewingMethod == "" {
		room.ViewingMethod = "public"
	}
	if room.PlaybackMethod == "" {
		room.PlaybackMethod = "post_end"
	}
	if room.PlaybackValidity == "" {
		room.PlaybackValidity = "unlimited"
	}

	return s.db.WithContext(ctx).Create(room).Error
}

// UpdateRoom 更新直播间
func (s *liveRoomService) UpdateRoom(ctx context.Context, room *entity.LiveRoom) error {
	if room.ID == 0 {
		return errors.New("直播间ID不能为空")
	}

	// 检查直播间是否存在
	var existingRoom entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", room.ID, "Y").First(&existingRoom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("直播间不存在")
		}
		return err
	}

	// 如果直播间正在直播中，限制某些字段的修改
	if existingRoom.Status == "live" {
		room.RoomType = existingRoom.RoomType
		room.BroadcastFormat = existingRoom.BroadcastFormat
		room.StreamName = existingRoom.StreamName
	}

	return s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("ID = ?", room.ID).Updates(room).Error
}

// DeleteRoom 删除直播间（软删除）
func (s *liveRoomService) DeleteRoom(ctx context.Context, id uint) error {
	// 检查直播间是否存在
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("直播间不存在")
		}
		return err
	}

	// 如果直播间正在直播中，不允许删除
	if room.Status == "live" {
		return errors.New("直播间正在直播中，无法删除")
	}

	// 软删除
	return s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("ID = ?", id).Update("IS_ACTIVE", "N").Error
}

// GetRoom 获取直播间详情
func (s *liveRoomService) GetRoom(ctx context.Context, id uint) (*entity.LiveRoom, error) {
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("直播间不存在")
		}
		return nil, err
	}
	return &room, nil
}

// GetRoomByStreamName 通过流名称获取直播间
func (s *liveRoomService) GetRoomByStreamName(ctx context.Context, streamName string) (*entity.LiveRoom, error) {
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("STREAM_NAME = ? AND IS_ACTIVE = ?", streamName, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("直播间不存在")
		}
		return nil, err
	}
	return &room, nil
}

// ListRooms 查询直播间列表
func (s *liveRoomService) ListRooms(ctx context.Context, filter *RoomFilter) ([]*entity.LiveRoom, int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	if filter.CompanyID > 0 {
		query = query.Where("SYS_COMPANY_ID = ?", filter.CompanyID)
	}
	if filter.RoomType != "" {
		query = query.Where("ROOM_TYPE = ?", filter.RoomType)
	}
	if filter.Status != "" {
		query = query.Where("STATUS = ?", filter.Status)
	}
	if filter.RoomStage != "" {
		query = query.Where("ROOM_STAGE = ?", filter.RoomStage)
	}
	if filter.ViewingMethod != "" {
		query = query.Where("VIEWING_METHOD = ?", filter.ViewingMethod)
	}
	if filter.Keyword != "" {
		query = query.Where("ROOM_NAME LIKE ?", "%"+filter.Keyword+"%")
	}
	if filter.StartTimeFrom != nil {
		query = query.Where("START_TIME >= ?", filter.StartTimeFrom)
	}
	if filter.StartTimeTo != nil {
		query = query.Where("START_TIME <= ?", filter.StartTimeTo)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var rooms []*entity.LiveRoom
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("CREATE_TIME DESC").Offset(offset).Limit(filter.PageSize).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// StartLive 开始直播
func (s *liveRoomService) StartLive(ctx context.Context, id uint) error {
	// 检查直播间是否存在
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("直播间不存在")
		}
		return err
	}

	// 检查状态
	if room.Status == "live" {
		return errors.New("直播间已经在直播中")
	}
	if room.Status == "ended" || room.Status == "archived" {
		return errors.New("直播间已结束，无法重新开始")
	}

	// 更新状态为直播中
	now := time.Now()
	updates := map[string]interface{}{
		"STATUS":     "live",
		"START_TIME": now,
	}

	return s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("ID = ?", id).Updates(updates).Error
}

// EndLive 结束直播
func (s *liveRoomService) EndLive(ctx context.Context, id uint) error {
	// 检查直播间是否存在
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("直播间不存在")
		}
		return err
	}

	// 检查状态
	if room.Status != "live" {
		return fmt.Errorf("直播间当前状态为%s，无法结束", room.Status)
	}

	// 计算直播时长
	now := time.Now()
	var duration int
	if room.StartTime != nil {
		duration = int(now.Sub(*room.StartTime).Seconds())
	}

	// 更新状态为已结束
	updates := map[string]interface{}{
		"STATUS":   "ended",
		"END_TIME": now,
		"DURATION": duration,
	}

	return s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("ID = ?", id).Updates(updates).Error
}

// UpdateViewerCount 更新观看人数
func (s *liveRoomService) UpdateViewerCount(ctx context.Context, id uint, count int) error {
	// 获取当前直播间信息
	var room entity.LiveRoom
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("直播间不存在")
		}
		return err
	}

	// 更新观看人数和峰值
	updates := map[string]interface{}{
		"VIEWER_COUNT": count,
	}
	if count > room.PeakViewerCount {
		updates["PEAK_VIEWER_COUNT"] = count
	}

	return s.db.WithContext(ctx).Model(&entity.LiveRoom{}).Where("ID = ?", id).Updates(updates).Error
}
