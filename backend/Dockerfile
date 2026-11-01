# ==========================================
# 0. 前端构建阶段 (Web / Desktop 客户端)
# ==========================================
# 必须和 Go 二进制打进同一个镜像：mim-server 以 ./dist 为根目录托管 Web 客户端，
# 缺少产物时容器依然 healthy，但首页直接 404。
FROM docker.m.daocloud.io/library/node:20-alpine AS web-builder
WORKDIR /web

COPY clients/desktop-web/package.json clients/desktop-web/package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY clients/desktop-web/ ./
RUN npm run build

# ==========================================
# 1. Go 编译阶段
# ==========================================
FROM docker.m.daocloud.io/library/golang:1.26-alpine AS builder
WORKDIR /build

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# 编译所有微服务与单体二进制 (纯 Go 静态编译 CGO_ENABLED=0)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-gateway ./cmd/gateway && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-logic ./cmd/logic && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/mim-transfer ./cmd/transfer

# ==========================================
# 2. 生产运行镜像
# ==========================================
FROM docker.m.daocloud.io/library/alpine:3.19 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 从构建阶段拷贝全量可执行二进制
COPY --from=builder /build/bin/mim-server ./mim-server
COPY --from=builder /build/bin/mim-gateway ./mim-gateway
COPY --from=builder /build/bin/mim-logic ./mim-logic
COPY --from=builder /build/bin/mim-transfer ./mim-transfer

# 前端静态资源
COPY --from=web-builder /web/dist ./dist

# 目录与权限初始化
RUN mkdir -p uploads logs && \
    chmod +x ./mim-server ./mim-gateway ./mim-logic ./mim-transfer

EXPOSE 8080 8081 9091 9092

CMD ["./mim-server", "-port", "8080", "-db", "mim.db", "-storage", "uploads"]
