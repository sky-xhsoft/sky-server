package action

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/executor"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"gorm.io/gorm"
)

// Service 动作执行服务接口
type Service interface {
	// 执行动作
	ExecuteAction(ctx context.Context, actionID uint, params map[string]interface{}, userID uint) (*ActionResult, error)

	// 根据名称执行动作
	ExecuteActionByName(ctx context.Context, tableName, actionName string, params map[string]interface{}, userID uint) (*ActionResult, error)

	// 批量执行动作
	BatchExecuteAction(ctx context.Context, actionID uint, batchParams []map[string]interface{}, userID uint) ([]*ActionResult, error)

	// 获取动作定义
	GetAction(ctx context.Context, actionID uint) (*entity.SysAction, error)
}

// ActionResult 动作执行结果
type ActionResult struct {
	Success  bool                   `json:"success"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data"`
	Duration time.Duration          `json:"duration"`
	Error    string                 `json:"error"`
}

// service 动作执行服务实现
type service struct {
	db              *gorm.DB
	metadataService metadata.Service
	groupsService   groups.Service
	urlExecutor     *executor.URLExecutor
	spExecutor      *executor.SPExecutor
	scriptTimeout   time.Duration
}

// NewService 创建动作执行服务
func NewService(
	db *gorm.DB,
	metadataService metadata.Service,
	groupsService groups.Service,
	scriptTimeout int,
) Service {
	return &service{
		db:              db,
		metadataService: metadataService,
		groupsService:   groupsService,
		urlExecutor:     executor.NewURLExecutor(time.Duration(scriptTimeout) * time.Second),
		spExecutor:      executor.NewSPExecutor(db),
		scriptTimeout:   time.Duration(scriptTimeout) * time.Second,
	}
}

// ExecuteAction 执行动作
func (s *service) ExecuteAction(ctx context.Context, actionID uint, params map[string]interface{}, userID uint) (*ActionResult, error) {
	start := time.Now()

	// 获取动作定义
	action, err := s.GetAction(ctx, actionID)
	if err != nil {
		return &ActionResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
		}, nil
	}

	// 检查权限（如果有关联表）
	if action.SysTableID > 0 {
		// 动作执行需要更新权限（write权限）
		hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, uint(action.SysTableID), groups.PermUpdate)
		if err != nil {
			return &ActionResult{
				Success:  false,
				Error:    "权限检查失败: " + err.Error(),
				Duration: time.Since(start),
			}, nil
		}
		if !hasPermission {
			return &ActionResult{
				Success:  false,
				Error:    "无权限执行此动作",
				Duration: time.Since(start),
			}, nil
		}
	}

	// 根据动作类型执行
	var result *ActionResult
	switch action.ActionType {
	case "url":
		result, err = s.executeURL(ctx, action, params)
	case "sp":
		result, err = s.executeSP(ctx, action, params)
	case "js":
		result, err = s.executeScript(ctx, action, params, executor.ScriptTypeJavaScript)
	case "py":
		result, err = s.executeScript(ctx, action, params, executor.ScriptTypePython)
	case "go":
		result, err = s.executeScript(ctx, action, params, executor.ScriptTypeGo)
	case "bsh":
		result, err = s.executeScript(ctx, action, params, executor.ScriptTypeBash)
	default:
		return &ActionResult{
			Success:  false,
			Error:    "不支持的动作类型: " + action.ActionType,
			Duration: time.Since(start),
		}, nil
	}

	if err != nil {
		result = &ActionResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// ExecuteActionByName 根据名称执行动作
func (s *service) ExecuteActionByName(ctx context.Context, tableName, actionName string, params map[string]interface{}, userID uint) (*ActionResult, error) {
	// 获取表定义
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   "表不存在: " + err.Error(),
		}, nil
	}

	// 获取动作列表
	actions, err := s.metadataService.GetActions(table.ID)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   "获取动作失败: " + err.Error(),
		}, nil
	}

	// 查找指定名称的动作
	var targetAction *entity.SysAction
	for _, action := range actions {
		if action.Name == actionName {
			targetAction = action
			break
		}
	}

	if targetAction == nil {
		return &ActionResult{
			Success: false,
			Error:   "动作不存在: " + actionName,
		}, nil
	}

	return s.ExecuteAction(ctx, targetAction.ID, params, userID)
}

// BatchExecuteAction 批量执行动作
func (s *service) BatchExecuteAction(ctx context.Context, actionID uint, batchParams []map[string]interface{}, userID uint) ([]*ActionResult, error) {
	results := make([]*ActionResult, 0, len(batchParams))

	for _, params := range batchParams {
		result, _ := s.ExecuteAction(ctx, actionID, params, userID)
		results = append(results, result)
	}

	return results, nil
}

// GetAction 获取动作定义
func (s *service) GetAction(ctx context.Context, actionID uint) (*entity.SysAction, error) {
	var action entity.SysAction
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", actionID, "Y").First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "动作不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询动作失败", err)
	}

	return &action, nil
}

// executeURL 执行URL动作
func (s *service) executeURL(ctx context.Context, action *entity.SysAction, params map[string]interface{}) (*ActionResult, error) {
	// 解析动作内容为URL请求配置
	var urlReq executor.URLRequest
	if err := json.Unmarshal([]byte(action.Content), &urlReq); err != nil {
		return &ActionResult{
			Success: false,
			Error:   "解析URL配置失败: " + err.Error(),
		}, nil
	}

	// 合并参数
	if urlReq.Params == nil {
		urlReq.Params = make(map[string]interface{})
	}
	for k, v := range params {
		urlReq.Params[k] = v
	}

	// 执行URL调用
	urlResp, err := s.urlExecutor.Execute(ctx, &urlReq)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	result := &ActionResult{
		Success:  urlResp.Success,
		Message:  urlResp.Body,
		Data:     urlResp.BodyJSON,
		Duration: urlResp.Duration,
	}

	if !urlResp.Success {
		result.Error = urlResp.Error
	}

	return result, nil
}

// executeSP 执行存储过程
func (s *service) executeSP(ctx context.Context, action *entity.SysAction, params map[string]interface{}) (*ActionResult, error) {
	// 解析动作内容为存储过程请求配置
	var spReq executor.SPRequest
	if err := json.Unmarshal([]byte(action.Content), &spReq); err != nil {
		return &ActionResult{
			Success: false,
			Error:   "解析存储过程配置失败: " + err.Error(),
		}, nil
	}

	// 合并参数
	if spReq.InParams == nil {
		spReq.InParams = make(map[string]interface{})
	}
	for k, v := range params {
		spReq.InParams[k] = v
	}

	// 执行存储过程
	spResp, err := s.spExecutor.Execute(ctx, &spReq)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	result := &ActionResult{
		Success:  spResp.Success,
		Duration: spResp.Duration,
		Data: map[string]interface{}{
			"outParams":  spResp.OutParams,
			"resultSets": spResp.ResultSets,
			"rowsAffected": spResp.RowsAffected,
		},
	}

	if !spResp.Success {
		result.Error = spResp.Error
	}

	return result, nil
}

// executeScript 执行脚本
func (s *service) executeScript(ctx context.Context, action *entity.SysAction, params map[string]interface{}, scriptType executor.ScriptType) (*ActionResult, error) {
	// 创建脚本执行器
	scriptExecutor := executor.NewScriptExecutor(scriptType, s.scriptTimeout)

	// 执行脚本
	execResult, err := scriptExecutor.Execute(ctx, action.Content, params)
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	result := &ActionResult{
		Success:  execResult.Success,
		Message:  execResult.Output,
		Duration: execResult.Duration,
		Data:     execResult.Data,
	}

	if !execResult.Success {
		result.Error = execResult.Error
	}

	return result, nil
}
