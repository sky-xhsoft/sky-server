package schedule

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/executor"
)

// NewTaskRunner 创建一个默认的 TaskRunner:按 actionType 分派到 executor。
// 支持 url/sp/js/py/go/bsh,content 分别是 URL 配置/存储过程配置/脚本/Go 扩展名。
// 该 runner 供调度器复用 executor(与 action.Service 同一套执行机制)。
func NewTaskRunner() TaskRunner {
	urlEx := executor.NewURLExecutor(5 * time.Minute)
	return func(ctx context.Context, actionType, actionContent string, sysCompanyID uint) (string, error) {
		params := map[string]interface{}{"sysCompanyId": sysCompanyID, "scheduleRunAt": time.Now()}
		switch actionType {
		case "url":
			var req executor.URLRequest
			if err := json.Unmarshal([]byte(actionContent), &req); err != nil {
				return "", err
			}
			if req.Params == nil {
				req.Params = map[string]interface{}{}
			}
			for k, v := range params {
				req.Params[k] = v
			}
			resp, err := urlEx.Execute(ctx, &req)
			if err != nil {
				return "", err
			}
			if !resp.Success {
				respErr := resp.Error
				if respErr == "" {
					respErr = "URL 执行失败"
				}
				return resp.Body, &ScheduleError{msg: respErr}
			}
			return resp.Body, nil
		case "sp":
			return "", &ScheduleError{msg: "sp 存储过程已废弃,请改用 go 扩展程序"}
		case "go":
			se := executor.NewScriptExecutor(executor.ScriptTypeGo, 5*time.Minute)
			res, err := se.Execute(ctx, actionContent, params)
			if err != nil {
				return "", err
			}
			if !res.Success {
				return res.Output, &ScheduleError{msg: res.Error}
			}
			return res.Output, nil
		case "js", "py", "bsh":
			var st executor.ScriptType
			switch actionType {
			case "js":
				st = executor.ScriptTypeJavaScript
			case "py":
				st = executor.ScriptTypePython
			case "bsh":
				st = executor.ScriptTypeBash
			}
			se := executor.NewScriptExecutor(st, 5*time.Minute)
			res, err := se.Execute(ctx, actionContent, params)
			if err != nil {
				return "", err
			}
			if !res.Success {
				return res.Output, &ScheduleError{msg: res.Error}
			}
			return res.Output, nil
		default:
			return "", &ScheduleError{msg: "不支持的任务动作类型: " + actionType}
		}
	}
}