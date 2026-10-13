package crud

import (
	"fmt"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// uniqueColumn 需要在有效记录中全局唯一的字段
type uniqueColumn struct {
	Column string // 字段名
	Label  string // 重复时提示用的名称
}

// uniqueColumns 按表名（大写）配置的唯一字段。
// 用户名、手机号全局唯一：业务表的 CREATE_BY 记录用户名，登录按用户名 / 手机号查找，重复会串到其他账号。
// 注册入口另有同样的校验（company.Register）。
var uniqueColumns = map[string][]uniqueColumn{
	"SYS_USER": {{Column: "USERNAME", Label: "用户名"}, {Column: "PHONE", Label: "手机号"}},
}

// checkUniqueColumns 新增（id 为 0）或修改时检查唯一字段是否已被其他有效记录使用。
// 空值不检查；修改时只检查值有变化的字段，避免历史重复数据导致该记录的其他字段无法修改。
func checkUniqueColumns(db *gorm.DB, tableName string, id uint, data map[string]interface{}) error {
	cols := uniqueColumns[strings.ToUpper(tableName)]
	if len(cols) == 0 {
		return nil
	}
	var current map[string]interface{}
	for _, col := range cols {
		raw, ok := data[col.Column]
		if !ok || raw == nil {
			continue
		}
		value := uniqueValueString(raw)
		if value == "" {
			continue
		}
		if id > 0 {
			if current == nil {
				current = map[string]interface{}{}
				if err := db.Table(tableName).Where("ID = ?", id).Take(&current).Error; err != nil && err != gorm.ErrRecordNotFound {
					return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
				}
			}
			if old, ok := current[col.Column]; ok && old != nil && strings.EqualFold(uniqueValueString(old), value) {
				continue
			}
		}
		var n int64
		q := db.Table(tableName).Where(col.Column+" = ? AND IS_ACTIVE = ?", value, "Y")
		if id > 0 {
			q = q.Where("ID <> ?", id)
		}
		if err := q.Count(&n).Error; err != nil {
			return errors.Wrap(errors.ErrDatabase, "检查重复失败", err)
		}
		if n > 0 {
			return errors.New(errors.ErrResourceExists, fmt.Sprintf("该%s已被使用", col.Label))
		}
	}
	return nil
}

// uniqueValueString 统一转成去掉首尾空白的字符串（MySQL 驱动读出的文本可能是 []byte）
func uniqueValueString(v interface{}) string {
	if b, ok := v.([]byte); ok {
		return strings.TrimSpace(string(b))
	}
	return strings.TrimSpace(fmt.Sprint(v))
}
