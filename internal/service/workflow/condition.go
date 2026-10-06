package workflow

import (
	"fmt"
	"strconv"
	"strings"
)

// evalCondition 计算流转条件表达式，不执行任何代码，只做比较：
//
//	条件 := 与式 ( "||" 与式 )*
//	与式 := 比较 ( "&&" 比较 )*
//	比较 := 变量 运算符 值        运算符: == != > >= < <=
//	值   := 数字 | '字符串' | "字符串" | true | false
//
// 例：amount > 10000 && dept == "sales"。空条件恒为真；无法解析或变量缺失时为假（不走该分支）。
func evalCondition(condition string, variables map[string]interface{}) (bool, error) {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return true, nil
	}
	for _, orPart := range strings.Split(condition, "||") {
		all := true
		for _, andPart := range strings.Split(orPart, "&&") {
			ok, err := evalComparison(strings.TrimSpace(andPart), variables)
			if err != nil {
				return false, err
			}
			if !ok {
				all = false
				break
			}
		}
		if all {
			return true, nil
		}
	}
	return false, nil
}

// 按长度从长到短匹配，避免 ">=" 被识别成 ">"
var comparisonOps = []string{"==", "!=", ">=", "<=", ">", "<"}

func evalComparison(expr string, variables map[string]interface{}) (bool, error) {
	for _, op := range comparisonOps {
		idx := strings.Index(expr, op)
		if idx <= 0 {
			continue
		}
		name := strings.TrimSpace(expr[:idx])
		literal := strings.TrimSpace(expr[idx+len(op):])
		if name == "" || literal == "" {
			return false, fmt.Errorf("条件格式错误: %q", expr)
		}
		actual, exists := variables[name]
		if !exists || actual == nil {
			return false, nil
		}
		expected := parseLiteral(literal)
		return compare(actual, op, expected)
	}
	return false, fmt.Errorf("条件缺少比较运算符: %q", expr)
}

// parseLiteral 解析字面量：带引号为字符串，true/false 为布尔，其余尝试数字，否则按字符串
func parseLiteral(s string) interface{} {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	}
	return 0, false
}

func compare(actual interface{}, op string, expected interface{}) (bool, error) {
	// 期望值是数字且实际值可转为数字：按数值比较（JSON 变量反序列化后数字为 float64）
	if e, ok := expected.(float64); ok {
		if a, ok := toFloat(actual); ok {
			switch op {
			case "==":
				return a == e, nil
			case "!=":
				return a != e, nil
			case ">":
				return a > e, nil
			case ">=":
				return a >= e, nil
			case "<":
				return a < e, nil
			case "<=":
				return a <= e, nil
			}
		}
	}
	// 其余按字符串比较，只支持相等判断
	a, e := fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)
	switch op {
	case "==":
		return a == e, nil
	case "!=":
		return a != e, nil
	}
	return false, fmt.Errorf("非数值不支持运算符 %s", op)
}
