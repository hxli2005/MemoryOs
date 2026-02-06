# MemoryOS 项目结构总览

> 最后更新: 2026-02-06

## 📋 目录

- [项目简介](#项目简介)
- [目录结构](#目录结构)
- [核心架构](#核心架构)
- [快速导航](#快速导航)
- [开发指南](#开发指南)

---

## 🎯 项目简介

**MemoryOS** 是一个基于 RAG (Retrieval-Augmented Generation) 架构的 AI Agent 长期记忆系统，使用 Go 语言开发。

### 核心特性

- ✅ **混合记忆召回**：结合对话、主题、用户画像三层记忆
- ✅ **向量检索**：支持 Milvus/pgvector 向量数据库
- ✅ **LLM 集成**：支持 OpenAI/Gemini API（含国内中转）
- ✅ **完整监控**：Prometheus + Grafana 可观测性体系
- ✅ **云端部署**：Docker Compose + CI/CD 自动化部署
- ✅ **成本优化**：Embedding 批处理、限流、降维策略

---

## 📂 目录结构

```
MemoryOs/
├── .github/                    # GitHub 配置
│   ├── workflows/              
│   │   └── deploy.yml          # CI/CD 流水线
│   └── copilot-instructions.md # AI 助手指南
│
├── cmd/                        # 服务入口
│   └── server/
│       └── main.go             # HTTP 服务主程序
│
├── internal/                   # 内部核心代码（不可外部引用）
│   ├── adapter/                # 外部适配器层
│   │   └── eino.go             # Embedding 服务适配
│   ├── bootstrap/              # 应用启动
│   │   └── app.go              # 依赖注入 & 初始化
│   ├── config/                 # 配置管理
│   │   └── config.go           # 配置结构体 & 加载
│   ├── handler/                # HTTP 处理层（Controller）
│   │   └── memory.go           # 记忆相关 API
│   ├── llm/                    # LLM 抽象层
│   │   ├── interface.go        # LLM 接口定义
│   │   ├── openai.go           # OpenAI 实现
│   │   └── gemini.go           # Gemini 实现
│   ├── metrics/                # 监控指标
│   │   └── metrics.go          # Prometheus 埋点定义
│   ├── model/                  # 领域模型
│   │   └── memory.go           # Memory 结构体
│   ├── service/                # 业务逻辑层
│   │   └── memory/
│   │       └── manager.go      # 记忆管理核心逻辑
│   ├── storage/                # 存储层
│   │   ├── postgres/           # PostgreSQL 适配
│   │   │   ├── metadata_store.go
│   │   │   ├── models.go
│   │   │   └── converter.go
│   │   └── milvus/             # Milvus 向量库适配
│   │       └── vector_store.go
│   └── mock/                   # 测试桩（Mock模式）
│       └── stores.go
│
├── pkg/                        # 公共库（可被外部引用）
│   ├── chatbot/                # 聊天机器人适配器
│   │   ├── interface.go
│   │   └── adapter.go
│   └── queue/                  # 消息队列组件
│       ├── redis_stream_queue.go
│       └── rate_limiter.go
│
├── config/                     # 配置文件
│   ├── config.example.yaml     # 配置模板
│   ├── config.docker.yaml      # Docker 环境配置
│   └── config.yaml             # 实际配置（不提交）
│
├── docker/                     # Docker 相关
│   ├── app/
│   │   └── Dockerfile          # 应用镜像
│   ├── postgres/
│   │   ├── init.sql            # 数据库初始化
│   │   ├── fix_schema.sql
│   │   └── migrations/         # 数据库迁移
│   └── README.md
│
├── deploy/                     # 部署配置
│   └── nginx.conf              # Nginx 反向代理配置
│
├── monitoring/                 # 监控配置
│   ├── prometheus.yml          # Prometheus 抓取配置
│   ├── alerts.yml              # 告警规则
│   ├── dashboards/
│   │   ├── dashboard-provider.yml
│   │   └── memoryos-overview.json  # Grafana 仪表盘
│   └── datasources/
│       └── prometheus.yml      # Grafana 数据源
│
├── scripts/                    # 自动化脚本
│   ├── deploy.sh               # 一键部署脚本
│   ├── init-server.sh          # 服务器初始化
│   ├── init-db.sql             # PostgreSQL 初始化
│   ├── build/                  # 构建脚本
│   └── dev/                    # 开发辅助脚本
│
├── docs/                       # 完整文档
│   ├── api/                    # API 文档
│   │   └── API_GUIDE.md
│   ├── deployment/             # 部署文档
│   │   ├── DEPLOYMENT_GUIDE.md
│   │   └── MONITORING_DEPLOYMENT.md
│   ├── guides/                 # 技术指南
│   │   ├── GEMINI_SETUP.md
│   │   ├── GEMINI_COST_OPTIMIZATION.md
│   │   └── MILVUS_IMPLEMENTATION.md
│   ├── dev/                    # 开发文档
│   │   ├── PROJECT_STRUCTURE.md
│   │   ├── CONTRIBUTING.md
│   │   ├── BUG_REPORT.md
│   │   ├── EMBEDDING_ERROR_HANDLING.md
│   │   ├── MESSAGE_QUEUE_GUIDE.md
│   │   └── MONITORING_M1_M2_REPORT.md
│   └── PROJECT_OVERVIEW.md     # 本文档
│
├── logs/                       # 日志目录
│   └── .gitkeep
│
├── bin/                        # 编译产物
│   └── server.exe
│
├── docker-compose.yml          # 本地开发环境
├── docker-compose.4c4g.yml     # 腾讯云 4C4G 生产环境
├── docker-compose.monitoring.yml  # 监控服务
├── Dockerfile                  # 生产镜像定义
├── Makefile                    # 构建命令
├── go.mod & go.sum             # Go 依赖管理
├── README.md                   # 项目主文档
├── CHANGELOG.md                # 版本变更记录
├── LICENSE                     # 开源协议
└── .gitignore                  # Git 忽略规则
```

---

## 🏗️ 核心架构

### 分层架构（6 层）

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Handler                        │  API 层
│                   (Gin Framework)                        │
├─────────────────────────────────────────────────────────┤
│                   Service Layer                          │  业务逻辑层
│          (Manager: 召回策略 & 记忆管理)                   │
├────────────┬────────────────────────┬────────────────────┤
│   LLM      │     Embedding          │    Storage         │  基础设施层
│  (OpenAI/  │    (Eino Adapter)      │  (Postgres/Milvus) │
│   Gemini)  │                        │                    │
├────────────┴────────────────────────┴────────────────────┤
│                  Metrics Layer                           │  可观测性
│           (Prometheus + Grafana)                         │
├─────────────────────────────────────────────────────────┤
│              Configuration & Bootstrap                   │  基础设施
│          (YAML Config + Dependency Injection)            │
└─────────────────────────────────────────────────────────┘
```

### 技术栈

| 层级 | 技术选型 |
|------|---------|
| **Web 框架** | Gin (高性能 HTTP 路由) |
| **数据库** | PostgreSQL 14 + pgvector |
| **向量库** | Milvus 2.3 (可选) |
| **缓存** | Redis 7 (LRU策略) |
| **LLM** | OpenAI API / Gemini API (中转接口) |
| **Embedding** | 火山引擎 Eino SDK (qwen3-embedding-4b) |
| **监控** | Prometheus 2.54 + Grafana 11.4 |
| **容器化** | Docker + Docker Compose |
| **CI/CD** | GitHub Actions |
| **反向代理** | Nginx |

---

## 🚀 快速导航

### 新手入门

1. **环境搭建**: [DEPLOYMENT_GUIDE.md](deployment/DEPLOYMENT_GUIDE.md)
2. **API 文档**: [API_GUIDE.md](api/API_GUIDE.md)
3. **配置指南**: [GEMINI_SETUP.md](guides/GEMINI_SETUP.md)

### 核心功能

- **记忆创建**: `POST /api/v1/memories`
- **混合召回**: `POST /api/v1/recall/hybrid`
- **监控指标**: `GET /metrics` (Prometheus 格式)

### 运维部署

- **一键部署**: `bash scripts/deploy.sh`
- **监控面板**: http://\<server-ip>:3000 (Grafana)
- **告警配置**: [monitoring/alerts.yml](../monitoring/alerts.yml)

### 开发指南

- **项目结构**: [PROJECT_STRUCTURE.md](dev/PROJECT_STRUCTURE.md)
- **贡献指南**: [CONTRIBUTING.md](dev/CONTRIBUTING.md)
- **问题报告**: [BUG_REPORT.md](dev/BUG_REPORT.md)

---

## 💻 开发指南

### 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/hxli2005/MemoryOs.git
cd MemoryOs

# 2. 配置环境
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 填写 API Key

# 3. 启动依赖服务
docker-compose up -d postgres redis

# 4. 运行服务
go run cmd/server/main.go

# 5. 测试 API
curl http://localhost:8080/health
```

### 生产部署

```bash
# 1. 初始化服务器
bash scripts/init-server.sh

# 2. 配置环境变量
cat > config/config.yaml << EOF
# ... 填写生产配置
EOF

# 3. 一键部署
bash scripts/deploy.sh
```

### 监控查看

- **Prometheus**: http://\<server-ip>:9090
- **Grafana**: http://\<server-ip>:3000
  - 账号: `admin`
  - 密码: `memoryos123`

---

## 📊 性能指标

### 资源消耗（4C4G 服务器）

| 服务 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| MemoryOS API | 1.5 核 | 1 GB | - |
| PostgreSQL | 1.0 核 | 768 MB | 20 GB |
| Redis | 0.5 核 | 256 MB | 1 GB |
| Prometheus | 0.5 核 | 512 MB | 10 GB |
| Grafana | 0.5 核 | 256 MB | 1 GB |
| **总计** | **4 核** | **~2.5 GB** | **32 GB** |

### 性能基准

- **记忆创建**: <500ms (P95)
- **混合召回**: <1s (P99)
- **Embedding生成**: <300ms (批处理 10条)
- **LLM调用**: 取决于模型 (Gemini Flash ~2s)

---

## 🔗 相关链接

- **GitHub仓库**: https://github.com/hxli2005/MemoryOs
- **Gemini API**: https://ai.google.dev/
- **火山引擎 Eino**: https://www.volcengine.com/
- **Prometheus文档**: https://prometheus.io/docs/
- **Grafana文档**: https://grafana.com/docs/

---

## 📝 变更日志

详见 [CHANGELOG.md](../CHANGELOG.md)

---

## 📄 许可证

本项目采用 [MIT License](../LICENSE)

---

**维护者**: hxli2005  
**最后更新**: 2026-02-06
