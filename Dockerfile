# 使用官方的 Golang 镜像作为构建环境
FROM golang:1.22.3-alpine3.20 as builder

# 设置工作目录
WORKDIR /app

# 将 go.mod 和 go.sum 文件复制到工作目录
COPY go.mod go.sum ./

# 下载所有依赖
RUN go mod download

# 将源代码复制到工作目录
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用 scratch 作为最小化的运行环境
FROM scratch

# 从 builder 阶段复制二进制文件和其他必要的文件到当前阶段
COPY --from=builder /app/main .

# 暴露端口，这应该与你的应用实际使用的端口一致
EXPOSE 8888

# 运行应用
CMD ["./main"]