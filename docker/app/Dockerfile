# ==================== 多阶段构建 Dockerfile ====================
# 阶段 1: 编译阶段
# 阶段 2: 运行阶段

# ==================== 阶段 1: 构建阶段 ====================
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /build

# 配置 Go 代理（国内镜像）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# 安装编译依赖
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（静态链接，减小镜像体积）
RUN GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o memoryos \
    ./cmd/server/main.go

# ==================== 阶段 2: 运行阶段 ====================
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 memoryos && \
    adduser -D -u 1000 -G memoryos memoryos

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /build/memoryos .

# 复制配置文件（将被 volume 覆盖）
COPY --from=builder /build/config ./config

# 创建日志目录
RUN mkdir -p /app/logs && \
    chown -R memoryos:memoryos /app

# 切换到非 root 用户
USER memoryos

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./memoryos"]
