package executor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// SPExecutor 存储过程执行器
type SPExecutor struct {
	db *gorm.DB
}

// NewSPExecutor 创建存储过程执行器
func NewSPExecutor(db *gorm.DB) *SPExecutor {
	return &SPExecutor{
		db: db,
	}
}

// SPRequest 存储过程请求
type SPRequest struct {
	Name      string                 `json:"name"`      // 存储过程名称
	InParams  map[string]interface{} `json:"inParams"`  // 输入参数
	OutParams []string               `json:"outParams"` // 输出参数名称列表
}

// SPResponse 存储过程响应
type SPResponse struct {
	Success    bool                   `json:"success"`
	OutParams  map[string]interface{} `json:"outParams"`  // 输出参数值
	ResultSets [][]map[string]interface{} `json:"resultSets"` // 结果集
	RowsAffected int64                `json:"rowsAffected"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error"`
}

// Execute 执行存储过程
func (e *SPExecutor) Execute(ctx context.Context, req *SPRequest) (*SPResponse, error) {
	start := time.Now()

	response := &SPResponse{
		Success:    false,
		OutParams:  make(map[string]interface{}),
		ResultSets: [][]map[string]interface{}{},
	}

	// 获取原始数据库连接
	sqlDB, err := e.db.DB()
	if err != nil {
		response.Error = fmt.Sprintf("获取数据库连接失败: %v", err)
		response.Duration = time.Since(start)
		return response, nil
	}

	// 构建CALL语句
	callStmt, args := e.buildCallStatement(req)

	// 执行存储过程
	rows, err := sqlDB.QueryContext(ctx, callStmt, args...)
	if err != nil {
		response.Error = fmt.Sprintf("执行存储过程失败: %v", err)
		response.Duration = time.Since(start)
		return response, nil
	}
	defer rows.Close()

	// 读取所有结果集
	for {
		// 读取当前结果集
		resultSet, err := e.fetchResultSet(rows)
		if err != nil {
			response.Error = fmt.Sprintf("读取结果集失败: %v", err)
			response.Duration = time.Since(start)
			return response, nil
		}

		if len(resultSet) > 0 {
			response.ResultSets = append(response.ResultSets, resultSet)
		}

		// 检查是否还有更多结果集
		if !rows.NextResultSet() {
			break
		}
	}

	response.Success = true
	response.Duration = time.Since(start)

	// TODO: 处理输出参数（需要根据具体数据库类型实现）
	// MySQL使用SELECT @out_param获取输出参数
	// PostgreSQL使用函数返回值

	return response, nil
}

// buildCallStatement 构建CALL语句
func (e *SPExecutor) buildCallStatement(req *SPRequest) (string, []interface{}) {
	var args []interface{}
	var placeholders []string

	// 按顺序构建输入参数
	if req.InParams != nil {
		for _, value := range req.InParams {
			args = append(args, value)
			placeholders = append(placeholders, "?")
		}
	}

	// 添加输出参数占位符
	for range req.OutParams {
		placeholders = append(placeholders, "@?")
	}

	callStmt := fmt.Sprintf("CALL %s(%s)", req.Name, strings.Join(placeholders, ", "))
	return callStmt, args
}

// fetchResultSet 读取结果集
func (e *SPExecutor) fetchResultSet(rows *sql.Rows) ([]map[string]interface{}, error) {
	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	// 读取每一行
	for rows.Next() {
		// 创建值容器
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 构建结果map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// 处理字节数组
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteFunction 执行数据库函数（返回单个值）
func (e *SPExecutor) ExecuteFunction(ctx context.Context, funcName string, params map[string]interface{}) (interface{}, error) {
	// 构建SELECT语句
	var args []interface{}
	var placeholders []string

	if params != nil {
		for _, value := range params {
			args = append(args, value)
			placeholders = append(placeholders, "?")
		}
	}

	query := fmt.Sprintf("SELECT %s(%s)", funcName, strings.Join(placeholders, ", "))

	// 执行查询
	var result interface{}
	if err := e.db.WithContext(ctx).Raw(query, args...).Scan(&result).Error; err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, "执行函数失败", err)
	}

	return result, nil
}
