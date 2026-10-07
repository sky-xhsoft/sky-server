package crud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"github.com/sky-xhsoft/sky-server/internal/pkg/mask"
	"golang.org/x/crypto/bcrypt"
)

// 系统审计字段（无论MASK如何配置都应包含）
const (
	colID           = "ID"
	colSysCompanyID = "SYS_COMPANY_ID"
	colCreateBy     = "CREATE_BY"
	colCreateTime   = "CREATE_TIME"
	colUpdateBy     = "UPDATE_BY"
	colUpdateTime   = "UPDATE_TIME"
	colIsActive     = "IS_ACTIVE"
	colStatus       = "STATUS"
	colOrderNo      = "ORDERNO"
)

var standardFields = map[string]bool{
	colID:           true,
	colSysCompanyID: true,
	colCreateBy:     true,
	colCreateTime:   true,
	colUpdateBy:     true,
	colUpdateTime:   true,
	colIsActive:     true,
}

// normalizeFieldValue 规范化前端传入的字段值：
// 数值列与外键列的空字符串视为 NULL（导入时留空的可选外键原先被当作 "" 校验而整行失败）；
// 数值列传入非数字时返回参数错误（原先直接写库，MySQL 报错返回 500）
func normalizeFieldValue(col *entity.SysColumn, value interface{}) (interface{}, error) {
	str, isStr := value.(string)
	if !isStr {
		return value, nil
	}
	isNumber := inferDBType(col) == "number"
	if !isNumber && col.SetValueType != "fk" {
		return value, nil
	}
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return nil, nil
	}
	if isNumber {
		if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
			name := col.DisplayName
			if name == "" {
				name = col.DbName
			}
			return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s 必须是数字", name))
		}
		return trimmed, nil
	}
	return value, nil
}

// skippedUpdateFields 请求中的业务字段在处理后一个都没有保留时，返回这些字段的显示名；
// 只要有字段被保留就返回空（部分字段只读属于正常情况）。
// 密码/加密字段留空表示不修改，不计入。
func skippedUpdateFields(columns []*entity.SysColumn, data, processed map[string]interface{}) []string {
	var requested []string
	for _, col := range columns {
		if standardFields[col.DbName] {
			continue
		}
		value, ok := data[col.DbName]
		if !ok {
			continue
		}
		if col.SetValueType == "password" || col.SetValueType == "encrypt" {
			if s, isStr := value.(string); isStr && s == "" {
				continue
			}
		}
		if _, kept := processed[col.DbName]; kept {
			return nil
		}
		name := col.DisplayName
		if name == "" {
			name = col.DbName
		}
		requested = append(requested, name)
	}
	return requested
}

// buildSelectFields 构建查询字段列表（根据MASK控制）
func (s *service) buildSelectFields(columns []*entity.SysColumn, userID uint, operation string) (string, error) {
	var fields []string

	addedStandardFields := make(map[string]bool)
	grade := s.userFieldGrade(userID)

	for _, col := range columns {
		// 跳过UI-only字段（不在真实数据库表中）
		if col.DisplayType == "button" || col.DisplayType == "blank" || col.DisplayType == "hr" {
			continue
		}
		// 字段访问级别高于用户级别：不返回
		if !standardFields[col.DbName] && !fieldAllowed(col.Sgrade, grade) {
			continue
		}

		// 敏感字段（密码/密钥）永不返回给前端
		if col.SetValueType == "password" || col.SetValueType == "encrypt" {
			continue
		}

		// C4: 虚拟字段 → 用表达式生成计算列(仅合法白名单表达式)
		if col.IsVirtual == "Y" {
			if expr, ok := sanitizeVirtualExpr(col.Expression); ok && expr != "" {
				fields = append(fields, fmt.Sprintf("%s AS `%s`", expr, col.DbName))
			} else if col.IsAK != "Y" {
				// 无有效表达式则不输出该列
				continue
			}
			continue
		}

		isPrimaryKey := col.IsAK == "Y" || col.SetValueType == "pk"
		isStandardField := standardFields[col.DbName]
		isForeignKey := col.SetValueType == "fk"

		if isStandardField {
			addedStandardFields[col.DbName] = true
		}

		if !isPrimaryKey && !isStandardField && !isForeignKey {
			if col.Mask != "" {
				fieldMask := mask.ParseMask(col.Mask)
				if !fieldMask.IsVisible(operation) {
					continue
				}
			}
		}

		fields = append(fields, col.DbName)
	}

	for fieldName := range standardFields {
		if !addedStandardFields[fieldName] {
			fields = append(fields, fieldName)
		}
	}

	if len(fields) == 0 {
		return "*", nil
	}

	return strings.Join(fields, ", "), nil
}

// processFieldsForCreate 处理创建时的字段
func (s *service) processFieldsForCreate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	grade := s.userFieldGrade(userID)

	for _, col := range columns {
		// 字段访问级别高于用户级别：忽略提交的值
		if !standardFields[col.DbName] && !fieldAllowed(col.Sgrade, grade) {
			continue
		}
		// 跳过UI-only字段
		if col.DisplayType == "button" || col.DisplayType == "blank" || col.DisplayType == "hr" {
			continue
		}

		// C4: 跳过虚拟字段(只读展示不落库)
		if col.IsVirtual == "Y" {
			continue
		}

		value, exists := data[col.DbName]
		if exists {
			normalized, err := normalizeFieldValue(col, value)
			if err != nil {
				return nil, err
			}
			data[col.DbName] = normalized
			value = normalized
		}

		// SetValueType 策略处理（必须在 NullAble 之前，因为自动填充字段不需要前端传值）
		switch col.SetValueType {
		case "ignore":
			delete(data, col.DbName)
			continue
		case "sysdate":
			now := time.Now()
			data[col.DbName] = now
			value = now
			exists = true
		case "createBy":
			data[col.DbName] = userID
			value = userID
			exists = true
		case "operator":
			if user, err := s.userRepo.GetUserByID(userID); err == nil && user != nil {
				data[col.DbName] = user.Username
				value = user.Username
				exists = true
			}
		case "password":
			if strVal, ok := value.(string); ok && strVal != "" {
				hashed, err := bcrypt.GenerateFromPassword([]byte(strVal), bcrypt.DefaultCost)
				if err != nil {
					return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("字段 %s: 密码加密失败", col.DisplayName), err)
				}
				data[col.DbName] = string(hashed)
				value = string(hashed)
			}
		case "encrypt":
			if strVal, ok := value.(string); ok && strVal != "" {
				encrypted, err := fieldcrypto.EncryptField(strVal)
				if err != nil {
					return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("字段 %s: 加密失败", col.DisplayName), err)
				}
				data[col.DbName] = encrypted
				value = encrypted
			}
		case "docno":
			if col.Seq != "" {
				docNo, err := s.sequenceService.NextValue(col.Seq)
				if err != nil {
					return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("字段 %s: 生成单据编号失败", col.DisplayName), err)
				}
				data[col.DbName] = docNo
				value = docNo
				exists = true
			}
		}

		// 默认值填充（在 NullAble 之前，因为默认值可以满足必填要求）
		if !exists && col.DefaultValue != "" {
			value = col.DefaultValue
			exists = true
			data[col.DbName] = col.DefaultValue
		}

		isAutoFill := col.SetValueType == "sysdate" || col.SetValueType == "createBy" || col.SetValueType == "operator" || col.SetValueType == "docno" || col.SetValueType == "pk" || col.DefaultValue != ""

		if col.NullAble == "N" && !isAutoFill && (!exists || value == nil) {
			return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s 不能为空", col.DisplayName))
		}
		if col.NullAble == "N" && !isAutoFill && exists && value != nil {
			if str, ok := value.(string); ok && str == "" {
				return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s 不能为空", col.DisplayName))
			}
		}

		if col.RegExpression != "" && exists && value != nil {
			if strVal, ok := value.(string); ok {
				re := regexCache.get(col.RegExpression)
				if !re.MatchString(strVal) {
					return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s 格式不正确", col.DisplayName))
				}
			}
		}

		switch col.SetValueType {
		case "select":
			if col.SysDictID != "" && exists && value != nil {
				dictID := parseDictID(col.SysDictID)
				if dictID > 0 {
					items, err := s.dictService.GetDictItems(dictID)
					if err == nil && len(items) > 0 {
						found := false
						for _, item := range items {
							if item.Value == fmt.Sprintf("%v", value) {
								found = true
								break
							}
						}
						if !found {
							return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: 值不在字典选项中", col.DisplayName))
						}
					}
				}
			}
		}

		if exists && col.SetValueType == "fk" && col.RefTableID != nil {
			if value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
			continue
		}

		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("add") {
				continue
			}
		}

		if exists {
			if col.DisplayType == "json" && value != nil {
				switch v := value.(type) {
				case map[string]interface{}, []interface{}:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				case string:
					processedData[col.DbName] = v
				default:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				}
			} else {
				if col.IsUppercase == "Y" && value != nil {
					if strValue, ok := value.(string); ok {
						processedData[col.DbName] = strings.ToUpper(strValue)
					} else {
						processedData[col.DbName] = value
					}
				} else {
					processedData[col.DbName] = value
				}
			}
		}
	}

	return processedData, nil
}

// processFieldsForUpdate 处理更新时的字段
func (s *service) processFieldsForUpdate(columns []*entity.SysColumn, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	processedData := make(map[string]interface{})

	grade := s.userFieldGrade(userID)

	for _, col := range columns {
		// 字段访问级别高于用户级别：忽略提交的值
		if !standardFields[col.DbName] && !fieldAllowed(col.Sgrade, grade) {
			continue
		}
		// 跳过UI-only字段
		if col.DisplayType == "button" || col.DisplayType == "blank" || col.DisplayType == "hr" {
			continue
		}

		// C4: 跳过虚拟字段(只读展示不落库)
		if col.IsVirtual == "Y" {
			continue
		}

		value, exists := data[col.DbName]
		if exists {
			normalized, err := normalizeFieldValue(col, value)
			if err != nil {
				return nil, err
			}
			data[col.DbName] = normalized
			value = normalized
		}

		if col.RegExpression != "" && exists && value != nil {
			if strVal, ok := value.(string); ok {
				re := regexCache.get(col.RegExpression)
				if !re.MatchString(strVal) {
					return nil, errors.New(errors.ErrInvalidParam, fmt.Sprintf("字段 %s 格式不正确", col.DisplayName))
				}
			}
		}

		if exists && col.SetValueType == "fk" && col.RefTableID != nil {
			if value != nil {
				if err := s.validateForeignKeyValue(*col.RefTableID, value); err != nil {
					return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: %s", col.DisplayName, err.Error()), err)
				}
			}
			processedData[col.DbName] = value
			continue
		}

		if col.Mask != "" {
			fieldMask := mask.ParseMask(col.Mask)
			if !fieldMask.IsEditable("edit") {
				continue
			}
		}

		// 敏感字段：查询时不返回，编辑时留空表示不修改；有值则 bcrypt/AES 加密后更新
		if col.SetValueType == "password" {
			if strVal, ok := value.(string); ok && strVal != "" {
				hashed, err := bcrypt.GenerateFromPassword([]byte(strVal), bcrypt.DefaultCost)
				if err != nil {
					return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("字段 %s: 密码加密失败", col.DisplayName), err)
				}
				processedData[col.DbName] = string(hashed)
			}
			continue
		}
		if col.SetValueType == "encrypt" {
			if strVal, ok := value.(string); ok && strVal != "" {
				encrypted, err := fieldcrypto.EncryptField(strVal)
				if err != nil {
					return nil, errors.Wrap(errors.ErrInternal, fmt.Sprintf("字段 %s: 加密失败", col.DisplayName), err)
				}
				processedData[col.DbName] = encrypted
			}
			continue
		}

		if exists {
			if col.DisplayType == "json" && value != nil {
				switch v := value.(type) {
				case map[string]interface{}, []interface{}:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				case string:
					processedData[col.DbName] = v
				default:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return nil, errors.Wrap(errors.ErrInvalidParam, fmt.Sprintf("字段 %s: JSON 序列化失败", col.DisplayName), err)
					}
					processedData[col.DbName] = string(jsonBytes)
				}
			} else {
				if col.IsUppercase == "Y" && value != nil {
					if strValue, ok := value.(string); ok {
						processedData[col.DbName] = strings.ToUpper(strValue)
					} else {
						processedData[col.DbName] = value
					}
				} else {
					processedData[col.DbName] = value
				}
			}
		}
	}

	return processedData, nil
}

var regexCache = &regexCacheImpl{cache: make(map[string]*regexp.Regexp)}

type regexCacheImpl struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

func (r *regexCacheImpl) get(pattern string) *regexp.Regexp {
	r.mu.RLock()
	if re, ok := r.cache[pattern]; ok {
		r.mu.RUnlock()
		return re
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if re, ok := r.cache[pattern]; ok {
		return re
	}
	re := regexp.MustCompile(pattern)
	r.cache[pattern] = re
	return re
}

func parseDictID(s string) uint {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

// sanitizeVirtualExpr 校验虚拟字段表达式语法安全。
// 仅允许白名单函数 + 标识符/DIGIT/括号/逗号/单引号/运算符,返回规范化表达式。
// 非法或含空格外的危险字符则返回 ok=false(调用方跳过该列,不拼入 SQL)。
var virtualFnWhitelist = map[string]bool{
	"CONCAT": true, "UPPER": true, "LOWER": true, "TRIM": true, "LEN": true,
	"ROUND": true, "CEIL": true, "FLOOR": true, "ABS": true,
	"DATE_FORMAT": true, "YEAR": true, "MONTH": true, "DAY": true,
	"NOW": true, "CURDATE": true, "IFNULL": true, "COALESCE": true, "CAST": true,
}

var virtualTokRe = regexp.MustCompile(`^[A-Za-z0-9_ 	(),.'+-]+$`)

func sanitizeVirtualExpr(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", true
	}
	if len(expr) > 500 {
		return "", false
	}
	if !virtualTokRe.MatchString(expr) {
		return "", false
	}
	// 逐个 token 校验:标识符(列名)、数字、函数名(白名单)、括号/逗号/运算符
	reTok := regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*|\d+(\.\d+)?|[-+*/&=<>(),' ]`)
	for _, tok := range reTok.FindAllString(expr, -1) {
		t := strings.TrimSpace(tok)
		if t == "" {
			continue
		}
		if reNum.MatchString(t) || t == "(" || t == ")" || t == "," || t == "'" ||
			t == "+" || t == "-" || t == "*" || t == "/" || t == "=" || t == "<" || t == ">" {
			continue
		}
		// 函数:必须白名单;否则视为列名(仅字母数字下划线)
		if strings.HasSuffix(t, "(") {
			fn := strings.TrimSuffix(strings.ToUpper(t), "(")
			if !virtualFnWhitelist[fn] {
				return "", false
			}
			continue
		}
		if !virtualTokRe.MatchString(t) {
			return "", false
		}
	}
	return expr, true
}
