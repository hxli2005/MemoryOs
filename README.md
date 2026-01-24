# 🧠 MemoryOS

> **基于 RAG 架构的 AI Agent 长期记忆系统**  
> 生产级实现 | Docker 一键部署 | QQ Bot 完整示例

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](docker-compose.yml)
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

### 方式一：Docker 部署（推荐）

```powershell
# 1. 克隆仓库
git clone https://github.com/yourusername/MemoryOS.git
cd MemoryOS

# 2. 配置 API Key（编辑 .env 文件）
LLM_API_KEY=your-api-key-here
EMBEDDING_API_KEY=your-api-key-here

# 3. 启动完整技术栈（PostgreSQL + Redis + Milvus）
docker-compose up -d

# 4. 查看日志
docker logs -f memoryos-app
```

访问 http://localhost:8080 验证 API 服务运行正常。

### 方式二：QQ Bot 示例

体验完整的长期记忆对话机器人：

```powershell
# 1. 配置 NapCat（参考 examples/qqbot/NAPCAT_SETUP.md）

# 2. 启动 QQ Bot
docker-compose -f docker-compose.qqbot.yaml up -d

# 3. 通过 QQ 私聊测试
# Bot 会记住你的对话历史、兴趣偏好和聊天风格
```

详见 [QQ Bot 使用指南](examples/qqbot/README.md)

### 方式三：本地开发

```bash
# 1. 配置环境
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml，填写 LLM API Key

# 2. 启动服务
go run cmd/server/main.go

# 3. 测试 API
curl http://localhost:8080/health
```

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
| 语言          | Go 1.24+                         | 核心开发语言              |
| AI 框架       | [Eino](https://github.com/cloudwego/eino) | LLM/Embedding 抽象层 |
| 向量数据库    | Milvus 2.3.3                     | 高性能向量检索            |
| 关系型数据库  | PostgreSQL + pgvector            | 元数据与向量存储          |
| 缓存          | Redis 7                          | 会话缓存                  |
| 消息队列      | NapCat WebSocket                 | QQ Bot 消息接入           |
| 容器编排      | Docker Compose                   | 一键部署全栈              |
| 监控方案      | Prometheus + Grafana（规划中）    | 指标采集与可视化          |

## 🎯 完整示例

### QQ Bot - 生产级长期记忆对话机器人

完整实现包括：
- ✅ NapCat WebSocket 消息收发
- ✅ Persona 配置化人设（4 个预设角色可切换）
- ✅ 三段式记忆召回（对话/主题/画像）
- ✅ 好感度系统与用户画像
- ✅ Docker 部署与数据持久化

详见 **[examples/qqbot/README.md](examples/qqbot/README.md)**

### 监控方案（规划中）

基于 Prometheus + Grafana 的可视化监控：
- 记忆操作 QPS 与召回耗时
- LLM API 调用统计与 Token 消耗
- 系统资源（Goroutine/内存/数据库连接池）
- 业务指标（活跃用户数/好感度分布）

可行性分析详见项目文档。

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
- [x] 真实 PostgreSQL Store 实现（含 pgvector 扩展）
- [x] 真实 Milvus Vector Store 实现
- [x] Docker Compose 编排（PostgreSQL + Redis + Milvus + etcd + MinIO）
- [x] QQ Bot 生产级示例（NapCat 集成）
- [x] Persona 配置化人设系统
- [x] 好感度与长期记忆召回

### Phase C（进行中 🚧）
- [x] Docker 一键部署
- [x] 数据持久化与备份
- [ ] 记忆压缩与老化（Decay）
- [ ] Prometheus + Grafana 监控
- [ ] pgvector VectorStore 实现（当前仅 Milvus）

### Phase D（规划中 📝）
- [ ] Web UI 管理后台（Vue/React）
- [ ] 多平台 Chatbot 适配器（微信/Telegram/Discord）
- [ ] 性能优化（缓存策略、批处理）
- [ ] 分布式部署文档

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
