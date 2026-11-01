# ==========================================
# 1. 编译阶段 (Multi-Stage Builder)
# ==========================================
FROM golang:1.26-alpine AS builder
WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译所有微服务与单体二进制 (纯 Go 静态编译 CGO_ENABLED=0)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-gateway ./cmd/gateway && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-logic ./cmd/logic && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-transfer ./cmd/transfer

# ==========================================
# 2. 生产运行镜像 (Production Runner)
# ==========================================
FROM alpine:3.19 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 从构建阶段拷贝全量可执行二进制
COPY --from=builder /build/bin/mim-server ./mim-server
COPY --from=builder /build/bin/mim-gateway ./mim-gateway
COPY --from=builder /build/bin/mim-logic ./mim-logic
COPY --from=builder /build/bin/mim-transfer ./mim-transfer

# 前端静态资源：mim-server 以 ./dist 为根目录托管 Web 客户端
# (缺少这一步时 /app/dist 是空目录，首页直接 404)
COPY dist ./dist

# 目录与权限初始化
RUN mkdir -p uploads logs && \
    chmod +x ./mim-server ./mim-gateway ./mim-logic ./mim-transfer

EXPOSE 8080 8081 9091 9092

CMD ["./mim-server", "-port", "8080", "-db", "mim.db", "-storage", "uploads"]
