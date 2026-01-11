package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client WebSocket客户端
type Client struct {
	UserID     uint
	Conn       *websocket.Conn
	Send       chan []byte
	Manager    *Manager
	LastActive time.Time
}

// Manager WebSocket连接管理器
type Manager struct {
	clients    map[uint]*Client     // userID -> Client
	broadcast  chan *BroadcastMsg   // 广播消息通道
	register   chan *Client         // 注册客户端通道
	unregister chan *Client         // 注销客户端通道
	mu         sync.RWMutex         // 读写锁
	logger     *zap.Logger
}

// BroadcastMsg 广播消息
type BroadcastMsg struct {
	UserIDs []uint      // 目标用户ID列表（空表示广播给所有人）
	Data    interface{} // 消息数据
}

// MessageType 消息类型
type MessageType string

const (
	// 消息类型常量
	TypeNewMessage      MessageType = "NEW_MESSAGE"       // 新消息
	TypeMessageRead     MessageType = "MESSAGE_READ"      // 消息已读
	TypeMessageDeleted  MessageType = "MESSAGE_DELETED"   // 消息删除
	TypeUnreadCount     MessageType = "UNREAD_COUNT"      // 未读消息数更新
	TypeSystemNotify    MessageType = "SYSTEM_NOTIFY"     // 系统通知
	TypeHeartbeat       MessageType = "HEARTBEAT"         // 心跳
	TypeHeartbeatReply  MessageType = "HEARTBEAT_REPLY"   // 心跳响应
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// NewManager 创建WebSocket管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		clients:    make(map[uint]*Client),
		broadcast:  make(chan *BroadcastMsg, 256),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		logger:     logger,
	}
}

// Run 运行管理器（在goroutine中运行）
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case msg := <-m.broadcast:
			m.broadcastMessage(msg)
		}
	}
}

// registerClient 注册客户端
func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果用户已有连接，先关闭旧连接
	if oldClient, exists := m.clients[client.UserID]; exists {
		close(oldClient.Send)
		oldClient.Conn.Close()
		m.logger.Info("Closed old WebSocket connection",
			zap.Uint("userID", client.UserID))
	}

	m.clients[client.UserID] = client
	m.logger.Info("WebSocket client registered",
		zap.Uint("userID", client.UserID),
		zap.Int("totalClients", len(m.clients)))
}

// unregisterClient 注销客户端
func (m *Manager) unregisterClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[client.UserID]; exists {
		delete(m.clients, client.UserID)
		close(client.Send)
		m.logger.Info("WebSocket client unregistered",
			zap.Uint("userID", client.UserID),
			zap.Int("totalClients", len(m.clients)))
	}
}

// broadcastMessage 广播消息
func (m *Manager) broadcastMessage(msg *BroadcastMsg) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 如果指定了目标用户，只发送给这些用户
	if len(msg.UserIDs) > 0 {
		for _, userID := range msg.UserIDs {
			if client, exists := m.clients[userID]; exists {
				select {
				case client.Send <- m.marshalMessage(msg.Data):
				default:
					// 发送队列已满，跳过
					m.logger.Warn("Client send channel full, message dropped",
						zap.Uint("userID", userID))
				}
			}
		}
	} else {
		// 广播给所有在线用户
		for userID, client := range m.clients {
			select {
			case client.Send <- m.marshalMessage(msg.Data):
			default:
				m.logger.Warn("Client send channel full, message dropped",
					zap.Uint("userID", userID))
			}
		}
	}
}

// marshalMessage 序列化消息
func (m *Manager) marshalMessage(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		m.logger.Error("Failed to marshal WebSocket message", zap.Error(err))
		return []byte("{}")
	}
	return bytes
}

// SendToUser 发送消息给指定用户
func (m *Manager) SendToUser(userID uint, msgType MessageType, data interface{}) {
	msg := &WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	m.broadcast <- &BroadcastMsg{
		UserIDs: []uint{userID},
		Data:    msg,
	}
}

// SendToUsers 发送消息给多个用户
func (m *Manager) SendToUsers(userIDs []uint, msgType MessageType, data interface{}) {
	msg := &WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	m.broadcast <- &BroadcastMsg{
		UserIDs: userIDs,
		Data:    msg,
	}
}

// BroadcastToAll 广播消息给所有在线用户
func (m *Manager) BroadcastToAll(msgType MessageType, data interface{}) {
	msg := &WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	m.broadcast <- &BroadcastMsg{
		UserIDs: nil, // nil表示广播给所有人
		Data:    msg,
	}
}

// GetOnlineCount 获取在线用户数
func (m *Manager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// IsUserOnline 检查用户是否在线
func (m *Manager) IsUserOnline(userID uint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.clients[userID]
	return exists
}

// GetOnlineUsers 获取所有在线用户ID
func (m *Manager) GetOnlineUsers() []uint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userIDs := make([]uint, 0, len(m.clients))
	for userID := range m.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// Register 注册客户端（公开方法）
func (m *Manager) Register(client *Client) {
	m.register <- client
}

// ReadPump 读取客户端消息（心跳检测）
func (c *Client) ReadPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	// 设置读取超时
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastActive = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.Error("WebSocket read error",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
			}
			break
		}

		// 更新活跃时间
		c.LastActive = time.Now()

		// 处理客户端消息（如心跳响应）
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == TypeHeartbeatReply {
				// 心跳响应，更新活跃时间即可
				continue
			}
		}
	}
}

// WritePump 向客户端写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second) // 心跳间隔30秒
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 设置写入超时
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// 通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 写入消息
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.Manager.logger.Error("WebSocket write error",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
				return
			}

		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			heartbeat := &WSMessage{
				Type:      TypeHeartbeat,
				Data:      map[string]interface{}{"ping": "pong"},
				Timestamp: time.Now().Unix(),
			}
			heartbeatBytes, _ := json.Marshal(heartbeat)
			if err := c.Conn.WriteMessage(websocket.TextMessage, heartbeatBytes); err != nil {
				c.Manager.logger.Error("Failed to send heartbeat",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
				return
			}
		}
	}
}
