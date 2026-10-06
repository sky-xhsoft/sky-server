package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/pkg/volcengine"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"go.uber.org/zap"
)

// SpeechTaskHandler 声影同传任务处理器
type SpeechTaskHandler struct {
	speechTaskService liveService.SpeechTaskService
}

// NewSpeechTaskHandler 创建声影同传任务处理器
func NewSpeechTaskHandler(speechTaskService liveService.SpeechTaskService) *SpeechTaskHandler {
	return &SpeechTaskHandler{speechTaskService: speechTaskService}
}

// getCompanyID 从请求中获取公司ID(由 dashboard_handler.go 提供)

// CreateSpeechTask 创建声影同传任务
// @Summary 创建声影同传任务
// @Description 调用火山引擎 CreateSpeechTask 创建声影同传任务
// @Tags Live-SpeechTask
// @Accept json
// @Produce json
// @Param request body volcengine.CreateSpeechTaskRequest true "创建声影同传任务请求"
// @Success 200 {object} utils.Response{data=string}
// @Router /api/v1/live/speech-tasks [post]
func (h *SpeechTaskHandler) CreateSpeechTask(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}

	var req volcengine.CreateSpeechTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Name == "" {
		utils.BadRequest(c, "任务名称不能为空")
		return
	}
	if req.Rule == nil {
		utils.BadRequest(c, "任务规则不能为空")
		return
	}
	if req.Rule.Pipe == "" {
		utils.BadRequest(c, "任务类型不能为空")
		return
	}
	if req.Rule.Source == nil || req.Rule.Source.Url == "" {
		utils.BadRequest(c, "拉流源地址不能为空")
		return
	}
	if req.Rule.Output == nil || len(req.Rule.Output.Url) == 0 || req.Rule.Output.Url[0] == "" {
		utils.BadRequest(c, "转推输出地址不能为空")
		return
	}
	if req.Rule.StartTime == "" {
		utils.BadRequest(c, "开始时间不能为空")
		return
	}
	if req.Rule.EndTime == "" {
		utils.BadRequest(c, "结束时间不能为空")
		return
	}

	// 临时调试日志：确认后端实际收到的 Pipe 类型与完整请求
	if b, err := json.Marshal(&req); err == nil {
		zap.L().Info("CreateSpeechTask 收到请求", zap.String("pipe", req.Rule.Pipe), zap.ByteString("body", b))
	}

	taskID, err := h.speechTaskService.CreateSpeechTask(c.Request.Context(), companyID, &req)
	if err != nil {
		zap.L().Error("创建声影同传任务失败", zap.Error(err))
		utils.FailFromError(c, "创建声影同传任务失败: ", err)
		return
	}

	utils.Success(c, taskID)
}

// DeleteSpeechTask 删除声影同传任务
// @Summary 删除声影同传任务
// @Description 调用火山引擎 DeleteSpeechTask 删除声影同传任务
// @Tags Live-SpeechTask
// @Accept json
// @Produce json
// @Param request body volcengine.DeleteSpeechTaskRequest true "删除声影同传任务请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/live/speech-tasks/delete [post]
func (h *SpeechTaskHandler) DeleteSpeechTask(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}

	var req volcengine.DeleteSpeechTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.TaskID == "" {
		utils.BadRequest(c, "任务ID不能为空")
		return
	}

	if err := h.speechTaskService.DeleteSpeechTask(c.Request.Context(), companyID, &req); err != nil {
		zap.L().Error("删除声影同传任务失败", zap.Error(err))
		utils.FailFromError(c, "删除声影同传任务失败: ", err)
		return
	}

	utils.Success(c, nil)
}

// ListSpeechTasks 查询声影同传任务列表
// @Summary 查询声影同传任务列表
// @Description 查询当前直播间下已创建的字幕同传任务
// @Tags Live-SpeechTask
// @Produce json
// @Param roomId query int false "直播间ID"
// @Success 200 {object} utils.Response{data=[]live.SpeechTaskRecord}
// @Router /api/v1/live/speech-tasks [get]
func (h *SpeechTaskHandler) ListSpeechTasks(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.Unauthorized(c, "未获取到公司信息")
		return
	}

	var roomID uint
	if v, err := strconv.ParseUint(c.Query("roomId"), 10, 32); err == nil {
		roomID = uint(v)
	}

	items, err := h.speechTaskService.ListSpeechTasks(c.Request.Context(), companyID, roomID)
	if err != nil {
		zap.L().Error("查询声影同传任务列表失败", zap.Error(err))
		utils.FailFromError(c, "查询声影同传任务列表失败: ", err)
		return
	}
	if items == nil {
		items = []liveService.SpeechTaskRecord{}
	}
	utils.Success(c, items)
}