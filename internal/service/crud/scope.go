package crud

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"gorm.io/gorm"
)

// 公司隔离（多租户）规则，通用读写接口共用：
//   - 读：本公司数据 + 公共数据（SYS_COMPANY_ID 为 0 / -1）
//   - 写（修改、删除、提交、作废等）：只能写本公司数据，公共数据只有平台公司可以改；
//     同时要在权限组的数据过滤条件和表过滤条件之内（与查询一致）
//   - 新增：所属公司一律取登录用户的公司，请求里的值被忽略；
//     平台公司可以显式写 0 / -1 建公共数据，或在平台维护的表里为其他公司建记录
//   - 修改：所属公司不能改
//   - 平台维护的表（如 SYS_COMPANY_CONF 公司配置）：平台公司读写不受公司隔离，其他公司照常隔离
//
// 表是否有公司字段按元数据判断，元数据没有登记时再看物理表（防止漏登记的表绕过隔离）。

// platformCompanyID 平台公司（可使用系统中心、维护公共数据）
const platformCompanyID uint = 1

// platformManagedTables 平台跨公司维护的表
var platformManagedTables = map[string]bool{"SYS_COMPANY_CONF": true}

func isPlatformManaged(tableName string) bool {
	return platformManagedTables[strings.ToUpper(tableName)]
}

// userCompanyID 登录用户所属公司；查不到用户时返回 false（调用方按无权处理）
func (s *service) userCompanyID(userID uint) (uint, bool) {
	if s.userRepo == nil {
		return 0, false
	}
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return 0, false
	}
	return user.SysCompanyID, true
}

// hasCompanyColumn 表是否有 SYS_COMPANY_ID：先看元数据，没有登记时看物理表（结果缓存）
func (s *service) hasCompanyColumn(tableName string, columns []*entity.SysColumn) bool {
	for _, col := range columns {
		if col.DbName == colSysCompanyID {
			return true
		}
	}
	key := strings.ToUpper(tableName)
	if s.companyColCache != nil {
		if v, ok := s.companyColCache.Load(key); ok {
			return v.(bool)
		}
	}
	has := false
	if rows, err := s.db.Table(tableName).Select("*").Limit(0).Rows(); err == nil {
		if names, err := rows.Columns(); err == nil {
			for _, n := range names {
				if strings.EqualFold(n, colSysCompanyID) {
					has = true
					break
				}
			}
		}
		_ = rows.Close()
	}
	if s.companyColCache != nil {
		s.companyColCache.Store(key, has)
	}
	return has
}

// applyCompanyScope 读取范围：本公司与公共数据；查不到当前用户时不返回任何数据
func (s *service) applyCompanyScope(query *gorm.DB, tableName string, columns []*entity.SysColumn, userID uint) *gorm.DB {
	company, ok := s.userCompanyID(userID)
	if !ok {
		return query.Where("1 = 0")
	}
	if company == platformCompanyID && isPlatformManaged(tableName) {
		return query
	}
	if !s.hasCompanyColumn(tableName, columns) {
		return query
	}
	return query.Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1))", company)
}

// applyWriteScope 写入范围：本公司数据；平台公司还可以写公共数据
func (s *service) applyWriteScope(query *gorm.DB, tableName string, columns []*entity.SysColumn, company uint) *gorm.DB {
	if !s.hasCompanyColumn(tableName, columns) || (company == platformCompanyID && isPlatformManaged(tableName)) {
		return query
	}
	if company == platformCompanyID {
		return query.Where("(SYS_COMPANY_ID = ? OR SYS_COMPANY_ID IN (0, -1))", company)
	}
	return query.Where("SYS_COMPANY_ID = ?", company)
}

// ensureWritable 写操作前确认目标记录都在当前用户的写入范围内：公司范围、表过滤条件、权限组数据过滤条件。
// 写操作只按 ID 定位记录，缺少这一步会让有权限的用户改动范围外（其他公司、其他部门）的数据。
func (s *service) ensureWritable(ctx context.Context, table *entity.SysTable, columns []*entity.SysColumn, userID uint, ids ...uint) error {
	uniq := make(map[uint]bool, len(ids))
	list := make([]uint, 0, len(ids))
	for _, id := range ids {
		if !uniq[id] {
			uniq[id] = true
			list = append(list, id)
		}
	}
	if len(list) == 0 {
		return nil
	}
	company, ok := s.userCompanyID(userID)
	if !ok {
		return errors.New(errors.ErrPermissionDenied, "无法确定当前用户所属公司")
	}
	q := s.applyWriteScope(s.db.Table(table.Name).Where("ID IN ?", list), table.Name, columns, company)
	if table.Filter != "" {
		q = s.applyDataFilter(q, table.Filter, columns)
	}
	filter, err := s.userTableDataFilter(ctx, userID, table.ID)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "获取数据过滤条件失败", err)
	}
	if len(filter) > 0 {
		f := make(map[string]interface{}, len(filter))
		for k, v := range filter {
			if k != colIsActive { // 与 GetOne 一致：停用的记录也可以维护
				f[k] = v
			}
		}
		if len(f) > 0 {
			q = s.applyFilters(q, f, columns)
		}
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}
	if n == int64(len(list)) {
		return nil
	}
	// 区分「公共数据不能改」和「记录不存在」，后者不透露范围外记录是否存在
	if company != platformCompanyID && s.hasCompanyColumn(table.Name, columns) {
		var shared int64
		if err := s.db.Table(table.Name).Where("ID IN ? AND SYS_COMPANY_ID IN (0, -1)", list).Count(&shared).Error; err == nil && shared > 0 {
			return errors.New(errors.ErrPermissionDenied, "公共数据只有平台可以修改")
		}
	}
	return errors.New(errors.ErrResourceNotFound, "记录不存在")
}

// ensureWritableByID 同 ensureWritable，字段定义由这里加载
func (s *service) ensureWritableByID(ctx context.Context, table *entity.SysTable, userID uint, ids ...uint) error {
	columns, err := s.metadataService.GetColumns(table.ID)
	if err != nil {
		return err
	}
	return s.ensureWritable(ctx, table, columns, userID, ids...)
}

// parseCompanyValue 把请求里的公司值（数字、字符串）转成整数；空值返回 false
func parseCompanyValue(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	str := strings.TrimSpace(fmt.Sprint(v))
	if str == "" {
		return 0, false
	}
	if n, err := strconv.ParseInt(str, 10, 64); err == nil {
		return n, true
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil && f == float64(int64(f)) {
		return int64(f), true
	}
	return 0, false
}

// createCompanyID 新增记录的所属公司（见文件头规则）
func createCompanyID(tableName string, requested interface{}, own uint) interface{} {
	if own == platformCompanyID {
		if n, ok := parseCompanyValue(requested); ok {
			if n == 0 || n == -1 || (n > 0 && isPlatformManaged(tableName)) {
				return n
			}
		}
	}
	return own
}

// checkCompanyUnchanged 修改时提交的所属公司必须与当前值一致（整条提交时会带上原值）
func (s *service) checkCompanyUnchanged(tableName string, id uint, requested interface{}) error {
	want, ok := parseCompanyValue(requested)
	if !ok {
		return nil
	}
	var cur []int64
	if err := s.db.Table(tableName).Where("ID = ?", id).Limit(1).Pluck(colSysCompanyID, &cur).Error; err != nil {
		return errors.Wrap(errors.ErrDatabase, "查询记录失败", err)
	}
	if len(cur) > 0 && cur[0] != want {
		return errors.New(errors.ErrInvalidParam, "所属公司不能修改")
	}
	return nil
}

// ensureRefReadable 外键只能指向当前用户读取范围内（本公司或公共）的记录
func (s *service) ensureRefReadable(refTable *entity.SysTable, value interface{}, userID uint) error {
	columns, err := s.metadataService.GetColumns(refTable.ID)
	if err != nil {
		return err
	}
	var n int64
	q := s.applyCompanyScope(s.db.Table(refTable.Name).Where("ID = ?", value), refTable.Name, columns, userID)
	if err := q.Count(&n).Error; err != nil {
		return fmt.Errorf("验证失败: %v", err)
	}
	if n == 0 {
		return fmt.Errorf("关联记录不存在或已失效")
	}
	return nil
}

// requirePermission 表级权限检查
func (s *service) requirePermission(ctx context.Context, userID uint, table *entity.SysTable, perm int, msg string) error {
	ok, err := s.groupsService.CheckUserTablePermission(ctx, userID, table.ID, perm)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "权限检查失败", err)
	}
	if !ok {
		return errors.New(errors.ErrPermissionDenied, msg)
	}
	return nil
}

// newCompanyColCache 物理表公司字段缓存（表结构变更需重启生效，与元数据缓存一致）
func newCompanyColCache() *sync.Map { return &sync.Map{} }
