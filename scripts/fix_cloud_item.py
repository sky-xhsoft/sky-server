#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
修复云盘录制文件脚本

功能：批量处理 live_callback_event 表中的 recording_file 类型事件，
      自动生成对应的 cloud_item 文件夹和文件记录。

使用方法：
    python fix_cloud_item.py

数据库配置：
    主机: 127.0.0.1
    端口: 3306
    用户: root
    密码: ***REMOVED***
    库名: skyserver

依赖安装：
    pip install pymysql
"""

import json
import pymysql
from typing import Optional, Dict, Any

# ==================== 数据库配置 ====================
DB_CONFIG = {
    'host': '127.0.0.1',
    'port': 3306,
    'user': 'root',
    'password': '***REMOVED***',
    'database': 'skyserver',
    'charset': 'utf8mb4',
    'cursorclass': pymysql.cursors.DictCursor
}

# 默认直播间父文件夹ID（与配置一致）
DEFAULT_ROOM_PARENT_ID = 391


# ==================== 数据库操作 ====================

def get_db_connection():
    """获取数据库连接"""
    return pymysql.connect(**DB_CONFIG)


def query_recording_file_events(connection) -> list:
    """查询所有录制文件事件"""
    sql = """
        SELECT ID, EVENT_TYPE, ROOM_NAME, STREAM_ID, EVENT_TIME, EVENT_DATA
        FROM live_callback_event
        WHERE EVENT_TYPE = 'recording_file'
        ORDER BY ID ASC
    """
    with connection.cursor() as cursor:
        cursor.execute(sql)
        return cursor.fetchall()


def check_file_exists(connection, access_url: str) -> bool:
    """检查文件是否已存在"""
    sql = """
        SELECT COUNT(*) as cnt FROM cloud_item
        WHERE ITEM_TYPE = 'file' AND ACCESS_URL = %s AND IS_ACTIVE = 'Y'
    """
    with connection.cursor() as cursor:
        cursor.execute(sql, (access_url,))
        result = cursor.fetchone()
        return result['cnt'] > 0


def get_live_room(connection, stream_id: str) -> Optional[Dict[str, Any]]:
    """获取直播间信息"""
    # 先尝试转为数字查询
    try:
        room_id = int(stream_id)
        sql = """
            SELECT ID, ROOM_NAME, CREATE_BY
            FROM live_room
            WHERE ID = %s AND IS_ACTIVE = 'Y'
        """
        with connection.cursor() as cursor:
            cursor.execute(sql, (room_id,))
            result = cursor.fetchone()
            if result:
                return result
    except (ValueError, TypeError):
        pass

    # 数字查询失败，尝试字符串查询（兼容旧数据）
    sql = """
        SELECT ID, ROOM_NAME, CREATE_BY
        FROM live_room
        WHERE ID = %s AND IS_ACTIVE = 'Y'
    """
    with connection.cursor() as cursor:
        cursor.execute(sql, (stream_id,))
        return cursor.fetchone()


def get_user_id_by_username(connection, username: str) -> Optional[int]:
    """根据用户名获取用户ID"""
    sql = """
        SELECT ID FROM sys_user WHERE USERNAME = %s AND IS_ACTIVE = 'Y'
    """
    with connection.cursor() as cursor:
        cursor.execute(sql, (username,))
        result = cursor.fetchone()
        return result['ID'] if result else None


def get_or_create_room_folder(connection, room_name: str, user_id: int) -> int:
    """获取或创建直播间文件夹"""
    # 先查询是否已存在
    check_sql = """
        SELECT ID FROM cloud_item
        WHERE ITEM_TYPE = 'folder' AND NAME = %s AND OWNER_ID = %s AND PARENT_ID = %s AND IS_ACTIVE = 'Y'
    """
    with connection.cursor() as cursor:
        cursor.execute(check_sql, (room_name, user_id, DEFAULT_ROOM_PARENT_ID))
        result = cursor.fetchone()
        if result:
            return result['ID']

    # 不存在则创建
    path = "/" + room_name
    insert_sql = """
        INSERT INTO cloud_item (
            ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
            FILE_COUNT, TOTAL_SIZE, CREATE_BY, UPDATE_BY, IS_ACTIVE
        ) VALUES (%s, %s, %s, %s, %s, 0, 0, %s, %s, 'Y')
    """
    with connection.cursor() as cursor:
        cursor.execute(insert_sql, ("folder", room_name, DEFAULT_ROOM_PARENT_ID, path, user_id, "admin", "admin"))
        connection.commit()
        return cursor.lastrowid


def get_or_create_recording_folder(connection, parent_id: int, room_name: str, user_id: int) -> int:
    """获取或创建直播录制文件夹"""
    folder_name = "直播录制"

    # 先查询是否已存在
    check_sql = """
        SELECT ID FROM cloud_item
        WHERE ITEM_TYPE = 'folder' AND NAME = %s AND OWNER_ID = %s AND PARENT_ID = %s AND IS_ACTIVE = 'Y'
    """
    with connection.cursor() as cursor:
        cursor.execute(check_sql, (folder_name, user_id, parent_id))
        result = cursor.fetchone()
        if result:
            return result['ID']

    # 不存在则创建
    path = "/" + room_name + "/" + folder_name
    insert_sql = """
        INSERT INTO cloud_item (
            ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
            FILE_COUNT, TOTAL_SIZE, CREATE_BY, UPDATE_BY, IS_ACTIVE
        ) VALUES (%s, %s, %s, %s, %s, 0, 0, %s, %s, 'Y')
    """
    with connection.cursor() as cursor:
        cursor.execute(insert_sql, ("folder", folder_name, parent_id, path, user_id, "admin", "admin"))
        connection.commit()
        return cursor.lastrowid


def create_recording_file(connection, parent_id: int, room_name: str, user_id: int,
                          event_time: int, event_data: Dict[str, Any]) -> int:
    """创建录制文件记录"""
    file_format = event_data.get('file_format', 'mp4')
    file_name = f"录制_{event_time}.{file_format}"
    file_path = "/" + room_name + "/直播录制/" + file_name
    file_type = "video/" + file_format
    file_ext = "." + file_format
    video_url = event_data.get('video_url', '')
    file_size = event_data.get('file_size', 0)

    insert_sql = """
        INSERT INTO cloud_item (
            ITEM_TYPE, NAME, PARENT_ID, PATH, OWNER_ID,
            STORAGE_TYPE, STORAGE_PATH, FILE_SIZE, FILE_TYPE, FILE_EXT,
            ACCESS_URL, DOWNLOAD_COUNT, CREATE_BY, UPDATE_BY, IS_ACTIVE
        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, 0, %s, %s, 'Y')
    """
    with connection.cursor() as cursor:
        cursor.execute(insert_sql, (
            "file", file_name, parent_id, file_path, user_id,
            "oss", video_url, file_size, file_type, file_ext,
            video_url, "admin", "admin"
        ))
        connection.commit()
        return cursor.lastrowid


# ==================== 主程序 ====================

def main():
    print("=" * 80)
    print("云盘录制文件修复脚本")
    print("=" * 80)
    print()

    # 连接数据库
    try:
        connection = get_db_connection()
        print("✅ 数据库连接成功！")
        print()
    except Exception as e:
        print(f"❌ 数据库连接失败: {e}")
        return

    try:
        # 查询所有录制事件
        events = query_recording_file_events(connection)
        print(f"找到 {len(events)} 条录制文件事件")
        print()

        if len(events) == 0:
            print("没有需要处理的事件")
            return

        success_count = 0
        skip_count = 0
        fail_count = 0

        # 处理每个事件
        for event in events:
            event_id = event['ID']
            stream_id = event['STREAM_ID']
            room_name = event['ROOM_NAME'] or ''

            print(f"处理事件 ID={event_id}, StreamID={stream_id}, RoomName={room_name}")

            try:
                # 解析事件数据
                event_data = json.loads(event['EVENT_DATA'])
                video_url = event_data.get('video_url', '')

                if not video_url:
                    print("  ⚠️  视频URL为空，跳过")
                    skip_count += 1
                    continue

                # 检查文件是否已存在
                if check_file_exists(connection, video_url):
                    print("  ⏭️  文件已存在，跳过")
                    skip_count += 1
                    continue

                # 获取直播间信息
                room = get_live_room(connection, stream_id)
                if not room:
                    print(f"  ❌ 未找到直播间: {stream_id}")
                    fail_count += 1
                    continue
                print(f"  ✅ 找到直播间: ID={room['ID']}, Name={room['ROOM_NAME']}, CreateBy={room['CREATE_BY']}")

                # 获取用户ID
                user_id = get_user_id_by_username(connection, room['CREATE_BY'])
                if not user_id:
                    print(f"  ❌ 未找到用户: {room['CREATE_BY']}")
                    fail_count += 1
                    continue
                print(f"  ✅ 用户ID: {user_id}")

                # 获取或创建直播间文件夹
                room_folder_id = get_or_create_room_folder(connection, room['ROOM_NAME'], user_id)
                print(f"  ✅ 直播间文件夹ID: {room_folder_id}")

                # 获取或创建直播录制文件夹
                recording_folder_id = get_or_create_recording_folder(connection, room_folder_id, room['ROOM_NAME'], user_id)
                print(f"  ✅ 直播录制文件夹ID: {recording_folder_id}")

                # 确定事件时间
                event_time = event_data.get('event_time', 0)
                if event_time == 0:
                    end_time = event_data.get('end_time', 0)
                    if end_time > 0:
                        event_time = end_time
                    else:
                        import time
                        event_time = int(time.time())

                # 创建文件记录
                file_id = create_recording_file(connection, recording_folder_id, room['ROOM_NAME'], user_id, event_time, event_data)
                print(f"  ✅ 文件创建成功! ID={file_id}")

                success_count += 1

            except Exception as e:
                print(f"  ❌ 处理失败: {e}")
                fail_count += 1

            print()

        # 输出统计
        print("=" * 80)
        print(f"处理完成！成功: {success_count}, 跳过: {skip_count}, 失败: {fail_count}")
        print()

    except Exception as e:
        print(f"❌ 程序执行失败: {e}")
        import traceback
        traceback.print_exc()
    finally:
        connection.close()


if __name__ == "__main__":
    main()
