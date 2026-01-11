package entity

import "time"

// Menu 菜单
type Menu struct {
	ID           uint      `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	SysCompanyID uint      `gorm:"column:SYS_COMPANY_ID" json:"sysCompanyId"`
	CreateBy     string    `gorm:"column:CREATE_BY;size:80" json:"createBy"`
	CreateTime   time.Time `gorm:"column:CREATE_TIME" json:"createTime"`
	UpdateBy     string    `gorm:"column:UPDATE_BY;size:80" json:"updateBy"`
	UpdateTime   time.Time `gorm:"column:UPDATE_TIME" json:"updateTime"`
	IsActive     string    `gorm:"column:IS_ACTIVE;size:1;not null;default:'Y'" json:"isActive"`

	MenuName   string `gorm:"column:MENU_NAME;size:100;not null" json:"menuName"`                      // 菜单名称
	ParentID   uint   `gorm:"column:PARENT_ID;index" json:"parentId"`                                  // 父菜单ID
	MenuType   string `gorm:"column:MENU_TYPE;size:20;not null;index" json:"menuType"`                 // 菜单类型(dir:目录,menu:菜单,button:按钮)
	Path       string `gorm:"column:PATH;size:200" json:"path"`                                        // 路由路径
	Component  string `gorm:"column:COMPONENT;size:200" json:"component"`                              // 组件路径
	PermCode   string `gorm:"column:PERM_CODE;size:100;index" json:"permCode"`                         // 权限编码(关联权限表)
	Icon       string `gorm:"column:ICON;size:100" json:"icon"`                                        // 图标
	SortOrder  int    `gorm:"column:SORT_ORDER;default:0" json:"sortOrder"`                            // 排序号
	IsVisible  string `gorm:"column:IS_VISIBLE;size:1;not null;default:'Y'" json:"isVisible"`          // 是否可见(Y/N)
	IsCache    string `gorm:"column:IS_CACHE;size:1;not null;default:'N'" json:"isCache"`              // 是否缓存(Y/N)
	IsFrame    string `gorm:"column:IS_FRAME;size:1;not null;default:'N'" json:"isFrame"`              // 是否外链(Y/N)
	Status     string `gorm:"column:STATUS;size:20;not null;default:'enabled';index" json:"status"`    // 状态(enabled:启用,disabled:禁用)
	Redirect   string `gorm:"column:REDIRECT;size:200" json:"redirect"`                                // 重定向路径
	AlwaysShow string `gorm:"column:ALWAYS_SHOW;size:1;not null;default:'N'" json:"alwaysShow"`        // 是否总是显示(Y/N,有子菜单时是否显示父菜单)
	Remark     string `gorm:"column:REMARK;size:500" json:"remark"`                                    // 备注
}

// TableName 指定表名
func (Menu) TableName() string {
	return "sys_menu"
}

// MenuType 菜单类型常量
const (
	MenuTypeDir    = "dir"    // 目录
	MenuTypeMenu   = "menu"   // 菜单
	MenuTypeButton = "button" // 按钮
)

// MenuStatus 菜单状态常量
const (
	MenuStatusEnabled  = "enabled"  // 启用
	MenuStatusDisabled = "disabled" // 禁用
)

// MenuNode 菜单树节点
type MenuNode struct {
	*Menu
	Children []*MenuNode `json:"children"`
}

// Meta 前端路由Meta信息
type Meta struct {
	Title      string `json:"title"`                // 菜单标题
	Icon       string `json:"icon,omitempty"`       // 图标
	NoCache    bool   `json:"noCache,omitempty"`    // 不缓存
	AlwaysShow bool   `json:"alwaysShow,omitempty"` // 总是显示
	Hidden     bool   `json:"hidden,omitempty"`     // 隐藏
}

// RouterVO 前端路由对象
type RouterVO struct {
	Name      string      `json:"name"`                // 路由名称
	Path      string      `json:"path"`                // 路由路径
	Hidden    bool        `json:"hidden"`              // 是否隐藏
	Redirect  string      `json:"redirect,omitempty"`  // 重定向
	Component string      `json:"component"`           // 组件路径
	Meta      Meta        `json:"meta"`                // Meta信息
	Children  []*RouterVO `json:"children,omitempty"`  // 子路由
}
