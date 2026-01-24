# Test 目录说明

本目录包含 MemoryOS 项目的所有测试文件。

## 目录结构

```
test/
├── integration/      # 集成测试（需要真实数据库）
│   ├── aggregation_test.go
│   ├── complete_test.go
│   ├── db_test.go
│   ├── embedding_test.go
│   ├── milvus_test.go
│   ├── memory_test.go
│   ├── simple_test.go
│   ├── debug_test.go
│   └── embedding_debug_test.go
└── e2e/              # 端到端测试
    ├── create_memory_test.go
    └── openai_api_test.go
```

## 测试分类

### 集成测试 (integration/)
测试多个组件协同工作，需要：
- PostgreSQL 数据库连接
- Milvus 向量数据库
- LLM API 配置

**运行方式**：
```bash
# 使用集成测试配置
go test ./test/integration/... -v

# 或使用配置文件
go test ./test/integration/... -config=config/test/integration.yaml
```

**关键测试**：
- `aggregation_test.go`: LLM 聚合功能（对话→话题→画像）
- `milvus_test.go`: 向量存储和检索
- `db_test.go`: PostgreSQL 元数据存储

### 端到端测试 (e2e/)
模拟真实用户场景，测试完整流程：
- `create_memory_test.go`: 记忆创建完整流程
- `openai_api_test.go`: OpenAI API 集成测试

## 配置文件

测试使用独立的配置文件，位于 `config/test/`：
- `mock.yaml`: Mock 模式（不需要真实服务）
- `integration.yaml`: 集成测试配置

## 最佳实践

### 1. 测试隔离
每个测试应该独立，不依赖其他测试的状态：
```go
func TestXxx(t *testing.T) {
    // Setup
    ctx := context.Background()
    app := setupTestApp(t)
    defer app.Shutdown()
    
    // Test logic
    // ...
    
    // Cleanup (if needed)
}
```

### 2. 使用 Subtests
对相关测试用例进行分组：
```go
func TestMemoryManager(t *testing.T) {
    t.Run("CreateMemory", func(t *testing.T) { ... })
    t.Run("GetMemory", func(t *testing.T) { ... })
    t.Run("UpdateMemory", func(t *testing.T) { ... })
}
```

### 3. 表驱动测试
对于多个相似场景：
```go
func TestValidation(t *testing.T) {
    tests := []struct{
        name string
        input string
        want error
    }{
        {"valid", "test", nil},
        {"empty", "", ErrEmpty},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Validate(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## 运行测试

### 运行所有测试
```bash
go test ./test/...
```

### 运行特定测试
```bash
go test ./test/integration/aggregation_test.go -v
```

### 查看覆盖率
```bash
go test ./test/... -cover
go test ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 调试测试

### 启用详细日志
```bash
go test ./test/integration/... -v -args -debug
```

### 使用 Delve 调试
```bash
dlv test ./test/integration/ -- -test.run TestAggregation
```

---

**注意**: 运行集成测试前，请确保 Docker 服务已启动（`scripts/dev/start_docker.bat`）。
