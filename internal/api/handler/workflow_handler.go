package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/workflow"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService workflow.Service
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowService workflow.Service) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
	}
}

// CreateDefinition 创建流程定义
// @Summary 创建流程定义
// @Tags 工作流
// @Accept json
// @Produce json
// @Param request body entity.WfDefinition true "流程定义"
// @Success 200 {object} entity.WfDefinition
// @Router /api/v1/workflow/definitions [post]
func (h *WorkflowHandler) CreateDefinition(c *gin.Context) {
	var def entity.WfDefinition
	if err := c.ShouldBindJSON(&def); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	def.IsActive = "Y"
	if err := h.workflowService.CreateDefinition(c.Request.Context(), &def); err != nil {
		utils.InternalError(c, "创建流程定义失败: "+err.Error())
		return
	}

	utils.Success(c, def)
}

// GetDefinition 获取流程定义
// @Summary 获取流程定义
// @Tags 工作流
// @Produce json
// @Param id path int true "流程定义ID"
// @Success 200 {object} entity.WfDefinition
// @Router /api/v1/workflow/definitions/{id} [get]
func (h *WorkflowHandler) GetDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	def, err := h.workflowService.GetDefinition(c.Request.Context(), uint(id))
	if err != nil {
		utils.NotFound(c, "流程定义不存在")
		return
	}

	utils.Success(c, def)
}

// UpdateDefinition 更新流程定义
// @Summary 更新流程定义
// @Tags 工作流
// @Accept json
// @Produce json
// @Param id path int true "流程定义ID"
// @Param request body entity.WfDefinition true "流程定义"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/definitions/{id} [put]
func (h *WorkflowHandler) UpdateDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var def entity.WfDefinition
	if err := c.ShouldBindJSON(&def); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	def.ID = uint(id)
	if err := h.workflowService.UpdateDefinition(c.Request.Context(), &def); err != nil {
		utils.InternalError(c, "更新流程定义失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "更新成功"})
}

// PublishDefinition 发布流程定义
// @Summary 发布流程定义
// @Tags 工作流
// @Produce json
// @Param id path int true "流程定义ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/definitions/{id}/publish [post]
func (h *WorkflowHandler) PublishDefinition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	if err := h.workflowService.PublishDefinition(c.Request.Context(), uint(id)); err != nil {
		utils.InternalError(c, "发布流程定义失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "发布成功"})
}

// ListDefinitions 查询流程定义列表
// @Summary 查询流程定义列表
// @Tags 工作流
// @Produce json
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/definitions [get]
func (h *WorkflowHandler) ListDefinitions(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	defs, total, err := h.workflowService.ListDefinitions(c.Request.Context(), status, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询流程定义列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     defs,
	})
}

// CreateNode 创建流程节点
// @Summary 创建流程节点
// @Tags 工作流
// @Accept json
// @Produce json
// @Param request body entity.WfNode true "流程节点"
// @Success 200 {object} entity.WfNode
// @Router /api/v1/workflow/nodes [post]
func (h *WorkflowHandler) CreateNode(c *gin.Context) {
	var node entity.WfNode
	if err := c.ShouldBindJSON(&node); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	node.IsActive = "Y"
	if err := h.workflowService.CreateNode(c.Request.Context(), &node); err != nil {
		utils.InternalError(c, "创建流程节点失败: "+err.Error())
		return
	}

	utils.Success(c, node)
}

// GetNodes 获取流程节点列表
// @Summary 获取流程节点列表
// @Tags 工作流
// @Produce json
// @Param definitionId query int true "流程定义ID"
// @Success 200 {array} entity.WfNode
// @Router /api/v1/workflow/nodes [get]
func (h *WorkflowHandler) GetNodes(c *gin.Context) {
	definitionIDStr := c.Query("definitionId")
	definitionID, err := strconv.ParseUint(definitionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "流程定义ID格式错误")
		return
	}

	nodes, err := h.workflowService.GetNodes(c.Request.Context(), uint(definitionID))
	if err != nil {
		utils.InternalError(c, "查询流程节点失败: "+err.Error())
		return
	}

	utils.Success(c, nodes)
}

// UpdateNode 更新流程节点
// @Summary 更新流程节点
// @Tags 工作流
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Param request body entity.WfNode true "流程节点"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/nodes/{id} [put]
func (h *WorkflowHandler) UpdateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var node entity.WfNode
	if err := c.ShouldBindJSON(&node); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	node.ID = uint(id)
	if err := h.workflowService.UpdateNode(c.Request.Context(), &node); err != nil {
		utils.InternalError(c, "更新流程节点失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "更新成功"})
}

// DeleteNode 删除流程节点
// @Summary 删除流程节点
// @Tags 工作流
// @Produce json
// @Param id path int true "节点ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/nodes/{id} [delete]
func (h *WorkflowHandler) DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	if err := h.workflowService.DeleteNode(c.Request.Context(), uint(id)); err != nil {
		utils.InternalError(c, "删除流程节点失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// CreateTransition 创建流程流转
// @Summary 创建流程流转
// @Tags 工作流
// @Accept json
// @Produce json
// @Param request body entity.WfTransition true "流程流转"
// @Success 200 {object} entity.WfTransition
// @Router /api/v1/workflow/transitions [post]
func (h *WorkflowHandler) CreateTransition(c *gin.Context) {
	var transition entity.WfTransition
	if err := c.ShouldBindJSON(&transition); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	transition.IsActive = "Y"
	if err := h.workflowService.CreateTransition(c.Request.Context(), &transition); err != nil {
		utils.InternalError(c, "创建流程流转失败: "+err.Error())
		return
	}

	utils.Success(c, transition)
}

// GetTransitions 获取流程流转列表
// @Summary 获取流程流转列表
// @Tags 工作流
// @Produce json
// @Param definitionId query int true "流程定义ID"
// @Success 200 {array} entity.WfTransition
// @Router /api/v1/workflow/transitions [get]
func (h *WorkflowHandler) GetTransitions(c *gin.Context) {
	definitionIDStr := c.Query("definitionId")
	definitionID, err := strconv.ParseUint(definitionIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "流程定义ID格式错误")
		return
	}

	transitions, err := h.workflowService.GetTransitions(c.Request.Context(), uint(definitionID))
	if err != nil {
		utils.InternalError(c, "查询流程流转失败: "+err.Error())
		return
	}

	utils.Success(c, transitions)
}

// DeleteTransition 删除流程流转
// @Summary 删除流程流转
// @Tags 工作流
// @Produce json
// @Param id path int true "流转ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/transitions/{id} [delete]
func (h *WorkflowHandler) DeleteTransition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	if err := h.workflowService.DeleteTransition(c.Request.Context(), uint(id)); err != nil {
		utils.InternalError(c, "删除流程流转失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// StartProcess 启动流程
// @Summary 启动流程
// @Tags 工作流
// @Accept json
// @Produce json
// @Param request body workflow.StartProcessRequest true "启动流程请求"
// @Success 200 {object} entity.WfInstance
// @Router /api/v1/workflow/instances/start [post]
func (h *WorkflowHandler) StartProcess(c *gin.Context) {
	var req workflow.StartProcessRequest
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
	req.StartUserID = userID.(uint)

	instance, err := h.workflowService.StartProcess(c.Request.Context(), &req)
	if err != nil {
		utils.InternalError(c, "启动流程失败: "+err.Error())
		return
	}

	utils.Success(c, instance)
}

// GetInstance 获取流程实例
// @Summary 获取流程实例
// @Tags 工作流
// @Produce json
// @Param id path int true "流程实例ID"
// @Success 200 {object} entity.WfInstance
// @Router /api/v1/workflow/instances/{id} [get]
func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	instance, err := h.workflowService.GetInstance(c.Request.Context(), uint(id))
	if err != nil {
		utils.NotFound(c, "流程实例不存在")
		return
	}

	utils.Success(c, instance)
}

// ListInstances 查询流程实例列表
// @Summary 查询流程实例列表
// @Tags 工作流
// @Produce json
// @Param definitionId query int false "流程定义ID"
// @Param status query string false "状态"
// @Param startUserId query int false "发起人ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/instances [get]
func (h *WorkflowHandler) ListInstances(c *gin.Context) {
	var req workflow.ListInstancesRequest
	definitionIDStr := c.Query("definitionId")
	if definitionIDStr != "" {
		id, _ := strconv.ParseUint(definitionIDStr, 10, 32)
		req.DefinitionID = uint(id)
	}

	req.Status = c.Query("status")

	startUserIDStr := c.Query("startUserId")
	if startUserIDStr != "" {
		id, _ := strconv.ParseUint(startUserIDStr, 10, 32)
		req.StartUserID = uint(id)
	}

	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	instances, total, err := h.workflowService.ListInstances(c.Request.Context(), &req)
	if err != nil {
		utils.InternalError(c, "查询流程实例列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"data":     instances,
	})
}

// TerminateInstance 终止流程实例
// @Summary 终止流程实例
// @Tags 工作流
// @Produce json
// @Param id path int true "流程实例ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/instances/{id}/terminate [post]
func (h *WorkflowHandler) TerminateInstance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.workflowService.TerminateInstance(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "终止流程失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "终止成功"})
}

// ListMyTasks 查询我的任务列表
// @Summary 查询我的任务列表
// @Tags 工作流
// @Produce json
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/tasks/my [get]
func (h *WorkflowHandler) ListMyTasks(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	tasks, total, err := h.workflowService.ListMyTasks(c.Request.Context(), userID.(uint), status, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询任务列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     tasks,
	})
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Tags 工作流
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} entity.WfTask
// @Router /api/v1/workflow/tasks/{id} [get]
func (h *WorkflowHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	task, err := h.workflowService.GetTask(c.Request.Context(), uint(id))
	if err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// CompleteTask 完成任务
// @Summary 完成任务
// @Tags 工作流
// @Accept json
// @Produce json
// @Param request body workflow.CompleteTaskRequest true "完成任务请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/tasks/complete [post]
func (h *WorkflowHandler) CompleteTask(c *gin.Context) {
	var req workflow.CompleteTaskRequest
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
	req.UserID = userID.(uint)

	if err := h.workflowService.CompleteTask(c.Request.Context(), &req); err != nil {
		utils.InternalError(c, "完成任务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "任务已完成"})
}

// ClaimTask 签收任务
// @Summary 签收任务
// @Tags 工作流
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/tasks/{id}/claim [post]
func (h *WorkflowHandler) ClaimTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未授权")
		return
	}

	if err := h.workflowService.ClaimTask(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		utils.InternalError(c, "签收任务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "签收成功"})
}

// TransferTask 转交任务
// @Summary 转交任务
// @Tags 工作流
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body TransferTaskRequest true "转交任务请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/tasks/{id}/transfer [post]
func (h *WorkflowHandler) TransferTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "ID格式错误")
		return
	}

	var req TransferTaskRequest
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

	if err := h.workflowService.TransferTask(c.Request.Context(), uint(id), userID.(uint), req.ToUserID, req.Comment); err != nil {
		utils.InternalError(c, "转交任务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "转交成功"})
}

// TransferTaskRequest 转交任务请求
type TransferTaskRequest struct {
	ToUserID uint   `json:"toUserId" binding:"required"`
	Comment  string `json:"comment"`
}
