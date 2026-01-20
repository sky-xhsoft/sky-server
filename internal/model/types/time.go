package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 自定义时间类型，用于格式化 JSON 输出
type LocalTime time.Time

const (
	// TimeFormat 时间格式：年-月-日 时:分:秒
	TimeFormat = "2006-01-02 15:04:05"
)

// MarshalJSON 实现 JSON 序列化
func (t LocalTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf(`"%s"`, time.Time(t).Format(TimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON 实现 JSON 反序列化
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	// 去掉引号
	str := string(data)
	if str == "null" || str == `""` {
		*t = LocalTime(time.Time{})
		return nil
	}

	// 移除首尾引号
	if len(str) > 2 {
		str = str[1 : len(str)-1]
	}

	// 尝试多种时间格式
	formats := []string{
		TimeFormat,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02",
	}

	var parsedTime time.Time
	var err error
	for _, format := range formats {
		parsedTime, err = time.Parse(format, str)
		if err == nil {
			*t = LocalTime(parsedTime)
			return nil
		}
	}

	return fmt.Errorf("无法解析时间: %s", str)
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (t LocalTime) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		*t = LocalTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = LocalTime(v)
		return nil
	case []byte:
		return t.UnmarshalJSON(v)
	case string:
		return t.UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("无法将 %T 转换为 LocalTime", value)
	}
}

// Time 转换为标准 time.Time
func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

// String 返回格式化的时间字符串
func (t LocalTime) String() string {
	return time.Time(t).Format(TimeFormat)
}

// IsZero 判断是否为零值
func (t LocalTime) IsZero() bool {
	return time.Time(t).IsZero()
}
