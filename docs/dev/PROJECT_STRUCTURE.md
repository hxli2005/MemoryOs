# MemoryOS 项目结构

## 📁 目录说明

```
MemoryOs/
├── .github/
│   └── copilot-instructions.md    # GitHub Copilot 指令配置
│
├── cmd/
│   └── server/
│       └── main.go                # 服务器主入口
│
├── config/
│   ├── config.yaml                # 实际配置（不提交）
│   └── config.example.yaml        # 配置模板
│
├── docker/
│   ├── app/
│   │   └── Dockerfile             # 应用容器化配置
│   ├── postgres/
│   │   └── init.sql              # PostgreSQL 初始化脚本
│   └── README.md                  # Docker 部署文档
│
├── docs/
│   ├── API_GUIDE.md              # API 使用指南
│   ├── CHATBOT_USAGE.md          # Chatbot 使用说明
│   ├── GEMINI_SETUP.md           # Gemini 配置指南
│   └── GEMINI_COST_OPTIMIZATION.md  # 成本优化建议
│
├── examples/
│   └── chatbot/
│       └── main.go               # Chatbot 示例
│
├── internal/
│   ├── adapter/
│   │   └── eino.go              # Eino 框架适配器
│   ├── bootstrap/
│   │   └── app.go               # 应用初始化
│   ├── config/
│   │   └── config.go            # 配置结构定义
│   ├── handler/
│   │   └── memory.go            # HTTP 请求处理器
│   ├── mock/
│   │   └── stores.go            # Mock 存储实现
│   ├── model/
│   │   └── memory.go            # 数据模型定义
│   └── service/
│       └── memory/
│           └── manager.go       # 核心业务逻辑
│
├── test/
│   └── index.html               # Web 测试页面
│
├── .dockerignore                # Docker 构建忽略文件
├── .env                         # 环境变量（不提交）
├── .gitignore                   # Git 忽略文件
├── docker-compose.yml           # Docker Compose 配置
├── docker-manage.sh             # Docker 管理脚本（Linux/Mac）
├── go.mod                       # Go 模块定义
├── go.sum                       # Go 依赖校验
├── Makefile                     # 构建任务定义
├── README.md                    # 项目说明
├── test_api.http                # API 测试文件（REST Client）
├── test_docker.bat              # Docker 环境测试（Windows）
├── start_docker.bat             # 启动 Docker 服务（Windows）
├── stop_docker.bat              # 停止 Docker 服务（Windows）
├── start_chatbot.bat            # 启动 Chatbot（Windows）
├── logs_docker.bat              # 查看 Docker 日志（Windows）
└── monitor_build.bat            # 监控构建进度（Windows）
```

## 📂 运行时目录（不提交到 Git）

```
MemoryOs/
├── data/                        # Docker 数据持久化
│   ├── postgres/               # PostgreSQL 数据
│   ├── redis/                  # Redis 数据
│   ├── milvus/                 # Milvus 数据
│   ├── etcd/                   # etcd 数据
│   └── minio/                  # MinIO 数据
│
└── logs/                        # 应用日志
```

## 🔧 核心文件说明

### 入口文件
- **cmd/server/main.go**: 应用启动入口，初始化并启动 HTTP 服务器

### 配置相关
- **config/config.yaml**: 实际配置文件（包含敏感信息，不提交）
- **config/config.example.yaml**: 配置模板，展示配置结构
- **.env**: Docker Compose 环境变量

### 业务逻辑
- **internal/service/memory/manager.go**: 核心记忆管理逻辑
- **internal/handler/memory.go**: HTTP API 处理器
- **internal/model/memory.go**: 数据模型定义

### 存储层
- **internal/mock/stores.go**: Mock 存储实现（开发/测试用）
- **internal/storage/** (待实现): 真实存储实现
  - postgres/: PostgreSQL + pgvector
  - redis/: Redis 缓存
  - milvus/: Milvus 向量检索

### Docker 相关
- **docker-compose.yml**: 服务编排配置
- **docker/app/Dockerfile**: 应用镜像构建
- **docker/postgres/init.sql**: 数据库初始化

### 辅助脚本
- **start_docker.bat**: 一键启动所有服务
- **test_docker.bat**: 验证环境配置
- **monitor_build.bat**: 监控镜像构建

## 📝 文档索引

| 文档 | 用途 |
|------|------|
| [README.md](../README.md) | 项目总览 |
| [API_GUIDE.md](API_GUIDE.md) | API 使用说明 |
| [CHATBOT_USAGE.md](CHATBOT_USAGE.md) | Chatbot 示例 |
| [docker/README.md](../docker/README.md) | Docker 部署指南 |
| [GEMINI_SETUP.md](GEMINI_SETUP.md) | Gemini 配置 |

## 🚀 快速命令

```bash
# 启动 Docker 环境
.\start_docker.bat

# 测试环境
.\test_docker.bat

# 启动 Chatbot 示例
.\start_chatbot.bat

# 查看日志
.\logs_docker.bat memoryos

# 停止服务
.\stop_docker.bat
```

## 🔄 开发工作流

1. **本地开发**: 使用 Mock 模式快速测试
2. **Docker 开发**: 连接真实数据库
3. **生产部署**: 云服务器 + 真实存储

---

**维护说明**：
- 定期更新文档与代码同步
- 清理不用的测试文件和临时文件
- 保持目录结构简洁明了
