#!/bin/bash

# ==========================================
# 云盘上传测试脚本
# ==========================================

# 配置
API_BASE="http://localhost:9090/api/v1"
TOKEN="YOUR_TOKEN_HERE"  # 请替换为真实的 JWT Token

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "云盘功能测试"
echo "=========================================="

# 检查 token
if [ "$TOKEN" = "YOUR_TOKEN_HERE" ]; then
    echo -e "${RED}❌ 请先在脚本中设置真实的 TOKEN${NC}"
    echo "获取 token 的方法："
    echo "1. 登录获取 token："
    echo "   curl -X POST $API_BASE/sso/login \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -d '{\"username\":\"admin\",\"password\":\"your_password\"}'"
    echo ""
    echo "2. 将返回的 accessToken 复制到脚本的 TOKEN 变量中"
    exit 1
fi

echo ""
echo "1️⃣  测试：查看配额"
echo "=========================================="
QUOTA_RESPONSE=$(curl -s -X GET "$API_BASE/cloud/quota" \
  -H "Authorization: Bearer $TOKEN")

echo "$QUOTA_RESPONSE" | jq '.'

# 解析配额
TOTAL_QUOTA=$(echo "$QUOTA_RESPONSE" | jq -r '.data.totalQuota // 0')
MAX_FILE_SIZE=$(echo "$QUOTA_RESPONSE" | jq -r '.data.maxFileSize // 0')

if [ "$TOTAL_QUOTA" != "0" ]; then
    TOTAL_GB=$(echo "scale=2; $TOTAL_QUOTA / (1024*1024*1024)" | bc)
    MAX_GB=$(echo "scale=2; $MAX_FILE_SIZE / (1024*1024*1024)" | bc)
    echo -e "${GREEN}✅ 配额加载成功${NC}"
    echo "   总空间: ${TOTAL_GB} GB"
    echo "   单文件限制: ${MAX_GB} GB"
else
    echo -e "${RED}❌ 配额加载失败${NC}"
fi

echo ""
echo "2️⃣  测试：查看文件夹列表"
echo "=========================================="
FOLDERS_RESPONSE=$(curl -s -X GET "$API_BASE/cloud/folders" \
  -H "Authorization: Bearer $TOKEN")

echo "$FOLDERS_RESPONSE" | jq '.'

FOLDER_COUNT=$(echo "$FOLDERS_RESPONSE" | jq '.data | length // 0')
echo "当前文件夹数量: $FOLDER_COUNT"

echo ""
echo "3️⃣  测试：创建文件夹"
echo "=========================================="
FOLDER_RESPONSE=$(curl -s -X POST "$API_BASE/cloud/folders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试文件夹",
    "description": "自动化测试创建"
  }')

echo "$FOLDER_RESPONSE" | jq '.'

FOLDER_ID=$(echo "$FOLDER_RESPONSE" | jq -r '.data.id // empty')

if [ -n "$FOLDER_ID" ]; then
    echo -e "${GREEN}✅ 文件夹创建成功，ID: $FOLDER_ID${NC}"
else
    echo -e "${YELLOW}⚠️  文件夹创建失败或已存在${NC}"
    # 尝试获取现有文件夹
    FOLDER_ID=$(echo "$FOLDERS_RESPONSE" | jq -r '.data[0].id // empty')
    if [ -n "$FOLDER_ID" ]; then
        echo -e "${GREEN}使用现有文件夹，ID: $FOLDER_ID${NC}"
    fi
fi

echo ""
echo "4️⃣  测试：创建测试文件"
echo "=========================================="

# 创建 5MB 测试文件
TEST_FILE="test_5mb.bin"
if [ ! -f "$TEST_FILE" ]; then
    echo "创建 5MB 测试文件..."
    dd if=/dev/zero of="$TEST_FILE" bs=1M count=5 2>/dev/null
    echo -e "${GREEN}✅ 测试文件创建成功${NC}"
else
    echo "测试文件已存在"
fi

echo ""
echo "5️⃣  测试：上传文件到根目录"
echo "=========================================="
UPLOAD_ROOT_RESPONSE=$(curl -s -X POST "$API_BASE/cloud/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$TEST_FILE" \
  -F "folderId=")

echo "$UPLOAD_ROOT_RESPONSE" | jq '.'

UPLOAD_ROOT_CODE=$(echo "$UPLOAD_ROOT_RESPONSE" | jq -r '.code // 0')
if [ "$UPLOAD_ROOT_CODE" = "201" ] || [ "$UPLOAD_ROOT_CODE" = "200" ]; then
    echo -e "${GREEN}✅ 上传到根目录成功${NC}"
else
    echo -e "${RED}❌ 上传到根目录失败${NC}"
    ERROR_MSG=$(echo "$UPLOAD_ROOT_RESPONSE" | jq -r '.message // "未知错误"')
    echo "错误信息: $ERROR_MSG"
fi

if [ -n "$FOLDER_ID" ]; then
    echo ""
    echo "6️⃣  测试：上传文件到指定文件夹"
    echo "=========================================="
    UPLOAD_FOLDER_RESPONSE=$(curl -s -X POST "$API_BASE/cloud/files" \
      -H "Authorization: Bearer $TOKEN" \
      -F "file=@$TEST_FILE" \
      -F "folderId=$FOLDER_ID")

    echo "$UPLOAD_FOLDER_RESPONSE" | jq '.'

    UPLOAD_FOLDER_CODE=$(echo "$UPLOAD_FOLDER_RESPONSE" | jq -r '.code // 0')
    if [ "$UPLOAD_FOLDER_CODE" = "201" ] || [ "$UPLOAD_FOLDER_CODE" = "200" ]; then
        echo -e "${GREEN}✅ 上传到文件夹成功${NC}"
    else
        echo -e "${RED}❌ 上传到文件夹失败${NC}"
        ERROR_MSG=$(echo "$UPLOAD_FOLDER_RESPONSE" | jq -r '.message // "未知错误"')
        echo "错误信息: $ERROR_MSG"
    fi
fi

echo ""
echo "7️⃣  测试：查看文件列表"
echo "=========================================="
FILES_RESPONSE=$(curl -s -X GET "$API_BASE/cloud/files?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN")

echo "$FILES_RESPONSE" | jq '.'

FILE_COUNT=$(echo "$FILES_RESPONSE" | jq '.data.list | length // 0')
echo "当前文件数量: $FILE_COUNT"

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="

# 清理测试文件
read -p "是否删除测试文件 $TEST_FILE? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f "$TEST_FILE"
    echo "测试文件已删除"
fi
