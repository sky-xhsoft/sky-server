package entity

// SysMessage 系统消息
type SysMessage struct {
	BaseModel
	Title       string `gorm:"column:TITLE;size:255;not null" json:"title"`                         // 消息标题
	Content     string `gorm:"column:CONTENT;type:text;not null" json:"content"`                    // 消息内容
	MessageType string `gorm:"column:MESSAGE_TYPE;size:50;not null;index" json:"messageType"`       // 消息类型: system, workflow, business, notice
	Priority    int    `gorm:"column:PRIORITY;default:0;index" json:"priority"`                     // 优先级: 0=普通, 1=重要, 2=紧急
	Category    string `gorm:"column:CATEGORY;size:50;index" json:"category"`                       // 消息分类
	SenderID    *uint  `gorm:"column:SENDER_ID;index" json:"senderId"`                              // 发送者ID（系统消息为NULL）
	SenderName  string `gorm:"column:SENDER_NAME;size:100" json:"senderName"`                       // 发送者姓名
	TargetType  string `gorm:"column:TARGET_TYPE;size:20;default:user" json:"targetType"`           // 目标类型: user, role, group, all
	TargetIDs   string `gorm:"column:TARGET_IDS;size:1000" json:"targetIds"`                        // 目标ID列表（逗号分隔）
	LinkURL     string `gorm:"column:LINK_URL;size:500" json:"linkUrl"`                             // 关联URL
	LinkType    string `gorm:"column:LINK_TYPE;size:50" json:"linkType"`                            // 链接类型: internal, external
	Params      string `gorm:"column:PARAMS;type:text" json:"params"`                               // 消息参数（JSON）
	TemplateID  *uint  `gorm:"column:TEMPLATE_ID;index" json:"templateId"`                          // 消息模板ID
	ReadCount   int    `gorm:"column:READ_COUNT;default:0" json:"readCount"`                        // 已读人数
	TotalCount  int    `gorm:"column:TOTAL_COUNT;default:0" json:"totalCount"`                      // 总接收人数
	ExpireTime  string `gorm:"column:EXPIRE_TIME;type:datetime" json:"expireTime"`                  // 过期时间
	Status      string `gorm:"column:STATUS;size:20;default:active" json:"status"`                  // 状态: active, expired, deleted
}

// TableName 指定表名
func (SysMessage) TableName() string {
	return "sys_message"
}

// SysUserMessage 用户消息关联
type SysUserMessage struct {
	BaseModel
	MessageID  uint   `gorm:"column:MESSAGE_ID;index:idx_user_msg;not null" json:"messageId"`      // 消息ID
	UserID     uint   `gorm:"column:USER_ID;index:idx_user_msg;not null" json:"userId"`            // 用户ID
	IsRead     string `gorm:"column:IS_READ;size:1;default:N;index" json:"isRead"`                 // 是否已读 Y/N
	ReadTime   string `gorm:"column:READ_TIME;type:datetime" json:"readTime"`                      // 读取时间
	IsStarred  string `gorm:"column:IS_STARRED;size:1;default:N" json:"isStarred"`                 // 是否星标 Y/N
	IsArchived string `gorm:"column:IS_ARCHIVED;size:1;default:N" json:"isArchived"`               // 是否归档 Y/N
	DeletedAt  string `gorm:"column:DELETED_AT;type:datetime" json:"deletedAt"`                    // 删除时间（软删除）
}

// TableName 指定表名
func (SysUserMessage) TableName() string {
	return "sys_user_message"
}

// SysMessageTemplate 消息模板
type SysMessageTemplate struct {
	BaseModel
	Code        string `gorm:"column:CODE;size:50;uniqueIndex;not null" json:"code"`                // 模板代码
	Name        string `gorm:"column:NAME;size:100;not null" json:"name"`                           // 模板名称
	MessageType string `gorm:"column:MESSAGE_TYPE;size:50;not null" json:"messageType"`             // 消息类型
	Title       string `gorm:"column:TITLE;size:255;not null" json:"title"`                         // 标题模板
	Content     string `gorm:"column:CONTENT;type:text;not null" json:"content"`                    // 内容模板
	Variables   string `gorm:"column:VARIABLES;size:500" json:"variables"`                          // 变量列表（逗号分隔）
	Description string `gorm:"column:DESCRIPTION;size:500" json:"description"`                      // 描述
	IsEnabled   string `gorm:"column:IS_ENABLED;size:1;default:Y" json:"isEnabled"`                 // 是否启用 Y/N
	Category    string `gorm:"column:CATEGORY;size:50" json:"category"`                             // 分类
}

// TableName 指定表名
func (SysMessageTemplate) TableName() string {
	return "sys_message_template"
}

// SysEmailConfig 邮件配置
type SysEmailConfig struct {
	BaseModel
	SmtpHost     string `gorm:"column:SMTP_HOST;size:100;not null" json:"smtpHost"`                 // SMTP服务器地址
	SmtpPort     int    `gorm:"column:SMTP_PORT;not null" json:"smtpPort"`                          // SMTP端口
	SmtpUser     string `gorm:"column:SMTP_USER;size:100;not null" json:"smtpUser"`                 // SMTP用户名
	SmtpPassword string `gorm:"column:SMTP_PASSWORD;size:255;not null" json:"smtpPassword"`         // SMTP密码（加密存储）
	FromEmail    string `gorm:"column:FROM_EMAIL;size:100;not null" json:"fromEmail"`               // 发件人邮箱
	FromName     string `gorm:"column:FROM_NAME;size:100" json:"fromName"`                          // 发件人名称
	UseTLS       string `gorm:"column:USE_TLS;size:1;default:Y" json:"useTls"`                      // 是否使用TLS Y/N
	IsDefault    string `gorm:"column:IS_DEFAULT;size:1;default:N" json:"isDefault"`                // 是否默认配置 Y/N
	Description  string `gorm:"column:DESCRIPTION;size:500" json:"description"`                     // 描述
}

// TableName 指定表名
func (SysEmailConfig) TableName() string {
	return "sys_email_config"
}

// SysNotificationLog 通知日志
type SysNotificationLog struct {
	BaseModel
	MessageID    uint   `gorm:"column:MESSAGE_ID;index" json:"messageId"`                           // 消息ID
	UserID       uint   `gorm:"column:USER_ID;index" json:"userId"`                                 // 接收用户ID
	NotifyType   string `gorm:"column:NOTIFY_TYPE;size:20;not null" json:"notifyType"`              // 通知类型: websocket, email, sms
	Status       string `gorm:"column:STATUS;size:20;not null" json:"status"`                       // 状态: pending, sent, failed, read
	SentTime     string `gorm:"column:SENT_TIME;type:datetime" json:"sentTime"`                     // 发送时间
	ReadTime     string `gorm:"column:READ_TIME;type:datetime" json:"readTime"`                     // 读取时间
	ErrorMessage string `gorm:"column:ERROR_MESSAGE;size:500" json:"errorMessage"`                  // 错误信息
	RetryCount   int    `gorm:"column:RETRY_COUNT;default:0" json:"retryCount"`                     // 重试次数
}

// TableName 指定表名
func (SysNotificationLog) TableName() string {
	return "sys_notification_log"
}
