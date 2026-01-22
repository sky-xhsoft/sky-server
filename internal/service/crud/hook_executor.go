package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/executor"
	"gorm.io/gorm"
)

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
