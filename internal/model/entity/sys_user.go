package entity

// SysUser 系统用户
type SysUser struct {
	BaseModel
	TrueName string `gorm:"column:TRUE_NAME;size:255" json:"trueName"`
	Username string `gorm:"column:USERNAME;size:255;uniqueIndex;not null" json:"username"`
	Password string `gorm:"column:PASSWORD;size:255;not null" json:"-"` // 密码不返回给前端
	Phone    string `gorm:"column:PHONE;size:20" json:"phone"`
	Email    string `gorm:"column:EMAIL;size:255" json:"email"`
	Language string `gorm:"column:LANGUAGE;size:255" json:"language"`
	IsAdmin  string `gorm:"column:IS_ADMIN;size:2;default:N" json:"isAdmin"` // Y/N
	Sgrade   int    `gorm:"column:SGRADE" json:"sgrade"`                      // 字段访问级别
}

// TableName 指定表名
func (SysUser) TableName() string {
	return "sys_user"
}
