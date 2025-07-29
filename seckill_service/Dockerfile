# 多阶段构建 Dockerfile for Go 秒杀服务
# Multi-stage build Dockerfile for Go seckill service

# 构建阶段 / Build stage
FROM golang:1.21-alpine AS builder

# 设置工作目录 / Set working directory
WORKDIR /app

# 安装必要的包 / Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件 / Copy go mod files
COPY go.mod go.sum ./

# 下载依赖 / Download dependencies
RUN go mod download

# 复制源代码 / Copy source code
COPY . .

# 构建应用 / Build application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o seckill-server \
    ./cmd/server

# 运行阶段 / Runtime stage
FROM alpine:latest

# 安装ca证书和时区数据 / Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# 设置时区 / Set timezone
ENV TZ=Asia/Shanghai

# 创建非root用户 / Create non-root user
RUN addgroup -g 1001 -S seckill && \
    adduser -u 1001 -S seckill -G seckill

# 设置工作目录 / Set working directory
WORKDIR /app

# 从构建阶段复制二进制文件 / Copy binary from build stage
COPY --from=builder /app/seckill-server .

# 更改文件所有者 / Change file ownership
RUN chown -R seckill:seckill /app

# 切换到非root用户 / Switch to non-root user
USER seckill

# 暴露端口 / Expose port
EXPOSE 8080

# 健康检查 / Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用 / Start application
CMD ["./seckill-server"]
