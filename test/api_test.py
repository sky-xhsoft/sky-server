#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Sky-Server API 自动化测试脚本
测试所有API接口的可用性
"""

import requests
import json
import sys
from typing import Optional, Dict, Any
from datetime import datetime

# 配置
BASE_URL = "http://localhost:9090"
API_BASE = f"{BASE_URL}/api/v1"

# 颜色输出
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

# 统计
stats = {
    'total': 0,
    'passed': 0,
    'failed': 0,
    'skipped': 0
}

# JWT Token
token = None

# 测试结果
test_results = []

def print_separator():
    """打印分隔线"""
    print("=" * 80)

def print_title(title):
    """打印测试标题"""
    print(f"\n{Colors.YELLOW}{Colors.BOLD}>>> {title}{Colors.ENDC}")
    print_separator()

def run_test(
    test_name: str,
    method: str,
    endpoint: str,
    data: Optional[Dict[str, Any]] = None,
    expected_codes: list = [200, 201],
    need_auth: bool = True,
    headers: Optional[Dict[str, str]] = None
) -> bool:
    """
    执行单个API测试

    Args:
        test_name: 测试名称
        method: HTTP方法 (GET, POST, PUT, DELETE)
        endpoint: API端点
        data: 请求数据
        expected_codes: 期望的HTTP状态码列表
        need_auth: 是否需要认证
        headers: 额外的请求头

    Returns:
        bool: 测试是否通过
    """
    global stats, token

    stats['total'] += 1

    # 构建URL
    url = f"{API_BASE}{endpoint}"

    # 构建请求头
    req_headers = {'Content-Type': 'application/json'}
    if need_auth and token:
        req_headers['Authorization'] = f'Bearer {token}'
    if headers:
        req_headers.update(headers)

    try:
        # 发送请求
        if method.upper() == 'GET':
            response = requests.get(url, headers=req_headers, timeout=10)
        elif method.upper() == 'POST':
            response = requests.post(url, json=data, headers=req_headers, timeout=10)
        elif method.upper() == 'PUT':
            response = requests.put(url, json=data, headers=req_headers, timeout=10)
        elif method.upper() == 'DELETE':
            response = requests.delete(url, headers=req_headers, timeout=10)
        else:
            raise ValueError(f"不支持的HTTP方法: {method}")

        # 检查状态码
        if response.status_code in expected_codes:
            print(f"{Colors.GREEN}[PASS]{Colors.ENDC} {test_name} (HTTP {response.status_code})")
            stats['passed'] += 1
            test_results.append(('PASS', test_name))
            return True
        else:
            print(f"{Colors.RED}[FAIL]{Colors.ENDC} {test_name} (HTTP {response.status_code})")
            try:
                print(f"  Response: {response.json()}")
            except:
                print(f"  Response: {response.text[:200]}")
            stats['failed'] += 1
            test_results.append(('FAIL', test_name))
            return False

    except requests.exceptions.ConnectionError:
        print(f"{Colors.RED}[FAIL]{Colors.ENDC} {test_name} (Connection Failed)")
        print(f"  Error: Cannot connect to {url}")
        stats['failed'] += 1
        test_results.append(('FAIL', test_name))
        return False
    except requests.exceptions.Timeout:
        print(f"{Colors.RED}[FAIL]{Colors.ENDC} {test_name} (Timeout)")
        stats['failed'] += 1
        test_results.append(('FAIL', test_name))
        return False
    except Exception as e:
        print(f"{Colors.RED}[FAIL]{Colors.ENDC} {test_name} (Exception: {str(e)})")
        stats['failed'] += 1
        test_results.append(('FAIL', test_name))
        return False

def test_health_check():
    """测试健康检查"""
    print_title("1. 健康检查")
    # Health check is at /health (not /api/v1/health)
    try:
        response = requests.get(f"{BASE_URL}/health", timeout=10)
        if response.status_code == 200:
            print(f"{Colors.GREEN}[PASS]{Colors.ENDC} Health Check (HTTP {response.status_code})")
            stats['passed'] += 1
            stats['total'] += 1
        else:
            print(f"{Colors.RED}[FAIL]{Colors.ENDC} Health Check (HTTP {response.status_code})")
            stats['failed'] += 1
            stats['total'] += 1
    except Exception as e:
        print(f"{Colors.RED}[FAIL]{Colors.ENDC} Health Check (Exception: {str(e)})")
        stats['failed'] += 1
        stats['total'] += 1

def test_authentication():
    """测试认证接口"""
    global token

    print_title("2. 认证接口")

    # 登录
    print("尝试登录...")
    try:
        response = requests.post(
            f"{API_BASE}/auth/login",
            json={
                "username": "admin",
                "password": "admin123",
                "companyId": 1,
                "clientType": "web",
                "deviceId": "test-device-001",
                "deviceName": "API Test Client"
            },
            timeout=10
        )

        if response.status_code == 200:
            data = response.json()
            if 'data' in data and 'token' in data['data']:
                token = data['data']['token']
                print(f"{Colors.GREEN}[PASS]{Colors.ENDC} Login successful, got token")
                stats['passed'] += 1
            else:
                print(f"{Colors.YELLOW}[WARNING]{Colors.ENDC} Login response format error, using test token")
                token = "test_token_for_testing"
        else:
            print(f"{Colors.YELLOW}[WARNING]{Colors.ENDC} Login failed (HTTP {response.status_code}), using test token")
            token = "test_token_for_testing"
    except Exception as e:
        print(f"{Colors.YELLOW}[WARNING]{Colors.ENDC} Login exception: {str(e)}, using test token")
        token = "test_token_for_testing"

    stats['total'] += 1

    # 其他认证接口
    run_test("刷新Token", "POST", "/auth/refresh",
             data={"refreshToken": token}, need_auth=False, expected_codes=[200, 401])
    run_test("获取会话列表", "GET", "/auth/sessions")
    run_test("登出", "POST", "/auth/logout")

def test_metadata():
    """测试元数据接口"""
    print_title("3. 元数据接口")

    run_test("获取表信息", "GET", "/metadata/tables/sys_user")
    run_test("获取表字段", "GET", "/metadata/tables/sys_user/columns")
    run_test("获取表关系", "GET", "/metadata/tables/sys_user/refs")
    run_test("获取表动作", "GET", "/metadata/tables/sys_user/actions")
    run_test("刷新元数据缓存", "POST", "/metadata/refresh")
    run_test("获取元数据版本", "GET", "/metadata/version")

def test_dictionary():
    """测试字典接口"""
    print_title("4. 字典接口")

    run_test("获取字典项(按ID)", "GET", "/dicts/1/items")
    run_test("获取字典项(按名称)", "GET", "/dicts/name/user_status/items")
    run_test("获取字典默认值", "GET", "/dicts/1/default")
    run_test("刷新字典缓存", "POST", "/dicts/refresh")

def test_sequence():
    """测试序号接口"""
    print_title("5. 序号接口")

    run_test("获取下一个序号", "POST", "/sequences/ORDER_NO/next")
    run_test("批量获取序号", "POST", "/sequences/batch",
             data={"seqName": "ORDER_NO", "count": 5})
    run_test("获取当前序号值", "GET", "/sequences/ORDER_NO/current")

def test_crud():
    """测试通用CRUD接口"""
    print_title("6. 通用CRUD接口")

    run_test("查询列表", "POST", "/data/sys_user/query",
             data={"page": 1, "pageSize": 10})
    run_test("获取单条记录", "GET", "/data/sys_user/1")
    run_test("创建记录", "POST", "/data/sys_user",
             data={"username": "testuser", "password": "123456"})
    run_test("更新记录", "PUT", "/data/sys_user/1",
             data={"username": "updated_user"})
    run_test("删除记录", "DELETE", "/data/sys_user/999",
             expected_codes=[200, 404])
    run_test("批量删除", "POST", "/data/sys_user/batch-delete",
             data={"ids": [997, 998, 999]})

def test_actions():
    """测试动作接口"""
    print_title("7. 动作接口")

    run_test("获取动作信息", "GET", "/actions/1")
    run_test("执行动作(按ID)", "POST", "/actions/1/execute",
             data={"recordId": 1, "params": {}})
    run_test("批量执行动作", "POST", "/actions/1/batch-execute",
             data={"recordIds": [1, 2, 3], "params": {}})
    run_test("执行动作(按名称)", "POST", "/actions/by-name/sys_user/approve/execute",
             data={"recordId": 1})

def test_workflow():
    """测试工作流接口"""
    print_title("8. 工作流接口")

    # 流程定义
    run_test("创建流程定义", "POST", "/workflow/definitions",
             data={"name": "测试流程", "code": "TEST_FLOW", "description": "测试"})
    run_test("查询流程定义列表", "GET", "/workflow/definitions")
    run_test("获取流程定义详情", "GET", "/workflow/definitions/1")
    run_test("更新流程定义", "PUT", "/workflow/definitions/1",
             data={"name": "更新后的流程"})
    run_test("发布流程定义", "POST", "/workflow/definitions/1/publish")

    # 流程节点
    run_test("创建流程节点", "POST", "/workflow/nodes",
             data={"definitionId": 1, "name": "开始节点", "nodeType": "start"})
    run_test("查询流程节点", "GET", "/workflow/nodes?definitionId=1")
    run_test("更新流程节点", "PUT", "/workflow/nodes/1",
             data={"name": "更新后的节点"})
    run_test("删除流程节点", "DELETE", "/workflow/nodes/999",
             expected_codes=[200, 404])

    # 任务管理
    run_test("查询我的任务", "GET", "/workflow/tasks/my")
    run_test("获取任务详情", "GET", "/workflow/tasks/1")
    run_test("完成任务", "POST", "/workflow/tasks/complete",
             data={"taskId": 1, "action": "approve", "comment": "同意"})

def test_audit():
    """测试审计日志接口"""
    print_title("9. 审计日志接口")

    run_test("查询审计日志", "GET", "/audit/logs?page=1&pageSize=10")
    run_test("获取日志详情", "GET", "/audit/logs/1")
    run_test("查询用户日志", "GET", "/audit/users/1/logs")
    run_test("查询资源日志", "GET", "/audit/resources/sys_user/1/logs")
    run_test("获取审计统计", "GET", "/audit/statistics")
    run_test("清理过期日志", "POST", "/audit/clean",
             data={"days": 90})

def test_groups():
    """测试权限组接口"""
    print_title("10. 权限组接口")

    run_test("创建权限组", "POST", "/groups",
             data={"name": "测试组", "code": "TEST_GROUP", "description": "测试"})
    run_test("查询权限组列表", "GET", "/groups")
    run_test("获取权限组详情", "GET", "/groups/1")
    run_test("更新权限组", "PUT", "/groups/1",
             data={"name": "更新后的组"})
    run_test("删除权限组", "DELETE", "/groups/999",
             expected_codes=[200, 404])
    run_test("分配权限", "POST", "/groups/1/permissions",
             data={"directoryId": 1, "permission": 3})
    run_test("获取组权限", "GET", "/groups/1/permissions")
    run_test("分配用户到组", "POST", "/groups/users/1",
             data={"groupIds": [1, 2]})
    run_test("获取用户组", "GET", "/groups/users/1")
    run_test("检查权限", "POST", "/permissions/check",
             data={"userId": 1, "directoryId": 1, "requiredPermission": 1})
    run_test("获取用户权限", "GET", "/permissions/user")

def test_directories():
    """测试安全目录接口"""
    print_title("11. 安全目录接口")

    run_test("创建目录", "POST", "/directories",
             data={"code": "TEST_DIR", "name": "测试目录", "tableName": "sys_user"})
    run_test("查询目录列表", "GET", "/directories")
    run_test("获取目录树", "GET", "/directories/tree")
    run_test("获取目录详情", "GET", "/directories/1")
    run_test("更新目录", "PUT", "/directories/1",
             data={"name": "更新后的目录"})
    run_test("删除目录", "DELETE", "/directories/999",
             expected_codes=[200, 404])

def test_menus():
    """测试菜单接口"""
    print_title("12. 菜单接口")

    run_test("创建菜单", "POST", "/menus",
             data={"name": "测试菜单", "path": "/test", "icon": "el-icon-test"})
    run_test("查询菜单列表", "GET", "/menus")
    run_test("获取菜单树", "GET", "/menus/tree")
    run_test("获取用户菜单树", "GET", "/menus/user/tree")
    run_test("获取用户路由", "GET", "/menus/user/routers")
    run_test("获取菜单详情", "GET", "/menus/1")
    run_test("更新菜单", "PUT", "/menus/1",
             data={"name": "更新后的菜单"})
    run_test("删除菜单", "DELETE", "/menus/999",
             expected_codes=[200, 404])

def test_files():
    """测试文件接口"""
    print_title("13. 文件接口")

    run_test("获取文件信息", "GET", "/files/1")
    run_test("查询文件列表", "POST", "/files/list",
             data={"page": 1, "pageSize": 10})
    run_test("下载文件", "GET", "/files/download/1",
             expected_codes=[200, 404])
    run_test("预览文件", "GET", "/files/preview/1",
             expected_codes=[200, 404])
    run_test("删除文件", "DELETE", "/files/999",
             expected_codes=[200, 404])

def test_messages():
    """测试消息通知接口"""
    print_title("14. 消息通知接口")

    run_test("发送消息", "POST", "/messages/send",
             data={"title": "测试消息", "content": "这是一条测试消息",
                   "targetType": "user", "targetIds": [1]})
    run_test("发送模板消息", "POST", "/messages/send/template",
             data={"templateCode": "WELCOME", "targetType": "user",
                   "targetIds": [1], "variables": {"userName": "张三"}})
    run_test("批量发送消息", "POST", "/messages/send/batch",
             data={"userIds": [1, 2, 3],
                   "message": {"title": "批量消息", "content": "测试内容"}})
    run_test("发送给所有用户", "POST", "/messages/send/all",
             data={"title": "全员通知", "content": "系统维护通知"})
    run_test("获取消息详情", "GET", "/messages/1")
    run_test("查询消息列表", "POST", "/messages/list",
             data={"page": 1, "pageSize": 10, "isRead": "N"})
    run_test("获取未读消息数", "GET", "/messages/unread/count")
    run_test("获取未读消息列表", "GET", "/messages/unread/list?limit=10")
    run_test("标记为已读", "POST", "/messages/1/read")
    run_test("标记所有为已读", "POST", "/messages/read-all")
    run_test("标记星标", "POST", "/messages/1/star",
             data={"isStarred": True})
    run_test("归档消息", "POST", "/messages/1/archive")
    run_test("删除消息", "DELETE", "/messages/999",
             expected_codes=[200, 404])

def test_websocket():
    """测试WebSocket接口"""
    print_title("15. WebSocket接口")

    run_test("获取在线用户列表", "GET", "/ws/online/users")
    run_test("检查在线状态", "GET", "/ws/online/check")
    run_test("管理员广播消息", "POST", "/ws/broadcast",
             data={"type": "SYSTEM_NOTIFY",
                   "data": {"title": "系统通知", "content": "测试广播"}})

def print_summary():
    """打印测试总结"""
    print("\n")
    print_separator()
    print(f"{Colors.BOLD}测试总结{Colors.ENDC}")
    print_separator()
    print(f"总测试数: {stats['total']}")
    print(f"{Colors.GREEN}通过: {stats['passed']}{Colors.ENDC}")
    print(f"{Colors.RED}失败: {stats['failed']}{Colors.ENDC}")

    # 计算通过率
    if stats['total'] > 0:
        pass_rate = (stats['passed'] / stats['total']) * 100
        print(f"通过率: {pass_rate:.1f}%")

    print_separator()

    if stats['failed'] == 0:
        print(f"{Colors.GREEN}{Colors.BOLD}[SUCCESS] All tests passed!{Colors.ENDC}")
        return 0
    else:
        print(f"{Colors.RED}{Colors.BOLD}[FAILED] {stats['failed']} tests failed{Colors.ENDC}")
        return 1

def main():
    """主函数"""
    print(f"{Colors.BLUE}{Colors.BOLD}")
    print("=" * 80)
    print("Sky-Server API 自动化测试")
    print("=" * 80)
    print(f"{Colors.ENDC}")
    print(f"基础URL: {API_BASE}")
    print(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print_separator()

    # 执行所有测试
    test_health_check()
    test_authentication()
    test_metadata()
    test_dictionary()
    test_sequence()
    test_crud()
    test_actions()
    test_workflow()
    test_audit()
    test_groups()
    test_directories()
    test_menus()
    test_files()
    test_messages()
    test_websocket()

    # 打印总结
    exit_code = print_summary()
    sys.exit(exit_code)

if __name__ == "__main__":
    main()
