package entity

import "time"

// CloudItemPermission 云盘项目（文件/文件夹）用户权限
// 授权对象类型: user / group / role
// 权限位参考 Linux 文件权限：r=读(查看/预览/下载)  w=写(上传/删除/改名/移动)
type CloudItemPermission struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemID       uint      `gorm:"column:ITEM_ID;index:idx_item_target,unique;not null" json:"itemId"`                 // 云盘项目ID（文件/文件夹）
	TargetType   string    `gorm:"column:TARGET_TYPE;size:20;index:idx_item_target,unique;not null" json:"targetType"` // 授权对象类型: user, group, role
	TargetID     uint      `gorm:"column:TARGET_ID;index:idx_item_target,unique;not null" json:"targetId"`             // 授权对象ID（用户ID/组ID/角色ID）
	PermRead     string    `gorm:"column:PERM_READ;size:1;default:N" json:"permRead"`                                  // 读权限 Y/N
	PermWrite    string    `gorm:"column:PERM_WRITE;size:1;default:N" json:"permWrite"`                                // 写权限 Y/N
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID;index" json:"sysCompanyId"`                                    // 所属公司
	CreateBy     string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME;autoCreateTime" json:"createTime"`
	UpdateBy     string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME;autoUpdateTime" json:"updateTime"`
	IsActive     string    `gorm:"column:IS_ACTIVE;size:1;default:Y" json:"isActive"`
}

// TableName 指定表名
func (CloudItemPermission) TableName() string {
	return "cloud_item_permission"
}
