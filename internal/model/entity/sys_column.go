package entity

// SysColumn 系统表字段定义
type SysColumn struct {
	BaseModel
	SysTableID     uint   `gorm:"column:SYS_TABLE_ID;index:idx_column_table;not null" json:"sysTableId"`
	DisplayName    string `gorm:"column:DISPLAY_NAME;size:255" json:"displayName"`
	DbName         string `gorm:"column:DB_NAME;size:255" json:"dbName"`
	FullName       string `gorm:"column:FULL_NAME;size:255;uniqueIndex:idx_column_full_name" json:"fullName"`
	ColType        string `gorm:"column:COL_TYPE;size:255" json:"colType"`
	ColLength      int    `gorm:"column:COL_LENGTH" json:"colLength"`
	ColPrecision   int    `gorm:"column:COL_PRECISION" json:"colPrecision"`
	IsDK           string `gorm:"column:IS_DK;size:1" json:"isDk"`       // Y/N
	IsAK           string `gorm:"column:IS_AK;size:1" json:"isAk"`       // Y/N
	NullAble       string `gorm:"column:NULL_ABLE;size:1" json:"nullAble"` // Y/N
	IsUppercase    string `gorm:"column:IS_UPPERCASE;size:1" json:"isUppercase"` // Y/N
	IsQuery        string `gorm:"column:IS_QUERY;size:1" json:"isQuery"` // Y/N
	Orderno        int    `gorm:"column:ORDERNO" json:"orderno"`
	Mask           string `gorm:"column:MASK;size:10" json:"mask"` // 10位字段读写规则
	ModifiAble     string `gorm:"column:MODIFI_ABLE;size:1" json:"modifiAble"`
	SetValueType   string `gorm:"column:SET_VALUE_TYPE;size:255" json:"setValueType"` // pk,docno,createBy,byPage,select,fk,sysdate,operator,password,ignore
	RefTableID     *uint  `gorm:"column:REF_TABLE_ID" json:"refTableId"`
	RefColumnID    *uint  `gorm:"column:REF_COLUMN_ID" json:"refColumnId"`
	RefOnDelete    string `gorm:"column:REF_ON_DELETE;size:255" json:"refOnDelete"`
	Seq            string `gorm:"column:SEQ;size:80" json:"seq"`
	SysDictID      string `gorm:"column:SYS_DICT_ID;size:80" json:"sysDictId"`
	DefaultValue   string `gorm:"column:DEFAULT_VALUE;size:255" json:"defaultValue"`
	RegExpression  string `gorm:"column:REG_EXPRESSION;size:255" json:"regExpression"`
	ErrMsg         string `gorm:"column:ERR_MSG;size:255" json:"errMsg"`
	Filter         string `gorm:"column:FILTER;size:255" json:"filter"`
	DisplayType    string `gorm:"column:DISPLAY_TYPE;size:255" json:"displayType"` // blank,button,hr,check,file,image,select,text,textarea,date,datetime,clob,xml,json
	DisplayCols    int    `gorm:"column:DISPLAY_COLS" json:"displayCols"`
	DisplayRows    int    `gorm:"column:DISPLAY_ROWS" json:"displayRows"`
	Props          string `gorm:"column:PROPS;size:2000" json:"props"` // 扩展属性（JSON）
	IsShowTitle    string `gorm:"column:IS_SHOW_TITLE;size:3" json:"isShowTitle"`
	Submethod      string `gorm:"column:SUBMETHOD;size:1" json:"submethod"`
	HrColumnID     *int   `gorm:"column:HR_COLUMN_ID" json:"hrColumnId"`
	ShowColumnID   *int   `gorm:"column:SHOW_COLUMN_ID" json:"showColumnId"`
	ShowColumnVal  string `gorm:"column:SHOW_COLUMN_VAL;size:255" json:"showColumnVal"`
	Description    string `gorm:"column:DESCRIPTION;size:255" json:"description"`
	Sgrade         int    `gorm:"column:SGRADE" json:"sgrade"` // 字段访问级别
}

// TableName 指定表名
func (SysColumn) TableName() string {
	return "sys_column"
}
