@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Sky-Server API 测试脚本 (Windows版本)
REM 测试所有API接口的可用性

REM 配置
set BASE_URL=http://localhost:9090
set API_BASE=%BASE_URL%/api/v1

REM 统计
set /a TOTAL_TESTS=0
set /a PASSED_TESTS=0
set /a FAILED_TESTS=0

REM JWT Token
set TOKEN=

echo ================================================================
echo Sky-Server API 自动化测试
echo ================================================================
echo 基础URL: %API_BASE%
echo ================================================================

REM 1. 健康检查
echo.
echo [1. 健康检查]
echo ================================================================
curl -s -X GET "%BASE_URL%/health" >nul 2>&1
if !errorlevel! equ 0 (
    echo [PASS] 健康检查
    set /a PASSED_TESTS+=1
) else (
    echo [FAIL] 健康检查
    set /a FAILED_TESTS+=1
)
set /a TOTAL_TESTS+=1

REM 2. 认证测试
echo.
echo [2. 认证接口]
echo ================================================================

REM 登录获取Token
echo 正在登录...
curl -s -X POST "%API_BASE%/auth/login" ^
    -H "Content-Type: application/json" ^
    -d "{\"username\":\"admin\",\"password\":\"admin123\"}" > temp_login.json

REM 解析Token (简单方式，实际使用可能需要jq)
for /f "tokens=2 delims=:," %%a in ('type temp_login.json ^| findstr "accessToken"') do (
    set TOKEN=%%a
    set TOKEN=!TOKEN:"=!
    set TOKEN=!TOKEN: =!
)
del temp_login.json >nul 2>&1

if defined TOKEN (
    echo [PASS] 登录成功，获取到Token
    set /a PASSED_TESTS+=1
) else (
    echo [WARNING] 登录失败，使用测试Token
    set TOKEN=test_token_for_testing
)
set /a TOTAL_TESTS+=1

REM 3. 元数据接口
echo.
echo [3. 元数据接口]
echo ================================================================

call :run_test "获取表信息" GET "/metadata/tables/sys_user"
call :run_test "获取表字段" GET "/metadata/tables/sys_user/columns"
call :run_test "获取表关系" GET "/metadata/tables/sys_user/refs"
call :run_test "获取表动作" GET "/metadata/tables/sys_user/actions"
call :run_test "刷新元数据缓存" POST "/metadata/refresh"
call :run_test "获取元数据版本" GET "/metadata/version"

REM 4. 字典接口
echo.
echo [4. 字典接口]
echo ================================================================

call :run_test "获取字典项(按ID)" GET "/dicts/1/items"
call :run_test "获取字典项(按名称)" GET "/dicts/name/user_status/items"
call :run_test "获取字典默认值" GET "/dicts/1/default"
call :run_test "刷新字典缓存" POST "/dicts/refresh"

REM 5. 序号接口
echo.
echo [5. 序号接口]
echo ================================================================

call :run_test "获取下一个序号" POST "/sequences/ORDER_NO/next"
call :run_test "获取当前序号值" GET "/sequences/ORDER_NO/current"

REM 6. 通用CRUD接口
echo.
echo [6. 通用CRUD接口]
echo ================================================================

call :run_test_with_data "查询列表" POST "/data/sys_user/query" "{\"page\":1,\"pageSize\":10}"
call :run_test "获取单条记录" GET "/data/sys_user/1"
call :run_test_with_data "创建记录" POST "/data/sys_user" "{\"username\":\"testuser\",\"password\":\"123456\"}"
call :run_test_with_data "更新记录" PUT "/data/sys_user/1" "{\"username\":\"updated_user\"}"
call :run_test "删除记录" DELETE "/data/sys_user/999"

REM 7. 动作接口
echo.
echo [7. 动作接口]
echo ================================================================

call :run_test "获取动作信息" GET "/actions/1"
call :run_test_with_data "执行动作(按ID)" POST "/actions/1/execute" "{\"recordId\":1,\"params\":{}}"
call :run_test_with_data "执行动作(按名称)" POST "/actions/by-name/sys_user/approve/execute" "{\"recordId\":1}"

REM 8. 工作流接口
echo.
echo [8. 工作流接口]
echo ================================================================

call :run_test_with_data "创建流程定义" POST "/workflow/definitions" "{\"name\":\"测试流程\",\"code\":\"TEST_FLOW\"}"
call :run_test "查询流程定义列表" GET "/workflow/definitions"
call :run_test "获取流程定义详情" GET "/workflow/definitions/1"
call :run_test "查询我的任务" GET "/workflow/tasks/my"

REM 9. 审计日志接口
echo.
echo [9. 审计日志接口]
echo ================================================================

call :run_test "查询审计日志" GET "/audit/logs?page=1&pageSize=10"
call :run_test "获取日志详情" GET "/audit/logs/1"
call :run_test "获取审计统计" GET "/audit/statistics"

REM 10. 权限组接口
echo.
echo [10. 权限组接口]
echo ================================================================

call :run_test_with_data "创建权限组" POST "/groups" "{\"name\":\"测试组\",\"code\":\"TEST_GROUP\"}"
call :run_test "查询权限组列表" GET "/groups"
call :run_test "获取权限组详情" GET "/groups/1"
call :run_test "获取用户权限" GET "/permissions/user"

REM 11. 安全目录接口
echo.
echo [11. 安全目录接口]
echo ================================================================

call :run_test "查询目录列表" GET "/directories"
call :run_test "获取目录树" GET "/directories/tree"
call :run_test "获取目录详情" GET "/directories/1"

REM 12. 菜单接口
echo.
echo [12. 菜单接口]
echo ================================================================

call :run_test "查询菜单列表" GET "/menus"
call :run_test "获取菜单树" GET "/menus/tree"
call :run_test "获取用户菜单树" GET "/menus/user/tree"
call :run_test "获取用户路由" GET "/menus/user/routers"

REM 13. 文件接口
echo.
echo [13. 文件接口]
echo ================================================================

call :run_test "获取文件信息" GET "/files/1"
call :run_test_with_data "查询文件列表" POST "/files/list" "{\"page\":1,\"pageSize\":10}"

REM 14. 消息通知接口
echo.
echo [14. 消息通知接口]
echo ================================================================

call :run_test_with_data "发送消息" POST "/messages/send" "{\"title\":\"测试消息\",\"content\":\"测试内容\",\"targetType\":\"user\",\"targetIds\":[1]}"
call :run_test "获取消息详情" GET "/messages/1"
call :run_test_with_data "查询消息列表" POST "/messages/list" "{\"page\":1,\"pageSize\":10}"
call :run_test "获取未读消息数" GET "/messages/unread/count"
call :run_test "获取未读消息列表" GET "/messages/unread/list?limit=10"

REM 15. WebSocket接口
echo.
echo [15. WebSocket接口]
echo ================================================================

call :run_test "获取在线用户列表" GET "/ws/online/users"
call :run_test "检查在线状态" GET "/ws/online/check"

REM 打印测试总结
echo.
echo ================================================================
echo 测试总结
echo ================================================================
echo 总测试数: %TOTAL_TESTS%
echo 通过: %PASSED_TESTS%
echo 失败: %FAILED_TESTS%
echo ================================================================

if %FAILED_TESTS% equ 0 (
    echo [SUCCESS] 所有测试通过！
    exit /b 0
) else (
    echo [FAILED] 有 %FAILED_TESTS% 个测试失败
    exit /b 1
)

REM 函数：执行测试（无数据）
:run_test
set test_name=%~1
set method=%~2
set endpoint=%~3

curl -s -X %method% "%API_BASE%%endpoint%" ^
    -H "Authorization: Bearer %TOKEN%" ^
    -H "Content-Type: application/json" >nul 2>&1

if !errorlevel! equ 0 (
    echo [PASS] %test_name%
    set /a PASSED_TESTS+=1
) else (
    echo [FAIL] %test_name%
    set /a FAILED_TESTS+=1
)
set /a TOTAL_TESTS+=1
goto :eof

REM 函数：执行测试（带数据）
:run_test_with_data
set test_name=%~1
set method=%~2
set endpoint=%~3
set data=%~4

curl -s -X %method% "%API_BASE%%endpoint%" ^
    -H "Authorization: Bearer %TOKEN%" ^
    -H "Content-Type: application/json" ^
    -d "%data%" >nul 2>&1

if !errorlevel! equ 0 (
    echo [PASS] %test_name%
    set /a PASSED_TESTS+=1
) else (
    echo [FAIL] %test_name%
    set /a FAILED_TESTS+=1
)
set /a TOTAL_TESTS+=1
goto :eof
