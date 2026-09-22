package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"gorm.io/gorm"
)

// ScheduleHandler 自动任务处理器
type ScheduleHandler struct {
	db       *gorm.DB
	runner   func(scheduleID uint) (string, error)  // 手动执行入口(注入)
	cronNext func(expr string) (time.Time, error)   // cron 解析
}

// NewScheduleHandler 创建自动任务处理器
// runner 由 wire 注入(Scheduler.RunNow 的封装)
func NewScheduleHandler(db *gorm.DB, runner func(scheduleID uint) (string, error), cronNext func(expr string) (time.Time, error)) *ScheduleHandler {
	return &ScheduleHandler{db: db, runner: runner, cronNext: cronNext}
}

// Run 手动执行任务 POST /schedule/:id/run
func (h *ScheduleHandler) Run(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}
	if h.runner == nil {
		utils.InternalError(c, "调度器未初始化")
		return
	}
	msg, rerr := h.runner(uint(id))
	if rerr != nil {
		utils.Success(c, gin.H{"status": "FAIL", "message": msg, "error": rerr.Error()})
		return
	}
	utils.Success(c, gin.H{"status": "SUCCESS", "message": msg})
}

// CronVerify 校验 cron 表达式并返回下次运行时间 POST /schedule/cron-verify
// body: {cron}
func (h *ScheduleHandler) CronVerify(c *gin.Context) {
	var req struct {
		Cron string `json:"cron"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if h.cronNext == nil {
		utils.InternalError(c, "cron 解析未初始化")
		return
	}
	next, err := h.cronNext(req.Cron)
	if err != nil {
		utils.Success(c, gin.H{"valid": false, "error": err.Error()})
		return
	}
	utils.Success(c, gin.H{"valid": true, "next": next.Format("2006-01-02 15:04:05")})
}

// Logs 查询任务执行日志 GET /schedule/:id/logs?page=&pageSize=
func (h *ScheduleHandler) Logs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	h.db.Model(&entity.SysScheduleLog{}).Where("SCHEDULE_ID = ?", id).Count(&total)
	var rows []entity.SysScheduleLog
	h.db.Where("SCHEDULE_ID = ?", id).Order("ID DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows)
	utils.Success(c, gin.H{"data": rows, "total": total})
}