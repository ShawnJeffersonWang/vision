FROM golang:alpine AS builder

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件
RUN go build -o agricultural_vision .

###################
# 接下来创建一个小镜像
###################
FROM debian:bookworm-slim

# 设置工作目录
WORKDIR /app

# 复制脚本和配置文件到工作目录
COPY ./wait-for.sh /app/wait-for.sh
COPY ./conf /app/conf

# 从builder镜像中把可执行文件拷贝到当前工作目录
COPY --from=builder /build/agricultural_vision /app/

# 更新包列表，安装 netcat-openbsd，设置脚本权限
RUN set -eux \
    && apt-get update \
    && apt-get install -y --no-install-recommends netcat-openbsd \
    && chmod 755 /app/wait-for.sh \
    && ls -l /app/wait-for.sh

# 设置入口点
ENTRYPOINT ["/app/agricultural_vision"]
CMD ["conf/config.yaml"]