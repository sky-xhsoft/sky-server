package crud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/sqlutils"
)

// processForeignKeys 处理结果集中的外键字段，将ID转换为显示名称（批量查询版本）
func (s *service) processForeignKeys(ctx context.Context, columns []*entity.SysColumn, results []map[string]interface{}, userID uint) error {
	if len(results) == 0 {
		return nil
	}

	// 找出所有FK字段
	fkColumns := make([]*entity.SysColumn, 0)
	for _, col := range columns {
		if col.SetValueType == "fk" && col.RefTableID != nil {
			fkColumns = append(fkColumns, col)
		}
	}

	if len(fkColumns) == 0 {
		return nil
	}

	// 按 (refTableID, refColumnID) 分组收集所有FK值
	type fkBatchKey struct {
		RefTableID  uint
		RefColumnID *uint
	}
	batchGroups := make(map[fkBatchKey]map[interface{}]bool)

	for _, fkCol := range fkColumns {
		key := fkBatchKey{RefTableID: *fkCol.RefTableID, RefColumnID: fkCol.RefColumnID}
		if batchGroups[key] == nil {
			batchGroups[key] = make(map[interface{}]bool)
		}
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				batchGroups[key][val] = true
			}
		}
	}

	// 批量查询每个分组
	batchResults := make(map[fkBatchKey]map[interface{}]string)
	for key, idSet := range batchGroups {
		if len(idSet) == 0 {
			continue
		}

		ids := make([]interface{}, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}

		// 获取关联表元数据
		refTable, err := s.metadataService.GetTableByID(key.RefTableID)
		if err != nil || refTable == nil {
			continue
		}
		refColumns, err := s.metadataService.GetColumns(key.RefTableID)
		if err != nil {
			continue
		}

		// 确定显示列
		var displayCol *entity.SysColumn
		if key.RefColumnID != nil {
			for _, col := range refColumns {
				if col.ID == *key.RefColumnID {
					displayCol = col
					break
				}
			}
		}
		if displayCol == nil {
			for _, col := range refColumns {
				if col.IsDK == "Y" {
					displayCol = col
					break
				}
			}
		}
		if displayCol == nil {
			for _, col := range refColumns {
				if col.ColType == "varchar" || col.ColType == "char" {
					displayCol = col
					break
				}
			}
		}
		if displayCol == nil {
			continue
		}

		// 批量查询显示值
		var rows []map[string]interface{}
		s.db.Table(refTable.Name).Select(fmt.Sprintf("ID, %s", displayCol.DbName)).Where("ID IN ?", ids).Find(&rows)

		labelMap := make(map[interface{}]string)
		for _, row := range rows {
			if id, ok := row["ID"]; ok {
				if label, ok := row[displayCol.DbName]; ok {
					labelMap[id] = fmt.Sprintf("%v", label)
				}
			}
		}
		batchResults[key] = labelMap
	}

	// 填充结果
	for _, fkCol := range fkColumns {
		key := fkBatchKey{RefTableID: *fkCol.RefTableID, RefColumnID: fkCol.RefColumnID}
		labelMap := batchResults[key]

		refTable, err := s.metadataService.GetTableByID(*fkCol.RefTableID)
		var refTableName string
		if err == nil && refTable != nil {
			refTableName = refTable.Name
		}

		displayFieldName := fkCol.DbName + "_display"
		refFieldName := fkCol.DbName + "_ref"
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				if labelMap != nil {
					if label, ok := labelMap[val]; ok {
						record[displayFieldName] = label
					}
				}
				if refTableName != "" {
					record[refFieldName] = map[string]interface{}{
						"table_name": refTableName,
						"table_id":   *fkCol.RefTableID,
						"record_id":  val,
					}
				}
			}
		}
	}

	return nil
}

// fkRef 外键引用信息
type fkRef struct {
	TableName   string
	ColName     string
	RefOnDelete string
}

// checkForeignKeyReferences 检查外键引用，返回需要级联处理的FK引用列表
// restrict策略：直接阻断删除
// cascade/setNull策略：返回引用列表由调用方在事务中处理
func (s *service) checkForeignKeyReferences(ctx context.Context, tableID uint, recordID uint) ([]fkRef, error) {
	var fkRefs []fkRef
	s.db.Raw(`
		SELECT t.NAME as table_name, c.DB_NAME as col_name, COALESCE(c.REF_ON_DELETE, '') as ref_on_delete
		FROM sys_column c
		JOIN sys_table t ON c.SYS_TABLE_ID = t.ID
		WHERE c.SET_VALUE_TYPE = 'fk' AND c.REF_TABLE_ID = ?
		AND t.ID != ?
	`, tableID, tableID).Scan(&fkRefs)

	for _, ref := range fkRefs {
		safeCol, _ := sqlutils.SanitizeIdentifier(ref.ColName)
		if safeCol == "" {
			continue
		}

		if ref.RefOnDelete == "noAction" || ref.RefOnDelete == "" {
			var count int64
			s.db.Table(ref.TableName).Where(fmt.Sprintf("`%s` = ? AND IS_ACTIVE = 'Y'", safeCol), recordID).Count(&count)
			if count > 0 {
				return nil, errors.New(errors.ErrResourceConflict, fmt.Sprintf("无法删除：有 %d 条 %s 记录正在引用", count, ref.TableName))
			}
		}
	}

	return fkRefs, nil
}

// validateForeignKeyValue 验证FK字段的值是否在关联表中存在
func (s *service) validateForeignKeyValue(refTableID uint, value interface{}) error {
	// 获取关联表信息
	refTable, err := s.metadataService.GetTableByID(refTableID)
	if err != nil {
		return fmt.Errorf("关联表不存在")
	}

	// 查询记录是否存在
	var count int64
	if err := s.db.Table(refTable.Name).Where("ID = ? AND IS_ACTIVE = 'Y'", value).Count(&count).Error; err != nil {
		return fmt.Errorf("验证失败: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("关联记录不存在或已失效")
	}

	return nil
}
