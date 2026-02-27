package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func replaceInFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 替换查询条件
		line = regexp.MustCompile(`Where\("STREAM_NAME = \? AND IS_ACTIVE = \?", event\.StreamName, "Y"\)`).ReplaceAllString(line, `Where("ID = ? AND IS_ACTIVE = ?", event.StreamID, "Y")`)
		// 替换日志消息
		line = regexp.MustCompile(`zap\.String\("streamName", event\.StreamName\)`).ReplaceAllString(line, `zap.String("streamId", event.StreamID)`)
		line = regexp.MustCompile(`zap\.String\("streamName", streamName\)`).ReplaceAllString(line, `zap.String("streamId", event.StreamID)`)
		// 替换注释
		line = regexp.MustCompile(`// 通过流名称获取直播间信息`).ReplaceAllString(line, `// 通过 stream_id 获取直播间信息`)

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	file, err = os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return w.Flush()
}

func main() {
	files := []string{"api/handler/live_callback_handler.go", "api/handler/live_callback_handler_ext.go"}
	for _, filePath := range files {
		fmt.Printf("正在处理文件: %s\n", filePath)
		err := replaceInFile(filePath)
		if err != nil {
			fmt.Printf("处理文件 %s 失败: %v\n", filePath, err)
		} else {
			fmt.Printf("文件 %s 处理成功\n", filePath)
		}
	}
}
