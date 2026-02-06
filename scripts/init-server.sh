#!/bin/bash

###############################################################################
# MemoryOS 服务器初始化脚本
# 适用于: Ubuntu 20.04/22.04, Debian 11/12
# 功能: 安装 Docker, Docker Compose, Nginx, 配置防火墙等
###############################################################################

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then 
    log_error "请使用 root 用户运行此脚本"
    exit 1
fi

log_info "开始初始化 MemoryOS 服务器环境..."

# 1. 更新系统
log_info "步骤 1/8: 更新系统软件包..."
apt update && apt upgrade -y

# 2. 安装基础工具
log_info "步骤 2/8: 安装基础工具..."
apt install -y \
    curl \
    wget \
    git \
    vim \
    htop \
    ufw \
    ca-certificates \
    gnupg \
    lsb-release

# 3. 安装 Docker
log_info "步骤 3/8: 安装 Docker..."
if command -v docker &> /dev/null; then
    log_warn "Docker 已安装，跳过"
else
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    log_info "Docker 安装完成: $(docker --version)"
fi

# 4. 安装 Docker Compose
log_info "步骤 4/8: 安装 Docker Compose..."
if command -v docker-compose &> /dev/null; then
    log_warn "Docker Compose 已安装，跳过"
else
    DOCKER_COMPOSE_VERSION="v2.24.0"
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    log_info "Docker Compose 安装完成: $(docker-compose --version)"
fi

# 5. 安装 Nginx
log_info "步骤 5/8: 安装 Nginx..."
if command -v nginx &> /dev/null; then
    log_warn "Nginx 已安装，跳过"
else
    apt install -y nginx
    systemctl enable nginx
    systemctl start nginx
    log_info "Nginx 安装完成: $(nginx -v)"
fi

# 6. 安装 Certbot (Let's Encrypt SSL)
log_info "步骤 6/8: 安装 Certbot..."
apt install -y certbot python3-certbot-nginx

# 7. 配置防火墙
log_info "步骤 7/8: 配置防火墙..."
ufw --force enable
ufw allow 22/tcp comment 'SSH'
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'
ufw status

# 8. 配置 Swap (2GB)
log_info "步骤 8/8: 配置 Swap 空间..."
if [ -f /swapfile ]; then
    log_warn "Swap 文件已存在，跳过"
else
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    log_info "Swap 配置完成: $(free -h | grep Swap)"
fi

# 9. 配置 Docker 日志轮转
log_info "配置 Docker 日志轮转..."
cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker

# 10. 克隆项目（可选）
log_info "是否克隆 MemoryOS 项目？(y/n)"
read -r CLONE_PROJECT
if [ "$CLONE_PROJECT" = "y" ]; then
    log_info "请输入 Git 仓库地址:"
    read -r GIT_REPO
    cd /root
    git clone "$GIT_REPO" MemoryOs
    cd MemoryOs
    log_info "项目克隆完成: /root/MemoryOs"
fi

# 完成
log_info "====================================="
log_info "✅ 服务器初始化完成！"
log_info "====================================="
log_info "已安装组件:"
log_info "  - Docker: $(docker --version)"
log_info "  - Docker Compose: $(docker-compose --version)"
log_info "  - Nginx: $(nginx -v 2>&1)"
log_info "  - Certbot: $(certbot --version)"
log_info ""
log_info "下一步操作:"
log_info "  1. 配置环境变量: cd /root/MemoryOs && cp config/config.example.yaml config/config.yaml"
log_info "  2. 启动服务: bash scripts/deploy.sh"
log_info "  3. 配置 SSL: certbot --nginx -d your-domain.com"
log_info ""
log_warn "重要提示:"
log_warn "  - 请修改默认密码（数据库、Grafana 等）"
log_warn "  - 配置 GitHub Secrets 以启用 CI/CD"
log_warn "  - 定期备份数据库数据"
