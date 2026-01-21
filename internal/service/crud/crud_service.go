package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/executor"
	"github.com/sky-xhsoft/sky-server/internal/pkg/mask"
	"github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

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

// GetOne 查询单条记录
func (s *service) GetOne(ctx context.Context, tableName string, id uint, userID uint) (map[string]interface{}, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查读权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermRead)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无查询权限")
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 构建查询字段（根据MASK控制）
	selectFields, err := s.buildSelectFields(columns, userID, "edit")
	if err != nil {
		return nil, err
	}
	fmt.Printf("[DEBUG] GetOne - 查询字段: %s\n", selectFields)

	// 获取数据过滤条件
	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	// 构建查询
	query := s.db.Table(table.Name).Select(selectFields)

	// 添加ID条件
	query = query.Where("ID = ?", id)

	// 添加数据过滤条件
	if dataFilter != nil && len(dataFilter) > 0 {
		query = s.applyFilters(query, dataFilter, columns)
	}

	// 添加IS_ACTIVE条件
	query = query.Where("IS_ACTIVE = ?", "Y")

	// 执行查询
	var result map[string]interface{}
	if err := query.Take(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "记录不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
	}

	// 处理外键字段，将ID转换为显示名称
	results := []map[string]interface{}{result}
	if err := s.processForeignKeys(ctx, columns, results, userID); err != nil {
		// FK 转换失败不影响主查询，只记录错误
		fmt.Printf("[WARN] 处理外键字段失败: %v\n", err)
	}

	return result, nil
}

// GetList 查询列表
func (s *service) GetList(ctx context.Context, req *QueryRequest, userID uint) (*QueryResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取表元数据
	table, err := s.metadataService.GetTable(req.TableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查读权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermRead)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无查询权限")
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 构建查询字段（根据MASK控制）
	selectFields, err := s.buildSelectFields(columns, userID, "list")
	if err != nil {
		return nil, err
	}

	// 获取数据过滤条件
	dataFilter, err := s.groupsService.GetUserDataFilter(ctx, userID, table.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}

	// 构建查询
	query := s.db.Table(table.Name).Select(selectFields)

	// 添加数据过滤条件
	if dataFilter != nil && len(dataFilter) > 0 {
		query = s.applyFilters(query, dataFilter, columns)
	}

	// 添加IS_ACTIVE条件
	query = query.Where("IS_ACTIVE = ?", "Y")

	// 添加过滤条件
	if req.Filters != nil && len(req.Filters) > 0 {
		query = s.applyFilters(query, req.Filters, columns)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询总数失败", err)
	}

	// 添加排序
	if req.OrderBy != "" {
		order := "ASC"
		if strings.ToUpper(req.Order) == "DESC" {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", req.OrderBy, order))
	} else {
		query = query.Order("ID DESC")
	}

	// 添加分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Limit(req.PageSize).Offset(offset)

	// 执行查询
	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
	}

	// 处理外键字段，将ID转换为显示名称
	if err := s.processForeignKeys(ctx, columns, results, userID); err != nil {
		// FK 转换失败不影响主查询，只记录错误
		fmt.Printf("[WARN] 处理外键字段失败: %v\n", err)
	}

	return &QueryResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     results,
	}, nil
}

// Create 创建记录
func (s *service) Create(ctx context.Context, tableName string, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查创建权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermCreate)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无创建权限")
	}

	// 获取字段定义（在事务外，避免长时间持有锁）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 验证和处理字段（在事务外）
	processedData, err := s.processFieldsForCreate(columns, data, userID)
	if err != nil {
		return nil, err
	}

	// 生成新的ID（在事务外，避免长时间持有锁）
	newID, err := s.idgenService.GetNextID(ctx, table.Name)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "生成ID失败", err)
	}
	processedData["ID"] = newID

	// 添加审计字段
	processedData["IS_ACTIVE"] = "Y"

	// 获取用户信息以填充审计字段
	user, userErr := s.userRepo.GetUserByID(userID)
	if userErr == nil && user != nil {
		// 设置创建人和公司ID
		processedData["CREATE_BY"] = user.Username
		processedData["SYS_COMPANY_ID"] = user.SysCompanyID
		fmt.Printf("[DEBUG] 设置审计字段: CREATE_BY=%s, SYS_COMPANY_ID=%d\n", user.Username, user.SysCompanyID)
	} else {
		fmt.Printf("[DEBUG] 获取用户信息失败: userID=%d, err=%v\n", userID, userErr)
	}
	// 设置创建时间
	processedData["CREATE_TIME"] = time.Now()

	fmt.Printf("[DEBUG] 准备插入的数据: %+v\n", processedData)

	// 在事务中执行：before钩子 + 插入 + after钩子
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "begin", data); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 执行插入（在事务中，ID已经预先生成）
		if err := tx.Table(table.Name).Create(&processedData).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "创建失败", err)
		}

		// 执行after钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "A", "end", processedData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 获取创建记录的ID（已在前面生成）
	recordID := newID

	// 返回创建的记录
	if recordID == 0 {
		return processedData, nil
	}

	return s.GetOne(ctx, tableName, recordID, userID)
}

// Update 更新记录
func (s *service) Update(ctx context.Context, tableName string, id uint, data map[string]interface{}, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查更新权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermUpdate)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无修改权限")
	}

	// 添加ID到数据中供钩子使用
	data["ID"] = id

	// 获取字段定义（在事务外）
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return err
	}

	// 验证和处理字段（在事务外）
	processedData, err := s.processFieldsForUpdate(columns, data, userID)
	if err != nil {
		return err
	}

	// 添加审计字段
	// 获取用户信息以填充审计字段
	user, err := s.userRepo.GetUserByID(userID)
	if err == nil && user != nil {
		// 设置更新人
		processedData["UPDATE_BY"] = user.Username
	}
	// 设置更新时间
	processedData["UPDATE_TIME"] = time.Now()

	// 获取要更新的字段列表（支持零值更新）
	updateFields := make([]string, 0, len(processedData))
	for field := range processedData {
		updateFields = append(updateFields, field)
	}

	// 在事务中执行：before钩子 + 更新 + after钩子
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "begin", data); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 执行更新（在事务中，使用 Select 明确指定要更新的字段，包括零值）
		result := tx.Table(table.Name).
			Where("ID = ? AND IS_ACTIVE = ?", id, "Y").
			Select(updateFields).
			Updates(processedData)
		if result.Error != nil {
			return errors.Wrap(errors.ErrDatabase, "更新失败", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New(errors.ErrResourceNotFound, "记录不存在")
		}

		// 执行after钩子（在事务中）
		processedData["ID"] = id
		if err := s.executeHooksInTx(ctx, tx, table.ID, "M", "end", processedData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	return err
}

// Delete 删除记录（物理删除）
func (s *service) Delete(ctx context.Context, tableName string, id uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查删除权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermDelete)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无删除权限")
	}

	// 检查外键引用（REF_ON_DELETE 保护）
	if err := s.checkForeignKeyReferences(ctx, table.ID, id); err != nil {
		return err
	}

	// 在事务中执行：before钩子 + 删除 + after钩子
	deleteData := map[string]interface{}{"ID": id}
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 执行before钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
		}

		// 执行物理删除（在事务中）
		result := tx.Table(table.Name).Where("ID = ?", id).Delete(nil)
		if result.Error != nil {
			return errors.Wrap(errors.ErrDatabase, "删除失败", result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New(errors.ErrResourceNotFound, "记录不存在")
		}

		// 执行after钩子（在事务中）
		if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
			return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
		}

		return nil
	})

	return err
}

// BatchDelete 批量删除
func (s *service) BatchDelete(ctx context.Context, tableName string, ids []uint, userID uint) error {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 检查删除权限
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermDelete)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !hasPermission {
		return errors.New(errors.ErrPermissionDenied, "无删除权限")
	}

	// 检查每个ID的外键引用
	for _, id := range ids {
		if err := s.checkForeignKeyReferences(ctx, table.ID, id); err != nil {
			return errors.Wrap(errors.ErrInternal, fmt.Sprintf("ID=%d: %s", id, err.Error()), err)
		}
	}

	// 在事务中执行批量删除
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 对每个ID执行before钩子（在事务中）
		for _, id := range ids {
			deleteData := map[string]interface{}{"ID": id}
			if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "begin", deleteData); err != nil {
				return errors.Wrap(errors.ErrInternal, fmt.Sprintf("执行ID=%d的before钩子失败", id), err)
			}
		}

		// 执行批量物理删除（在事务中）
		result := tx.Table(table.Name).Where("ID IN ?", ids).Delete(nil)
		if result.Error != nil {
			return errors.Wrap(errors.ErrDatabase, "批量删除失败", result.Error)
		}

		// 对每个ID执行after钩子（在事务中）
		for _, id := range ids {
			deleteData := map[string]interface{}{"ID": id}
			if err := s.executeHooksInTx(ctx, tx, table.ID, "D", "end", deleteData); err != nil {
				return errors.Wrap(errors.ErrInternal, fmt.Sprintf("执行ID=%d的after钩子失败", id), err)
			}
		}

		return nil
	})

	return err
}

// buildSelectFields 构建查询字段列表（根据MASK控制）
func (s *service) buildSelectFields(columns []*entity.SysColumn, userID uint, operation string) (string, error) {
	var fields []string

	// 定义标准的系统审计字段（无论MASK如何配置都应包含）
	standardFields := map[string]bool{
		"ID":             true,
		"SYS_COMPANY_ID": true,
		"CREATE_BY":      true,
		"CREATE_TIME":    true,
		"UPDATE_BY":      true,
		"UPDATE_TIME":    true,
		"IS_ACTIVE":      true,
	}

	// 跟踪哪些系统字段已经被添加
	addedStandardFields := make(map[string]bool)

	for _, col := range columns {
		// TODO: 检查字段权限（基于SGRADE）- 需要集成groups权限服务
		// 暂时允许所有字段访问

		// 主键字段和系统审计字段始终包含，不受MASK限制
		isPrimaryKey := col.IsAK == "Y" || col.SetValueType == "pk"
		isStandardField := standardFields[col.DbName]

		if isStandardField {
			addedStandardFields[col.DbName] = true
		}

		if !isPrimaryKey && !isStandardField {
			// 对于业务字段，检查MASK可见性
			if col.Mask != "" {
				fieldMask := mask.ParseMask(col.Mask)
				if !fieldMask.IsVisible(operation) {
					continue
				}
			}
		}

		fields = append(fields, col.DbName)
	}

	// 确保所有系统审计字段都被包含（即使它们不在 sys_column 中定义）
	for fieldName := range standardFields {
		if !addedStandardFields[fieldName] {
			fields = append(fields, fieldName)
			fmt.Printf("[DEBUG] 强制添加系统字段: %s\n", fieldName)
		}
	}

	if len(fields) == 0 {
		return "*", nil
	}

	return strings.Join(fields, ", "), nil
}

// processFieldsForCreate 处理创建时的字段
func (s *service) processFieldsForCreate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	for _, col := range columns {
		// TODO: 检查字段权限 - 需要集成groups权限服务

		// 检查MASK可编辑性
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("add") {
				continue
			}
		}

		// 获取字段值
		value, exists := data[col.DbName]
		if exists {
			// 验证FK字段值
			if col.SetValueType == "fk" && col.RefTableID != nil && value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
		}
	}

	return processedData, nil
}

// processFieldsForUpdate 处理更新时的字段
func (s *service) processFieldsForUpdate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	for _, col := range columns {
		// TODO: 检查字段权限 - 需要集成groups权限服务

		// 检查MASK可编辑性
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("edit") {
				continue
			}
		}

		// 获取字段值
		value, exists := data[col.DbName]
		if exists {
			// 验证FK字段值
			if col.SetValueType == "fk" && col.RefTableID != nil && value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
		}
	}

	return processedData, nil
}

// applyDataFilter 应用数据过滤条件
func (s *service) applyDataFilter(query *gorm.DB, filterJSON string, columns []*entity.SysColumn) *gorm.DB {
	if filterJSON == "" {
		return query
	}

	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return query
	}

	return s.applyFilters(query, filter, columns)
}

// applyFilters 应用过滤条件
// 对于 text 类型字段使用 LIKE 模糊匹配，其他类型使用精确匹配
func (s *service) applyFilters(query *gorm.DB, filters map[string]interface{}, columns []*entity.SysColumn) *gorm.DB {
	// 构建字段类型映射，方便快速查找
	columnTypeMap := make(map[string]string)
	if columns != nil {
		for _, col := range columns {
			columnTypeMap[col.DbName] = col.DisplayType
		}
	}

	for field, value := range filters {
		// 跳过 nil 值
		if value == nil {
			continue
		}

		// 获取字段的显示类型
		displayType := columnTypeMap[field]

		// 对于 text 类型字段（text, textarea），使用 LIKE 模糊查询
		// 根据 DisplayType 字段定义：blank,button,hr,check,file,image,select,text,textarea,date,datetime,clob,xml,json
		if displayType == "text" || displayType == "textarea" || displayType == "clob" {
			// 转换为字符串并添加通配符
			if strValue, ok := value.(string); ok && strValue != "" {
				query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+strValue+"%")
			}
		} else {
			// 其他类型使用精确匹配（select, date, datetime, check 等）
			query = query.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}

	return query
}

// executeHooks 执行表命令钩子
func (s *service) executeHooks(ctx context.Context, tableID uint, action, event string, data map[string]interface{}) error {
	// 获取钩子列表
	hooks, err := s.metadataRepo.GetTableCmdsByAction(tableID, action, event)
	if err != nil {
		return err
	}

	// 按顺序执行钩子
	for _, hook := range hooks {
		if err := s.executeHook(ctx, hook, data, s.db); err != nil {
			return err
		}
	}

	return nil
}

// executeHooksInTx 在事务中执行钩子
func (s *service) executeHooksInTx(ctx context.Context, tx *gorm.DB, tableID uint, action, event string, data map[string]interface{}) error {
	// 获取钩子列表
	hooks, err := s.metadataRepo.GetTableCmdsByAction(tableID, action, event)
	if err != nil {
		return err
	}

	// 按顺序执行钩子（在事务中）
	for _, hook := range hooks {
		if err := s.executeHook(ctx, hook, data, tx); err != nil {
			return err
		}
	}

	return nil
}

// executeHook 执行单个钩子
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
	// 根据ContentType执行不同类型的钩子
	switch hook.ContentType {
	case "js", "py", "go", "bsh":
		return s.executeScriptHook(ctx, hook, data, db)
	case "url":
		return s.executeURLHook(ctx, hook, data)
	case "sp":
		return s.executeSPHook(ctx, hook, data, db)
	default:
		return nil
	}
}

// executeScriptHook 执行脚本钩子
func (s *service) executeScriptHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
	var scriptType executor.ScriptType
	switch hook.ContentType {
	case "js":
		scriptType = executor.ScriptTypeJavaScript
	case "py":
		scriptType = executor.ScriptTypePython
	case "go":
		scriptType = executor.ScriptTypeGo
	case "bsh":
		scriptType = executor.ScriptTypeBash
	}

	// 对于 Go 类型的钩子，需要将数据库连接传递给钩子函数
	params := make(map[string]interface{})
	for k, v := range data {
		params[k] = v
	}

	// 对于 Go 钩子，将数据库连接加入到参数中
	if hook.ContentType == "go" && db != nil {
		params["__db__"] = db
	}

	scriptExecutor := executor.NewScriptExecutor(scriptType, 5*time.Minute)
	result, err := scriptExecutor.Execute(ctx, hook.Content, params)
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("钩子执行失败: %s", result.Error)
	}

	return nil
}

// executeURLHook 执行URL钩子
func (s *service) executeURLHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error {
	var urlReq executor.URLRequest
	if err := json.Unmarshal([]byte(hook.Content), &urlReq); err != nil {
		return err
	}

	// 合并数据到参数
	if urlReq.Params == nil {
		urlReq.Params = make(map[string]interface{})
	}
	for k, v := range data {
		urlReq.Params[k] = v
	}

	urlExecutor := executor.NewURLExecutor(5 * time.Minute)
	resp, err := urlExecutor.Execute(ctx, &urlReq)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("URL钩子执行失败: %s", resp.Error)
	}

	return nil
}

// executeSPHook 执行存储过程钩子
func (s *service) executeSPHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}, db *gorm.DB) error {
	var spReq executor.SPRequest
	if err := json.Unmarshal([]byte(hook.Content), &spReq); err != nil {
		return err
	}

	// 合并数据到输入参数
	if spReq.InParams == nil {
		spReq.InParams = make(map[string]interface{})
	}
	for k, v := range data {
		spReq.InParams[k] = v
	}

	spExecutor := executor.NewSPExecutor(db)
	resp, err := spExecutor.Execute(ctx, &spReq)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("存储过程钩子执行失败: %s", resp.Error)
	}

	return nil
}

// processForeignKeys 处理结果集中的外键字段，将ID转换为显示名称
func (s *service) processForeignKeys(ctx context.Context, columns []*entity.SysColumn, results []map[string]interface{}, userID uint) error {
	if len(results) == 0 {
		return nil
	}

	// 找出所有FK字段
	fkColumns := make([]*entity.SysColumn, 0)
	for _, col := range columns {
		if col.SetValueType == "fk" && col.RefTableID != nil {
			fkColumns = append(fkColumns, col)
		}
	}

	if len(fkColumns) == 0 {
		return nil
	}

	// 为每个FK字段收集所有需要查询的ID
	for _, fkCol := range fkColumns {
		// 收集所有唯一的ID值
		idSet := make(map[interface{}]bool)
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				idSet[val] = true
			}
		}

		if len(idSet) == 0 {
			continue
		}

		// 批量查询这些ID对应的显示值
		idToLabelMap := make(map[interface{}]string)
		for id := range idSet {
			// 转换为字符串进行查询
			idStr := fmt.Sprintf("%v", id)
			label, err := s.metadataService.GetForeignKeyDisplayValue(*fkCol.RefTableID, idStr, fkCol.RefColumnID, userID)
			if err != nil {
				// 单个FK查询失败不影响其他，使用原始值
				idToLabelMap[id] = idStr
			} else {
				idToLabelMap[id] = label
			}
		}

		// 在结果集中添加 _display 字段
		displayFieldName := fkCol.DbName + "_display"
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				if label, ok := idToLabelMap[val]; ok {
					record[displayFieldName] = label
				}
			}
		}
	}

	return nil
}
// checkForeignKeyReferences 检查外键引用，防止删除被引用的记录
func (s *service) checkForeignKeyReferences(ctx context.Context, tableID uint, recordID uint) error {
	// 查询所有引用当前表的FK字段
	allTables, err := s.metadataRepo.GetAllTables()
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询表列表失败", err)
	}

	fmt.Printf("[DEBUG] 检查外键引用：tableID=%d, recordID=%d, 总表数=%d\n", tableID, recordID, len(allTables))

	for _, refTable := range allTables {
		// 跳过当前表自己（不检查自己引用自己）
		if refTable.ID == tableID {
			fmt.Printf("[DEBUG] 跳过当前表自己: %s (ID=%d)\n", refTable.Name, refTable.ID)
			continue
		}

		columns, err := s.metadataService.GetColumns(refTable.ID)
		if err != nil {
			fmt.Printf("[DEBUG] 获取表 %s (ID=%d) 的列失败: %v\n", refTable.Name, refTable.ID, err)
			continue
		}

		// 检查引用表是否有 IS_ACTIVE 字段
		hasIsActive := false
		for _, col := range columns {
			if col.DbName == "IS_ACTIVE" {
				hasIsActive = true
				break
			}
		}

		for _, col := range columns {
			// 检查是否是引用当前表的FK字段
			if col.SetValueType == "fk" && col.RefTableID != nil && *col.RefTableID == tableID {
				fmt.Printf("[DEBUG] 发现FK字段：表=%s, 字段=%s, RefTableID=%d, RefOnDelete=%s\n",
					refTable.Name, col.DbName, *col.RefTableID, col.RefOnDelete)

				// 检查 REF_ON_DELETE 策略
				if col.RefOnDelete == "noAction" || col.RefOnDelete == "" {
					// 检查是否有记录引用
					var count int64
					var query string
					if hasIsActive {
						query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND IS_ACTIVE = 'Y'", refTable.Name, col.DbName)
					} else {
						query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", refTable.Name, col.DbName)
					}

					fmt.Printf("[DEBUG] 执行查询: %s, recordID=%d\n", query, recordID)

					if err := s.db.Raw(query, recordID).Scan(&count).Error; err != nil {
						fmt.Printf("[DEBUG] 查询失败: %v\n", err)
						return errors.Wrap(errors.ErrDatabase, "检查外键引用失败", err)
					}

					fmt.Printf("[DEBUG] 查询结果: count=%d\n", count)

					if count > 0 {
						return errors.New(errors.ErrResourceConflict, fmt.Sprintf("无法删除：有 %d 条 %s 记录正在引用", count, refTable.DisplayName))
					}
				}
				// cascade 和 setNull 策略暂不实现，需要在事务中处理
			}
		}
	}

	fmt.Printf("[DEBUG] 外键引用检查通过\n")
	return nil
}

// validateForeignKeyValue 验证FK字段的值是否在关联表中存在
func (s *service) validateForeignKeyValue(refTableID uint, value interface{}) error {
	// 获取关联表信息
	refTable, err := s.metadataService.GetTableByID(refTableID)
	if err != nil {
		return fmt.Errorf("关联表不存在")
	}

	// 查询记录是否存在
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE ID = ? AND IS_ACTIVE = 'Y'", refTable.Name)
	if err := s.db.Raw(query, value).Scan(&count).Error; err != nil {
		return fmt.Errorf("验证失败: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("关联记录不存在或已失效")
	}

	return nil
}
