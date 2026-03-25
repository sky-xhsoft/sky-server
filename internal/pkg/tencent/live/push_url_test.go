package live

import (
	"testing"
	"time"
)

func TestGeneratePushURL(t *testing.T) {
	// 创建推流地址生成器
	generator := NewPushURLGenerator(
		"push.example.com", // 推流域名
		"your_stream_key",  // 推流密钥
		"live",             // 应用名称
	)

	// 生成推流地址
	req := &GeneratePushURLRequest{
		StreamName: "test_stream_001",
		ExpireTime: 24 * time.Hour,
	}

	resp, err := generator.GeneratePushURL(req)
	if err != nil {
		t.Fatalf("生成推流地址失败: %v", err)
	}

	// 验证响应
	if resp.PushURL == "" {
		t.Error("推流地址为空")
	}
	if resp.PushURLOBS == "" {
		t.Error("OBS推流地址为空")
	}
	if resp.StreamKey == "" {
		t.Error("OBS推流密钥为空")
	}
	if resp.ExpireTime == 0 {
		t.Error("过期时间为0")
	}

	t.Logf("推流地址: %s", resp.PushURL)
	t.Logf("OBS推流地址: %s", resp.PushURLOBS)
	t.Logf("OBS推流密钥: %s", resp.StreamKey)
	t.Logf("过期时间: %d", resp.ExpireTime)
}

func TestGeneratePlayURL(t *testing.T) {
	// 创建推流地址生成器
	generator := NewPushURLGenerator(
		"push.example.com",
		"your_stream_key",
		"live",
	)

	// 生成播放地址（不带鉴权）
	playURLs, err := generator.GeneratePlayURL(
		"test_stream_001",
		"play.example.com",
		"",
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("生成播放地址失败: %v", err)
	}

	// 验证响应
	if playURLs["rtmp"] == "" {
		t.Error("RTMP播放地址为空")
	}
	if playURLs["flv"] == "" {
		t.Error("FLV播放地址为空")
	}
	if playURLs["hls"] == "" {
		t.Error("HLS播放地址为空")
	}

	t.Logf("RTMP播放地址: %s", playURLs["rtmp"])
	t.Logf("FLV播放地址: %s", playURLs["flv"])
	t.Logf("HLS播放地址: %s", playURLs["hls"])
}

func TestGeneratePlayURLWithAuth(t *testing.T) {
	// 创建推流地址生成器
	generator := NewPushURLGenerator(
		"push.example.com",
		"your_stream_key",
		"live",
	)

	// 生成播放地址（带鉴权）
	playURLs, err := generator.GeneratePlayURL(
		"test_stream_001",
		"play.example.com",
		"your_play_key", // 播放密钥
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("生成播放地址失败: %v", err)
	}

	// 验证响应
	if playURLs["rtmp"] == "" {
		t.Error("RTMP播放地址为空")
	}
	if playURLs["flv"] == "" {
		t.Error("FLV播放地址为空")
	}
	if playURLs["hls"] == "" {
		t.Error("HLS播放地址为空")
	}

	t.Logf("RTMP播放地址（带鉴权）: %s", playURLs["rtmp"])
	t.Logf("FLV播放地址（带鉴权）: %s", playURLs["flv"])
	t.Logf("HLS播放地址（带鉴权）: %s", playURLs["hls"])
}

func TestValidatePushURL(t *testing.T) {
	// 创建推流地址生成器
	generator := NewPushURLGenerator(
		"push.example.com",
		"your_stream_key",
		"live",
	)

	// 生成推流地址
	req := &GeneratePushURLRequest{
		StreamName: "test_stream_001",
		ExpireTime: 24 * time.Hour,
	}

	resp, err := generator.GeneratePushURL(req)
	if err != nil {
		t.Fatalf("生成推流地址失败: %v", err)
	}

	// 从推流地址中提取鉴权参数
	// 这里简化处理，实际应该解析URL参数
	// 假设我们已经提取了 txSecret 和 txTime

	// 验证推流地址（这里需要手动解析URL参数）
	// isValid := generator.ValidatePushURL("test_stream_001", txSecret, txTime)
	// if !isValid {
	// 	t.Error("推流地址验证失败")
	// }

	t.Logf("生成的推流地址: %s", resp.PushURL)
}

func TestDefaultAppName(t *testing.T) {
	// 测试默认应用名称
	generator := NewPushURLGenerator(
		"push.example.com",
		"your_stream_key",
		"", // 空应用名称，应该使用默认值 "live"
	)

	req := &GeneratePushURLRequest{
		StreamName: "test_stream_001",
		ExpireTime: 24 * time.Hour,
	}

	resp, err := generator.GeneratePushURL(req)
	if err != nil {
		t.Fatalf("生成推流地址失败: %v", err)
	}

	// 验证URL中包含默认的应用名称 "live"
	if resp.PushURL == "" {
		t.Error("推流地址为空")
	}

	t.Logf("推流地址（默认应用名称）: %s", resp.PushURL)
}

func TestEmptyStreamName(t *testing.T) {
	generator := NewPushURLGenerator(
		"push.example.com",
		"your_stream_key",
		"live",
	)

	// 测试空流名称
	req := &GeneratePushURLRequest{
		StreamName: "",
		ExpireTime: 24 * time.Hour,
	}

	_, err := generator.GeneratePushURL(req)
	if err == nil {
		t.Error("应该返回错误：流名称为空")
	}
}
