package entity

import "time"

// AuditLog 审计日志
type AuditLog struct {
	ID            uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UserID        uint      `gorm:"column:USER_ID;index" json:"userId"`                       // 操作用户ID
	Username      string    `gorm:"column:USERNAME;size:80" json:"username"`                  // 操作用户名
	Action        string    `gorm:"column:ACTION;size:50;not null;index" json:"action"`       // 操作类型
	Resource      string    `gorm:"column:RESOURCE;size:100;index" json:"resource"`           // 资源类型
	ResourceID    string    `gorm:"column:RESOURCE_ID;size:100;index" json:"resourceId"`      // 资源ID
	ResourceName  string    `gorm:"column:RESOURCE_NAME;size:255" json:"resourceName"`        // 资源名称
	Method        string    `gorm:"column:METHOD;size:10" json:"method"`                      // HTTP方法
	Path          string    `gorm:"column:PATH;size:500" json:"path"`                         // 请求路径
	IP            string    `gorm:"column:IP;size:50" json:"ip"`                              // 客户端IP
	UserAgent     string    `gorm:"column:USER_AGENT;size:500" json:"userAgent"`              // 用户代理
	Status        string    `gorm:"column:STATUS;size:20;not null;index" json:"status"`       // 操作状态: success, failure
	ErrorMessage  string    `gorm:"column:ERROR_MESSAGE;size:2000" json:"errorMessage"`       // 错误信息
	RequestBody   string    `gorm:"column:REQUEST_BODY;type:text" json:"requestBody"`         // 请求体
	ResponseBody  string    `gorm:"column:RESPONSE_BODY;type:text" json:"responseBody"`       // 响应体
	OldValue      string    `gorm:"column:OLD_VALUE;type:text" json:"oldValue"`               // 修改前的值(JSON)
	NewValue      string    `gorm:"column:NEW_VALUE;type:text" json:"newValue"`               // 修改后的值(JSON)
	Duration      int64     `gorm:"column:DURATION" json:"duration"`                          // 执行时长(毫秒)
	Tags          string    `gorm:"column:TAGS;size:500" json:"tags"`                         // 标签(用于分类和搜索)
	CreatedAt     time.Time `gorm:"column:CREATED_AT;index" json:"createdAt"`                 // 创建时间
	SysCompanyID  uint      `gorm:"column:SYS_COMPANY_ID" json:"sysCompanyId"`                // 所属公司
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_log"
}

// AuditAction 审计操作类型常量
const (
	// 认证相关
	ActionLogin       = "login"        // 登录
	ActionLogout      = "logout"       // 登出
	ActionRefresh     = "refresh"      // 刷新token
	ActionKickDevice  = "kick_device"  // 踢出设备

	// CRUD操作
	ActionCreate = "create" // 创建
	ActionRead   = "read"   // 读取
	ActionUpdate = "update" // 更新
	ActionDelete = "delete" // 删除
	ActionQuery  = "query"  // 查询

	// 动作执行
	ActionExecute      = "execute"       // 执行动作
	ActionBatchExecute = "batch_execute" // 批量执行

	// 工作流操作
	ActionStartProcess     = "start_process"     // 启动流程
	ActionCompleteTask     = "complete_task"     // 完成任务
	ActionClaimTask        = "claim_task"        // 签收任务
	ActionTransferTask     = "transfer_task"     // 转交任务
	ActionTerminateProcess = "terminate_process" // 终止流程
	ActionPublishWorkflow  = "publish_workflow"  // 发布流程

	// 权限操作
	ActionGrantPermission  = "grant_permission"  // 授予权限
	ActionRevokePermission = "revoke_permission" // 撤销权限

	// 配置操作
	ActionUpdateConfig  = "update_config"  // 更新配置
	ActionRefreshCache  = "refresh_cache"  // 刷新缓存
	ActionResetSequence = "reset_sequence" // 重置序号
)

// AuditStatus 审计状态常量
const (
	StatusSuccess = "success" // 成功
	StatusFailure = "failure" // 失败
)

// AuditResource 审计资源类型常量
const (
	ResourceUser       = "user"        // 用户
	ResourceTable      = "table"       // 数据表
	ResourceAction     = "action"      // 动作
	ResourceWorkflow   = "workflow"    // 工作流
	ResourceTask       = "task"        // 任务
	ResourceDict       = "dict"        // 字典
	ResourceSequence   = "sequence"    // 序号
	ResourcePermission = "permission"  // 权限
)
