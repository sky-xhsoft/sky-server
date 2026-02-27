#!/usr/bin/env python3
import re
import os

def add_room_name_field(file_path):
    print(f"Processing file: {file_path}")
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 找到所有事件处理函数
    function_pattern = r'func.*?Handle.*?{.*?}'
    functions = re.findall(function_pattern, content, re.DOTALL)
    
    # 为每个事件处理函数添加 RoomName 字段
    for func in functions:
        # 检查函数是否已经包含了 RoomName 字段
        if 'RoomName' in func:
            print(f"Function already has RoomName field: {func.split('(')[0]}")
            continue
        
        # 找到创建 LiveCallbackEvent 的代码块
        match = re.search(r'callbackEvent := &entity\.LiveCallbackEvent\{(.*?)\}', func, re.DOTALL)
        if match:
            event_str = match.group(1)
            print(f"Found LiveCallbackEvent creation in: {func.split('(')[0]}")
            
            # 在 LiveCallbackEvent 结构体中添加 RoomName 字段
            if 'RoomName' not in event_str:
                # 在 IsActive 之前添加 RoomName 字段
                new_event_str = re.sub(r'IsActive:', r'RoomName:     room.RoomName,\n\tIsActive:', event_str)
                content = content.replace(event_str, new_event_str)
            
            # 检查是否已经有通过 stream_id 获取直播间信息的代码
            if '通过 stream_id 获取直播间信息' not in func:
                # 找到保存事件的代码位置
                save_match = re.search(r'// 保存事件到数据库|if err := h\.db\.Create.*?Error', func)
                if save_match:
                    # 在保存事件之前添加获取直播间信息的代码
                    room_code = '''\t// 通过 stream_id 获取直播间信息
\tvar room entity.LiveRoom
\tif err := h.db.Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y").First(&room).Error; err == nil {
\t\tlogger.Info("找到对应的直播间",
\t\t\tzap.String("streamId", event.StreamID),
\t\t\tzap.String("roomName", room.RoomName))
\t} else {
\t\tlogger.Warn("未找到对应的直播间",
\t\t\tzap.String("streamId", event.StreamID))
\t}

'''
                    content = content.replace(save_match.group(0), room_code + save_match.group(0))
    
    # 写入修改后的内容
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"File processed successfully: {file_path}")

if __name__ == "__main__":
    files = ["api/handler/live_callback_handler.go", "api/handler/live_callback_handler_ext.go"]
    for file_path in files:
        if os.path.exists(file_path):
            add_room_name_field(file_path)
        else:
            print(f"File not found: {file_path}")
