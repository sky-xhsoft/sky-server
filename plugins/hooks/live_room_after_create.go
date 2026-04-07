package hooks

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
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

		// 获取创建者用户名
		createBy, _ := GetStringFromParams(params, "CREATE_BY")

		// 获取公司 ID（可选）
		companyID := GetUintOrZero(params, "SYS_COMPANY_ID")

		// 查找创建者用户ID
		var creatorUser entity.SysUser
		var userID uint = 1 // 默认使用系统用户
		if createBy != "" {
			if err := txDB.Where("USERNAME = ? AND IS_ACTIVE = ?", createBy, "Y").First(&creatorUser).Error; err == nil {
				userID = creatorUser.ID
			}
		}

		// 直接使用数据库操作创建云盘文件夹
		var roomFolder entity.CloudItem

		// 检查直播间文件夹是否存在
		if err := txDB.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID IS NULL AND IS_ACTIVE = ?",
			"folder", roomName, userID, "Y").First(&roomFolder).Error; err != nil {
			// 直播间文件夹不存在，创建
			roomFolder = entity.CloudItem{
				BaseModel: entity.BaseModel{
					SysCompanyID: companyID,
					CreateBy:     createBy,
					UpdateBy:     createBy,
					IsActive:     "Y",
				},
				ItemType: "folder",
				Name:     roomName,
				ParentID: nil,
				Path:     "/" + roomName,
				OwnerID:  userID,
			}
			if err := txDB.Create(&roomFolder).Error; err != nil {
				logger.Error("创建直播间文件夹失败", zap.Error(err), zap.String("roomName", roomName))
				// 不影响主流程，继续执行
			} else {
				logger.Info("创建直播间根目录成功", zap.String("roomName", roomName), zap.Uint("folderID", roomFolder.ID))
			}
		}

		// 创建直播录制子目录（如果直播间文件夹创建成功或已存在）
		if roomFolder.ID > 0 {
			recordingFolderName := "直播录制"
			var recordingFolder entity.CloudItem
			if err := txDB.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = ?",
				"folder", recordingFolderName, userID, roomFolder.ID, "Y").First(&recordingFolder).Error; err != nil {
				recordingFolder = entity.CloudItem{
					BaseModel: entity.BaseModel{
						SysCompanyID: companyID,
						CreateBy:     createBy,
						UpdateBy:     createBy,
						IsActive:     "Y",
					},
					ItemType: "folder",
					Name:     recordingFolderName,
					ParentID: &roomFolder.ID,
					Path:     roomFolder.Path + "/" + recordingFolderName,
					OwnerID:  userID,
				}
				if err := txDB.Create(&recordingFolder).Error; err != nil {
					logger.Error("创建直播录制子目录失败", zap.Error(err), zap.String("roomName", roomName))
				} else {
					logger.Info("创建直播录制子目录成功", zap.String("roomName", roomName))
				}
			}

			// 创建直播切片子目录
			clipFolderName := "直播切片"
			var clipFolder entity.CloudItem
			if err := txDB.Where("ITEM_TYPE = ? AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = ?",
				"folder", clipFolderName, userID, roomFolder.ID, "Y").First(&clipFolder).Error; err != nil {
				clipFolder = entity.CloudItem{
					BaseModel: entity.BaseModel{
						SysCompanyID: companyID,
						CreateBy:     createBy,
						UpdateBy:     createBy,
						IsActive:     "Y",
					},
					ItemType: "folder",
					Name:     clipFolderName,
					ParentID: &roomFolder.ID,
					Path:     roomFolder.Path + "/" + clipFolderName,
					OwnerID:  userID,
				}
				if err := txDB.Create(&clipFolder).Error; err != nil {
					logger.Error("创建直播切片子目录失败", zap.Error(err), zap.String("roomName", roomName))
				} else {
					logger.Info("创建直播切片子目录成功", zap.String("roomName", roomName))
				}
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
