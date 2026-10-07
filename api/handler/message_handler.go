package handler

import (
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/message"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	messageService message.Service
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(messageService message.Service) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendMessage 发送消息
// @Summary 发送消息给指定用户
// @Tags message
// @Accept json
// @Produce json
// @Param request body message.SendMessageRequest true "消息内容"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/send [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req message.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取发送者ID
	userID, exists := c.Get("userID")
	var senderID *uint
	if exists {
		uid := userID.(uint)
		senderID = &uid
	}

	msg, err := h.messageService.SendMessage(c.Request.Context(), &req, senderID)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "发送成功",
		"data":    msg,
	})
}

// GetMessage 获取消息详情
// @Summary 获取消息详情
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "消息ID格式错误",
		})
		return
	}

	userID, _ := c.Get("userID")

	detail, err := h.messageService.GetMessage(c.Request.Context(), uint(messageID), userID.(uint))
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data":    detail,
	})
}

// ListUserMessages 查询用户消息列表
// @Summary 查询用户消息列表
// @Tags message
// @Accept json
// @Produce json
// @Param request body message.ListMessagesRequest true "查询请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/list [post]
func (h *MessageHandler) ListUserMessages(c *gin.Context) {
	var req message.ListMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	userID, _ := c.Get("userID")

	items, total, err := h.messageService.ListUserMessages(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data": gin.H{
			"items":    items,
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	})
}

// MarkAsRead 标记为已读
// @Summary 标记消息为已读
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/{id}/read [post]
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "消息ID格式错误",
		})
		return
	}

	userID, _ := c.Get("userID")

	if err := h.messageService.MarkAsRead(c.Request.Context(), uint(messageID), userID.(uint)); err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "标记成功",
	})
}

// MarkAllAsRead 标记所有为已读
// @Summary 标记所有消息为已读
// @Tags message
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/read-all [post]
func (h *MessageHandler) MarkAllAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.messageService.MarkAllAsRead(c.Request.Context(), userID.(uint)); err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "标记成功",
	})
}

// GetUnreadCount 获取未读消息数
// @Summary 获取未读消息数量
// @Tags message
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/unread/count [get]
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")

	count, err := h.messageService.GetUnreadCount(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data": gin.H{
			"count": count,
		},
	})
}

// GetUnreadMessages 获取最新未读消息
// @Summary 获取最新未读消息列表
// @Tags message
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/unread/list [get]
func (h *MessageHandler) GetUnreadMessages(c *gin.Context) {
	userID, _ := c.Get("userID")

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items, err := h.messageService.GetUnreadMessages(c.Request.Context(), userID.(uint), limit)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "查询成功",
		"data":    items,
	})
}

// DeleteMessage 删除消息
// @Summary 删除指定消息
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "消息ID格式错误",
		})
		return
	}

	userID, _ := c.Get("userID")

	if err := h.messageService.DeleteMessage(c.Request.Context(), uint(messageID), userID.(uint)); err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// StarMessage 标记/取消星标
// @Summary 标记或取消消息星标
// @Tags message
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Param request body object true "星标请求 {isStarred: boolean}"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/{id}/star [post]
func (h *MessageHandler) StarMessage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "消息ID格式错误",
		})
		return
	}

	var req struct {
		IsStarred bool `json:"isStarred"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	userID, _ := c.Get("userID")

	if err := h.messageService.StarMessage(c.Request.Context(), uint(messageID), userID.(uint), req.IsStarred); err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "操作成功",
	})
}

// ArchiveMessage 归档消息
// @Summary 归档指定消息
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/{id}/archive [post]
func (h *MessageHandler) ArchiveMessage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "消息ID格式错误",
		})
		return
	}

	userID, _ := c.Get("userID")

	if err := h.messageService.ArchiveMessage(c.Request.Context(), uint(messageID), userID.(uint)); err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "归档成功",
	})
}

// SendTemplateMessage 发送模板消息
// @Summary 使用模板发送消息
// @Tags message
// @Accept json
// @Produce json
// @Param request body message.SendTemplateMessageRequest true "模板消息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/send/template [post]
func (h *MessageHandler) SendTemplateMessage(c *gin.Context) {
	var req message.SendTemplateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取发送者ID
	userID, exists := c.Get("userID")
	var senderID *uint
	if exists {
		uid := userID.(uint)
		senderID = &uid
	}

	msg, err := h.messageService.SendTemplateMessage(c.Request.Context(), &req, senderID)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "发送成功",
		"data":    msg,
	})
}

// SendBatchMessage 批量发送消息
// @Summary 批量发送消息给多个用户
// @Tags message
// @Accept json
// @Produce json
// @Param request body object true "批量发送请求 {userIds, message}"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/send/batch [post]
func (h *MessageHandler) SendBatchMessage(c *gin.Context) {
	var req struct {
		UserIDs []uint                     `json:"userIds" binding:"required"`
		Message message.SendMessageRequest `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取发送者ID
	userID, exists := c.Get("userID")
	var senderID *uint
	if exists {
		uid := userID.(uint)
		senderID = &uid
	}

	messages, err := h.messageService.SendBatchMessage(c.Request.Context(), req.UserIDs, &req.Message, senderID)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "批量发送成功",
		"data": gin.H{
			"successCount": len(messages),
			"messages":     messages,
		},
	})
}

// SendToAll 发送给所有用户
// @Summary 群发消息给所有用户
// @Tags message
// @Accept json
// @Produce json
// @Param request body message.SendMessageRequest true "消息内容"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/messages/send/all [post]
func (h *MessageHandler) SendToAll(c *gin.Context) {
	var req message.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取发送者ID
	userID, exists := c.Get("userID")
	var senderID *uint
	if exists {
		uid := userID.(uint)
		senderID = &uid
	}

	msg, err := h.messageService.SendToAll(c.Request.Context(), &req, senderID)
	if err != nil {
		c.JSON(utils.HTTPStatusFor(err), gin.H{
			"code":    errors.GetCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "群发成功",
		"data":    msg,
	})
}
