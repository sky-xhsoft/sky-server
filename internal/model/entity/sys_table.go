package entity

// SysTable 系统表单定义
type SysTable struct {
	BaseModel
	Name               string `gorm:"column:NAME;size:255;uniqueIndex:idx_systable_name;not null" json:"NAME"`
	DisplayName        string `gorm:"column:DISPLAY_NAME;size:255" json:"DISPLAY_NAME"`
	RealTableID        *uint  `gorm:"column:REAL_TABLE_ID" json:"REAL_TABLE_ID"`
	Filter             string `gorm:"column:FILTER;size:2000" json:"FILTER"`
	AkColumnID         *int   `gorm:"column:AK_COLUMN_ID" json:"AK_COLUMN_ID"`
	DkColumnID         *uint  `gorm:"column:DK_COLUMN_ID" json:"DK_COLUMN_ID"`
	Mask               string `gorm:"column:MASK;size:10" json:"MASK"` // A:新增,M:修改,D:删除,Q:查询,S:提交,U:反提交,V:作废,E:导出,I:导入
	SysTableCategoryID *uint  `gorm:"column:SYS_TABLECATEGORY_ID" json:"SYS_TABLECATEGORY_ID"`
	URL                string `gorm:"column:URL;size:255" json:"URL"`
	RpcName            string `gorm:"column:RPC_NAME;size:255" json:"RPC_NAME"`
	IsMenu             string `gorm:"column:IS_MENU;size:1" json:"IS_MENU"`         // Y/N
	IcoImg             string `gorm:"column:ICO_IMG;size:255" json:"ICO_IMG"`       // 表单图标
	IsDropdown         string `gorm:"column:IS_DROPDOWN;size:1" json:"IS_DROPDOWN"` // Y/N
	SysObjUIConfID     *int   `gorm:"column:SYS_OBJUICONF_ID" json:"SYS_OBJUICONF_ID"`
	SysDirectoryID     *uint  `gorm:"column:SYS_DIRECTORY_ID" json:"SYS_DIRECTORY_ID"`     // 安全目录
	SysParentTableID   *uint  `gorm:"column:SYS_PARENT_TABLE_ID" json:"SYS_PARENT_TABLE_ID"` // 父表
	RowCnt             *int   `gorm:"column:ROWCNT" json:"ROWCNT"`                         // 统计行数
	IsBig              string `gorm:"column:IS_BIG;size:1" json:"IS_BIG"`                   // Y/N 是否海量
	Props              string `gorm:"column:PROPS;size:2000" json:"PROPS"`                 // 扩展属性（JSON）
	Description        string `gorm:"column:DESCRIPTION;size:2000" json:"DESCRIPTION"`     // 备注
	OrderNo            int    `gorm:"column:ORDERNO" json:"ORDERNO"`                       // 排序
}

// TableName 指定表名
func (SysTable) TableName() string {
	return "sys_table"
}
