// repair-cloud-media 修复直播回调写入的云盘视频：大小为 0（回调不带大小，旧版本没有去 COS 查询）、
// 名称只有扩展名（如 ".mp4"，旧版本切片标题为空时生成）。同时修复切片、录制列表中大小为 0 的记录。
//
// 大小用 HEAD 请求向 COS 查询（只查腾讯云域名）；查不到（如文件已删除）的保持不变并列出。
// 加 -update-quota 时，补上的大小同时计入文件所有者的云盘已用空间（与正常上传一致）；默认不改已用空间。
//
// 用法：go run ./cmd/repair-cloud-media                       只统计、预览，不写库
//
//	go run ./cmd/repair-cloud-media -apply                写库（只修大小和名称）
//	go run ./cmd/repair-cloud-media -apply -update-quota  写库，并把补上的大小计入已用空间
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
	"gorm.io/gorm"
)

var (
	apply       = flag.Bool("apply", false, "写库（默认只预览）")
	updateQuota = flag.Bool("update-quota", false, "把补上的大小计入文件所有者的云盘已用空间（默认不改）")
)

// 与 api/handler 中回调使用的域名一致
var tencentMediaHostSuffixes = []string{".myqcloud.com", ".vod-qcloud.com", ".qcloud.com", ".tencentcos.cn"}

var headClient = &http.Client{
	Timeout:       10 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

type cloudRow struct {
	ID          uint   `gorm:"column:ID"`
	Name        string `gorm:"column:NAME"`
	FileSize    int64  `gorm:"column:FILE_SIZE"`
	FileExt     string `gorm:"column:FILE_EXT"`
	OwnerID     uint   `gorm:"column:OWNER_ID"`
	StoragePath string `gorm:"column:STORAGE_PATH"`
	ParentID    *uint  `gorm:"column:PARENT_ID"`
	Path        string `gorm:"column:PATH"`
}

type mediaRow struct {
	ID       uint   `gorm:"column:ID"`
	URL      string `gorm:"column:URL"`
	FileSize int64  `gorm:"column:FILE_SIZE"`
}

func main() {
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	log, err := logger.Init(&cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	db, err := mysql.Init(&cfg.Database.MySQL, log)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer mysql.CloseDB(db)

	mode := "预览（不写库）"
	if *apply {
		mode = "写库"
	}
	fmt.Printf("模式：%s\n\n", mode)

	sizes := map[string]int64{} // 同一地址只查一次
	headSize := func(raw string) int64 {
		if v, ok := sizes[raw]; ok {
			return v
		}
		v := mediaFileSize(raw)
		sizes[raw] = v
		return v
	}

	repairCloud(db, headSize)
	repairMedia(db, "live_cut", "CLIP_URL", headSize)
	repairMedia(db, "live_recode", "RECORD_URL", headSize)
}

// repairCloud 云盘中大小为 0 或名称只有扩展名的视频文件
func repairCloud(db *gorm.DB, headSize func(string) int64) {
	var rows []cloudRow
	if err := db.Table("cloud_item").
		Select("ID, NAME, IFNULL(FILE_SIZE, 0) AS FILE_SIZE, IFNULL(FILE_EXT, '') AS FILE_EXT, OWNER_ID, IFNULL(STORAGE_PATH, '') AS STORAGE_PATH, PARENT_ID, IFNULL(PATH, '') AS PATH").
		Where("ITEM_TYPE = 'file' AND IS_ACTIVE = 'Y'").
		Where("IFNULL(FILE_SIZE, 0) = 0 OR NAME LIKE '.%'").
		Where("LOWER(IFNULL(FILE_TYPE, '')) LIKE 'video/%' OR LOWER(IFNULL(FILE_EXT, '')) IN ('mp4', '.mp4', 'flv', '.flv', 'm3u8', '.m3u8')").
		Order("ID").Find(&rows).Error; err != nil {
		fmt.Printf("查询云盘文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("== 云盘视频文件：待检查 %d 个\n", len(rows))

	var sizeFixed, nameFixed, notFound int
	ownerDelta := map[uint]int64{}
	for _, r := range rows {
		updates := map[string]interface{}{}
		if r.FileSize == 0 && isTencentMediaURL(r.StoragePath) {
			if n := headSize(r.StoragePath); n > 0 {
				updates["FILE_SIZE"] = n
				ownerDelta[r.OwnerID] += n
				sizeFixed++
			} else {
				notFound++
				fmt.Printf("  ID=%d 取不到大小（文件可能已删除）：%s\n", r.ID, r.StoragePath)
			}
		}
		if strings.HasPrefix(r.Name, ".") {
			newName := fixedName(db, r)
			updates["NAME"] = newName
			if strings.HasSuffix(r.Path, r.Name) {
				updates["PATH"] = strings.TrimSuffix(r.Path, r.Name) + newName
			}
			nameFixed++
			fmt.Printf("  ID=%d 名称 %q → %q\n", r.ID, r.Name, newName)
		}
		if len(updates) == 0 || !*apply {
			continue
		}
		if err := db.Table("cloud_item").Where("ID = ?", r.ID).Updates(updates).Error; err != nil {
			fmt.Printf("  ID=%d 更新失败: %v\n", r.ID, err)
		}
	}
	fmt.Printf("  补大小 %d 个，补名称 %d 个，取不到大小 %d 个\n", sizeFixed, nameFixed, notFound)
	for owner, delta := range ownerDelta {
		if !*updateQuota {
			fmt.Printf("  用户 %d 补上的大小合计 %s（未加 -update-quota，不改已用空间）\n", owner, humanSize(delta))
			continue
		}
		fmt.Printf("  用户 %d 已用空间增加 %s\n", owner, humanSize(delta))
		if *apply {
			if err := db.Table("cloud_quota").Where("USER_ID = ?", owner).
				Update("USED_SPACE", gorm.Expr("USED_SPACE + ?", delta)).Error; err != nil {
				fmt.Printf("  用户 %d 更新已用空间失败: %v\n", owner, err)
			}
		}
	}
	fmt.Println()
}

// fixedName 只有扩展名的文件名：用同一地址的切片名称，找不到用「高光切片_ID」
func fixedName(db *gorm.DB, r cloudRow) string {
	ext := path.Ext(r.Name)
	if ext == "" {
		ext = ".mp4"
	}
	var clipName string
	db.Table("live_cut").Select("CLIP_NAME").Where("CLIP_URL = ? AND IFNULL(CLIP_NAME, '') <> ''", r.StoragePath).Limit(1).Scan(&clipName)
	base := sanitizeName(clipName)
	if base == "" {
		base = fmt.Sprintf("高光切片_%d", r.ID)
	}
	return base + ext
}

// repairMedia 切片、录制列表中大小为 0 的记录
func repairMedia(db *gorm.DB, table, urlColumn string, headSize func(string) int64) {
	var rows []mediaRow
	if err := db.Table(table).
		Select(fmt.Sprintf("ID, IFNULL(%s, '') AS URL, IFNULL(FILE_SIZE, 0) AS FILE_SIZE", urlColumn)).
		Where("IS_ACTIVE = 'Y' AND IFNULL(FILE_SIZE, 0) = 0").
		Order("ID").Find(&rows).Error; err != nil {
		fmt.Printf("查询 %s 失败: %v\n", table, err)
		return
	}
	var fixed, notFound int
	for _, r := range rows {
		if !isTencentMediaURL(r.URL) {
			continue
		}
		n := headSize(r.URL)
		if n <= 0 {
			notFound++
			continue
		}
		fixed++
		if *apply {
			if err := db.Table(table).Where("ID = ?", r.ID).Update("FILE_SIZE", n).Error; err != nil {
				fmt.Printf("  %s ID=%d 更新失败: %v\n", table, r.ID, err)
			}
		}
	}
	fmt.Printf("== %s：大小为 0 的 %d 条，补大小 %d 条，取不到大小 %d 条\n\n", table, len(rows), fixed, notFound)
}

func sanitizeName(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', '*', '?', '"', '<', '>', '|':
			return -1
		case ':':
			return '_'
		}
		return r
	}, s))
}

func isTencentMediaURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, suffix := range tencentMediaHostSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

func mediaFileSize(raw string) int64 {
	resp, err := headClient.Head(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0
	}
	if resp.ContentLength > 0 {
		return resp.ContentLength
	}
	n, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	return n
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
