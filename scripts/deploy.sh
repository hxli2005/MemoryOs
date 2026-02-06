#!/bin/bash

###############################################################################
# MemoryOS 一键部署脚本
# 功能: 快速部署/更新 MemoryOS 服务
###############################################################################

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR" || exit 1

log_info "====================================="
log_info "  MemoryOS 一键部署脚本"
log_info "====================================="
log_info "项目目录: $PROJECT_DIR"
log_info ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    log_error "Docker 未安装，请先运行 scripts/init-server.sh"
    exit 1
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose 未安装，请先运行 scripts/init-server.sh"
    exit 1
fi

# 1. 检查配置文件
log_info "步骤 1/7: 检查配置文件..."
if [ ! -f "config/config.yaml" ]; then
    log_warn "config.yaml 不存在，正在从示例文件创建..."
    if [ -f "config/config.example.yaml" ]; then
        cp config/config.example.yaml config/config.yaml
        log_warn "请编辑 config/config.yaml 填写正确的配置信息"
        log_warn "按任意键继续或 Ctrl+C 取消..."
        read -r
    else
        log_error "config.example.yaml 不存在"
        exit 1
    fi
fi

# 2. 拉取最新代码（如果是 Git 仓库）
if [ -d ".git" ]; then
    log_info "步骤 2/7: 拉取最新代码..."
    git pull origin main || git pull origin master || log_warn "代码拉取失败，继续部署..."
else
    log_warn "步骤 2/7: 不是 Git 仓库，跳过代码拉取"
fi

# 3. 停止旧服务（可选）
log_info "步骤 3/7: 停止旧服务..."
docker-compose -f docker-compose.4c4g.yml down || log_warn "没有运行中的服务"

# 4. 构建镜像
log_info "步骤 4/7: 构建 Docker 镜像..."
docker-compose -f docker-compose.4c4g.yml build --no-cache memoryos

# 5. 启动服务
log_info "步骤 5/7: 启动服务..."
docker-compose -f docker-compose.4c4g.yml up -d

# 6. 等待服务启动
log_info "步骤 6/7: 等待服务启动..."
sleep 15

# 7. 健康检查
log_info "步骤 7/7: 执行健康检查..."
MAX_RETRIES=10
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_info "✅ 健康检查通过"
        break
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        log_error "❌ 健康检查失败"
        log_error "查看日志: docker-compose -f docker-compose.4c4g.yml logs memoryos"
        exit 1
    fi
    
    log_warn "等待服务就绪... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 3
done

# 显示服务状态
log_info ""
log_info "====================================="
log_info "📊 服务运行状态"
log_info "====================================="
docker-compose -f docker-compose.4c4g.yml ps

# 显示访问地址
log_info ""
log_info "====================================="
log_info "🚀 部署完成！"
log_info "====================================="
log_info "访问地址:"
log_info "  - API 健康检查: http://localhost:8080/health"
log_info "  - Prometheus: http://localhost:9090"
log_info "  - Grafana: http://localhost:3000 (admin / memoryos123)"
log_info ""
log_info "查看日志:"
log_info "  docker-compose -f docker-compose.4c4g.yml logs -f"
log_info ""
log_info "停止服务:"
log_info "  docker-compose -f docker-compose.4c4g.yml down"
log_info ""
log_warn "提示: 请配置 Nginx 反向代理和 SSL 证书以启用 HTTPS 访问"
