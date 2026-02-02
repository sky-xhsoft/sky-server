package main

import (
	"context"
	"log"

	"github.com/sky-xhsoft/sky-server/internal/config"
	tencentLive "github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
	"github.com/sky-xhsoft/sky-server/internal/service/live"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 检查是否启用直播功能
	if !cfg.TencentCloud.Live.Enabled {
		log.Println("直播功能未启用，请在配置文件中设置 tencentCloud.live.enabled = true")
		return
	}

	// 创建直播服务
	liveService, err := live.NewService(&cfg.TencentCloud)
	if err != nil {
		log.Fatalf("创建直播服务失败: %v", err)
	}

	ctx := context.Background()

	// 示例1: 查询域名列表
	log.Println("=== 查询域名列表 ===")
	domains, err := liveService.ListDomains(ctx, nil)
	if err != nil {
		log.Printf("查询域名列表失败: %v", err)
	} else {
		log.Printf("找到 %d 个域名", len(domains))
		for _, domain := range domains {
			log.Printf("  - 域名: %s, 类型: %d, 状态: %d", domain.Name, domain.Type, domain.Status)
		}
	}

	// 示例2: 查询拉流任务
	log.Println("\n=== 查询拉流任务 ===")
	tasks, err := liveService.DescribePullStreamTasks(ctx, nil)
	if err != nil {
		log.Printf("查询拉流任务失败: %v", err)
	} else {
		log.Printf("找到 %d 个拉流任务", len(tasks))
		for _, task := range tasks {
			log.Printf("  - 任务ID: %s, 状态: %s, 流名称: %s", task.TaskID, task.Status, task.StreamName)
		}
	}

	// 示例3: 查询录制模板
	log.Println("\n=== 查询录制模板 ===")
	templates, err := liveService.DescribeRecordTemplates(ctx)
	if err != nil {
		log.Printf("查询录制模板失败: %v", err)
	} else {
		log.Printf("找到 %d 个录制模板", len(templates))
		for _, tpl := range templates {
			log.Printf("  - 模板ID: %d, 名称: %s", tpl.TemplateID, tpl.TemplateName)
		}
	}

	// 示例4: 创建录制模板（取消注释以测试）
	/*
		log.Println("\n=== 创建录制模板 ===")
		templateID, err := liveService.CreateRecordTemplate(ctx, &tencentLive.CreateRecordTemplateRequest{
			TemplateName: "测试录制模板",
			Description:  "通过 Go SDK 创建的测试模板",
			HlsParam: &tencentLive.RecordParam{
				RecordInterval: 3600,  // 1小时
				StorageTime:    86400, // 24小时
				Enable:         1,     // 启用
			},
		})
		if err != nil {
			log.Printf("创建录制模板失败: %v", err)
		} else {
			log.Printf("创建录制模板成功，模板ID: %d", templateID)
		}
	*/

	log.Println("\n=== 测试完成 ===")
}
