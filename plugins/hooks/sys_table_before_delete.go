package hooks

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// SysTableBeforeDeleteHook sys_table 删除前钩子
type SysTableBeforeDeleteHook struct {
	*BaseHook
}

// 在 init() 中自动注册
func init() {
	hook := &SysTableBeforeDeleteHook{
		BaseHook: NewBaseHook("SYS_TABLE_BEFORE_DELETE", sysTableBeforeDeleteHandler),
	}
	Register(hook)
}

// sysTableBeforeDeleteHandler 处理 sys_table 删除前的逻辑
func sysTableBeforeDeleteHandler(manager *core.Manager) func(map[string]interface{}) (interface{}, error) {
	return func(params map[string]interface{}) (interface{}, error) {
		logger.Info("执行 SYS_TABLE_BEFORE_DELETE 钩子", zap.Any("params", params))

		// 获取数据库连接（事务连接）
		txDB, err := GetDBFromParams(params)
		if err != nil {
			return nil, err
		}

		// 获取记录 ID
		recordID, err := GetUintFromParams(params, "ID")
		if err != nil {
			return nil, err
		}

		// 获取公司 ID（可选）
		companyID := GetUintOrZero(params, "SYS_COMPANY_ID")

		// 构造插件数据
		pluginData := core.PluginData{
			TableName: "sys_table",
			Action:    "delete",
			Timing:    "before",
			RecordID:  recordID,
			CompanyID: companyID,
			Data:      params,
		}

		// 执行插件（使用事务连接）
		ctx := context.Background()
		if err := manager.ExecuteWithDB(ctx, txDB, pluginData); err != nil {
			logger.Error("执行插件失败", zap.Error(err))
			return nil, err
		}

		return SuccessResult("sys_table 删除前钩子执行成功"), nil
	}
}
