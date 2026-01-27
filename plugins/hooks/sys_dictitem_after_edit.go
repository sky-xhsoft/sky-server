package hooks

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// SysDictItemAfterEditHook sys_dict_item 新增/修改后钩子
type SysDictItemAfterEditHook struct {
	*BaseHook
}

// 在 init() 中自动注册
func init() {
	hook := &SysDictItemAfterEditHook{
		BaseHook: NewBaseHook("SYS_DICT_ITEM_AFTER_EDIT", sysDictItemAfterEditHandler),
	}
	Register(hook)
}

// sysDictItemAfterEditHandler 处理 sys_dict_item 新增/修改后的逻辑
// 功能：
// 1. 如果当前记录的 IS_DEFAULT_VALUE = 'Y'，则更新 sys_dict 表中对应记录的 DEFAULT_VALUE 为当前记录的 VALUE
// 2. 同时将同一 SYS_DICT_ID 下的其他记录的 IS_DEFAULT_VALUE 设置为 'N'
func sysDictItemAfterEditHandler(manager *core.Manager) func(map[string]interface{}) (interface{}, error) {
	return func(params map[string]interface{}) (interface{}, error) {
		logger.Info("执行 SYS_DICT_ITEM_AFTER_EDIT 钩子", zap.Any("params", params))

		// 获取数据库连接（事务连接）
		txDB, err := GetDBFromParams(params)
		if err != nil {
			logger.Error("获取数据库连接失败", zap.Error(err))
			return nil, err
		}

		// 获取记录 ID
		recordID, err := GetUintFromParams(params, "ID")
		if err != nil {
			logger.Error("获取记录 ID 失败", zap.Error(err))
			return nil, err
		}

		// 从数据库查询完整的记录信息
		var record map[string]interface{}
		if err := txDB.Table("sys_dict_item").Where("ID = ?", recordID).Find(&record).Error; err != nil {
			logger.Error("查询记录失败", zap.Error(err), zap.Uint("recordID", recordID))
			return nil, err
		}

		// 检查记录是否存在
		if len(record) == 0 {
			logger.Error("记录不存在", zap.Uint("recordID", recordID))
			return nil, fmt.Errorf("记录不存在: ID=%d", recordID)
		}

		logger.Info("查询到的记录", zap.Any("record", record))

		// 获取 IS_DEFAULT_VALUE 字段
		isDefaultValue := ""
		if val, ok := record["IS_DEFAULT_VALUE"].(string); ok {
			isDefaultValue = val
		}

		// 如果不是默认值，直接返回
		if isDefaultValue != "Y" {
			logger.Info("IS_DEFAULT_VALUE 不为 Y，跳过处理", zap.Uint("recordID", recordID))
			return SuccessResult("sys_dict_item 编辑后钩子执行成功（无需处理）"), nil
		}

		// 获取 SYS_DICT_ID
		sysDictID, err := GetUintFromMap(record, "SYS_DICT_ID")
		if err != nil {
			logger.Error("获取 SYS_DICT_ID 失败", zap.Error(err))
			return nil, err
		}

		// 获取 VALUE
		value, err := GetStringFromMap(record, "VALUE")
		if err != nil {
			logger.Error("获取 VALUE 失败", zap.Error(err))
			return nil, err
		}

		logger.Info("开始处理默认值更新",
			zap.Uint("recordID", recordID),
			zap.Uint("sysDictID", sysDictID),
			zap.String("value", value))

		// 1. 更新 sys_dict 表的 DEFAULT_VALUE
		if err := txDB.Table("sys_dict").
			Where("ID = ?", sysDictID).
			Update("DEFAULT_VALUE", value).Error; err != nil {
			logger.Error("更新 sys_dict DEFAULT_VALUE 失败", zap.Error(err))
			return nil, err
		}

		logger.Info("更新 sys_dict DEFAULT_VALUE 成功",
			zap.Uint("sysDictID", sysDictID),
			zap.String("defaultValue", value))

		// 2. 将同一 SYS_DICT_ID 下的其他记录的 IS_DEFAULT_VALUE 设置为 'N'
		if err := txDB.Table("sys_dict_item").
			Where("SYS_DICT_ID = ? AND ID != ?", sysDictID, recordID).
			Update("IS_DEFAULT_VALUE", "N").Error; err != nil {
			logger.Error("更新其他记录的 IS_DEFAULT_VALUE 失败", zap.Error(err))
			return nil, err
		}

		logger.Info("更新其他记录的 IS_DEFAULT_VALUE 成功",
			zap.Uint("sysDictID", sysDictID),
			zap.Uint("excludeRecordID", recordID))

		// 获取公司 ID（可选）
		companyID := GetUintOrZero(params, "SYS_COMPANY_ID")

		// 构造插件数据
		pluginData := core.PluginData{
			TableName: "sys_dict_item",
			Action:    "edit",
			Timing:    "after",
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

		return SuccessResult("sys_dict_item 编辑后钩子执行成功"), nil
	}
}
