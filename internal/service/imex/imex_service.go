package imex

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"github.com/sky-xhsoft/sky-server/pkg/errors"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Service 导入导出服务接口
type Service interface {
	// ExportToExcel 导出数据到Excel
	ExportToExcel(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) (string, error)

	// ImportFromExcel 从Excel导入数据
	ImportFromExcel(ctx context.Context, tableName string, file *multipart.FileHeader, userID uint) (*ImportResult, error)

	// GenerateTemplate 生成Excel导入模板
	GenerateTemplate(ctx context.Context, tableName string) (string, error)
}

// ImportResult 导入结果
type ImportResult struct {
	Total   int      `json:"total"`   // 总行数
	Success int      `json:"success"` // 成功导入数
	Failed  int      `json:"failed"`  // 失败数
	Errors  []string `json:"errors"`  // 错误信息
}

// service 导入导出服务实现
type service struct {
	db              *gorm.DB
	metadataService metadata.Service
	exportDir       string // 导出文件目录
}

// NewService 创建导入导出服务
func NewService(db *gorm.DB, metadataService metadata.Service, exportDir string) Service {
	if exportDir == "" {
		exportDir = "./exports"
	}
	return &service{
		db:              db,
		metadataService: metadataService,
		exportDir:       exportDir,
	}
}

// ExportToExcel 导出数据到Excel
func (s *service) ExportToExcel(ctx context.Context, tableName string, filters map[string]interface{}, userID uint) (string, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return "", err
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	f.SetSheetName(f.GetSheetName(0), sheetName)

	// 写入表头
	for i, col := range columns {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheetName, cell, col.DisplayName)
	}

	// 查询数据
	query := s.db.WithContext(ctx).Table(table.Name).Where("IS_ACTIVE = ?", "Y")

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		return "", errors.Wrap(errors.ErrDatabase, "查询数据失败", err)
	}

	// 写入数据
	for rowIdx, row := range results {
		for colIdx, col := range columns {
			cell := string(rune('A'+colIdx)) + strconv.Itoa(rowIdx+2)
			value := row[col.DbName]
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// 生成文件名
	filename := fmt.Sprintf("%s_%s.xlsx", tableName, time.Now().Format("20060102150405"))
	filepath := fmt.Sprintf("%s/%s", s.exportDir, filename)

	// 保存文件
	if err := f.SaveAs(filepath); err != nil {
		return "", errors.Wrap(errors.ErrInternal, "保存Excel文件失败", err)
	}

	return filepath, nil
}

// ImportFromExcel 从Excel导入数据
func (s *service) ImportFromExcel(ctx context.Context, tableName string, fileHeader *multipart.FileHeader, userID uint) (*ImportResult, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return nil, err
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "打开文件失败", err)
	}
	defer file.Close()

	// 读取Excel文件
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "解析Excel文件失败", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "读取Excel行失败", err)
	}

	if len(rows) == 0 {
		return nil, errors.New(errors.ErrInvalidParam, "Excel文件为空")
	}

	// 第一行是表头，建立列名映射
	header := rows[0]
	colMap := make(map[int]*entity.SysColumn)
	for i, headerName := range header {
		for _, col := range columns {
			if col.DisplayName == headerName || col.DbName == headerName {
				colMap[i] = col
				break
			}
		}
	}

	result := &ImportResult{
		Total:   len(rows) - 1,
		Success: 0,
		Failed:  0,
		Errors:  make([]string, 0),
	}

	// 逐行导入数据
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		data := make(map[string]interface{})

		// 解析行数据
		for colIdx, cellValue := range row {
			col, exists := colMap[colIdx]
			if !exists {
				continue
			}

			// 类型转换
			var value interface{}
			switch col.FieldType {
			case "int":
				if cellValue != "" {
					v, err := strconv.Atoi(cellValue)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("第%d行第%d列: 整数格式错误", rowIdx+1, colIdx+1))
						continue
					}
					value = v
				}
			case "decimal":
				if cellValue != "" {
					v, err := strconv.ParseFloat(cellValue, 64)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("第%d行第%d列: 小数格式错误", rowIdx+1, colIdx+1))
						continue
					}
					value = v
				}
			case "date", "datetime":
				if cellValue != "" {
					value = cellValue
				}
			default:
				value = cellValue
			}

			data[col.DbName] = value
		}

		// 添加审计字段
		data["CREATE_BY"] = fmt.Sprintf("user_%d", userID)
		data["UPDATE_BY"] = fmt.Sprintf("user_%d", userID)
		data["IS_ACTIVE"] = "Y"

		// 插入数据
		if err := s.db.WithContext(ctx).Table(table.Name).Create(data).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: 插入失败 - %s", rowIdx+1, err.Error()))
		} else {
			result.Success++
		}
	}

	return result, nil
}

// GenerateTemplate 生成Excel导入模板
func (s *service) GenerateTemplate(ctx context.Context, tableName string) (string, error) {
	// 获取表元数据
	table, err := s.metadataService.GetTable(tableName)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}

	// 获取字段定义
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return "", err
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	f.SetSheetName(f.GetSheetName(0), sheetName)

	// 写入表头
	for i, col := range columns {
		// 跳过系统字段
		if strings.HasPrefix(col.DbName, "CREATE_") ||
			strings.HasPrefix(col.DbName, "UPDATE_") ||
			col.DbName == "IS_ACTIVE" ||
			col.DbName == "ID" {
			continue
		}

		cell := string(rune('A'+i)) + "1"
		// 使用displayName作为表头，并添加注释
		f.SetCellValue(sheetName, cell, col.DisplayName)

		// 添加备注说明字段类型
		comment := fmt.Sprintf("字段: %s\n类型: %s\n", col.DbName, col.FieldType)
		if col.IsRequired {
			comment += "必填: 是\n"
		}
		if col.DefaultValue != "" {
			comment += fmt.Sprintf("默认值: %s\n", col.DefaultValue)
		}
		f.AddComment(sheetName, cell, comment)
	}

	// 生成文件名
	filename := fmt.Sprintf("%s_template.xlsx", tableName)
	filepath := fmt.Sprintf("%s/%s", s.exportDir, filename)

	// 保存文件
	if err := f.SaveAs(filepath); err != nil {
		return "", errors.Wrap(errors.ErrInternal, "保存模板文件失败", err)
	}

	return filepath, nil
}
