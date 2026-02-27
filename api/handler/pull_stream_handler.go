package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
)

// PullStreamHandler 拉流任务处理器
type PullStreamHandler struct {
	liveService       liveService.Service
	pullStreamTaskService liveService.PullStreamTaskService
}

// NewPullStreamHandler 创建拉流任务处理器
func NewPullStreamHandler(service liveService.Service, pullStreamTaskService liveService.PullStreamTaskService) *PullStreamHandler {
	return &PullStreamHandler{
		liveService:       service,
		pullStreamTaskService: pullStreamTaskService,
	}
}

// CreatePullStreamTask 创建拉流任务
// @Summary 创建拉流任务
// @Description 创建拉流转推任务
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param request body live.CreatePullStreamTaskRequest true "创建拉流任务请求"
// @Success 200 {object} utils.Response{data=string}
// @Router /api/v1/live/pull-stream/tasks [post]
func (h *PullStreamHandler) CreatePullStreamTask(c *gin.Context) {
	var req live.CreatePullStreamTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.SourceType == "" {
		utils.BadRequest(c, "拉流源类型不能为空")
		return
	}
	if len(req.SourceURLs) == 0 {
		utils.BadRequest(c, "拉流源URL不能为空")
		return
	}
	// 验证目标地址相关字段：要么提供完整的 ToUrl，要么提供 DomainName、AppName、StreamName
	if req.ToUrl == "" {
		if req.DomainName == "" {
			utils.BadRequest(c, "推流域名不能为空")
			return
		}
		if req.AppName == "" {
			utils.BadRequest(c, "推流应用名称不能为空")
			return
		}
		if req.StreamName == "" {
			utils.BadRequest(c, "推流流名称不能为空")
			return
		}
	} else {
		// 如果提供了 ToUrl，则 DomainName、AppName、StreamName 应该是空字符串
		// 这里不强制要求，但建议按照腾讯云 API 文档的建议设置为空字符串
	}
	if req.StartTime == "" {
		utils.BadRequest(c, "开始时间不能为空")
		return
	}
	if req.EndTime == "" {
		utils.BadRequest(c, "结束时间不能为空")
		return
	}
	if req.Operator == "" {
		utils.BadRequest(c, "操作者不能为空")
		return
	}

	// 创建腾讯云拉流任务
	taskID, err := h.liveService.CreatePullStreamTask(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("创建拉流任务失败", zap.Error(err))
		utils.InternalError(c, "创建拉流任务失败: "+err.Error())
		return
	}

	// 保存到本地数据库
	var targetURL string
	if req.ToUrl != "" {
		targetURL = req.ToUrl
	} else {
		// 如果没有提供完整的 ToUrl，则拼接成 RTMP URL
		targetURL = "rtmp://" + req.DomainName + "/" + req.AppName + "/" + req.StreamName
		if req.PushArgs != "" {
			targetURL += "?" + req.PushArgs
		}
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		zap.L().Error("解析开始时间失败", zap.Error(err))
		utils.InternalError(c, "解析开始时间失败: "+err.Error())
		return
	}

	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		zap.L().Error("解析结束时间失败", zap.Error(err))
		utils.InternalError(c, "解析结束时间失败: "+err.Error())
		return
	}

	task := &entity.PullStreamTask{
		TaskID:     taskID,
		Comment:    req.Comment,
		Region:     req.Region,
		SourceType: req.SourceType,
		SourceURL:  req.SourceURLs[0], // 只保存第一个源地址
		TargetURL:  targetURL,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     "enable",
		Operator:   req.Operator,
		IsActive:   "Y",
		RoomID:     req.RoomID,
		RoomName:   req.RoomName,
	}

	if err := h.pullStreamTaskService.CreatePullStreamTask(c.Request.Context(), task); err != nil {
		zap.L().Error("保存拉流任务到数据库失败", zap.Error(err))
		// 这里不返回错误，因为腾讯云任务已经创建成功
	}

	utils.Success(c, taskID)
}

// GetPullStreamTasks 查询拉流任务列表
// @Summary 查询拉流任务列表
// @Description 查询拉流转推任务列表
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param taskId query string false "任务ID"
// @Param pageNum query int false "取得第几页，默认值：1"
// @Param pageSize query int false "分页大小，默认值：10，取值范围：1~20"
// @Param specifyTaskId query string false "使用指定任务 ID 查询任务信息"
// @Success 200 {object} utils.Response{data=live.DescribePullStreamTasksResponse}
// @Router /api/v1/live/pull-stream/tasks [get]
func (h *PullStreamHandler) GetPullStreamTasks(c *gin.Context) {
	var req live.DescribePullStreamTasksRequest

	// 从查询参数中获取值
	taskId := c.Query("taskId")
	if taskId != "" {
		req.TaskID = &taskId
	}

	pageNumStr := c.Query("pageNum")
	if pageNumStr != "" {
		pageNum, err := strconv.ParseUint(pageNumStr, 10, 64)
		if err == nil {
			req.PageNum = &pageNum
		}
	}

	pageSizeStr := c.Query("pageSize")
	if pageSizeStr != "" {
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err == nil {
			// 确保 pageSize 在 1~20 之间
			if pageSize < 1 {
				pageSize = 1
			} else if pageSize > 20 {
				pageSize = 20
			}
			req.PageSize = &pageSize
		}
	}

	specifyTaskId := c.Query("specifyTaskId")
	if specifyTaskId != "" {
		req.SpecifyTaskId = &specifyTaskId
	}

	// 首先从腾讯云查询任务
	tasks, err := h.liveService.DescribePullStreamTasks(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("查询拉流任务列表失败", zap.Error(err))
		utils.InternalError(c, "查询拉流任务列表失败: "+err.Error())
		return
	}

	// 如果有任务ID参数，并且在数据库中没有找到该任务，则尝试从腾讯云同步
	if specifyTaskId != "" {
		_, err := h.pullStreamTaskService.GetPullStreamTask(c.Request.Context(), specifyTaskId)
		if err != nil {
			// 如果任务在数据库中不存在，则尝试同步
			if len(tasks.TaskInfos) > 0 {
				taskInfo := tasks.TaskInfos[0]
				// 解析时间
				startTime, _ := time.Parse("2006-01-02 15:04:05", taskInfo.StartTime)
				endTime, _ := time.Parse("2006-01-02 15:04:05", taskInfo.EndTime)

				// 保存到数据库
				task := &entity.PullStreamTask{
					TaskID:     taskInfo.TaskID,
					Comment:    taskInfo.Comment,
					Region:     taskInfo.Region,
					SourceType: taskInfo.SourceType,
					SourceURL:  taskInfo.SourceURLs[0], // 只保存第一个源地址
					TargetURL:  taskInfo.ToUrl,
					StartTime:  startTime,
					EndTime:    endTime,
					Status:     taskInfo.Status,
					Operator:   taskInfo.CreateBy,
					IsActive:   "Y",
				}

				if err := h.pullStreamTaskService.CreatePullStreamTask(c.Request.Context(), task); err != nil {
					zap.L().Error("同步拉流任务到数据库失败", zap.Error(err))
				}
			}
		}
	}

	// 为每个任务添加 RoomID 和 RoomName 信息
	for _, taskInfo := range tasks.TaskInfos {
		// 从本地数据库中获取任务信息
		localTask, err := h.pullStreamTaskService.GetPullStreamTask(c.Request.Context(), taskInfo.TaskID)
		if err == nil {
			taskInfo.RoomID = localTask.RoomID
			taskInfo.RoomName = localTask.RoomName
		}
	}

	utils.Success(c, tasks)
}

// UpdatePullStreamTask 更新拉流任务
// @Summary 更新拉流任务
// @Description 更新拉流转推任务
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param request body live.UpdatePullStreamTaskRequest true "更新拉流任务请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/pull-stream/tasks/{id} [put]
func (h *PullStreamHandler) UpdatePullStreamTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	var req live.UpdatePullStreamTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置任务ID
	req.TaskID = taskID

	// 验证必填字段
	if req.Operator == "" {
		utils.BadRequest(c, "操作者不能为空")
		return
	}

	// 更新腾讯云拉流任务
	err := h.liveService.UpdatePullStreamTask(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("更新拉流任务失败", zap.Error(err))
		utils.InternalError(c, "更新拉流任务失败: "+err.Error())
		return
	}

	// 更新本地数据库
	task := &entity.PullStreamTask{
		TaskID:     taskID,
		Comment:    req.Comment,
		Operator:   req.Operator,
		RoomID:     req.RoomID,
		RoomName:   req.RoomName,
	}

	if req.SourceURLs != nil && len(req.SourceURLs) > 0 {
		task.SourceURL = req.SourceURLs[0]
	}

	if req.ToUrl != "" {
		task.TargetURL = req.ToUrl
	}

	if req.StartTime != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err == nil {
			task.StartTime = startTime
		}
	}

	if req.EndTime != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err == nil {
			task.EndTime = endTime
		}
	}

	if req.Status != "" {
		task.Status = req.Status
	}

	if err := h.pullStreamTaskService.UpdatePullStreamTask(c.Request.Context(), task); err != nil {
		zap.L().Error("更新拉流任务到数据库失败", zap.Error(err))
		// 这里不返回错误，因为腾讯云任务已经更新成功
	}

	utils.Success(c, nil)
}

// DeletePullStreamTask 删除拉流任务
// @Summary 删除拉流任务
// @Description 删除拉流转推任务
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param operator query string true "操作者"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/pull-stream/tasks/{id} [delete]
func (h *PullStreamHandler) DeletePullStreamTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	operator := c.Query("operator")
	if operator == "" {
		utils.BadRequest(c, "操作者不能为空")
		return
	}

	// 删除腾讯云拉流任务
	err := h.liveService.DeletePullStreamTask(c.Request.Context(), taskID, operator)
	if err != nil {
		zap.L().Error("删除拉流任务失败", zap.Error(err))
		utils.InternalError(c, "删除拉流任务失败: "+err.Error())
		return
	}

	// 更新本地数据库（软删除）
	if err := h.pullStreamTaskService.DeletePullStreamTask(c.Request.Context(), taskID); err != nil {
		zap.L().Error("删除拉流任务到数据库失败", zap.Error(err))
		// 这里不返回错误，因为腾讯云任务已经删除成功
	}

	utils.Success(c, nil)
}

// GetPullStreamTaskStatus 查询拉流任务状态
// @Summary 查询拉流任务状态
// @Description 查询拉流转推任务状态
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response{data=live.PullStreamTaskStatus}
// @Router /api/v1/live/pull-stream/tasks/{id}/status [get]
func (h *PullStreamHandler) GetPullStreamTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	status, err := h.liveService.DescribePullStreamTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		zap.L().Error("查询拉流任务状态失败", zap.Error(err))
		utils.InternalError(c, "查询拉流任务状态失败: "+err.Error())
		return
	}

	utils.Success(c, status)
}

// RestartPullStreamTask 重启拉流任务
// @Summary 重启拉流任务
// @Description 重启拉流转推任务
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param operator query string true "操作者"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/pull-stream/tasks/{id}/restart [post]
func (h *PullStreamHandler) RestartPullStreamTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	operator := c.Query("operator")
	if operator == "" {
		utils.BadRequest(c, "操作者不能为空")
		return
	}

	err := h.liveService.RestartPullStreamTask(c.Request.Context(), taskID, operator)
	if err != nil {
		zap.L().Error("重启拉流任务失败", zap.Error(err))
		utils.InternalError(c, "重启拉流任务失败: "+err.Error())
		return
	}

	utils.Success(c, nil)
}

// DescribePullTransformPushInfoList 查询拉流转推任务流数据
// @Summary 查询拉流转推任务流数据
// @Description 查询拉流转推任务流数据统计信息
// @Tags Live-PullStream
// @Accept json
// @Produce json
// @Param request body live.DescribePullTransformPushInfoListRequest true "查询拉流转推任务流数据请求"
// @Success 200 {object} utils.Response{data=live.DescribePullTransformPushInfoListResponse}
// @Router /api/v1/live/pull-stream/tasks/transform-push-info [post]
func (h *PullStreamHandler) DescribePullTransformPushInfoList(c *gin.Context) {
	var req live.DescribePullTransformPushInfoListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.StartTime == "" {
		utils.BadRequest(c, "开始时间不能为空")
		return
	}
	if req.EndTime == "" {
		utils.BadRequest(c, "结束时间不能为空")
		return
	}
	if req.TaskId == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	resp, err := h.liveService.DescribePullTransformPushInfoList(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("查询拉流转推任务流数据失败", zap.Error(err))
		utils.InternalError(c, "查询拉流转推任务流数据失败: "+err.Error())
		return
	}

	utils.Success(c, resp)
}
