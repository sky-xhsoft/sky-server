package cloud

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func TestFileCategoryOf(t *testing.T) {
	cases := map[[2]string]string{
		{".MP4", ""}:              "video",
		{"mov", ""}:               "video",
		{".jpg", ""}:              "image",
		{".pdf", ""}:              "document",
		{".docx", ""}:             "document",
		{".csv", ""}:              "document",
		{".zip", ""}:              "other",
		{".mp3", ""}:              "other",
		{"", "video/quicktime"}:   "video",
		{"", "text/plain"}:        "document",
		{"", "application/x-bin"}: "other",
	}
	for in, want := range cases {
		if got := FileCategoryOf(in[0], in[1]); got != want {
			t.Errorf("FileCategoryOf(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

// 统计与按类型列出覆盖全部目录，只统计本人有效文件
func TestDriveStatsAndCategory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entity.CloudItem{}); err != nil {
		t.Fatal(err)
	}
	str := func(v string) *string { return &v }
	size := func(v int64) *int64 { return &v }
	parent := uint(1)
	items := []entity.CloudItem{
		{ItemType: "folder", Name: "素材", OwnerID: 7},
		{ItemType: "file", Name: "a.mp4", OwnerID: 7, FileExt: str(".mp4"), FileSize: size(100)},
		{ItemType: "file", Name: "b.mov", OwnerID: 7, FileExt: str(".mov"), FileSize: size(200), ParentID: &parent},
		{ItemType: "file", Name: "c.jpg", OwnerID: 7, FileExt: str(".jpg"), FileSize: size(10), ParentID: &parent},
		{ItemType: "file", Name: "d.pdf", OwnerID: 7, FileExt: str(".pdf"), FileSize: size(5)},
		{ItemType: "file", Name: "e.zip", OwnerID: 7, FileExt: str(".zip"), FileSize: size(1)},
		{ItemType: "file", Name: "x.mp4", OwnerID: 8, FileExt: str(".mp4"), FileSize: size(999)},
	}
	for i := range items {
		items[i].IsActive = "Y"
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	deleted := entity.CloudItem{ItemType: "file", Name: "old.mp4", OwnerID: 7, FileExt: str(".mp4")}
	deleted.IsActive = "N"
	db.Create(&deleted)

	s := &service{db: db}
	stats, err := s.GetDriveStats(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	want := DriveStats{All: 5, Video: 2, Image: 1, Document: 1, Other: 1, Folders: 1, TotalSize: 316}
	if *stats != want {
		t.Fatalf("stats = %+v, want %+v", *stats, want)
	}

	videos, err := s.ListFilesByCategory(context.Background(), 7, "video")
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 2 {
		t.Fatalf("视频应为 2 个（含子目录），实际 %d 个", len(videos))
	}
	if _, err := s.ListFilesByCategory(context.Background(), 7, "unknown"); err == nil {
		t.Error("未知分类应返回错误")
	}
}
