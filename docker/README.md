# MemoryOS Docker 部署指南

## 🚀 快速启动

### 1. 启动所有服务
```powershell
.\start_docker.bat
```

或手动启动：
```powershell
docker compose up -d
```

### 2. 查看服务状态
```powershell
docker compose ps
```

### 3. 测试连接
```powershell
.\test_docker.bat
```

---

## 📦 服务列表

| 服务 | 端口 | 访问地址 | 用途 |
|------|------|----------|------|
| **PostgreSQL** | 5432 | `postgres:5432` (容器内) | 元数据存储 + pgvector |
| **Redis** | 6379 | `redis:6379` (容器内) | 缓存 |
| **Milvus** | 19530 | `milvus:19530` (容器内) | 向量检索 |
| **MinIO** | 9000 | http://localhost:9001 | 对象存储控制台 |
| **MemoryOS** | 8080 | http://localhost:8080 | 主应用 API |

---

## 🔧 常用命令

### 查看日志
```powershell
# 所有服务
docker compose logs -f

# 特定服务
docker compose logs -f memoryos
docker compose logs -f postgres
docker compose logs -f milvus
```

### 重启服务
```powershell
# 重启所有
docker compose restart

# 重启特定服务
docker compose restart memoryos
```

### 停止服务
```powershell
.\stop_docker.bat

# 或
docker compose down
```

### 完全清理（包括数据）
```powershell
docker compose down -v
rmdir /s /q data
```

---

## 🗄️ 数据持久化

所有数据存储在项目目录下的 `data/` 文件夹：

```
data/
├── postgres/     # PostgreSQL 数据文件
├── redis/        # Redis 持久化文件
├── milvus/       # Milvus 向量数据
├── etcd/         # etcd 元数据
└── minio/        # MinIO 对象存储
```

**备份**：直接复制 `data/` 文件夹即可

---

## 🐛 故障排查

### 问题 1: PostgreSQL 启动失败
**现象**：容器反复重启

**检查日志**：
```powershell
docker compose logs postgres
```

**常见原因**：
- 端口 5432 被占用
- 数据目录权限问题

**解决方案**：
```powershell
# 检查端口占用
netstat -ano | findstr "5432"

# 重置数据（危险操作）
docker compose down
rmdir /s /q data\postgres
docker compose up -d postgres
```

### 问题 2: Milvus 启动超时
**现象**：Milvus 健康检查失败

**原因**：
- Milvus 首次启动需要 1-2 分钟初始化
- etcd 或 MinIO 未就绪

**解决方案**：
```powershell
# 等待 2 分钟后检查
docker compose logs -f milvus

# 确保依赖服务正常
docker compose ps etcd minio
```

### 问题 3: MemoryOS 连接数据库失败
**现象**：应用日志显示连接错误

**检查步骤**：
```powershell
# 1. 确认数据库运行
docker exec memoryos-postgres psql -U memoryos -d memoryos -c "SELECT 1"

# 2. 检查环境变量
docker compose exec memoryos env | findstr DB_

# 3. 查看应用日志
docker compose logs memoryos
```

### 问题 4: 磁盘空间不足
**检查占用**：
```powershell
docker system df
```

**清理未使用资源**：
```powershell
docker system prune -a --volumes
```

---

## 📊 性能优化

### 限制资源使用
编辑 `docker-compose.yml`，添加资源限制：

```yaml
services:
  milvus:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

### 调整 PostgreSQL 参数
编辑 `data/postgres/postgresql.conf`（首次启动后生成）：

```ini
shared_buffers = 512MB
effective_cache_size = 2GB
max_connections = 100
```

---

## 🔒 安全建议

### 生产环境部署
1. **修改默认密码**（编辑 `.env`）：
   ```env
   POSTGRES_PASSWORD=your_strong_password
   MINIO_ROOT_PASSWORD=your_minio_password
   ```

2. **启用 TLS**（PostgreSQL）：
   - 生成证书
   - 配置 `sslmode=require`

3. **网络隔离**：
   - 不暴露数据库端口（删除 `ports` 配置）
   - 只暴露应用端口 8080

---

## 📈 扩展部署

### 水平扩展（多实例）
```yaml
services:
  memoryos:
    deploy:
      replicas: 3  # 运行 3 个实例
```

### 负载均衡
使用 Nginx 或 Traefik 进行反向代理

---

## 🔄 更新部署

### 更新应用代码
```powershell
# 1. 拉取最新代码
git pull

# 2. 重新构建镜像
docker compose build memoryos

# 3. 重启应用
docker compose up -d memoryos
```

### 更新镜像版本
编辑 `docker-compose.yml`，修改镜像版本：
```yaml
milvus:
  image: milvusdb/milvus:v2.3.4  # 更新版本号
```

然后执行：
```powershell
docker compose pull
docker compose up -d
```

---

## 📝 开发工作流

### 开发模式（热重载）
```yaml
services:
  memoryos:
    volumes:
      - ./:/app  # 挂载源码
    command: go run cmd/server/main.go
```

### 生产模式（当前配置）
使用多阶段构建，编译后的二进制文件运行

---

## 🎯 下一步

- [ ] 连接到数据库，验证表结构
- [ ] 实现真实的 PostgreSQL Store
- [ ] 实现 Milvus Vector Store
- [ ] 性能压测
- [ ] 部署到云服务器

---

**快速命令参考**：
```powershell
# 启动
.\start_docker.bat

# 测试
.\test_docker.bat

# 日志
.\logs_docker.bat memoryos

# 停止
.\stop_docker.bat
```
