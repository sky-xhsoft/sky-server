// sanitize-audit-log 用与审计中间件相同的规则，遮盖 audit_log 历史记录里请求/响应体中的密码、密钥、令牌（幂等）
//
// 用法：go run ./cmd/sanitize-audit-log [-dry-run]
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sensitive"
	"github.com/sky-xhsoft/sky-server/internal/repository/mysql"
)

var dryRun = flag.Bool("dry-run", false, "只统计需要遮盖的记录数，不写库")

const batchSize = 500

type auditRow struct {
	ID           uint   `gorm:"column:ID"`
	RequestBody  string `gorm:"column:REQUEST_BODY"`
	ResponseBody string `gorm:"column:RESPONSE_BODY"`
}

// maskedCount 统计已遮盖值的个数，用于判断 Filter 是否遮盖了新的敏感值
func maskedCount(s string) int {
	return strings.Count(s, `"***"`) + strings.Count(s, "=***")
}

// sanitize 返回遮盖后的内容及是否有新的敏感值被遮盖（未变化时保留原文，避免无谓改写 JSON 格式）
func sanitize(body string) (string, bool) {
	if body == "" {
		return body, false
	}
	filtered := sensitive.Filter(body)
	if maskedCount(filtered) > maskedCount(body) {
		return filtered, true
	}
	return body, false
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

	var lastID uint
	scanned, changed := 0, 0
	for {
		var rows []auditRow
		if err := db.Raw(`SELECT ID, IFNULL(REQUEST_BODY, '') AS REQUEST_BODY, IFNULL(RESPONSE_BODY, '') AS RESPONSE_BODY
			FROM audit_log WHERE ID > ? AND (IFNULL(REQUEST_BODY, '') <> '' OR IFNULL(RESPONSE_BODY, '') <> '')
			ORDER BY ID LIMIT ?`, lastID, batchSize).Scan(&rows).Error; err != nil {
			fmt.Printf("查询审计日志失败: %v\n", err)
			os.Exit(1)
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			lastID = r.ID
			scanned++
			req, reqChanged := sanitize(r.RequestBody)
			resp, respChanged := sanitize(r.ResponseBody)
			if !reqChanged && !respChanged {
				continue
			}
			changed++
			if *dryRun {
				continue
			}
			if err := db.Exec("UPDATE audit_log SET REQUEST_BODY = ?, RESPONSE_BODY = ? WHERE ID = ?", req, resp, r.ID).Error; err != nil {
				fmt.Printf("更新审计日志 ID=%d 失败: %v\n", r.ID, err)
				os.Exit(1)
			}
		}
	}

	action := "已遮盖"
	if *dryRun {
		action = "待遮盖"
	}
	fmt.Printf("扫描 %d 条带请求/响应体的记录，%s %d 条\n", scanned, action, changed)
}
