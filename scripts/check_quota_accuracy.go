package main

import (
	"fmt"
	"log"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/database"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 查询所有活跃的配额记录
	var quotas []entity.CloudQuota
	if err := db.Where("IS_ACTIVE = ?", "Y").Find(&quotas).Error; err != nil {
		log.Fatalf("查询配额失败: %v", err)
	}

	fmt.Println("=== 配额准确性检查 ===\n")

	for _, quota := range quotas {
		fmt.Printf("用户 ID: %d\n", quota.UserID)
		fmt.Println("----------------------------------------")

		// 配额表中的记录
		fmt.Printf("配额表记录:\n")
		fmt.Printf("  文件夹数量: %d\n", quota.FolderCount)
		fmt.Printf("  文件数量:   %d\n", quota.FileCount)
		fmt.Printf("  已用空间:   %d bytes (%.2f GB)\n", quota.UsedSpace, float64(quota.UsedSpace)/(1024*1024*1024))
		fmt.Printf("  总配额:     %d bytes (%.2f GB)\n", quota.TotalQuota, float64(quota.TotalQuota)/(1024*1024*1024))

		// 实际文件夹数量
		var actualFolderCount int64
		db.Model(&entity.CloudFolder{}).
			Where("OWNER_ID = ? AND IS_ACTIVE = ?", quota.UserID, "Y").
			Count(&actualFolderCount)

		// 实际文件数量和空间使用
		type FileStats struct {
			FileCount int64
			UsedSpace int64
		}
		var fileStats FileStats
		db.Model(&entity.CloudItem{}).
			Select("COUNT(*) as file_count, COALESCE(SUM(FILE_SIZE), 0) as used_space").
			Where("OWNER_ID = ? AND IS_ACTIVE = ? AND ITEM_TYPE = ?", quota.UserID, "Y", "file").
			Scan(&fileStats)

		// 实际数据
		fmt.Printf("\n实际数据统计:\n")
		fmt.Printf("  文件夹数量: %d\n", actualFolderCount)
		fmt.Printf("  文件数量:   %d\n", fileStats.FileCount)
		fmt.Printf("  已用空间:   %d bytes (%.2f GB)\n", fileStats.UsedSpace, float64(fileStats.UsedSpace)/(1024*1024*1024))

		// 差异检查
		fmt.Printf("\n差异检查:\n")
		folderDiff := int64(quota.FolderCount) - actualFolderCount
		fileDiff := int64(quota.FileCount) - fileStats.FileCount
		spaceDiff := quota.UsedSpace - fileStats.UsedSpace

		if folderDiff != 0 {
			fmt.Printf("  ❌ 文件夹数量不一致: 差异 %d\n", folderDiff)
		} else {
			fmt.Printf("  ✅ 文件夹数量一致\n")
		}

		if fileDiff != 0 {
			fmt.Printf("  ❌ 文件数量不一致: 差异 %d\n", fileDiff)
		} else {
			fmt.Printf("  ✅ 文件数量一致\n")
		}

		if spaceDiff != 0 {
			fmt.Printf("  ❌ 已用空间不一致: 差异 %d bytes (%.2f GB)\n", spaceDiff, float64(spaceDiff)/(1024*1024*1024))
		} else {
			fmt.Printf("  ✅ 已用空间一致\n")
		}

		fmt.Println()
	}

	// 如果发现不一致，提供修复建议
	fmt.Println("=== 修复建议 ===")
	fmt.Println("如果发现数据不一致，请运行修复脚本:")
	fmt.Println("  go run scripts/fix_quota_counts.go")
	fmt.Println()
}
