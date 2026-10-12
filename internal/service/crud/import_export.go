package crud

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// Export 导出表数据到 Excel（支持过滤条件）
func (s *service) Export(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) (*excelize.File, error) {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermExport)
	if err != nil || !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无导出权限")
	}

	allColumns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}
	// 字段访问级别不足的列不导出（查询结果里本来就没有值，表头也不应出现）
	grade := s.userFieldGrade(userID)
	columns := make([]*entity.SysColumn, 0, len(allColumns))
	for _, col := range allColumns {
		if standardFields[col.DbName] || fieldAllowed(col.Sgrade, grade) {
			columns = append(columns, col)
		}
	}

	records, err := s.exportRecords(ctx, tableName, filters, userID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := "Sheet1"
	// SYS_COMPANY_ID 相关：预载公司 ID->名称 映射，供导出显示公司名（FK 关联翻译）
	companyNameMap := s.loadCompanyNameMap()
	// 第1行: 中文显示名(表头); 第2行: DB_NAME 内部映射行; 第3行起: 数据
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col.DisplayName)
		dbCell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, dbCell, col.DbName)
	}

	// 预载所有字典(DICT_ID -> {VALUE:DISPLAY_NAME ReverseDict})，供翻译
	dictVals, _ := s.loadDictMaps(columns)

	for rowIdx, record := range records {
		for colIdx, col := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+3)
			val, ok := record[col.DbName]
			if !ok {
				continue
			}
			// 外键列：优先使用 GetList 已翻译的 _display（显示名）
			if col.SetValueType == "fk" {
				if dp, ok := record[col.DbName+"_display"]; ok && dp != nil {
					f.SetCellValue(sheet, cell, fmt.Sprintf("%v", dp))
					continue
				}
			}
			// 字典列：VALUE -> DISPLAY_NAME
			if col.SetValueType == "select" && col.SysDictID != "" {
				if dm, ok := dictVals[col.SysDictID]; ok {
					if lbl, ok2 := dm[fmt.Sprintf("%v", val)]; ok2 {
						f.SetCellValue(sheet, cell, lbl)
						continue
					}
				}
			}
			// SYS_COMPANY_ID：按公司名显示（FK 关联翻译）
			if col.DbName == "SYS_COMPANY_ID" {
				if nm, ok := companyNameMap[fmt.Sprintf("%v", val)]; ok {
					f.SetCellValue(sheet, cell, nm)
					continue
				}
			}
			f.SetCellValue(sheet, cell, fmt.Sprintf("%v", val))
		}
	}

	return f, nil
}

// Import 从 Excel 导入数据
func (s *service) Import(ctx context.Context, tableName string, file multipart.File, userID uint) (int, error) {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return 0, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermImport)
	if err != nil || !hasPermission {
		return 0, errors.New(errors.ErrPermissionDenied, "无导入权限")
	}

	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return 0, err
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return 0, errors.New(errors.ErrInvalidParam, "无法解析Excel文件")
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return 0, errors.New(errors.ErrInvalidParam, "无法读取工作表数据")
	}
	if len(rows) < 2 {
		return 0, errors.New(errors.ErrInvalidParam, "缺少表头/字段映射行")
	}

	// 第1行：中文显示名；第2行：DB_NAME 内部映射（优先）；否则回退按位置
	// 建立 表头中文名 -> DbName 映射，以及 列位置 -> DbName 映射
	displayToDb := make(map[string]string)
	for _, col := range columns {
		if col.DisplayName != "" {
			displayToDb[col.DisplayName] = col.DbName
		}
	}

	// 决定每个 Excel 列对应的 DbName
	header := rows[0]
	var dbNameRow []string
	if len(rows) > 1 {
		dbNameRow = rows[1] // 第2行是 DB_NAME 映射行
	}

	colDbName := make([]string, len(header))
	hasDbNameRow := false
	for j := range header {
		if j < len(dbNameRow) && dbNameRow[j] != "" {
			// 第2行有值：当作 DB_NAME 直接使用
			colDbName[j] = dbNameRow[j]
			hasDbNameRow = true
		} else if j < len(header) {
			// 回退：用中文表头匹配 DbName
			colDbName[j] = displayToDb[header[j]]
		}
	}
	// 若没有任何一列从 DB_NAME 行解析出，且表头可全部映射，则识别为「仅中文表头」的模板，
	// 数据从第2行开始
	dataStartRow := 2
	if hasDbNameRow {
		dataStartRow = 3
	}

	imported := 0
	failed := 0
	var rowErrors []string // 失败行原因（最多返回 maxImportRowErrors 条）

	// 预载字典反向映射(DISPLAY_NAME -> VALUE)，供导入翻译
	_, dictRev := s.loadDictMaps(columns)
	// 预载 FK 反向映射：{dbName: {displayName: id}}，供导入将显示名转回ID
	fkRev := s.loadFkReverseMaps(columns, userID)
	// 预载公司反向映射：{公司名称: 公司ID}，供 SYS_COMPANY_ID FK 关联（显示名 -> ID）
	companyRev := s.loadCompanyReverseMap()
	logger.Info("import companyRev", zap.Int("size", len(companyRev)))

	for i := dataStartRow; i < len(rows); i++ {
		row := rows[i]
		// 跳过整行空白
		empty := true
		for _, cell := range row {
			if cell != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}

		data := make(map[string]interface{})
		for j, cell := range row {
			if j >= len(colDbName) {
				break
			}
			dbName := colDbName[j]
			if dbName == "" {
				continue
			}
			target := findColumnByDbName(columns, dbName)
			// 只导入「新增可见可改」的字段；系统/审计字段直接跳过
			if target == nil || !isImportEditableColumn(target) {
				continue
			}
			// 字典列：显示名 -> 原始 VALUE
			if cell != "" && target.SetValueType == "select" && target.SysDictID != "" {
				if dr, ok := dictRev[target.SysDictID]; ok {
					cell = s.dictLabelToValue(dr, cell)
				}
			}
			// FK 列：显示名 -> 关联ID；SYS_COMPANY_ID 也做 FK 关联（公司名称 -> 公司ID）
			if cell != "" && target.SetValueType == "fk" {
				if frm, ok := fkRev[dbName]; ok {
					if id, ok2 := frm[cell]; ok2 {
						cell = id
					}
				}
			}
			// SYS_COMPANY_ID 专用 FK 关联：填写公司名称/显示名 -> 公司ID（留空则 Create 兜底用户公司）
			if cell != "" && dbName == "SYS_COMPANY_ID" {
				if id, ok := companyRev[cell]; ok {
					cell = id
				} else if n, err := strconv.ParseInt(cell, 10, 64); err == nil {
					cell = fmt.Sprintf("%d", n)
				}
			}
			// 类型转换：数字列转 float64
			if cell != "" {
				if target.ColType != "" && isNumericType(target.ColType) {
					if n, convErr := strconv.ParseFloat(cell, 64); convErr == nil {
						data[dbName] = n
						continue
					}
				}
			}
			data[dbName] = cell
		}
		logger.Info("import row data", zap.Int("row", i+1), zap.Any("data", data))
		if _, err := s.Create(ctx, tableName, data, userID); err != nil {
			logger.Warn("import row failed", zap.String("table", tableName), zap.Int("row", i+1), zap.Any("data", data), zap.Error(err))
			failed++
			if len(rowErrors) < maxImportRowErrors {
				msg := err.Error()
				var appErr *errors.AppError
				if errors.As(err, &appErr) {
					msg = appErr.Message
				}
				rowErrors = append(rowErrors, fmt.Sprintf("第 %d 行：%s", i+1, msg))
			}
			continue
		}
		imported++
	}

	if failed > 0 {
		msg := fmt.Sprintf("导入完成：成功 %d 条，失败 %d 条；%s", imported, failed, strings.Join(rowErrors, "；"))
		if failed > len(rowErrors) {
			msg += "；……"
		}
		return imported, errors.New(errors.ErrInvalidParam, msg)
	}
	return imported, nil
}

// maxImportRowErrors 导入失败时最多列出的行数
const maxImportRowErrors = 10

// findColumnByDbName 按 DbName 查找列配置
func findColumnByDbName(columns []*entity.SysColumn, dbName string) *entity.SysColumn {
	for _, c := range columns {
		if c.DbName == dbName {
			return c
		}
	}
	return nil
}

// isNumericType 判断数据库列类型是否为数字
func isNumericType(colType string) bool {
	t := strings.ToLower(colType)
	switch {
	case strings.Contains(t, "int"), strings.Contains(t, "decimal"),
		strings.Contains(t, "float"), strings.Contains(t, "double"),
		strings.Contains(t, "numeric"), strings.Contains(t, "number"),
		strings.Contains(t, "bigint"), strings.Contains(t, "tinyint"),
		strings.Contains(t, "smallint"), strings.Contains(t, "mediumint"):
		return true
	}
	return false
}

// loadDictMaps 预载所有字典项映射，返回：
//
//	dictVals: DICT_ID -> {VALUE: DISPLAY_NAME}   （导出/翻译：值转显示名）
//	dictRev:  DICT_ID -> {DISPLAY_NAME: VALUE}   （导入：显示名转回值）
func (s *service) loadDictMaps(columns []*entity.SysColumn) (map[string]map[string]string, map[string]map[string]string) {
	dictVals := make(map[string]map[string]string)
	dictRev := make(map[string]map[string]string)
	seen := make(map[string]bool)
	for _, col := range columns {
		if col.SetValueType != "select" || col.SysDictID == "" {
			continue
		}
		dictID := col.SysDictID
		if seen[dictID] {
			continue
		}
		seen[dictID] = true
		var items []*entity.SysDictItem
		if err := s.db.Where("SYS_DICT_ID = ? AND (IS_ACTIVE IS NULL OR IS_ACTIVE='Y')", dictID).Find(&items).Error; err != nil {
			continue
		}
		fwd := make(map[string]string)
		rev := make(map[string]string)
		for _, it := range items {
			fwd[it.Value] = it.DisplayName
			rev[it.DisplayName] = it.Value
		}
		dictVals[dictID] = fwd
		dictRev[dictID] = rev
	}
	return dictVals, dictRev
}

// maxExportRows 单次导出上限
const maxExportRows = 10000

// exportRecords 按列表查询的分页上限（每页 100 条）分批取数，直到取完或达到导出上限；
// 原先一次请求 10000 条，被列表查询截断为 100 条，导出只有第一页
func (s *service) exportRecords(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) ([]map[string]interface{}, error) {
	const pageSize = 100
	var records []map[string]interface{}
	for page := 1; len(records) < maxExportRows; page++ {
		resp, err := s.GetList(ctx, &QueryRequest{
			TableName: tableName,
			Page:      page,
			PageSize:  pageSize,
			Filters:   filters,
		}, userID)
		if err != nil {
			return nil, err
		}
		records = append(records, resp.Data...)
		if len(resp.Data) < pageSize || int64(len(records)) >= resp.Total {
			break
		}
	}
	if len(records) > maxExportRows {
		records = records[:maxExportRows]
	}
	return records, nil
}

// ImportTemplate 生成导入模板（第1行中文表头、第2行 DB_NAME 映射、含 1 行示例空数据）
func (s *service) ImportTemplate(ctx context.Context, tableName string, userID uint) (*excelize.File, error) {
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}
	hasPermission, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, groups.PermImport)
	if err != nil || !hasPermission {
		return nil, errors.New(errors.ErrPermissionDenied, "无导入权限")
	}
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := "Sheet1"
	idx := 0
	for _, col := range columns {
		// 模板只列出「新增可见可改」的字段
		if !isImportEditableColumn(col) {
			continue
		}
		idx++
		cell, _ := excelize.CoordinatesToCellName(idx, 1)
		f.SetCellValue(sheet, cell, col.DisplayName)
		dbCell, _ := excelize.CoordinatesToCellName(idx, 2)
		f.SetCellValue(sheet, dbCell, col.DbName)
		// 示例行：FK/字典列填显示名提示，其余留空
		sample, _ := excelize.CoordinatesToCellName(idx, 3)
		if col.SetValueType == "fk" {
			f.SetCellValue(sheet, sample, fmt.Sprintf("(填写%s关联ID或显示名)", col.DisplayName))
		} else if col.SetValueType == "select" && col.SysDictID != "" {
			f.SetCellValue(sheet, sample, fmt.Sprintf("(字典: %s)", col.DisplayName))
		} else if col.DbName == "SYS_COMPANY_ID" {
			f.SetCellValue(sheet, sample, "(填写公司名称或ID，留空则归属当前用户)")
		}
	}
	return f, nil
}

// dictValueToKey 将 Excel 单元格(可能是显示名)转回原始 VALE；(用于导入字典列反向翻译)
// 若找不到映射则原样返回（可能是原始值）。
func (s *service) dictLabelToValue(dictRev map[string]string, cell string) string {
	if dictRev != nil {
		if v, ok := dictRev[cell]; ok {
			return v
		}
	}
	return cell
}

// loadFkReverseMaps 预载外键反向映射：{dbName: {显示名: ID}}
// 用于导入时把 Excel 中外键列的「显示名」转回关联记录 ID。
// 只取当前用户读取范围内（本公司或公共）的关联记录，名称不会匹配到其他公司的数据
func (s *service) loadFkReverseMaps(columns []*entity.SysColumn, userID uint) map[string]map[string]string {
	res := make(map[string]map[string]string)
	seen := make(map[string]bool)
	for _, col := range columns {
		if col.SetValueType != "fk" || col.RefTableID == nil || col.RefColumnID == nil {
			continue
		}
		key := fmt.Sprintf("%d-%d", *col.RefTableID, *col.RefColumnID)
		if seen[key] {
			continue
		}
		seen[key] = true
		refTable, err := s.metadataService.GetTableByID(*col.RefTableID)
		if err != nil || refTable == nil {
			continue
		}
		displayCol, err := s.metadataService.ResolveDisplayColumn(*col.RefTableID, col.RefColumnID)
		if err != nil {
			continue
		}
		refColumns, err := s.metadataService.GetColumns(refTable.ID)
		if err != nil {
			continue
		}
		var rows []map[string]interface{}
		q := s.applyCompanyScope(s.db.Table(refTable.Name).Select("ID, "+displayCol.DbName), refTable.Name, refColumns, userID)
		if err := q.Find(&rows).Error; err != nil {
			continue
		}
		m := make(map[string]string)
		for _, r := range rows {
			if id, ok := r["ID"]; ok {
				if lbl, ok2 := r[displayCol.DbName]; ok2 && lbl != nil {
					m[fmt.Sprintf("%v", lbl)] = fmt.Sprintf("%v", id)
				}
			}
		}
		res[col.DbName] = m
	}
	return res
}

// loadCompanyReverseMap 预载公司反向映射：{公司名称/NAME: 公司ID}
func (s *service) loadCompanyReverseMap() map[string]string {
	res := make(map[string]string)
	if t, err := s.metadataService.GetTable("SYS_COMPANY"); err != nil || t == nil {
		return res
	}
	var rows []map[string]interface{}
	if err := s.db.Table("SYS_COMPANY").Select("ID, NAME").Find(&rows).Error; err != nil {
		return res
	}
	for _, r := range rows {
		if id, ok := r["ID"]; ok {
			if lbl, ok2 := r["NAME"]; ok2 && lbl != nil {
				res[fmt.Sprintf("%v", lbl)] = fmt.Sprintf("%v", id)
			}
		}
	}
	return res
}

// loadCompanyNameMap 预载公司 ID->名称 映射，供导出时把 SYS_COMPANY_ID 翻译为公司名（FK 关联显示）
func (s *service) loadCompanyNameMap() map[string]string {
	res := make(map[string]string)
	var rows []map[string]interface{}
	if err := s.db.Table("SYS_COMPANY").Select("ID, NAME").Find(&rows).Error; err != nil {
		return res
	}
	for _, r := range rows {
		if id, ok := r["ID"]; ok {
			if lbl, ok2 := r["NAME"]; ok2 && lbl != nil {
				res[fmt.Sprintf("%v", id)] = fmt.Sprintf("%v", lbl)
			}
		}
	}
	return res
}

// isImportEditableColumn 判断某列在「新增导入」场景下是否可见且可修改。
// 复用了前端新增表单的编辑判定（MODIFI_ABLE!='N' + 排除系统赋值字段）。
// 排除主键、审计、以及由系统自动赋值的字段（pk/companyId/byPage/createBy/sysdate/operator/ignore/docno/select* 等）。
// SYS_COMPANY_ID 例外：入库若未提供由 Create 兜底用户公司，故不算“系统赋值”，可导入（走 FK 反翻）。
func isImportEditableColumn(col *entity.SysColumn) bool {
	if col == nil || col.DbName == "" {
		return false
	}
	db := col.DbName
	// 主键/审计/系统归属字段一律排除（SYS_COMPANY_ID 由 Create 兜底用户公司，模板不含）
	switch db {
	case "ID",
		"CREATE_BY", "CREATE_TIME",
		"UPDATE_BY", "UPDATE_TIME",
		"IS_ACTIVE", "SYS_COMPANY_ID":
		return false
	}
	// MODIFI_ABLE=N 不可改
	if col.ModifiAble == "N" {
		return false
	}
	// 系统自动赋值类型不可人工导入
	switch col.SetValueType {
	case "pk", "companyId", "createBy", "sysdate", "operator", "ignore", "docno", "docNo", "seq":
		return false
	}
	return true
}
