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
	"github.com/sky-xhsoft/sky-server/internal/pkg/mask"
	"golang.org/x/crypto/bcrypt"
)

// 系统审计字段（无论MASK如何配置都应包含）
const (
	colID            = "ID"
	colSysCompanyID  = "SYS_COMPANY_ID"
	colCreateBy      = "CREATE_BY"
	colCreateTime    = "CREATE_TIME"
	colUpdateBy      = "UPDATE_BY"
	colUpdateTime    = "UPDATE_TIME"
	colIsActive      = "IS_ACTIVE"
	colStatus        = "STATUS"
	colOrderNo       = "ORDERNO"
)

var standardFields = map[string]bool{
	colID:            true,
	colSysCompanyID:  true,
	colCreateBy:      true,
	colCreateTime:    true,
	colUpdateBy:      true,
	colUpdateTime:    true,
	colIsActive:      true,
}

// buildSelectFields 构建查询字段列表（根据MASK控制）
func (s *service) buildSelectFields(columns []*entity.SysColumn, userID uint, operation string) (string, error) {
	var fields []string

	addedStandardFields := make(map[string]bool)

	for _, col := range columns {
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

	for _, col := range columns {
		value, exists := data[col.DbName]

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

		isAutoFill := col.SetValueType == "sysdate" || col.SetValueType == "createBy" || col.SetValueType == "operator" || col.SetValueType == "docno" || col.SetValueType == "pk"

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

		if !exists && col.DefaultValue != "" {
			value = col.DefaultValue
			exists = true
			data[col.DbName] = col.DefaultValue
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

	for _, col := range columns {
		value, exists := data[col.DbName]

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
