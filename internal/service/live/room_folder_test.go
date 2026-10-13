package live

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"gorm.io/gorm"
)

func setupRoomFolderDB(t *testing.T) *gorm.DB {
	t.Helper()
	if _, err := logger.Init(&config.LogConfig{Level: "error"}); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.CloudItem{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newFolder(t *testing.T, db *gorm.DB, name, path string, owner, company uint) *entity.CloudItem {
	t.Helper()
	f := &entity.CloudItem{BaseModel: entity.BaseModel{SysCompanyID: company, IsActive: "Y"}, ItemType: "folder", Name: name, Path: path, OwnerID: owner}
	if err := db.Create(f).Error; err != nil {
		t.Fatal(err)
	}
	return f
}

func TestEnsureRoomFolder_NoStorage_UsesRoot(t *testing.T) {
	db := setupRoomFolderDB(t)
	f, err := EnsureRoomFolder(db, RoomFolderSpec{RoomName: "年会", CreateBy: "zhang", CompanyID: 7, OwnerID: 10}, RoomFolderRecording)
	if err != nil {
		t.Fatal(err)
	}
	if f.Path != "/年会/直播录制" || f.SysCompanyID != 7 || f.OwnerID != 10 {
		t.Errorf("未配置存储位置应放在根目录: path=%q company=%d owner=%d", f.Path, f.SysCompanyID, f.OwnerID)
	}
}

func TestEnsureRoomFolder_StorageFolder_UsedAsParent(t *testing.T) {
	db := setupRoomFolderDB(t)
	storage := newFolder(t, db, "直播资料", "/项目/直播资料", 10, 7)
	f, err := EnsureRoomFolder(db, RoomFolderSpec{RoomName: "年会", CompanyID: 7, OwnerID: 10, StorageFolderID: &storage.ID}, RoomFolderClip)
	if err != nil {
		t.Fatal(err)
	}
	if f.Path != "/项目/直播资料/年会/直播切片" {
		t.Errorf("应保存在直播间配置的存储位置下，实际 %q", f.Path)
	}
	var room entity.CloudItem
	db.First(&room, *f.ParentID)
	if room.ParentID == nil || *room.ParentID != storage.ID {
		t.Errorf("直播间目录的上级应为存储位置 %d", storage.ID)
	}
}

// 存储位置属于其他用户时不能写入，改为根目录
func TestEnsureRoomFolder_OthersFolder_FallsBackToRoot(t *testing.T) {
	db := setupRoomFolderDB(t)
	others := newFolder(t, db, "别人的目录", "/别人的目录", 99, 7)
	f, err := EnsureRoomFolder(db, RoomFolderSpec{RoomName: "年会", CompanyID: 7, OwnerID: 10, StorageFolderID: &others.ID}, RoomFolderRecording)
	if err != nil {
		t.Fatal(err)
	}
	if f.Path != "/年会/直播录制" {
		t.Errorf("存储位置不属于直播间创建人时应改为根目录，实际 %q", f.Path)
	}
}

// 已存在的目录直接复用，不重复创建
func TestEnsureRoomFolder_ReusesExisting(t *testing.T) {
	db := setupRoomFolderDB(t)
	spec := RoomFolderSpec{RoomName: "年会", CompanyID: 7, OwnerID: 10}
	a, _ := EnsureRoomFolder(db, spec, RoomFolderRecording)
	b, _ := EnsureRoomFolder(db, spec, RoomFolderRecording)
	if _, err := EnsureRoomFolder(db, spec, RoomFolderClip); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&entity.CloudItem{}).Count(&n)
	if a.ID != b.ID || n != 3 {
		t.Errorf("应复用已有目录：两次 ID %d/%d，共 %d 条（期望 3：直播间目录+两个子目录）", a.ID, b.ID, n)
	}
}

// 不同公司有同名用户时，按直播间所属公司找创建人
func TestRoomCreatorID_SameUsernameInOtherCompany(t *testing.T) {
	db := setupRoomFolderDB(t)
	// 实体声明了 USERNAME 唯一索引，但实际库中用户名只在公司内唯一，这里按实际库建表
	if err := db.Exec("CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, USERNAME TEXT, SYS_COMPANY_ID INTEGER, IS_ACTIVE TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO sys_user VALUES (31, 'sajida', 51, 'Y'), (38, 'sajida', 58, 'Y')").Error; err != nil {
		t.Fatal(err)
	}
	if id, ok := RoomCreatorID(db, "sajida", 58); !ok || id != 38 {
		t.Errorf("公司 58 的 sajida 应为 38，实际 %d（%v）", id, ok)
	}
	if _, ok := RoomCreatorID(db, "sajida", 99); ok {
		t.Error("所属公司没有该用户、全库又有多个同名用户时不应猜测")
	}
	if err := db.Exec("INSERT INTO sys_user VALUES (40, 'zhang', 3, 'Y')").Error; err != nil {
		t.Fatal(err)
	}
	if id, ok := RoomCreatorID(db, "zhang", 7); !ok || id != 40 {
		t.Errorf("创建人已调到其他公司、全库唯一时应找到 40，实际 %d（%v）", id, ok)
	}
	if _, ok := RoomCreatorID(db, "", 58); ok {
		t.Error("用户名为空时不应找到")
	}
}

// 所有者在其他公司已有同名文件夹时不能复用（避免录制文件写进其他公司的目录）：报错或新建，都不能返回其他公司的目录
func TestEnsureRoomFolder_DoesNotReuseOtherCompanyFolder(t *testing.T) {
	db := setupRoomFolderDB(t)
	other := newFolder(t, db, "3", "/3", 31, 51)
	f, err := EnsureRoomFolder(db, RoomFolderSpec{RoomName: "3", CompanyID: 58, OwnerID: 31}, RoomFolderRecording)
	if err != nil {
		return
	}
	var room entity.CloudItem
	db.First(&room, *f.ParentID)
	if room.ID == other.ID || room.SysCompanyID != 58 {
		t.Errorf("复用了公司 51 的目录：直播间目录 ID=%d 公司=%d", room.ID, room.SysCompanyID)
	}
}
