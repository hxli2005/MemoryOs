# 🧠 MemoryOS

> **基于三层记忆架构的 AI Agent 长期记忆系统**  
> 专为 Chatbot 场景设计，对抗"人格与话题连续性的熵增"

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](docs/dev/CONTRIBUTING.md)

## ✨ 核心特性

### 🎯 三层记忆架构

```
Profile Layer (画像层)     → 用户是谁？长期认知
    ↑
Topic Layer (话题层)       → 聊过什么？跨会话记忆
    ↑
Dialogue Layer (对话层)    → 刚说什么？即时上下文
```

### 🔍 智能混合召回

根据对话阶段自动选择召回策略：
- **session_start**: 画像 70% + 话题 30%（会话开始，强化人格认知）
- **topic_deepening**: 话题 60% + 画像 30% + 对话 10%（深度讨论）
- **multi_turn**: 对话 50% + 话题 30% + 画像 20%（平衡三层记忆）

### 📊 11 种记忆类型

| Layer     | Memory Types                                      |
|-----------|--------------------------------------------------|
| Profile   | 用户身份、兴趣偏好、沟通风格、知识水平              |
| Topic     | 技术讨论、闲聊、问答、任务请求                      |
| Dialogue  | 用户消息、助手回复、系统事件                       |

## 🚀 快速开始

### 1️⃣ 配置 API Key

MemoryOS 支持 **Google Gemini**（推荐，免费）或 **OpenAI**：

#### 方式一：使用 Google Gemini（推荐）

```yaml
# 编辑 config/config.yaml
llm:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"  # 从 https://aistudio.google.com/app/apikey 获取
  model: "gemini-2.0-flash-exp"   # 免费且强大

embedder:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"
  model: "text-embedding-004"
  dimension: 768
```

> 💡 **为什么选择 Gemini？** 完全免费 + 性能优秀 + 无需信用卡

#### 方式二：使用 OpenAI

```yaml
# 编辑 config/config.yaml
llm:
  provider: "openai"
  api_key: "sk-YOUR_OPENAI_API_KEY"  # 从 https://platform.openai.com/api-keys 获取
  model: "gpt-4o-mini"

embedder:
  provider: "openai"
  api_key: "sk-YOUR_OPENAI_API_KEY"
  model: "text-embedding-3-small"
  dimension: 1536
```

详细配置指南：[Gemini 配置指南](docs/GEMINI_SETUP.md)

### 2️⃣ 启动 HTTP API 服务器

```bash
# 启动 API 服务器
go run cmd/server/main.go

# 服务将在 http://localhost:8080 启动
```

### 3️⃣ 测试 API

使用提供的测试文件快速验证功能：

```bash
# 查看 Web 测试界面
open test/index.html  # Mac/Linux
start test/index.html # Windows

# 或使用 REST Client 测试
# 在 VSCode 中打开 test/test_api.http
```

详见 [API 使用指南](docs/api/API_GUIDE.md)

## 🏗️ 项目架构

```
MemoryOs/
├── cmd/
│   └── server/          # HTTP API 服务器
│       └── main.go
├── internal/
│   ├── model/           # 数据模型（Memory, Layer, Type）
│   ├── service/memory/  # 核心业务逻辑（Manager）
│   ├── adapter/         # 外部服务适配器（Eino）
│   ├── handler/         # HTTP 处理器
│   ├── bootstrap/       # 应用初始化
│   └── mock/            # Mock 实现（开发用）

├── config/
│   ├── config.yaml      # 配置文件
│   └── config.example.yaml
├── docs/
│   ├── API_GUIDE.md     # API 文档
│   └── CHATBOT_USAGE.md # Chatbot 使用指南
└── test/
    ├── index.html       # Web 测试界面
    └── test_api.http    # REST Client 测试文件
```

## 📖 核心 API

### 创建记忆

```bash
POST /api/v1/memories
Content-Type: application/json

{
  "user_id": "user_alice",
  "layer": "profile",
  "type": "user_identity",
  "content": "我是一名软件工程师，正在准备校招",
  "metadata": {
    "category": "identity",
    "confidence_level": 0.9,
    "is_pinned": true
  }
}
```

### 混合召回

```bash
POST /api/v1/recall/hybrid
Content-Type: application/json

{
  "user_id": "user_alice",
  "session_id": "session_123",
  "query": "我在学习 Go 语言的并发",
  "dialog_stage": "multi_turn",
  "max_tokens": 4000
}
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "dialogue_memories": [...],
    "topic_memories": [...],
    "profile_memories": [...],
    "strategy": "multi_turn"
  }
}
```

完整 API 文档：[docs/API_GUIDE.md](docs/API_GUIDE.md)

## 🛠️ 技术栈

| 组件          | 技术选型                          | 用途                     |
|--------------|----------------------------------|--------------------------|
| 语言          | Go 1.23.3                        | 核心开发语言              |
| AI 框架       | [Eino](https://github.com/cloudwego/eino) | LLM/Embedding 抽象层 |
| Web 框架      | Gin                              | HTTP API                 |
| 向量数据库    | Milvus (planned)                 | 向量存储与检索            |
| 关系型数据库  | PostgreSQL + pgvector            | 元数据存储                |
| 缓存          | Redis                            | 热点画像缓存              |

## 🔧 开发模式

### Mock 模式（默认）

无需配置数据库，数据存在内存中（程序重启会丢失）。

适用场景：
- ✅ 快速体验
- ✅ 开发调试
- ✅ 单元测试

配置方式：
```yaml
database:
  host: ""  # 留空即启用 Mock 模式
```

### 生产模式

配置真实数据库，实现数据持久化。

配置方式：
```yaml
database:
  host: "localhost"  # 非空即启用 PostgreSQL
  port: 5432
  user: "postgres"
  password: "your_password"
  database: "memoryos"

redis:
  host: "localhost:6379"

vector_db:
  host: "localhost:19530"
```

> ⚠️ 注意：真实 Store 实现待补充（见 Roadmap）

## 📚 文档

- [Gemini 配置指南](docs/guides/GEMINI_SETUP.md) ⭐ **推荐阅读**
- [API 文档](docs/api/API_GUIDE.md)
- [开发指南](docs/dev/CONTRIBUTING.md)
- [架构设计](docs/ARCHITECTURE.md)（待补充）

## 🗺️ Roadmap

### Phase A（已完成 ✅）
- [x] 三层记忆架构设计
- [x] 核心 Manager 方法
- [x] HTTP API 接口
- [x] Mock 模式实现
- [x] 集成测试与验证

### Phase B（已完成 ✅）
- [x] 真实 PostgreSQL Store 实现
- [x] 真实 Milvus Vector Store 实现
- [x] LLM 驱动的话题聚合
- [x] LLM 驱动的画像提取

### Phase C（规划中 📝）
- [ ] 记忆压缩与老化（Decay）
- [ ] Web UI（Vue/React）
- [ ] 部署文档（Docker Compose）
- [ ] 性能优化与监控

## 💡 设计理念

### 对抗"人格与话题连续性的熵增"

传统 Chatbot 的问题：
- ❌ 对话越长，越容易"失忆"
- ❌ 跨会话无法记住用户
- ❌ 话题切换后无法回溯

MemoryOS 的解决方案：
- ✅ **Profile Layer**: 长期记住"用户是谁"（身份、偏好、风格）
- ✅ **Topic Layer**: 跨会话记住"聊过什么"（话题线索）
- ✅ **Dialogue Layer**: 即时记住"刚说什么"（上下文窗口）

### Layer-aware Importance

不同层级的记忆重要性不同：
- `Profile`: 1.0（最重要，决定人格）
- `Topic`: 0.8（重要，保证连续性）
- `Dialogue`: 0.6（普通，即时上下文）

### Adaptive Hybrid Recall

根据对话阶段动态调整召回策略：
```
session_start     → 强化人格认知（画像为主）
  ↓
multi_turn        → 平衡三层记忆
  ↓
topic_deepening   → 聚焦话题线索
```

## 📚 文档导航

- **快速开始**: 见上方 [快速开始](#-快速开始) 章节
- **API 文档**: [docs/api/API_GUIDE.md](docs/api/API_GUIDE.md)
- **开发指南**: [docs/dev/CONTRIBUTING.md](docs/dev/CONTRIBUTING.md)
- **项目结构**: [docs/dev/PROJECT_STRUCTURE.md](docs/dev/PROJECT_STRUCTURE.md)
- **更新日志**: [CHANGELOG.md](CHANGELOG.md)

## 🤝 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feat/amazing-feature`)
3. 遵循代码规范 (`gofmt`, `golangci-lint`)
4. 编写测试并确保通过 (`go test ./...`)
5. 提交更改 (`git commit -m 'feat: add amazing feature'`)
6. 推送到分支 (`git push origin feat/amazing-feature`)
7. 打开 Pull Request

详细贡献指南请查看 [CONTRIBUTING.md](docs/dev/CONTRIBUTING.md)

## 📄 License

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 📧 联系方式

- 项目目的：AI Agent 长期记忆系统研究与实现
- 技术栈：Go 1.21+ / Eino / RAG / PostgreSQL / Milvus
- 问题反馈：通过 GitHub Issues 提交

---

**让每一次对话都有记忆，让每一个 Agent 都有人格** 🚀
