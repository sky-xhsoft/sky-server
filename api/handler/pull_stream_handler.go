package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
)

// PullStreamHandler 拉流任务处理器
type PullStreamHandler struct {
	liveService liveService.Service
}

// NewPullStreamHandler 创建拉流任务处理器
func NewPullStreamHandler(service liveService.Service) *PullStreamHandler {
	return &PullStreamHandler{
		liveService: service,
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

	taskID, err := h.liveService.CreatePullStreamTask(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("创建拉流任务失败", zap.Error(err))
		utils.InternalError(c, "创建拉流任务失败: "+err.Error())
		return
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
// @Success 200 {object} utils.Response{data=[]live.PullStreamTaskInfo}
// @Router /api/v1/live/pull-stream/tasks [get]
func (h *PullStreamHandler) GetPullStreamTasks(c *gin.Context) {
	taskID := c.Query("taskId")

	var taskIDPtr *string
	if taskID != "" {
		taskIDPtr = &taskID
	}

	tasks, err := h.liveService.DescribePullStreamTasks(c.Request.Context(), taskIDPtr)
	if err != nil {
		zap.L().Error("查询拉流任务列表失败", zap.Error(err))
		utils.InternalError(c, "查询拉流任务列表失败: "+err.Error())
		return
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

	err := h.liveService.UpdatePullStreamTask(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("更新拉流任务失败", zap.Error(err))
		utils.InternalError(c, "更新拉流任务失败: "+err.Error())
		return
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

	err := h.liveService.DeletePullStreamTask(c.Request.Context(), taskID, operator)
	if err != nil {
		zap.L().Error("删除拉流任务失败", zap.Error(err))
		utils.InternalError(c, "删除拉流任务失败: "+err.Error())
		return
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
