package crud

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/xuri/excelize/v2"
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

	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	resp, err := s.GetList(ctx, &QueryRequest{
		TableName: tableName,
		Page:      1,
		PageSize:  10000,
		Filters:   filters,
	}, userID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := "Sheet1"
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col.DisplayName)
	}

	for rowIdx, record := range resp.Data {
		for colIdx, col := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if val, ok := record[col.DbName]; ok {
				f.SetCellValue(sheet, cell, fmt.Sprintf("%v", val))
			}
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

	imported := 0
	for i, row := range rows {
		if i == 0 {
			continue
		}
		data := make(map[string]interface{})
		for j, cell := range row {
			if j < len(columns) {
				data[columns[j].DbName] = cell
			}
		}
		if _, err := s.Create(ctx, tableName, data, userID); err != nil {
			continue
		}
		imported++
	}

	return imported, nil
}
