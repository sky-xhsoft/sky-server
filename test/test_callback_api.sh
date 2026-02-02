#!/bin/bash

# 测试直播回调事件查询接口
# 使用方法: ./test_callback_api.sh [your_jwt_token]

BASE_URL="http://c.skyzhou.cn:3000"
TOKEN="${1:-your_jwt_token_here}"

echo "=========================================="
echo "测试直播回调事件查询接口"
echo "=========================================="
echo ""

# 测试1: 查询高光切片事件
echo "1. 测试查询高光切片事件..."
curl -s -X GET "${BASE_URL}/api/v1/live/callback/events?eventType=highlight&pageNum=1&pageSize=20" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" | jq '.'

echo ""
echo "=========================================="
echo ""

# 测试2: 查询录制文件事件
echo "2. 测试查询录制文件事件..."
curl -s -X GET "${BASE_URL}/api/v1/live/callback/events?eventType=recording_file&pageNum=1&pageSize=20" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" | jq '.'

echo ""
echo "=========================================="
echo ""

# 测试3: 查询所有事件
echo "3. 测试查询所有事件..."
curl -s -X GET "${BASE_URL}/api/v1/live/callback/events?pageNum=1&pageSize=10" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" | jq '.'

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="
