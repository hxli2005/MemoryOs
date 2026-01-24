# Contributing to MemoryOS

感谢你考虑为 MemoryOS 做出贡献！本文档提供了贡献指南。

## 开发环境设置

### 前置要求
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+（或使用 Docker）
- Milvus 2.3+（或使用 Docker）

### 快速开始
```bash
# 1. 克隆仓库
git clone https://github.com/your-org/MemoryOS.git
cd MemoryOS

# 2. 启动依赖服务
cd scripts/dev
./start_docker.bat  # Windows
# or
./start_docker.sh   # Linux/Mac

# 3. 配置环境
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml，填写 API Key

# 4. 运行测试
go test ./test/integration/...
```

## 代码规范

### Go 代码风格
- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 运行 `go vet` 检查潜在问题
- 使用 `golangci-lint` 进行静态分析

### 项目结构
```
MemoryOS/
├── cmd/           # 应用入口
├── internal/      # 内部包
│   ├── adapter/   # 外部适配器
│   ├── llm/       # LLM 客户端
│   ├── service/   # 业务逻辑
│   └── storage/   # 存储层
├── test/          # 测试文件
│   ├── integration/  # 集成测试
│   └── e2e/          # 端到端测试
└── examples/      # 示例代码
```

### 单元测试
- 测试文件命名：`*_test.go`
- 测试函数命名：`TestXxx(t *testing.T)`
- 覆盖率目标：> 70%

```go
// 示例：internal/llm/gemini_test.go
func TestGeminiClient_SummarizeDialogues(t *testing.T) {
    // Arrange
    client, _ := NewGeminiClient(cfg)
    
    // Act
    summary, err := client.SummarizeDialogues(ctx, dialogues)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, summary.Title)
}
```

### 集成测试
- 使用 `test/integration/` 目录
- 需要真实的数据库连接
- 使用 `config/test/integration.yaml` 配置

## 提交规范

### Commit Message 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型（type）**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具链

**示例**：
```
feat(llm): 添加 OpenAI LLM Client 支持

- 实现 SummarizeDialogues 方法
- 支持灵娅AI中转接口
- 添加自动路径补全（/v1）

Closes #42
```

## Pull Request 流程

1. **Fork 仓库** 并创建特性分支
   ```bash
   git checkout -b feat/your-feature
   ```

2. **编写代码** 并确保：
   - 通过所有测试：`go test ./...`
   - 代码格式化：`gofmt -w .`
   - 无 lint 错误：`golangci-lint run`

3. **提交更改**
   ```bash
   git add .
   git commit -m "feat(scope): description"
   ```

4. **推送分支**
   ```bash
   git push origin feat/your-feature
   ```

5. **创建 Pull Request**
   - 填写 PR 模板
   - 关联相关 Issue
   - 等待 Code Review

## Code Review 标准

PR 合并前需要：
- ✅ 至少 1 位 Maintainer 的批准
- ✅ 通过 CI/CD 检查
- ✅ 代码覆盖率不降低
- ✅ 文档已更新（如适用）

## 常见任务

### 添加新的 LLM Provider
1. 在 `internal/llm/` 创建 `provider.go`
2. 实现 `LLMClient` 接口
3. 在 `bootstrap/app.go` 注册 Provider
4. 添加单元测试 `provider_test.go`
5. 更新文档 `docs/guides/`

### 添加新的记忆类型
1. 在 `internal/model/memory.go` 添加常量
2. 更新 Schema（`docker/postgres/migrations/`）
3. 添加业务逻辑（`internal/service/memory/`）
4. 更新 API 文档

## 获取帮助

- 📖 查看 [docs/](../docs/) 目录
- 💬 在 Issue 中提问
- 📧 联系维护者：your-email@example.com

---

**再次感谢你的贡献！** 🎉
