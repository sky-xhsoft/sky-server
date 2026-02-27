package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

// LiveRoomHandler 直播间处理器
type LiveRoomHandler struct {
	roomService live.LiveRoomService
}

// NewLiveRoomHandler 创建直播间处理器实例
func NewLiveRoomHandler(roomService live.LiveRoomService) *LiveRoomHandler {
	return &LiveRoomHandler{
		roomService: roomService,
	}
}

// CreateRoomRequest 创建直播间请求
type CreateRoomRequest struct {
	RoomName          string     `json:"roomName" binding:"required"`
	RoomType          string     `json:"roomType" binding:"required"`
	BroadcastFormat   string     `json:"broadcastFormat" binding:"required"`
	RoomStage         string     `json:"roomStage" binding:"required"`
	DisplayMode       string     `json:"displayMode"`
	StartTime         *time.Time `json:"startTime"`
	CoverImage        string     `json:"coverImage"`
	ViewingMethod     string     `json:"viewingMethod"`
	ViewingPassword   string     `json:"viewingPassword"`
	ViewingPrice      *float64   `json:"viewingPrice"`
	PlaybackMethod    string     `json:"playbackMethod"`
	PlaybackValidity  string     `json:"playbackValidity"`
	PlaybackStartTime *time.Time `json:"playbackStartTime"`
	PlaybackEndTime   *time.Time `json:"playbackEndTime"`
	Description       string     `json:"description"`
}

// UpdateRoomRequest 更新直播间请求
type UpdateRoomRequest struct {
	ID                uint       `json:"id" binding:"required"`
	RoomName          string     `json:"roomName"`
	DisplayMode       string     `json:"displayMode"`
	StartTime         *time.Time `json:"startTime"`
	CoverImage        string     `json:"coverImage"`
	ViewingMethod     string     `json:"viewingMethod"`
	ViewingPassword   string     `json:"viewingPassword"`
	ViewingPrice      *float64   `json:"viewingPrice"`
	PlaybackMethod    string     `json:"playbackMethod"`
	PlaybackValidity  string     `json:"playbackValidity"`
	PlaybackStartTime *time.Time `json:"playbackStartTime"`
	PlaybackEndTime   *time.Time `json:"playbackEndTime"`
	Description       string     `json:"description"`
}

// ListRoomsRequest 查询直播间列表请求
type ListRoomsRequest struct {
	RoomType      string `form:"roomType"`
	Status        string `form:"status"`
	RoomStage     string `form:"roomStage"`
	ViewingMethod string `form:"viewingMethod"`
	Keyword       string `form:"keyword"`
	StartTimeFrom string `form:"startTimeFrom"`
	StartTimeTo   string `form:"startTimeTo"`
	Page          int    `form:"page"`
	PageSize      int    `form:"pageSize"`
}

// CreateRoom 创建直播间
// @Summary 创建直播间
// @Tags 直播间管理
// @Accept json
// @Produce json
// @Param request body CreateRoomRequest true "创建直播间请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms [post]
func (h *LiveRoomHandler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取公司ID（从上下文中获取）
	companyID, _ := c.Get("companyId")
	userID, _ := c.Get("userId")

	room := &entity.LiveRoom{
		BaseModel: entity.BaseModel{
			SysCompanyID: companyID.(uint),
			CreateBy:     userID.(string),
		},
		RoomName:         req.RoomName,
		RoomType:         req.RoomType,
		BroadcastFormat:  req.BroadcastFormat,
		RoomStage:        req.RoomStage,
		DisplayMode:      req.DisplayMode,
		StartTime:        req.StartTime,
		CoverImage:       req.CoverImage,
		ViewingMethod:    req.ViewingMethod,
		ViewingPassword:  req.ViewingPassword,
		ViewingPrice:     req.ViewingPrice,
		PlaybackMethod:   req.PlaybackMethod,
		PlaybackValidity: req.PlaybackValidity,
		Description:      req.Description,
	}

	if err := h.roomService.CreateRoom(c.Request.Context(), room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    room,
	})
}

// UpdateRoom 更新直播间
// @Summary 更新直播间
// @Tags 直播间管理
// @Accept json
// @Produce json
// @Param request body UpdateRoomRequest true "更新直播间请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms [put]
func (h *LiveRoomHandler) UpdateRoom(c *gin.Context) {
	var req UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userId")

	room := &entity.LiveRoom{
		BaseModel: entity.BaseModel{
			ID:       req.ID,
			UpdateBy: userID.(string),
		},
		RoomName:         req.RoomName,
		DisplayMode:      req.DisplayMode,
		StartTime:        req.StartTime,
		CoverImage:       req.CoverImage,
		ViewingMethod:    req.ViewingMethod,
		ViewingPassword:  req.ViewingPassword,
		ViewingPrice:     req.ViewingPrice,
		PlaybackMethod:   req.PlaybackMethod,
		PlaybackValidity: req.PlaybackValidity,
		Description:      req.Description,
	}

	if err := h.roomService.UpdateRoom(c.Request.Context(), room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// DeleteRoom 删除直播间
// @Summary 删除直播间
// @Tags 直播间管理
// @Produce json
// @Param id path int true "直播间ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms/{id} [delete]
func (h *LiveRoomHandler) DeleteRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.roomService.DeleteRoom(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// GetRoom 获取直播间详情
// @Summary 获取直播间详情
// @Tags 直播间管理
// @Produce json
// @Param id path int true "直播间ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms/{id} [get]
func (h *LiveRoomHandler) GetRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	room, err := h.roomService.GetRoom(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": room,
	})
}

// ListRooms 查询直播间列表
// @Summary 查询直播间列表
// @Tags 直播间管理
// @Produce json
// @Param roomType query string false "直播间类型"
// @Param status query string false "状态"
// @Param roomStage query string false "直播间阶段"
// @Param viewingMethod query string false "观看方式"
// @Param keyword query string false "搜索关键词"
// @Param startTimeFrom query string false "开始时间起"
// @Param startTimeTo query string false "开始时间止"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms [get]
func (h *LiveRoomHandler) ListRooms(c *gin.Context) {
	var req ListRoomsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 获取公司ID
	companyID, _ := c.Get("companyId")

	filter := &live.RoomFilter{
		CompanyID:     companyID.(uint),
		RoomType:      req.RoomType,
		Status:        req.Status,
		RoomStage:     req.RoomStage,
		ViewingMethod: req.ViewingMethod,
		Keyword:       req.Keyword,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}

	// 解析时间范围
	if req.StartTimeFrom != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.StartTimeFrom); err == nil {
			filter.StartTimeFrom = &t
		}
	}
	if req.StartTimeTo != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.StartTimeTo); err == nil {
			filter.StartTimeTo = &t
		}
	}

	rooms, total, err := h.roomService.ListRooms(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rooms,
		"pagination": gin.H{
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	})
}

// StartLive 开始直播
// @Summary 开始直播
// @Tags 直播间管理
// @Produce json
// @Param id path int true "直播间ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms/{id}/start [post]
func (h *LiveRoomHandler) StartLive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.roomService.StartLive(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "直播已开始",
	})
}

// EndLive 结束直播
// @Summary 结束直播
// @Tags 直播间管理
// @Produce json
// @Param id path int true "直播间ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms/{id}/end [post]
func (h *LiveRoomHandler) EndLive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.roomService.EndLive(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "直播已结束",
	})
}

// UpdateViewerCount 更新观看人数
// @Summary 更新观看人数
// @Tags 直播间管理
// @Accept json
// @Produce json
// @Param id path int true "直播间ID"
// @Param request body map[string]int true "观看人数"
// @Success 200 {object} map[string]interface{}
// @Router /api/live/rooms/{id}/viewer-count [put]
func (h *LiveRoomHandler) UpdateViewerCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Count int `json:"count" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roomService.UpdateViewerCount(c.Request.Context(), uint(id), req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}
