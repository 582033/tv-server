# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/web

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates，用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 创建缓存目录
RUN mkdir -p /app/cache

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 设置环境变量
ENV GIN_MODE=release

# 暴露端口
EXPOSE 10101

# 设置用户
RUN adduser -D appuser
RUN chown -R appuser:appuser /app
USER appuser

# 运行应用
CMD ["./server", "-c", "/config.json"]
