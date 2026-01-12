package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "github.com/sky-xhsoft/sky-server/internal/pkg/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域（生产环境应该限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	manager *ws.Manager
	logger  *zap.Logger
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(manager *ws.Manager, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
		logger:  logger,
	}
}

// HandleConnection 处理WebSocket连接
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// 从上下文获取用户ID（由JWT中间件设置）
	userIDInterface, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("WebSocket connection attempt without authentication")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	userID := userIDInterface.(uint)

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 创建客户端
	client := &ws.Client{
		UserID:     userID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Manager:    h.manager,
		LastActive: time.Now(),
	}

	// 注册客户端
	h.manager.Register(client)

	// 发送欢迎消息
	welcomeMsg := &ws.WSMessage{
		Type: "CONNECTED",
		Data: map[string]interface{}{
			"userID":    userID,
			"message":   "WebSocket connected successfully",
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	}
	h.manager.SendToUser(userID, "CONNECTED", welcomeMsg.Data)

	h.logger.Info("WebSocket connection established",
		zap.Uint("userID", userID),
		zap.String("remoteAddr", c.Request.RemoteAddr))

	// 启动读写goroutine
	go client.WritePump()
	go client.ReadPump()
}

// GetOnlineUsers 获取在线用户列表
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	onlineUsers := h.manager.GetOnlineUsers()
	onlineCount := h.manager.GetOnlineCount()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data": gin.H{
			"onlineCount": onlineCount,
			"users":       onlineUsers,
		},
	})
}

// CheckUserOnline 检查用户是否在线
func (h *WebSocketHandler) CheckUserOnline(c *gin.Context) {
	userID := c.GetUint("userID")
	isOnline := h.manager.IsUserOnline(userID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data": gin.H{
			"userID":   userID,
			"isOnline": isOnline,
		},
	})
}

// BroadcastMessage 广播消息给所有在线用户（管理员功能）
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	var req struct {
		Type string      `json:"type" binding:"required"`
		Data interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.manager.BroadcastToAll(ws.MessageType(req.Type), req.Data)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "广播成功",
		"data": gin.H{
			"recipients": h.manager.GetOnlineCount(),
		},
	})
}
