#!/bin/bash

# 录制文件回调完整测试脚本
# 用于测试录制列表页面的字段显示

BASE_URL="${1:-http://localhost:9090}"
TIMESTAMP=$(date +%s)

echo "==========================================="
echo "发送录制文件回调测试数据"
echo "==========================================="
echo ""
echo "目标地址: ${BASE_URL}/api/v1/live/callback/recording-file"
echo "时间戳: ${TIMESTAMP}"
echo ""

# 发送完整的录制文件回调数据
curl -X POST "${BASE_URL}/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_001",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP}',
    "sign": "test_sign_'${TIMESTAMP}'",
    "event_time": '${TIMESTAMP}',
    "video_url": "https://example.com/record/test_stream_001.flv",
    "file_size": 10485760,
    "duration": 3600,
    "file_format": "flv",
    "start_time": '$((TIMESTAMP - 3600))',
    "end_time": '${TIMESTAMP}',
    "stream_param": "txSecret=abc123&txTime=5C2A3CFF",
    "video_id": "video_'${TIMESTAMP}'",
    "record_file_id": "file_'${TIMESTAMP}'",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "test_stream"
  }'

echo ""
echo ""
echo "==========================================="
echo "发送第二条测试数据（不同的流）"
echo "==========================================="
echo ""

sleep 1
TIMESTAMP2=$(date +%s)

curl -X POST "${BASE_URL}/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_002",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP2}',
    "sign": "test_sign_'${TIMESTAMP2}'",
    "event_time": '${TIMESTAMP2}',
    "video_url": "https://example.com/record/test_stream_002.mp4",
    "file_size": 20971520,
    "duration": 7200,
    "file_format": "mp4",
    "start_time": '$((TIMESTAMP2 - 7200))',
    "end_time": '${TIMESTAMP2}',
    "stream_param": "txSecret=def456&txTime=5C2A3D00",
    "video_id": "video_'${TIMESTAMP2}'",
    "record_file_id": "file_'${TIMESTAMP2}'",
    "push_domain": "push.example.com",
    "app_name": "live",
    "stream_name": "another_stream"
  }'

echo ""
echo ""
echo "==========================================="
echo "发送第三条测试数据（HLS格式）"
echo "==========================================="
echo ""

sleep 1
TIMESTAMP3=$(date +%s)

curl -X POST "${BASE_URL}/api/v1/live/callback/recording-file" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": 100,
    "stream_id": "test_stream_003",
    "channel_id": "test_channel",
    "t": '${TIMESTAMP3}',
    "sign": "test_sign_'${TIMESTAMP3}'",
    "event_time": '${TIMESTAMP3}',
    "video_url": "https://example.com/record/test_stream_003.m3u8",
    "file_size": 15728640,
    "duration": 5400,
    "file_format": "hls",
    "start_time": '$((TIMESTAMP3 - 5400))',
    "end_time": '${TIMESTAMP3}',
    "stream_param": "",
    "video_id": "video_'${TIMESTAMP3}'",
    "record_file_id": "file_'${TIMESTAMP3}'",
    "push_domain": "live.example.com",
    "app_name": "myapp",
    "stream_name": "hls_stream"
  }'

echo ""
echo ""
echo "==========================================="
echo "测试完成！"
echo "==========================================="
echo ""
echo "已发送3条测试数据："
echo "1. FLV格式 - test_stream (1小时, 10MB)"
echo "2. MP4格式 - another_stream (2小时, 20MB)"
echo "3. HLS格式 - hls_stream (1.5小时, 15MB)"
echo ""
echo "请刷新录制列表页面查看结果"
echo ""
