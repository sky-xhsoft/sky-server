#!/usr/bin/env python3
"""
Sky-Server 云盘服务端功能测试脚本
使用 Python 测试云盘 API 功能
"""

import os
import sys
import time
import json
import hashlib
import random
import string
from pathlib import Path
from typing import Optional, Dict, Any, List, Tuple
from dataclasses import dataclass
from datetime import datetime

# 修复Windows控制台编码问题
if sys.platform == "win32":
    import io
    # 设置标准输出编码为UTF-8
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

# ==================== 配置 ====================

BASE_URL = os.getenv("SKY_SERVER_URL", "http://localhost:9090")
API_BASE = f"{BASE_URL}/api/v1"

# 测试账号配置
TEST_USERNAME = os.getenv("TEST_USERNAME", "admin")
TEST_PASSWORD = os.getenv("TEST_PASSWORD", "admin123")
TEST_COMPANY_ID = int(os.getenv("TEST_COMPANY_ID", "1"))

# 测试数据目录
TEST_DATA_DIR = Path(__file__).parent / "test_data"

# ==================== 数据类 ====================

@dataclass
class TestContext:
    """测试上下文"""
    session: requests.Session
    token: str = ""
    user_id: int = 0
    company_id: int = 1


@dataclass
class CloudItem:
    """云盘项目（文件或文件夹）"""
    id: int
    item_type: str
    name: str
    parent_id: Optional[int]
    owner_id: int
    path: str
    is_active: str
    file_size: Optional[int] = None
    file_type: Optional[str] = None
    storage_type: Optional[str] = None
    access_url: Optional[str] = None


@dataclass
class TestResult:
    """测试结果"""
    name: str
    success: bool
    message: str = ""
    duration: float = 0.0
    data: Any = None


# ==================== 测试工具类 ====================

class SkyServerTester:
    """Sky-Server 云盘测试客户端"""

    def __init__(self, base_url: str = API_BASE):
        self.base_url = base_url
        self.session = self._create_session()
        self.token = ""
        self.context = TestContext(session=self.session)
        self.results: List[TestResult] = []
        self.test_files: List[Path] = []

    def _create_session(self) -> requests.Session:
        """创建带重试机制的HTTP会话"""
        session = requests.Session()
        retry_strategy = Retry(
            total=3,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504]
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        session.mount("http://", adapter)
        session.mount("https://", adapter)
        return session

    def _headers(self) -> Dict[str, str]:
        """获取请求头"""
        headers = {
            "Content-Type": "application/json",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers

    def _check_response(self, response: requests.Response, expected_code: int = 200) -> Dict[str, Any]:
        """检查并解析响应"""
        try:
            data = response.json()
        except json.JSONDecodeError:
            data = {"text": response.text}

        # 接受200和201状态码
        if response.status_code != expected_code and not (expected_code == 200 and response.status_code == 201):
            raise Exception(f"Request failed: {response.status_code} - {data}")

        return data

    def login(self, username: str, password: str, company_id: Optional[int] = None) -> bool:
        """登录获取Token"""
        print(f"正在登录用户: {username}")
        url = f"{self.base_url}/auth/login"
        payload = {
            "username": username,
            "password": password,
            "clientType": "web"
        }
        if company_id and company_id > 0:
            payload["companyId"] = company_id

        try:
            response = self.session.post(url, json=payload, timeout=30)
            data = self._check_response(response)

            if data.get("code") == 0 or data.get("success") or data.get("data"):
                result_data = data.get("data", data)
                self.token = result_data.get("token", result_data.get("accessToken", ""))
                if not self.token:
                    # 尝试从其他字段获取
                    self.token = result_data.get("access_token", "")
                self.context.token = self.token
                print(f"登录成功！Token: {self.token[:20]}...")
                return True
            else:
                print(f"登录失败: {data}")
                return False
        except Exception as e:
            print(f"登录异常: {e}")
            return False

    # ==================== 文件夹管理 ====================

    def create_folder(self, name: str, parent_id: Optional[int] = None, description: str = "") -> Optional[CloudItem]:
        """创建文件夹"""
        url = f"{self.base_url}/cloud/items"
        payload = {
            "itemType": "folder",
            "name": name,
            "description": description
        }
        if parent_id and parent_id > 0:
            payload["parentId"] = parent_id

        try:
            response = self.session.post(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)

            # 解析响应
            item_data = data.get("data", data)
            if isinstance(item_data, dict):
                return CloudItem(
                    id=item_data.get("id", 0),
                    item_type=item_data.get("itemType", "folder"),
                    name=item_data.get("name", name),
                    parent_id=item_data.get("parentId"),
                    owner_id=item_data.get("ownerId", 0),
                    path=item_data.get("path", ""),
                    is_active=item_data.get("isActive", "Y")
                )
            return None
        except Exception as e:
            print(f"创建文件夹失败: {e}")
            return None

    def list_items(self, parent_id: Optional[int] = None) -> Tuple[List[CloudItem], List[CloudItem]]:
        """列出项目"""
        url = f"{self.base_url}/cloud/items"
        params = {}
        if parent_id and parent_id > 0:
            params["parentId"] = str(parent_id)

        try:
            response = self.session.get(
                url,
                params=params,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)

            folders_data = data.get("data", {}).get("folders", [])
            files_data = data.get("data", {}).get("files", [])

            folders = []
            for f in folders_data:
                folders.append(CloudItem(
                    id=f.get("id", 0),
                    item_type="folder",
                    name=f.get("name", ""),
                    parent_id=f.get("parentId"),
                    owner_id=f.get("ownerId", 0),
                    path=f.get("path", ""),
                    is_active=f.get("isActive", "Y")
                ))

            files = []
            for f in files_data:
                files.append(CloudItem(
                    id=f.get("id", 0),
                    item_type="file",
                    name=f.get("name", ""),
                    parent_id=f.get("parentId"),
                    owner_id=f.get("ownerId", 0),
                    path=f.get("path", ""),
                    is_active=f.get("isActive", "Y"),
                    file_size=f.get("fileSize"),
                    file_type=f.get("fileType"),
                    access_url=f.get("accessUrl")
                ))

            return folders, files
        except Exception as e:
            print(f"列出项目失败: {e}")
            return [], []

    def delete_item(self, item_id: int) -> bool:
        """删除项目"""
        url = f"{self.base_url}/cloud/items/{item_id}"
        try:
            response = self.session.delete(
                url,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"删除项目失败: {e}")
            return False

    def rename_item(self, item_id: int, new_name: str) -> bool:
        """重命名项目"""
        url = f"{self.base_url}/cloud/items/{item_id}/rename"
        payload = {"newName": new_name}
        try:
            response = self.session.put(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"重命名项目失败: {e}")
            return False

    def move_item(self, item_id: int, target_parent_id: Optional[int]) -> bool:
        """移动项目"""
        url = f"{self.base_url}/cloud/items/{item_id}/move"
        payload = {"targetParentId": target_parent_id}
        try:
            response = self.session.put(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"移动项目失败: {e}")
            return False

    def upload_file(self, file_path: Path, folder_id: Optional[int] = None) -> Optional[CloudItem]:
        """上传文件"""
        url = f"{self.base_url}/cloud/files/upload"
        try:
            with open(file_path, "rb") as f:
                files = {"file": (file_path.name, f)}
                data = {}
                if folder_id and folder_id > 0:
                    data["folderId"] = str(folder_id)

                # 上传文件时不使用默认headers（避免设置application/json）
                headers = {}
                if self.token:
                    headers["Authorization"] = f"Bearer {self.token}"

                response = self.session.post(
                    url,
                    files=files,
                    data=data,
                    headers=headers,
                    timeout=60
                )
                data = self._check_response(response)

                item_data = data.get("data", data)
                if isinstance(item_data, dict):
                    return CloudItem(
                        id=item_data.get("id", 0),
                        item_type="file",
                        name=item_data.get("name", file_path.name),
                        parent_id=item_data.get("parentId"),
                        owner_id=item_data.get("ownerId", 0),
                        path=item_data.get("path", ""),
                        is_active=item_data.get("isActive", "Y"),
                        file_size=item_data.get("fileSize"),
                        file_type=item_data.get("fileType"),
                        storage_type=item_data.get("storageType"),
                        access_url=item_data.get("accessUrl")
                    )
            return None
        except Exception as e:
            print(f"上传文件失败: {e}")
            return None

    def download_file(self, file_id: int, save_path: Path) -> bool:
        """下载文件"""
        url = f"{self.base_url}/cloud/files/{file_id}/download"
        try:
            # 下载文件时不使用默认headers，避免设置application/json
            headers = {}
            if self.token:
                headers["Authorization"] = f"Bearer {self.token}"

            response = self.session.get(
                url,
                headers=headers,
                timeout=60,
                stream=True
            )
            if response.status_code != 200:
                raise Exception(f"下载文件失败: {response.status_code}")

            with open(save_path, "wb") as f:
                for chunk in response.iter_content(chunk_size=8192):
                    if chunk:
                        f.write(chunk)
            return True
        except Exception as e:
            print(f"下载文件失败: {e}")
            return False

    def get_user_quota(self) -> Optional[Dict[str, Any]]:
        """获取用户配额"""
        url = f"{self.base_url}/cloud/quota"
        try:
            response = self.session.get(
                url,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", data)
        except Exception as e:
            print(f"获取用户配额失败: {e}")
            return None

    def create_share(self, resource_type: str, resource_id: int,
                     share_type: str = "public", password: str = "",
                     expire_days: int = 0, max_downloads: int = 0) -> Optional[Dict[str, Any]]:
        """创建分享"""
        url = f"{self.base_url}/cloud/shares"
        payload = {
            "resourceType": resource_type,
            "resourceId": resource_id,
            "shareType": share_type,
            "password": password,
            "expireDays": expire_days,
            "maxDownloads": max_downloads
        }
        try:
            response = self.session.post(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", data)
        except Exception as e:
            print(f"创建分享失败: {e}")
            return None

    def get_my_shares(self) -> List[Dict[str, Any]]:
        """获取我的分享列表"""
        url = f"{self.base_url}/cloud/shares"
        try:
            response = self.session.get(
                url,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", [])
        except Exception as e:
            print(f"获取分享列表失败: {e}")
            return []

    def cancel_share(self, share_id: int) -> bool:
        """取消分享"""
        url = f"{self.base_url}/cloud/shares/{share_id}"
        try:
            response = self.session.delete(
                url,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"取消分享失败: {e}")
            return False

    # ==================== 存储配置管理 ====================

    def create_storage_config(self, company_id: int, storage_type: str = "local",
                            local_base_path: str = "/tmp/storage",
                            local_base_url: str = "http://localhost:9090/storage") -> Optional[Dict[str, Any]]:
        """创建存储配置"""
        url = f"{self.base_url}/cloud/storage/config"
        payload = {
            "sysCompanyId": company_id,
            "storageType": storage_type,
            "localBasePath": local_base_path,
            "localBaseUrl": local_base_url
        }

        # 根据存储类型添加相应的配置
        if storage_type == "aliyunOSS":
            payload.update({
                "aliyunOSSEndpoint": "oss-cn-hangzhou.aliyuncs.com",
                "aliyunOSSAccessKeyId": "test_access_key",
                "aliyunOSSAccessKeySecret": "test_secret_key",
                "aliyunOSSBucketName": "test-bucket",
                "aliyunOSSCDNDomain": "cdn.example.com"
            })
        elif storage_type == "tencentCOS":
            payload.update({
                "tencentCOSBucketUrl": "https://test-bucket.cos.ap-guangzhou.myqcloud.com",
                "tencentCOSSecretId": "test_secret_id",
                "tencentCOSSecretKey": "test_secret_key",
                "tencentCOSBucketName": "test-bucket",
                "tencentCOSRegion": "ap-guangzhou",
                "tencentCOSCDNDomain": "cdn.example.com"
            })

        try:
            response = self.session.post(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", data)
        except Exception as e:
            print(f"创建存储配置失败: {e}")
            return None

    def update_storage_config(self, config_id: int, company_id: int,
                            storage_type: str = "local",
                            local_base_path: str = "/tmp/storage",
                            local_base_url: str = "http://localhost:9090/storage") -> Optional[Dict[str, Any]]:
        """更新存储配置"""
        url = f"{self.base_url}/cloud/storage/config/{config_id}"
        payload = {
            "id": config_id,
            "sysCompanyId": company_id,
            "storageType": storage_type,
            "localBasePath": local_base_path,
            "localBaseUrl": local_base_url
        }

        if storage_type == "aliyunOSS":
            payload.update({
                "aliyunOSSEndpoint": "oss-cn-hangzhou.aliyuncs.com",
                "aliyunOSSAccessKeyId": "test_access_key_updated",
                "aliyunOSSAccessKeySecret": "test_secret_key_updated",
                "aliyunOSSBucketName": "test-bucket-updated",
                "aliyunOSSCDNDomain": "cdn-updated.example.com"
            })
        elif storage_type == "tencentCOS":
            payload.update({
                "tencentCOSBucketUrl": "https://test-bucket-updated.cos.ap-guangzhou.myqcloud.com",
                "tencentCOSSecretId": "test_secret_id_updated",
                "tencentCOSSecretKey": "test_secret_key_updated",
                "tencentCOSBucketName": "test-bucket-updated",
                "tencentCOSRegion": "ap-guangzhou",
                "tencentCOSCDNDomain": "cdn-updated.example.com"
            })

        try:
            response = self.session.put(
                url,
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", data)
        except Exception as e:
            print(f"更新存储配置失败: {e}")
            return None

    def get_company_storage_config(self, company_id: int) -> Optional[Dict[str, Any]]:
        """获取公司存储配置"""
        url = f"{self.base_url}/cloud/storage/config/company/{company_id}"
        try:
            response = self.session.get(
                url,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", data)
        except Exception as e:
            print(f"获取公司存储配置失败: {e}")
            return None

    def get_all_storage_configs(self) -> List[Dict[str, Any]]:
        """获取所有存储配置"""
        url = f"{self.base_url}/cloud/storage/config"
        try:
            response = self.session.get(
                url,
                headers=self._headers(),
                timeout=30
            )
            data = self._check_response(response)
            return data.get("data", [])
        except Exception as e:
            print(f"获取所有存储配置失败: {e}")
            return []

    def delete_storage_config(self, config_id: int) -> bool:
        """删除存储配置"""
        url = f"{self.base_url}/cloud/storage/config/{config_id}"
        try:
            response = self.session.delete(
                url,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"删除存储配置失败: {e}")
            return False

    def refresh_storage_cache(self, company_id: int) -> bool:
        """刷新存储配置缓存"""
        url = f"{self.base_url}/cloud/storage/config/{company_id}/refresh"
        try:
            response = self.session.post(
                url,
                headers=self._headers(),
                timeout=30
            )
            self._check_response(response)
            return True
        except Exception as e:
            print(f"刷新存储配置缓存失败: {e}")
            return False

    def check_storage_config_api(self) -> bool:
        """检查存储配置API是否可用"""
        url = f"{self.base_url}/cloud/storage/config"
        try:
            response = self.session.get(
                url,
                headers=self._headers(),
                timeout=5
            )
            # 如果返回404或500以外的状态码，认为API可用
            if response.status_code == 404:
                return False
            return True
        except Exception as e:
            print(f"检查存储配置API失败: {e}")
            return False

    # ==================== 测试文件生成 ====================

    def create_test_file(self, size: int, name: Optional[str] = None) -> Path:
        """创建测试文件"""
        TEST_DATA_DIR.mkdir(exist_ok=True)
        if not name:
            name = f"test_{size}_{random_string(8)}.dat"
        file_path = TEST_DATA_DIR / name

        # 生成随机内容
        chunk_size = 1024 * 1024
        with open(file_path, "wb") as f:
            remaining = size
            while remaining > 0:
                write_size = min(chunk_size, remaining)
                f.write(random_bytes(write_size))
                remaining -= write_size

        self.test_files.append(file_path)
        return file_path

    def cleanup(self):
        """清理测试文件"""
        for file_path in self.test_files:
            try:
                if file_path.exists():
                    file_path.unlink()
            except Exception:
                pass
        self.test_files.clear()

        # 清理测试数据目录
        if TEST_DATA_DIR.exists() and not any(TEST_DATA_DIR.iterdir()):
            try:
                TEST_DATA_DIR.rmdir()
            except Exception:
                pass


# ==================== 辅助函数 ====================

def random_string(length: int) -> str:
    """生成随机字符串"""
    chars = string.ascii_letters + string.digits
    return "".join(random.choice(chars) for _ in range(length))


def random_bytes(length: int) -> bytes:
    """生成随机字节"""
    return bytes(random.getrandbits(8) for _ in range(length))


def format_size(size_bytes: int) -> str:
    """格式化文件大小"""
    for unit in ["B", "KB", "MB", "GB", "TB"]:
        if size_bytes < 1024.0:
            return f"{size_bytes:.2f} {unit}"
        size_bytes /= 1024.0
    return f"{size_bytes:.2f} PB"


# ==================== 测试用例 ====================

def run_test(tester: SkyServerTester, test_name: str, test_func):
    """运行单个测试"""
    print(f"\n{'='*60}")
    print(f"运行测试: {test_name}")
    print(f"{'='*60}")

    start_time = time.time()
    success = False
    message = ""
    data = None

    try:
        result = test_func(tester)
        if isinstance(result, tuple) and len(result) >= 2:
            success, message = result[0], result[1]
            if len(result) >= 3:
                data = result[2]
        else:
            success = bool(result)
            message = "测试完成" if success else "测试失败"
    except Exception as e:
        success = False
        message = f"异常: {e}"
        import traceback
        traceback.print_exc()

    duration = time.time() - start_time

    result = TestResult(
        name=test_name,
        success=success,
        message=message,
        duration=duration,
        data=data
    )
    tester.results.append(result)

    status = "[PASS]" if success else "[FAIL]"
    print(f"\n{status} - {test_name}")
    print(f"   消息: {message}")
    print(f"   耗时: {duration:.2f}秒")

    return result


def test_login(tester: SkyServerTester) -> Tuple[bool, str]:
    """测试登录"""
    success = tester.login(TEST_USERNAME, TEST_PASSWORD, TEST_COMPANY_ID)
    return success, "登录成功" if success else "登录失败"


def test_create_folder(tester: SkyServerTester) -> Tuple[bool, str, Any]:
    """测试创建文件夹"""
    folder_name = f"测试文件夹_{random_string(6)}"
    folder = tester.create_folder(folder_name)
    if folder:
        return True, f"创建文件夹成功: {folder_name} (ID: {folder.id})", folder
    return False, "创建文件夹失败", None


def test_list_items(tester: SkyServerTester) -> Tuple[bool, str]:
    """测试列出项目"""
    folders, files = tester.list_items()
    return True, f"列出项目成功: {len(folders)} 个文件夹, {len(files)} 个文件"


def test_delete_items(tester: SkyServerTester, items: List[CloudItem]) -> Tuple[bool, str]:
    """测试删除项目"""
    success_count = 0
    fail_count = 0

    for item in items:
        print(f"删除 {item.item_type}: {item.name} (ID: {item.id})")
        # 如果是文件并且已经在已删除的文件夹中，可能已经被级联删除了，跳过删除
        if item.item_type == "file":
            # 检查文件是否还存在
            try:
                folders, files = tester.list_items()
                file_exists = any(f.id == item.id for f in files)
                if not file_exists:
                    print(f"文件已不存在（可能已被级联删除）: {item.name}")
                    success_count += 1
                    continue
            except:
                pass

        # 尝试删除
        if tester.delete_item(item.id):
            success_count += 1
        else:
            fail_count += 1

    if fail_count == 0:
        return True, f"删除成功: {success_count} 个项目"
    return False, f"删除: {success_count} 成功, {fail_count} 失败"


def test_upload_file(tester: SkyServerTester) -> Tuple[bool, str, Any]:
    """测试上传文件"""
    file_size = 1024 * 1024  # 1MB
    test_file = tester.create_test_file(file_size)
    file = tester.upload_file(test_file)
    if file:
        return True, f"上传文件成功: {file.name} (ID: {file.id})", file
    return False, "上传文件失败", None


def test_download_file(tester: SkyServerTester, file: CloudItem) -> Tuple[bool, str]:
    """测试下载文件"""
    save_path = TEST_DATA_DIR / f"download_{file.name}"
    if tester.download_file(file.id, save_path):
        # 验证下载的文件大小
        downloaded_size = save_path.stat().st_size
        if downloaded_size == file.file_size:
            save_path.unlink()
            return True, f"下载文件成功: {file.name}"
        else:
            save_path.unlink()
            return False, f"文件大小不匹配: 预期 {file.file_size}, 实际 {downloaded_size}"
    return False, "下载文件失败"


def test_rename_item(tester: SkyServerTester, item: CloudItem) -> Tuple[bool, str]:
    """测试重命名项目"""
    new_name = f"重命名_{random_string(6)}.txt" if item.item_type == "file" else f"重命名文件夹_{random_string(6)}"
    if tester.rename_item(item.id, new_name):
        return True, f"重命名成功: {item.name} -> {new_name}"
    return False, f"重命名失败: {item.name}"


def test_move_item(tester: SkyServerTester, source_item: CloudItem, target_folder: Optional[CloudItem]) -> Tuple[bool, str]:
    """测试移动项目"""
    target_id = target_folder.id if target_folder and target_folder.id != 0 else None
    if tester.move_item(source_item.id, target_id):
        target_name = target_folder.name if target_folder else "根目录"
        return True, f"移动成功: {source_item.name} -> {target_name}"
    return False, f"移动失败: {source_item.name}"


def test_get_user_quota(tester: SkyServerTester) -> Tuple[bool, str, Any]:
    """测试获取用户配额"""
    quota = tester.get_user_quota()
    if quota:
        used_gb = quota.get("usedSpace", 0) / (1024 * 1024 * 1024)
        total_gb = quota.get("totalQuota", 0) / (1024 * 1024 * 1024)
        return True, f"配额查询成功: {used_gb:.2f} GB / {total_gb:.2f} GB", quota
    return False, "配额查询失败", None


def test_create_and_cancel_share(tester: SkyServerTester, item: CloudItem) -> Tuple[bool, str, Any]:
    """测试创建和取消分享"""
    share = tester.create_share(
        resource_type=item.item_type,
        resource_id=item.id,
        share_type="password",
        password="test123",
        expire_days=7,
        max_downloads=10
    )
    if share:
        share_code = share.get("shareCode", "")
        if tester.cancel_share(share.get("id", 0)):
            return True, f"分享创建和取消成功: 分享码 {share_code}", share
        return False, f"创建分享成功但取消失败: {share_code}", share
    return False, "创建分享失败", None


def test_create_local_storage_config(tester: SkyServerTester, company_id: int = 1) -> Tuple[bool, str, Any]:
    """测试创建本地存储配置"""
    config = tester.create_storage_config(
        company_id=company_id,
        storage_type="local",
        local_base_path=f"/test/storage_{random_string(6)}",
        local_base_url=f"http://localhost:9090/storage_{random_string(6)}"
    )
    if config:
        return True, f"创建本地存储配置成功: ID {config.get('id')}", config
    return False, "创建本地存储配置失败", None


def test_create_aliyun_oss_config(tester: SkyServerTester, company_id: int = 1) -> Tuple[bool, str, Any]:
    """测试创建阿里云OSS存储配置"""
    config = tester.create_storage_config(
        company_id=company_id,
        storage_type="aliyunOSS"
    )
    if config:
        return True, f"创建阿里云OSS存储配置成功: ID {config.get('id')}", config
    return False, "创建阿里云OSS存储配置失败", None


def test_create_tencent_cos_config(tester: SkyServerTester, company_id: int = 1) -> Tuple[bool, str, Any]:
    """测试创建腾讯云COS存储配置"""
    config = tester.create_storage_config(
        company_id=company_id,
        storage_type="tencentCOS"
    )
    if config:
        return True, f"创建腾讯云COS存储配置成功: ID {config.get('id')}", config
    return False, "创建腾讯云COS存储配置失败", None


def test_get_company_storage_config(tester: SkyServerTester, company_id: int = 1) -> Tuple[bool, str, Any]:
    """测试获取公司存储配置"""
    config = tester.get_company_storage_config(company_id)
    if config:
        return True, f"获取公司存储配置成功: {config.get('storageType')}", config
    return False, "获取公司存储配置失败", None


def test_get_all_storage_configs(tester: SkyServerTester) -> Tuple[bool, str, Any]:
    """测试获取所有存储配置"""
    configs = tester.get_all_storage_configs()
    if configs:
        count = len(configs)
        types = ", ".join([c.get('storageType', 'unknown') for c in configs])
        return True, f"获取存储配置列表成功: {count} 条记录 ({types})", configs
    return False, "获取存储配置列表失败", []


def test_update_storage_config(tester: SkyServerTester, config: Dict[str, Any]) -> Tuple[bool, str, Any]:
    """测试更新存储配置"""
    updated_config = tester.update_storage_config(
        config_id=config.get("id"),
        company_id=config.get("sysCompanyId"),
        storage_type=config.get("storageType"),
        local_base_path=f"/test/updated_{random_string(6)}",
        local_base_url=f"http://localhost:9090/updated_{random_string(6)}"
    )
    if updated_config:
        return True, f"更新存储配置成功: ID {updated_config.get('id')}", updated_config
    return False, "更新存储配置失败", None


def test_refresh_storage_cache(tester: SkyServerTester, company_id: int = 1) -> Tuple[bool, str]:
    """测试刷新存储配置缓存"""
    if tester.refresh_storage_cache(company_id):
        return True, "刷新存储配置缓存成功"
    return False, "刷新存储配置缓存失败"


def test_delete_storage_config(tester: SkyServerTester, config: Dict[str, Any]) -> Tuple[bool, str]:
    """测试删除存储配置"""
    if tester.delete_storage_config(config.get("id")):
        return True, f"删除存储配置成功: ID {config.get('id')}"
    return False, f"删除存储配置失败: ID {config.get('id')}"


# ==================== 主函数 ====================

def main():
    """主函数 - 运行所有测试"""
    print("""
============================================================
             Sky-Server 云盘功能测试
============================================================
    """)

    print(f"服务器地址: {BASE_URL}")
    print(f"测试用户: {TEST_USERNAME}")
    print("")

    # 创建测试客户端
    tester = SkyServerTester()

    created_items: List[CloudItem] = []
    test_file: Optional[CloudItem] = None
    test_folder: Optional[CloudItem] = None
    test_folder2: Optional[CloudItem] = None

    try:
        # ==================== 测试流程 ====================

        # 1. 登录测试
        run_test(tester, "1. 用户登录", test_login)

        # 检查是否登录成功
        if not tester.token:
            print("\n登录失败，无法继续测试")
            return

        # ==================== 云盘存储配置测试 ====================
        created_configs: List[Dict[str, Any]] = []

        # 检查云盘存储配置API是否可用
        storage_config_available = tester.check_storage_config_api()

        if storage_config_available:
            print("\n" + "="*60)
            print("云盘存储配置API可用，开始存储配置测试")
            print("="*60)

            # 12. 测试获取所有存储配置
            list_result = run_test(tester, "12. 获取所有存储配置",
                                 lambda t: test_get_all_storage_configs(t))

            # 13. 测试获取公司存储配置
            config_result = run_test(tester, "13. 获取公司存储配置",
                                   lambda t: test_get_company_storage_config(t, TEST_COMPANY_ID))

            # 14. 测试创建本地存储配置
            local_config_result = run_test(tester, "14. 创建本地存储配置",
                                         lambda t: test_create_local_storage_config(t, TEST_COMPANY_ID))
            if local_config_result.success and local_config_result.data:
                created_configs.append(local_config_result.data)

            # 15. 测试创建阿里云OSS存储配置（可选，需要真实的OSS配置）
            oss_config_result = run_test(tester, "15. 创建阿里云OSS存储配置",
                                        lambda t: test_create_aliyun_oss_config(t, TEST_COMPANY_ID))
            if oss_config_result.success and oss_config_result.data:
                created_configs.append(oss_config_result.data)

            # 16. 测试创建腾讯云COS存储配置（可选，需要真实的COS配置）
            cos_config_result = run_test(tester, "16. 创建腾讯云COS存储配置",
                                        lambda t: test_create_tencent_cos_config(t, TEST_COMPANY_ID))
            if cos_config_result.success and cos_config_result.data:
                created_configs.append(cos_config_result.data)

            # 17. 测试更新存储配置
            if created_configs:
                update_result = run_test(tester, "17. 更新存储配置",
                                       lambda t: test_update_storage_config(t, created_configs[0]))
                if update_result.success and update_result.data:
                    # 更新配置列表
                    created_configs[0] = update_result.data

            # 18. 测试刷新存储配置缓存
            run_test(tester, "18. 刷新存储配置缓存",
                    lambda t: test_refresh_storage_cache(t, TEST_COMPANY_ID))
        else:
            print("\n" + "="*60)
            print("云盘存储配置API不可用，跳过存储配置测试")
            print("="*60)

        # 根据存储配置API可用性，确定传统云盘操作的起始编号
        base_test_number = 12 if storage_config_available else 2

        # ==================== 传统云盘操作测试 ====================

        # 创建文件夹
        result = run_test(tester, f"{base_test_number}. 创建文件夹", test_create_folder)
        if result.success and result.data:
            test_folder = result.data
            created_items.append(result.data)

        # 列出项目
        run_test(tester, f"{base_test_number + 1}. 列出项目", test_list_items)

        # 获取用户配额
        result_quota = run_test(tester, f"{base_test_number + 2}. 获取用户配额", test_get_user_quota)

        # 上传文件
        result_upload = run_test(tester, f"{base_test_number + 3}. 上传文件", test_upload_file)
        if result_upload.success and result_upload.data:
            test_file = result_upload.data
            created_items.append(result_upload.data)

        # 下载文件
        if test_file:
            run_test(tester, f"{base_test_number + 4}. 下载文件", lambda t: test_download_file(t, test_file))

        # 重命名文件
        if test_file:
            run_test(tester, f"{base_test_number + 5}. 重命名文件", lambda t: test_rename_item(t, test_file))

        # 重命名文件夹
        if test_folder:
            run_test(tester, f"{base_test_number + 6}. 重命名文件夹", lambda t: test_rename_item(t, test_folder))

        # 创建第二个文件夹用于移动测试
        if test_folder:
            folder_name2 = f"测试文件夹2_{random_string(6)}"
            folder2 = tester.create_folder(folder_name2)
            if folder2:
                test_folder2 = folder2
                created_items.append(folder2)

        # 移动文件
        if test_file and test_folder2:
            run_test(tester, f"{base_test_number + 8}. 移动文件", lambda t: test_move_item(t, test_file, test_folder2))

        # 创建分享并取消
        if test_file:
            run_test(tester, f"{base_test_number + 9}. 创建和取消分享", lambda t: test_create_and_cancel_share(t, test_file))

        # 将文件移回根目录（准备删除）
        if test_file and test_file.parent_id:
            run_test(tester, f"{base_test_number + 10}. 将文件移回根目录", lambda t: test_move_item(t, test_file, CloudItem(id=0, item_type="folder", name="根目录", parent_id=None, owner_id=0, path="/", is_active="Y")))

        # 删除测试项目
        if created_items:
            run_test(tester, f"{base_test_number + 11}. 删除测试项目",
                     lambda t: test_delete_items(t, list(reversed(created_items))))

        # ==================== 云盘存储配置清理 ====================
        if storage_config_available and created_configs:
            for config in reversed(created_configs):
                run_test(tester, f"删除存储配置: ID {config.get('id')}",
                         lambda t, cfg=config: test_delete_storage_config(t, cfg))

        # ==================== 测试总结 ====================
        print("\n" + "="*70)
        print("测试总结")
        print("="*70)

        total_tests = len(tester.results)
        passed_tests = sum(1 for r in tester.results if r.success)
        failed_tests = total_tests - passed_tests
        total_duration = sum(r.duration for r in tester.results)

        print(f"\n总测试数: {total_tests}")
        print(f"[PASS] 通过: {passed_tests}")
        print(f"[FAIL] 失败: {failed_tests}")
        print(f"总耗时: {total_duration:.2f}秒")

        if failed_tests > 0:
            print("\n失败的测试:")
            for r in tester.results:
                if not r.success:
                    print(f"  - {r.name}: {r.message}")

        print("\n" + "="*70)
        if failed_tests == 0:
            print("所有测试通过!")
        else:
            print(f"有 {failed_tests} 个测试失败")
        print("="*70)

    except KeyboardInterrupt:
        print("\n\n测试被用户中断")
    except Exception as e:
        print(f"\n\n测试过程发生异常: {e}")
        import traceback
        traceback.print_exc()
    finally:
        # 清理
        print("\n清理测试文件...")
        tester.cleanup()
        print("清理完成")


def print_usage():
    """打印使用说明"""
    print("""
Sky-Server 云盘功能测试脚本

使用方法:
    python cloud_test.py [选项]

环境变量:
    SKY_SERVER_URL    服务器地址 (默认: http://localhost:9090)
    TEST_USERNAME     测试用户名 (默认: admin)
    TEST_PASSWORD     测试密码 (默认: admin123)
    TEST_COMPANY_ID   测试公司ID (默认: 1)

示例:
    # 使用默认配置
    python cloud_test.py

    # 自定义服务器地址
    SKY_SERVER_URL=http://192.168.1.100:9090 python cloud_test.py

    # 自定义账号
    TEST_USERNAME=testuser TEST_PASSWORD=testpass python cloud_test.py
    """)


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] in ("-h", "--help", "help"):
        print_usage()
    else:
        main()
