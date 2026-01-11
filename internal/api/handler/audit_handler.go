package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/audit"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService audit.Service
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService audit.Service) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// QueryLogsRequest 查询审计日志请求
type QueryLogsRequest struct {
	UserID     uint   `form:"userId"`
	Username   string `form:"username"`
	Action     string `form:"action"`
	Resource   string `form:"resource"`
	ResourceID string `form:"resourceId"`
	Status     string `form:"status"`
	IP         string `form:"ip"`
	StartTime  string `form:"startTime"` // 格式: 2006-01-02 15:04:05
	EndTime    string `form:"endTime"`   // 格式: 2006-01-02 15:04:05
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	SortBy     string `form:"sortBy"`
	SortOrder  string `form:"sortOrder"`
}

// QueryLogs 查询审计日志
func (h *AuditHandler) QueryLogs(c *gin.Context) {
	var req QueryLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 转换时间参数
	var startTime, endTime time.Time
	var err error
	if req.StartTime != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    errors.ErrInvalidParam,
				"message": "开始时间格式错误",
			})
			return
		}
	}
	if req.EndTime != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    errors.ErrInvalidParam,
				"message": "结束时间格式错误",
			})
			return
		}
	}

	// 构建查询请求
	queryReq := &audit.QueryRequest{
		UserID:     req.UserID,
		Username:   req.Username,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Status:     req.Status,
		IP:         req.IP,
		StartTime:  startTime,
		EndTime:    endTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	logs, total, err := h.auditService.QueryLogs(c.Request.Context(), queryReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  queryReq.Page,
		},
	})
}

// GetLog 获取单条审计日志
func (h *AuditHandler) GetLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的ID",
		})
		return
	}

	log, err := h.auditService.GetLog(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    log,
	})
}

// GetUserLogs 获取用户的审计日志
func (h *AuditHandler) GetUserLogs(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "无效的用户ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.auditService.GetUserLogs(c.Request.Context(), uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  page,
		},
	})
}

// GetResourceLogs 获取资源的审计日志
func (h *AuditHandler) GetResourceLogs(c *gin.Context) {
	resource := c.Param("resource")
	resourceID := c.Param("resourceId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.auditService.GetResourceLogs(c.Request.Context(), resource, resourceID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  page,
		},
	})
}

// StatisticsRequest 统计请求
type StatisticsRequest struct {
	StartTime string `form:"startTime"` // 格式: 2006-01-02 15:04:05
	EndTime   string `form:"endTime"`   // 格式: 2006-01-02 15:04:05
	GroupBy   string `form:"groupBy"`   // action, resource, user, date
}

// GetStatistics 获取审计统计
func (h *AuditHandler) GetStatistics(c *gin.Context) {
	var req StatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 转换时间参数
	var startTime, endTime time.Time
	var err error
	if req.StartTime != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    errors.ErrInvalidParam,
				"message": "开始时间格式错误",
			})
			return
		}
	}
	if req.EndTime != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    errors.ErrInvalidParam,
				"message": "结束时间格式错误",
			})
			return
		}
	}

	// 构建统计请求
	statsReq := &audit.StatisticsRequest{
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   req.GroupBy,
	}

	stats, err := h.auditService.GetStatistics(c.Request.Context(), statsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// CleanExpiredLogsRequest 清理过期日志请求
type CleanExpiredLogsRequest struct {
	BeforeDate string `json:"beforeDate" binding:"required"` // 格式: 2006-01-02
}

// CleanExpiredLogs 清理过期日志
func (h *AuditHandler) CleanExpiredLogs(c *gin.Context) {
	var req CleanExpiredLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 转换日期参数
	beforeDate, err := time.Parse("2006-01-02", req.BeforeDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    errors.ErrInvalidParam,
			"message": "日期格式错误,应为: YYYY-MM-DD",
		})
		return
	}

	count, err := h.auditService.CleanExpiredLogs(c.Request.Context(), beforeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrDatabase,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"deletedCount": count,
		},
	})
}
