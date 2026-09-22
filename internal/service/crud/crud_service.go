package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/sky-xhsoft/sky-server/internal/domain/search"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/dict"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
	"github.com/xuri/excelize/v2"
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

	// 高级搜索（支持多条件组合、多种运算符）
	Search(ctx context.Context, req *search.SearchRequest, tableName string, userID uint) (*QueryResponse, error)

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

	// 作废记录
	Void(ctx context.Context, tableName string, id uint, userID uint) error

	// 导出到Excel（支持过滤条件，nil=全部）
	Export(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) (*excelize.File, error)

	// 从Excel导入
	Import(ctx context.Context, tableName string, file multipart.File, userID uint) (int, error)

	// 生成导入模板Excel
	ImportTemplate(ctx context.Context, tableName string, userID uint) (*excelize.File, error)
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
	dictService     dict.Service
	sequenceService sequence.Service
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

		// 只更新传入的字段（前端只传变更字段）
		updateFields := make([]string, 0, len(mainRecord))
		for field := range mainRecord {
			updateFields = append(updateFields, field)
		}

		if err := tx.Table(req.TableName).Where("id = ?", mainRecordID).Select(updateFields).Updates(&mainRecord).Error; err != nil {
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

		if req.Mode == "update" {
			// 收集前端传入的已有记录 ID
			incomingIDs := make(map[uint]bool)
			for _, rec := range detail.Records {
				if id, ok := rec["ID"]; ok {
					if uid, ok := toUint(id); ok && uid > 0 {
						incomingIDs[uid] = true
					}
				}
			}

			// 删除数据库中不在传入列表中旧记录
			if len(incomingIDs) > 0 {
				var existingIDs []uint
				if err := tx.Table(detail.TableName).Where(detail.RefField+" = ?", mainRecordID).Pluck("ID", &existingIDs).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("查询已有子表记录失败: %s", err.Error())
				}
				for _, eid := range existingIDs {
					if !incomingIDs[eid] {
						if err := tx.Table(detail.TableName).Where("ID = ?", eid).Delete(nil).Error; err != nil {
							tx.Rollback()
							return nil, fmt.Errorf("删除子表记录失败: %s", err.Error())
						}
					}
				}
			} else {
				// 没有传入任何已有ID，全部删除
				if err := tx.Table(detail.TableName).Where(detail.RefField+" = ?", mainRecordID).Delete(nil).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}

		savedRecords := make([]map[string]interface{}, 0, len(detail.Records))
		for _, record := range detail.Records {
			record[detail.RefField] = mainRecordID

			if req.Mode == "update" && hasID(record) {
				// 已有记录：更新
				rid := toUintVal(record["ID"])
				delete(record, "ID")
				processed, err := s.processFieldsForUpdate(detailColumns, record, userID)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				// 只更新传入的字段
				updateFields := make([]string, 0, len(processed))
				for field := range processed {
					updateFields = append(updateFields, field)
				}
				if err := tx.Table(detail.TableName).Where("ID = ?", rid).Select(updateFields).Updates(&processed).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
				// 取回最新数据
				var updated map[string]interface{}
				if err := tx.Table(detail.TableName).Where("ID = ?", rid).Take(&updated).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
				savedRecords = append(savedRecords, updated)
			} else {
				// 新记录：插入
				processed, err := s.processFieldsForCreate(detailColumns, record, userID)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				if err := tx.Table(detail.TableName).Create(&processed).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
				savedRecords = append(savedRecords, processed)
			}
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

func hasID(record map[string]interface{}) bool {
	id, ok := record["ID"]
	if !ok {
		return false
	}
	_, ok = toUint(id)
	return ok
}

func toUintVal(v interface{}) uint {
	uid, _ := toUint(v)
	return uid
}

func toUint(v interface{}) (uint, bool) {
	switch val := v.(type) {
	case uint:
		return val, val > 0
	case uint64:
		return uint(val), val > 0
	case uint32:
		return uint(val), val > 0
	case int:
		return uint(val), val > 0
	case int64:
		return uint(val), val > 0
	case int32:
		return uint(val), val > 0
	case float64:
		return uint(val), val > 0
	case float32:
		return uint(val), val > 0
	case string:
		n, err := strconv.ParseUint(val, 10, 64)
		return uint(n), err == nil && n > 0
	case json.Number:
		n, err := val.Int64()
		return uint(n), err == nil && n > 0
	default:
		s := fmt.Sprintf("%v", val)
		n, err := strconv.ParseUint(s, 10, 64)
		return uint(n), err == nil && n > 0
	}
}

// NewService 创建通用CRUD服务
func NewService(
	db *gorm.DB,
	metadataService metadata.Service,
	groupsService groups.Service,
	dictService dict.Service,
	sequenceService sequence.Service,
	metadataRepo repository.MetadataRepository,
	userRepo repository.UserRepository,
	idgenService idgen.Service,
) Service {
	return &service{
		db:              db,
		metadataService: metadataService,
		groupsService:   groupsService,
		dictService:     dictService,
		sequenceService: sequenceService,
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
