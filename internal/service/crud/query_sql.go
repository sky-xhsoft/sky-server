package crud

import (
	"fmt"
	"regexp"
	"strings"
)

// ============================================================
// C1 查询区智能转换(仿 Bos Portal QueryUtils.toSQLClause,MySQL 方言)
// 输入语义:
//   "=张"       -> 精确匹配        `col` = '张'
//   "张"        -> 模糊匹配        `col` LIKE '%张%'
//   "张*"/"张%" -> 前缀通配        `col` LIKE '张%'
//   "*张*"      -> 通配符 '*'->'%' `col` LIKE '%张%'
//   "10~20"     -> 区间(NUMBER)    `col` BETWEEN 10 AND 20
//   ">25","<5",">=3","!=5" -> 比较
//   数字"25"    -> 精确(NUMBER)   `col` = 25
//   日期范围 "2026-01-01~2026-02-01" / 数字日期 yyyymmdd -> BETWEEN
// ============================================================

var (
	reEq     = regexp.MustCompile(`^=(.+)$`)
	reRange  = regexp.MustCompile(`^([<>]=?|!?=?=)(.+)$`)
	reDate   = regexp.MustCompile(`^\d{8}$`) // yyyymmdd
	reNum    = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
)

// toSQLClause 把一个字段+用户输入转成语义化 WHERE 条件。
// 返回 SQL 子句与参数;无法识别时返回精确匹配。
// dbType: string / number / date / text(模糊)
func toSQLClause(columnName, input string, dbType string) (string, []interface{}) {
	col := "`" + columnName + "`"
	trimmed := strings.TrimSpace(input)

	if dbType == "number" {
		return numberClause(col, trimmed)
	}
	if dbType == "date" {
		return dateClause(col, trimmed)
	}

	// 字符串/text:支持 =精确、*通配、默认模糊
	return stringClause(col, trimmed)
}

func numberClause(col, in string) (string, []interface{}) {
	// 区间 10~20
	if i := strings.Index(in, "~"); i >= 0 {
		lo := strings.TrimSpace(in[:i])
		hi := strings.TrimSpace(in[i+1:])
		if lo != "" && hi != "" && reNum.MatchString(lo) && reNum.MatchString(hi) {
			return col + " BETWEEN ? AND ?", []interface{}{lo, hi}
		}
	}
	// 比较 >25, <5, >=3, !=5
	if m := reRange.FindStringSubmatch(in); m != nil && len(m) >= 3 {
		op := m[1]
		val := strings.TrimSpace(m[2])
		if reNum.MatchString(val) {
			return col + " " + op + " ?", []interface{}{val}
		}
	}
	// 列表逗号
	if strings.Contains(in, ",") {
		parts := splitNonEmpty(in)
		if len(parts) > 0 {
			ph := strings.Repeat("?,", len(parts))
			return col + " IN (" + ph[:len(ph)-1] + ")", toIfaces(parts)
		}
	}
	// 精确
	if reNum.MatchString(in) {
		return col + " = ?", []interface{}{in}
	}
	return col + " = ?", []interface{}{in}
}

func dateClause(col, in string) (string, []interface{}) {
	// 日期范围 a~b,支持 yyyymmdd 或 yyyy-mm-dd
	if i := strings.Index(in, "~"); i >= 0 {
		lo := strings.TrimSpace(in[:i])
		hi := strings.TrimSpace(in[i+1:])
		lo = normalizeDate(lo)
		hi = normalizeDate(hi)
		if lo != "" && hi != "" {
			return col + " BETWEEN ? AND ?", []interface{}{lo, hi}
		}
	}
	// 精确日期 / 数字日期
	d := normalizeDate(in)
	if d != "" {
		return col + " = ?", []interface{}{d}
	}
	return col + " = ?", []interface{}{in}
}

func stringClause(col, in string) (string, []interface{}) {
	// =精确
	if m := reEq.FindStringSubmatch(in); m != nil {
		return col + " = ?", []interface{}{m[1]}
	}
	// 含 * 当作 LIKE 通配(*->%)
	if strings.Contains(in, "*") {
		return col + " LIKE ?", []interface{}{strings.ReplaceAll(in, "*", "%")}
	}
	// 含 ? 当作单字符通配
	if strings.Contains(in, "?") {
		return col + " LIKE ?", []interface{}{strings.ReplaceAll(in, "?", "_")}
	}
	// 默认模糊
	return col + " LIKE ?", []interface{}{"%" + in + "%"}
}

// normalizeDate 把 yyyymmdd → yyyy-mm-dd;已是 yyyy-mm-dd 原样;否则返回空
func normalizeDate(s string) string {
	s = strings.TrimSpace(s)
	if reDate.MatchString(s) {
		return fmt.Sprintf("%s-%s-%s", s[0:4], s[4:6], s[6:8])
	}
	if len(s) == 10 && s[4] == '-' && s[7] == '-' {
		return s
	}
	return ""
}

func splitNonEmpty(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toIfaces(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}