package crud

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"gorm.io/gorm"
)

// SaveWithDetailsRequest 同时保存主表和明细请求（service层）
type SaveWithDetailsRequest struct {
	TableName    string                 `json:"tableName"`
	MainRecord   map[string]interface{} `json:"mainRecord"`
	Details      []DetailTableRequest   `json:"details"`
	Mode         string                 `json:"mode"`
	MainRecordID uint                   `json:"mainRecordId"`
}

// DetailTableRequest 子表请求数据
type DetailTableRequest struct {
	TableName string                   `json:"tableName"`
	Records   []map[string]interface{} `json:"records"`
	AssoType  string                   `json:"assoType"`
	RefField  string                   `json:"refField"`
}

// SaveWithDetailsResponse 保存响应
type SaveWithDetailsResponse struct {
	MainRecord map[string]interface{} `json:"mainRecord"`
	Details    []DetailTableResponse  `json:"details"`
}

// DetailTableResponse 子表响应数据
type DetailTableResponse struct {
	TableName string                   `json:"tableName"`
	Records   []map[string]interface{} `json:"records"`
}

// Service 通用CRUD服务接口
type Service interface {
	// 查询单条记录
	GetOne(ctx context.Context, tableName string, id uint, userID uint) (map[string]interface{}, error)

	// 查询列表（支持分页、排序、过滤）
	GetList(ctx context.Context, req *QueryRequest, userID uint) (*QueryResponse, error)

	// 创建记录
	Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) (map[string]interface{}, error)

	// 更新记录
	Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error

	// 删除记录（软删除）
	Delete(ctx context.Context, tableName string, id uint, userID uint) error

	// 批量删除
	BatchDelete(ctx context.Context, tableName string, ids []uint, userID uint) error

	// 提交记录
	Submit(ctx context.Context, tableName string, id uint, userID uint) error

	// 反提交记录
	Unsubmit(ctx context.Context, tableName string, id uint, userID uint) error

	// 同时保存主表和明细
	SaveWithDetails(ctx context.Context, req *SaveWithDetailsRequest, userID uint) (*SaveWithDetailsResponse, error)
}

// QueryRequest 查询请求
type QueryRequest struct {
	TableName string                 `json:"tableName" binding:"required"`
	Page      int                    `json:"page"`     // 页码，从1开始
	PageSize  int                    `json:"pageSize"` // 每页大小
	OrderBy   string                 `json:"orderBy"`  // 排序字段
	Order     string                 `json:"order"`    // 排序方向: asc, desc
	Filters   map[string]interface{} `json:"filters"`  // 过滤条件
	Include   []string               `json:"include"`  // 包含的关联表
}

// QueryResponse 查询响应
type QueryResponse struct {
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
	Data     []map[string]interface{} `json:"data"`
}

// service 通用CRUD服务实现
type service struct {
	db              *gorm.DB
	metadataService metadata.Service
	groupsService   groups.Service
	metadataRepo    repository.MetadataRepository
	userRepo        repository.UserRepository
	idgenService    idgen.Service
}

// SaveWithDetails 同时保存主表和明细
func (s *service) SaveWithDetails(ctx context.Context, req *SaveWithDetailsRequest, userID uint) (*SaveWithDetailsResponse, error) {
	// 开启事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var mainRecordID uint
	mainRecord := req.MainRecord

	// 1. 保存主表
	if req.Mode == "create" {
		// 新增主表
		// 生成新的ID（和基础CRUD逻辑保持一致）
		newID, err := s.idgenService.GetNextID(ctx, req.TableName)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		mainRecord["ID"] = newID
		mainRecordID = newID

		// 填充系统字段
		mainRecord["CREATE_BY"] = userID
		mainRecord["UPDATE_BY"] = userID
		mainRecord["IS_ACTIVE"] = "Y"

		if err := tx.Table(req.TableName).Create(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if req.Mode == "update" {
		// 更新主表
		mainRecordID = req.MainRecordID
		mainRecord["UPDATE_BY"] = userID

		if err := tx.Table(req.TableName).Where("id = ?", mainRecordID).Updates(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		// 查询最新的主表数据
		if err := tx.Table(req.TableName).Where("id = ?", mainRecordID).Take(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		tx.Rollback()
		return nil, gorm.ErrInvalidTransaction
	}

	// 2. 保存所有子表
	detailResponses := make([]DetailTableResponse, len(req.Details))
	for i, detail := range req.Details {
		// 先删除旧的子表数据（更新模式下）
		if req.Mode == "update" {
			if err := tx.Table(detail.TableName).Where(detail.RefField+" = ?", mainRecordID).Delete(nil).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		// 处理子表记录，设置主表关联ID
		savedRecords := make([]map[string]interface{}, len(detail.Records))
		for j, record := range detail.Records {
			// 设置关联字段值为主表ID
			record[detail.RefField] = mainRecordID
			// 填充系统字段
			record["CREATE_BY"] = userID
			record["UPDATE_BY"] = userID
			record["IS_ACTIVE"] = "Y"

			// 保存子表记录
			if err := tx.Table(detail.TableName).Create(&record).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			savedRecords[j] = record
		}

		detailResponses[i] = DetailTableResponse{
			TableName: detail.TableName,
			Records:   savedRecords,
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 返回结果
	return &SaveWithDetailsResponse{
		MainRecord: mainRecord,
		Details:    detailResponses,
	}, nil
}

// NewService 创建通用CRUD服务
func NewService(
	db *gorm.DB,
	metadataService metadata.Service,
	groupsService groups.Service,
	metadataRepo repository.MetadataRepository,
	userRepo repository.UserRepository,
	idgenService idgen.Service,
) Service {
	return &service{
		db:              db,
		metadataService: metadataService,
		groupsService:   groupsService,
		metadataRepo:    metadataRepo,
		userRepo:        userRepo,
		idgenService:    idgenService,
	}
}
