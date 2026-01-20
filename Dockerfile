# 使用Alpine基础镜像，保持轻量（适合Koyeb免费阶层）
FROM golang:1.21-alpine AS builder

# 安装依赖
RUN apk add --no-cache git curl bash

# 下载最新Xray-core（动态获取最新版本）
RUN curl -s https://api.github.com/repos/XTLS/Xray-core/releases/latest | \
    grep "browser_download_url.*linux-64.zip" | cut -d '"' -f 4 | \
    xargs curl -L -o xray.zip && \
    unzip xray.zip -d /usr/bin/ && \
    chmod +x /usr/bin/xray && \
    rm xray.zip

# 复制Go源代码并构建
WORKDIR /app
COPY main.go .
RUN go mod init koyeb-reality-proxy && \
    go mod tidy && \
    go build -o proxy-bin main.go

# 最终镜像
FROM alpine:latest

# 复制Xray和Go binary
COPY --from=builder /usr/bin/xray /usr/bin/xray
COPY --from=builder /app/proxy-bin /usr/bin/proxy-bin

# 复制启动脚本和模板
COPY entrypoint.sh /entrypoint.sh
COPY config.json.template /config.json.template

# 设置权限
RUN chmod +x /entrypoint.sh /usr/bin/proxy-bin

# 暴露端口
EXPOSE 443

# 入口点
ENTRYPOINT ["/entrypoint.sh"]