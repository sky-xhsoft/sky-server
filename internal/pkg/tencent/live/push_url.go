package live

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PushURLGenerator 推流地址生成器
type PushURLGenerator struct {
	pushDomain string // 推流域名
	streamKey  string // 推流密钥
	appName    string // 应用名称
}

// NewPushURLGenerator 创建推流地址生成器
func NewPushURLGenerator(pushDomain, streamKey, appName string) *PushURLGenerator {
	if appName == "" {
		appName = "live"
	}
	return &PushURLGenerator{
		pushDomain: pushDomain,
		streamKey:  streamKey,
		appName:    appName,
	}
}

// GeneratePushURLRequest 生成推流地址请求
type GeneratePushURLRequest struct {
	StreamName string        // 流名称（必填）
	ExpireTime time.Duration // 过期时间（可选，默认24小时）
}

// GeneratePushURLResponse 生成推流地址响应
type GeneratePushURLResponse struct {
	PushURL       string `json:"pushUrl"`       // RTMP推流地址
	PushURLOBS    string `json:"pushUrlObs"`    // OBS推流地址（服务器）
	StreamKey     string `json:"streamKey"`     // OBS推流密钥
	PushURLSRT    string `json:"pushUrlSrt"`    // SRT推流地址
	PushURLWebRTC string `json:"pushUrlWebRtc"` // WebRTC推流地址
	ExpireTime    int64  `json:"expireTime"`    // 过期时间戳
}

// GeneratePushURL 生成推流地址
// 腾讯云直播推流鉴权算法：
// 1. 计算过期时间戳 txTime（十六进制）
// 2. 计算签名 txSecret = MD5(key + StreamName + txTime)
// 3. 推流地址格式：rtmp://domain/AppName/StreamName?txSecret=xxx&txTime=xxx
func (g *PushURLGenerator) GeneratePushURL(req *GeneratePushURLRequest) (*GeneratePushURLResponse, error) {
	if req.StreamName == "" {
		return nil, fmt.Errorf("stream name is required")
	}

	// 设置默认过期时间为24小时
	expireTime := req.ExpireTime
	if expireTime == 0 {
		expireTime = 24 * time.Hour
	}

	// 计算过期时间戳（Unix时间戳）
	expireTimestamp := time.Now().Add(expireTime).Unix()

	// 将过期时间戳转换为十六进制（腾讯云要求）
	txTime := strconv.FormatInt(expireTimestamp, 16)

	// 计算签名：MD5(key + StreamName + txTime)
	// 注意：这里的key是推流密钥，StreamName是流名称
	signStr := g.streamKey + req.StreamName + txTime
	hash := md5.Sum([]byte(signStr))
	txSecret := fmt.Sprintf("%x", hash)

	// 生成RTMP推流地址
	pushURL := fmt.Sprintf("rtmp://%s/%s/%s?txSecret=%s&txTime=%s",
		g.pushDomain,
		g.appName,
		req.StreamName,
		txSecret,
		strings.ToUpper(txTime),
	)

	// 生成OBS推流地址（服务器部分）
	pushURLOBS := fmt.Sprintf("rtmp://%s/%s/",
		g.pushDomain,
		g.appName,
	)

	// 生成OBS推流密钥（流名称 + 鉴权参数）
	streamKey := fmt.Sprintf("%s?txSecret=%s&txTime=%s",
		req.StreamName,
		txSecret,
		strings.ToUpper(txTime),
	)

	// 生成SRT推流地址
	// SRT格式：srt://domain:port?streamid=#!::h=domain/AppName/StreamName,txSecret=xxx,txTime=xxx
	pushURLSRT := fmt.Sprintf("srt://%s:9000?streamid=#!::h=%s/%s/%s,txSecret=%s,txTime=%s",
		g.pushDomain,
		g.pushDomain,
		g.appName,
		req.StreamName,
		txSecret,
		strings.ToUpper(txTime),
	)

	// 生成WebRTC推流地址
	// WebRTC格式：webrtc://domain/AppName/StreamName?txSecret=xxx&txTime=xxx
	pushURLWebRTC := fmt.Sprintf("webrtc://%s/%s/%s?txSecret=%s&txTime=%s",
		g.pushDomain,
		g.appName,
		req.StreamName,
		txSecret,
		strings.ToUpper(txTime),
	)

	return &GeneratePushURLResponse{
		PushURL:       pushURL,
		PushURLOBS:    pushURLOBS,
		StreamKey:     streamKey,
		PushURLSRT:    pushURLSRT,
		PushURLWebRTC: pushURLWebRTC,
		ExpireTime:    expireTimestamp,
	}, nil
}

// GeneratePlayURL 生成播放地址
// 如果播放域名配置了鉴权，需要使用此方法生成带鉴权的播放地址
func (g *PushURLGenerator) GeneratePlayURL(streamName string, playDomain string, playKey string, expireTime time.Duration) (map[string]string, error) {
	if streamName == "" {
		return nil, fmt.Errorf("stream name is required")
	}
	if playDomain == "" {
		return nil, fmt.Errorf("play domain is required")
	}

	// 设置默认过期时间为24小时
	if expireTime == 0 {
		expireTime = 24 * time.Hour
	}

	// 计算过期时间戳
	expireTimestamp := time.Now().Add(expireTime).Unix()
	txTime := strconv.FormatInt(expireTimestamp, 16)

	// 生成播放地址（支持多种协议）
	playURLs := make(map[string]string)

	// 如果配置了播放密钥，生成带鉴权的播放地址
	if playKey != "" {
		// 计算签名：MD5(key + StreamName + txTime)
		signStr := playKey + streamName + txTime
		hash := md5.Sum([]byte(signStr))
		txSecret := fmt.Sprintf("%x", hash)

		// RTMP播放地址
		playURLs["rtmp"] = fmt.Sprintf("rtmp://%s/%s/%s?txSecret=%s&txTime=%s",
			playDomain,
			g.appName,
			streamName,
			txSecret,
			strings.ToUpper(txTime),
		)

		// FLV播放地址
		playURLs["flv"] = fmt.Sprintf("http://%s/%s/%s.flv?txSecret=%s&txTime=%s",
			playDomain,
			g.appName,
			streamName,
			txSecret,
			strings.ToUpper(txTime),
		)

		// HLS播放地址
		playURLs["hls"] = fmt.Sprintf("http://%s/%s/%s.m3u8?txSecret=%s&txTime=%s",
			playDomain,
			g.appName,
			streamName,
			txSecret,
			strings.ToUpper(txTime),
		)
	} else {
		// 不带鉴权的播放地址
		playURLs["rtmp"] = fmt.Sprintf("rtmp://%s/%s/%s",
			playDomain,
			g.appName,
			streamName,
		)

		playURLs["flv"] = fmt.Sprintf("http://%s/%s/%s.flv",
			playDomain,
			g.appName,
			streamName,
		)

		playURLs["hls"] = fmt.Sprintf("http://%s/%s/%s.m3u8",
			playDomain,
			g.appName,
			streamName,
		)
	}

	return playURLs, nil
}

// ValidatePushURL 验证推流地址是否有效
func (g *PushURLGenerator) ValidatePushURL(streamName, txSecret, txTime string) bool {
	// 将十六进制时间戳转换为十进制
	expireTimestamp, err := strconv.ParseInt(txTime, 16, 64)
	if err != nil {
		return false
	}

	// 检查是否过期
	if time.Now().Unix() > expireTimestamp {
		return false
	}

	// 计算签名并验证
	signStr := g.streamKey + streamName + txTime
	hash := md5.Sum([]byte(signStr))
	expectedSecret := fmt.Sprintf("%x", hash)

	return txSecret == expectedSecret
}
