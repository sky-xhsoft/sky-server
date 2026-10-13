package hooks

import (
	"context"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	liveService "github.com/sky-xhsoft/sky-server/internal/service/live"
	"github.com/sky-xhsoft/sky-server/plugins/core"
	"go.uber.org/zap"
)

// LiveRoomAfterCreateHook live_room 创建后钩子
type LiveRoomAfterCreateHook struct {
	*BaseHook
}

// 在 init() 中自动注册
func init() {
	hook := &LiveRoomAfterCreateHook{
		BaseHook: NewBaseHook("LIVE_ROOM_AFTER_CREATE", liveRoomAfterCreateHandler),
	}
	Register(hook)
}

// liveRoomAfterCreateHandler 处理 live_room 创建后的逻辑
func liveRoomAfterCreateHandler(manager *core.Manager, cfg *config.Config) func(map[string]interface{}) (interface{}, error) {
	return func(params map[string]interface{}) (interface{}, error) {
		logger.Info("执行 LIVE_ROOM_AFTER_CREATE 钩子")

		// 获取数据库连接（事务连接）
		txDB, err := GetDBFromParams(params)
		if err != nil {
			return nil, err
		}

		// 获取记录 ID
		recordID, err := GetUintFromParams(params, "ID")
		if err != nil {
			return nil, err
		}

		// 获取直播间名称
		roomName, err := GetStringFromParams(params, "ROOM_NAME")
		if err != nil {
			logger.Error("获取直播间名称失败", zap.Error(err))
			return nil, err
		}

		// 获取创建者用户名
		createBy, _ := GetStringFromParams(params, "CREATE_BY")

		// 获取公司 ID（可选）
		companyID := GetUintOrZero(params, "SYS_COMPANY_ID")

		// 查找创建者用户 ID（按直播间所属公司查找：不同公司可以有同名用户）。
		// 找不到时不建目录，录制 / 切片回调到达时会再补建，避免把目录记到其他用户名下
		userID, ok := liveService.RoomCreatorID(txDB, createBy, companyID)
		if !ok {
			logger.Warn("未找到直播间创建人，暂不创建云盘目录", zap.String("createBy", createBy), zap.Uint("companyId", companyID))
		}

		// 开启录制 / 切片时分别创建云盘目录：
		// 录制存储位置/直播间名称/直播录制，切片存储位置/直播间名称/直播切片
		targets := []struct {
			enabledKey, folderKey, sub string
		}{
			{"IS_RECORD", "RECORD_FOLDER_ID", liveService.RoomFolderRecording},
			{"IS_CLIP", "CLIP_FOLDER_ID", liveService.RoomFolderClip},
		}
		for _, target := range targets {
			if !ok {
				break
			}
			if strings.EqualFold(GetStringOrEmpty(params, target.enabledKey), "N") {
				continue
			}
			spec := liveService.RoomFolderSpec{
				RoomName:  roomName,
				CreateBy:  createBy,
				CompanyID: companyID,
				OwnerID:   userID,
			}
			if folderID := GetUintOrZero(params, target.folderKey); folderID > 0 {
				spec.StorageFolderID = &folderID
			}
			// 不影响直播间创建：失败只记录日志，回调到达时会再次补建
			if _, err := liveService.EnsureRoomFolder(txDB, spec, target.sub); err != nil {
				logger.Error("创建直播间云盘目录失败", zap.Error(err), zap.String("roomName", roomName), zap.String("sub", target.sub))
			}
		}

		// 构造插件数据
		pluginData := core.PluginData{
			TableName: "live_room",
			Action:    "create",
			Timing:    "after",
			RecordID:  recordID,
			CompanyID: companyID,
			Data:      params,
		}

		// 执行插件（使用事务连接）
		ctx := context.Background()
		if err := manager.ExecuteWithDB(ctx, txDB, pluginData); err != nil {
			logger.Error("执行插件失败", zap.Error(err))
			return nil, err
		}

		return SuccessResult("live_room 创建后钩子执行成功"), nil
	}
}
