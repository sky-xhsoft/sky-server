package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"gorm.io/gorm"
)

// Service 工作流服务接口
type Service interface {
	// 流程定义管理
	CreateDefinition(ctx context.Context, def *entity.WfDefinition) error
	GetDefinition(ctx context.Context, id uint) (*entity.WfDefinition, error)
	UpdateDefinition(ctx context.Context, def *entity.WfDefinition) error
	PublishDefinition(ctx context.Context, id uint) error
	ListDefinitions(ctx context.Context, status string, page, pageSize int) ([]*entity.WfDefinition, int64, error)

	// 流程节点管理
	CreateNode(ctx context.Context, node *entity.WfNode) error
	GetNodes(ctx context.Context, definitionID uint) ([]*entity.WfNode, error)
	UpdateNode(ctx context.Context, node *entity.WfNode) error
	DeleteNode(ctx context.Context, id uint) error

	// 流程流转管理
	CreateTransition(ctx context.Context, transition *entity.WfTransition) error
	GetTransitions(ctx context.Context, definitionID uint) ([]*entity.WfTransition, error)
	DeleteTransition(ctx context.Context, id uint) error

	// 流程实例管理
	StartProcess(ctx context.Context, req *StartProcessRequest) (*entity.WfInstance, error)
	GetInstance(ctx context.Context, id uint) (*entity.WfInstance, error)
	ListInstances(ctx context.Context, req *ListInstancesRequest) ([]*entity.WfInstance, int64, error)
	TerminateInstance(ctx context.Context, id uint, userID uint) error
	SuspendInstance(ctx context.Context, id uint) error
	ResumeInstance(ctx context.Context, id uint) error

	// 任务管理
	GetTask(ctx context.Context, id uint) (*entity.WfTask, error)
	ListMyTasks(ctx context.Context, userID uint, status string, page, pageSize int) ([]*entity.WfTask, int64, error)
	CompleteTask(ctx context.Context, req *CompleteTaskRequest) error
	ClaimTask(ctx context.Context, taskID, userID uint) error
	TransferTask(ctx context.Context, taskID, fromUserID, toUserID uint, comment string) error
}

// StartProcessRequest 启动流程请求
type StartProcessRequest struct {
	DefinitionID uint                   `json:"definitionId" binding:"required"`
	SysTableID   int                    `json:"sysTableId"`
	BusinessID   uint                   `json:"businessId"`
	StartUserID  uint                   `json:"startUserId" binding:"required"`
	Title        string                 `json:"title"`
	Variables    map[string]interface{} `json:"variables"`
}

// ListInstancesRequest 查询流程实例请求
type ListInstancesRequest struct {
	DefinitionID uint   `json:"definitionId"`
	Status       string `json:"status"`
	StartUserID  uint   `json:"startUserId"`
	Page         int    `json:"page"`
	PageSize     int    `json:"pageSize"`
}

// CompleteTaskRequest 完成任务请求
type CompleteTaskRequest struct {
	TaskID    uint                   `json:"taskId" binding:"required"`
	UserID    uint                   `json:"userId" binding:"required"`
	Action    string                 `json:"action" binding:"required"` // approve, reject
	Comment   string                 `json:"comment"`
	Variables map[string]interface{} `json:"variables"`
}

// service 工作流服务实现
type service struct {
	db            *gorm.DB
	actionService action.Service
}

// NewService 创建工作流服务
func NewService(db *gorm.DB, actionService action.Service) Service {
	return &service{
		db:            db,
		actionService: actionService,
	}
}

// CreateDefinition 创建流程定义
func (s *service) CreateDefinition(ctx context.Context, def *entity.WfDefinition) error {
	if def.Version == 0 {
		def.Version = 1
	}
	if def.Status == "" {
		def.Status = "draft"
	}

	if err := s.db.WithContext(ctx).Create(def).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建流程定义失败", err)
	}

	return nil
}

// GetDefinition 获取流程定义
func (s *service) GetDefinition(ctx context.Context, id uint) (*entity.WfDefinition, error) {
	var def entity.WfDefinition
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&def).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "流程定义不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询流程定义失败", err)
	}

	return &def, nil
}

// UpdateDefinition 更新流程定义
func (s *service) UpdateDefinition(ctx context.Context, def *entity.WfDefinition) error {
	// 只有草稿状态才能更新
	var existing entity.WfDefinition
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", def.ID, "Y").First(&existing).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询流程定义失败", err)
	}

	if existing.Status != "draft" {
		return errors.New(errors.ErrValidation, "只能更新草稿状态的流程定义")
	}

	if err := s.db.WithContext(ctx).Save(def).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新流程定义失败", err)
	}

	return nil
}

// PublishDefinition 发布流程定义
func (s *service) PublishDefinition(ctx context.Context, id uint) error {
	// 验证流程定义完整性
	def, err := s.GetDefinition(ctx, id)
	if err != nil {
		return err
	}

	if def.Status != "draft" {
		return errors.New(errors.ErrValidation, "只能发布草稿状态的流程定义")
	}

	// 验证至少有一个开始节点和一个结束节点
	nodes, err := s.GetNodes(ctx, id)
	if err != nil {
		return err
	}

	hasStart, hasEnd := false, false
	for _, node := range nodes {
		if node.NodeType == "start" {
			hasStart = true
		}
		if node.NodeType == "end" {
			hasEnd = true
		}
	}

	if !hasStart || !hasEnd {
		return errors.New(errors.ErrValidation, "流程定义必须包含开始节点和结束节点")
	}

	// 更新状态为已发布
	if err := s.db.WithContext(ctx).Model(&entity.WfDefinition{}).
		Where("ID = ?", id).
		Update("STATUS", "published").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "发布流程定义失败", err)
	}

	return nil
}

// ListDefinitions 查询流程定义列表
func (s *service) ListDefinitions(ctx context.Context, status string, page, pageSize int) ([]*entity.WfDefinition, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&entity.WfDefinition{}).Where("IS_ACTIVE = ?", "Y")
	if status != "" {
		query = query.Where("STATUS = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询流程定义总数失败", err)
	}

	var defs []*entity.WfDefinition
	offset := (page - 1) * pageSize
	if err := query.Order("ID DESC").Limit(pageSize).Offset(offset).Find(&defs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询流程定义列表失败", err)
	}

	return defs, total, nil
}

// CreateNode 创建流程节点
func (s *service) CreateNode(ctx context.Context, node *entity.WfNode) error {
	if err := s.db.WithContext(ctx).Create(node).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建流程节点失败", err)
	}

	return nil
}

// GetNodes 获取流程节点列表
func (s *service) GetNodes(ctx context.Context, definitionID uint) ([]*entity.WfNode, error) {
	var nodes []*entity.WfNode
	if err := s.db.WithContext(ctx).
		Where("WF_DEFINITION_ID = ? AND IS_ACTIVE = ?", definitionID, "Y").
		Find(&nodes).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询流程节点失败", err)
	}

	return nodes, nil
}

// UpdateNode 更新流程节点
func (s *service) UpdateNode(ctx context.Context, node *entity.WfNode) error {
	if err := s.db.WithContext(ctx).Save(node).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新流程节点失败", err)
	}

	return nil
}

// DeleteNode 删除流程节点
func (s *service) DeleteNode(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.WfNode{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除流程节点失败", err)
	}

	return nil
}

// CreateTransition 创建流程流转
func (s *service) CreateTransition(ctx context.Context, transition *entity.WfTransition) error {
	if err := s.db.WithContext(ctx).Create(transition).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建流程流转失败", err)
	}

	return nil
}

// GetTransitions 获取流程流转列表
func (s *service) GetTransitions(ctx context.Context, definitionID uint) ([]*entity.WfTransition, error) {
	var transitions []*entity.WfTransition
	if err := s.db.WithContext(ctx).
		Where("WF_DEFINITION_ID = ? AND IS_ACTIVE = ?", definitionID, "Y").
		Order("ORDERNO ASC").
		Find(&transitions).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "查询流程流转失败", err)
	}

	return transitions, nil
}

// DeleteTransition 删除流程流转
func (s *service) DeleteTransition(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.WfTransition{}).
		Where("ID = ?", id).
		Update("IS_ACTIVE", "N").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "删除流程流转失败", err)
	}

	return nil
}

// StartProcess 启动流程
func (s *service) StartProcess(ctx context.Context, req *StartProcessRequest) (*entity.WfInstance, error) {
	// 获取流程定义
	def, err := s.GetDefinition(ctx, req.DefinitionID)
	if err != nil {
		return nil, err
	}

	if def.Status != "published" {
		return nil, errors.New(errors.ErrValidation, "只能启动已发布的流程定义")
	}

	// 获取开始节点
	nodes, err := s.GetNodes(ctx, req.DefinitionID)
	if err != nil {
		return nil, err
	}

	var startNode *entity.WfNode
	for _, node := range nodes {
		if node.NodeType == "start" {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return nil, errors.New(errors.ErrValidation, "流程定义缺少开始节点")
	}

	// 序列化变量
	variablesJSON, _ := json.Marshal(req.Variables)

	// 创建流程实例
	instance := &entity.WfInstance{
		WfDefinitionID: req.DefinitionID,
		SysTableID:     req.SysTableID,
		BusinessID:     req.BusinessID,
		Status:         "running",
		CurrentNodeID:  startNode.ID,
		StartUserID:    req.StartUserID,
		StartTime:      time.Now(),
		Variables:      string(variablesJSON),
		Title:          req.Title,
	}
	instance.IsActive = "Y"

	if err := s.db.WithContext(ctx).Create(instance).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "创建流程实例失败", err)
	}

	// 移动到下一个节点
	if err := s.moveToNext(ctx, instance, startNode, req.Variables); err != nil {
		return nil, err
	}

	return instance, nil
}

// moveToNext 移动到下一个节点
func (s *service) moveToNext(ctx context.Context, instance *entity.WfInstance, currentNode *entity.WfNode, variables map[string]interface{}) error {
	// 获取当前节点的所有流转
	transitions, err := s.GetTransitions(ctx, instance.WfDefinitionID)
	if err != nil {
		return err
	}

	var nextTransitions []*entity.WfTransition
	for _, t := range transitions {
		if t.FromNodeID == currentNode.ID {
			nextTransitions = append(nextTransitions, t)
		}
	}

	if len(nextTransitions) == 0 {
		// 没有后续流转,检查是否是结束节点
		if currentNode.NodeType == "end" {
			// 流程结束
			instance.Status = "completed"
			instance.EndTime = time.Now()
			return s.db.WithContext(ctx).Save(instance).Error
		}
		return errors.New(errors.ErrValidation, "流程定义错误：节点没有后续流转")
	}

	// 找到符合条件的第一个流转
	var nextTransition *entity.WfTransition
	for _, t := range nextTransitions {
		if s.evaluateCondition(t.Condition, variables) {
			nextTransition = t
			break
		}
	}

	if nextTransition == nil {
		return errors.New(errors.ErrValidation, "没有符合条件的流转")
	}

	// 获取下一个节点
	var nextNode entity.WfNode
	if err := s.db.WithContext(ctx).First(&nextNode, nextTransition.ToNodeID).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询下一个节点失败", err)
	}

	// 更新实例当前节点
	instance.CurrentNodeID = nextNode.ID
	if err := s.db.WithContext(ctx).Save(instance).Error; err != nil {
		return err
	}

	// 根据节点类型执行操作
	switch nextNode.NodeType {
	case "user":
		// 创建用户任务
		return s.createUserTask(ctx, instance, &nextNode)
	case "auto":
		// 执行自动任务
		return s.executeAutoTask(ctx, instance, &nextNode, variables)
	case "end":
		// 流程结束
		instance.Status = "completed"
		instance.EndTime = time.Now()
		return s.db.WithContext(ctx).Save(instance).Error
	default:
		// 继续流转
		return s.moveToNext(ctx, instance, &nextNode, variables)
	}
}

// evaluateCondition 评估流转条件
func (s *service) evaluateCondition(condition string, variables map[string]interface{}) bool {
	// 如果没有条件,默认为true
	if condition == "" {
		return true
	}

	// TODO: 实现复杂的条件表达式评估
	// 简单实现：支持 变量名==值 的格式
	// 生产环境应使用表达式引擎如 govaluate
	return true
}

// createUserTask 创建用户任务
func (s *service) createUserTask(ctx context.Context, instance *entity.WfInstance, node *entity.WfNode) error {
	// 获取任务执行人
	assigneeID, err := s.getAssignee(ctx, node, instance)
	if err != nil {
		return err
	}

	task := &entity.WfTask{
		WfInstanceID: instance.ID,
		WfNodeID:     node.ID,
		AssigneeID:   assigneeID,
		Status:       "pending",
		Variables:    instance.Variables,
	}
	task.IsActive = "Y"

	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建任务失败", err)
	}

	return nil
}

// getAssignee 获取任务执行人
func (s *service) getAssignee(ctx context.Context, node *entity.WfNode, instance *entity.WfInstance) (uint, error) {
	switch node.AssignType {
	case "user":
		// 直接指定用户ID
		var assigneeID uint
		fmt.Sscanf(node.AssignValue, "%d", &assigneeID)
		return assigneeID, nil
	case "starter":
		// 流程发起人
		return instance.StartUserID, nil
	default:
		// TODO: 支持更多分配类型: role, expression等
		return 0, errors.New(errors.ErrValidation, "不支持的任务分配类型")
	}
}

// executeAutoTask 执行自动任务
func (s *service) executeAutoTask(ctx context.Context, instance *entity.WfInstance, node *entity.WfNode, variables map[string]interface{}) error {
	if node.ActionID == 0 {
		return errors.New(errors.ErrValidation, "自动任务必须配置动作")
	}

	// 执行关联的动作
	_, err := s.actionService.ExecuteAction(ctx, node.ActionID, variables, instance.StartUserID)
	if err != nil {
		return err
	}

	// 继续流转到下一个节点
	return s.moveToNext(ctx, instance, node, variables)
}

// GetInstance 获取流程实例
func (s *service) GetInstance(ctx context.Context, id uint) (*entity.WfInstance, error) {
	var instance entity.WfInstance
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "流程实例不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询流程实例失败", err)
	}

	return &instance, nil
}

// ListInstances 查询流程实例列表
func (s *service) ListInstances(ctx context.Context, req *ListInstancesRequest) ([]*entity.WfInstance, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&entity.WfInstance{}).Where("IS_ACTIVE = ?", "Y")
	if req.DefinitionID > 0 {
		query = query.Where("WF_DEFINITION_ID = ?", req.DefinitionID)
	}
	if req.Status != "" {
		query = query.Where("STATUS = ?", req.Status)
	}
	if req.StartUserID > 0 {
		query = query.Where("START_USER_ID = ?", req.StartUserID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询流程实例总数失败", err)
	}

	var instances []*entity.WfInstance
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("ID DESC").Limit(req.PageSize).Offset(offset).Find(&instances).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询流程实例列表失败", err)
	}

	return instances, total, nil
}

// TerminateInstance 终止流程实例
func (s *service) TerminateInstance(ctx context.Context, id uint, userID uint) error {
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	if instance.Status != "running" && instance.Status != "suspended" {
		return errors.New(errors.ErrValidation, "只能终止运行中或挂起的流程")
	}

	instance.Status = "terminated"
	instance.EndTime = time.Now()

	if err := s.db.WithContext(ctx).Save(instance).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "终止流程失败", err)
	}

	return nil
}

// SuspendInstance 挂起流程实例
func (s *service) SuspendInstance(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.WfInstance{}).
		Where("ID = ? AND STATUS = ?", id, "running").
		Update("STATUS", "suspended").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "挂起流程失败", err)
	}

	return nil
}

// ResumeInstance 恢复流程实例
func (s *service) ResumeInstance(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&entity.WfInstance{}).
		Where("ID = ? AND STATUS = ?", id, "suspended").
		Update("STATUS", "running").Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "恢复流程失败", err)
	}

	return nil
}

// GetTask 获取任务
func (s *service) GetTask(ctx context.Context, id uint) (*entity.WfTask, error) {
	var task entity.WfTask
	if err := s.db.WithContext(ctx).Where("ID = ? AND IS_ACTIVE = ?", id, "Y").First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrResourceNotFound, "任务不存在")
		}
		return nil, errors.Wrap(errors.ErrDatabase, "查询任务失败", err)
	}

	return &task, nil
}

// ListMyTasks 查询我的任务列表
func (s *service) ListMyTasks(ctx context.Context, userID uint, status string, page, pageSize int) ([]*entity.WfTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&entity.WfTask{}).
		Where("ASSIGNEE_ID = ? AND IS_ACTIVE = ?", userID, "Y")

	if status != "" {
		query = query.Where("STATUS = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询任务总数失败", err)
	}

	var tasks []*entity.WfTask
	offset := (page - 1) * pageSize
	if err := query.Order("ID DESC").Limit(pageSize).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, errors.Wrap(errors.ErrDatabase, "查询任务列表失败", err)
	}

	return tasks, total, nil
}

// CompleteTask 完成任务
func (s *service) CompleteTask(ctx context.Context, req *CompleteTaskRequest) error {
	// 获取任务
	task, err := s.GetTask(ctx, req.TaskID)
	if err != nil {
		return err
	}

	// 验证任务执行人
	if task.AssigneeID != req.UserID {
		return errors.New(errors.ErrPermissionDenied, "只能完成分配给自己的任务")
	}

	if task.Status != "pending" {
		return errors.New(errors.ErrValidation, "任务已处理")
	}

	// 获取流程实例
	instance, err := s.GetInstance(ctx, task.WfInstanceID)
	if err != nil {
		return err
	}

	if instance.Status != "running" {
		return errors.New(errors.ErrValidation, "流程实例不在运行状态")
	}

	// 更新任务状态
	task.Status = "completed"
	task.Action = req.Action
	task.Comment = req.Comment
	task.CompleteTime = time.Now()

	if err := s.db.WithContext(ctx).Save(task).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "更新任务失败", err)
	}

	// 如果是拒绝,终止流程
	if req.Action == "reject" {
		instance.Status = "terminated"
		instance.EndTime = time.Now()
		return s.db.WithContext(ctx).Save(instance).Error
	}

	// 获取当前节点
	var currentNode entity.WfNode
	if err := s.db.WithContext(ctx).First(&currentNode, task.WfNodeID).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询当前节点失败", err)
	}

	// 合并变量
	var instanceVars map[string]interface{}
	if instance.Variables != "" {
		json.Unmarshal([]byte(instance.Variables), &instanceVars)
	}
	if instanceVars == nil {
		instanceVars = make(map[string]interface{})
	}
	for k, v := range req.Variables {
		instanceVars[k] = v
	}

	// 更新实例变量
	variablesJSON, _ := json.Marshal(instanceVars)
	instance.Variables = string(variablesJSON)
	s.db.WithContext(ctx).Save(instance)

	// 继续流转
	return s.moveToNext(ctx, instance, &currentNode, instanceVars)
}

// ClaimTask 签收任务
func (s *service) ClaimTask(ctx context.Context, taskID, userID uint) error {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.AssigneeID != userID {
		return errors.New(errors.ErrPermissionDenied, "只能签收分配给自己的任务")
	}

	if task.Status != "pending" {
		return errors.New(errors.ErrValidation, "任务已处理")
	}

	task.ClaimTime = time.Now()

	if err := s.db.WithContext(ctx).Save(task).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "签收任务失败", err)
	}

	return nil
}

// TransferTask 转交任务
func (s *service) TransferTask(ctx context.Context, taskID, fromUserID, toUserID uint, comment string) error {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.AssigneeID != fromUserID {
		return errors.New(errors.ErrPermissionDenied, "只能转交自己的任务")
	}

	if task.Status != "pending" {
		return errors.New(errors.ErrValidation, "任务已处理")
	}

	task.AssigneeID = toUserID
	task.Status = "transferred"
	task.Comment = comment

	if err := s.db.WithContext(ctx).Save(task).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "转交任务失败", err)
	}

	// 创建新任务
	newTask := &entity.WfTask{
		WfInstanceID: task.WfInstanceID,
		WfNodeID:     task.WfNodeID,
		AssigneeID:   toUserID,
		Status:       "pending",
		Variables:    task.Variables,
	}
	newTask.IsActive = "Y"

	if err := s.db.WithContext(ctx).Create(newTask).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "创建转交任务失败", err)
	}

	return nil
}
