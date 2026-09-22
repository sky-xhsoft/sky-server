package hooks

import (
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// ============================================================
// 第8章 存储过程 → 扩展程序 示例
// 以前用存储过程(CALL ...)实现的动作/钩子,现在改为注册一个 Go 扩展:
//   在 sys_table_cmd / sys_action / sys_schedule 里用 ContentType/go 类型、
//   Content 填 "CRUD_SUBMIT_DEMO" 即可调用下面的 Go 实现。
// 这复用 executor.goExecutor 的 GoFuncRegistry 机制,无需新增框架。
// ============================================================

// CRUDSubmitDemoHook 示例扩展:替代"提交"存储过程。
// 仅做演示:打印参数并返回成功;真实场景在此写 gorm 事务逻辑(用 GetDBFromParams 拿事务连接)。
type CRUDSubmitDemoHook struct {
	*BaseHook
}

// 在 init() 中自动注册
func init() {
	hook := &CRUDSubmitDemoHook{
		BaseHook: NewBaseHook("CRUD_SUBMIT_DEMO", crudSubmitDemoHandler),
	}
	Register(hook)
}

// crudSubmitDemoHandler 实现"提交"扩展程序逻辑
func crudSubmitDemoHandler(manager *core.Manager, cfg *config.Config) func(map[string]interface{}) (interface{}, error) {
	return func(params map[string]interface{}) (interface{}, error) {
		logger.Info("执行 CRUD_SUBMIT_DEMO 扩展(替代存储过程)", zap.Any("params", params))

		// 获取事务数据库连接(若在 CRUD 钩子事务中调用)
		if txDB, err := GetDBFromParams(params); err == nil && txDB != nil {
			_ = txDB // 可在此执行 gorm 逻辑,如更新业务流程字段
		}

		// 获取记录 ID / 表名
		recordID := GetUintOrZero(params, "ID")

		return SuccessResult("CRUD_SUBMIT_DEMO 扩展执行成功(替代存储过程), recordId=" + itoa(recordID)), nil
	}
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}