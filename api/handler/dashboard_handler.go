package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/api/middleware"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/dashboard"
)

// DashboardHandler Dashboard 聚合数据处理器
type DashboardHandler struct {
	dashboardService dashboard.Service
}

// NewDashboardHandler 创建 Dashboard 处理器
func NewDashboardHandler(dashboardService dashboard.Service) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// getCompanyID 从上下文中获取当前公司 ID
func getCompanyID(c *gin.Context) uint {
	if id := middleware.GetCompanyID(c); id != nil {
		return *id
	}
	return 0
}

// GetStats 获取 Dashboard 统计卡
// @Summary 获取 Dashboard 统计卡
// @Description 返回直播间总数、今日直播场次、云盘已用空间、权限组数
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} dashboard.Stats
// @Router /api/v1/dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.BadRequest(c, "无法识别公司")
		return
	}

	stats, err := h.dashboardService.GetStats(c, companyID)
	if err != nil {
		utils.InternalError(c, "获取统计信息失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetTrend 获取近 7 日趋势
// @Summary 获取近 7 日趋势
// @Description 返回近 7 天每日直播场次与新增录制数
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param days query int false "天数，默认 7"
// @Success 200 {object} dashboard.Trend
// @Router /api/v1/dashboard/trend [get]
func (h *DashboardHandler) GetTrend(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.BadRequest(c, "无法识别公司")
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 {
		days = 7
	}

	trend, err := h.dashboardService.GetTrend(c, companyID, days)
	if err != nil {
		utils.InternalError(c, "获取趋势数据失败: "+err.Error())
		return
	}

	utils.Success(c, trend)
}

// GetRecentRooms 获取最近直播
// @Summary 获取最近直播
// @Description 返回最近开播/更新的直播间列表
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param limit query int false "数量，默认 5"
// @Success 200 {array} dashboard.RecentRoom
// @Router /api/v1/dashboard/recent-rooms [get]
func (h *DashboardHandler) GetRecentRooms(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.BadRequest(c, "无法识别公司")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit <= 0 {
		limit = 5
	}

	rooms, err := h.dashboardService.GetRecentRooms(c, companyID, limit)
	if err != nil {
		utils.InternalError(c, "获取最近直播失败: "+err.Error())
		return
	}

	utils.Success(c, rooms)
}

// GetActivities 获取最新动态
// @Summary 获取最新动态
// @Description 返回最近系统操作动态
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param limit query int false "数量，默认 10"
// @Success 200 {array} dashboard.Activity
// @Router /api/v1/dashboard/activities [get]
func (h *DashboardHandler) GetActivities(c *gin.Context) {
	companyID := getCompanyID(c)
	if companyID == 0 {
		utils.BadRequest(c, "无法识别公司")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}

	activities, err := h.dashboardService.GetActivities(c, companyID, limit)
	if err != nil {
		utils.InternalError(c, "获取最新动态失败: "+err.Error())
		return
	}

	utils.Success(c, activities)
}
