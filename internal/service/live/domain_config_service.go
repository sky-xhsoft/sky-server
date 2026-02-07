package live

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PushURLResponse 推流地址响应
type PushURLResponse struct {
	PushURL       string `json:"pushUrl"`       // RTMP推流地址
	PushURLSRT    string `json:"pushUrlSrt"`    // SRT推流地址
	PushURLWebRTC string `json:"pushUrlWebRtc"` // WebRTC推流地址
}

// PushURLService 推流地址生成服务接口
type PushURLService interface {
	// GeneratePushURL 生成推流地址
	GeneratePushURL(domainName, appName, streamName, streamKey string, expireTime int64) (*PushURLResponse, error)

	// GeneratePlayURL 生成播放地址
	GeneratePlayURL(playDomain, appName, streamName, playKey string, expireTime int64) (map[string]string, error)
}

// pushURLService 推流地址生成服务实现
type pushURLService struct{}

// NewPushURLService 创建推流地址生成服务
func NewPushURLService() PushURLService {
	return &pushURLService{}
}

// GeneratePushURL 生成推流地址
// domainName: 推流域名
// appName: 应用名称
// streamName: 流名称
// streamKey: 推流密钥
// expireTime: 过期时间（Unix时间戳，秒）
func (s *pushURLService) GeneratePushURL(domainName, appName, streamName, streamKey string, expireTime int64) (*PushURLResponse, error) {
	if domainName == "" || appName == "" || streamName == "" {
		return nil, fmt.Errorf("域名、应用名称和流名称不能为空")
	}

	// 如果没有设置过期时间，默认24小时后过期
	if expireTime == 0 {
		expireTime = time.Now().Unix() + 86400
	}

	response := &PushURLResponse{}

	// 生成推流地址
	if streamKey != "" {
		// 有密钥时，生成带鉴权的推流地址
		// 鉴权格式: txTime=十六进制时间戳&txSecret=MD5(key+StreamName+txTime)
		// 注意：根据腾讯云官方文档，txTime 必须先转为大写，然后用大写的 txTime 计算 MD5
		txTime := strings.ToUpper(strconv.FormatInt(expireTime, 16))
		txSecret := s.generateTxSecret(streamKey, streamName, txTime)

		// RTMP推流地址
		response.PushURL = fmt.Sprintf("rtmp://%s/%s/%s?txSecret=%s&txTime=%s",
			domainName, appName, streamName, txSecret, txTime)

		// SRT推流地址
		// SRT格式：srt://domain:port?streamid=#!::h=domain/AppName/StreamName,txSecret=xxx,txTime=xxx
		response.PushURLSRT = fmt.Sprintf("srt://%s:9000?streamid=#!::h=%s/%s/%s,txSecret=%s,txTime=%s",
			domainName, domainName, appName, streamName, txSecret, txTime)

		// WebRTC推流地址
		// WebRTC格式：webrtc://domain/AppName/StreamName?txSecret=xxx&txTime=xxx
		response.PushURLWebRTC = fmt.Sprintf("webrtc://%s/%s/%s?txSecret=%s&txTime=%s",
			domainName, appName, streamName, txSecret, txTime)
	} else {
		// 无密钥时，生成普通推流地址
		response.PushURL = fmt.Sprintf("rtmp://%s/%s/%s", domainName, appName, streamName)
		response.PushURLSRT = fmt.Sprintf("srt://%s:9000?streamid=#!::h=%s/%s/%s",
			domainName, domainName, appName, streamName)
		response.PushURLWebRTC = fmt.Sprintf("webrtc://%s/%s/%s",
			domainName, appName, streamName)
	}

	return response, nil
}

// GeneratePlayURL 生成播放地址
// playDomain: 播放域名
// appName: 应用名称
// streamName: 流名称
// playKey: 播放密钥
// expireTime: 过期时间（Unix时间戳，秒）
func (s *pushURLService) GeneratePlayURL(playDomain, appName, streamName, playKey string, expireTime int64) (map[string]string, error) {
	if playDomain == "" || appName == "" || streamName == "" {
		return nil, fmt.Errorf("播放域名、应用名称和流名称不能为空")
	}

	// 如果没有设置过期时间，默认24小时后过期
	if expireTime == 0 {
		expireTime = time.Now().Unix() + 86400
	}

	result := make(map[string]string)

	// 生成鉴权参数
	var authParams string
	if playKey != "" {
		// 注意：根据腾讯云官方文档，txTime 必须先转为大写，然后用大写的 txTime 计算 MD5
		txTime := strings.ToUpper(strconv.FormatInt(expireTime, 16))
		txSecret := s.generateTxSecret(playKey, streamName, txTime)
		authParams = fmt.Sprintf("?txSecret=%s&txTime=%s", txSecret, txTime)
	}

	// RTMP播放地址
	result["rtmp"] = fmt.Sprintf("rtmp://%s/%s/%s%s", playDomain, appName, streamName, authParams)

	// FLV播放地址
	result["flv"] = fmt.Sprintf("http://%s/%s/%s.flv%s", playDomain, appName, streamName, authParams)

	// HLS播放地址
	result["hls"] = fmt.Sprintf("http://%s/%s/%s.m3u8%s", playDomain, appName, streamName, authParams)

	return result, nil
}

// generateTxSecret 生成腾讯云推流/播放鉴权的txSecret
// key: 密钥
// streamName: 流名称
// txTime: 十六进制时间戳
func (s *pushURLService) generateTxSecret(key, streamName, txTime string) string {
	// txSecret = MD5(key+StreamName+txTime)
	//str := key + streamName + txTime
	//hash := md5.Sum([]byte(str))
	//return hex.EncodeToString(hash[:])

	txSecret := md5.Sum([]byte(key + streamName + txTime))
	txSecretStr := fmt.Sprintf("%x", txSecret)
	return txSecretStr
}
