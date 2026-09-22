package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/utils"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

// LiveClipHandler 直播切片/录制处理器
type LiveClipHandler struct {
	clipService live.ClipService
}

// NewLiveClipHandler 创建直播切片/录制处理器
func NewLiveClipHandler(clipService live.ClipService) *LiveClipHandler {
	return &LiveClipHandler{
		clipService: clipService,
	}
}

// ListClipsRequest 切片列表查询请求
type ListClipsRequest struct {
	RoomID   string `form:"roomId"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// ListClips 查询直播切片列表
// @Summary 查询直播切片列表
// @Description 查询精彩片段/切片列表，支持按直播间、状态、名称搜索
// @Tags Live-Clip
// @Accept json
// @Produce json
// @Param roomId query string false "直播间ID"
// @Param status query string false "状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} utils.Response{data=utils.PageResult}
// @Router /api/v1/live/clips [get]
func (h *LiveClipHandler) ListClips(c *gin.Context) {
	var req ListClipsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	filter, err := h.buildClipFilter(&req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	ctx := withCompanyID(c)
	clips, total, err := h.clipService.ListClips(ctx, filter)
	if err != nil {
		utils.InternalError(c, "查询切片列表失败")
		return
	}

	utils.PageResponse(c, clips, filter.Page, filter.PageSize, int(total))
}

// ListRecordingsRequest 录制列表查询请求
type ListRecordingsRequest struct {
	RoomID   string `form:"roomId"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// ListRecordings 查询直播录制列表
// @Summary 查询直播录制列表
// @Description 查询直播录制文件列表，支持按直播间、状态、关键词搜索
// @Tags Live-Recording
// @Accept json
// @Produce json
// @Param roomId query string false "直播间ID"
// @Param status query string false "状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页大小" default(20)
// @Success 200 {object} utils.Response{data=utils.PageResult}
// @Router /api/v1/live/recordings [get]
func (h *LiveClipHandler) ListRecordings(c *gin.Context) {
	var req ListRecordingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	clipReq := &ListClipsRequest{
		RoomID:   req.RoomID,
		Status:   req.Status,
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	filter, err := h.buildClipFilter(clipReq)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	recordingFilter := &live.RecordingFilter{
		RoomID:   filter.RoomID,
		Status:   filter.Status,
		Keyword:  filter.Keyword,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	ctx := withCompanyID(c)
	recordings, total, err := h.clipService.ListRecordings(ctx, recordingFilter)
	if err != nil {
		utils.InternalError(c, "查询录制列表失败")
		return
	}

	utils.PageResponse(c, recordings, recordingFilter.Page, recordingFilter.PageSize, int(total))
}

func (h *LiveClipHandler) buildClipFilter(req *ListClipsRequest) (*live.ClipFilter, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filter := &live.ClipFilter{
		Status:   req.Status,
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.RoomID != "" {
		roomID, err := strconv.ParseInt(req.RoomID, 10, 64)
		if err != nil {
			return nil, errors.New("直播间ID格式错误")
		}
		filter.RoomID = &roomID
	}

	return filter, nil
}
