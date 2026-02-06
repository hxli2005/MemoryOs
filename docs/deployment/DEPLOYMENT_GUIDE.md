# MemoryOS 腾讯云 4C4G 部署指南

## 📋 部署清单

### 已创建的配置文件

```
MemoryOs/
├── docker-compose.4c4g.yml       # 4C4G 优化配置
├── .github/workflows/deploy.yml  # GitHub Actions CI/CD
├── deploy/nginx.conf             # Nginx 反向代理配置
├── scripts/
│   ├── init-server.sh           # 服务器初始化脚本
│   └── deploy.sh                # 一键部署脚本
└── .dockerignore                # Docker 构建忽略文件
```

---

## 🚀 快速开始（3 步完成部署）

### 步骤 1: 上传代码到 GitHub

```powershell
# 在本地 Windows 环境执行

# 1. 初始化 Git 仓库（如果还没有）
git init
git add .
git commit -m "feat(deploy): 添加腾讯云 4C4G 部署配置"

# 2. 关联远程仓库
git remote add origin https://github.com/YOUR_USERNAME/MemoryOs.git

# 3. 推送代码
git push -u origin main
```

---

### 步骤 2: 服务器初始化

```bash
# SSH 登录到腾讯云服务器
ssh root@YOUR_SERVER_IP

# 下载初始化脚本
curl -O https://raw.githubusercontent.com/YOUR_USERNAME/MemoryOs/main/scripts/init-server.sh

# 或者直接克隆仓库
git clone https://github.com/YOUR_USERNAME/MemoryOs.git
cd MemoryOs

# 执行初始化脚本
chmod +x scripts/init-server.sh
bash scripts/init-server.sh
```

**初始化脚本会自动安装**：
- ✅ Docker & Docker Compose
- ✅ Nginx
- ✅ Certbot (SSL 证书)
- ✅ UFW 防火墙配置
- ✅ 2GB Swap 空间

---

### 步骤 3: 配置并部署

```bash
# 1. 配置环境变量
cp config/config.example.yaml config/config.yaml
nano config/config.yaml

# 必须修改的配置项:
# - database.password: 强密码
# - embedding.api_key: 你的 Embedding API Key
# - llm.api_key: 你的 LLM API Key

# 2. 一键部署
chmod +x scripts/deploy.sh
bash scripts/deploy.sh

# 3. 验证部署
curl http://localhost:8080/health
```

---

## 🔧 配置 Nginx + SSL（可选但推荐）

### 1. 配置域名解析

在腾讯云 DNS 控制台添加 A 记录:
```
主机记录: memoryos (或 @)
记录类型: A
记录值: YOUR_SERVER_IP
TTL: 600
```

### 2. 部署 Nginx 配置

```bash
# 复制 Nginx 配置
cp deploy/nginx.conf /etc/nginx/sites-available/memoryos

# 修改域名
sed -i 's/YOUR_DOMAIN/memoryos.yourdomain.com/g' /etc/nginx/sites-available/memoryos

# 启用配置
ln -s /etc/nginx/sites-available/memoryos /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

### 3. 申请 SSL 证书

```bash
# 使用 Let's Encrypt 免费证书
certbot --nginx -d memoryos.yourdomain.com

# 自动续期测试
certbot renew --dry-run
```

---

## 🤖 配置 GitHub Actions CI/CD

### 1. 在 GitHub 仓库配置 Secrets

导航: **Settings → Secrets and variables → Actions → New repository secret**

添加以下 Secrets:

| Secret 名称 | 值 | 说明 |
|------------|-----|------|
| `TENCENT_CCR_USERNAME` | 你的腾讯云账号 ID | 容器镜像服务用户名 |
| `TENCENT_CCR_PASSWORD` | 生成的访问凭证 | 容器镜像服务密码 |
| `SERVER_HOST` | 服务器公网 IP | 如: 43.xxx.xxx.xxx |
| `SERVER_USER` | `root` | SSH 登录用户 |
| `SERVER_SSH_KEY` | SSH 私钥内容 | 见下方生成方法 |
| `SERVER_DOMAIN` | 域名 (可选) | 如: memoryos.yourdomain.com |

### 2. 生成 SSH 密钥

```bash
# 在服务器上生成密钥对
ssh-keygen -t rsa -b 4096 -C "github-actions" -f ~/.ssh/github_actions_key

# 将公钥添加到授权列表
cat ~/.ssh/github_actions_key.pub >> ~/.ssh/authorized_keys

# 复制私钥内容（用于 GitHub Secret）
cat ~/.ssh/github_actions_key

# 将完整输出（包括 -----BEGIN RSA PRIVATE KEY----- 和 -----END RSA PRIVATE KEY-----）
# 复制到 GitHub Secret: SERVER_SSH_KEY
```

### 3. 配置腾讯云容器镜像服务

1. 登录腾讯云控制台
2. 搜索"容器镜像服务 TCR"
3. 创建个人版命名空间: `memoryos`
4. 生成访问凭证: 账号设置 → 访问凭证 → 新建
5. 记录用户名和密码，填入 GitHub Secrets

### 4. 触发自动部署

```bash
# 本地修改代码后推送
git add .
git commit -m "feat: 新功能"
git push origin main

# GitHub Actions 会自动:
# 1. 运行测试
# 2. 构建 Docker 镜像
# 3. 推送到腾讯云容器镜像仓库
# 4. SSH 登录服务器
# 5. 拉取新镜像并滚动更新
# 6. 执行健康检查
```

在 GitHub 仓库的 **Actions** 标签页可以查看部署进度。

---

## 📊 访问服务

部署成功后，可以通过以下地址访问:

| 服务 | HTTP 地址 | HTTPS 地址 (配置 SSL 后) |
|------|-----------|-------------------------|
| **API 文档** | http://YOUR_IP:8080/api/v1/ | https://YOUR_DOMAIN/api/v1/ |
| **健康检查** | http://YOUR_IP:8080/health | https://YOUR_DOMAIN/health |
| **Prometheus** | http://YOUR_IP:9090 | https://YOUR_DOMAIN/prometheus/ |
| **Grafana** | http://YOUR_IP:3000 | https://YOUR_DOMAIN/grafana/ |

**默认登录凭证**:
- Grafana: `admin` / `memoryos123`（首次登录后请修改）

---

## 🔍 常用运维命令

### 查看服务状态

```bash
cd /root/MemoryOs
docker-compose -f docker-compose.4c4g.yml ps
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.4c4g.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.4c4g.yml logs -f memoryos
docker-compose -f docker-compose.4c4g.yml logs -f postgres
```

### 重启服务

```bash
# 重启所有服务
docker-compose -f docker-compose.4c4g.yml restart

# 重启单个服务
docker-compose -f docker-compose.4c4g.yml restart memoryos
```

### 进入容器

```bash
# 进入 API 容器
docker exec -it memoryos-api sh

# 进入数据库容器
docker exec -it memoryos-postgres psql -U memoryos
```

### 备份数据库

```bash
# 导出数据库
docker exec memoryos-postgres pg_dump -U memoryos memoryos > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker exec -i memoryos-postgres psql -U memoryos memoryos < backup_20260206.sql
```

### 清理磁盘空间

```bash
# 清理未使用的镜像
docker image prune -a -f

# 清理未使用的卷
docker volume prune -f

# 清理所有未使用的资源
docker system prune -a -f
```

---

## 🛡️ 安全建议

### 1. 修改默认密码

```bash
# 修改数据库密码
nano config/config.yaml  # 修改 database.password

# 修改 Grafana 密码
# 登录 Grafana → 右上角头像 → Change Password
```

### 2. 限制端口访问

```bash
# 仅允许 Nginx 反向代理访问
# 修改 docker-compose.4c4g.yml，移除端口映射:
# ports:
#   - "8080:8080"  # 删除此行

# 仅绑定 localhost
ports:
  - "127.0.0.1:8080:8080"
```

### 3. 启用 Fail2Ban

```bash
apt install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban
```

---

## 📈 性能监控

### 查看资源使用

```bash
# Docker 容器资源占用
docker stats

# 系统资源
htop

# 磁盘使用
df -h
du -sh /var/lib/docker
```

### Prometheus 关键指标

访问 Prometheus (http://YOUR_IP:9090/graph) 执行以下查询:

```promql
# 召回延迟 P95
histogram_quantile(0.95, rate(memory_recall_duration_seconds_bucket[5m]))

# 记忆创建成功率
sum(rate(memory_create_total{status="success"}[5m])) / sum(rate(memory_create_total[5m]))

# Embedding 生成 QPS
rate(embedding_requests_total[1m])

# 系统内存使用
memory_usage_bytes / 1024 / 1024
```

---

## 🆘 故障排查

### 问题 1: 服务启动失败

```bash
# 查看详细日志
docker-compose -f docker-compose.4c4g.yml logs memoryos

# 常见原因:
# - 配置文件错误: 检查 config/config.yaml
# - 端口被占用: lsof -i:8080
# - 内存不足: free -h
```

### 问题 2: 数据库连接失败

```bash
# 检查 PostgreSQL 状态
docker exec memoryos-postgres pg_isready -U memoryos

# 查看数据库日志
docker logs memoryos-postgres

# 手动连接测试
docker exec -it memoryos-postgres psql -U memoryos
```

### 问题 3: GitHub Actions 部署失败

```bash
# 检查 SSH 连接
ssh -i ~/.ssh/github_actions_key root@YOUR_SERVER_IP

# 检查镜像仓库登录
docker login ccr.ccs.tencentyun.com

# 手动拉取镜像测试
docker pull ccr.ccs.tencentyun.com/memoryos/memoryos-api:latest
```

---

## 📞 支持

遇到问题？
1. 查看日志: `docker-compose logs -f`
2. 检查 GitHub Actions 构建日志
3. 提交 Issue: https://github.com/YOUR_USERNAME/MemoryOs/issues

---

**祝部署顺利！🎉**
