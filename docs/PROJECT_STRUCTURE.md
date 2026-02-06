# MemoryOS 项目结构

> 本文档描述了 MemoryOS 项目的完整目录结构和各模块职责

## 📁 总体结构

```
MemoryOS/
├── .github/                    # GitHub 配置
│   ├── copilot-instructions.md # Copilot 指令
│   └── workflows/              # CI/CD 工作流
│       └── deploy.yml          # 自动部署配置
│
├── cmd/                        # 应用程序入口
│   └── server/
│       └── main.go             # HTTP 服务器启动
│
├── internal/                   # 内部私有代码（不可被外部引用）
│   ├── adapter/                # 外部服务适配器
│   │   └── eino.go             # Eino Embedding 适配器
│   ├── bootstrap/              # 应用启动器
│   │   └── app.go              # 依赖注入与初始化
│   ├── config/                 # 配置管理
│   │   └── config.go           # 配置结构与加载
│   ├── handler/                # HTTP 请求处理器
│   │   └── memory.go           # 记忆管理 API
│   ├── llm/                    # LLM 服务集成
│   │   ├── interface.go        # LLM 接口定义
│   │   ├── openai.go           # OpenAI 实现
│   │   └── gemini.go           # Gemini 实现
│   ├── metrics/                # Prometheus 监控指标
│   │   └── metrics.go          # 指标定义与采集
│   ├── mock/                   # Mock 实现（测试/开发）
│   │   └── stores.go           # Mock 存储层
│   ├── model/                  # 数据模型
│   │   └── memory.go           # 记忆数据结构
│   ├── service/                # 业务逻辑层
│   │   └── memory/
│   │       └── manager.go      # 记忆管理核心逻辑
│   └── storage/                # 存储层
│       ├── milvus/             # 向量数据库
│       │   └── vector_store.go
│       └── postgres/           # 关系型数据库
│           ├── models.go       # GORM 模型
│           ├── metadata_store.go # 元数据存储
│           └── converter.go    # 数据转换
│
├── config/                     # 配置文件目录
│   ├── config.yaml             # 本地开发配置（不提交）
│   ├── config.example.yaml     # 配置模板
│   └── config.docker.yaml      # Docker 环境配置
│
├── deploy/                     # 部署相关配置
│   └── nginx.conf              # Nginx 反向代理配置
│
├── docker/                     # Docker 相关文件
│   ├── app/
│   │   └── Dockerfile          # 应用容器镜像
│   └── postgres/
│       ├── init.sql            # 数据库初始化脚本
│       ├── fix_schema.sql      # Schema 修复脚本
│       └── migrations/         # 数据库迁移
│           ├── 001_add_memory_uuid.sql
│           └── 002_add_embedding.sql
│
├── docs/                       # 项目文档
│   ├── api/                    # API 文档
│   │   └── API_GUIDE.md
│   ├── dev/                    # 开发者文档
│   │   ├── CONTRIBUTING.md     # 贡献指南
│   │   ├── BUG_REPORT.md       # Bug 报告模板
│   │   └── PROJECT_STRUCTURE.md
│   ├── guides/                 # 使用指南
│   │   ├── GEMINI_SETUP.md
│   │   ├── GEMINI_COST_OPTIMIZATION.md
│   │   └── MILVUS_IMPLEMENTATION.md
│   ├── DEPLOYMENT_GUIDE.md     # 生产部署指南
│   ├── EMBEDDING_ERROR_HANDLING.md
│   ├── MESSAGE_QUEUE_GUIDE.md
│   ├── MONITORING_DEPLOYMENT.md
│   └── MONITORING_M1_M2_REPORT.md
│
├── monitoring/                 # 监控配置
│   ├── prometheus.yml          # Prometheus 采集配置
│   ├── alerts.yml              # 告警规则
│   ├── dashboards/             # Grafana 仪表盘
│   │   ├── dashboard-provider.yml
│   │   └── memoryos-overview.json
│   └── datasources/            # 数据源配置
│       └── prometheus.yml
│
├── scripts/                    # 自动化脚本
│   ├── deploy.sh               # 一键部署脚本
│   ├── init-server.sh          # 服务器环境初始化
│   ├── init-db.sql             # PostgreSQL 初始化
│   ├── dev/                    # 开发环境脚本
│   │   ├── start_docker.bat
│   │   ├── stop_docker.bat
│   │   └── logs_docker.bat
│   └── build/
│       └── monitor_build.bat
│
├── pkg/                        # 公共库（可被外部项目引用）
│   └── (预留，当前为空)
│
├── Dockerfile                  # 多阶段构建配置
├── docker-compose.yml          # 本地开发环境
├── docker-compose.monitoring.yml  # 监控栈（独立）
├── docker-compose.4c4g.yml     # 腾讯云 4C4G 生产配置
├── go.mod                      # Go 模块依赖
├── go.sum                      # 依赖校验文件
├── Makefile                    # 构建命令
├── README.md                   # 项目说明
├── CHANGELOG.md                # 变更日志
└── .gitignore                  # Git 忽略规则

```

## 🏗️ 架构分层

### 1. **Handler 层** (HTTP 入口)
- 路径：`internal/handler/`
- 职责：接收 HTTP 请求，参数验证，调用 Service 层
- 依赖：Service 层

### 2. **Service 层** (业务逻辑)
- 路径：`internal/service/`
- 职责：核心业务逻辑，流程编排，事务管理
- 依赖：Storage 层、LLM 层、Adapter 层

### 3. **Storage 层** (数据持久化)
- 路径：`internal/storage/`
- 职责：数据库操作，CRUD 封装
- 实现：
  - `postgres/`: 元数据存储 (PostgreSQL + pgvector)
  - `milvus/`: 向量检索 (Milvus)

### 4. **LLM 层** (AI 能力)
- 路径：`internal/llm/`
- 职责：LLM API 调用封装
- 实现：OpenAI, Gemini

### 5. **Adapter 层** (外部服务)
- 路径：`internal/adapter/`
- 职责：第三方服务适配
- 示例：Eino Embedding Service

### 6. **Metrics 层** (可观测性)
- 路径：`internal/metrics/`
- 职责：Prometheus 指标采集

## 📝 配置文件说明

| 文件 | 用途 | 提交到 Git |
|------|------|-----------|
| `config.yaml` | 本地开发配置（含敏感信息） | ❌ |
| `config.example.yaml` | 配置模板 | ✅ |
| `config.docker.yaml` | Docker Compose 环境 | ✅ |
| `.env` | 环境变量（含密钥） | ❌ |

## 🐳 Docker 配置文件

| 文件 | 用途 |
|------|------|
| `Dockerfile` | 生产环境多阶段构建 |
| `docker-compose.yml` | 本地开发（PostgreSQL + Redis + 应用） |
| `docker-compose.monitoring.yml` | 监控栈（Prometheus + Grafana） |
| `docker-compose.4c4g.yml` | 腾讯云 4C4G 生产部署 |

## 📚 文档组织

```
docs/
├── api/              # API 接口文档
├── dev/              # 开发者指南
├── guides/           # 功能使用指南
└── *.md              # 核心文档（部署、监控等）
```

## 🚀 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/hxli2005/MemoryOs.git
cd MemoryOs
```

### 2. 配置环境
```bash
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 填入 API 密钥
```

### 3. 启动服务
```bash
# 开发环境
docker-compose up -d

# 生产环境（4C4G）
docker-compose -f docker-compose.4c4g.yml up -d
```

### 4. 验证部署
```bash
curl http://localhost:8080/health
```

## 🔗 相关文档

- [API 使用指南](docs/api/API_GUIDE.md)
- [部署指南](docs/DEPLOYMENT_GUIDE.md)
- [监控部署](docs/MONITORING_DEPLOYMENT.md)
- [贡献指南](docs/dev/CONTRIBUTING.md)

## 📊 技术栈

- **语言**: Go 1.24
- **Web 框架**: Gin
- **数据库**: PostgreSQL 14 (pgvector)
- **向量数据库**: Milvus / Qdrant
- **缓存**: Redis 7
- **监控**: Prometheus + Grafana
- **容器化**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **反向代理**: Nginx

---

最后更新：2026-02-06
