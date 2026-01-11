package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
)

// ActionHandler 动作处理器
type ActionHandler struct {
	actionService action.Service
}

// NewActionHandler 创建动作处理器
func NewActionHandler(actionService action.Service) *ActionHandler {
	return &ActionHandler{
		actionService: actionService,
	}
}

// ExecuteAction 执行动作
// @Summary 执行动作
// @Description 根据动作ID执行动作
// @Tags 动作
// @Accept json
// @Produce json
// @Param actionId path int true "动作ID"
// @Param request body ExecuteActionRequest true "执行请求"
// @Success 200 {object} action.ActionResult
// @Router /api/v1/actions/{actionId}/execute [post]
func (h *ActionHandler) ExecuteAction(c *gin.Context) {
	actionIDStr := c.Param("actionId")
	actionID, err := strconv.ParseUint(actionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "动作ID格式错误")
		return
	}

	var req ExecuteActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.actionService.ExecuteAction(c.Request.Context(), uint(actionID), req.Params, userID.(uint))
	if err != nil {
		utils.InternalError(c, "执行动作失败: "+err.Error())
		return
	}

	if !result.Success {
		utils.InternalError(c, result.Error)
		return
	}

	utils.Success(c, result)
}

// ExecuteActionByName 根据名称执行动作
// @Summary 根据名称执行动作
// @Description 根据表名和动作名称执行动作
// @Tags 动作
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param actionName path string true "动作名称"
// @Param request body ExecuteActionRequest true "执行请求"
// @Success 200 {object} action.ActionResult
// @Router /api/v1/actions/{tableName}/{actionName}/execute [post]
func (h *ActionHandler) ExecuteActionByName(c *gin.Context) {
	tableName := c.Param("tableName")
	actionName := c.Param("actionName")

	if tableName == "" || actionName == "" {
		utils.BadRequest(c, "表名或动作名称不能为空")
		return
	}

	var req ExecuteActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	result, err := h.actionService.ExecuteActionByName(c.Request.Context(), tableName, actionName, req.Params, userID.(uint))
	if err != nil {
		utils.InternalError(c, "执行动作失败: "+err.Error())
		return
	}

	if !result.Success {
		utils.InternalError(c, result.Error)
		return
	}

	utils.Success(c, result)
}

// BatchExecuteAction 批量执行动作
// @Summary 批量执行动作
// @Description 批量执行指定动作
// @Tags 动作
// @Accept json
// @Produce json
// @Param actionId path int true "动作ID"
// @Param request body BatchExecuteActionRequest true "批量执行请求"
// @Success 200 {array} action.ActionResult
// @Router /api/v1/actions/{actionId}/batch-execute [post]
func (h *ActionHandler) BatchExecuteAction(c *gin.Context) {
	actionIDStr := c.Param("actionId")
	actionID, err := strconv.ParseUint(actionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "动作ID格式错误")
		return
	}

	var req BatchExecuteActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	results, err := h.actionService.BatchExecuteAction(c.Request.Context(), uint(actionID), req.BatchParams, userID.(uint))
	if err != nil {
		utils.InternalError(c, "批量执行失败: "+err.Error())
		return
	}

	utils.Success(c, results)
}

// GetAction 获取动作定义
// @Summary 获取动作定义
// @Description 根据动作ID获取动作定义
// @Tags 动作
// @Accept json
// @Produce json
// @Param actionId path int true "动作ID"
// @Success 200 {object} entity.SysAction
// @Router /api/v1/actions/{actionId} [get]
func (h *ActionHandler) GetAction(c *gin.Context) {
	actionIDStr := c.Param("actionId")
	actionID, err := strconv.ParseUint(actionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "动作ID格式错误")
		return
	}

	action, err := h.actionService.GetAction(c.Request.Context(), uint(actionID))
	if err != nil {
		utils.NotFound(c, "动作不存在")
		return
	}

	utils.Success(c, action)
}

// ExecuteActionRequest 执行动作请求
type ExecuteActionRequest struct {
	Params map[string]interface{} `json:"params"`
}

// BatchExecuteActionRequest 批量执行请求
type BatchExecuteActionRequest struct {
	BatchParams []map[string]interface{} `json:"batchParams" binding:"required"`
}
