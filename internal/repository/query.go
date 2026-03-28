package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	db         *gorm.DB
	table      string
	columns    []string
	conditions []Condition
	orderBys   []OrderBy
	groupBys   []string
	having     []Condition
	limit      int
	offset     int
	forUpdate  bool
}

// Condition 查询条件
type Condition struct {
	Column   string
	Operator string
	Value    interface{}
}

// OrderBy 排序条件
type OrderBy struct {
	Column    string
	Direction string // "asc" or "desc"
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:         db,
		columns:    []string{},
		conditions: []Condition{},
		orderBys:   []OrderBy{},
		groupBys:   []string{},
		having:     []Condition{},
		limit:      0,
		offset:     0,
	}
}

// Select 选择字段
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = append(qb.columns, columns...)
	return qb
}

// Table 设置表名
func (qb *QueryBuilder) Table(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Where 添加WHERE条件
func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// WhereIn 添加IN条件
func (qb *QueryBuilder) WhereIn(column string, values interface{}) *QueryBuilder {
	return qb.Where(column, "IN", values)
}

// WhereNotIn 添加NOT IN条件
func (qb *QueryBuilder) WhereNotIn(column string, values interface{}) *QueryBuilder {
	return qb.Where(column, "NOT IN", values)
}

// WhereLike 添加LIKE条件
func (qb *QueryBuilder) WhereLike(column string, value string) *QueryBuilder {
	return qb.Where(column, "LIKE", "%"+value+"%")
}

// WhereIsNull 添加IS NULL条件
func (qb *QueryBuilder) WhereIsNull(column string) *QueryBuilder {
	return qb.Where(column, "IS NULL", nil)
}

// WhereIsNotNull 添加IS NOT NULL条件
func (qb *QueryBuilder) WhereIsNotNull(column string) *QueryBuilder {
	return qb.Where(column, "IS NOT NULL", nil)
}

// WhereBetween 添加BETWEEN条件
func (qb *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder {
	return qb.Where(column, "BETWEEN", []interface{}{min, max})
}

// OrderBy 添加排序
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "asc"
	}
	direction = strings.ToLower(direction)
	if direction != "asc" && direction != "desc" {
		direction = "asc"
	}

	qb.orderBys = append(qb.orderBys, OrderBy{
		Column:    column,
		Direction: direction,
	})
	return qb
}

// GroupBy 添加分组
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBys = append(qb.groupBys, columns...)
	return qb
}

// Having 添加HAVING条件
func (qb *QueryBuilder) Having(column string, operator string, value interface{}) *QueryBuilder {
	qb.having = append(qb.having, Condition{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// Limit 设置LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset 设置OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Page 设置分页
func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	qb.limit = pageSize
	qb.offset = (page - 1) * pageSize
	return qb
}

// ForUpdate 设置FOR UPDATE
func (qb *QueryBuilder) ForUpdate() *QueryBuilder {
	qb.forUpdate = true
	return qb
}

// Build 构建查询
func (qb *QueryBuilder) Build() *gorm.DB {
	db := qb.db

	if qb.table != "" {
		db = db.Table(qb.table)
	}

	if len(qb.columns) > 0 {
		db = db.Select(qb.columns)
	}

	for _, cond := range qb.conditions {
		switch strings.ToUpper(cond.Operator) {
		case "IN":
			db = db.Where(cond.Column+" IN ?", cond.Value)
		case "NOT IN":
			db = db.Where(cond.Column+" NOT IN ?", cond.Value)
		case "LIKE":
			db = db.Where(cond.Column+" LIKE ?", cond.Value)
		case "IS NULL":
			db = db.Where(cond.Column + " IS NULL")
		case "IS NOT NULL":
			db = db.Where(cond.Column + " IS NOT NULL")
		case "BETWEEN":
			db = db.Where(cond.Column+" BETWEEN ? AND ?", cond.Value.([]interface{})[0], cond.Value.([]interface{})[1])
		default:
			db = db.Where(cond.Column+" "+cond.Operator+" ?", cond.Value)
		}
	}

	for _, orderBy := range qb.orderBys {
		db = db.Order(orderBy.Column + " " + orderBy.Direction)
	}

	if len(qb.groupBys) > 0 {
		db = db.Group(strings.Join(qb.groupBys, ", "))
	}

	for _, having := range qb.having {
		db = db.Having(having.Column+" "+having.Operator+" ?", having.Value)
	}

	if qb.limit > 0 {
		db = db.Limit(qb.limit)
	}

	if qb.offset > 0 {
		db = db.Offset(qb.offset)
	}

	if qb.forUpdate {
		db = db.Clauses(gorm.Locking{Strength: "UPDATE"})
	}

	return db
}

// GetOne 获取一条记录
func (qb *QueryBuilder) GetOne(dest interface{}) error {
	return qb.Build().First(dest).Error
}

// GetAll 获取所有记录
func (qb *QueryBuilder) GetAll(dest interface{}) error {
	return qb.Build().Find(dest).Error
}

// GetPage 获取分页记录
func (qb *QueryBuilder) GetPage(page, pageSize int, dest interface{}) (*PageResult, error) {
	qb.Page(page, pageSize)

	// 统计总数
	countQuery := qb.Build().Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// 查询数据
	if err := qb.Build().Find(dest).Error; err != nil {
		return nil, err
	}

	return CreatePageResult(dest, page, pageSize, total), nil
}

// Count 统计数量
func (qb *QueryBuilder) Count() (int64, error) {
	var total int64
	err := qb.Build().Count(&total).Error
	return total, err
}

// Exists 判断是否存在
func (qb *QueryBuilder) Exists() (bool, error) {
	count, err := qb.Count()
	return count > 0, err
}

// Create 创建记录
func (qb *QueryBuilder) Create(entity interface{}) error {
	return qb.db.Create(entity).Error
}

// Update 更新记录
func (qb *QueryBuilder) Update(entity interface{}) error {
	return qb.Build().Updates(entity).Error
}

// Delete 删除记录
func (qb *QueryBuilder) Delete(entity interface{}) error {
	return qb.Build().Delete(entity).Error
}

// FirstOrCreate 获取或创建
func (qb *QueryBuilder) FirstOrCreate(dest interface{}, conds ...interface{}) error {
	return qb.db.FirstOrCreate(dest, conds...).Error
}

// UpdateOrCreate 更新或创建
func (qb *QueryBuilder) UpdateOrCreate(dest interface{}, conds ...interface{}) error {
	return qb.db.Assign(dest).FirstOrCreate(dest, conds...).Error
}

// Exec 执行原生SQL
func (qb *QueryBuilder) Exec(sql string, values ...interface{}) error {
	return qb.db.Exec(sql, values...).Error
}

// Raw 执行原生SQL查询
func (qb *QueryBuilder) Raw(sql string, values ...interface{}) *gorm.DB {
	return qb.db.Raw(sql, values...)
}

// Scopes 应用查询作用域
func (qb *QueryBuilder) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *QueryBuilder {
	for _, f := range funcs {
		qb.db = f(qb.db)
	}
	return qb
}

// Clone 克隆查询构建器
func (qb *QueryBuilder) Clone() *QueryBuilder {
	newQB := *qb
	newQB.conditions = make([]Condition, len(qb.conditions))
	copy(newQB.conditions, qb.conditions)
	newQB.orderBys = make([]OrderBy, len(qb.orderBys))
	copy(newQB.orderBys, qb.orderBys)
	newQB.groupBys = make([]string, len(qb.groupBys))
	copy(newQB.groupBys, qb.groupBys)
	newQB.having = make([]Condition, len(qb.having))
	copy(newQB.having, qb.having)
	return &newQB
}

// Reset 重置查询构建器
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.columns = []string{}
	qb.conditions = []Condition{}
	qb.orderBys = []OrderBy{}
	qb.groupBys = []string{}
	qb.having = []Condition{}
	qb.limit = 0
	qb.offset = 0
	qb.forUpdate = false
	return qb
}

// WithContext 设置上下文
func (qb *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	qb.db = qb.db.WithContext(ctx)
	return qb
}

// GetDB 获取DB实例
func (qb *QueryBuilder) GetDB() *gorm.DB {
	return qb.db
}

// Preload 预加载关系
func (qb *QueryBuilder) Preload(query string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Preload(query, args...)
	return qb
}

// Joins 连接表
func (qb *QueryBuilder) Joins(query string, args ...interface{}) *QueryBuilder {
	qb.db = qb.db.Joins(query, args...)
	return qb
}

// Unscoped 忽略软删除
func (qb *QueryBuilder) Unscoped() *QueryBuilder {
	qb.db = qb.db.Unscoped()
	return qb
}

// WithDeleted 包含已删除的记录
func (qb *QueryBuilder) WithDeleted() *QueryBuilder {
	return qb.Unscoped()
}

// Pluck 提取单个字段
func (qb *QueryBuilder) Pluck(column string, dest interface{}) error {
	return qb.Build().Pluck(column, dest).Error
}

// Distinct 去重
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.db = qb.db.Distinct()
	return qb
}

// Debug 调试模式
func (qb *QueryBuilder) Debug() *QueryBuilder {
	qb.db = qb.db.Debug()
	return qb
}
