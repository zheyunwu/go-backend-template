# 二开推荐阅读[如何提高项目构建效率](https://developers.weixin.qq.com/miniprogram/dev/wxcloudrun/src/scene/build/speed.html)
# 选择构建用基础镜像（选择原则：在包含所有用到的依赖前提下尽可能体积小）
FROM golang:1.24-alpine as builder

# 指定构建过程中的工作目录
WORKDIR /app

# 将go.mod和go.sum复制到工作目录
# 先复制依赖文件，利用Docker缓存机制提高构建速度
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将当前目录下所有文件都拷贝到工作目录下（.dockerignore中文件除外）
COPY . ./

# 执行代码编译命令
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/

# 选用运行时所用基础镜像
FROM alpine:3.18

# 容器默认时区为UTC，如果需要设置为其他时区，可以取消下面注释并修改时区
# RUN apk add --no-cache tzdata && \
#     cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime && \
#     echo Europe/Berlin > /etc/timezone

# 使用 HTTPS 协议访问容器云调用证书安装
RUN apk add --no-cache ca-certificates

# 指定运行时的工作目录
WORKDIR /app

# 将构建产物拷贝到运行时的工作目录中
COPY --from=builder /app/main ./
COPY --from=builder /app/config ./config

# 声明服务端口
EXPOSE 8080

# 设置环境变量
ENV APP_ENV=dev

# 执行启动命令，优先运行数据库迁移，然后启动服务器
CMD ["/app/main", "server"]
