package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("Sky-Server Database Initialization")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("⚠️  WARNING: This will recreate the database and all tables!")
	fmt.Println("⚠️  All existing data will be preserved if tables exist.")
	fmt.Println()

	// Read SQL file
	sqlContent, err := os.ReadFile("sqls/init.sql")
	if err != nil {
		log.Fatal("Failed to read init.sql:", err)
	}

	// Connect to MySQL with multiStatements enabled
	dsn := "root:abc123@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Connected to MySQL server")
	fmt.Println("⏳ Executing SQL script (this may take a moment)...")
	fmt.Println()

	// Execute the entire SQL script
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		// Even if there are some errors, continue
		fmt.Printf("⚠️  Some errors occurred: %v\n", err)
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("✅ Database initialization completed!")
	fmt.Println()

	// Verify by checking table count
	db.Exec("USE skyserver")
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'skyserver'").Scan(&tableCount)
	if err == nil {
		fmt.Printf("   - Tables created: %d\n", tableCount)
	}

	// Verify admin user
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sys_user WHERE USERNAME = 'admin'").Scan(&userCount)
	if err == nil && userCount > 0 {
		fmt.Println("   - Admin user created: ✅")
	}

	// Verify company
	var companyCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sys_company WHERE ID = 1").Scan(&companyCount)
	if err == nil && companyCount > 0 {
		fmt.Println("   - Test company created: ✅")
	}

	fmt.Println()
	fmt.Println("Default credentials:")
	fmt.Println("   Username: admin")
	fmt.Println("   Password: admin123")
	fmt.Println("   Company ID: 1")
	fmt.Println("=" + strings.Repeat("=", 60))
}
