# Sky-Server Makefile

.PHONY: help init build run test clean tidy swagger metadata-init test-cloud test-cloud-coverage test-cloud-integration test-cloud-bench test-all-cloud

# 显示帮助信息
help:
	@echo "Sky-Server Makefile Commands:"
	@echo "  make init          - 初始化项目（安装依赖）"
	@echo "  make build         - 编译项目"
	@echo "  make run           - 运行项目"
	@echo "  make test          - 运行所有测试"
	@echo "  make test-cloud    - 运行云盘服务单元测试"
	@echo "  make test-cloud-coverage  - 运行云盘服务测试并生成覆盖率报告"
	@echo "  make test-cloud-integration - 运行云盘服务集成测试"
	@echo "  make test-cloud-bench - 运行云盘服务性能基准测试"
	@echo "  make test-all-cloud - 运行所有云盘相关测试"
	@echo "  make clean         - 清理编译产物"
	@echo "  make tidy          - 整理依赖"
	@echo "  make swagger       - 生成Swagger文档"
	@echo "  make metadata-init - 从数据库初始化元数据"

# 初始化项目
init:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Dependencies installed successfully!"

# 编译项目
build:
	@echo "Building Sky-Server..."
	go build -o bin/sky-server cmd/server/main.go
	@echo "Build completed: bin/sky-server"

# 运行项目
run:
	@echo "Starting Sky-Server..."
	go run cmd/server/main.go

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 云盘服务测试
.PHONY: test-cloud
test-cloud: ## 运行云盘服务单元测试
	@echo "Running Cloud Service Tests..."
	go test -v -race ./internal/service/cloud/...

.PHONY: test-cloud-coverage
test-cloud-coverage: ## 运行云盘服务测试并生成覆盖率报告
	@echo "Running Cloud Service Tests with Coverage..."
	go test -v -coverprofile=coverage-cloud.out -covermode=atomic ./internal/service/cloud/...
	go tool cover -html=coverage-cloud.out -o coverage-cloud.html
	@echo "Coverage report: coverage-cloud.html"

.PHONY: test-cloud-integration
test-cloud-integration: ## 运行云盘服务集成测试
	@echo "Running Cloud Service Integration Tests..."
	go test -v -tags=integration ./internal/service/cloud/...

.PHONY: test-cloud-bench
test-cloud-bench: ## 运行云盘服务性能基准测试
	@echo "Running Cloud Service Benchmark Tests..."
	go test -v -bench=. -benchmem ./internal/service/cloud/...

.PHONY: test-all-cloud
test-all-cloud: test-cloud test-cloud-coverage test-cloud-bench ## 运行所有云盘相关测试
	@echo "All cloud service tests completed!"

# 清理编译产物
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	@echo "Clean completed!"

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied!"

# 生成Swagger文档
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o api/swagger
	@echo "Swagger documentation generated!"

# 从数据库初始化元数据
metadata-init:
	@echo "Initializing metadata from database..."
	go run cmd/metadata-init/main.go
	@echo "Metadata initialization completed!"
