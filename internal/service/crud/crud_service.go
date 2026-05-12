package crud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
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
	mainTable, err := s.metadataService.GetTable(req.TableName)
	if err != nil {
		return nil, fmt.Errorf("表不存在: %s", req.TableName)
	}
	mainColumns, err := s.metadataService.GetColumns(mainTable.ID)
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %s", err.Error())
	}

	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var mainRecordID uint
	mainRecord := req.MainRecord

	if req.Mode == "create" {
		newID, err := s.idgenService.GetNextID(ctx, req.TableName)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		mainRecord["ID"] = newID
		mainRecordID = newID

		processed, err := s.processFieldsForCreate(mainColumns, mainRecord, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		mainRecord = processed

		if err := tx.Table(req.TableName).Create(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if req.Mode == "update" {
		mainRecordID = req.MainRecordID

		processed, err := s.processFieldsForUpdate(mainColumns, mainRecord, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		mainRecord = processed

		if err := tx.Table(req.TableName).Where("id = ?", mainRecordID).Updates(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Table(req.TableName).Where("id = ?", mainRecordID).Take(&mainRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		tx.Rollback()
		return nil, gorm.ErrInvalidTransaction
	}

	detailResponses := make([]DetailTableResponse, len(req.Details))
	for i, detail := range req.Details {
		if req.Mode == "update" {
			if err := tx.Table(detail.TableName).Where(detail.RefField+" = ?", mainRecordID).Delete(nil).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		detailTable, err := s.metadataService.GetTable(detail.TableName)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("子表不存在: %s", detail.TableName)
		}
		detailColumns, err := s.metadataService.GetColumns(detailTable.ID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("获取子表列信息失败: %s", err.Error())
		}

		savedRecords := make([]map[string]interface{}, len(detail.Records))
		for j, record := range detail.Records {
			record[detail.RefField] = mainRecordID

			processed, err := s.processFieldsForCreate(detailColumns, record, userID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			if err := tx.Table(detail.TableName).Create(&processed).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			savedRecords[j] = processed
		}

		detailResponses[i] = DetailTableResponse{
			TableName: detail.TableName,
			Records:   savedRecords,
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

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

// resolveTableName 解析虚拟表映射到真实物理表
func resolveTableName(table *entity.SysTable) string {
	if table.RealTableID != nil && *table.RealTableID > 0 {
		return fmt.Sprintf("table_%d", *table.RealTableID)
	}
	return table.Name
}
