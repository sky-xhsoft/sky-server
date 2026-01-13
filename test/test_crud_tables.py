#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Sky-Server CRUD API 测试脚本
测试 sys_table 和 sys_column 的增删改查操作
"""

import requests
import json
import sys
from datetime import datetime

# 配置
BASE_URL = "http://pan.xh-tec.cn:9090/api/v1"
USERNAME = "admin"
PASSWORD = "admin123"
COMPANY_ID = 1

# 颜色输出
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    BOLD = '\033[1m'
    END = '\033[0m'

def print_header(text):
    """打印测试标题"""
    print(f"\n{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}")
    print(f"{Colors.BLUE}{Colors.BOLD}{text}{Colors.END}")
    print(f"{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}\n")

def print_success(text):
    """打印成功信息"""
    print(f"{Colors.GREEN}[PASS]{Colors.END} {text}")

def print_fail(text, details=None):
    """打印失败信息"""
    print(f"{Colors.RED}[FAIL]{Colors.END} {text}")
    if details:
        print(f"  {Colors.RED}详情: {details}{Colors.END}")

def print_info(text):
    """打印信息"""
    print(f"{Colors.YELLOW}[INFO]{Colors.END} {text}")

class APITester:
    def __init__(self, base_url):
        self.base_url = base_url
        self.token = None
        self.headers = {"Content-Type": "application/json"}
        self.test_results = {"passed": 0, "failed": 0}

    def login(self):
        """登录并获取token"""
        print_header(">>> 1. 登录测试")

        url = f"{self.base_url}/auth/login"
        data = {
            "username": USERNAME,
            "password": PASSWORD,
            "clientType": "web",
            "deviceId": "test-device-001",
            "deviceName": "API Test Client"
        }

        try:
            response = requests.post(url, json=data)
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                self.token = result["data"]["token"]
                self.headers["Authorization"] = f"Bearer {self.token}"
                print_success(f"登录成功 (HTTP {response.status_code})")
                print_info(f"Token: {self.token[:50]}...")
                print_info(f"用户: {result['data']['user']['username']} ({result['data']['user']['trueName']})")
                self.test_results["passed"] += 1
                return True
            else:
                print_fail(f"登录失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
                return False
        except Exception as e:
            print_fail("登录异常", str(e))
            self.test_results["failed"] += 1
            return False

    def test_sys_table_crud(self):
        """测试 sys_table 的 CRUD 操作"""
        print_header(">>> 2. sys_table CRUD 测试")

        table_id = None

        # CREATE - 创建测试表
        print_info("步骤 1: 创建测试表")
        create_data = {
            "NAME": f"TEST_TABLE_{datetime.now().strftime('%Y%m%d%H%M%S')}",
            "DISPLAY_NAME": "测试表",
            "DESCRIPTION": "这是一个测试表",
            "MASK": "AMDSQ",
            "SYS_TABLECATEGORY_ID": 1,
            "IS_ACTIVE": "Y",
            "ORDERNO": 9999  # 添加ORDERNO字段，避免插件自动生成时查询错误
        }

        try:
            response = requests.post(
                f"{self.base_url}/data/sys_table",
                json=create_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code in [200, 201] and result.get("code") in [200, 201]:
                table_id = result["data"].get("id") or result["data"].get("ID") or result["data"].get("@id")
                print_success(f"创建表成功 (HTTP {response.status_code}), ID: {table_id}")
                print_info(f"表名: {create_data['NAME']}")
                self.test_results["passed"] += 1
            else:
                print_fail(f"创建表失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
                return None
        except Exception as e:
            print_fail("创建表异常", str(e))
            self.test_results["failed"] += 1
            return None

        # READ - 查询单条记录
        print_info(f"\n步骤 2: 查询表记录 (ID: {table_id})")
        try:
            response = requests.get(
                f"{self.base_url}/data/sys_table/{table_id}",
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                print_success(f"查询表成功 (HTTP {response.status_code})")
                data = result["data"]
                print_info(f"表名: {data.get('NAME')}")
                print_info(f"显示名称: {data.get('DISPLAY_NAME')}")
                print_info(f"描述: {data.get('DESCRIPTION')}")
                self.test_results["passed"] += 1
            else:
                print_fail(f"查询表失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("查询表异常", str(e))
            self.test_results["failed"] += 1

        # LIST - 查询列表
        print_info("\n步骤 3: 查询表列表")
        try:
            query_data = {
                "tableName": "sys_table",
                "page": 1,
                "pageSize": 10,
                "conditions": {
                    "NAME": create_data["NAME"]
                }
            }
            response = requests.post(
                f"{self.base_url}/data/sys_table/query",
                json=query_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                total = result["data"]["total"]
                print_success(f"查询列表成功 (HTTP {response.status_code}), 共 {total} 条记录")
                self.test_results["passed"] += 1
            else:
                print_fail(f"查询列表失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("查询列表异常", str(e))
            self.test_results["failed"] += 1

        # UPDATE - 更新记录
        print_info(f"\n步骤 4: 更新表记录 (ID: {table_id})")
        update_data = {
            "DISPLAY_NAME": "测试表(已更新)",
            "DESCRIPTION": "这是一个更新后的测试表"
        }

        try:
            response = requests.put(
                f"{self.base_url}/data/sys_table/{table_id}",
                json=update_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                print_success(f"更新表成功 (HTTP {response.status_code})")
                print_info(f"新显示名称: {update_data['DISPLAY_NAME']}")
                self.test_results["passed"] += 1
            else:
                print_fail(f"更新表失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("更新表异常", str(e))
            self.test_results["failed"] += 1

        # 验证更新
        print_info(f"\n步骤 5: 验证更新结果")
        try:
            response = requests.get(
                f"{self.base_url}/data/sys_table/{table_id}",
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                data = result["data"]
                if data.get("DISPLAY_NAME") == update_data["DISPLAY_NAME"]:
                    print_success("验证更新成功，数据已更新")
                    self.test_results["passed"] += 1
                else:
                    print_fail("验证更新失败", "更新的数据不匹配")
                    self.test_results["failed"] += 1
            else:
                print_fail(f"验证更新失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("验证更新异常", str(e))
            self.test_results["failed"] += 1

        return table_id

    def test_sys_column_crud(self, table_id):
        """测试 sys_column 的 CRUD 操作"""
        print_header(">>> 3. sys_column CRUD 测试")

        if not table_id:
            print_fail("跳过 sys_column 测试", "没有可用的 table_id")
            return

        column_id = None

        # CREATE - 创建测试字段
        print_info(f"步骤 1: 为表 (ID: {table_id}) 创建测试字段")
        create_data = {
            "SYS_TABLE_ID": table_id,
            "DB_NAME": f"TEST_FIELD_{datetime.now().strftime('%H%M%S')}",
            "DISPLAY_NAME": "测试字段",
            "COL_TYPE": "varchar",
            "COL_LENGTH": 255,
            "NULL_ABLE": "Y",
            "IS_ACTIVE": "Y",
            "DISPLAY_TYPE": "text",
            "SET_VALUE_TYPE": "byPage",
            "ORDERNO": 999
        }

        try:
            response = requests.post(
                f"{self.base_url}/data/sys_column",
                json=create_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code in [200, 201] and result.get("code") in [200, 201]:
                column_id = result["data"].get("id") or result["data"].get("ID") or result["data"].get("@id")
                print_success(f"创建字段成功 (HTTP {response.status_code}), ID: {column_id}")
                print_info(f"字段名: {create_data['DB_NAME']}")
                self.test_results["passed"] += 1
            else:
                print_fail(f"创建字段失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
                return None
        except Exception as e:
            print_fail("创建字段异常", str(e))
            self.test_results["failed"] += 1
            return None

        # READ - 查询单条记录
        print_info(f"\n步骤 2: 查询字段记录 (ID: {column_id})")
        try:
            response = requests.get(
                f"{self.base_url}/data/sys_column/{column_id}",
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                print_success(f"查询字段成功 (HTTP {response.status_code})")
                data = result["data"]
                print_info(f"字段名: {data.get('DB_NAME')}")
                print_info(f"显示名称: {data.get('NAME')}")
                print_info(f"数据类型: {data.get('COL_TYPE')}({data.get('COL_LENGTH')})")
                self.test_results["passed"] += 1
            else:
                print_fail(f"查询字段失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("查询字段异常", str(e))
            self.test_results["failed"] += 1

        # LIST - 查询字段列表
        print_info(f"\n步骤 3: 查询表 {table_id} 的字段列表")
        try:
            query_data = {
                "tableName": "sys_column",
                "page": 1,
                "pageSize": 10,
                "conditions": {
                    "SYS_TABLE_ID": table_id
                }
            }
            response = requests.post(
                f"{self.base_url}/data/sys_column/query",
                json=query_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                total = result["data"]["total"]
                print_success(f"查询字段列表成功 (HTTP {response.status_code}), 共 {total} 条记录")
                self.test_results["passed"] += 1
            else:
                print_fail(f"查询字段列表失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("查询字段列表异常", str(e))
            self.test_results["failed"] += 1

        # UPDATE - 更新字段
        print_info(f"\n步骤 4: 更新字段记录 (ID: {column_id})")
        update_data = {
            "DISPLAY_NAME": "测试字段(已更新)",
            "COL_LENGTH": 500
        }

        try:
            response = requests.put(
                f"{self.base_url}/data/sys_column/{column_id}",
                json=update_data,
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                print_success(f"更新字段成功 (HTTP {response.status_code})")
                print_info(f"新显示名称: {update_data['DISPLAY_NAME']}")
                print_info(f"新长度: {update_data['COL_LENGTH']}")
                self.test_results["passed"] += 1
            else:
                print_fail(f"更新字段失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("更新字段异常", str(e))
            self.test_results["failed"] += 1

        # 验证更新
        print_info(f"\n步骤 5: 验证更新结果")
        try:
            response = requests.get(
                f"{self.base_url}/data/sys_column/{column_id}",
                headers=self.headers
            )
            result = response.json()

            if response.status_code == 200 and result.get("code") == 200:
                data = result["data"]
                if (data.get("DISPLAY_NAME") == update_data["DISPLAY_NAME"] and
                    data.get("COL_LENGTH") == update_data["COL_LENGTH"]):
                    print_success("验证更新成功，数据已更新")
                    self.test_results["passed"] += 1
                else:
                    print_fail("验证更新失败", "更新的数据不匹配")
                    self.test_results["failed"] += 1
            else:
                print_fail(f"验证更新失败 (HTTP {response.status_code})", result.get("message"))
                self.test_results["failed"] += 1
        except Exception as e:
            print_fail("验证更新异常", str(e))
            self.test_results["failed"] += 1

        return column_id

    def cleanup(self, table_id, column_id):
        """清理测试数据"""
        print_header(">>> 4. 清理测试数据")

        # DELETE - 删除字段
        if column_id:
            print_info(f"删除测试字段 (ID: {column_id})")
            try:
                response = requests.delete(
                    f"{self.base_url}/data/sys_column/{column_id}",
                    headers=self.headers
                )
                result = response.json()

                if response.status_code == 200 and result.get("code") == 200:
                    print_success(f"删除字段成功 (HTTP {response.status_code})")
                    self.test_results["passed"] += 1
                else:
                    print_fail(f"删除字段失败 (HTTP {response.status_code})", result.get("message"))
                    self.test_results["failed"] += 1
            except Exception as e:
                print_fail("删除字段异常", str(e))
                self.test_results["failed"] += 1

        # DELETE - 删除表
        if table_id:
            print_info(f"删除测试表 (ID: {table_id})")
            try:
                response = requests.delete(
                    f"{self.base_url}/data/sys_table/{table_id}",
                    headers=self.headers
                )
                result = response.json()

                if response.status_code == 200 and result.get("code") == 200:
                    print_success(f"删除表成功 (HTTP {response.status_code})")
                    self.test_results["passed"] += 1
                else:
                    print_fail(f"删除表失败 (HTTP {response.status_code})", result.get("message"))
                    self.test_results["failed"] += 1
            except Exception as e:
                print_fail("删除表异常", str(e))
                self.test_results["failed"] += 1

    def print_summary(self):
        """打印测试汇总"""
        print_header("测试汇总")
        total = self.test_results["passed"] + self.test_results["failed"]
        pass_rate = (self.test_results["passed"] / total * 100) if total > 0 else 0

        print(f"总测试数: {total}")
        print(f"{Colors.GREEN}通过: {self.test_results['passed']}{Colors.END}")
        print(f"{Colors.RED}失败: {self.test_results['failed']}{Colors.END}")
        print(f"通过率: {pass_rate:.1f}%")
        print(f"{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}\n")

        if self.test_results["failed"] > 0:
            print(f"{Colors.RED}{Colors.BOLD}[FAILED] {self.test_results['failed']} tests failed{Colors.END}")
            sys.exit(1)
        else:
            print(f"{Colors.GREEN}{Colors.BOLD}[SUCCESS] All tests passed!{Colors.END}")

def main():
    """主函数"""
    print(f"\n{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}")
    print(f"{Colors.BLUE}{Colors.BOLD}Sky-Server CRUD API 测试脚本{Colors.END}")
    print(f"{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}")
    print(f"基础URL: {BASE_URL}")
    print(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"{Colors.BLUE}{Colors.BOLD}{'='*80}{Colors.END}\n")

    tester = APITester(BASE_URL)

    # 1. 登录
    if not tester.login():
        print_fail("登录失败，终止测试")
        sys.exit(1)

    # 2. 测试 sys_table CRUD
    table_id = tester.test_sys_table_crud()

    # 3. 测试 sys_column CRUD
    column_id = tester.test_sys_column_crud(table_id)

    # 4. 清理测试数据
    tester.cleanup(table_id, column_id)

    # 5. 打印汇总
    tester.print_summary()

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n\n{Colors.YELLOW}测试被用户中断{Colors.END}")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n{Colors.RED}测试异常: {str(e)}{Colors.END}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
