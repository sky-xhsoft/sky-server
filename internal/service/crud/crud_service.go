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
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
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
	Page      int                    `json:"page"`      // 页码，从1开始
	PageSize  int                    `json:"pageSize"`  // 每页大小
	OrderBy   string                 `json:"orderBy"`   // 排序字段
	Order     string                 `json:"order"`     // 排序方向: asc, desc
	Filters   map[string]interface{} `json:"filters"`   // 过滤条件
	Include   []string               `json:"include"`   // 包含的关联表
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
}

// NewService 创建通用CRUD服务
func NewService(
	db *gorm.DB,
	metadataService metadata.Service,
	groupsService groups.Service,
	metadataRepo repository.MetadataRepository,
) Service {
	return &service{
		db:              db,
		metadataService: metadataService,
		groupsService:   groupsService,
		metadataRepo:    metadataRepo,
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
		query = s.applyFilters(query, dataFilter)
	}

	// 添加IS_ACTIVE条件
	query = query.Where("IS_ACTIVE = ?", "Y")

	// 执行查询
	var result map[string]interface{}
	if err := query.First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "记录不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询失败", err)
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
		query = s.applyFilters(query, dataFilter)
	}

	// 添加IS_ACTIVE条件
	query = query.Where("IS_ACTIVE = ?", "Y")

	// 添加过滤条件
	if req.Filters != nil && len(req.Filters) > 0 {
		query = s.applyFilters(query, req.Filters)
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

	// 执行before钩子
	if err := s.executeHooks(ctx, table.ID, "A", "begin", data); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 验证和处理字段（根据MASK和权限）
	processedData, err := s.processFieldsForCreate(columns, data, userID)
	if err != nil {
		return nil, err
	}

	// 添加审计字段
	// TODO: 从context获取用户名
	processedData["IS_ACTIVE"] = "Y"

	// 执行插入
	if err := s.db.Table(table.Name).Create(&processedData).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建失败", err)
	}

	// 执行after钩子
	if err := s.executeHooks(ctx, table.ID, "A", "end", processedData); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
	}

	// 返回创建的记录
	id := processedData["ID"]
	if id == nil {
		return processedData, nil
	}

	return s.GetOne(ctx, tableName, id.(uint), userID)
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

	// 执行before钩子
	if err := s.executeHooks(ctx, table.ID, "M", "begin", data); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return err
	}

	// 验证和处理字段（根据MASK和权限）
	processedData, err := s.processFieldsForUpdate(columns, data, userID)
	if err != nil {
		return err
	}

	// 添加审计字段
	// TODO: 从context获取用户名

	// 执行更新
	result := s.db.Table(table.Name).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").Updates(processedData)
	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "更新失败", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrResourceNotFound, "记录不存在")
	}

	// 执行after钩子
	processedData["ID"] = id
	if err := s.executeHooks(ctx, table.ID, "M", "end", processedData); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
	}

	return nil
}

// Delete 删除记录（软删除）
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

	// 执行before钩子
	deleteData := map[string]interface{}{"ID": id}
	if err := s.executeHooks(ctx, table.ID, "D", "begin", deleteData); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
	}

	// 执行软删除
	result := s.db.Table(table.Name).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").Update("IS_ACTIVE", "N")
	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "删除失败", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrResourceNotFound, "记录不存在")
	}

	// 执行after钩子
	if err := s.executeHooks(ctx, table.ID, "D", "end", deleteData); err != nil {
		return errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
	}

	return nil
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

	// 执行批量软删除
	result := s.db.Table(table.Name).Where("ID IN ? AND IS_ACTIVE = ?", ids, "Y").Update("IS_ACTIVE", "N")
	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "批量删除失败", result.Error)
	}

	return nil
}

// buildSelectFields 构建查询字段列表（根据MASK控制）
func (s *service) buildSelectFields(columns []*entity.SysColumn, userID uint, operation string) (string, error) {
	var fields []string

	for _, col := range columns {
		// TODO: 检查字段权限（基于SGRADE）- 需要集成groups权限服务
		// 暂时允许所有字段访问

		// 检查MASK可见性
		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsVisible(operation) {
				continue
			}
		}

		fields = append(fields, col.DbName)
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
			processedData[col.DbName] = value
		}
	}

	return processedData, nil
}

// applyDataFilter 应用数据过滤条件
func (s *service) applyDataFilter(query *gorm.DB, filterJSON string) *gorm.DB {
	if filterJSON == "" {
		return query
	}

	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return query
	}

	return s.applyFilters(query, filter)
}

// applyFilters 应用过滤条件
func (s *service) applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for field, value := range filters {
		// 简单的等值过滤
		// TODO: 支持更复杂的过滤操作符（like, in, between, etc.）
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
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
		if err := s.executeHook(ctx, hook, data); err != nil {
			return err
		}
	}

	return nil
}

// executeHook 执行单个钩子
func (s *service) executeHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error {
	// 根据ContentType执行不同类型的钩子
	switch hook.ContentType {
	case "js", "py", "go", "bsh":
		return s.executeScriptHook(ctx, hook, data)
	case "url":
		return s.executeURLHook(ctx, hook, data)
	case "sp":
		return s.executeSPHook(ctx, hook, data)
	default:
		return nil
	}
}

// executeScriptHook 执行脚本钩子
func (s *service) executeScriptHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error {
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

	scriptExecutor := executor.NewScriptExecutor(scriptType, 5*time.Minute)
	result, err := scriptExecutor.Execute(ctx, hook.Content, data)
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
func (s *service) executeSPHook(ctx context.Context, hook *entity.SysTableCmd, data map[string]interface{}) error {
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

	spExecutor := executor.NewSPExecutor(s.db)
	resp, err := spExecutor.Execute(ctx, &spReq)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("存储过程钩子执行失败: %s", resp.Error)
	}

	return nil
}
