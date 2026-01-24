# Scripts 目录说明

本目录包含开发、构建和测试所需的自动化脚本。

## 目录结构

```
scripts/
├── dev/              # 开发环境相关
│   ├── start_docker.bat
│   ├── stop_docker.bat
│   ├── logs_docker.bat
│   └── start_chatbot.bat
├── build/            # 构建相关
│   └── monitor_build.bat
└── test/             # 测试相关
    ├── test_docker.bat
    ├── test_milvus.bat
    └── test_storage.ps1
```

## 开发脚本 (dev/)

### start_docker.bat
启动所有依赖服务（PostgreSQL、Milvus、Redis、etcd）

```bash
# Windows
cd scripts/dev
.\start_docker.bat

# Linux/Mac
chmod +x start_docker.sh
./start_docker.sh
```

### stop_docker.bat
停止并清理 Docker 容器

### logs_docker.bat
查看 Docker 容器日志

### start_chatbot.bat
快速启动 Chatbot 示例

## 构建脚本 (build/)

### monitor_build.bat
监控构建过程并输出详细日志

## 测试脚本 (test/)

### test_docker.bat
测试 Docker 服务连通性

### test_milvus.bat
测试 Milvus 向量数据库功能

### test_storage.ps1
测试存储层完整性（PostgreSQL + Milvus）

---

**注意**: 某些脚本可能需要管理员权限或 Docker Desktop 正在运行。
