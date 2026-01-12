package hotload

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// Compiler 插件编译器
// 负责将 Go 源码编译成 .so 插件文件
type Compiler struct {
	runtimeDir  string // 插件源码目录
	compiledDir string // 编译输出目录
	modulePath  string // Go 模块路径
}

// CompileResult 编译结果
type CompileResult struct {
	Success      bool      // 是否成功
	OutputPath   string    // 输出文件路径
	SourceHash   string    // 源码哈希值
	CompileTime  time.Time // 编译时间
	Error        error     // 错误信息
	CompileDuration time.Duration // 编译耗时
}

// NewCompiler 创建编译器
func NewCompiler(runtimeDir, compiledDir, modulePath string) *Compiler {
	return &Compiler{
		runtimeDir:  runtimeDir,
		compiledDir: compiledDir,
		modulePath:  modulePath,
	}
}

// Compile 编译插件
// pluginPath: 插件源码相对路径（相对于 runtimeDir）
func (c *Compiler) Compile(pluginPath string) *CompileResult {
	startTime := time.Now()
	result := &CompileResult{
		CompileTime: startTime,
	}

	// 1. 获取源码目录
	sourceDir := filepath.Join(c.runtimeDir, pluginPath)
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		result.Error = fmt.Errorf("获取源码路径失败: %w", err)
		return result
	}

	// 2. 检查源码目录是否存在
	if _, err := os.Stat(absSourceDir); os.IsNotExist(err) {
		result.Error = fmt.Errorf("源码目录不存在: %s", absSourceDir)
		return result
	}

	// 3. 计算源码哈希值
	hash, err := c.calculateSourceHash(absSourceDir)
	if err != nil {
		result.Error = fmt.Errorf("计算源码哈希失败: %w", err)
		return result
	}
	result.SourceHash = hash

	logger.Info("开始编译插件",
		zap.String("plugin", pluginPath),
		zap.String("sourceDir", absSourceDir),
		zap.String("hash", hash))

	// 4. 确保编译输出目录存在
	absCompiledDir, err := filepath.Abs(c.compiledDir)
	if err != nil {
		result.Error = fmt.Errorf("获取编译目录路径失败: %w", err)
		return result
	}
	if err := os.MkdirAll(absCompiledDir, 0755); err != nil {
		result.Error = fmt.Errorf("创建编译目录失败: %w", err)
		return result
	}

	// 5. 生成输出文件名
	pluginName := filepath.Base(pluginPath)
	outputFileName := fmt.Sprintf("%s.so", pluginName)
	outputPath := filepath.Join(absCompiledDir, outputFileName)
	result.OutputPath = outputPath

	// 6. 执行编译
	// 使用 go build -buildmode=plugin 编译成 .so 文件
	cmd := exec.Command("go", "build",
		"-buildmode=plugin",
		"-o", outputPath,
		absSourceDir)

	// 设置环境变量
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1", // 必须启用 CGO
	)

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Errorf("编译失败: %w\n%s", err, string(output))
		logger.Error("插件编译失败",
			zap.String("plugin", pluginPath),
			zap.Error(err),
			zap.String("output", string(output)))
		return result
	}

	// 7. 验证输出文件是否存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("编译输出文件不存在: %s", outputPath)
		return result
	}

	// 8. 编译成功
	result.Success = true
	result.CompileDuration = time.Since(startTime)

	logger.Info("插件编译成功",
		zap.String("plugin", pluginPath),
		zap.String("output", outputPath),
		zap.Duration("duration", result.CompileDuration))

	return result
}

// NeedsRecompile 检查是否需要重新编译
// 比较当前源码哈希与上次编译的哈希
func (c *Compiler) NeedsRecompile(pluginPath string, lastHash string) (bool, error) {
	sourceDir := filepath.Join(c.runtimeDir, pluginPath)
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return false, fmt.Errorf("获取源码路径失败: %w", err)
	}

	currentHash, err := c.calculateSourceHash(absSourceDir)
	if err != nil {
		return false, fmt.Errorf("计算源码哈希失败: %w", err)
	}

	return currentHash != lastHash, nil
}

// calculateSourceHash 计算源码目录的哈希值
// 遍历所有 .go 文件，计算内容哈希
func (c *Compiler) calculateSourceHash(dir string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理 .go 文件
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 读取文件内容
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 写入文件路径（相对路径）
		relPath, _ := filepath.Rel(dir, path)
		hasher.Write([]byte(relPath))

		// 写入文件内容
		if _, err := io.Copy(hasher, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// Clean 清理编译输出
// 删除编译后的 .so 文件
func (c *Compiler) Clean(pluginName string) error {
	outputFileName := fmt.Sprintf("%s.so", pluginName)
	outputPath := filepath.Join(c.compiledDir, outputFileName)

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}

	if err := os.Remove(outputPath); err != nil {
		return fmt.Errorf("清理编译文件失败: %w", err)
	}

	logger.Info("清理编译文件",
		zap.String("plugin", pluginName),
		zap.String("file", outputPath))

	return nil
}
