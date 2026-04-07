package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
)

// LiveStreamHandler 直播流处理器
type LiveStreamHandler struct {
	liveService liveService.Service
}

// NewLiveStreamHandler 创建直播流处理器
func NewLiveStreamHandler(service liveService.Service) *LiveStreamHandler {
	return &LiveStreamHandler{
		liveService: service,
	}
}

// GetOnlineStreams 查询在线流列表
// @Summary 查询在线流列表
// @Description 查询正在直播中的流列表
// @Tags Live-Stream
// @Accept json
// @Produce json
// @Param domainName query string false "域名"
// @Param appName query string false "应用名称"
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(10)
// @Success 200 {object} utils.Response{data=live.DescribeStreamOnlineListResponse}
// @Router /api/v1/live/streams/online [get]
func (h *LiveStreamHandler) GetOnlineStreams(c *gin.Context) {
	// 获取查询参数
	domainName := c.Query("domainName")
	appName := c.Query("appName")
	pageNumStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 转换分页参数
	pageNum, err := strconv.ParseInt(pageNumStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "页码格式错误")
		return
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "每页大小格式错误")
		return
	}

	// 构建请求
	req := &live.DescribeStreamOnlineListRequest{
		PageNum:  &pageNum,
		PageSize: &pageSize,
	}

	if domainName != "" {
		req.DomainName = &domainName
	}
	if appName != "" {
		req.AppName = &appName
	}

	// 调用服务
	ctx := withCompanyID(c)
	result, err := h.liveService.DescribeStreamOnlineList(ctx, req)
	if err != nil {
		zap.L().Error("查询在线流列表失败",
			zap.String("domainName", domainName),
			zap.String("appName", appName),
			zap.Error(err))
		utils.InternalError(c, "查询在线流列表失败")
		return
	}

	utils.Success(c, result)
}

// GetHistoryStreams 查询历史流列表
// @Summary 查询历史流列表
// @Description 查询历史推流记录
// @Tags Live-Stream
// @Accept json
// @Produce json
// @Param domainName query string false "域名"
// @Param appName query string false "应用名称"
// @Param streamName query string false "流名称"
// @Param startTime query string false "开始时间（格式：2019-01-01T00:00:00Z）"
// @Param endTime query string false "结束时间（格式：2019-01-01T23:59:59Z）"
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(10)
// @Success 200 {object} utils.Response{data=live.DescribeStreamHistoryListResponse}
// @Router /api/v1/live/streams/history [get]
func (h *LiveStreamHandler) GetHistoryStreams(c *gin.Context) {
	// 获取查询参数
	domainName := c.Query("domainName")
	appName := c.Query("appName")
	streamName := c.Query("streamName")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	pageNumStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 验证必填参数
	if domainName == "" {
		utils.BadRequest(c, "域名不能为空")
		return
	}
	if startTime == "" {
		utils.BadRequest(c, "开始时间不能为空")
		return
	}
	if endTime == "" {
		utils.BadRequest(c, "结束时间不能为空")
		return
	}

	// 转换分页参数
	pageNum, err := strconv.ParseInt(pageNumStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "页码格式错误")
		return
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "每页大小格式错误")
		return
	}

	// 构建请求
	req := &live.DescribeStreamHistoryListRequest{
		DomainName: domainName,
		StartTime:  startTime,
		EndTime:    endTime,
		PageNum:    &pageNum,
		PageSize:   &pageSize,
	}

	if appName != "" {
		req.AppName = &appName
	}
	if streamName != "" {
		req.StreamName = &streamName
	}

	// 调用服务
	ctx := withCompanyID(c)
	result, err := h.liveService.DescribeStreamHistoryList(ctx, req)
	if err != nil {
		zap.L().Error("查询历史流列表失败",
			zap.String("domainName", domainName),
			zap.String("appName", appName),
			zap.String("streamName", streamName),
			zap.Error(err))
		utils.InternalError(c, "查询历史流列表失败")
		return
	}

	utils.Success(c, result)
}

// DropLiveStreamRequest 断开推流请求
type DropLiveStreamRequest struct {
	StreamName string `json:"streamName" binding:"required"` // 流名称
	DomainName string `json:"domainName" binding:"required"` // 推流域名
	AppName    string `json:"appName" binding:"required"`    // 推流路径
}

// DropStream 断开直播推流
// @Summary 断开直播推流
// @Description 断开指定的直播推流
// @Tags Live-Stream
// @Accept json
// @Produce json
// @Param request body DropLiveStreamRequest true "断开推流请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/streams/drop [post]
func (h *LiveStreamHandler) DropStream(c *gin.Context) {
	var req DropLiveStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 构建请求
	dropReq := &live.DropLiveStreamRequest{
		StreamName: req.StreamName,
		DomainName: req.DomainName,
		AppName:    req.AppName,
	}

	// 调用服务
	ctx := withCompanyID(c)
	err := h.liveService.DropLiveStream(ctx, dropReq)
	if err != nil {
		zap.L().Error("断开推流失败",
			zap.String("domainName", req.DomainName),
			zap.String("appName", req.AppName),
			zap.String("streamName", req.StreamName),
			zap.Error(err))
		utils.InternalError(c, "断开推流失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "断开推流成功",
	})
}

// GetStreamEvents 查询推断流事件
// @Summary 查询推断流事件
// @Description 查询推流断开的历史记录
// @Tags Live-Stream
// @Accept json
// @Produce json
// @Param startTime query string true "起始时间（UTC格式，例如：2018-12-29T19:00:00Z）"
// @Param endTime query string true "结束时间（UTC格式，例如：2018-12-29T20:00:00Z）"
// @Param domainName query string false "推流域名"
// @Param appName query string false "应用名称"
// @Param streamName query string false "流名称"
// @Param isFilter query int false "是否过滤（0-不过滤，1-只返回开播成功的）"
// @Param pageNum query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(10)
// @Success 200 {object} utils.Response{data=live.DescribeStreamEventListResponse}
// @Router /api/v1/live/streams/events [get]
func (h *LiveStreamHandler) GetStreamEvents(c *gin.Context) {
	// 获取查询参数
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")
	domainName := c.Query("domainName")
	appName := c.Query("appName")
	streamName := c.Query("streamName")
	isFilterStr := c.Query("isFilter")
	pageNumStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 验证必填参数
	if startTime == "" || endTime == "" {
		utils.BadRequest(c, "起始时间和结束时间不能为空")
		return
	}

	// 转换分页参数
	pageNum, err := strconv.ParseInt(pageNumStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "页码格式错误")
		return
	}

	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "每页大小格式错误")
		return
	}

	// 构建请求
	req := &live.DescribeStreamEventListRequest{
		StartTime: startTime,
		EndTime:   endTime,
		PageNum:   &pageNum,
		PageSize:  &pageSize,
	}

	if domainName != "" {
		req.DomainName = &domainName
	}
	if appName != "" {
		req.AppName = &appName
	}
	if streamName != "" {
		req.StreamName = &streamName
	}
	if isFilterStr != "" {
		isFilter, err := strconv.ParseInt(isFilterStr, 10, 64)
		if err == nil {
			req.IsFilter = &isFilter
		}
	}

	// 调用服务
	ctx := withCompanyID(c)
	result, err := h.liveService.DescribeStreamEventList(ctx, req)
	if err != nil {
		zap.L().Error("查询推断流事件失败",
			zap.String("startTime", startTime),
			zap.String("endTime", endTime),
			zap.String("domainName", domainName),
			zap.Error(err))
		utils.InternalError(c, "查询推断流事件失败")
		return
	}

	utils.Success(c, result)
}

// ResumeLiveStreamRequest 恢复推流请求
type ResumeLiveStreamRequest struct {
	StreamName string `json:"streamName" binding:"required"` // 流名称
	DomainName string `json:"domainName" binding:"required"` // 推流域名
	AppName    string `json:"appName" binding:"required"`    // 推流路径
}

// ResumeStream 恢复直播推流
// @Summary 恢复直播推流
// @Description 恢复之前被断开的直播推流
// @Tags Live-Stream
// @Accept json
// @Produce json
// @Param request body ResumeLiveStreamRequest true "恢复推流请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/streams/resume [post]
func (h *LiveStreamHandler) ResumeStream(c *gin.Context) {
	var req ResumeLiveStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 构建请求
	resumeReq := &live.ResumeLiveStreamRequest{
		StreamName: req.StreamName,
		DomainName: req.DomainName,
		AppName:    req.AppName,
	}

	// 调用服务
	ctx := withCompanyID(c)
	err := h.liveService.ResumeLiveStream(ctx, resumeReq)
	if err != nil {
		zap.L().Error("恢复推流失败",
			zap.String("domainName", req.DomainName),
			zap.String("appName", req.AppName),
			zap.String("streamName", req.StreamName),
			zap.Error(err))
		utils.InternalError(c, "恢复推流失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "恢复推流成功",
	})
}
