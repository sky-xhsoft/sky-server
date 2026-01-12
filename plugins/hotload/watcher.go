package hotload

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"go.uber.org/zap"
)

// FileChangeEvent 文件变化事件
type FileChangeEvent struct {
	PluginPath string    // 插件路径（相对 runtime 目录）
	EventType  string    // 事件类型: create, write, remove
	Timestamp  time.Time // 事件时间
}

// ChangeHandler 文件变化处理函数
type ChangeHandler func(event FileChangeEvent)

// Watcher 文件监听器
// 监听插件源码目录的文件变化
type Watcher struct {
	watcher       *fsnotify.Watcher
	runtimeDir    string
	debounceTime  time.Duration              // 防抖动时间
	changeHandler ChangeHandler              // 变化处理函数
	debounceMap   map[string]*time.Timer     // 防抖动定时器
	debounceMu    sync.Mutex                 // 防抖动锁
	stopCh        chan struct{}              // 停止信号
	running       bool                       // 是否运行中
	runningMu     sync.RWMutex               // 运行状态锁
}

// NewWatcher 创建文件监听器
func NewWatcher(runtimeDir string, debounceTime time.Duration, handler ChangeHandler) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监听器失败: %w", err)
	}

	w := &Watcher{
		watcher:       fsWatcher,
		runtimeDir:    runtimeDir,
		debounceTime:  debounceTime,
		changeHandler: handler,
		debounceMap:   make(map[string]*time.Timer),
		stopCh:        make(chan struct{}),
		running:       false,
	}

	return w, nil
}

// Start 启动监听
func (w *Watcher) Start() error {
	w.runningMu.Lock()
	if w.running {
		w.runningMu.Unlock()
		return fmt.Errorf("监听器已在运行中")
	}
	w.running = true
	w.runningMu.Unlock()

	// 添加监听目录
	absRuntimeDir, err := filepath.Abs(w.runtimeDir)
	if err != nil {
		return fmt.Errorf("获取 runtime 目录绝对路径失败: %w", err)
	}

	if err := w.watcher.Add(absRuntimeDir); err != nil {
		return fmt.Errorf("添加监听目录失败: %w", err)
	}

	// 递归监听子目录
	if err := w.addSubDirectories(absRuntimeDir); err != nil {
		logger.Warn("添加子目录监听失败", zap.Error(err))
	}

	logger.Info("文件监听器已启动",
		zap.String("dir", absRuntimeDir),
		zap.Duration("debounce", w.debounceTime))

	// 启动监听循环
	go w.watchLoop()

	return nil
}

// Stop 停止监听
func (w *Watcher) Stop() error {
	w.runningMu.Lock()
	defer w.runningMu.Unlock()

	if !w.running {
		return nil
	}

	close(w.stopCh)
	w.running = false

	logger.Info("文件监听器已停止")

	return w.watcher.Close()
}

// watchLoop 监听循环
func (w *Watcher) watchLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("文件监听器错误", zap.Error(err))

		case <-w.stopCh:
			return
		}
	}
}

// handleEvent 处理文件事件
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// 只关心 .go 文件
	if !strings.HasSuffix(event.Name, ".go") {
		return
	}

	// 获取插件路径（相对于 runtime 目录）
	absRuntimeDir, _ := filepath.Abs(w.runtimeDir)
	relPath, err := filepath.Rel(absRuntimeDir, event.Name)
	if err != nil {
		logger.Error("计算相对路径失败", zap.Error(err))
		return
	}

	// 提取插件名称（第一级目录名）
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) == 0 {
		return
	}
	pluginPath := parts[0]

	// 判断事件类型
	var eventType string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = "write"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "remove"
	default:
		return // 忽略其他事件
	}

	logger.Debug("检测到文件变化",
		zap.String("file", event.Name),
		zap.String("plugin", pluginPath),
		zap.String("event", eventType))

	// 防抖动处理
	w.debounce(pluginPath, eventType)
}

// debounce 防抖动处理
// 同一个插件在短时间内的多次变化，只触发一次处理
func (w *Watcher) debounce(pluginPath string, eventType string) {
	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()

	// 取消之前的定时器
	if timer, exists := w.debounceMap[pluginPath]; exists {
		timer.Stop()
	}

	// 创建新的定时器
	w.debounceMap[pluginPath] = time.AfterFunc(w.debounceTime, func() {
		// 触发变化处理
		w.changeHandler(FileChangeEvent{
			PluginPath: pluginPath,
			EventType:  eventType,
			Timestamp:  time.Now(),
		})

		// 清理定时器
		w.debounceMu.Lock()
		delete(w.debounceMap, pluginPath)
		w.debounceMu.Unlock()
	})
}

// addSubDirectories 递归添加子目录监听
func (w *Watcher) addSubDirectories(dir string) error {
	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// 检查是否是目录
		info, err := filepath.Glob(entry)
		if err != nil || len(info) == 0 {
			continue
		}

		// 递归添加监听
		if err := w.watcher.Add(entry); err != nil {
			logger.Warn("添加子目录监听失败",
				zap.String("dir", entry),
				zap.Error(err))
		} else {
			logger.Debug("添加子目录监听",
				zap.String("dir", entry))
		}
	}

	return nil
}

// IsRunning 检查是否运行中
func (w *Watcher) IsRunning() bool {
	w.runningMu.RLock()
	defer w.runningMu.RUnlock()
	return w.running
}
