#!/usr/bin/env python3
import re
import sys

def replace_in_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # 替换通过 stream_name 获取直播间信息的代码块
        # 找到所有匹配的模式并替换
        patterns = [
            # 1. 主要的查询模式
            (r'Where\("STREAM_NAME = \? AND IS_ACTIVE = \?", event\.StreamName, "Y"\)', 
             r'Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y")'),
            # 2. 日志消息模式
            (r'zap\.String\("streamName", event\.StreamName\)', 
             r'zap.String("streamId", event.StreamID)'),
            (r'zap\.String\("streamName", streamName\)', 
             r'zap.String("streamId", event.StreamID)'),
            # 3. 注释模式
            (r'// 通过流名称获取直播间信息', 
             r'// 通过 stream_id 获取直播间信息'),
        ]
        
        for old, new in patterns:
            content = re.sub(old, new, content)
        
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        print(f"Successfully processed: {file_path}")
    except Exception as e:
        print(f"Error processing {file_path}: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 replace.py <file1> <file2> ...")
        sys.exit(1)
    
    for file_path in sys.argv[1:]:
        replace_in_file(file_path)
