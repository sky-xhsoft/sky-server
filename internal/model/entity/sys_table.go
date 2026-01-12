package entity

// SysTable 系统表单定义
type SysTable struct {
	BaseModel
	Name               string `gorm:"column:NAME;size:255;uniqueIndex:idx_systable_name;not null" json:"name"`
	DisplayName        string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	RealTableID        *uint  `gorm:"column:REAL_TABLE_ID" json:"realTableId"`
	Filter             string `gorm:"column:FILTER;size:2000" json:"filter"`
	AkColumnID         *int   `gorm:"column:AK_COLUMN_ID" json:"akColumnId"`
	DkColumnID         *uint  `gorm:"column:DK_COLUMN_ID" json:"dkColumnId"`
	Mask               string `gorm:"column:MASK;size:10" json:"mask"` // A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,E:导出,I:导入
	SysTableCategoryID *uint  `gorm:"column:SYS_TABLECATEGORY_ID" json:"sysTableCategoryId"`
	URL                string `gorm:"column:URL;size:255" json:"url"`
	RpcName            string `gorm:"column:RPC_NAME;size:255" json:"rpcName"`
	IsMenu             string `gorm:"column:IS_MENU;size:1" json:"isMenu"`         // Y/N
	IcoImg             string `gorm:"column:ICO_IMG;size:255" json:"icoImg"`       // 表单图标
	IsDropdown         string `gorm:"column:IS_DROPDOWN;size:1" json:"isDropdown"` // Y/N
	SysObjUIConfID     *int   `gorm:"column:SYS_OBJUICONF_ID" json:"sysObjUiConfId"`
	SysDirectoryID     *uint  `gorm:"column:SYS_DIRECTORY_ID" json:"sysDirectoryId"`     // 安全目录
	SysParentTableID   *uint  `gorm:"column:SYS_PARENT_TABLE_ID" json:"sysParentTableId"` // 父表
	RowCnt             *int   `gorm:"column:ROWCNT" json:"rowcnt"`                         // 统计行数
	IsBig              string `gorm:"column:IS_BIG;size:1" json:"isBig"`                   // Y/N 是否海量
	Props              string `gorm:"column:PROPS;size:2000" json:"props"`                 // 扩展属性（JSON）
	Description        string `gorm:"column:DESCRIPTION;size:2000" json:"description"`     // 备注
	OrderNo            int    `gorm:"column:ORDERNO" json:"orderno"`                       // 排序
}

// TableName 指定表名
func (SysTable) TableName() string {
	return "sys_table"
}
