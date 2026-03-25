# 腾讯云直播回调事件测试

本文档提供了测试各类回调事件的curl命令示例。

## 环境变量

```bash
export BASE_URL="http://localhost:9090"
export CALLBACK_KEY="your_callback_key"
export TIMESTAMP=$(date +%s)
```

## 1. 推流事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/push" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 1,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "sequence": "seq_001",
    "node": "192.168.1.1",
    "user_ip": "1.2.3.4",
    "stream_param": "",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 2. 断流事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/disconnect" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 0,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "duration": 3600,
    "reason": "normal",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 3. 录制文件事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "video_url": "https://example.com/record.flv",
    "file_size": 1024000,
    "duration": 3600,
    "file_format": "flv",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 4. 截图事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/screenshot" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 200,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "pic_url": "https://example.com/screenshot.jpg",
    "width": 1920,
    "height": 1080,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 5. 画面审核事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/video-audit" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 317,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "confidence": 95,
    "label": "normal",
    "suggestion": "pass",
    "screenshot_url": "https://example.com/audit.jpg",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 6. 音频审核事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/audio-audit" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 318,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "confidence": 90,
    "label": "normal",
    "suggestion": "pass",
    "audio_text": "测试音频内容",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 7. 质检事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-inspection" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 319,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "diagnose_type": "video_freeze",
    "level": "warning",
    "description": "检测到视频卡顿",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 8. 评测阈值事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-threshold" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 320,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "metric_type": "bitrate",
    "threshold": 2000.0,
    "current_value": 1500.0,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 9. 评测平均分事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/quality-average" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 321,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "score": 85.5,
    "duration": 3600,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 10. 智能擦除事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/smart-erase" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 322,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "task_id": "task_001",
    "status": "success",
    "output_url": "https://example.com/erased.mp4",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 11. 直播字幕事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/subtitle" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 323,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "text": "这是实时生成的字幕内容",
    "language": "zh-CN",
    "start_time": '${TIMESTAMP}',
    "end_time": '$((TIMESTAMP + 5))',
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 12. 直播摘要事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/summary" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 324,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "summary": "本次直播的主要内容摘要",
    "keywords": ["关键词1", "关键词2", "关键词3"],
    "duration": 3600,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 13. 高光切片事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/highlight" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 325,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "clip_url": "https://example.com/highlight.mp4",
    "start_time": '${TIMESTAMP}',
    "end_time": '$((TIMESTAMP + 30))',
    "score": 9.5,
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 14. 推流异常事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/push-exception" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 326,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "error_code": 1001,
    "error_msg": "推流连接超时",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 15. 录制异常事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/record-exception" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 327,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "error_code": 2001,
    "error_msg": "录制存储空间不足",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 16. 拉流转推事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/pull-stream" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 328,
    "task_id": "task_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "status": "start",
    "source_url": "rtmp://source.example.com/live/stream",
    "target_url": "rtmp://target.example.com/live/stream",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 17. 监播事件测试

```bash
curl -X POST "${BASE_URL}/api/v1/live/callback/monitor" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 329,
    "stream_id": "test_stream_001",
    "t": '${TIMESTAMP}',
    "sign": "test_sign",
    "event_time": '${TIMESTAMP}',
    "alert_type": "black_screen",
    "alert_level": "error",
    "description": "检测到黑屏",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'
```

## 查询回调事件

```bash
# 需要先获取JWT token
export TOKEN="your_jwt_token"

# 查询所有事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events" \
  -H "Authorization: Bearer ${TOKEN}"

# 查询特定类型的事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events?eventType=push_stream" \
  -H "Authorization: Bearer ${TOKEN}"

# 查询特定流的事件
curl -X GET "${BASE_URL}/api/v1/live/callback/events?streamId=test_stream_001" \
  -H "Authorization: Bearer ${TOKEN}"

# 分页查询
curl -X GET "${BASE_URL}/api/v1/live/callback/events?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer ${TOKEN}"
```

## 批量测试脚本

创建一个测试脚本 `test_callbacks.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:9090"
TIMESTAMP=$(date +%s)

echo "测试推流事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/push" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":1,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "测试断流事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/disconnect" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":0,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"duration\":3600,\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "测试截图事件..."
curl -s -X POST "${BASE_URL}/api/v1/live/callback/screenshot" \
  -H "Content-Type: application/json" \
  -d "{\"event_type\":200,\"stream_id\":\"test_001\",\"t\":${TIMESTAMP},\"sign\":\"test\",\"event_time\":${TIMESTAMP},\"pic_url\":\"https://example.com/pic.jpg\",\"width\":1920,\"height\":1080,\"push_domain\":\"push.example.com\",\"app_name\":\"live\",\"stream_name\":\"test\"}" \
  | jq .

echo "所有测试完成！"
```

运行测试：

```bash
chmod +x test_callbacks.sh
./test_callbacks.sh
```
