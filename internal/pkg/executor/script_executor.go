package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// ScriptExecutor 脚本执行器接口
type ScriptExecutor interface {
	// 执行脚本
	Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error)
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success  bool                   `json:"success"`
	Output   string                 `json:"output"`
	Error    string                 `json:"error"`
	ExitCode int                    `json:"exitCode"`
	Duration time.Duration          `json:"duration"`
	Data     map[string]interface{} `json:"data"`
}

// ScriptType 脚本类型
type ScriptType string

const (
	ScriptTypeJavaScript ScriptType = "js"
	ScriptTypePython     ScriptType = "py"
	ScriptTypeGo         ScriptType = "go"
	ScriptTypeBash       ScriptType = "bsh"
)

// NewScriptExecutor 创建脚本执行器
func NewScriptExecutor(scriptType ScriptType, timeout time.Duration) ScriptExecutor {
	switch scriptType {
	case ScriptTypeJavaScript:
		return &jsExecutor{timeout: timeout}
	case ScriptTypePython:
		return &pythonExecutor{timeout: timeout}
	case ScriptTypeGo:
		return &goExecutor{timeout: timeout}
	case ScriptTypeBash:
		return &bashExecutor{timeout: timeout}
	default:
		return &bashExecutor{timeout: timeout}
	}
}

// bashExecutor Bash脚本执行器
type bashExecutor struct {
	timeout time.Duration
}

func (e *bashExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	// 创建临时脚本文件
	tmpFile, err := e.createTempScript(script, params)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建临时脚本失败", err)
	}
	defer os.Remove(tmpFile)

	// 设置超时
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 执行脚本
	start := time.Now()
	cmd := exec.CommandContext(execCtx, "bash", tmpFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range params {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", key, value))
	}

	err = cmd.Run()
	duration := time.Since(start)

	result := &ExecutionResult{
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
		Data:     make(map[string]interface{}),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}

func (e *bashExecutor) createTempScript(script string, params map[string]interface{}) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("script_%d.sh", time.Now().UnixNano()))

	// 添加参数注释
	content := "#!/bin/bash\n"
	content += "# Auto-generated script\n"
	content += "# Parameters:\n"
	for key, value := range params {
		content += fmt.Sprintf("# %s=%v\n", key, value)
	}
	content += "\n"
	content += script

	if err := os.WriteFile(tmpFile, []byte(content), 0755); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// pythonExecutor Python脚本执行器
type pythonExecutor struct {
	timeout time.Duration
}

func (e *pythonExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	// 创建临时脚本文件
	tmpFile, err := e.createTempScript(script, params)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建临时脚本失败", err)
	}
	defer os.Remove(tmpFile)

	// 设置超时
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 执行脚本
	start := time.Now()
	cmd := exec.CommandContext(execCtx, "python3", tmpFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range params {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", key, value))
	}

	err = cmd.Run()
	duration := time.Since(start)

	result := &ExecutionResult{
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
		Data:     make(map[string]interface{}),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}

func (e *pythonExecutor) createTempScript(script string, params map[string]interface{}) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("script_%d.py", time.Now().UnixNano()))

	// 添加参数导入
	content := "#!/usr/bin/env python3\n"
	content += "# -*- coding: utf-8 -*-\n"
	content += "# Auto-generated script\n"
	content += "import os\n"
	content += "import sys\n"
	content += "import json\n\n"

	// 从环境变量读取参数
	content += "# Read parameters from environment\n"
	content += "params = {}\n"
	for key := range params {
		content += fmt.Sprintf("params['%s'] = os.getenv('%s')\n", key, key)
	}
	content += "\n"
	content += script

	if err := os.WriteFile(tmpFile, []byte(content), 0755); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// jsExecutor JavaScript脚本执行器（使用Node.js）
type jsExecutor struct {
	timeout time.Duration
}

func (e *jsExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	// 创建临时脚本文件
	tmpFile, err := e.createTempScript(script, params)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "创建临时脚本失败", err)
	}
	defer os.Remove(tmpFile)

	// 设置超时
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 执行脚本
	start := time.Now()
	cmd := exec.CommandContext(execCtx, "node", tmpFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range params {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", key, value))
	}

	err = cmd.Run()
	duration := time.Since(start)

	result := &ExecutionResult{
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
		Data:     make(map[string]interface{}),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}

func (e *jsExecutor) createTempScript(script string, params map[string]interface{}) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("script_%d.js", time.Now().UnixNano()))

	// 添加参数
	content := "// Auto-generated script\n"
	content += "// Parameters from environment:\n"
	content += "const params = {};\n"
	for key := range params {
		content += fmt.Sprintf("params.%s = process.env.%s;\n", key, key)
	}
	content += "\n"
	content += script

	if err := os.WriteFile(tmpFile, []byte(content), 0755); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// goExecutor Go方法调用执行器
type goExecutor struct {
	timeout time.Duration
}

// GoFuncRegistry Go函数注册表
var GoFuncRegistry = make(map[string]func(map[string]interface{}) (interface{}, error))

// RegisterGoFunc 注册Go函数
func RegisterGoFunc(name string, fn func(map[string]interface{}) (interface{}, error)) {
	GoFuncRegistry[name] = fn
}

func (e *goExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (*ExecutionResult, error) {
	start := time.Now()

	result := &ExecutionResult{
		Duration: 0,
		Data:     make(map[string]interface{}),
	}

	// script 是要调用的 Go 函数名
	funcName := script

	// 从注册表中查找函数
	fn, exists := GoFuncRegistry[funcName]
	if !exists {
		result.Success = false
		result.Error = fmt.Sprintf("Go函数 '%s' 未注册", funcName)
		result.Duration = time.Since(start)
		return result, nil
	}

	// 使用 channel 来处理超时
	type execResult struct {
		data interface{}
		err  error
	}
	resultChan := make(chan execResult, 1)

	// 执行函数
	go func() {
		data, err := fn(params)
		resultChan <- execResult{data: data, err: err}
	}()

	// 等待执行结果或超时
	select {
	case <-ctx.Done():
		result.Success = false
		result.Error = "执行超时或被取消"
		result.Duration = time.Since(start)
		return result, nil

	case <-time.After(e.timeout):
		result.Success = false
		result.Error = fmt.Sprintf("执行超时（%v）", e.timeout)
		result.Duration = time.Since(start)
		return result, nil

	case res := <-resultChan:
		result.Duration = time.Since(start)
		if res.err != nil {
			result.Success = false
			result.Error = res.err.Error()
			return result, nil
		}

		result.Success = true
		result.ExitCode = 0

		// 将返回数据转换为 map
		if res.data != nil {
			if dataMap, ok := res.data.(map[string]interface{}); ok {
				result.Data = dataMap
			} else {
				result.Data["result"] = res.data
			}
			result.Output = fmt.Sprintf("%v", res.data)
		}

		return result, nil
	}
}
