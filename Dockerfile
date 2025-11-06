FROM golang:1.19-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/rst ./cmd/rst

# 创建最终镜像
FROM alpine:latest

# 安装 CA 证书
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -S app && adduser -S app -G app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/rst /app/rst

# 复制示例文件
COPY --from=builder /app/examples /app/examples

# 更改文件所有权
RUN chown -R app:app /app

# 切换到非 root 用户
USER app

# 设置入口点
ENTRYPOINT ["/app/rst"]

# 默认命令
CMD ["--help"]