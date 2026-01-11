package entity

// SysSeq 序号生成器
type SysSeq struct {
	BaseModel
	Name        string `gorm:"column:NAME;size:255;uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	VFormat     string `gorm:"column:VFORMAT;size:255" json:"vformat"` // 格式：PO{YYYY}{MM}{DD}{0000}
	Incre       int    `gorm:"column:INCRE" json:"incre"`              // 递增步长
	CycleType   string `gorm:"column:CYCLETYPE;size:1" json:"cycleType"` // D:日, M:月, Y:年, N:不循环
	Prefix      string `gorm:"column:PREFIX;size:10" json:"prefix"`
	Suffix      string `gorm:"column:SUFFIX;size:10" json:"suffix"`
	CurDate     string `gorm:"column:CUR_DATE;size:20" json:"curDate"` // 当前周期值
	CurNum      int    `gorm:"column:CUR_NUM" json:"curNum"`            // 当前流水号
}

// TableName 指定表名
func (SysSeq) TableName() string {
	return "sys_seq"
}
