# Sky-Server Makefile

.PHONY: help init build run test clean tidy swagger metadata-init

# 显示帮助信息
help:
	@echo "Sky-Server Makefile Commands:"
	@echo "  make init          - 初始化项目（安装依赖）"
	@echo "  make build         - 编译项目"
	@echo "  make run           - 运行项目"
	@echo "  make test          - 运行测试"
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
