# QQ Bot Docker 镜像
# 多阶段构建，最终镜像约 20MB

# ===== 构建阶段 =====
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 设置 Go 代理（解决网络问题）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GO111MODULE=on

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译（静态链接）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /qqbot ./examples/qqbot/

# ===== 运行阶段 =====
FROM alpine:latest

WORKDIR /app

# 安装 CA 证书（HTTPS 请求需要）
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制
COPY --from=builder /qqbot /app/qqbot

# 复制配置文件（可被挂载覆盖）
COPY config/config.yaml /app/config/config.yaml
COPY examples/qqbot/persona.yaml /app/config/persona.yaml

# 环境变量（可覆盖默认配置）
ENV CQHTTP_WS_URL=ws://host.docker.internal:6700
ENV CONFIG_PATH=/app/config/config.yaml
ENV PERSONA_PATH=/app/config/persona.yaml

ENTRYPOINT ["/app/qqbot"]
