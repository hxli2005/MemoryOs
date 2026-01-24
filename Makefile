.PHONY: help build run test clean docker-up docker-down

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@echo "🔨 编译 MemoryOS..."
	@go build -o bin/server cmd/server/main.go

run: ## 运行服务
	@echo "🚀 启动 MemoryOS..."
	@go run cmd/server/main.go

test: ## 运行测试
	@echo "🧪 运行测试..."
	@go test -v ./...

clean: ## 清理编译产物
	@echo "🧹 清理..."
	@rm -rf bin/

deps: ## 下载依赖
	@echo "📦 下载依赖..."
	@go mod download
	@go mod tidy

docker-up: ## 启动 Docker 环境 (PostgreSQL, Redis, Milvus)
	@echo "🐳 启动 Docker 容器..."
	@docker-compose up -d

docker-down: ## 停止 Docker 环境
	@echo "🛑 停止 Docker 容器..."
	@docker-compose down

lint: ## 代码检查
	@echo "🔍 代码检查..."
	@golangci-lint run
