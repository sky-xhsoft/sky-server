package repository

import (
	"context"
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/pkg/test"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TestBaseRepositoryCRUD 测试基础Repository的CRUD功能
func TestBaseRepositoryCRUD(t *testing.T) {
	db := test.SetupGormTestDB(t)
	logger := test.NewTestLogger()
	repo := NewBaseRepository(db, logger, &NoopCache{}, time.Minute)

	test.AssertNotNil(t, repo, "BaseRepository should not be nil")

	// 测试简单查询功能
	ctx := context.Background()
	db.Exec("CREATE TABLE IF NOT EXISTS test_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT)")

	user := struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:200"`
		Email string `gorm:"size:200"`
	}{
		Name:  "Test User",
		Email: "test@example.com",
	}

	test.AssertNoError(t, repo.Create(ctx, &user), "Create failed")
	test.AssertNotNil(t, user.ID, "User should have ID after creation")

	var retrievedUser struct {
		ID    uint
		Name  string
		Email string
	}
	test.AssertNoError(t, repo.GetByID(ctx, &retrievedUser, user.ID), "GetByID failed")
	test.AssertEqual(t, user.ID, retrievedUser.ID, "ID mismatch")
	test.AssertEqual(t, user.Name, retrievedUser.Name, "Name mismatch")
	test.AssertEqual(t, user.Email, retrievedUser.Email, "Email mismatch")
}

// TestQueryBuilder 测试查询构建器功能
func TestQueryBuilder(t *testing.T) {
	db := test.SetupGormTestDB(t)

	// 创建测试表
	test.AssertNoError(t, db.Exec("CREATE TABLE IF NOT EXISTS test_products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, category TEXT, price REAL, quantity INTEGER)").Error, "Create test table failed")

	// 插入测试数据
	products := []struct {
		Name     string
		Category string
		Price    float64
		Quantity int
	}{
		{Name: "Product 1", Category: "Electronics", Price: 99.99, Quantity: 10},
		{Name: "Product 2", Category: "Electronics", Price: 199.99, Quantity: 5},
		{Name: "Product 3", Category: "Books", Price: 29.99, Quantity: 20},
		{Name: "Product 4", Category: "Books", Price: 19.99, Quantity: 30},
	}

	for _, p := range products {
		test.AssertNoError(t, db.Exec("INSERT INTO test_products (name, category, price, quantity) VALUES (?, ?, ?, ?)",
			p.Name, p.Category, p.Price, p.Quantity).Error, "Insert failed")
	}

	// 使用查询构建器查询
	qb := NewQueryBuilder(db).Table("test_products").Select("id", "name", "price").Where("category", "=", "Electronics").OrderBy("price", "asc")

	var results []struct {
		ID    uint
		Name  string
		Price float64
	}
	test.AssertNoError(t, qb.GetAll(&results), "Query failed")

	if len(results) == 0 {
		t.Fatal("Expected products to be returned")
	}

	t.Logf("Found %d products", len(results))
	for _, product := range results {
		t.Logf("- %s: $%.2f", product.Name, product.Price)
	}
}

// TestPageQuery 测试分页查询
func TestPageQuery(t *testing.T) {
	db := test.SetupGormTestDB(t)

	// 创建测试表
	test.AssertNoError(t, db.Exec("CREATE TABLE IF NOT EXISTS test_items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, value INTEGER)").Error, "Create test table failed")

	// 插入大量测试数据
	for i := 1; i <= 25; i++ {
		test.AssertNoError(t, db.Exec("INSERT INTO test_items (name, value) VALUES (?, ?)",
			ItemName(i), i*10).Error, "Insert failed")
	}

	// 测试分页查询
	repo := NewPageRepository(db)
	query := &PageQuery{
		Page:     2,
		PageSize: 10,
		OrderBy:  "value",
		Order:    "desc",
	}

	dest := []struct {
		ID    uint
		Name  string
		Value int
	}{}

	result, err := repo.GetPage(context.Background(), query, &dest)
	test.AssertNoError(t, err, "GetPage failed")

	test.AssertEqual(t, int(result.Total), 25, "Total count mismatch")
	test.AssertEqual(t, int(result.Page), 2, "Page number mismatch")
	test.AssertEqual(t, int(result.PageSize), 10, "Page size mismatch")
	test.AssertNotNil(t, result.Data, "Data should not be nil")

	t.Logf("Retrieved %d items from page %d (total %d)", len(dest), result.Page, result.Total)
}

func ItemName(i int) string {
	return "Item " + string('A'-1+i)
}

// TestNoopCache 测试空操作缓存
func TestNoopCache(t *testing.T) {
	cache := &NoopCache{}

	test.AssertNotNil(t, cache, "NoopCache should not be nil")

	var value interface{}
	err := cache.Get("test-key", &value)
	test.AssertError(t, err, "Get should return error")

	err = cache.Set("test-key", "value", time.Minute)
	test.AssertNoError(t, err, "Set failed")

	err = cache.Delete("test-key")
	test.AssertNoError(t, err, "Delete failed")

	exists, err := cache.Exists("test-key")
	test.AssertNoError(t, err, "Exists failed")
	test.AssertEqual(t, exists, false, "Key should not exist")
}

// TestLoggerInterface 测试Logger接口
func TestLoggerInterface(t *testing.T) {
	logger := &NoopLogger{}
	test.AssertNotNil(t, logger, "NoopLogger should not be nil")

	ctx := context.Background()
	logger.Debug(ctx, "Debug message")
	logger.Info(ctx, "Info message")
	logger.Warn(ctx, "Warn message")
	logger.Error(ctx, "Error message")
}

// TestTransaction 测试事务
func TestTransaction(t *testing.T) {
	db := test.SetupGormTestDB(t)

	logger := test.NewTestLogger()
	repo := NewBaseRepository(db, logger, &NoopCache{}, time.Minute)

	err := repo.Transaction(context.Background(), func(tx *gorm.DB) error {
		// 创建一个简单操作
		tx.Exec("CREATE TABLE IF NOT EXISTS transaction_test (id INTEGER PRIMARY KEY, value TEXT)")

		result := tx.Exec("INSERT INTO transaction_test (value) VALUES ('test1'), ('test2')")
		if result.Error != nil {
			return result.Error
		}

		var count int
		tx.Raw("SELECT COUNT(*) FROM transaction_test").Scan(&count)
		if count != 2 {
			tx.Rollback()
			return gorm.ErrInvalidData
		}

		return nil
	})

	test.AssertNoError(t, err, "Transaction failed")
}

// TestTransactionRollback 测试事务回滚
func TestTransactionRollback(t *testing.T) {
	db := test.SetupGormTestDB(t)

	logger := test.NewTestLogger()
	repo := NewBaseRepository(db, logger, &NoopCache{}, time.Minute)

	test.AssertNoError(t, db.Exec("CREATE TABLE IF NOT EXISTS rollback_test (id INTEGER PRIMARY KEY, value TEXT)").Error, "Create table failed")

	err := repo.Transaction(context.Background(), func(tx *gorm.DB) error {
		tx.Exec("INSERT INTO rollback_test (value) VALUES ('should-not-exist')")
		return gorm.ErrInvalidData
	})

	test.AssertError(t, err, "Transaction should fail")

	var count int
	test.AssertNoError(t, db.Raw("SELECT COUNT(*) FROM rollback_test").Scan(&count).Error, "Count failed")
	test.AssertEqual(t, count, 0, "No data should be present after rollback")
}

// TestMultipleQueries 测试多个查询
func TestMultipleQueries(t *testing.T) {
	db := test.SetupGormTestDB(t)

	test.AssertNoError(t, db.Exec("CREATE TABLE IF NOT EXISTS items (id INTEGER PRIMARY KEY, name TEXT, category TEXT, quantity INTEGER)").Error, "Create table failed")

	items := []struct {
		ID       uint
		Name     string
		Category string
		Quantity int
	}{
		{ID: 1, Name: "Item 1", Category: "Electronics", Quantity: 5},
		{ID: 2, Name: "Item 2", Category: "Electronics", Quantity: 10},
		{ID: 3, Name: "Item 3", Category: "Clothing", Quantity: 15},
		{ID: 4, Name: "Item 4", Category: "Clothing", Quantity: 20},
		{ID: 5, Name: "Item 5", Category: "Home", Quantity: 25},
	}

	for _, item := range items {
		test.AssertNoError(t, db.Exec("INSERT INTO items (id, name, category, quantity) VALUES (?, ?, ?, ?)",
			item.ID, item.Name, item.Category, item.Quantity).Error, "Insert failed")
	}

	qb := NewQueryBuilder(db).
		Table("items").
		Select("id", "name", "quantity").
		Where("category", "=", "Electronics").
		OrderBy("quantity", "desc")

	var electronicsItems []struct {
		ID       uint
		Name     string
		Quantity int
	}

	test.AssertNoError(t, qb.GetAll(&electronicsItems), "GetAll failed")

	test.AssertNotNil(t, electronicsItems, "Electronics items should not be nil")
	t.Logf("Found %d electronics items", len(electronicsItems))

	for _, item := range electronicsItems {
		test.AssertEqual(t, "Electronics", item.Category, "Category should be Electronics")
	}
}

// TestBaseRepositoryOptions 测试基础Repository选项
func TestBaseRepositoryOptions(t *testing.T) {
	db := test.SetupGormTestDB(t)

	logger, err := zap.NewProduction()
	test.AssertNoError(t, err, "Logger creation failed")

	repo1 := NewBaseRepository(db, logger, &NoopCache{}, 0)
	test.AssertNotNil(t, repo1, "BaseRepository with logger should not be nil")

	repo2 := NewBaseRepository(db, nil, &NoopCache{}, 0)
	test.AssertNotNil(t, repo2, "BaseRepository with nil logger should not be nil")

	repo3 := NewBaseRepository(db, nil, nil, 0)
	test.AssertNotNil(t, repo3, "BaseRepository with nil params should not be nil")
}
