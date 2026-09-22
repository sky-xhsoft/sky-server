package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// cron 5 段解析:分 时 日 月 周 (0-59 0-23 1-31 1-12 0-6,0/7=周日)
// 支持: * 、 单一值、范围a-b、步长 /、列表, 和名(次要,主要支持数字)
// 返回该模式在给定时间后的下一次匹配时间,若永不匹配返回 zero。
func parseField(field string, min, max int) (map[int]bool, error) {
	allowed := make(map[int]bool)
	field = strings.TrimSpace(field)
	if field == "" {
		return nil, fmt.Errorf("空字段")
	}
	if field == "*" {
		for i := min; i <= max; i++ {
			allowed[i] = true
		}
		return allowed, nil
	}
	// 逗号分隔列表
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// 步长
		step := 1
		rangePart := part
		if idx := strings.Index(part, "/"); idx >= 0 {
			step, _ = strconv.Atoi(part[idx+1:])
			rangePart = part[:idx]
			if step <= 0 {
				step = 1
			}
		}
		var lo, hi int
		if strings.Contains(rangePart, "-") {
			segs := strings.SplitN(rangePart, "-", 2)
			lo, _ = strconv.Atoi(segs[0])
			hi, _ = strconv.Atoi(segs[1])
		} else {
			lo, _ = strconv.Atoi(rangePart)
			hi = lo
		}
		if lo < min {
			lo = min
		}
		if hi > max {
			hi = max
		}
		for v := lo; v <= hi; v += step {
			allowed[v] = true
		}
	}
	return allowed, nil
}

// cronSpec 解析后的 cron 计划
type cronSpec struct {
	min, hour, dom, mon, dow map[int]bool
}

func parseCron(expr string) (*cronSpec, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron 需 5 段(分 时 日 月 周),得到 %d 段: %q", len(fields), expr)
	}
	min, err := parseField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hour, err := parseField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}
	dom, err := parseField(fields[2], 1, 31)
	if err != nil {
		return nil, err
	}
	mon, err := parseField(fields[3], 1, 12)
	if err != nil {
		return nil, err
	}
	dow, err := parseField(fields[4], 0, 7)
	if err != nil {
		return nil, err
	}
	// 7 归一为 0(周日)
	if dow[7] {
		dow[0] = true
	}
	return &cronSpec{min: min, hour: hour, dom: dom, mon: mon, dow: dow}, nil
}

// next 计算 from 之后(>from)的下一次匹配时间
func (c *cronSpec) next(from time.Time) (time.Time, bool) {
	t := from.Add(time.Minute).Truncate(time.Minute)
	for i := 0; i < 366*24*60; i++ { // 上限约一年
		if c.mon[int(t.Month())] && c.dom[t.Day()] && c.dow[int(t.Weekday())] &&
			c.hour[t.Hour()] && c.min[t.Minute()] {
			return t, true
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, false
}

// NextRunAfter 解析 cron 并计算 from 之后的下一次运行时间(调度器与 API 共用)
func NextRunAfter(expr string, from time.Time) (time.Time, error) {
	spec, err := parseCron(expr)
	if err != nil {
		return time.Time{}, err
	}
	t, ok := spec.next(from)
	if !ok {
		return time.Time{}, fmt.Errorf("cron 表达式在未来一年内无匹配: %q", expr)
	}
	return t, nil
}