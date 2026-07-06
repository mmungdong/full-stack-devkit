# shop-app 前端登录注册设计

- **日期**：2026-07-04
- **范围**：在 `projects/shop-app/frontend` 新建 Next.js 静态导出前端，实现登录页 + 注册页；后端 API 统一加 `/api` 前缀；docker 单容器部署（Go embed 前端产物）
- **状态**：已确认，待生成实施计划

---

## 1. 背景与目标

shop-app 后端（gin + MariaDB）已跑通用户登录注册接口。现需补前端，先把登录注册页搞定。

目标：
- 在 `projects/shop-app/` 下新建 `frontend/` 目录，写 Next.js 前端，做静态导出
- 只写前端代码，后端继续用 Go
- 实现登录页 + 注册页，连通后端
- docker 单容器部署：Go 二进制 embed 前端产物，同时服务静态文件 + API

非目标（本次不做）：
- 商品、订单等业务页面
- Redis 缓存层（后续迭代）
- 后端 API 的功能扩展

---

## 2. 技术选型

### 2.1 构建工具决策

主人原问「Next.js 用 Vite 还是原生构建」——**Next.js 只能用其自带构建器（`next build`，Webpack/Turbopack），不能用 Vite**。Vite 是独立工具链，搭配纯 React SPA。选了 Next.js 静态导出，构建工具即 `next build`，无 Vite 选项。

### 2.2 选型矩阵

| 维度 | 决策 | 依据 |
|------|------|------|
| 框架 | Next.js 16（App Router） | 主人指定，静态导出原生支持 |
| 构建工具 | `next build --webpack` | Next 自带，模板已用 |
| 静态导出 | `output: "export"` | 产出纯静态 `out/` |
| 起点 | 复用根仓库 `frontend/template/web-nextjs` | 已有 axios/zod/Tailwind 4 分层，省搭架工作 |
| 状态管理 | zustand（persist 中间件） | 主人指定，token 持久化 |
| UI | shadcn/ui | 主人指定，基于 Radix + Tailwind |
| 校验 | Zod v4 | 模板已有 |
| 部署 | docker 单容器，Go embed 前端 | 主人指定，单二进制 |
| 后端路由 | 所有 API 统一加 `/api` 前缀 | 前后端路由隔离，避免冲突 |

### 2.3 技术栈版本（对齐模板）

Next.js 16.2.9 / React 19.2.4 / TypeScript 5 / Tailwind 4 / axios 1.17 / zod 4.4

---

## 3. 目录结构与模块边界

```
projects/shop-app/
├── frontend/                      # 前端（独立 npm 工程）
│   ├── package.json
│   ├── next.config.ts             # output: "export"
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx           # 首页（登录后跳转目标，占位）
│   │   │   ├── login/page.tsx     # 登录页
│   │   │   └── register/page.tsx  # 注册页
│   │   ├── components/ui/         # shadcn/ui 组件
│   │   ├── api/                   # axios 封装 + 认证拦截器
│   │   ├── stores/auth.ts         # zustand 认证 store
│   │   ├── schemas/               # zod 校验
│   │   └── lib/ utils/
│   └── out/                       # next build 产物（gitignore）
├── internal/apiserver/
│   └── httpserver.go              # 路由加 /api 前缀 + 静态托管
├── internal/web/
│   └── embed.go                   # go:embed frontend/out（新增包）
└── build/docker/shop-apiserver/
    └── Dockerfile                 # 三阶段：node + go + runtime
```

### 模块边界

- `frontend/` 是独立 npm 工程，与 Go 零编译期耦合，可单独 `npm run dev`
- `internal/web/` 是 Go 侧唯一的前端嵌入点，只有它知道静态文件位置
- `httpserver.go` 通过 `internal/web` 拿嵌入 FS，gin 挂载——不直接碰 `frontend/`
- `go:embed all:frontend/out` 要求构建前 `out/` 必须存在，由 Makefile 保证顺序

---

## 4. 前端登录注册数据流

### 4.1 API 契约（加 /api 前缀后）

| 接口 | 方法 | 入参 | 出参 |
|------|------|------|------|
| `/api/login` | POST | `{username, password}` | `{token, expireAt}` |
| `/api/v1/users` | POST | `{username, password, nickname, email, phone}` | `{userID}` |
| `/api/refresh-token` | PUT | （无，带旧 token） | `{token, expireAt}` |

### 4.2 zustand auth store（`src/stores/auth.ts`）

```ts
interface AuthState {
  token: string | null
  expireAt: number | null
  user: { userID: string; username: string } | null
  login: (username, password) => Promise<void>   // 调 /api/login，存 token
  register: (form) => Promise<string>            // 调 /api/v1/users，返回 userID
  logout: () => void
  isAuthenticated: () => boolean
}
```

- token 通过 zustand persist 中间件存 `localStorage`，刷新页面不丢
- 启动时从 localStorage rehydrate

### 4.3 axios 拦截器（`src/api/client/`）

- **请求拦截**：从 auth store 读 token，加 `Authorization: Bearer <token>`
- **响应拦截**：401 → `logout()` + 跳 `/login`；400 → 抛出后端 `message`；5xx → 统一提示
- **baseURL**：`NEXT_PUBLIC_API_BASE_URL`（默认 `http://localhost:5555/api`）

### 4.4 页面流程

**登录页 `/login`**：
1. 表单：username + password（zod 校验非空 + 长度）
2. 提交 → `auth.login()` → 成功跳 `/`，失败显示错误
3. 底部链接「去注册」→ `/register`

**注册页 `/register`**：
1. 表单：username + password + nickname + email + phone（zod 校验，phone 必填对齐后端）
2. 提交 → `auth.register()` → 成功跳 `/login` 并提示「注册成功」
3. 底部链接「去登录」→ `/login`

**首页 `/`（占位）**：
- 未登录 → 客户端重定向 `/login`
- 已登录 → 显示「欢迎 + userID」+ 登出按钮

### 4.5 静态导出下的路由守卫

`output: "export"` 不能用 Next 服务端中间件/重定向，采用**软守卫**（纯客户端）：
- 页面顶层 `useEffect` 检查 `auth.isAuthenticated()`，未登录 `router.replace('/login')`
- 安全性靠后端 API 鉴权保证，前端只管体验

---

## 5. Go 侧静态托管与后端 /api 前缀

### 5.1 后端路由加 /api 前缀

`InstallRESTAPI` 改为挂到 `api := engine.Group("/api")`：

```go
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
    InstallGenericAPI(engine)              // pprof、404、静态文件（不含 /api）

    api := engine.Group("/api")            // 所有业务 API 统一前缀
    authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}
    hdl := handler.NewHandler(c.biz, c.val, authMiddlewares...)

    api.GET("/healthz", hdl.Healthz)
    api.POST("/login", hdl.Login)
    api.PUT("/refresh-token", mw.AuthnMiddleware(c.retriever), hdl.RefreshToken)

    v1 := api.Group("/v1")                 // → /api/v1
    hdl.InstallAll(v1)
}
```

最终后端路径：`/api/healthz`、`/api/login`、`/api/refresh-token`、`/api/v1/users`...

### 5.2 Swagger 同步

- `cmd/shop-apiserver/main.go` 顶部注解 `@BasePath /api`
- 各 handler 的 `@Router` 注解加 `/api` 前缀（如 `@Router /api/login [post]`）
- 重新 `swag init` 生成 `docs/apidocs/`

### 5.3 Go embed 前端（`internal/web/embed.go`）

```go
package web

import (
    "embed"
    "io/fs"
)

//go:embed all:frontend/out
var Dist embed.FS

// DistFS 返回前端构建产物的 fs.FS
func DistFS() fs.FS {
    sub, _ := fs.Sub(Dist, "frontend/out")
    return sub
}
```

`all:frontend/out` 确保隐藏文件（`_next/`）也被嵌入。

### 5.4 gin 静态托管 + SPA fallback（`httpserver.go` `InstallGenericAPI`）

- 请求路径匹配 `/api/*` → 走 API 路由（未命中返回 JSON 404）
- 其他路径 → 尝试静态文件，找不到回退 `index.html`（SPA history mode）
- 实现用 `gin-contrib/static` 或自写 fallback

### 5.5 Makefile 协调

新增目标：
- `web-build`：`cd frontend && npm run build`
- `build-all`：先 `web-build` 再 `make build BINS=shop-apiserver`（保证 embed 前 `out/` 就绪）

---

## 6. Docker 单容器构建

### 6.1 Dockerfile 三阶段（`build/docker/shop-apiserver/Dockerfile`）

```dockerfile
# Stage 1: 前端构建
FROM node:20-alpine AS frontend
WORKDIR /web
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build          # 产物 /web/out

# Stage 2: Go 构建（含前端产物 embed）
FROM golang:1.25 AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /web/out ./frontend/out
RUN make build BINS=shop-apiserver

# Stage 3: runtime
FROM debian:bookworm
COPY --from=builder /workspace/_output/platforms/linux/amd64/shop-apiserver /app/shop-apiserver
COPY --from=builder /workspace/configs/shop-apiserver.yaml /app/configs/shop-apiserver.yaml
ENTRYPOINT ["/usr/bin/tini","--","/app/shop-apiserver","-c","/app/configs/shop-apiserver.yaml"]
```

### 6.2 容器内 MariaDB 连接

容器内 `127.0.0.1` 指向容器自身，需改 `configs/shop-apiserver.yaml` 的 `mysql.addr`：
- Linux 原生 docker：用宿主 IP 或 `--network host`
- Docker Desktop：用 `host.docker.internal:3306`
- 推荐用 docker compose 网络或环境变量覆盖，避免硬编码

### 6.3 镜像产物

单二进制 + 配置文件，Go 同时服务前端静态文件 + `/api` 后端接口，单容器单端口（5555）。

---

## 7. 错误处理

### 7.1 前端

- **API 层**（axios 响应拦截器）：
  - 网络错误 / 5xx → 统一提示「服务异常，请稍后重试」
  - 400 → 显示后端 `message`（如 `phone cannot be empty`）
  - 401 → `logout()` + 跳 `/login`
  - 409（用户名已存在）→ 注册页提示「用户名已被占用」
- **表单层**（zod + shadcn/ui Form）：字段级校验，失焦/提交触发，字段下显示错误
- **页面层**：提交按钮 loading 态防重复

### 7.2 后端

后端已有 `core.WriteResponse` + `errno` 体系，不改。加 `/api` 前缀后错误响应格式不变。

---

## 8. 测试策略

### 8.1 前端

- 单元测试：auth store 状态变更、zod schema 校验
- 组件测试：登录/注册表单提交（React Testing Library，初版可选跳过）
- 手工联调：`npm run dev` + 后端，走完整登录注册流程

### 8.2 后端

- 加 `/api` 前缀后用 curl 复测 `/api/login`、`/api/v1/users`
- 不新增 Go 测试（路由层改动靠联调验证）

### 8.3 集成（docker）

`docker build` 出单镜像 → 启动 → 浏览器 `/register` 注册 → `/login` 登录 → 跳首页，验证前端调通 `/api/*`。

---

## 9. 验收标准

1. ✅ `cd frontend && npm run build` 产出 `out/`，无报错
2. ✅ `make build-all` 产出含前端的 Go 二进制
3. ✅ `docker build` 成功，单容器启动
4. ✅ 浏览器 `/register` 注册成功 → 跳 `/login`
5. ✅ `/login` 登录成功 → 跳 `/` 显示用户信息
6. ✅ 登出 → 回 `/login`
7. ✅ Swagger UI（`/api` 前缀）接口可调试

---

## 10. 实施顺序（粗略，详细计划由 writing-plans 生成）

1. 复制模板到 `projects/shop-app/frontend`，清理模板无关内容
2. 后端路由加 `/api` 前缀 + Swagger 注解同步 + swag init + 复测
3. 前端：装 zustand + shadcn/ui，搭 auth store + axios 拦截器
4. 前端：登录页 + 注册页 + 首页占位 + 软守卫
5. Go embed 包 + gin 静态托管 + SPA fallback
6. Makefile 加 web-build / build-all
7. Dockerfile 三阶段改造
8. docker build + 端到端联调验收
