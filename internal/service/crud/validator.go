package crud

import (
	"context"
	"fmt"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// processForeignKeys 处理结果集中的外键字段，将ID转换为显示名称
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

	// 为每个FK字段收集所有需要查询的ID
	for _, fkCol := range fkColumns {
		// 收集所有唯一的ID值
		idSet := make(map[interface{}]bool)
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				idSet[val] = true
			}
		}

		if len(idSet) == 0 {
			continue
		}

		// 批量查询这些ID对应的显示值
		idToLabelMap := make(map[interface{}]string)
		for id := range idSet {
			// 转换为字符串进行查询
			idStr := fmt.Sprintf("%v", id)
			label, err := s.metadataService.GetForeignKeyDisplayValue(*fkCol.RefTableID, idStr, fkCol.RefColumnID, userID)
			if err != nil {
				// 单个FK查询失败不影响其他，使用原始值
				idToLabelMap[id] = idStr
			} else {
				idToLabelMap[id] = label
			}
		}

		// 获取关联表信息（用于跳转）
		refTable, err := s.metadataService.GetTableByID(*fkCol.RefTableID)
		var refTableName string
		if err == nil && refTable != nil {
			refTableName = refTable.Name
		}

		// 在结果集中添加 _display 和 _ref 字段
		displayFieldName := fkCol.DbName + "_display"
		refFieldName := fkCol.DbName + "_ref"
		for _, record := range results {
			if val, exists := record[fkCol.DbName]; exists && val != nil {
				// 添加显示值
				if label, ok := idToLabelMap[val]; ok {
					record[displayFieldName] = label
				}
				// 添加关联表信息（用于跳转）
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

// checkForeignKeyReferences 检查外键引用，防止删除被引用的记录
func (s *service) checkForeignKeyReferences(ctx context.Context, tableID uint, recordID uint) error {
	// 查询所有引用当前表的FK字段
	allTables, err := s.metadataRepo.GetAllTables()
	if err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询表列表失败", err)
	}

	fmt.Printf("[DEBUG] 检查外键引用：tableID=%d, recordID=%d, 总表数=%d\n", tableID, recordID, len(allTables))

	for _, refTable := range allTables {
		// 跳过当前表自己（不检查自己引用自己）
		if refTable.ID == tableID {
			fmt.Printf("[DEBUG] 跳过当前表自己: %s (ID=%d)\n", refTable.Name, refTable.ID)
			continue
		}

		columns, err := s.metadataService.GetColumns(refTable.ID)
		if err != nil {
			fmt.Printf("[DEBUG] 获取表 %s (ID=%d) 的列失败: %v\n", refTable.Name, refTable.ID, err)
			continue
		}

		// 检查引用表是否有 IS_ACTIVE 字段
		hasIsActive := false
		for _, col := range columns {
			if col.DbName == "IS_ACTIVE" {
				hasIsActive = true
				break
			}
		}

		for _, col := range columns {
			// 检查是否是引用当前表的FK字段
			if col.SetValueType == "fk" && col.RefTableID != nil && *col.RefTableID == tableID {
				fmt.Printf("[DEBUG] 发现FK字段：表=%s, 字段=%s, RefTableID=%d, RefOnDelete=%s\n",
					refTable.Name, col.DbName, *col.RefTableID, col.RefOnDelete)

				// 检查 REF_ON_DELETE 策略
				if col.RefOnDelete == "noAction" || col.RefOnDelete == "" {
					// 检查是否有记录引用
					var count int64
					var query string
					if hasIsActive {
						query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND IS_ACTIVE = 'Y'", refTable.Name, col.DbName)
					} else {
						query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", refTable.Name, col.DbName)
					}

					fmt.Printf("[DEBUG] 执行查询: %s, recordID=%d\n", query, recordID)

					if err := s.db.Raw(query, recordID).Scan(&count).Error; err != nil {
						fmt.Printf("[DEBUG] 查询失败: %v\n", err)
						return errors.Wrap(errors.ErrDatabase, "检查外键引用失败", err)
					}

					fmt.Printf("[DEBUG] 查询结果: count=%d\n", count)

					if count > 0 {
						return errors.New(errors.ErrResourceConflict, fmt.Sprintf("无法删除：有 %d 条 %s 记录正在引用", count, refTable.DisplayName))
					}
				}
				// cascade 和 setNull 策略暂不实现，需要在事务中处理
			}
		}
	}

	fmt.Printf("[DEBUG] 外键引用检查通过\n")
	return nil
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
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE ID = ? AND IS_ACTIVE = 'Y'", refTable.Name)
	if err := s.db.Raw(query, value).Scan(&count).Error; err != nil {
		return fmt.Errorf("验证失败: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("关联记录不存在或已失效")
	}

	return nil
}
