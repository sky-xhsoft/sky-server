package logger

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"go.uber.org/zap"
)

// 直接使用 zap.L() 的日志应写入项目日志，且调用位置指向调用方而不是 zap 内部
func TestInit_ReplacesGlobalZapLogger(t *testing.T) {
	// 输出到标准输出时，Init 会绑定当时的 os.Stdout，先换成管道以便读取
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	if _, err := Init(&config.LogConfig{Level: "info", Format: "console", Output: "stdout"}); err != nil {
		t.Fatal(err)
	}
	zap.L().Error("全局日志器测试")
	_ = w.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "全局日志器测试") {
		t.Fatalf("zap.L() 的日志未写入项目日志: %q", out)
	}
	if !strings.Contains(out, "global_test.go") {
		t.Errorf("调用位置应为 global_test.go，实际: %q", out)
	}
}
