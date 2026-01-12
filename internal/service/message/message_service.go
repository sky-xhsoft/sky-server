package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	ws "github.com/sky-xhsoft/sky-server/internal/pkg/websocket"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// Service 消息服务接口
type Service interface {
	// 消息管理
	SendMessage(ctx context.Context, req *SendMessageRequest, senderID *uint) (*entity.SysMessage, error)
	GetMessage(ctx context.Context, messageID uint, userID uint) (*MessageDetail, error)
	ListUserMessages(ctx context.Context, userID uint, req *ListMessagesRequest) ([]*UserMessageItem, int64, error)
	MarkAsRead(ctx context.Context, messageID uint, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	DeleteMessage(ctx context.Context, messageID uint, userID uint) error
	StarMessage(ctx context.Context, messageID uint, userID uint, isStarred bool) error
	ArchiveMessage(ctx context.Context, messageID uint, userID uint) error

	// 未读消息
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	GetUnreadMessages(ctx context.Context, userID uint, limit int) ([]*UserMessageItem, error)

	// 模板管理
	CreateTemplate(ctx context.Context, template *entity.SysMessageTemplate) error
	GetTemplate(ctx context.Context, code string) (*entity.SysMessageTemplate, error)
	SendTemplateMessage(ctx context.Context, req *SendTemplateMessageRequest, senderID *uint) (*entity.SysMessage, error)

	// 批量操作
	SendBatchMessage(ctx context.Context, userIDs []uint, req *SendMessageRequest, senderID *uint) ([]*entity.SysMessage, error)
	SendToAll(ctx context.Context, req *SendMessageRequest, senderID *uint) (*entity.SysMessage, error)
}

// service 消息服务实现
type service struct {
	db        *gorm.DB
	wsManager *ws.Manager // WebSocket管理器
}

// NewService 创建消息服务
func NewService(db *gorm.DB, wsManager *ws.Manager) Service {
	return &service{
		db:        db,
		wsManager: wsManager,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	MessageType string                 `json:"messageType"`
	Priority    int                    `json:"priority"`
	Category    string                 `json:"category"`
	TargetType  string                 `json:"targetType"` // user, role, group, all
	TargetIDs   []uint                 `json:"targetIds"`
	LinkURL     string                 `json:"linkUrl"`
	LinkType    string                 `json:"linkType"`
	Params      map[string]interface{} `json:"params"`
	ExpireDays  int                    `json:"expireDays"` // 过期天数（0=永久）
}

// SendTemplateMessageRequest 发送模板消息请求
type SendTemplateMessageRequest struct {
	TemplateCode string                 `json:"templateCode" binding:"required"`
	TargetType   string                 `json:"targetType"`
	TargetIDs    []uint                 `json:"targetIds"`
	Variables    map[string]interface{} `json:"variables"`
	LinkURL      string                 `json:"linkUrl"`
	ExpireDays   int                    `json:"expireDays"`
}

// ListMessagesRequest 消息列表请求
type ListMessagesRequest struct {
	Page        int    `json:"page"`
	PageSize    int    `json:"pageSize"`
	MessageType string `json:"messageType"`
	IsRead      string `json:"isRead"`      // Y/N/all
	IsStarred   string `json:"isStarred"`   // Y/N/all
	IsArchived  string `json:"isArchived"`  // Y/N/all
	Priority    *int   `json:"priority"`
	Category    string `json:"category"`
	Keyword     string `json:"keyword"`
}

// MessageDetail 消息详情
type MessageDetail struct {
	*entity.SysMessage
	UserMessage *entity.SysUserMessage `json:"userMessage"`
}

// UserMessageItem 用户消息列表项
type UserMessageItem struct {
	*entity.SysMessage
	IsRead     string `json:"isRead"`
	IsStarred  string `json:"isStarred"`
	IsArchived string `json:"isArchived"`
	ReadTime   string `json:"readTime"`
}

// SendMessage 发送消息
func (s *service) SendMessage(ctx context.Context, req *SendMessageRequest, senderID *uint) (*entity.SysMessage, error) {
	// 设置默认值
	if req.MessageType == "" {
		req.MessageType = "system"
	}
	if req.TargetType == "" {
		req.TargetType = "user"
	}

	// 序列化参数
	var paramsJSON string
	if len(req.Params) > 0 {
		bytes, _ := json.Marshal(req.Params)
		paramsJSON = string(bytes)
	}

	// 计算过期时间
	var expireTime string
	if req.ExpireDays > 0 {
		expireTime = time.Now().AddDate(0, 0, req.ExpireDays).Format("2006-01-02 15:04:05")
	}

	// 目标ID列表
	targetIDsStr := ""
	if len(req.TargetIDs) > 0 {
		targetIDsStr = s.uintsToString(req.TargetIDs)
	}

	// 创建消息
	message := &entity.SysMessage{
		BaseModel: entity.BaseModel{
			CreateBy: s.getSenderName(senderID),
			UpdateBy: s.getSenderName(senderID),
			IsActive: "Y",
		},
		Title:       req.Title,
		Content:     req.Content,
		MessageType: req.MessageType,
		Priority:    req.Priority,
		Category:    req.Category,
		SenderID:    senderID,
		SenderName:  s.getSenderName(senderID),
		TargetType:  req.TargetType,
		TargetIDs:   targetIDsStr,
		LinkURL:     req.LinkURL,
		LinkType:    req.LinkType,
		Params:      paramsJSON,
		TotalCount:  len(req.TargetIDs),
		ExpireTime:  expireTime,
		Status:      "active",
	}

	// 使用事务
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建消息记录
		if err := tx.Create(message).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "创建消息失败", err)
		}

		// 创建用户消息关联
		if req.TargetType == "user" && len(req.TargetIDs) > 0 {
			userMessages := make([]*entity.SysUserMessage, 0, len(req.TargetIDs))
			for _, userID := range req.TargetIDs {
				userMessages = append(userMessages, &entity.SysUserMessage{
					BaseModel: entity.BaseModel{
						CreateBy: s.getSenderName(senderID),
						UpdateBy: s.getSenderName(senderID),
						IsActive: "Y",
					},
					MessageID: message.ID,
					UserID:    userID,
					IsRead:    "N",
				})
			}

			if err := tx.CreateInBatches(userMessages, 100).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "创建用户消息关联失败", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// WebSocket推送新消息通知
	if s.wsManager != nil && req.TargetType == "user" && len(req.TargetIDs) > 0 {
		// 推送给目标用户
		s.wsManager.SendToUsers(req.TargetIDs, ws.TypeNewMessage, map[string]interface{}{
			"messageId":   message.ID,
			"title":       message.Title,
			"content":     message.Content,
			"messageType": message.MessageType,
			"priority":    message.Priority,
			"senderName":  message.SenderName,
			"linkUrl":     message.LinkURL,
			"createTime":  message.CreateTime,
		})

		// 推送未读消息数更新
		for _, userID := range req.TargetIDs {
			count, _ := s.GetUnreadCount(ctx, userID)
			s.wsManager.SendToUser(userID, ws.TypeUnreadCount, map[string]interface{}{
				"count": count,
			})
		}
	}

	return message, nil
}

// GetMessage 获取消息详情
func (s *service) GetMessage(ctx context.Context, messageID uint, userID uint) (*MessageDetail, error) {
	// 查询消息
	var message entity.SysMessage
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", messageID, "Y").First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "消息不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询消息失败", err)
	}

	// 查询用户消息关联
	var userMessage entity.SysUserMessage
	err := s.db.WithContext(ctx).
		Where("MESSAGE_ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", messageID, userID, "Y").
		First(&userMessage).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(errors.ErrDatabase, "查询用户消息关联失败", err)
	}

	return &MessageDetail{
		SysMessage:  &message,
		UserMessage: &userMessage,
	}, nil
}

// ListUserMessages 查询用户消息列表
func (s *service) ListUserMessages(ctx context.Context, userID uint, req *ListMessagesRequest) ([]*UserMessageItem, int64, error) {
	// 设置默认分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建查询
	query := s.db.WithContext(ctx).
		Table("sys_message m").
		Select("m.*, um.IS_READ, um.IS_STARRED, um.IS_ARCHIVED, um.READ_TIME").
		Joins("INNER JOIN sys_user_message um ON m.ID = um.MESSAGE_ID").
		Where("um.USER_ID = ? AND m.IS_ACTIVE = ? AND um.IS_ACTIVE = ? AND um.DELETED_AT IS NULL", userID, "Y", "Y")

	// 应用过滤条件
	if req.MessageType != "" {
		query = query.Where("m.MESSAGE_TYPE = ?", req.MessageType)
	}
	if req.IsRead != "" && req.IsRead != "all" {
		query = query.Where("um.IS_READ = ?", req.IsRead)
	}
	if req.IsStarred != "" && req.IsStarred != "all" {
		query = query.Where("um.IS_STARRED = ?", req.IsStarred)
	}
	if req.IsArchived != "" && req.IsArchived != "all" {
		query = query.Where("um.IS_ARCHIVED = ?", req.IsArchived)
	}
	if req.Priority != nil {
		query = query.Where("m.PRIORITY = ?", *req.Priority)
	}
	if req.Category != "" {
		query = query.Where("m.CATEGORY = ?", req.Category)
	}
	if req.Keyword != "" {
		query = query.Where("(m.TITLE LIKE ? OR m.CONTENT LIKE ?)", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询消息总数失败", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var items []*UserMessageItem
	if err := query.Order("m.PRIORITY DESC, m.CREATE_TIME DESC").
		Limit(req.PageSize).Offset(offset).
		Scan(&items).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询消息列表失败", err)
	}

	return items, total, nil
}

// MarkAsRead 标记为已读
func (s *service) MarkAsRead(ctx context.Context, messageID uint, userID uint) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	result := s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("MESSAGE_ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", messageID, userID, "Y").
		Updates(map[string]interface{}{
			"IS_READ":   "Y",
			"READ_TIME": now,
		})

	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "标记已读失败", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrResourceNotFound, "消息不存在")
	}

	// 更新消息已读人数
	s.db.WithContext(ctx).Model(&entity.SysMessage{}).
		Where("ID = ?", messageID).
		UpdateColumn("READ_COUNT", gorm.Expr("READ_COUNT + 1"))

	// WebSocket推送未读消息数更新
	if s.wsManager != nil {
		count, _ := s.GetUnreadCount(ctx, userID)
		s.wsManager.SendToUser(userID, ws.TypeUnreadCount, map[string]interface{}{
			"count": count,
		})
	}

	return nil
}

// MarkAllAsRead 标记所有未读为已读
func (s *service) MarkAllAsRead(ctx context.Context, userID uint) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	err := s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("USER_ID = ? AND IS_READ = ? AND IS_ACTIVE = ?", userID, "N", "Y").
		Updates(map[string]interface{}{
			"IS_READ":   "Y",
			"READ_TIME": now,
		}).Error

	if err != nil {
		return err
	}

	// WebSocket推送未读消息数更新（应该为0）
	if s.wsManager != nil {
		s.wsManager.SendToUser(userID, ws.TypeUnreadCount, map[string]interface{}{
			"count": 0,
		})
	}

	return nil
}

// GetUnreadCount 获取未读消息数
func (s *service) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("USER_ID = ? AND IS_READ = ? AND IS_ACTIVE = ?", userID, "N", "Y").
		Count(&count).Error

	if err != nil {
		return 0, errors.Wrap(errors.ErrDatabase, "查询未读消息数失败", err)
	}

	return count, nil
}

// GetUnreadMessages 获取最新未读消息
func (s *service) GetUnreadMessages(ctx context.Context, userID uint, limit int) ([]*UserMessageItem, error) {
	if limit <= 0 {
		limit = 10
	}

	var items []*UserMessageItem
	err := s.db.WithContext(ctx).
		Table("sys_message m").
		Select("m.*, um.IS_READ, um.IS_STARRED, um.IS_ARCHIVED, um.READ_TIME").
		Joins("INNER JOIN sys_user_message um ON m.ID = um.MESSAGE_ID").
		Where("um.USER_ID = ? AND um.IS_READ = ? AND m.IS_ACTIVE = ? AND um.IS_ACTIVE = ?", userID, "N", "Y", "Y").
		Order("m.PRIORITY DESC, m.CREATE_TIME DESC").
		Limit(limit).
		Scan(&items).Error

	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询未读消息失败", err)
	}

	return items, nil
}

// DeleteMessage 删除消息（软删除）
func (s *service) DeleteMessage(ctx context.Context, messageID uint, userID uint) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	result := s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("MESSAGE_ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", messageID, userID, "Y").
		Update("DELETED_AT", now)

	if result.Error != nil {
		return errors.Wrap(errors.ErrDatabase, "删除消息失败", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New(errors.ErrResourceNotFound, "消息不存在")
	}

	// WebSocket推送消息删除通知
	if s.wsManager != nil {
		s.wsManager.SendToUser(userID, ws.TypeMessageDeleted, map[string]interface{}{
			"messageId": messageID,
		})

		// 更新未读消息数
		count, _ := s.GetUnreadCount(ctx, userID)
		s.wsManager.SendToUser(userID, ws.TypeUnreadCount, map[string]interface{}{
			"count": count,
		})
	}

	return nil
}

// StarMessage 标记/取消星标
func (s *service) StarMessage(ctx context.Context, messageID uint, userID uint, isStarred bool) error {
	starred := "N"
	if isStarred {
		starred = "Y"
	}

	return s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("MESSAGE_ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", messageID, userID, "Y").
		Update("IS_STARRED", starred).Error
}

// ArchiveMessage 归档消息
func (s *service) ArchiveMessage(ctx context.Context, messageID uint, userID uint) error {
	return s.db.WithContext(ctx).Model(&entity.SysUserMessage{}).
		Where("MESSAGE_ID = ? AND USER_ID = ? AND IS_ACTIVE = ?", messageID, userID, "Y").
		Update("IS_ARCHIVED", "Y").Error
}

// CreateTemplate 创建消息模板
func (s *service) CreateTemplate(ctx context.Context, template *entity.SysMessageTemplate) error {
	return s.db.WithContext(ctx).Create(template).Error
}

// GetTemplate 获取消息模板
func (s *service) GetTemplate(ctx context.Context, code string) (*entity.SysMessageTemplate, error) {
	var template entity.SysMessageTemplate
	if err := s.db.WithContext(ctx).
		Where("CODE = ? AND IS_ENABLED = ? AND IS_ACTIVE = ?", code, "Y", "Y").
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "消息模板不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询消息模板失败", err)
	}
	return &template, nil
}

// SendTemplateMessage 发送模板消息
func (s *service) SendTemplateMessage(ctx context.Context, req *SendTemplateMessageRequest, senderID *uint) (*entity.SysMessage, error) {
	// 获取模板
	template, err := s.GetTemplate(ctx, req.TemplateCode)
	if err != nil {
		return nil, err
	}

	// 替换变量
	title := s.replaceVariables(template.Title, req.Variables)
	content := s.replaceVariables(template.Content, req.Variables)

	// 构建发送请求
	sendReq := &SendMessageRequest{
		Title:       title,
		Content:     content,
		MessageType: template.MessageType,
		Category:    template.Category,
		TargetType:  req.TargetType,
		TargetIDs:   req.TargetIDs,
		LinkURL:     req.LinkURL,
		Params:      req.Variables,
		ExpireDays:  req.ExpireDays,
	}

	return s.SendMessage(ctx, sendReq, senderID)
}

// SendBatchMessage 批量发送消息
func (s *service) SendBatchMessage(ctx context.Context, userIDs []uint, req *SendMessageRequest, senderID *uint) ([]*entity.SysMessage, error) {
	messages := make([]*entity.SysMessage, 0, len(userIDs))

	for _, userID := range userIDs {
		req.TargetIDs = []uint{userID}
		msg, err := s.SendMessage(ctx, req, senderID)
		if err != nil {
			continue // 忽略单个发送失败
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// SendToAll 发送给所有用户
func (s *service) SendToAll(ctx context.Context, req *SendMessageRequest, senderID *uint) (*entity.SysMessage, error) {
	// 查询所有活跃用户ID
	var userIDs []uint
	if err := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Where("IS_ACTIVE = ?", "Y").
		Pluck("ID", &userIDs).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询用户失败", err)
	}

	req.TargetType = "all"
	req.TargetIDs = userIDs

	message, err := s.SendMessage(ctx, req, senderID)
	if err != nil {
		return nil, err
	}

	// 广播给所有在线用户
	if s.wsManager != nil {
		s.wsManager.BroadcastToAll(ws.TypeNewMessage, map[string]interface{}{
			"messageId":   message.ID,
			"title":       message.Title,
			"content":     message.Content,
			"messageType": message.MessageType,
			"priority":    message.Priority,
			"senderName":  message.SenderName,
			"linkUrl":     message.LinkURL,
			"createTime":  message.CreateTime,
		})
	}

	return message, nil
}

// 辅助方法
func (s *service) getSenderName(senderID *uint) string {
	if senderID == nil {
		return "system"
	}
	return fmt.Sprintf("user_%d", *senderID)
}

func (s *service) uintsToString(nums []uint) string {
	strs := make([]string, len(nums))
	for i, num := range nums {
		strs[i] = fmt.Sprintf("%d", num)
	}
	return strings.Join(strs, ",")
}

func (s *service) replaceVariables(template string, variables map[string]interface{}) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}
