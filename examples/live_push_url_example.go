package main

import (
	"fmt"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/tencent/live"
)

func main() {
	fmt.Println("=== 腾讯云直播推流地址生成示例 ===\n")

	// 配置参数（请替换为您的实际配置）
	pushDomain := "push.example.com"  // 推流域名
	streamKey := "your_stream_key"    // 推流密钥（在腾讯云控制台获取）
	appName := "live"                 // 应用名称
	playDomain := "play.example.com"  // 播放域名

	// 创建推流地址生成器
	generator := live.NewPushURLGenerator(pushDomain, streamKey, appName)

	// 示例1：生成推流地址
	fmt.Println("【示例1】生成推流地址")
	fmt.Println("----------------------------------------")

	streamName := "test_stream_001"
	req := &live.GeneratePushURLRequest{
		StreamName: streamName,
		ExpireTime: 24 * time.Hour, // 24小时后过期
	}

	resp, err := generator.GeneratePushURL(req)
	if err != nil {
		fmt.Printf("❌ 生成推流地址失败: %v\n", err)
		return
	}

	fmt.Printf("流名称: %s\n", streamName)
	fmt.Printf("推流地址: %s\n", resp.PushURL)
	fmt.Printf("\n【OBS 推流配置】\n")
	fmt.Printf("服务器: %s\n", resp.PushURLOBS)
	fmt.Printf("串流密钥: %s\n", resp.StreamKey)
	fmt.Printf("过期时间: %s\n", time.Unix(resp.ExpireTime, 0).Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 示例2：生成播放地址（不带鉴权）
	fmt.Println("【示例2】生成播放地址（不带鉴权）")
	fmt.Println("----------------------------------------")

	playURLs, err := generator.GeneratePlayURL(
		streamName,
		playDomain,
		"", // 不使用播放鉴权
		24*time.Hour,
	)
	if err != nil {
		fmt.Printf("❌ 生成播放地址失败: %v\n", err)
		return
	}

	fmt.Printf("RTMP播放地址: %s\n", playURLs["rtmp"])
	fmt.Printf("FLV播放地址: %s\n", playURLs["flv"])
	fmt.Printf("HLS播放地址: %s\n", playURLs["hls"])
	fmt.Println()

	// 示例3：生成播放地址（带鉴权）
	fmt.Println("【示例3】生成播放地址（带鉴权）")
	fmt.Println("----------------------------------------")

	playKey := "your_play_key" // 播放密钥（如果播放域名开启了鉴权）
	playURLsWithAuth, err := generator.GeneratePlayURL(
		streamName,
		playDomain,
		playKey,
		24*time.Hour,
	)
	if err != nil {
		fmt.Printf("❌ 生成播放地址失败: %v\n", err)
		return
	}

	fmt.Printf("RTMP播放地址: %s\n", playURLsWithAuth["rtmp"])
	fmt.Printf("FLV播放地址: %s\n", playURLsWithAuth["flv"])
	fmt.Printf("HLS播放地址: %s\n", playURLsWithAuth["hls"])
	fmt.Println()

	// 示例4：生成多个流的推流地址
	fmt.Println("【示例4】批量生成推流地址")
	fmt.Println("----------------------------------------")

	streamNames := []string{"stream_001", "stream_002", "stream_003"}
	for i, name := range streamNames {
		req := &live.GeneratePushURLRequest{
			StreamName: name,
			ExpireTime: 24 * time.Hour,
		}

		resp, err := generator.GeneratePushURL(req)
		if err != nil {
			fmt.Printf("❌ 生成流 %s 的推流地址失败: %v\n", name, err)
			continue
		}

		fmt.Printf("%d. 流名称: %s\n", i+1, name)
		fmt.Printf("   推流地址: %s\n", resp.PushURL)
	}
	fmt.Println()

	// 示例5：生成不同过期时间的推流地址
	fmt.Println("【示例5】生成不同过期时间的推流地址")
	fmt.Println("----------------------------------------")

	expireTimes := []struct {
		duration time.Duration
		desc     string
	}{
		{1 * time.Hour, "1小时"},
		{6 * time.Hour, "6小时"},
		{24 * time.Hour, "24小时"},
		{7 * 24 * time.Hour, "7天"},
	}

	for _, et := range expireTimes {
		req := &live.GeneratePushURLRequest{
			StreamName: "test_stream",
			ExpireTime: et.duration,
		}

		resp, err := generator.GeneratePushURL(req)
		if err != nil {
			fmt.Printf("❌ 生成推流地址失败: %v\n", err)
			continue
		}

		expireTime := time.Unix(resp.ExpireTime, 0)
		fmt.Printf("过期时间: %s | 有效期至: %s\n",
			et.desc,
			expireTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// 示例6：使用说明
	fmt.Println("【使用说明】")
	fmt.Println("----------------------------------------")
	fmt.Println("1. 在腾讯云控制台获取推流域名和推流密钥")
	fmt.Println("2. 将上述配置参数替换为您的实际配置")
	fmt.Println("3. 运行程序生成推流地址")
	fmt.Println("4. 在 OBS 中配置推流：")
	fmt.Println("   - 服务器：使用 pushUrlObs")
	fmt.Println("   - 串流密钥：使用 streamKey")
	fmt.Println("5. 开始推流后，使用播放地址观看直播")
	fmt.Println()

	fmt.Println("=== 示例程序结束 ===")
}
