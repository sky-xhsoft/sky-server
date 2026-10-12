package crud

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"github.com/sky-xhsoft/sky-server/internal/service/idgen"
	"github.com/sky-xhsoft/sky-server/internal/service/metadata"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ---------- 测试桩：只实现用到的方法，其余由嵌入的空接口兜底（调用会 panic） ----------

type scopeMeta struct {
	metadata.Service
	tables map[string]*entity.SysTable
	cols   map[uint][]*entity.SysColumn
	refs   map[uint][]*entity.SysTableRef
}

func (m *scopeMeta) GetTable(name string) (*entity.SysTable, error) {
	if t, ok := m.tables[strings.ToUpper(name)]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("table %s not found", name)
}
func (m *scopeMeta) GetTableByID(id uint) (*entity.SysTable, error) {
	for _, t := range m.tables {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("table %d not found", id)
}
func (m *scopeMeta) GetColumns(id uint) ([]*entity.SysColumn, error) { return m.cols[id], nil }
func (m *scopeMeta) GetTableRefs(id uint) ([]*entity.SysTableRef, error) {
	return m.refs[id], nil
}

type scopeUsers struct {
	repository.UserRepository
	users map[uint]*entity.SysUser
}

func (u *scopeUsers) GetUserByID(id uint) (*entity.SysUser, error) {
	if x, ok := u.users[id]; ok {
		return x, nil
	}
	return nil, fmt.Errorf("user %d not found", id)
}

type scopeGroups struct {
	groups.Service
	deny    map[string]bool                 // "用户-表-权限位" 不授予
	filters map[uint]map[string]interface{} // 用户的数据过滤条件
}

func (g *scopeGroups) CheckUserTablePermission(_ context.Context, userID, tableID uint, perm int) (bool, error) {
	return !g.deny[fmt.Sprintf("%d-%d-%d", userID, tableID, perm)], nil
}
func (g *scopeGroups) GetUserDataFilter(_ context.Context, userID, _ uint) (map[string]interface{}, error) {
	return g.filters[userID], nil
}

type scopeRepo struct{ repository.MetadataRepository }

func (scopeRepo) GetTableCmdsByAction(uint, string, string) ([]*entity.SysTableCmd, error) {
	return nil, nil
}

type scopeIDGen struct {
	idgen.Service
	mu   sync.Mutex
	next uint
}

func (g *scopeIDGen) GetNextID(context.Context, string) (uint, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return g.next, nil
}

// ---------- 测试数据 ----------
// 公司：1 平台，46 甲公司，50 乙公司；0 为公共数据
// 用户：1 平台管理员，2 甲公司，3 乙公司，4 甲公司（只能看 DEPT = A）
const (
	uPlatform  uint = 1
	uA         uint = 2
	uB         uint = 3
	uAFiltered uint = 4

	tCompany  uint = 9
	tOrders   uint = 10
	tItem     uint = 11
	tCustomer uint = 12
	tLogs     uint = 13
	tConf     uint = 14
	tOther    uint = 15
)

func uptr(v uint) *uint { return &v }

func newScopeService(t *testing.T) (*service, *scopeGroups) {
	t.Helper()
	if _, err := logger.Init(&config.LogConfig{Level: "error"}); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	audit := `SYS_COMPANY_ID INTEGER, IS_ACTIVE TEXT DEFAULT 'Y', CREATE_BY TEXT, CREATE_TIME DATETIME, UPDATE_BY TEXT, UPDATE_TIME DATETIME`
	for _, stmt := range []string{
		`CREATE TABLE sys_user (ID INTEGER PRIMARY KEY, IS_ADMIN TEXT, SGRADE INTEGER)`,
		`CREATE TABLE sys_user_groups (ID INTEGER PRIMARY KEY, SYS_USER_ID INTEGER, SYS_DIRECTORY_ID INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_groups (ID INTEGER PRIMARY KEY, SGRADE INTEGER, IS_ACTIVE TEXT)`,
		`CREATE TABLE sys_directory (ID INTEGER PRIMARY KEY, SYS_TABLE_ID INTEGER, IS_ACTIVE TEXT)`,
		`INSERT INTO sys_directory VALUES (100, 10, 'Y')`,
		`CREATE TABLE SYS_COMPANY (ID INTEGER PRIMARY KEY, NAME TEXT, ` + audit + `)`,
		`INSERT INTO SYS_COMPANY (ID, NAME, SYS_COMPANY_ID) VALUES (1, '平台', 1), (46, '甲', 46), (50, '乙', 50)`,
		`CREATE TABLE ORDERS (ID INTEGER PRIMARY KEY, NAME TEXT, DEPT TEXT, CUSTOMER_ID INTEGER, ` + audit + `)`,
		`INSERT INTO ORDERS (ID, NAME, DEPT, SYS_COMPANY_ID) VALUES (1, '甲单A', 'A', 46), (2, '甲单B', 'B', 46), (3, '乙单', 'A', 50), (4, '公共', 'A', 0)`,
		`CREATE TABLE ORDER_ITEM (ID INTEGER PRIMARY KEY, ORDERS_ID INTEGER, QTY INTEGER, ` + audit + `)`,
		`INSERT INTO ORDER_ITEM (ID, ORDERS_ID, QTY, SYS_COMPANY_ID) VALUES (11, 1, 1, 46), (12, 1, 2, 46), (31, 3, 9, 50)`,
		`CREATE TABLE CUSTOMER (ID INTEGER PRIMARY KEY, NAME TEXT, ` + audit + `)`,
		`INSERT INTO CUSTOMER (ID, NAME, SYS_COMPANY_ID) VALUES (1, '甲客户', 46), (2, '乙客户', 50), (3, '公共客户', 0)`,
		// 物理表有公司字段、元数据没有登记（如 SYS_SCHEDULE / SYS_USER_PREF）
		`CREATE TABLE LOGS (ID INTEGER PRIMARY KEY, MSG TEXT, ` + audit + `)`,
		`INSERT INTO LOGS (ID, MSG, SYS_COMPANY_ID) VALUES (1, '甲', 46), (2, '乙', 50), (3, '公共', 0)`,
		`CREATE TABLE SYS_COMPANY_CONF (ID INTEGER PRIMARY KEY, SECRET_ID TEXT, ` + audit + `)`,
		`INSERT INTO SYS_COMPANY_CONF (ID, SECRET_ID, SYS_COMPANY_ID) VALUES (1, 'AKID-PLATFORM', 1), (46, 'AKID-A', 46), (50, 'AKID-B', 50)`,
		`CREATE TABLE OTHER (ID INTEGER PRIMARY KEY, X TEXT, ` + audit + `)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(stmt, err)
		}
	}

	all := "1111111111"
	col := func(id uint, name, svt, mask string, ref *uint) *entity.SysColumn {
		c := &entity.SysColumn{DbName: name, DisplayName: name, SetValueType: svt, Mask: mask, RefTableID: ref}
		c.ID = id
		return c
	}
	tbl := func(id uint, name string) *entity.SysTable {
		t := &entity.SysTable{Name: name, Mask: "AMDQ"}
		t.ID = id
		return t
	}
	meta := &scopeMeta{
		tables: map[string]*entity.SysTable{
			"SYS_COMPANY": tbl(tCompany, "SYS_COMPANY"), "ORDERS": tbl(tOrders, "ORDERS"), "ORDER_ITEM": tbl(tItem, "ORDER_ITEM"),
			"CUSTOMER": tbl(tCustomer, "CUSTOMER"), "LOGS": tbl(tLogs, "LOGS"), "SYS_COMPANY_CONF": tbl(tConf, "SYS_COMPANY_CONF"), "OTHER": tbl(tOther, "OTHER"),
		},
		cols: map[uint][]*entity.SysColumn{
			tCompany: {col(90, "ID", "pk", "", nil), col(91, "NAME", "byPage", all, nil), col(92, "SYS_COMPANY_ID", "object", "0000000000", nil)},
			// 与本地库一致：所属公司是外键、MASK 全 0
			tOrders: {col(100, "ID", "pk", "", nil), col(101, "NAME", "byPage", all, nil), col(102, "DEPT", "byPage", all, nil),
				col(103, "CUSTOMER_ID", "fk", all, uptr(tCustomer)), col(104, "SYS_COMPANY_ID", "fk", "0000000000", uptr(tCompany))},
			// 与本地库 12 张表一致：所属公司赋值方式 companyId、MASK 为空（默认可改）；指向主表的外键 MASK 全 0（如 SYS_DICT_ITEM.SYS_DICT_ID）
			tItem: {col(110, "ID", "pk", "", nil), col(111, "ORDERS_ID", "fk", "0000000000", uptr(tOrders)), col(112, "QTY", "byPage", all, nil),
				col(113, "SYS_COMPANY_ID", "companyId", "", nil)},
			tCustomer: {col(120, "ID", "pk", "", nil), col(121, "NAME", "byPage", all, nil), col(122, "SYS_COMPANY_ID", "fk", "0000000000", uptr(tCompany))},
			tLogs:     {col(130, "ID", "pk", "", nil), col(131, "MSG", "byPage", all, nil)},
			tConf:     {col(140, "ID", "pk", "", nil), col(141, "SECRET_ID", "byPage", all, nil), col(142, "SYS_COMPANY_ID", "object", "0000000000", nil)},
			tOther:    {col(150, "ID", "pk", "", nil), col(151, "X", "byPage", all, nil), col(152, "SYS_COMPANY_ID", "fk", "0000000000", uptr(tCompany))},
		},
		refs: map[uint][]*entity.SysTableRef{tCompany: {{SysTableID: int(tCompany), RefTableID: int(tConf), RefColumnID: 142, AssocType: "1"}}},
	}
	user := func(id, company uint, name string) *entity.SysUser {
		u := &entity.SysUser{Username: name, SysCompanyID: company}
		u.ID = id
		return u
	}
	g := &scopeGroups{deny: map[string]bool{}, filters: map[uint]map[string]interface{}{uAFiltered: {"DEPT": "A", "IS_ACTIVE": "Y"}}}
	s := &service{
		db:              db,
		metadataService: meta,
		groupsService:   g,
		metadataRepo:    scopeRepo{},
		userRepo: &scopeUsers{users: map[uint]*entity.SysUser{
			uPlatform: user(uPlatform, 1, "admin"), uA: user(uA, 46, "alice"), uB: user(uB, 50, "bob"), uAFiltered: user(uAFiltered, 46, "amy"),
		}},
		idgenService:    &scopeIDGen{next: 1000},
		companyColCache: newCompanyColCache(),
	}
	return s, g
}

func companyOf(t *testing.T, s *service, table string, id interface{}) int64 {
	t.Helper()
	var v []int64
	if err := s.db.Table(table).Where("ID = ?", id).Pluck("SYS_COMPANY_ID", &v).Error; err != nil || len(v) == 0 {
		t.Fatalf("%s#%v 不存在: %v", table, id, err)
	}
	return v[0]
}

func countRows(t *testing.T, s *service, table, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	if err := s.db.Table(table).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func wantCode(t *testing.T, err error, code int, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s：应失败，实际成功", what)
	}
	if got := errors.GetCode(err); got != code {
		t.Fatalf("%s：错误码应为 %d，实际 %d（%v）", what, code, got, err)
	}
}

// ---------- 新增：所属公司由服务端决定 ----------

func TestCreate_CompanyAssignedByServer(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()

	r, err := s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "跨公司", "SYS_COMPANY_ID": float64(50)}, uA)
	if err != nil {
		t.Fatal(err)
	}
	if got := companyOf(t, s, "ORDERS", r["ID"]); got != 46 {
		t.Errorf("甲公司用户传入乙公司ID，记录应建在甲公司，实际 %d", got)
	}
	if r["CREATE_BY"] != "alice" {
		t.Errorf("创建人应为当前用户, got %v", r["CREATE_BY"])
	}

	// 普通公司不能建公共数据
	r, err = s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "伪公共", "SYS_COMPANY_ID": "0"}, uA)
	if err != nil {
		t.Fatal(err)
	}
	if got := companyOf(t, s, "ORDERS", r["ID"]); got != 46 {
		t.Errorf("普通公司不能建公共数据, got %d", got)
	}

	// 平台可以建公共数据，但不能把普通表的记录建到其他公司
	r, _ = s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "公共", "SYS_COMPANY_ID": float64(0)}, uPlatform)
	if got := companyOf(t, s, "ORDERS", r["ID"]); got != 0 {
		t.Errorf("平台写 0 应建公共数据, got %d", got)
	}
	r, _ = s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "x", "SYS_COMPANY_ID": float64(50)}, uPlatform)
	if got := companyOf(t, s, "ORDERS", r["ID"]); got != 1 {
		t.Errorf("平台在普通表上也不能替其他公司建记录, got %d", got)
	}

	// 平台维护的表：平台可以为其他公司建记录；普通公司不行
	r, err = s.Create(ctx, "SYS_COMPANY_CONF", map[string]interface{}{"SECRET_ID": "p", "SYS_COMPANY_ID": float64(50)}, uPlatform)
	if err != nil {
		t.Fatal(err)
	}
	if got := companyOf(t, s, "SYS_COMPANY_CONF", r["ID"]); got != 50 {
		t.Errorf("平台在公司配置表上可指定公司, got %d", got)
	}
	if _, err := s.Create(ctx, "SYS_COMPANY_CONF", map[string]interface{}{"SECRET_ID": "p", "SYS_COMPANY_ID": float64(77)}, uPlatform); err == nil {
		t.Error("指定的公司不存在时应报错")
	}
	r, _ = s.Create(ctx, "SYS_COMPANY_CONF", map[string]interface{}{"SECRET_ID": "a", "SYS_COMPANY_ID": float64(50)}, uA)
	if got := companyOf(t, s, "SYS_COMPANY_CONF", r["ID"]); got != 46 {
		t.Errorf("普通公司不能替其他公司建配置, got %d", got)
	}
}

// ---------- 修改：所属公司不能改 ----------

func TestUpdate_CannotMoveCompany(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()

	wantCode(t, s.Update(ctx, "ORDERS", 1, map[string]interface{}{"SYS_COMPANY_ID": float64(50)}, uA), errors.ErrInvalidParam, "把记录改到其他公司")
	// 赋值方式 companyId、MASK 为空的表同样不能改
	wantCode(t, s.Update(ctx, "ORDER_ITEM", 11, map[string]interface{}{"QTY": 5, "SYS_COMPANY_ID": 50}, uA), errors.ErrInvalidParam, "明细改公司")
	if got := companyOf(t, s, "ORDER_ITEM", 11); got != 46 {
		t.Errorf("明细公司不应变化, got %d", got)
	}

	// 整条提交带原值：允许，其他字段正常保存
	if err := s.Update(ctx, "ORDERS", 1, map[string]interface{}{"NAME": "改名", "SYS_COMPANY_ID": "46"}, uA); err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, "ORDERS", "ID = 1 AND NAME = ? AND SYS_COMPANY_ID = 46", "改名") != 1 {
		t.Error("带原公司值的修改应成功且公司不变")
	}
}

// ---------- 写操作：公司范围、公共数据、数据过滤条件 ----------

func TestWrite_CompanyScope(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()

	wantCode(t, s.Update(ctx, "ORDERS", 3, map[string]interface{}{"NAME": "x"}, uA), errors.ErrResourceNotFound, "改其他公司的记录")
	wantCode(t, s.Delete(ctx, "ORDERS", 3, uA), errors.ErrResourceNotFound, "删其他公司的记录")
	wantCode(t, s.BatchDelete(ctx, "ORDERS", []uint{2, 3}, uA), errors.ErrResourceNotFound, "批量删除夹带其他公司的记录")
	if countRows(t, s, "ORDERS", "ID IN (2, 3)") != 2 {
		t.Error("批量删除失败时不应删除任何记录")
	}

	// 公共数据：普通公司只读，平台可改
	wantCode(t, s.Update(ctx, "ORDERS", 4, map[string]interface{}{"NAME": "x"}, uA), errors.ErrPermissionDenied, "普通公司改公共数据")
	wantCode(t, s.Delete(ctx, "ORDERS", 4, uA), errors.ErrPermissionDenied, "普通公司删公共数据")
	if err := s.Update(ctx, "ORDERS", 4, map[string]interface{}{"NAME": "平台改"}, uPlatform); err != nil {
		t.Errorf("平台应可改公共数据: %v", err)
	}

	// 公司配置：普通公司不能改别家的，平台可以
	wantCode(t, s.Update(ctx, "SYS_COMPANY_CONF", 50, map[string]interface{}{"SECRET_ID": "hack"}, uA), errors.ErrResourceNotFound, "改其他公司的配置")
	if err := s.Update(ctx, "SYS_COMPANY_CONF", 50, map[string]interface{}{"SECRET_ID": "平台改"}, uPlatform); err != nil {
		t.Errorf("平台应可维护各公司配置: %v", err)
	}

	if err := s.Delete(ctx, "ORDERS", 2, uA); err != nil {
		t.Errorf("删除本公司记录应成功: %v", err)
	}
}

func TestWrite_RespectsDataFilter(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()

	// 用户 4 只能看 DEPT = A：按 ID 也不能改、删 DEPT = B 的记录
	wantCode(t, s.Update(ctx, "ORDERS", 2, map[string]interface{}{"NAME": "x"}, uAFiltered), errors.ErrResourceNotFound, "改数据范围外的记录")
	wantCode(t, s.Delete(ctx, "ORDERS", 2, uAFiltered), errors.ErrResourceNotFound, "删数据范围外的记录")
	if err := s.Update(ctx, "ORDERS", 1, map[string]interface{}{"NAME": "范围内"}, uAFiltered); err != nil {
		t.Errorf("数据范围内的记录应可修改: %v", err)
	}
}

// ---------- 外键只能指向本公司或公共记录 ----------

func TestForeignKey_MustBeReadable(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()

	_, err := s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "x", "CUSTOMER_ID": 2}, uA)
	wantCode(t, err, errors.ErrInvalidParam, "引用其他公司的客户")
	wantCode(t, s.Update(ctx, "ORDERS", 1, map[string]interface{}{"CUSTOMER_ID": 2}, uA), errors.ErrInvalidParam, "修改为其他公司的客户")

	for _, id := range []int{1, 3} { // 本公司、公共
		if _, err := s.Create(ctx, "ORDERS", map[string]interface{}{"NAME": "x", "CUSTOMER_ID": id}, uA); err != nil {
			t.Errorf("引用客户 %d 应成功: %v", id, err)
		}
	}
}

// ---------- 读取：物理表上的公司字段、平台维护的表 ----------

func TestRead_CompanyScope(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()
	visible := func(table string, userID uint) []int64 {
		cols, _ := s.metadataService.GetColumns(s.metadataService.(*scopeMeta).tables[table].ID)
		var ids []int64
		s.applyCompanyScope(s.db.Table(table), table, cols, userID).Order("ID").Pluck("ID", &ids)
		return ids
	}

	// LOGS 元数据没有登记公司字段，但物理表有：仍按公司隔离
	if got := fmt.Sprint(visible("LOGS", uA)); got != "[1 3]" {
		t.Errorf("元数据漏登记公司字段的表也应隔离, got %s", got)
	}
	// 公司配置：普通公司只看自己的，平台看全部
	if got := fmt.Sprint(visible("SYS_COMPANY_CONF", uA)); got != "[46]" {
		t.Errorf("普通公司只能看自己的配置, got %s", got)
	}
	if got := fmt.Sprint(visible("SYS_COMPANY_CONF", uPlatform)); got != "[1 46 50]" {
		t.Errorf("平台可看各公司配置, got %s", got)
	}
	// 查不到当前用户：不返回任何数据
	if got := fmt.Sprint(visible("ORDERS", 999)); got != "[]" {
		t.Errorf("未知用户不应读到数据, got %s", got)
	}

	if _, err := s.GetOne(ctx, "SYS_COMPANY_CONF", 50, uA); err == nil {
		t.Error("GetOne 不应读到其他公司的配置")
	}
	if r, err := s.GetOne(ctx, "SYS_COMPANY_CONF", 46, uA); err != nil || r["SECRET_ID"] != "AKID-A" {
		t.Errorf("GetOne 应读到本公司配置: %v %v", r, err)
	}
}

// ---------- 主从保存 ----------

func TestSaveWithDetails_Create(t *testing.T) {
	s, _ := newScopeService(t)
	resp, err := s.SaveWithDetails(context.Background(), &SaveWithDetailsRequest{
		TableName: "ORDERS", Mode: "create",
		MainRecord: map[string]interface{}{"NAME": "主从", "SYS_COMPANY_ID": float64(50)},
		Details: []DetailTableRequest{{TableName: "ORDER_ITEM", RefField: "ORDERS_ID", Records: []map[string]interface{}{
			{"QTY": 1, "SYS_COMPANY_ID": 50}, {"QTY": 2, "ORDERS_ID": 3},
		}}},
	}, uA)
	if err != nil {
		t.Fatal(err)
	}
	mainID := resp.MainRecord["ID"]
	if got := companyOf(t, s, "ORDERS", mainID); got != 46 {
		t.Errorf("主表公司应为当前用户公司, got %d", got)
	}
	if n := countRows(t, s, "ORDER_ITEM", "ORDERS_ID = ? AND SYS_COMPANY_ID = 46 AND CREATE_BY = 'alice'", mainID); n != 2 {
		t.Errorf("两行明细都应挂在新主单、属于甲公司、有创建人, got %d", n)
	}
}

func TestSaveWithDetails_Update(t *testing.T) {
	s, _ := newScopeService(t)
	ctx := context.Background()
	_, err := s.SaveWithDetails(ctx, &SaveWithDetailsRequest{
		TableName: "ORDERS", Mode: "update", MainRecordID: 1,
		MainRecord: map[string]interface{}{"ID": 1, "NAME": "改主单"},
		Details: []DetailTableRequest{{TableName: "ORDER_ITEM", RefField: "ORDERS_ID", Records: []map[string]interface{}{
			{"ID": 11, "QTY": 7}, {"QTY": 3},
		}}},
	}, uA)
	if err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, "ORDER_ITEM", "ID = 11 AND QTY = 7") != 1 || countRows(t, s, "ORDER_ITEM", "ID = 12") != 0 || countRows(t, s, "ORDER_ITEM", "ORDERS_ID = 1") != 2 {
		t.Error("应更新第 11 行、删除未带上的第 12 行、新增一行")
	}
	if countRows(t, s, "ORDERS", "ID = 1 AND NAME = '改主单'") != 1 {
		t.Error("主单应更新")
	}
}

func TestSaveWithDetails_Rejects(t *testing.T) {
	ctx := context.Background()
	item := func(ref string, recs ...map[string]interface{}) []DetailTableRequest {
		return []DetailTableRequest{{TableName: "ORDER_ITEM", RefField: ref, Records: recs}}
	}
	cases := []struct {
		name  string
		req   SaveWithDetailsRequest
		user  uint
		deny  string
		code  int
		check func(t *testing.T, s *service)
	}{
		{name: "关联字段注入", user: uA, code: errors.ErrInvalidParam,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 1, Details: item("1=1 OR ORDERS_ID")},
			check: func(t *testing.T, s *service) {
				if countRows(t, s, "ORDER_ITEM", "1 = 1") != 3 {
					t.Error("不应删除任何明细")
				}
			}},
		{name: "关联字段不是指向主表的外键", user: uA, code: errors.ErrInvalidParam,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 1, Details: item("QTY")}},
		{name: "明细表不是子表", user: uA, code: errors.ErrInvalidParam,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "create", MainRecord: map[string]interface{}{"NAME": "x"},
				Details: []DetailTableRequest{{TableName: "OTHER", RefField: "SYS_COMPANY_ID", Records: []map[string]interface{}{{"X": "y"}}}}}},
		{name: "修改其他公司的主单", user: uA, code: errors.ErrResourceNotFound,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 3, MainRecord: map[string]interface{}{"NAME": "x"}}},
		{name: "修改公共主单", user: uA, code: errors.ErrPermissionDenied,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 4, MainRecord: map[string]interface{}{"NAME": "x"}}},
		{name: "明细不属于这张主单", user: uA, code: errors.ErrResourceNotFound,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 1, Details: item("ORDERS_ID",
				map[string]interface{}{"ID": 11}, map[string]interface{}{"ID": 12}, map[string]interface{}{"ID": 31, "QTY": 0})},
			check: func(t *testing.T, s *service) {
				if countRows(t, s, "ORDER_ITEM", "ID = 31 AND ORDERS_ID = 3 AND QTY = 9") != 1 {
					t.Error("其他单据的明细不应被改动")
				}
			}},
		{name: "没有主表新增权限", user: uA, deny: fmt.Sprintf("%d-%d-%d", uA, tOrders, groups.PermCreate), code: errors.ErrPermissionDenied,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "create", MainRecord: map[string]interface{}{"NAME": "x"}}},
		{name: "没有明细新增权限时主表也不保存", user: uA, deny: fmt.Sprintf("%d-%d-%d", uA, tItem, groups.PermCreate), code: errors.ErrPermissionDenied,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "create", MainRecord: map[string]interface{}{"NAME": "回滚"}, Details: item("ORDERS_ID", map[string]interface{}{"QTY": 1})},
			check: func(t *testing.T, s *service) {
				if countRows(t, s, "ORDERS", "NAME = '回滚'") != 0 {
					t.Error("事务应整体回滚")
				}
			}},
		{name: "没有明细删除权限", user: uA, deny: fmt.Sprintf("%d-%d-%d", uA, tItem, groups.PermDelete), code: errors.ErrPermissionDenied,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "update", MainRecordID: 1, Details: item("ORDERS_ID", map[string]interface{}{"ID": 11})}},
		{name: "未知模式", user: uA, code: errors.ErrInvalidParam,
			req: SaveWithDetailsRequest{TableName: "ORDERS", Mode: "upsert"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, g := newScopeService(t)
			if c.deny != "" {
				g.deny[c.deny] = true
			}
			_, err := s.SaveWithDetails(ctx, &c.req, c.user)
			wantCode(t, err, c.code, c.name)
			if c.check != nil {
				c.check(t, s)
			}
		})
	}
}

// 公司（主）→ 公司配置（1:1 明细，关联字段就是 SYS_COMPANY_ID）：本公司可以维护自己的配置
func TestSaveWithDetails_CompanyConf(t *testing.T) {
	s, _ := newScopeService(t)
	_, err := s.SaveWithDetails(context.Background(), &SaveWithDetailsRequest{
		TableName: "SYS_COMPANY", Mode: "update", MainRecordID: 46,
		MainRecord: map[string]interface{}{"ID": 46},
		Details:    []DetailTableRequest{{TableName: "SYS_COMPANY_CONF", RefField: "SYS_COMPANY_ID", Records: []map[string]interface{}{{"ID": 46, "SECRET_ID": "新"}}}},
	}, uA)
	if err != nil {
		t.Fatal(err)
	}
	if countRows(t, s, "SYS_COMPANY_CONF", "ID = 46 AND SECRET_ID = '新' AND SYS_COMPANY_ID = 46") != 1 {
		t.Error("本公司配置应更新")
	}
	// 别家公司：主表就不在写入范围内
	_, err = s.SaveWithDetails(context.Background(), &SaveWithDetailsRequest{
		TableName: "SYS_COMPANY", Mode: "update", MainRecordID: 50,
		Details: []DetailTableRequest{{TableName: "SYS_COMPANY_CONF", RefField: "SYS_COMPANY_ID", Records: []map[string]interface{}{{"ID": 50, "SECRET_ID": "hack"}}}},
	}, uA)
	wantCode(t, err, errors.ErrResourceNotFound, "改其他公司的配置")
}

// 导出：按每页 100 条分批取完（原先被截断为 100 条），且仍按公司隔离、不重复
func TestExport_AllPagesInScope(t *testing.T) {
	s, _ := newScopeService(t)
	for i := 0; i < 250; i++ {
		s.db.Exec(`INSERT INTO LOGS (MSG, SYS_COMPANY_ID) VALUES (?, 46), (?, 50)`, fmt.Sprintf("甲%d", i), fmt.Sprintf("乙%d", i))
	}
	f, err := s.Export(context.Background(), "LOGS", nil, uA)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	data := rows[2:] // 第 1 行表头、第 2 行字段名
	if len(data) != 252 {
		t.Fatalf("应导出本公司 251 条 + 公共 1 条 = 252 条, got %d", len(data))
	}
	seen := map[string]bool{}
	for _, r := range data {
		if seen[r[0]] {
			t.Fatalf("ID %s 重复导出", r[0])
		}
		seen[r[0]] = true
		if strings.HasPrefix(r[1], "乙") {
			t.Fatalf("导出了其他公司的数据: %v", r)
		}
	}
}
