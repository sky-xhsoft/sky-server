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
	// 逗号分隔列表；每项为 *、a、a-b，可带步长 /n（*/n、a-b/n、a/n 表示从 a 到上限每 n）
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		step := 1
		rangePart := part
		if idx := strings.Index(part, "/"); idx >= 0 {
			n, err := strconv.Atoi(part[idx+1:])
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("步长非法: %q", part)
			}
			step = n
			rangePart = part[:idx]
		}
		var lo, hi int
		switch {
		case rangePart == "*":
			lo, hi = min, max
		case strings.Contains(rangePart, "-"):
			segs := strings.SplitN(rangePart, "-", 2)
			a, err1 := strconv.Atoi(segs[0])
			b, err2 := strconv.Atoi(segs[1])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("范围非法: %q", part)
			}
			lo, hi = a, b
		default:
			v, err := strconv.Atoi(rangePart)
			if err != nil {
				return nil, fmt.Errorf("取值非法: %q", part)
			}
			lo, hi = v, v
			if step > 1 {
				hi = max
			}
		}
		if lo < min || hi > max || lo > hi {
			return nil, fmt.Errorf("取值超出范围 %d-%d: %q", min, max, part)
		}
		for v := lo; v <= hi; v += step {
			allowed[v] = true
		}
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("空字段")
	}
	return allowed, nil
}

// cronSpec 解析后的 cron 计划
type cronSpec struct {
	min, hour, dom, mon, dow map[int]bool
	// 日/周字段是否以 * 开头（不受限）。标准 cron：两者都受限时任一满足即可
	domAny, dowAny bool
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
	return &cronSpec{min: min, hour: hour, dom: dom, mon: mon, dow: dow,
		domAny: strings.HasPrefix(fields[2], "*"), dowAny: strings.HasPrefix(fields[4], "*")}, nil
}

// next 计算 from 之后(>from)的下一次匹配时间
func (c *cronSpec) next(from time.Time) (time.Time, bool) {
	t := from.Add(time.Minute).Truncate(time.Minute)
	for i := 0; i < 366*24*60; i++ { // 上限约一年
		if c.mon[int(t.Month())] && c.dayMatches(t) && c.hour[t.Hour()] && c.min[t.Minute()] {
			return t, true
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, false
}

// dayMatches 日与周的组合判断（与标准 cron 一致）
func (c *cronSpec) dayMatches(t time.Time) bool {
	dom, dow := c.dom[t.Day()], c.dow[int(t.Weekday())]
	switch {
	case c.domAny && c.dowAny:
		return true
	case c.domAny:
		return dow
	case c.dowAny:
		return dom
	default:
		return dom || dow
	}
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
