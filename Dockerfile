# MemoryOS 多阶段构建 Dockerfile
# 优化镜像体积并提升安全性

# 阶段 1: 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make gcc musl-dev

# 设置工作目录
WORKDIR /build

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用 Docker 缓存层）
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（禁用 CGO，静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always) -X main.BuildTime=$(date -u +%Y%m%d%H%M%S)" \
    -o server ./cmd/server

# 阶段 2: 运行阶段
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 memoryos && \
    adduser -D -u 1000 -G memoryos memoryos

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/server .

# 复制配置文件（如果需要）
# COPY --from=builder /build/config ./config

# 修改文件所有者
RUN chown -R memoryos:memoryos /app

# 切换到非 root 用户
USER memoryos

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
ENTRYPOINT ["./server"]
