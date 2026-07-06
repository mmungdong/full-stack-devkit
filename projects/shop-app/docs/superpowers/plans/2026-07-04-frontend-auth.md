# shop-app 前端登录注册 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `projects/shop-app/frontend` 新建 Next.js 静态导出前端，实现登录页 + 注册页连通后端；后端 API 统一加 `/api` 前缀；docker 单容器部署（Go embed 前端产物）。

**Architecture:** 前端复用根仓库 `frontend/template/web-nextjs` 模板（Next 16 + Tailwind 4 + axios + zod）连同 `frontend/shared` 包，自包含到 `projects/shop-app/{frontend,shared}`。后端 gin 所有业务路由挂到 `/api` group，`go:embed` 嵌入前端 `out/`，gin 同时服务静态文件 + API。Dockerfile 三阶段（node 构建 frontend → go 构建 embed 二进制 → runtime）。

**Tech Stack:** Next.js 16.2.9 / React 19.2.4 / TypeScript 5 / Tailwind 4 / axios 1.17 / zod 4.4 / zustand / shadcn/ui // Go 1.25 + gin + onexstack@v0.3.19 // Docker（node:20-alpine + golang:1.25 + debian:bookworm）

## Global Constraints

- **路径基准**：所有相对路径以 `projects/shop-app/` 为根。工作目录 `cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app` 后执行命令。
- **onexstack 版本锁定 v0.3.19**：禁止 `go get -u` / `@latest`，禁止删 `replace google.golang.org/grpc => v1.64.0`（见 `docs/conventions/pitfalls.md`）。
- **后端 API 前缀**：所有业务路由（healthz/login/refresh-token/v1/*）统一挂在 `/api` 下。
- **前端 baseURL**：`NEXT_PUBLIC_API_BASE_URL` 默认 `http://localhost:5555/api`。
- **静态导出**：`next.config.ts` 保持 `output: "export"`；前端路由用客户端软守卫（`useEffect` 重定向），无服务端中间件。
- **注释语言**：Go/TS 代码注释与现有代码库一致——Go 用中文，TS 模板用英文则保持英文。
- **Git 提交**：`projects/shop-app/` 被根仓库 `.gitignore` 排除。计划中的 commit 步骤是**逻辑检查点**：若主人在 `projects/shop-app/` 内单独 `git init`，则执行 commit；否则跳过 commit 步骤，仅做构建/测试验证。本计划默认跳过 commit，需提交时主人会指示。
- **MariaDB 运行中**：容器 `shop-mariadb`（admin/123456/shop-app）已启动，`127.0.0.1:3306` 可连。
- **后端服务**：现有 shop-apiserver 可能仍在后台运行（task `bg7z0m6qw`），改路由后需重启。

---

## File Structure

**新建文件：**
- `frontend/`（整个目录，从模板复制）—— 独立 npm 工程
- `shared/`（整个目录，从根仓库复制）—— `@devkit/shared` 本地包，frontend 依赖
- `frontend/src/stores/auth.ts` —— zustand 认证 store
- `frontend/src/schemas/auth.ts` —— 登录/注册 zod schema
- `frontend/src/app/login/page.tsx` —— 登录页
- `frontend/src/app/register/page.tsx` —— 注册页
- `frontend/src/components/ui/` —— shadcn/ui 组件（button/input/form/label/card）
- `frontend/src/lib/utils.ts` —— shadcn/ui 必需的 cn() 工具（若模板无）
- `internal/web/embed.go` —— `go:embed all:frontend/out`

**修改文件：**
- `internal/apiserver/httpserver.go` —— `InstallRESTAPI` 加 `/api` group；`InstallGenericAPI` 加静态托管 + SPA fallback
- `cmd/shop-apiserver/main.go` —— `@BasePath /api`
- `internal/apiserver/handler/user.go` —— `@Router` 注解加 `/api` 前缀
- `internal/apiserver/handler/healthz.go` —— `@Router` 注解加 `/api` 前缀
- `build/docker/shop-apiserver/Dockerfile` —— 三阶段改造
- `Makefile` —— 加 `web-build` / `build-all` 目标
- `frontend/package.json` —— `@devkit/shared` 路径改 `file:../shared`；加 zustand 依赖
- `frontend/.env.example` / `frontend/.env.local` —— baseURL 加 `/api`
- `frontend/src/api/client/index.ts` —— baseURL 拼接 `/api`；token 从 zustand 读
- `frontend/src/app/page.tsx` —— 首页占位 + 软守卫

---

## Task 1: 复制前端模板与 shared 包到 shop-app

**Files:**
- Create: `projects/shop-app/frontend/`（从 `frontend/template/web-nextjs` 复制，排除 node_modules/.next/out）
- Create: `projects/shop-app/shared/`（从 `frontend/shared` 复制，排除 node_modules）
- Modify: `projects/shop-app/frontend/package.json` —— `@devkit/shared` 路径改 `file:../shared`
- Modify: `projects/shop-app/frontend/.env.example`、`.env.local` —— baseURL 加 `/api`

**Interfaces:**
- Produces: 自包含的 `frontend/` npm 工程，`npm install` 可成功，`@devkit/shared` 解析到 `../shared`

- [ ] **Step 1: 复制模板（排除 node_modules/.next/out/package-lock.json）**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
rsync -a --exclude='node_modules' --exclude='.next' --exclude='out' --exclude='package-lock.json' --exclude='tsconfig.tsbuildinfo' \
  ../../frontend/template/web-nextjs/ frontend/
rsync -a --exclude='node_modules' --exclude='package-lock.json' \
  ../../frontend/shared/ shared/
```
Expected: `frontend/` 和 `shared/` 目录创建成功，`ls frontend/src` 显示 app/api/components 等。

- [ ] **Step 2: 修改 frontend/package.json 的 shared 依赖路径**

Edit `frontend/package.json`，把：
```json
"@devkit/shared": "file:../../shared",
```
改为：
```json
"@devkit/shared": "file:../shared",
```

- [ ] **Step 3: 修改 .env 的 baseURL 加 /api**

Edit `frontend/.env.example` 和 `frontend/.env.local`，把：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:5555
```
改为：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:5555/api
```

- [ ] **Step 4: 安装依赖**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/shared && npm install
cd ../frontend && npm install
```
Expected: 两个目录 `npm install` 成功，无报错。`node_modules/@devkit/shared` 软链到 `../shared`。

- [ ] **Step 5: 验证模板可构建**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm run build
```
Expected: `next build` 成功，生成 `out/` 目录，`ls out` 显示 `index.html`。

- [ ] **Step 6: 验证 typecheck**

Run:
```bash
npm run typecheck
```
Expected: 无类型错误。

---

## Task 2: 后端 API 加 /api 前缀 + Swagger 同步

**Files:**
- Modify: `internal/apiserver/httpserver.go` —— `InstallRESTAPI` 改用 `/api` group
- Modify: `cmd/shop-apiserver/main.go` —— `@BasePath /api`
- Modify: `internal/apiserver/handler/user.go` —— `@Router` 加 `/api`
- Modify: `internal/apiserver/handler/healthz.go` —— `@Router` 加 `/api`

**Interfaces:**
- Produces: 后端路由全部在 `/api/*` 下；Swagger 文档路径同步

- [ ] **Step 1: 改 httpserver.go 的 InstallRESTAPI 加 /api group**

Edit `internal/apiserver/httpserver.go`，把 `InstallRESTAPI` 函数体改为：

```go
// 注册 API 路由。所有业务接口统一挂在 /api 前缀下，与前端静态文件路由隔离.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口（pprof、404、静态文件等，不含 /api）
	InstallGenericAPI(engine)

	// 所有业务 API 统一 /api 前缀
	api := engine.Group("/api")

	// 认证和授权中间件
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}

	// 创建核心业务处理器
	hdl := handler.NewHandler(c.biz, c.val, authMiddlewares...)
	// 注册健康检查接口
	api.GET("/healthz", hdl.Healthz)
	// 注册用户登录和令牌刷新接口。这2个接口比较简单，所以没有 API 版本
	api.POST("/login", hdl.Login)
	// 注意：认证中间件要在 hdl.RefreshToken 之前加载
	api.PUT("/refresh-token", mw.AuthnMiddleware(c.retriever), hdl.RefreshToken)

	// 注册 v1 版本 API 路由分组（/api/v1）
	v1 := api.Group("/v1")
	// 注册资源路由
	hdl.InstallAll(v1)
}
```

- [ ] **Step 2: 改 main.go 的 @BasePath**

Edit `cmd/shop-apiserver/main.go`，把注解里的 `@BasePath /` 改为：
```go
// @BasePath        /api
```

- [ ] **Step 3: 改 handler/user.go 的 @Router 注解加 /api**

Edit `internal/apiserver/handler/user.go`，每条 `@Router` 前加 `/api`：
- `@Router /login [post]` → `@Router /api/login [post]`
- `@Router /refresh-token [put]` → `@Router /api/refresh-token [put]`
- `@Router /v1/users [post]` → `@Router /api/v1/users [post]`
- `@Router /v1/users/{userID}/change-password [put]` → `@Router /api/v1/users/{userID}/change-password [put]`
- `@Router /v1/users/{userID} [put]` → `@Router /api/v1/users/{userID} [put]`
- `@Router /v1/users/{userID} [delete]` → `@Router /api/v1/users/{userID} [delete]`
- `@Router /v1/users/{userID} [get]` → `@Router /api/v1/users/{userID} [get]`
- `@Router /v1/users [get]` → `@Router /api/v1/users [get]`

- [ ] **Step 4: 改 handler/healthz.go 的 @Router 注解**

Edit `internal/apiserver/handler/healthz.go`：
- `@Router /healthz [get]` → `@Router /api/healthz [get]`

- [ ] **Step 5: 重新生成 Swagger 文档**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
$(go env GOPATH)/bin/swag init -g main.go -d cmd/shop-apiserver,internal/apiserver/handler -o docs/apidocs --parseDependency --parseInternal 2>&1 | tail -5
```
Expected: 末尾显示 `create docs.go` / `create swagger.json`，无 error（warning 可忽略）。

- [ ] **Step 6: 构建后端**

Run:
```bash
make build BINS=shop-apiserver
```
Expected: 构建成功。

- [ ] **Step 7: 启动服务验证 /api 前缀**

先停掉旧服务（若运行中）：`pkill -f 'shop-apiserver -c' 2>/dev/null; sleep 1`

启动新服务（后台）：
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml
```
（用 run_in_background 启动，等待 4 秒）

验证：
```bash
curl -s -o /dev/null -w "GET /api/healthz -> HTTP %{http_code}\n" http://127.0.0.1:5555/api/healthz
curl -s -o /dev/null -w "GET /healthz (旧路径应404) -> HTTP %{http_code}\n" http://127.0.0.1:5555/healthz
```
Expected: `/api/healthz` 返回 200；`/healthz` 返回 404（JSON）。

- [ ] **Step 8: curl 复测登录注册**

```bash
# 注册
curl -s -X POST http://127.0.0.1:5555/api/v1/users -H "Content-Type: application/json" \
  -d '{"username":"apitest","password":"Admin@1234","nickname":"测试","email":"t@shop.app","phone":"13900000000"}'
# 登录
curl -s -X POST http://127.0.0.1:5555/api/login -H "Content-Type: application/json" \
  -d '{"username":"apitest","password":"Admin@1234"}'
```
Expected: 注册返回 `{"userID":"user-..."}`；登录返回含 `token` 的 JSON。

---

## Task 3: 前端 zustand auth store + zod schemas

**Files:**
- Create: `frontend/src/schemas/auth.ts` —— 登录/注册 zod schema
- Create: `frontend/src/stores/auth.ts` —— zustand 认证 store（persist）
- Modify: `frontend/package.json` —— 加 zustand 依赖

**Interfaces:**
- Produces: `useAuthStore`（含 `token`/`user`/`login`/`register`/`logout`/`isAuthenticated`），`loginSchema`/`registerSchema`

- [ ] **Step 1: 安装 zustand**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm install zustand
```
Expected: zustand 安装成功，`package.json` 出现 `zustand` 依赖。

- [ ] **Step 2: 创建 zod schemas**

Create `frontend/src/schemas/auth.ts`：

```ts
import { z } from "zod";

// 登录表单校验
export const loginSchema = z.object({
  username: z.string().min(3, "用户名至少 3 个字符"),
  password: z.string().min(8, "密码至少 8 个字符"),
});
export type LoginForm = z.infer<typeof loginSchema>;

// 注册表单校验（phone 必填，对齐后端）
export const registerSchema = z
  .object({
    username: z.string().min(3, "用户名至少 3 个字符"),
    password: z.string().min(8, "密码至少 8 个字符"),
    nickname: z.string().min(1, "昵称不能为空"),
    email: z.string().email("邮箱格式不正确"),
    phone: z
      .string()
      .min(1, "手机号不能为空")
      .regex(/^1[3-9]\d{9}$/, "手机号格式不正确"),
  });
export type RegisterForm = z.infer<typeof registerSchema>;

// API 响应 schema
export const loginResponseSchema = z.object({
  token: z.string(),
  expireAt: z.object({ seconds: z.number().optional(), nanos: z.number().optional() }).optional(),
});
export type LoginResponse = z.infer<typeof loginResponseSchema>;

export const registerResponseSchema = z.object({
  userID: z.string(),
});
export type RegisterResponse = z.infer<typeof registerResponseSchema>;
```

- [ ] **Step 3: 创建 zustand auth store**

Create `frontend/src/stores/auth.ts`：

```ts
"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { webPost } from "@/api";
import {
  loginSchema,
  registerSchema,
  loginResponseSchema,
  registerResponseSchema,
} from "@/schemas/auth";

interface AuthUser {
  userID: string;
  username: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  login: (username: string, password: string) => Promise<void>;
  register: (form: {
    username: string;
    password: string;
    nickname: string;
    email: string;
    phone: string;
  }) => Promise<string>;
  logout: () => void;
  isAuthenticated: () => boolean;
}

// 认证状态 store，token 持久化到 localStorage
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      login: async (username, password) => {
        loginSchema.parse({ username, password });
        const data = await webPost("/login", loginResponseSchema, { username, password });
        set({ token: data.token, user: { userID: "", username } });
      },
      register: async (form) => {
        registerSchema.parse(form);
        const data = await webPost("/v1/users", registerResponseSchema, form);
        return data.userID;
      },
      logout: () => set({ token: null, user: null }),
      isAuthenticated: () => get().token !== null,
    }),
    { name: "shop-app-auth" },
  ),
);
```

> ⚠️ `webPost` 来自模板 `src/api/request.ts`（web-bound 的 POST 帮助函数，带 zod 校验）。Step 4 会确认 `@/api` 确实 re-export 了 `webPost`；若没有，按 Step 4 修正 import 路径。

- [ ] **Step 4: 确认 webPost 导出路径**

模板 `src/api/request.ts` 导出 `webPost`，但 `src/api/index.ts` 是否 re-export 它需确认：

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
cat src/api/index.ts
```

若 `index.ts` 未 re-export `webPost`，则改 `src/stores/auth.ts` 的 import 为直接从 request 模块导入：
```ts
import { webPost } from "@/api/request";
```
若已 re-export，保持 `import { webPost } from "@/api"` 不变。

- [ ] **Step 5: typecheck 验证**

Run:
```bash
npm run typecheck
```
Expected: 无类型错误。若有，按报错修正 import 路径或 schema 类型。

---

## Task 4: 前端 axios client 接入 auth store token

**Files:**
- Modify: `frontend/src/api/client/index.ts` —— token 从 zustand 读（替代 localStorage 直接读）；baseURL 确认含 /api

**Interfaces:**
- Consumes: `useAuthStore`（Task 3）
- Produces: axios 请求自动带 Bearer token；401 自动 logout

- [ ] **Step 1: 改 client/index.ts 的 token 读取源**

Edit `frontend/src/api/client/index.ts`，把请求拦截器从 `localStorage.getItem("auth_token")` 改为读 zustand store。注意避免循环依赖（store 导入 client，client 不能直接导入 store——用动态读取）。

替换请求拦截器为：
```ts
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (typeof window !== "undefined") {
      // 从 localStorage 直接读持久化的 zustand state（避免循环依赖）
      const raw = localStorage.getItem("shop-app-auth");
      if (raw) {
        try {
          const parsed = JSON.parse(raw);
          const token = parsed?.state?.token;
          if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        } catch {
          // ignore parse error
        }
      }
    }
    return config;
  },
  (error: AxiosError) => Promise.reject(error),
);
```

> 说明：zustand persist 默认 key 为 `"shop-app-auth"`（Task 3 Step 3 配置），state 结构为 `{ state: { token, user }, version }`。直接读 localStorage 避免在 axios client 里导入 store 造成循环依赖。

- [ ] **Step 2: 改 401 响应拦截器清理 zustand state**

替换响应拦截器的 401 分支为：
```ts
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError<ApiErrorResponse>) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      // 清理持久化的 auth state
      localStorage.removeItem("shop-app-auth");
      // 跳登录页（避免在 client 层导入 router，用 location）
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  },
);
```

- [ ] **Step 3: 确认 baseURL 含 /api**

确认 `client/index.ts` 顶部：
```ts
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:5555/api";
```
（Task 1 已改 `.env`，这里默认值也加 `/api` 作为兜底）

- [ ] **Step 4: typecheck 验证**

Run:
```bash
npm run typecheck
```
Expected: 无类型错误。

---

## Task 5: shadcn/ui 初始化 + 登录页

**Files:**
- Create: `frontend/src/lib/utils.ts` —— cn() 工具（若不存在）
- Create: `frontend/src/components/ui/{button,input,label,card}.tsx` —— shadcn 组件
- Create: `frontend/src/app/login/page.tsx` —— 登录页
- Modify: `frontend/package.json` —— 加 shadcn 依赖（class-variance-authority/clsx/tailwind-merge/@radix-ui ）

**Interfaces:**
- Produces: `/login` 页面，提交调用 `useAuthStore.login`

- [ ] **Step 1: 安装 shadcn/ui 基础依赖**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm install class-variance-authority clsx tailwind-merge @radix-ui/react-label @radix-ui/react-slot
```
Expected: 依赖安装成功。

- [ ] **Step 2: 创建 lib/utils.ts（cn 工具）**

先确认是否已存在：
```bash
ls src/lib/utils.ts 2>/dev/null && echo "exists" || echo "missing"
```

若 missing，Create `frontend/src/lib/utils.ts`：
```ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

- [ ] **Step 3: 创建 shadcn ui 组件**

手动创建以下文件（shadcn/ui 标准组件源码，避免 init 交互）：

Create `frontend/src/components/ui/button.tsx`：
```tsx
import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: { default: "bg-primary text-primary-foreground hover:bg-primary/90", outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground", ghost: "hover:bg-accent hover:text-accent-foreground", destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90" },
      size: { default: "h-9 px-4 py-2", sm: "h-8 rounded-md px-3 text-xs", lg: "h-10 rounded-md px-8", icon: "h-9 w-9" },
    },
    defaultVariants: { variant: "default", size: "default" },
  },
);

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> { asChild?: boolean }

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />;
  },
);
Button.displayName = "Button";
export { Button, buttonVariants };
```

Create `frontend/src/components/ui/input.tsx`：
```tsx
import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, type, ...props }, ref) => (
    <input type={type} className={cn("flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50", className)} ref={ref} {...props} />
  ),
);
Input.displayName = "Input";
export { Input };
```

Create `frontend/src/components/ui/label.tsx`：
```tsx
"use client";
import * as React from "react";
import * as LabelPrimitive from "@radix-ui/react-label";
import { cn } from "@/lib/utils";

const Label = React.forwardRef<React.ElementRef<typeof LabelPrimitive.Root>, React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root>>(
  ({ className, ...props }, ref) => (<LabelPrimitive.Root ref={ref} className={cn("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", className)} {...props} />),
);
Label.displayName = LabelPrimitive.Root.displayName;
export { Label };
```

Create `frontend/src/components/ui/card.tsx`：
```tsx
import * as React from "react";
import { cn } from "@/lib/utils";

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({ className, ...props }, ref) => (<div ref={ref} className={cn("rounded-xl border bg-card text-card-foreground shadow", className)} {...props} />));
Card.displayName = "Card";
const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({ className, ...props }, ref) => (<div ref={ref} className={cn("flex flex-col space-y-1.5 p-6", className)} {...props} />));
CardHeader.displayName = "CardHeader";
const CardTitle = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({ className, ...props }, ref) => (<div ref={ref} className={cn("font-semibold leading-none tracking-tight", className)} {...props} />));
CardTitle.displayName = "CardTitle";
const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(({ className, ...props }, ref) => (<div ref={ref} className={cn("p-6 pt-0", className)} {...props} />));
CardContent.displayName = "CardContent";
export { Card, CardHeader, CardTitle, CardContent };
```

- [ ] **Step 4: 创建登录页**

Create `frontend/src/app/login/page.tsx`：
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth";
import { loginSchema } from "@/schemas/auth";

export default function LoginPage() {
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [form, setForm] = useState({ username: "", password: "" });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = loginSchema.safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "输入不合法");
      return;
    }
    setLoading(true);
    try {
      await login(form.username, form.password);
      router.push("/");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "登录失败";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>登录 Shop App</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input id="username" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input id="password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
            </div>
            {error && <p className="text-sm text-red-500">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "登录中..." : "登录"}
            </Button>
            <p className="text-center text-sm text-gray-500">
              还没有账号？<Link href="/register" className="text-blue-500 hover:underline">去注册</Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

- [ ] **Step 5: typecheck + 构建**

Run:
```bash
npm run typecheck && npm run build
```
Expected: 无类型错误；`out/login/index.html` 生成。

- [ ] **Step 6: 手工验证（dev 模式联调后端）**

启动后端（若未运行，见 Task 2 Step 7）。启动前端 dev：
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm run dev
```
浏览器访问 `http://localhost:3000/login`，用 Task 2 注册的 `apitest/Admin@1234` 登录。
Expected: 登录成功跳转 `/`（首页占位，可能报错因 page.tsx 还没改，先不管）。DevTools Application → Local Storage 看到 `shop-app-auth` 含 token。

---

## Task 6: 注册页 + 首页占位 + 软守卫

**Files:**
- Create: `frontend/src/app/register/page.tsx` —— 注册页
- Modify: `frontend/src/app/page.tsx` —— 首页占位 + 软守卫

**Interfaces:**
- Produces: `/register` 页面；`/` 首页（未登录跳 `/login`，已登录显示欢迎 + 登出）

- [ ] **Step 1: 创建注册页**

Create `frontend/src/app/register/page.tsx`：
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth";
import { registerSchema } from "@/schemas/auth";

export default function RegisterPage() {
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const [form, setForm] = useState({ username: "", password: "", nickname: "", email: "", phone: "" });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = registerSchema.safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "输入不合法");
      return;
    }
    setLoading(true);
    try {
      await register(form);
      router.push("/login");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "注册失败";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader><CardTitle>注册账号</CardTitle></CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input id="username" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input id="password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="nickname">昵称</Label>
              <Input id="nickname" value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <Input id="email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone">手机号</Label>
              <Input id="phone" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
            </div>
            {error && <p className="text-sm text-red-500">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "注册中..." : "注册"}
            </Button>
            <p className="text-center text-sm text-gray-500">
              已有账号？<Link href="/login" className="text-blue-500 hover:underline">去登录</Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

- [ ] **Step 2: 改首页 page.tsx 加软守卫 + 登出**

Read `frontend/src/app/page.tsx` 看现有内容，然后整体替换为：
```tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth";

export default function HomePage() {
  const router = useRouter();
  const { token, user, logout, isAuthenticated } = useAuthStore();

  // 客户端软守卫：未登录跳转 /login（静态导出下无服务端中间件）
  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
    }
  }, [isAuthenticated, router]);

  if (!token) return null;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 gap-4">
      <h1 className="text-2xl font-semibold">欢迎，{user?.username ?? "用户"}</h1>
      <p className="text-gray-500">登录成功，这是首页占位页。</p>
      <Button variant="outline" onClick={() => { logout(); router.push("/login"); }}>
        登出
      </Button>
    </div>
  );
}
```

- [ ] **Step 3: typecheck + 构建**

Run:
```bash
npm run typecheck && npm run build
```
Expected: 无错误；`out/register/index.html` 和 `out/index.html` 生成。

- [ ] **Step 4: 端到端手工验证（dev 模式）**

前端 `npm run dev` + 后端运行。浏览器流程：
1. 访问 `/` → 自动跳 `/login`
2. 点「去注册」→ `/register` 填表注册 → 成功跳 `/login`
3. 登录 → 跳 `/` 显示「欢迎，xxx」
4. 点「登出」→ 回 `/login`

Expected: 全流程通畅，localStorage 的 `shop-app-auth` 随登录/登出变化。

---

## Task 7: Go embed + gin 静态托管 + SPA fallback

**Files:**
- Create: `internal/web/embed.go` —— go:embed 前端产物
- Modify: `internal/apiserver/httpserver.go` —— `InstallGenericAPI` 加静态托管

**Interfaces:**
- Consumes: `frontend/out/`（Task 1/5/6 构建产物）
- Produces: Go 二进制含嵌入前端；gin 服务静态文件 + SPA fallback

- [ ] **Step 1: 先构建前端产物（embed 前提）**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm run build
```
Expected: `frontend/out/` 存在，含 `index.html`。

- [ ] **Step 2: 创建 internal/web/embed.go**

Create `internal/web/embed.go`：
```go
// Package web 提供前端静态文件的 go:embed 嵌入。
package web

import (
	"embed"
	"io/fs"
)

// Dist 嵌入前端构建产物 frontend/out。使用 all: 前缀确保 _next/ 等隐藏目录也被嵌入.
//
//go:embed all:frontend/out
var Dist embed.FS

// DistFS 返回前端构建产物的 fs.FS，供 gin 挂载静态文件.
func DistFS() fs.FS {
	sub, err := fs.Sub(Dist, "frontend/out")
	if err != nil {
		panic(err)
	}
	return sub
}
```

- [ ] **Step 3: 改 httpserver.go 加静态托管 + SPA fallback**

Edit `internal/apiserver/httpserver.go`：

1) 加 import：
```go
"net/http"
"strings"

"github.com/onexstack/shop-app/internal/web"
```

2) 替换 `InstallGenericAPI` 函数为：
```go
// InstallGenericAPI 注册业务无关的路由，例如 pprof、静态文件、404 处理等.
func InstallGenericAPI(engine *gin.Engine) {
	// 注册 pprof 路由
	pprof.Register(engine)

	// 注册 Swagger UI 路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 托管前端静态文件（go:embed 嵌入的 frontend/out）
	distFS := web.DistFS()
	fileServer := http.FileServer(http.FS(distFS))

	// SPA fallback：非 /api、非 /swagger 的请求交给静态文件服务器
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 路由未命中 → 返回 JSON 404
		if strings.HasPrefix(path, "/api/") {
			core.WriteResponse(c, errno.ErrPageNotFound, nil)
			return
		}
		// 尝试静态文件
		c.Request.URL.Path = path
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
```

> 说明：去掉了原 `NoRoute` 直接返回 JSON 404 的逻辑，改为按路径分流。`/api/*` 未命中返回 JSON（给前端 API 调用正确的 404），其他路径走静态文件。Next 静态导出 `trailingSlash: true`，`/login/` 会匹配 `out/login/index.html`。

- [ ] **Step 4: 构建 Go 二进制（embed 前提：frontend/out 已存在）**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
make build BINS=shop-apiserver
```
Expected: 构建成功（embed 把 `frontend/out` 嵌入二进制）。

- [ ] **Step 5: 验证静态托管 + API 共存**

停旧服务，启动新二进制：
```bash
pkill -f 'shop-apiserver -c' 2>/dev/null; sleep 1
```
（用 run_in_background 启动 `./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml`，等 4 秒）

验证：
```bash
curl -s -o /dev/null -w "前端首页 / -> HTTP %{http_code}\n" http://127.0.0.1:5555/
curl -s -o /dev/null -w "前端登录页 /login/ -> HTTP %{http_code}\n" http://127.0.0.1:5555/login/
curl -s -o /dev/null -w "API /api/healthz -> HTTP %{http_code}\n" http://127.0.0.1:5555/api/healthz
```
Expected: 全部 200。

- [ ] **Step 6: 浏览器端到端验证**

浏览器访问 `http://localhost:5555/login`（注意 trailingSlash，访问 `/login` 会 301 到 `/login/`）：
1. 登录页正常显示
2. 注册 → 登录 → 首页 全流程
Expected: 单端口 5555 同时服务前端 + API，全流程通畅。

---

## Task 8: Makefile 加 web-build / build-all

**Files:**
- Modify: `Makefile` —— 加 `web-build` / `build-all` 目标

**Interfaces:**
- Produces: `make web-build` 构建前端；`make build-all` 顺序构建前端+后端

- [ ] **Step 1: 查看 Makefile 结构找插入点**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
grep -nE "^build:|^build\.|^\.PHONY.*build" Makefile | head
```
找到 `build` 目标定义位置，在其附近加新目标。

- [ ] **Step 2: 加 web-build 和 build-all 目标**

在 Makefile 合适位置（如 `build` 目标之后）追加：
```makefile
## web-build: 构建前端静态产物到 frontend/out
web-build:
	@echo "===========> Building frontend"
	@cd frontend && npm run build

## build-all: 先构建前端，再构建 Go 二进制（embed 依赖前端产物）
build-all: web-build
	@$(MAKE) build BINS=shop-apiserver
```

并确保 `.PHONY` 行包含 `web-build build-all`（找到已有 `.PHONY` 追加，或新建）。

- [ ] **Step 3: 验证 make build-all**

Run:
```bash
make build-all
```
Expected: 先执行 `npm run build`，再 `make build`，最终二进制生成。

- [ ] **Step 4: 验证产物可运行**

```bash
pkill -f 'shop-apiserver -c' 2>/dev/null; sleep 1
```
（run_in_background 启动新二进制）
```bash
curl -s -o /dev/null -w "/ -> %{http_code}\n" http://127.0.0.1:5555/
curl -s -o /dev/null -w "/api/healthz -> %{http_code}\n" http://127.0.0.1:5555/api/healthz
```
Expected: 均 200。

---

## Task 9: Dockerfile 三阶段改造

**Files:**
- Modify: `build/docker/shop-apiserver/Dockerfile` —— 改三阶段

**Interfaces:**
- Produces: 单镜像，含 embed 前端的 Go 二进制 + 配置文件

- [ ] **Step 1: 备份现有 Dockerfile**

Run:
```bash
cp build/docker/shop-apiserver/Dockerfile build/docker/shop-apiserver/Dockerfile.bak
```

- [ ] **Step 2: 重写 Dockerfile 为三阶段**

Write `build/docker/shop-apiserver/Dockerfile`：
```dockerfile
# syntax=docker/dockerfile:1.7

# 0) Build args
ARG USER=noroot
ARG UID=65532
ARG GID=65532

# =============================================================================
# Stage 1: 前端构建（Next.js 静态导出）
# =============================================================================
FROM node:20-alpine AS frontend
WORKDIR /web

# 先拷依赖清单，利用层缓存
COPY shared/ /web/shared/
COPY frontend/package*.json frontend/package-lock.json* /web/frontend/
# shared 也需 install
RUN cd /web/shared && npm install
RUN cd /web/frontend && npm install

# 拷前端源码并构建
COPY frontend/ /web/frontend/
RUN cd /web/frontend && npm run build
# 产物在 /web/frontend/out

# =============================================================================
# Stage 2: Go 构建（embed 前端产物）
# =============================================================================
FROM golang:1.25 AS builder
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
```

- [ ] **Step 3: 准备 .dockerignore（避免把 node_modules/out 拷进构建上下文）**

检查/创建 `projects/shop-app/.dockerignore`：
```
frontend/node_modules
frontend/.next
frontend/out
shared/node_modules
_output
.git
*.test.go
docs/apidocs
```
> 说明：frontend/out 必须忽略，确保 Stage 2 用的是 Stage 1 构建的新产物，而非本地残留。

- [ ] **Step 4: 构建镜像**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
docker build -f build/docker/shop-apiserver/Dockerfile -t shop-apiserver:dev .
```
Expected: 三阶段构建成功（耗时较长，前端 npm install + go build）。

> ⚠️ 若构建机无 docker（docker 在 Windows），此步需主人在 Windows 终端执行。浮浮酱会提示。

---

## Task 10: docker 端到端验收

**Files:** 无（验证任务）

- [ ] **Step 1: 运行容器（连宿主 MariaDB）**

容器内连 MariaDB 需用宿主可达地址。Docker Desktop 用 `host.docker.internal`：
```bash
docker run -d --name shop-app-test -p 5555:5555 \
  -e MYSQL_HOST=host.docker.internal \
  shop-apiserver:dev
```

> ⚠️ 配置文件里 `mysql.addr: 127.0.0.1:3306` 在容器内指容器自身，连不到宿主 MariaDB。需改配置或用环境变量覆盖。最简方案：构建前把 `configs/shop-apiserver.yaml` 的 `mysql.addr` 改为 `host.docker.internal:3306`（Docker Desktop）或用 `--add-host=host.docker.internal:host-gateway`（Linux）。验收时按实际环境调整。

- [ ] **Step 2: 验证容器服务**

Run:
```bash
docker logs shop-app-test | tail -20
curl -s -o /dev/null -w "/ -> %{http_code}\n" http://127.0.0.1:5555/
curl -s -o /dev/null -w "/api/healthz -> %{http_code}\n" http://127.0.0.1:5555/api/healthz
```
Expected: 日志显示服务启动；两个请求都 200。

- [ ] **Step 3: 浏览器端到端验收**

浏览器 `http://localhost:5555/login`：
1. 注册新用户 → 跳 `/login`
2. 登录 → 跳 `/` 显示欢迎
3. 登出 → 回 `/login`
4. Swagger UI `http://localhost:5555/swagger/index.html` 接口可调试（路径含 `/api`）

Expected: 全部验收点通过（见 spec §9 验收标准 1-7）。

- [ ] **Step 4: 清理（可选）**

```bash
docker rm -f shop-app-test
```

---

## 验收标准对照（spec §9）

1. ✅ `cd frontend && npm run build` 产出 `out/` —— Task 1/5/6 验证
2. ✅ `make build-all` 产出含前端的 Go 二进制 —— Task 8
3. ✅ `docker build` 成功，单容器启动 —— Task 9/10
4. ✅ `/register` 注册成功 → 跳 `/login` —— Task 6/10
5. ✅ `/login` 登录成功 → 跳 `/` 显示用户信息 —— Task 5/6/10
6. ✅ 登出 → 回 `/login` —— Task 6/10
7. ✅ Swagger UI（`/api` 前缀）可调试 —— Task 2
