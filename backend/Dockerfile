# ptis 后端：Go 1.25 + Gin + pgx。多阶段构建 → alpine 运行（wget 健康检查）。
FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

WORKDIR /app

# 先单独 copy go.mod / go.sum 利用 docker 层缓存
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# 静态链接（CGO 关闭）+ 去除调试信息 + 路径裁剪。同时构建 server 与 seed。
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/seed   ./cmd/seed


FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata wget \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /out/server /app/server
COPY --from=builder --chown=app:app /out/seed   /app/seed

USER app
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["/app/server"]
