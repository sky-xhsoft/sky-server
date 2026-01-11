#!/bin/bash

# Sky-Server API 测试脚本
# 测试所有API接口的可用性

# 配置
BASE_URL="http://localhost:9090"
API_BASE="${BASE_URL}/api/v1"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# JWT Token (登录后获取)
TOKEN=""

# 测试结果记录
TEST_RESULTS=()

# 打印分隔线
print_separator() {
    echo "================================================================"
}

# 打印测试标题
print_title() {
    echo -e "\n${YELLOW}>>> $1${NC}"
    print_separator
}

# 执行测试
run_test() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=${5:-200}
    local need_auth=${6:-true}

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 构建curl命令
    local headers="-H 'Content-Type: application/json'"
    if [ "$need_auth" = "true" ] && [ -n "$TOKEN" ]; then
        headers="$headers -H 'Authorization: Bearer $TOKEN'"
    fi

    # 执行请求
    if [ -n "$data" ]; then
        response=$(eval curl -s -w "\n%{http_code}" -X $method "$headers" -d "'$data'" "${API_BASE}${endpoint}")
    else
        response=$(eval curl -s -w "\n%{http_code}" -X $method "$headers" "${API_BASE}${endpoint}")
    fi

    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')

    # 检查结果
    if [ "$http_code" -eq "$expected_code" ] || [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo -e "${GREEN}✓ PASS${NC} $test_name (HTTP $http_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        TEST_RESULTS+=("✓ $test_name")
    else
        echo -e "${RED}✗ FAIL${NC} $test_name (HTTP $http_code)"
        echo "  Response: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TEST_RESULTS+=("✗ $test_name")
    fi
}

# 1. 健康检查
print_title "1. 健康检查"
run_test "健康检查" "GET" "/health" "" 200 false

# 2. 认证测试
print_title "2. 认证接口"
login_response=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $login_response | grep -o '"accessToken":"[^"]*' | sed 's/"accessToken":"//')

if [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓ PASS${NC} 登录成功，获取到Token"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
else
    echo -e "${YELLOW}⚠ WARNING${NC} 登录失败或Token为空，使用测试Token"
    TOKEN="test_token_for_testing"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

run_test "刷新Token" "POST" "/auth/refresh" '{"refreshToken":"'$TOKEN'"}' 200 false
run_test "获取会话列表" "GET" "/auth/sessions" "" 200 true
run_test "登出" "POST" "/auth/logout" "" 200 true

# 3. 元数据接口
print_title "3. 元数据接口"
run_test "获取表信息" "GET" "/metadata/tables/sys_user" "" 200 true
run_test "获取表字段" "GET" "/metadata/tables/sys_user/columns" "" 200 true
run_test "获取表关系" "GET" "/metadata/tables/sys_user/refs" "" 200 true
run_test "获取表动作" "GET" "/metadata/tables/sys_user/actions" "" 200 true
run_test "刷新元数据缓存" "POST" "/metadata/refresh" "" 200 true
run_test "获取元数据版本" "GET" "/metadata/version" "" 200 true

# 4. 字典接口
print_title "4. 字典接口"
run_test "获取字典项(按ID)" "GET" "/dicts/1/items" "" 200 true
run_test "获取字典项(按名称)" "GET" "/dicts/name/user_status/items" "" 200 true
run_test "获取字典默认值" "GET" "/dicts/1/default" "" 200 true
run_test "刷新字典缓存" "POST" "/dicts/refresh" "" 200 true

# 5. 序号接口
print_title "5. 序号接口"
run_test "获取下一个序号" "POST" "/sequences/ORDER_NO/next" "" 200 true
run_test "批量获取序号" "POST" "/sequences/batch" '{"seqName":"ORDER_NO","count":5}' 200 true
run_test "获取当前序号值" "GET" "/sequences/ORDER_NO/current" "" 200 true

# 6. 通用CRUD接口
print_title "6. 通用CRUD接口"
run_test "查询列表" "POST" "/data/sys_user/query" '{"page":1,"pageSize":10}' 200 true
run_test "获取单条记录" "GET" "/data/sys_user/1" "" 200 true
run_test "创建记录" "POST" "/data/sys_user" '{"username":"testuser","password":"123456"}' 200 true
run_test "更新记录" "PUT" "/data/sys_user/1" '{"username":"updated_user"}' 200 true
run_test "删除记录" "DELETE" "/data/sys_user/999" "" 200 true
run_test "批量删除" "POST" "/data/sys_user/batch-delete" '{"ids":[997,998,999]}' 200 true

# 7. 动作接口
print_title "7. 动作接口"
run_test "获取动作信息" "GET" "/actions/1" "" 200 true
run_test "执行动作(按ID)" "POST" "/actions/1/execute" '{"recordId":1,"params":{}}' 200 true
run_test "批量执行动作" "POST" "/actions/1/batch-execute" '{"recordIds":[1,2,3],"params":{}}' 200 true
run_test "执行动作(按名称)" "POST" "/actions/by-name/sys_user/approve/execute" '{"recordId":1}' 200 true

# 8. 工作流接口
print_title "8. 工作流接口"

# 流程定义
run_test "创建流程定义" "POST" "/workflow/definitions" \
    '{"name":"测试流程","code":"TEST_FLOW","description":"测试流程"}' 200 true
run_test "查询流程定义列表" "GET" "/workflow/definitions" "" 200 true
run_test "获取流程定义详情" "GET" "/workflow/definitions/1" "" 200 true
run_test "更新流程定义" "PUT" "/workflow/definitions/1" '{"name":"更新后的流程"}' 200 true
run_test "发布流程定义" "POST" "/workflow/definitions/1/publish" "" 200 true

# 流程节点
run_test "创建流程节点" "POST" "/workflow/nodes" \
    '{"definitionId":1,"name":"开始节点","nodeType":"start"}' 200 true
run_test "查询流程节点" "GET" "/workflow/nodes?definitionId=1" "" 200 true
run_test "更新流程节点" "PUT" "/workflow/nodes/1" '{"name":"更新后的节点"}' 200 true
run_test "删除流程节点" "DELETE" "/workflow/nodes/999" "" 200 true

# 流程流转
run_test "创建流程流转" "POST" "/workflow/transitions" \
    '{"definitionId":1,"fromNodeId":1,"toNodeId":2}' 200 true
run_test "查询流程流转" "GET" "/workflow/transitions?definitionId=1" "" 200 true
run_test "删除流程流转" "DELETE" "/workflow/transitions/999" "" 200 true

# 流程实例
run_test "启动流程实例" "POST" "/workflow/instances/start" \
    '{"definitionId":1,"businessKey":"TEST001","variables":{}}' 200 true
run_test "查询流程实例列表" "GET" "/workflow/instances" "" 200 true
run_test "获取流程实例详情" "GET" "/workflow/instances/1" "" 200 true
run_test "终止流程实例" "POST" "/workflow/instances/1/terminate" '{"reason":"测试终止"}' 200 true

# 任务管理
run_test "查询我的任务" "GET" "/workflow/tasks/my" "" 200 true
run_test "获取任务详情" "GET" "/workflow/tasks/1" "" 200 true
run_test "完成任务" "POST" "/workflow/tasks/complete" \
    '{"taskId":1,"action":"approve","comment":"同意"}' 200 true
run_test "认领任务" "POST" "/workflow/tasks/1/claim" "" 200 true
run_test "转办任务" "POST" "/workflow/tasks/1/transfer" '{"targetUserId":2}' 200 true

# 9. 审计日志接口
print_title "9. 审计日志接口"
run_test "查询审计日志" "GET" "/audit/logs?page=1&pageSize=10" "" 200 true
run_test "获取日志详情" "GET" "/audit/logs/1" "" 200 true
run_test "查询用户日志" "GET" "/audit/users/1/logs" "" 200 true
run_test "查询资源日志" "GET" "/audit/resources/sys_user/1/logs" "" 200 true
run_test "获取审计统计" "GET" "/audit/statistics" "" 200 true
run_test "清理过期日志" "POST" "/audit/clean" '{"days":90}' 200 true

# 10. 权限组接口
print_title "10. 权限组接口"
run_test "创建权限组" "POST" "/groups" \
    '{"name":"测试组","code":"TEST_GROUP","description":"测试权限组"}' 200 true
run_test "查询权限组列表" "GET" "/groups" "" 200 true
run_test "获取权限组详情" "GET" "/groups/1" "" 200 true
run_test "更新权限组" "PUT" "/groups/1" '{"name":"更新后的组"}' 200 true
run_test "删除权限组" "DELETE" "/groups/999" "" 200 true
run_test "分配权限" "POST" "/groups/1/permissions" \
    '{"directoryId":1,"permission":3}' 200 true
run_test "获取组权限" "GET" "/groups/1/permissions" "" 200 true
run_test "分配用户到组" "POST" "/groups/users/1" '{"groupIds":[1,2]}' 200 true
run_test "获取用户组" "GET" "/groups/users/1" "" 200 true
run_test "检查权限" "POST" "/permissions/check" \
    '{"userId":1,"directoryId":1,"requiredPermission":1}' 200 true
run_test "获取用户权限" "GET" "/permissions/user" "" 200 true

# 11. 安全目录接口
print_title "11. 安全目录接口"
run_test "创建目录" "POST" "/directories" \
    '{"code":"TEST_DIR","name":"测试目录","tableName":"sys_user"}' 200 true
run_test "查询目录列表" "GET" "/directories" "" 200 true
run_test "获取目录树" "GET" "/directories/tree" "" 200 true
run_test "获取目录详情" "GET" "/directories/1" "" 200 true
run_test "更新目录" "PUT" "/directories/1" '{"name":"更新后的目录"}' 200 true
run_test "删除目录" "DELETE" "/directories/999" "" 200 true

# 12. 菜单接口
print_title "12. 菜单接口"
run_test "创建菜单" "POST" "/menus" \
    '{"name":"测试菜单","path":"/test","icon":"el-icon-test"}' 200 true
run_test "查询菜单列表" "GET" "/menus" "" 200 true
run_test "获取菜单树" "GET" "/menus/tree" "" 200 true
run_test "获取用户菜单树" "GET" "/menus/user/tree" "" 200 true
run_test "获取用户路由" "GET" "/menus/user/routers" "" 200 true
run_test "获取菜单详情" "GET" "/menus/1" "" 200 true
run_test "更新菜单" "PUT" "/menus/1" '{"name":"更新后的菜单"}' 200 true
run_test "删除菜单" "DELETE" "/menus/999" "" 200 true

# 13. 文件接口
print_title "13. 文件接口"
run_test "获取文件信息" "GET" "/files/1" "" 200 true
run_test "查询文件列表" "POST" "/files/list" '{"page":1,"pageSize":10}' 200 true
run_test "下载文件" "GET" "/files/download/1" "" 200 true
run_test "预览文件" "GET" "/files/preview/1" "" 200 true
run_test "删除文件" "DELETE" "/files/999" "" 200 true

# 14. 消息通知接口
print_title "14. 消息通知接口"
run_test "发送消息" "POST" "/messages/send" \
    '{"title":"测试消息","content":"这是一条测试消息","targetType":"user","targetIds":[1]}' 200 true
run_test "发送模板消息" "POST" "/messages/send/template" \
    '{"templateCode":"WELCOME","targetType":"user","targetIds":[1],"variables":{"userName":"张三"}}' 200 true
run_test "批量发送消息" "POST" "/messages/send/batch" \
    '{"userIds":[1,2,3],"message":{"title":"批量消息","content":"测试内容"}}' 200 true
run_test "发送给所有用户" "POST" "/messages/send/all" \
    '{"title":"全员通知","content":"系统维护通知"}' 200 true
run_test "获取消息详情" "GET" "/messages/1" "" 200 true
run_test "查询消息列表" "POST" "/messages/list" \
    '{"page":1,"pageSize":10,"isRead":"N"}' 200 true
run_test "获取未读消息数" "GET" "/messages/unread/count" "" 200 true
run_test "获取未读消息列表" "GET" "/messages/unread/list?limit=10" "" 200 true
run_test "标记为已读" "POST" "/messages/1/read" "" 200 true
run_test "标记所有为已读" "POST" "/messages/read-all" "" 200 true
run_test "标记星标" "POST" "/messages/1/star" '{"isStarred":true}' 200 true
run_test "归档消息" "POST" "/messages/1/archive" "" 200 true
run_test "删除消息" "DELETE" "/messages/999" "" 200 true

# 15. WebSocket接口
print_title "15. WebSocket接口"
run_test "获取在线用户列表" "GET" "/ws/online/users" "" 200 true
run_test "检查在线状态" "GET" "/ws/online/check" "" 200 true
run_test "管理员广播消息" "POST" "/ws/broadcast" \
    '{"type":"SYSTEM_NOTIFY","data":{"title":"系统通知","content":"测试广播"}}' 200 true

# 打印测试总结
print_separator
echo -e "\n${YELLOW}测试总结${NC}"
print_separator
echo "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 所有测试通过！${NC}"
    exit 0
else
    echo -e "\n${RED}❌ 有 $FAILED_TESTS 个测试失败${NC}"
    exit 1
fi
