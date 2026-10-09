package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/cloud"
	"gorm.io/gorm"
)

const testCallbackKey = "callback-test-key"

// fakeCloudService 只实现回调用到的 CreateItem：与真实实现一致，未指定公司时按属主所在公司归属
type fakeCloudService struct {
	cloud.Service
	db *gorm.DB
}

func (f *fakeCloudService) CreateItem(ctx context.Context, item *entity.CloudItem) error {
	if item.OwnerID != 0 && item.SysCompanyID == 0 {
		var user entity.SysUser
		if err := f.db.First(&user, item.OwnerID).Error; err == nil {
			item.SysCompanyID = user.SysCompanyID
		}
	}
	return f.db.Create(item).Error
}

// setupCallbackCompanyTest 直播间 501 属于公司 7，创建人 zhang 当前在公司 3（如调岗）：回调数据应跟随直播间归属公司 7
func setupCallbackCompanyTest(t *testing.T) (*LiveCallbackHandler, *gorm.DB, *entity.SysUser) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.LiveRoom{}, &entity.SysUser{}, &entity.CloudItem{}, &entity.LiveCallbackEvent{}); err != nil {
		t.Fatal(err)
	}

	owner := entity.SysUser{BaseModel: entity.BaseModel{IsActive: "Y"}, SysCompanyID: 3, Username: "zhang"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	room := entity.LiveRoom{
		BaseModel:  entity.BaseModel{ID: 501, SysCompanyID: 7, CreateBy: "zhang", IsActive: "Y"},
		RoomName:   "公司七的直播间",
		StreamName: "room-seven",
	}
	if err := db.Create(&room).Error; err != nil {
		t.Fatal(err)
	}

	h := NewLiveCallbackHandler(db, &config.Config{}, testCallbackKey, &fakeCloudService{db: db})
	return h, db, &owner
}

func postCallback(t *testing.T, handle gin.HandlerFunc, body map[string]interface{}) {
	t.Helper()
	exp := time.Now().Add(time.Minute).Unix()
	body["t"] = exp
	body["sign"] = signCallback(testCallbackKey, exp)
	raw, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	handle(c)
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"code":0`)) {
		t.Fatalf("回调未被接受: %d %s", w.Code, w.Body.String())
	}
}

func TestCallback_PushStream_EventBelongsToRoomCompany(t *testing.T) {
	h, db, _ := setupCallbackCompanyTest(t)
	postCallback(t, h.HandlePushStream, map[string]interface{}{"stream_id": "501", "event_time": time.Now().Unix()})

	var ev entity.LiveCallbackEvent
	if err := db.Where("EVENT_TYPE = ?", "push_stream").First(&ev).Error; err != nil {
		t.Fatal(err)
	}
	if ev.SysCompanyID != 7 {
		t.Errorf("推流事件应归属直播间所在公司 7，实际 %d", ev.SysCompanyID)
	}
	if ev.CreateBy != "zhang" || ev.UpdateBy != "zhang" {
		t.Errorf("回调没有操作人，创建人/更新人应取直播间创建人 zhang，实际 %q/%q", ev.CreateBy, ev.UpdateBy)
	}
	var room entity.LiveRoom
	db.First(&room, 501)
	if room.Status != "live" || room.UpdateBy != "zhang" {
		t.Errorf("推流后直播间应为 live、更新人 zhang，实际 %q/%q", room.Status, room.UpdateBy)
	}
}

// 拉流转推回调按流名称（非数字）匹配直播间，实体钩子无法从流名称推出公司，须取直播间的公司
func TestCallback_PullStream_EventBelongsToRoomCompany(t *testing.T) {
	h, db, _ := setupCallbackCompanyTest(t)
	postCallback(t, h.HandlePullStream, map[string]interface{}{"task_id": "task-1", "status": "start", "stream_name": "room-seven", "event_time": time.Now().Unix()})

	var ev entity.LiveCallbackEvent
	if err := db.Where("EVENT_TYPE = ?", "pull_stream").First(&ev).Error; err != nil {
		t.Fatal(err)
	}
	if ev.SysCompanyID != 7 {
		t.Errorf("拉流转推事件应归属直播间所在公司 7，实际 %d", ev.SysCompanyID)
	}
}

// 录制文件回调：目录与文件都归属直播间所在公司（不是创建人当前所在公司），属主为创建人
func TestCallback_RecordingFile_CloudItemsBelongToRoomCompany(t *testing.T) {
	h, db, owner := setupCallbackCompanyTest(t)
	postCallback(t, h.HandleRecordingFile, map[string]interface{}{
		"stream_id": "501", "event_time": time.Now().Unix(), "file_format": "mp4",
		"video_url": "https://example.com/r.mp4", "file_size": 1024,
	})

	var items []entity.CloudItem
	db.Order("ID").Find(&items)
	if len(items) != 3 {
		t.Fatalf("应创建 直播间目录/直播录制目录/录制文件 共 3 条，实际 %d", len(items))
	}
	for _, it := range items {
		if it.CreateBy != "zhang" {
			t.Errorf("%s %q 创建人应取直播间创建人 zhang，实际 %q", it.ItemType, it.Name, it.CreateBy)
		}
		if it.SysCompanyID != 7 || it.OwnerID != owner.ID {
			t.Errorf("%s %q 应归属直播间所在公司 7、属主 %d，实际公司 %d、属主 %d", it.ItemType, it.Name, owner.ID, it.SysCompanyID, it.OwnerID)
		}
	}
}

// countCallbackData 回调产生的数据：回调事件数、云盘记录数
func countCallbackData(db *gorm.DB) (events, items int64) {
	db.Model(&entity.LiveCallbackEvent{}).Count(&events)
	db.Model(&entity.CloudItem{}).Count(&items)
	return
}

func recordingFileBody() map[string]interface{} {
	return map[string]interface{}{
		"stream_id": "501", "event_time": time.Now().Unix(), "file_format": "mp4",
		"video_url": "https://example.com/r.mp4", "file_size": 1024,
	}
}

func highlightBody() map[string]interface{} {
	now := time.Now().Unix()
	return map[string]interface{}{
		"stream_id": "501",
		"items":     []map[string]interface{}{{"begin_time": now - 60, "end_time": now, "video_store_url": "https://example.com/c.mp4", "title": "精彩片段"}},
	}
}

// 直播间未开启录制：录制文件、录制状态、录制异常回调都不处理（不保存事件、不写云盘），但仍返回成功
func TestCallback_RecordingDisabled_Ignored(t *testing.T) {
	h, db, _ := setupCallbackCompanyTest(t)
	db.Model(&entity.LiveRoom{}).Where("ID = ?", 501).Update("IS_RECORD", "N")

	postCallback(t, h.HandleRecordingFile, recordingFileBody())
	postCallback(t, h.HandleRecordingStatus, map[string]interface{}{"stream_id": "501", "event_time": time.Now().Unix(), "status": 1})
	postCallback(t, h.HandleRecordException, map[string]interface{}{"stream_id": "501", "event_time": time.Now().Unix(), "error_code": 1})

	if events, items := countCallbackData(db); events != 0 || items != 0 {
		t.Errorf("未开启录制时不应处理录制回调，实际产生事件 %d 条、云盘记录 %d 条", events, items)
	}
}

// 直播间未开启切片：高光切片回调不处理；录制不受影响
func TestCallback_ClipDisabled_Ignored(t *testing.T) {
	h, db, _ := setupCallbackCompanyTest(t)
	db.Model(&entity.LiveRoom{}).Where("ID = ?", 501).Update("IS_CLIP", "N")

	postCallback(t, h.HandleHighlight, highlightBody())
	if events, items := countCallbackData(db); events != 0 || items != 0 {
		t.Fatalf("未开启切片时不应处理切片回调，实际产生事件 %d 条、云盘记录 %d 条", events, items)
	}

	postCallback(t, h.HandleRecordingFile, recordingFileBody())
	if events, items := countCallbackData(db); events != 1 || items != 3 {
		t.Errorf("录制已开启应正常处理，期望事件 1 条、云盘记录 3 条，实际 %d / %d", events, items)
	}
}

// 已开启（默认）时高光切片正常处理
func TestCallback_ClipEnabled_Processed(t *testing.T) {
	h, db, _ := setupCallbackCompanyTest(t)
	postCallback(t, h.HandleHighlight, highlightBody())
	if events, items := countCallbackData(db); events != 1 || items != 3 {
		t.Errorf("开启切片应保存事件并写入云盘（直播间目录/直播切片目录/切片文件），实际事件 %d 条、云盘记录 %d 条", events, items)
	}
}

// 直播间配置了录制存储位置：录制文件保存到「录制存储位置/直播间名称/直播录制」
func TestCallback_RecordingFile_UsesRoomStorageFolder(t *testing.T) {
	h, db, owner := setupCallbackCompanyTest(t)
	// 存储位置须与直播间同公司（直播间属于公司 7）
	storage := entity.CloudItem{BaseModel: entity.BaseModel{SysCompanyID: 7, IsActive: "Y"}, ItemType: "folder", Name: "直播资料", Path: "/直播资料", OwnerID: owner.ID}
	if err := db.Create(&storage).Error; err != nil {
		t.Fatal(err)
	}
	db.Model(&entity.LiveRoom{}).Where("ID = ?", 501).Update("RECORD_FOLDER_ID", storage.ID)

	postCallback(t, h.HandleRecordingFile, recordingFileBody())

	var file entity.CloudItem
	if err := db.Where("ITEM_TYPE = ?", "file").First(&file).Error; err != nil {
		t.Fatal(err)
	}
	want := "/直播资料/公司七的直播间/直播录制/"
	if len(file.Path) < len(want) || file.Path[:len(want)] != want {
		t.Errorf("录制文件应保存在直播间存储位置下，期望前缀 %q，实际 %q", want, file.Path)
	}
}

// 录制、切片分别设置存储位置：各自保存到对应的文件夹下
func TestCallback_RecordAndClipUseOwnFolders(t *testing.T) {
	h, db, owner := setupCallbackCompanyTest(t)
	recordDir := entity.CloudItem{BaseModel: entity.BaseModel{SysCompanyID: 7, IsActive: "Y"}, ItemType: "folder", Name: "录像", Path: "/录像", OwnerID: owner.ID}
	clipDir := entity.CloudItem{BaseModel: entity.BaseModel{SysCompanyID: 7, IsActive: "Y"}, ItemType: "folder", Name: "精彩", Path: "/素材/精彩", OwnerID: owner.ID}
	db.Create(&recordDir)
	db.Create(&clipDir)
	db.Model(&entity.LiveRoom{}).Where("ID = ?", 501).Updates(map[string]interface{}{"RECORD_FOLDER_ID": recordDir.ID, "CLIP_FOLDER_ID": clipDir.ID})

	postCallback(t, h.HandleRecordingFile, recordingFileBody())
	postCallback(t, h.HandleHighlight, highlightBody())

	var files []entity.CloudItem
	db.Where("ITEM_TYPE = ?", "file").Order("ID").Find(&files)
	if len(files) != 2 {
		t.Fatalf("应生成录制与切片文件各 1 个，实际 %d 个", len(files))
	}
	wants := []string{"/录像/公司七的直播间/直播录制/", "/素材/精彩/公司七的直播间/直播切片/"}
	for i, want := range wants {
		if !strings.HasPrefix(files[i].Path, want) {
			t.Errorf("文件 %d 应保存在 %q 下，实际 %q", i, want, files[i].Path)
		}
	}
}
