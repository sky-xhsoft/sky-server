package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// RecordingFileEvent 录制文件事件结构
type RecordingFileEvent struct {
	EventType    int    `json:"event_type"`
	StreamID     string `json:"stream_id"`
	ChannelID    string `json:"channel_id"`
	EventTime    int64  `json:"event_time"`
	VideoURL     string `json:"video_url"`
	FileSize     int64  `json:"file_size"`
	Duration     int64  `json:"duration"`
	FileFormat   string `json:"file_format"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	StreamParam  string `json:"stream_param"`
	VideoID      string `json:"video_id"`
	RecordFileID string `json:"record_file_id"`
	App          string `json:"app"`
	AppName      string `json:"appname"`
}

// LiveCallbackEvent 回调事件
type LiveCallbackEvent struct {
	ID        int64
	EventType string
	RoomName  string
	StreamID  string
	EventTime int64
	EventData string
}

// 数据库配置
const (
	DBHost     = "127.0.0.1"
	DBPort     = 3306
	DBUser     = "root"
	DBPassword = "***REMOVED***"
	DBName     = "skyserver"
)

// DefaultRoomParentID 默认直播间父文件夹ID（与配置一致）
const DefaultRoomParentID = 391

func main() {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DBUser, DBPassword, DBHost, DBPort, DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("数据库连接成功！")
	fmt.Println("开始修复录制文件的 cloud_item...")
	fmt.Println("=" + strings.Repeat("=", 80))

	// 1. 查询所有 recording_file 类型的事件
	events, err := queryRecordingFileEvents(db)
	if err != nil {
		log.Fatalf("查询事件失败: %v", err)
	}

	fmt.Printf("找到 %d 条录制文件事件\n\n", len(events))

	// 2. 处理每个事件
	successCount := 0
	skipCount := 0
	failCount := 0

	for _, event := range events {
		fmt.Printf("处理事件 ID=%d, StreamID=%s, RoomName=%s\n",
			event.ID, event.StreamID, event.RoomName)

		// 解析事件数据
		var fileEvent RecordingFileEvent
		if err := json.Unmarshal([]byte(event.EventData), &fileEvent); err != nil {
			fmt.Printf("  ❌ 解析事件数据失败: %v\n", err)
			failCount++
			continue
		}

		// 检查文件是否已存在
		exists, err := checkFileExists(db, fileEvent.VideoURL)
		if err != nil {
			fmt.Printf("  ❌ 检查文件是否存在失败: %v\n", err)
			failCount++
			continue
		}
		if exists {
			fmt.Printf("  ⏭️  文件已存在，跳过\n")
			skipCount++
			continue
		}

		// 获取直播间信息
		room, err := getLiveRoom(db, event.StreamID)
		if err != nil {
			fmt.Printf("  ❌ 获取直播间信息失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  ✅ 找到直播间: ID=%d, Name=%s, CreateBy=%s\n",
			room.ID, room.RoomName, room.CreateBy)

		// 获取用户ID
		userID, err := getUserIDByUsername(db, room.CreateBy)
		if err != nil {
			fmt.Printf("  ❌ 获取用户ID失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  ✅ 用户ID: %d\n", userID)

		// 获取或创建直播间文件夹
		roomFolderID, err := getOrCreateRoomFolder(db, room.RoomName, userID)
		if err != nil {
			fmt.Printf("  ❌ 获取/创建直播间文件夹失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  ✅ 直播间文件夹ID: %d\n", roomFolderID)

		// 获取或创建直播录制文件夹
		recordingFolderID, err := getOrCreateRecordingFolder(db, roomFolderID, room.RoomName, userID)
		if err != nil {
			fmt.Printf("  ❌ 获取/创建直播录制文件夹失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  ✅ 直播录制文件夹ID: %d\n", recordingFolderID)

		// 确定事件时间
		eventTime := fileEvent.EventTime
		if eventTime == 0 {
			if fileEvent.EndTime > 0 {
				eventTime = fileEvent.EndTime
			} else {
				eventTime = time.Now().Unix()
			}
		}

		// 创建文件记录
		fileID, err := createRecordingFile(db, recordingFolderID, room.RoomName, userID, eventTime, &fileEvent)
		if err != nil {
			fmt.Printf("  ❌ 创建录制文件失败: %v\n", err)
			failCount++
			continue
		}
		fmt.Printf("  ✅ 文件创建成功! ID=%d\n", fileID)

		successCount++
		fmt.Println()
	}

	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Printf("处理完成！成功: %d, 跳过: %d, 失败: %d\n",
		successCount, skipCount, failCount)
}

// 查询所有录制文件事件
func queryRecordingFileEvents(db *sql.DB) ([]*LiveCallbackEvent, error) {
	rows, err := db.Query(`
		SELECT ID, EVENT_TYPE, ROOM_NAME, STREAM_ID, EVENT_TIME, EVENT_DATA
		FROM live_callback_event
		WHERE EVENT_TYPE = 'recording_file'
		ORDER BY ID ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*LiveCallbackEvent
	for rows.Next() {
		var e LiveCallbackEvent
		err := rows.Scan(&e.ID, &e.EventType, &e.RoomName, &e.StreamID, &e.EventTime, &e.EventData)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

// 检查文件是否已存在
func checkFileExists(db *sql.DB, accessURL string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM cloud_item
		WHERE ITEM_TYPE = 'file' AND ACCESS_URL = ? AND IS_ACTIVE = 'Y'
	`, accessURL).Scan(&count)
	return count > 0, err
}

// 获取直播间信息
type LiveRoom struct {
	ID       int64
	RoomName string
	CreateBy string
}

func getLiveRoom(db *sql.DB, streamID string) (*LiveRoom, error) {
	// 先尝试转为数字查询
	id, err := strconv.ParseUint(streamID, 10, 32)
	if err == nil {
		var room LiveRoom
		err := db.QueryRow(`
			SELECT ID, ROOM_NAME, CREATE_BY
			FROM live_room
			WHERE ID = ? AND IS_ACTIVE = 'Y'
		`, uint(id)).Scan(&room.ID, &room.RoomName, &room.CreateBy)
		if err == nil {
			return &room, nil
		}
	}

	// 数字查询失败，尝试字符串查询（兼容旧数据）
	var room LiveRoom
	err = db.QueryRow(`
		SELECT ID, ROOM_NAME, CREATE_BY
		FROM live_room
		WHERE ID = ? AND IS_ACTIVE = 'Y'
	`, streamID).Scan(&room.ID, &room.RoomName, &room.CreateBy)
	if err != nil {
		return nil, fmt.Errorf("未找到直播间: %s", streamID)
	}
	return &room, nil
}

// 获取用户ID
func getUserIDByUsername(db *sql.DB, username string) (int64, error) {
	var userID int64
	err := db.QueryRow(`
		SELECT ID FROM sys_user WHERE USERNAME = ? AND IS_ACTIVE = 'Y'
	`, username).Scan(&userID)
	return userID, err
}

// 获取或创建直播间文件夹
func getOrCreateRoomFolder(db *sql.DB, roomName string, userID int64) (int64, error) {
	// 先查询是否已存在
	var folderID int64
	err := db.QueryRow(`
		SELECT ID FROM cloud_item
		WHERE ITEM_TYPE = 'folder' AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = 'Y'
	`, roomName, userID, DefaultRoomParentID).Scan(&folderID)
	if err == nil {
		return folderID, nil
	}

	// 不存在则创建
	path := "/" + roomName
	result, err := db.Exec(`
		INSERT INTO cloud_item (
			ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
			FILE_COUNT, TOTAL_SIZE, CREATE_BY, UPDATE_BY, IS_ACTIVE
		) VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?, 'Y')
	`, "folder", roomName, DefaultRoomParentID, path, userID, "admin", "admin")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// 获取或创建直播录制文件夹
func getOrCreateRecordingFolder(db *sql.DB, parentID int64, roomName string, userID int64) (int64, error) {
	folderName := "直播录制"

	// 先查询是否已存在
	var folderID int64
	err := db.QueryRow(`
		SELECT ID FROM cloud_item
		WHERE ITEM_TYPE = 'folder' AND NAME = ? AND OWNER_ID = ? AND PARENT_ID = ? AND IS_ACTIVE = 'Y'
	`, folderName, userID, parentID).Scan(&folderID)
	if err == nil {
		return folderID, nil
	}

	// 不存在则创建
	path := "/" + roomName + "/" + folderName
	result, err := db.Exec(`
		INSERT INTO cloud_item (
			ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
			FILE_COUNT, TOTAL_SIZE, CREATE_BY, UPDATE_BY, IS_ACTIVE
		) VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?, 'Y')
	`, "folder", folderName, parentID, path, userID, "admin", "admin")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// 创建录制文件
func createRecordingFile(db *sql.DB, parentID int64, roomName string, userID int64, eventTime int64, event *RecordingFileEvent) (int64, error) {
	fileName := fmt.Sprintf("录制_%d.%s", eventTime, event.FileFormat)
	filePath := "/" + roomName + "/直播录制/" + fileName
	fileType := "video/" + event.FileFormat
	fileExt := "." + event.FileFormat

	result, err := db.Exec(`
		INSERT INTO cloud_item (
			ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
			STORAGE_TYPE, STORAGE_PATH, FILE_SIZE, FILE_TYPE, FILE_EXT,
			ACCESS_URL, DOWNLOAD_COUNT, CREATE_BY, UPDATE_BY, IS_ACTIVE
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Y')
	`, "file", fileName, parentID, filePath, userID,
		"oss", event.VideoURL, event.FileSize, fileType, fileExt,
		event.VideoURL, 0, "admin", "admin")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
