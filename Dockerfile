FROM golang:1.25-alpine AS builder

# 查看Go版本
RUN go version

# 为我们的镜像设置必要的环境变量
# 启用新的JSON实现和新的垃圾收集器
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOEXPERIMENT=jsonv2,greenteagc

# 移动到工作目录：/build
WORKDIR /build

# 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
COPY go.mod go.sum ./
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件，添加优化参数
#RUN go build -o agricultural_vision .
RUN go build -ldflags="-w -s" -o agricultural_vision .

###################
# 接下来创建一个小镜像
###################
#FROM debian:bookworm-slim
FROM alpine:3.22

# 设置工作目录
WORKDIR /app

# 复制脚本和配置文件到工作目录
COPY ./scripts/wait-for.sh ./wait-for.sh
COPY ./conf ./conf

# 从builder镜像中把可执行文件拷贝到当前工作目录
COPY --from=builder /build/agricultural_vision ./

# 更新包列表，安装 netcat-openbsd 和 ca-certificates，设置脚本权限
#RUN set -eux \
#    && apt-get update \
#    && apt-get install -y --no-install-recommends netcat-openbsd ca-certificates \
#    && update-ca-certificates \
#    && chmod 755 /app/wait-for.sh \
#    && ls -l /app/wait-for.sh \
#    && rm -rf /var/lib/apt/lists/*

# 安装必要的运行时依赖，包括CA证书
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add \
    ca-certificates \
    netcat-openbsd \
    tzdata \
    && update-ca-certificates

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# 设置入口点
ENTRYPOINT ["./agricultural_vision"]
CMD ["conf/config.yaml"]