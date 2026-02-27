package hooks

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/storage"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
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
func liveRoomAfterCreateHandler(manager *core.Manager) func(map[string]interface{}) (interface{}, error) {
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

		// 获取公司 ID（可选）
		companyID := GetUintOrZero(params, "SYS_COMPANY_ID")

		// 加载配置
		cfg, err := config.Load()
		if err != nil {
			logger.Error("加载配置失败", zap.Error(err))
			return nil, err
		}

		// 创建云盘服务实例
		storageManager, err := storage.NewStorageManager(&cfg.Storage)
		if err != nil {
			logger.Error("创建存储管理器失败", zap.Error(err))
			return nil, err
		}
		defaultStorage, err := storageManager.GetDefaultStorage()
		if err != nil {
			logger.Error("获取默认存储失败", zap.Error(err))
			return nil, err
		}
		cloudService := cloud.NewService(txDB, defaultStorage)

		// 创建直播间目录（根目录）
		rootFolderReq := &cloud.CreateFolderRequest{
			Name:        roomName,
			ParentID:    nil, // 根目录
			Description: "直播间 " + roomName + " 的存储目录",
		}
		rootFolder, err := cloudService.CreateFolder(context.Background(), rootFolderReq, 1) // 0 表示系统用户
		if err != nil {
			logger.Error("创建直播间根目录失败", zap.Error(err), zap.String("roomName", roomName))
			return nil, err
		}
		logger.Info("创建直播间根目录成功", zap.String("roomName", roomName), zap.Uint("folderID", rootFolder.ID))

		// 创建直播录制子目录
		recordingFolderReq := &cloud.CreateFolderRequest{
			Name:        "直播录制",
			ParentID:    &rootFolder.ID,
			Description: "直播间 " + roomName + " 的直播录制存储目录",
		}
		_, err = cloudService.CreateFolder(context.Background(), recordingFolderReq, 1)
		if err != nil {
			logger.Error("创建直播录制子目录失败", zap.Error(err), zap.String("roomName", roomName))
			return nil, err
		}
		logger.Info("创建直播录制子目录成功", zap.String("roomName", roomName))

		// 创建直播切片子目录
		clipFolderReq := &cloud.CreateFolderRequest{
			Name:        "直播切片",
			ParentID:    &rootFolder.ID,
			Description: "直播间 " + roomName + " 的直播切片存储目录",
		}
		_, err = cloudService.CreateFolder(context.Background(), clipFolderReq, 1)
		if err != nil {
			logger.Error("创建直播切片子目录失败", zap.Error(err), zap.String("roomName", roomName))
			return nil, err
		}
		logger.Info("创建直播切片子目录成功", zap.String("roomName", roomName))

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
