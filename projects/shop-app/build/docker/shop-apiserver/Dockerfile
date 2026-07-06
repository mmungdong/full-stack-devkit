# syntax=docker/dockerfile:1.7

# 0) 构建参数
ARG USER=noroot
ARG UID=65532
ARG GID=65532

# =============================================================================
# Stage 1: 前端构建（Next.js 静态导出）
# =============================================================================
FROM node:20-alpine AS frontend
WORKDIR /web

# 先拷 shared 和依赖清单，利用层缓存
COPY shared/ /web/shared/
COPY frontend/package.json frontend/package-lock.json* /web/frontend/

# 安装 shared 和 frontend 依赖
RUN cd /web/shared && npm install
RUN cd /web/frontend && npm install

# 拷前端源码并构建
COPY frontend/ /web/frontend/
RUN cd /web/frontend && npm run build
# 产物在 /web/frontend/out

# =============================================================================
# Stage 2: Go 构建（embed 前端产物）
# =============================================================================
FROM golang:1.26 AS builder
ARG OS=linux
ARG ARCH=amd64

WORKDIR /workspace

# tini（与原 Dockerfile 一致）
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL -o /usr/bin/tini https://github.com/krallin/tini/releases/download/v0.19.0/tini-static-amd64 && chmod +x /usr/bin/tini

# Go 依赖缓存
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# 拷源码
COPY . .

# 前端产物就位（供 go:embed）
COPY --from=frontend /web/frontend/out ./frontend/out

ENV CGO_ENABLED=1 GOOS=${OS} GOARCH=${ARCH} GO111MODULE=on GOCACHE=/root/.cache/go-build GOMODCACHE=/go/pkg/mod
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build BINS=shop-apiserver

# =============================================================================
# Stage 3: runtime
# =============================================================================
FROM debian:bookworm AS runtime
ARG USER
ARG UID
ARG GID

WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget curl && rm -rf /var/lib/apt/lists/*
RUN groupadd -g ${GID} ${USER} 2>/dev/null || true && useradd -u ${UID} -g ${USER} ${USER} 2>/dev/null || true

COPY --from=builder --chown=0:0 /usr/bin/tini /usr/bin/tini
COPY --from=builder --chown=${UID}:${GID} /workspace/_output/platforms/linux/amd64/shop-apiserver /app/shop-apiserver
COPY --from=builder --chown=${UID}:${GID} /workspace/configs/shop-apiserver.yaml /app/configs/shop-apiserver.yaml

USER ${UID}:${GID}
ENTRYPOINT ["/usr/bin/tini", "--", "/app/shop-apiserver", "-c", "/app/configs/shop-apiserver.yaml"]
