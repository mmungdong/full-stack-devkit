# shop-app 前端样式重构与 i18n 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 shop-app 前端登录/注册/首页为暗色星尘风（CSS Modules），全站接入 react-i18next（zh/en 可扩展），后端新增 `/api/config` 提供默认语言，样式与 i18n 规范写入 CLAUDE.md。

**Architecture:** 前端 globals.css 作唯一全局入口（Tailwind + 暗色变量 + 星尘动画），各页面用 CSS Modules（单文件 + @media 分 PC/移动）。react-i18next 客户端初始化，默认语言优先级 localStorage → 后端 /api/config → zh。后端在 Config 加 DefaultLanguage 字段，yaml 配置 i18n.defaultLanguage。zod schema 改函数式 `makeXxxSchema(t)`，校验移到组件层。

**Tech Stack:** Next.js 16.2.9 / React 19 / TypeScript 5 / Tailwind 4 / CSS Modules / react-i18next / zustand // Go 1.26 + gin + onexstack@v0.3.19

## Global Constraints

- **路径基准**：相对路径以 `projects/shop-app/` 为根，命令在 `cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app` 后执行。
- **onexstack 锁定 v0.3.19**：禁止 `go get -u`/`@latest`，禁止删 `replace google.golang.org/grpc => v1.64.0`（见 `docs/conventions/pitfalls.md`）。
- **后端 API 前缀**：所有业务路由在 `/api` 下，新增 `/api/config` 同级。
- **样式约定**：全局唯一 `src/app/globals.css`；组件用 CSS Modules（`xxx.module.css` 同目录）；PC/移动用 `@media (min-width:768px)` 分区，移动优先；class 语义化。
- **i18n 约定**：react-i18next，语言文件 `src/locales/{lang}/common.json`；默认语言优先级 `localStorage["shop-app-lang"]` → `/api/config` defaultLanguage → `zh`；所有可见文案走 `t()`，禁止硬编码中文。
- **dev 脚本**：`next dev --webpack`（Turbopack 解析不了 @devkit/shared TS 源码，已改）。
- **Go embed**：前端产物在 `frontend/out/`，embed.go 在 `frontend/embed.go`（`//go:embed all:out`）。改前端后需 `npm run build` 再 `make build`。
- **Git 提交**：`projects/shop-app/` 被根仓库 gitignore 排除，commit 步骤跳过（除非主人在 shop-app 内 git init）。
- **MariaDB**：容器 `shop-mariadb`（admin/123456/shop-app）运行中。
- **后端服务**：可能在跑，改后端后需重启。docker 镜像 `shop-apiserver:dev` 已构建。

---

## File Structure

**后端新建：**
- `internal/apiserver/handler/config.go` —— GetConfig handler + ConfigResponse struct

**后端修改：**
- `configs/shop-apiserver.yaml` —— 加 `i18n.defaultLanguage: zh`
- `cmd/shop-apiserver/app/options/options.go` —— ServerOptions 加 `DefaultLanguage` 字段 + AddFlags
- `internal/apiserver/server.go` —— Config 加 `DefaultLanguage` 字段
- `cmd/shop-apiserver/app/options/options.go` 的 `Config()` 方法 —— 传递 DefaultLanguage
- `internal/apiserver/httpserver.go` —— InstallRESTAPI 注册 `api.GET("/config", ...)`；GetConfig handler（内联或调 handler）
- `internal/apiserver/handler/config.go` —— Swagger 注解 `@Router /api/config [get]`
- 重新 `swag init` → `docs/apidocs/`

**前端新建：**
- `src/i18n/config.ts` —— i18next 初始化
- `src/i18n/client.ts` —— initI18n + I18nBootstrap 组件
- `src/locales/zh/common.json` —— 中文文案
- `src/locales/en/common.json` —— 英文文案
- `src/components/LanguageSwitcher.tsx` + `LanguageSwitcher.module.css` —— 语言切换器
- `src/app/login/login.module.css` —— 登录页样式
- `src/app/register/register.module.css` —— 注册页样式
- `src/app/page.module.css` —— 首页样式

**前端修改：**
- `src/app/globals.css` —— 加暗色变量 + 星尘动画
- `src/app/layout.tsx` —— 包裹 I18nBootstrap
- `src/app/login/page.tsx` —— 重构结构 + i18n + module.css
- `src/app/register/page.tsx` —— 重构 + i18n + module.css
- `src/app/page.tsx` —— 首页 i18n + module.css
- `src/schemas/auth.ts` —— 改函数式 `makeLoginSchema(t)` / `makeRegisterSchema(t)`
- `src/stores/auth.ts` —— 移除 parse（校验移到组件层）
- `src/api/client/index.ts` —— 无需改（baseURL 已 /api）
- `package.json` —— 装 react-i18next / i18next

**文档修改：**
- `CLAUDE.md` —— 加「前端样式与 i18n 规范」一节

---

## Task 1: 后端 /api/config 接口

**Files:**
- Modify: `configs/shop-apiserver.yaml`
- Modify: `cmd/shop-apiserver/app/options/options.go`
- Modify: `internal/apiserver/server.go`（Config struct）
- Create: `internal/apiserver/handler/config.go`
- Modify: `internal/apiserver/httpserver.go`（注册路由）
- Regenerate: `docs/apidocs/`

**Interfaces:**
- Produces: `GET /api/config` 返回 `{"defaultLanguage":"zh"}`；`Config.DefaultLanguage` 字段供 handler 读取

- [ ] **Step 1: yaml 加 i18n 配置**

Edit `configs/shop-apiserver.yaml`，在末尾追加：
```yaml

# i18n 国际化配置
i18n:
  defaultLanguage: zh # 默认语言（zh/en）
```

- [ ] **Step 2: ServerOptions 加 DefaultLanguage 字段**

Edit `cmd/shop-apiserver/app/options/options.go`，在 `ServerOptions` struct 的 `SlogOptions` 字段后加：
```go
	// I18nOptions 包含国际化配置.
	I18nOptions *I18nOptions `json:"i18n" mapstructure:"i18n"`
```

在 `NewServerOptions` 函数的 `SlogOptions: genericoptions.NewSlogOptions(),` 行后加：
```go
		I18nOptions: NewI18nOptions(),
```

在文件末尾（`Config()` 方法后）追加 I18nOptions 定义：
```go
// I18nOptions 包含国际化相关配置.
type I18nOptions struct {
	// DefaultLanguage 定义默认语言（如 zh、en）.
	DefaultLanguage string `json:"defaultLanguage" mapstructure:"defaultLanguage"`
}

// NewI18nOptions 创建带默认值的 I18nOptions.
func NewI18nOptions() *I18nOptions {
	return &I18nOptions{
		DefaultLanguage: "zh",
	}
}

// AddFlags 绑定 i18n 选项到命令行标志.
func (o *I18nOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DefaultLanguage, "i18n.default-language", o.DefaultLanguage, "Default language for i18n (e.g. zh, en).")
}
```

在 `AddFlags` 方法的 `o.SlogOptions.AddFlags(fs, "slog")` 行后加：
```go
	o.I18nOptions.AddFlags(fs)
```

- [ ] **Step 3: Config struct 加 DefaultLanguage**

Edit `internal/apiserver/server.go`，在 `Config` struct 的 `MySQLOptions` 字段后加：
```go
	// DefaultLanguage 默认语言（i18n）.
	DefaultLanguage string
```

- [ ] **Step 4: opts.Config() 传递 DefaultLanguage**

Edit `cmd/shop-apiserver/app/options/options.go` 的 `Config()` 方法，在 `MySQLOptions: o.MySQLOptions,` 行后加：
```go
		DefaultLanguage: o.I18nOptions.DefaultLanguage,
```

- [ ] **Step 5: 创建 config handler**

Create `internal/apiserver/handler/config.go`：
```go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
)

// ConfigResponse 全局配置响应.
type ConfigResponse struct {
	// DefaultLanguage 默认语言.
	DefaultLanguage string `json:"defaultLanguage"`
}

// GetConfig 返回前端全局配置.
//
// @Summary      获取全局配置
// @Description  返回前端可用的全局配置（如默认语言）
// @Tags         系统
// @Produce      json
// @Success      200  {object}  ConfigResponse
// @Router       /api/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	core.WriteResponse(c, ConfigResponse{
		DefaultLanguage: h.defaultLanguage,
	}, nil)
}
```

- [ ] **Step 6: Handler 加 defaultLanguage 字段**

Edit `internal/apiserver/handler/handler.go`：

1) struct 加字段：
```go
type Handler struct {
	biz             biz.IBiz
	val             *validation.Validator
	mws             []gin.HandlerFunc
	defaultLanguage string
}
```

2) NewHandler 加参数：
```go
func NewHandler(biz biz.IBiz, val *validation.Validator, defaultLanguage string, mws ...gin.HandlerFunc) *Handler {
	return &Handler{biz: biz, val: val, defaultLanguage: defaultLanguage, mws: mws}
}
```

- [ ] **Step 7: httpserver.go 注册路由并传 defaultLanguage**

Edit `internal/apiserver/httpserver.go` 的 `InstallRESTAPI`：

1) 把 `hdl := handler.NewHandler(c.biz, c.val, authMiddlewares...)` 改为：
```go
	hdl := handler.NewHandler(c.biz, c.val, c.DefaultLanguage, authMiddlewares...)
```

2) 在 `api.GET("/healthz", hdl.Healthz)` 行后加：
```go
	// 注册全局配置接口（无需认证，登录页加载时调用）
	api.GET("/config", hdl.GetConfig)
```

- [ ] **Step 8: 重新生成 Swagger**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
$(go env GOPATH)/bin/swag init -g main.go -d cmd/shop-apiserver,internal/apiserver/handler -o docs/apidocs --parseDependency --parseInternal 2>&1 | tail -5
```
Expected: 末尾 create docs.go / swagger.json，无 error。

- [ ] **Step 9: 构建后端**

Run:
```bash
make build BINS=shop-apiserver
```
Expected: 构建成功。若 wire 报错（NewHandler 签名变了），检查 httpserver.go 调用是否同步改。

- [ ] **Step 10: 启动验证 /api/config**

停旧服务：`pkill -f 'shop-apiserver -c' 2>/dev/null; sleep 1`

run_in_background 启动：
```bash
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml
```
等 5 秒，验证：
```bash
curl -s http://127.0.0.1:5555/api/config
curl -s -o /dev/null -w "/api/healthz -> %{http_code}\n" http://127.0.0.1:5555/api/healthz
```
Expected: `/api/config` 返回 `{"defaultLanguage":"zh"}`（可能被 core.WriteResponse 包一层，看实际结构）；healthz 200。

> 注：core.WriteResponse 会把数据包在 `{code,data,message}` 信封里。若前端要直接读 defaultLanguage，需取 `.data.defaultLanguage`。Task 3 的 I18nBootstrap 会按实际响应结构解析。

---

## Task 2: 前端 i18n 基础设施

**Files:**
- Modify: `frontend/package.json`（装依赖）
- Create: `src/i18n/config.ts`
- Create: `src/i18n/client.ts`（含 I18nBootstrap）
- Create: `src/locales/zh/common.json`
- Create: `src/locales/en/common.json`
- Modify: `src/app/layout.tsx`（包裹 I18nBootstrap）

**Interfaces:**
- Produces: `initI18n(defaultLng)`、`I18nBootstrap` 组件；语言文件 key 结构供后续页面 `t()` 使用

- [ ] **Step 1: 安装 i18n 依赖**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm install i18next react-i18next
```
Expected: 安装成功，package.json 出现 i18next / react-i18next。

- [ ] **Step 2: 创建中文语言文件**

Create `src/locales/zh/common.json`：
```json
{
  "login": {
    "title": "登录 Shop App",
    "subtitle": "登录你的账户",
    "username": "用户名",
    "password": "密码",
    "submit": "登录",
    "submitting": "登录中...",
    "register": "还没有账号？",
    "registerLink": "去注册",
    "forgotPassword": "忘记密码？",
    "socialLogin": "第三方登录"
  },
  "register": {
    "title": "注册账号",
    "username": "用户名",
    "password": "密码",
    "nickname": "昵称",
    "email": "邮箱",
    "phone": "手机号",
    "submit": "注册",
    "submitting": "注册中...",
    "login": "已有账号？",
    "loginLink": "去登录"
  },
  "validation": {
    "usernameMin": "用户名至少 3 个字符",
    "passwordMin": "密码至少 8 个字符",
    "nicknameRequired": "昵称不能为空",
    "emailInvalid": "邮箱格式不正确",
    "phoneRequired": "手机号不能为空",
    "phoneInvalid": "手机号格式不正确"
  },
  "error": {
    "loginFailed": "登录失败",
    "registerFailed": "注册失败",
    "inputInvalid": "输入不合法",
    "comingSoon": "功能开发中"
  },
  "home": {
    "welcome": "欢迎，{{name}}",
    "loginSuccess": "登录成功，这是首页占位页。",
    "logout": "登出"
  },
  "language": {
    "zh": "中文",
    "en": "English"
  }
}
```

- [ ] **Step 3: 创建英文语言文件**

Create `src/locales/en/common.json`（key 与 zh 完全一致）：
```json
{
  "login": {
    "title": "Login Shop App",
    "subtitle": "Sign in to your account",
    "username": "Username",
    "password": "Password",
    "submit": "Login",
    "submitting": "Logging in...",
    "register": "No account yet?",
    "registerLink": "Register",
    "forgotPassword": "Forgot password?",
    "socialLogin": "Third-party login"
  },
  "register": {
    "title": "Register",
    "username": "Username",
    "password": "Password",
    "nickname": "Nickname",
    "email": "Email",
    "phone": "Phone",
    "submit": "Register",
    "submitting": "Registering...",
    "login": "Already have an account?",
    "loginLink": "Login"
  },
  "validation": {
    "usernameMin": "Username must be at least 3 characters",
    "passwordMin": "Password must be at least 8 characters",
    "nicknameRequired": "Nickname is required",
    "emailInvalid": "Invalid email format",
    "phoneRequired": "Phone is required",
    "phoneInvalid": "Invalid phone format"
  },
  "error": {
    "loginFailed": "Login failed",
    "registerFailed": "Registration failed",
    "inputInvalid": "Invalid input",
    "comingSoon": "Coming soon"
  },
  "home": {
    "welcome": "Welcome, {{name}}",
    "loginSuccess": "Login successful. This is the home placeholder.",
    "logout": "Logout"
  },
  "language": {
    "zh": "中文",
    "en": "English"
  }
}
```

- [ ] **Step 4: 创建 i18n/config.ts**

Create `src/i18n/config.ts`：
```ts
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import zh from "@/locales/zh/common.json";
import en from "@/locales/en/common.json";

// 支持的语言列表（新增语言只需在此注册 + 加 locales/{lang}/common.json）
export const supportedLanguages = ["zh", "en"] as const;
export type Language = (typeof supportedLanguages)[number];

export const defaultLanguage: Language = "zh";

// 初始化 i18next。defaultLng 为初始语言（由调用方按优先级决定）.
export function initI18n(defaultLng: string) {
  // 已初始化则只切语言，避免重复 init
  if (i18n.isInitialized) {
    void i18n.changeLanguage(defaultLng);
    return i18n;
  }
  return i18n.use(initReactI18next).init({
    resources: { zh: { common: zh }, en: { common: en } },
    lng: defaultLng,
    fallbackLng: defaultLanguage,
    defaultNS: "common",
    ns: ["common"],
    interpolation: { escapeValue: false },
  });
}
```

- [ ] **Step 5: 创建 i18n/client.ts（I18nBootstrap 组件）**

Create `src/i18n/client.ts`：
```ts
"use client";

import { useEffect, useState, type ReactNode } from "react";
import { I18nextProvider } from "react-i18next";
import i18n from "i18next";
import { initI18n, defaultLanguage, type Language } from "./config";
import { webGet } from "@/api";
import { configResponseSchema } from "@/schemas/config";

const LANG_STORAGE_KEY = "shop-app-lang";

// 从 localStorage 取用户选择的语言
function getStoredLang(): Language | null {
  if (typeof window === "undefined") return null;
  const v = localStorage.getItem(LANG_STORAGE_KEY);
  if (v === "zh" || v === "en") return v;
  return null;
}

// 从后端 /api/config 取默认语言。
// webGet 返回 { data, error }，data 已是 zod schema 校验后的对象。
// 注意：若后端 core.WriteResponse 把数据包在 {code,data,message} 信封里，
// 需先 curl 确认 /api/config 实际返回结构，再调整 configResponseSchema（见 schemas/config.ts）。
async function fetchServerLang(): Promise<Language> {
  try {
    const { data, error } = await webGet("/config", configResponseSchema);
    if (error || !data) return defaultLanguage;
    return data.defaultLanguage === "en" ? "en" : "zh";
  } catch {
    return defaultLanguage;
  }
}

// I18nBootstrap：挂载时确定语言并初始化 i18n，未就绪时显示 loading
export function I18nBootstrap({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(i18n.isInitialized);

  useEffect(() => {
    if (i18n.isInitialized) {
      setReady(true);
      return;
    }
    (async () => {
      const stored = getStoredLang();
      const lng = stored ?? (await fetchServerLang());
      await initI18n(lng);
      setReady(true);
    })();
  }, []);

  if (!ready) {
    return <div style={{ minHeight: "100vh", background: "#181828" }} />;
  }
  return <I18nextProvider i18n={i18n}>{children}</I18nextProvider>;
}
```

- [ ] **Step 6: 创建 config schema（供 client.ts 用）**

Create `src/schemas/config.ts`：
```ts
import { z } from "zod";

// /api/config 响应 schema。
// ⚠️ 后端 core.WriteResponse 会把数据包在 {code,data,message} 信封里。
// 实现前先 curl `GET /api/config` 确认实际返回结构：
//   - 若返回 {"defaultLanguage":"zh"}（无信封）→ 用下面的 configResponseSchema
//   - 若返回 {"code":0,"data":{"defaultLanguage":"zh"},"message":""}（有信封）
//     → webGet 的 schema 应匹配信封，且 fetchServerLang 取 data.defaultLanguage 需调整
// 先按无信封假设写，curl 后按实际修正。
export const configResponseSchema = z.object({
  defaultLanguage: z.string(),
});
```

> **实现时第一步**：Task 1 Step 10 已 curl 确认 /api/config 返回结构。根据实际结果修正此 schema：
> - 若 core.WriteResponse 返回 `{"code":200,"data":{"defaultLanguage":"zh"},"message":""}`，则 schema 改为 `z.object({ code: z.number(), data: z.object({ defaultLanguage: z.string() }), message: z.string().optional() })`，且 `fetchServerLang` 取 `data.data.defaultLanguage`。
> - 此判断在 Task 2 开始前完成。

- [ ] **Step 7: 改 layout.tsx 包裹 I18nBootstrap**

Edit `src/app/layout.tsx`，整体替换为：
```tsx
import type { Metadata } from "next";
import "./globals.css";
import { I18nBootstrap } from "@/i18n/client";

export const metadata: Metadata = {
  title: "Shop App",
  description: "基于 OneX 技术栈的电商后端服务",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh" className="h-full antialiased">
      <body className="min-h-full flex flex-col font-sans">
        <I18nBootstrap>{children}</I18nBootstrap>
      </body>
    </html>
  );
}
```

- [ ] **Step 8: typecheck**

Run:
```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app/frontend
npm run typecheck
```
Expected: 无错误。若有 `apiGet` 未导出错误，检查 `src/api/index.ts` 是否导出 `apiGet`（模板导出的是 `webGet`，需用 `webGet` 或确认 shared 导出 `apiGet`）。

> 修正：模板 `src/api/request.ts` 导出 `webGet/webPost/...`（web-bound），`@devkit/shared` 导出 `apiGet/apiPost/...`。client.ts 和 config schema 用 `webGet` 更合适（已绑定 apiClient）。若 typecheck 报 apiGet 未导出，改 import 为 `import { webGet } from "@/api"`，调用 `webGet("/config", schema)`。

---

## Task 3: globals.css 暗色变量 + 星尘动画

**Files:**
- Modify: `src/app/globals.css`

**Interfaces:**
- Produces: 全局 CSS 变量（`--bg/--accent/--card-bg/--border`）+ 星尘 class（`.starsec/.starthird/.starfourth/.starfifth`）供页面引用

- [ ] **Step 1: 重写 globals.css**

Edit `src/app/globals.css`，整体替换为：
```css
@import "tailwindcss";

:root {
  /* 暗色调色板 */
  --bg: #181828;
  --foreground: #ededed;
  --accent: #00ffaaed;
  --card-bg: #141421;
  --card-border: #2e2e4c;
  --input-bg: rgba(28, 31, 47, 0.16);
  --input-border: #2e344d;
  --btn-bg: #1c1f2f;
  --muted: rgba(255, 255, 255, 0.41);
}

@theme inline {
  --color-background: var(--bg);
  --color-foreground: var(--foreground);
  --font-sans: var(--font-geist-sans);
  --font-mono: var(--font-geist-mono);
}

body {
  background: var(--bg);
  color: var(--foreground);
  font-family: Arial, Helvetica, sans-serif;
}

/* ===== 星尘动画（全局，登录/注册/首页复用）===== */
.starsec,
.starthird,
.starfourth,
.starfifth {
  content: " ";
  position: absolute;
  background: transparent;
  z-index: 0;
}

.starsec {
  width: 3px;
  height: 3px;
  box-shadow: 571px 173px #00BCD4, 1732px 143px #00BCD4, 1745px 454px #FF5722, 234px 784px #00BCD4, 1793px 1123px #FF9800, 1076px 504px #03A9F4, 633px 601px #FF5722, 350px 630px #FFEB3B, 1164px 782px #00BCD4, 76px 690px #3F51B5, 1825px 701px #CDDC39, 1646px 578px #FFEB3B, 544px 293px #2196F3, 445px 1061px #673AB7, 928px 47px #00BCD4, 168px 1410px #8BC34A, 777px 782px #9C27B0, 1235px 1941px #9C27B0, 104px 1690px #8BC34A, 1167px 1338px #E91E63, 345px 1652px #009688, 1682px 1196px #F44336, 1995px 494px #8BC34A, 428px 798px #FF5722, 340px 1623px #F44336, 605px 349px #9C27B0, 1339px 1344px #673AB7, 1102px 1745px #3F51B5, 1592px 1676px #2196F3, 419px 1024px #FF9800, 630px 1033px #4CAF50, 1995px 1644px #00BCD4, 1092px 712px #9C27B0, 1355px 606px #F44336, 622px 1881px #CDDC39, 1481px 621px #9E9E9E, 19px 1348px #8BC34A, 864px 1780px #E91E63, 442px 1136px #2196F3, 67px 712px #FF5722, 89px 1406px #F44336, 275px 321px #009688, 592px 630px #E91E63, 1012px 1690px #9C27B0, 1749px 23px #673AB7, 94px 1542px #FFEB3B, 1201px 1657px #3F51B5, 1505px 692px #2196F3, 1799px 601px #03A9F4, 656px 811px #00BCD4, 701px 597px #00BCD4, 1202px 46px #FF5722, 890px 569px #FF5722, 1613px 813px #2196F3, 223px 252px #FF9800, 983px 1093px #F44336, 726px 1029px #FFC107, 1764px 778px #CDDC39, 622px 1643px #F44336, 174px 1559px #673AB7, 212px 517px #00BCD4, 340px 505px #FFF, 1700px 39px #FFF, 1768px 516px #F44336, 849px 391px #FF9800, 228px 1824px #FFF, 1119px 1680px #FFC107, 812px 1480px #3F51B5, 1438px 1585px #CDDC39, 137px 1397px #FFF, 1080px 456px #673AB7, 1208px 1437px #03A9F4, 857px 281px #F44336, 1254px 1306px #CDDC39, 987px 990px #4CAF50, 1655px 911px #00BCD4, 1102px 1216px #FF5722, 1807px 1044px #FFF, 660px 435px #03A9F4, 299px 678px #4CAF50, 1193px 115px #FF9800, 918px 290px #CDDC39, 1447px 1422px #FFEB3B, 91px 1273px #9C27B0, 108px 223px #FFEB3B, 146px 754px #00BCD4, 461px 1446px #FF5722, 1004px 391px #673AB7, 1529px 516px #F44336, 1206px 845px #CDDC39, 347px 583px #009688, 1102px 1332px #F44336, 709px 1756px #00BCD4, 1972px 248px #FFF, 1669px 1344px #FF5722, 1132px 406px #F44336, 320px 1076px #CDDC39, 126px 943px #FFEB3B, 263px 604px #FF5722, 1546px 692px #F44336;
  animation: animStar 150s linear infinite;
}

.starthird {
  width: 3px;
  height: 3px;
  box-shadow: 571px 173px #00BCD4, 1732px 143px #00BCD4, 1745px 454px #FF5722, 234px 784px #00BCD4, 1793px 1123px #FF9800, 1076px 504px #03A9F4, 633px 601px #FF5722, 350px 630px #FFEB3B, 1164px 782px #00BCD4, 76px 690px #3F51B5, 1825px 701px #CDDC39, 1646px 578px #FFEB3B, 544px 293px #2196F3, 445px 1061px #673AB7, 928px 47px #00BCD4, 168px 1410px #8BC34A, 777px 782px #9C27B0, 1235px 1941px #9C27B0, 104px 1690px #8BC34A, 1167px 1338px #E91E63, 345px 1652px #009688, 1682px 1196px #F44336, 1995px 494px #8BC34A, 428px 798px #FF5722, 340px 1623px #F44336, 605px 349px #9C27B0, 1339px 1344px #673AB7, 1102px 1745px #3F51B5, 1592px 1676px #2196F3, 419px 1024px #FF9800, 630px 1033px #4CAF50, 1995px 1644px #00BCD4, 1092px 712px #9C27B0, 1355px 606px #F44336, 622px 1881px #CDDC39, 1481px 621px #9E9E9E, 19px 1348px #8BC34A, 864px 1780px #E91E63, 442px 1136px #2196F3, 67px 712px #FF5722, 89px 1406px #F44336, 275px 321px #009688, 592px 630px #E91E63, 1012px 1690px #9C27B0, 1749px 23px #673AB7, 94px 1542px #FFEB3B, 1201px 1657px #3F51B5, 1505px 692px #2196F3, 1799px 601px #03A9F4, 656px 811px #00BCD4, 701px 597px #00BCD4, 1202px 46px #FF5722, 890px 569px #FF5722, 1613px 813px #2196F3, 223px 252px #FF9800, 983px 1093px #F44336, 726px 1029px #FFC107, 1764px 778px #CDDC39, 622px 1643px #F44336, 174px 1559px #673AB7, 212px 517px #00BCD4, 340px 505px #FFF, 1700px 39px #FFF, 1768px 516px #F44336, 849px 391px #FF9800, 228px 1824px #FFF, 1119px 1680px #FFC107, 812px 1480px #3F51B5, 1438px 1585px #CDDC39, 137px 1397px #FFF, 1080px 456px #673AB7, 1208px 1437px #03A9F4, 857px 281px #F44336, 1254px 1306px #CDDC39, 987px 990px #4CAF50, 1655px 911px #00BCD4, 1102px 1216px #FF5722, 1807px 1044px #FFF, 660px 435px #03A9F4, 299px 678px #4CAF50, 1193px 115px #FF9800, 918px 290px #CDDC39, 1447px 1422px #FFEB3B, 91px 1273px #9C27B0, 108px 223px #FFEB3B, 146px 754px #00BCD4, 461px 1446px #FF5722, 1004px 391px #673AB7, 1529px 516px #F44336, 1206px 845px #CDDC39, 347px 583px #009688, 1102px 1332px #F44336, 709px 1756px #00BCD4, 1972px 248px #FFF, 1669px 1344px #FF5722, 1132px 406px #F44336, 320px 1076px #CDDC39, 126px 943px #FFEB3B, 263px 604px #FF5722, 1546px 692px #F44336;
  animation: animStar 10s linear infinite;
}

.starfourth {
  width: 2px;
  height: 2px;
  box-shadow: 571px 173px #00BCD4, 1732px 143px #00BCD4, 1745px 454px #FF5722, 234px 784px #00BCD4, 1793px 1123px #FF9800, 1076px 504px #03A9F4, 633px 601px #FF5722, 350px 630px #FFEB3B, 1164px 782px #00BCD4, 76px 690px #3F51B5, 1825px 701px #CDDC39, 1646px 578px #FFEB3B, 544px 293px #2196F3, 445px 1061px #673AB7, 928px 47px #00BCD4, 168px 1410px #8BC34A, 777px 782px #9C27B0, 1235px 1941px #9C27B0, 104px 1690px #8BC34A, 1167px 1338px #E91E63, 345px 1652px #009688, 1682px 1196px #F44336, 1995px 494px #8BC34A, 428px 798px #FF5722, 340px 1623px #F44336, 605px 349px #9C27B0, 1339px 1344px #673AB7, 1102px 1745px #3F51B5, 1592px 1676px #2196F3, 419px 1024px #FF9800, 630px 1033px #4CAF50, 1995px 1644px #00BCD4, 1092px 712px #9C27B0, 1355px 606px #F44336, 622px 1881px #CDDC39, 1481px 621px #9E9E9E, 19px 1348px #8BC34A, 864px 1780px #E91E63, 442px 1136px #2196F3, 67px 712px #FF5722, 89px 1406px #F44336, 275px 321px #009688, 592px 630px #E91E63, 1012px 1690px #9C27B0, 1749px 23px #673AB7, 94px 1542px #FFEB3B, 1201px 1657px #3F51B5, 1505px 692px #2196F3, 1799px 601px #03A9F4, 656px 811px #00BCD4, 701px 597px #00BCD4, 1202px 46px #FF5722, 890px 569px #FF5722, 1613px 813px #2196F3, 223px 252px #FF9800, 983px 1093px #F44336, 726px 1029px #FFC107, 1764px 778px #CDDC39, 622px 1643px #F44336, 174px 1559px #673AB7, 212px 517px #00BCD4, 340px 505px #FFF, 1700px 39px #FFF, 1768px 516px #F44336, 849px 391px #FF9800, 228px 1824px #FFF, 1119px 1680px #FFC107, 812px 1480px #3F51B5, 1438px 1585px #CDDC39, 137px 1397px #FFF, 1080px 456px #673AB7, 1208px 1437px #03A9F4, 857px 281px #F44336, 1254px 1306px #CDDC39, 987px 990px #4CAF50, 1655px 911px #00BCD4, 1102px 1216px #FF5722, 1807px 1044px #FFF, 660px 435px #03A9F4, 299px 678px #4CAF50, 1193px 115px #FF9800, 918px 290px #CDDC39, 1447px 1422px #FFEB3B, 91px 1273px #9C27B0, 108px 223px #FFEB3B, 146px 754px #00BCD4, 461px 1446px #FF5722, 1004px 391px #673AB7, 1529px 516px #F44336, 1206px 845px #CDDC39, 347px 583px #009688, 1102px 1332px #F44336, 709px 1756px #00BCD4, 1972px 248px #FFF, 1669px 1344px #FF5722, 1132px 406px #F44336, 320px 1076px #CDDC39, 126px 943px #FFEB3B, 263px 604px #FF5722, 1546px 692px #F44336;
  animation: animStar 50s linear infinite;
}

.starfifth {
  width: 1px;
  height: 1px;
  box-shadow: 571px 173px #00BCD4, 1732px 143px #00BCD4, 1745px 454px #FF5722, 234px 784px #00BCD4, 1793px 1123px #FF9800, 1076px 504px #03A9F4, 633px 601px #FF5722, 350px 630px #FFEB3B, 1164px 782px #00BCD4, 76px 690px #3F51B5, 1825px 701px #CDDC39, 1646px 578px #FFEB3B, 544px 293px #2196F3, 445px 1061px #673AB7, 928px 47px #00BCD4, 168px 1410px #8BC34A, 777px 782px #9C27B0, 1235px 1941px #9C27B0, 104px 1690px #8BC34A, 1167px 1338px #E91E63, 345px 1652px #009688, 1682px 1196px #F44336, 1995px 494px #8BC34A, 428px 798px #FF5722, 340px 1623px #F44336, 605px 349px #9C27B0, 1339px 1344px #673AB7, 1102px 1745px #3F51B5, 1592px 1676px #2196F3, 419px 1024px #FF9800, 630px 1033px #4CAF50, 1995px 1644px #00BCD4, 1092px 712px #9C27B0, 1355px 606px #F44336, 622px 1881px #CDDC39, 1481px 621px #9E9E9E, 19px 1348px #8BC34A, 864px 1780px #E91E63, 442px 1136px #2196F3, 67px 712px #FF5722, 89px 1406px #F44336, 275px 321px #009688, 592px 630px #E91E63, 1012px 1690px #9C27B0, 1749px 23px #673AB7, 94px 1542px #FFEB3B, 1201px 1657px #3F51B5, 1505px 692px #2196F3, 1799px 601px #03A9F4, 656px 811px #00BCD4, 701px 597px #00BCD4, 1202px 46px #FF5722, 890px 569px #FF5722, 1613px 813px #2196F3, 223px 252px #FF9800, 983px 1093px #F44336, 726px 1029px #FFC107, 1764px 778px #CDDC39, 622px 1643px #F44336, 174px 1559px #673AB7, 212px 517px #00BCD4, 340px 505px #FFF, 1700px 39px #FFF, 1768px 516px #F44336, 849px 391px #FF9800, 228px 1824px #FFF, 1119px 1680px #FFC107, 812px 1480px #3F51B5, 1438px 1585px #CDDC39, 137px 1397px #FFF, 1080px 456px #673AB7, 1208px 1437px #03A9F4, 857px 281px #F44336, 1254px 1306px #CDDC39, 987px 990px #4CAF50, 1655px 911px #00BCD4, 1102px 1216px #FF5722, 1807px 1044px #FFF, 660px 435px #03A9F4, 299px 678px #4CAF50, 1193px 115px #FF9800, 918px 290px #CDDC39, 1447px 1422px #FFEB3B, 91px 1273px #9C27B0, 108px 223px #FFEB3B, 146px 754px #00BCD4, 461px 1446px #FF5722, 1004px 391px #673AB7, 1529px 516px #F44336, 1206px 845px #CDDC39, 347px 583px #009688, 1102px 1332px #F44336, 709px 1756px #00BCD4, 1972px 248px #FFF, 1669px 1344px #FF5722, 1132px 406px #F44336, 320px 1076px #CDDC39, 126px 943px #FFEB3B, 263px 604px #FF5722, 1546px 692px #F44336;
  animation: animStar 80s linear infinite;
}

@keyframes animStar {
  0% { transform: translateY(0px); }
  100% { transform: translateY(-2000px); }
}
```

- [ ] **Step 2: build 验证**

Run:
```bash
npm run build
```
Expected: 构建成功，out/ 生成。

---

## Task 4: LanguageSwitcher 组件

**Files:**
- Create: `src/components/LanguageSwitcher.tsx`
- Create: `src/components/LanguageSwitcher.module.css`

**Interfaces:**
- Produces: `LanguageSwitcher` 组件，切换 i18n 语言 + 存 localStorage

- [ ] **Step 1: 创建 LanguageSwitcher.module.css**

Create `src/components/LanguageSwitcher.module.css`：
```css
.switcher {
  position: fixed;
  right: 16px;
  bottom: 16px;
  z-index: 100;
  display: flex;
  gap: 8px;
  background: rgba(20, 20, 33, 0.8);
  border: 1px solid #2e2e4c;
  border-radius: 20px;
  padding: 6px 10px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4);
}

.langBtn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.41);
  cursor: pointer;
  font-size: 0.8rem;
  padding: 4px 8px;
  border-radius: 12px;
  transition: all 0.3s;
}

.langBtn:hover {
  color: #00ffaaed;
}

.active {
  color: #00ffaaed;
  background: rgba(0, 255, 170, 0.1);
}

@media (min-width: 768px) {
  .switcher {
    right: 24px;
    bottom: 24px;
  }
}
```

- [ ] **Step 2: 创建 LanguageSwitcher.tsx**

Create `src/components/LanguageSwitcher.tsx`：
```tsx
"use client";

import { useTranslation } from "react-i18next";
import { supportedLanguages, type Language } from "@/i18n/config";
import styles from "./LanguageSwitcher.module.css";

const LANG_STORAGE_KEY = "shop-app-lang";

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation("common");

  function change(lng: Language) {
    void i18n.changeLanguage(lng);
    if (typeof window !== "undefined") {
      localStorage.setItem(LANG_STORAGE_KEY, lng);
    }
  }

  return (
    <div className={styles.switcher}>
      {supportedLanguages.map((lng) => (
        <button
          key={lng}
          className={`${styles.langBtn} ${i18n.language === lng ? styles.active : ""}`}
          onClick={() => change(lng)}
        >
          {t(`language.${lng}`)}
        </button>
      ))}
    </div>
  );
}
```

- [ ] **Step 3: typecheck**

Run:
```bash
npm run typecheck
```
Expected: 无错误。

---

## Task 5: zod schema 函数式 + store 剥离校验

**Files:**
- Modify: `src/schemas/auth.ts`
- Modify: `src/stores/auth.ts`

**Interfaces:**
- Produces: `makeLoginSchema(t)` / `makeRegisterSchema(t)` 函数式 schema；store 的 `login`/`register` 不再 parse

- [ ] **Step 1: 重写 schemas/auth.ts**

Edit `src/schemas/auth.ts`，整体替换为：
```ts
import { z } from "zod";

// 翻译函数类型（接收 key 返回文案）
type TFunc = (key: string) => string;

// 登录表单校验（函数式，message 走 i18n）
export const makeLoginSchema = (t: TFunc) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
  });
export type LoginForm = { username: string; password: string };

// 注册表单校验（phone 必填，对齐后端）
export const makeRegisterSchema = (t: TFunc) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
    nickname: z.string().min(1, t("validation.nicknameRequired")),
    email: z.string().email(t("validation.emailInvalid")),
    phone: z
      .string()
      .min(1, t("validation.phoneRequired"))
      .regex(/^1[3-9]\d{9}$/, t("validation.phoneInvalid")),
  });
export type RegisterForm = {
  username: string;
  password: string;
  nickname: string;
  email: string;
  phone: string;
};

// API 响应 schema（不变）
export const loginResponseSchema = z.object({
  token: z.string(),
  expireAt: z
    .object({ seconds: z.number().optional(), nanos: z.number().optional() })
    .optional(),
});
export type LoginResponse = z.infer<typeof loginResponseSchema>;

export const registerResponseSchema = z.object({
  userID: z.string(),
});
export type RegisterResponse = z.infer<typeof registerResponseSchema>;
```

- [ ] **Step 2: 改 stores/auth.ts 移除 parse**

Edit `src/stores/auth.ts`，移除 `loginSchema`/`registerSchema` 的 import 和 parse 调用。修改后的 `login`/`register`：
```ts
      login: async (username, password) => {
        const { data, error } = await webPost("/login", loginResponseSchema, { username, password });
        if (error) throw error;
        set({ token: data!.token, user: { userID: "", username } });
      },
      register: async (form) => {
        const { data, error } = await webPost("/v1/users", registerResponseSchema, form);
        if (error) throw error;
        return data!.userID;
      },
```

并移除文件顶部 `import { loginSchema, registerSchema, ... } from "@/schemas/auth"` 中已不用的 `loginSchema`/`registerSchema`（保留 `loginResponseSchema`/`registerResponseSchema`）。

- [ ] **Step 3: typecheck**

Run:
```bash
npm run typecheck
```
Expected: 无错误。注意此时 login/register page 仍用旧的静态 schema import，可能报错——Task 6/7 会修。若 typecheck 因 login/page.tsx 报错，先注释掉那两行 import 临时通过，Task 6 会重写。

---

## Task 6: 登录页重构

**Files:**
- Create: `src/app/login/login.module.css`
- Modify: `src/app/login/page.tsx`（重写）

**Interfaces:**
- Consumes: `useTranslation`、`useAuthStore.login`、`makeLoginSchema`、`LanguageSwitcher`、星尘全局 class

- [ ] **Step 1: 创建 login.module.css**

Create `src/app/login/login.module.css`：
```css
.container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #181828;
  padding: 16px;
  position: relative;
  overflow: hidden;
}

.card {
  position: relative;
  z-index: 1;
  background: #141421;
  border: 1px solid #2e2e4c;
  border-radius: 10px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), -3px -3px 10px rgba(255, 255, 255, 0.06),
    inset 14px 14px 26px rgb(0, 0, 0, 0.3), inset -3px -3px 15px rgba(255, 255, 255, 0.05);
  padding: 24px;
  width: 100%;
  max-width: 320px;
}

.title {
  color: #00ffaaed;
  text-align: center;
  font-size: 1.25rem;
  margin: 0 0 4px 0;
}

.subtitle {
  color: rgba(255, 255, 255, 0.41);
  text-align: center;
  font-size: 0.85rem;
  margin: 0 0 16px 0;
}

.input {
  background: rgba(28, 31, 47, 0.16);
  border: 1px solid #2e344d;
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  margin-top: 15px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s;
}

.input:focus {
  outline: 0;
  border: 1px solid #344d2e;
  background: rgb(17, 20, 31);
}

.loginBtn {
  background: #1c1f2f;
  border: 1px solid #2e344d;
  border-radius: 30px;
  width: 100%;
  padding: 11px;
  margin-top: 24px;
  color: #fff;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0px 2px 26px rgb(0, 0, 0, 0.5), 0px 7px 13px rgba(255, 255, 255, 0.03);
}

.loginBtn:hover:not(:disabled) {
  border-radius: 50px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), -3px -3px 10px rgba(255, 255, 255, 0.06),
    inset 14px 14px 26px rgb(0, 0, 0, 0.3), inset -3px -3px 15px rgba(255, 255, 255, 0.05);
}

.loginBtn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error {
  color: #ff5555;
  font-size: 0.8rem;
  margin-top: 12px;
  text-align: center;
}

.forgotLink {
  display: block;
  text-align: center;
  color: #2b7a19;
  font-size: 0.8rem;
  margin-top: 12px;
  cursor: pointer;
  text-decoration: underline;
}

.socialList {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-top: 2rem;
}

.socialBtn {
  width: 40px;
  height: 40px;
  border-radius: 10%;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), -3px -3px 10px rgba(255, 255, 255, 0.06),
    inset 14px 14px 26px rgb(0, 0, 0, 0.3), inset -3px -3px 15px rgba(255, 255, 255, 0.05);
  transition: all 0.3s;
}

.socialBtn:hover {
  transform: scale(1.1);
}

.registerLink {
  text-align: center;
  color: rgba(255, 255, 255, 0.41);
  font-size: 0.85rem;
  margin-top: 16px;
}

.registerLink a {
  color: #00ffaaed;
  text-decoration: none;
  margin-left: 4px;
}

.registerLink a:hover {
  text-decoration: underline;
}

/* PC 段 */
@media (min-width: 768px) {
  .card {
    max-width: 420px;
    padding: 36px;
  }
  .title {
    font-size: 1.5rem;
  }
  .subtitle {
    font-size: 0.95rem;
  }
}
```

- [ ] **Step 2: 重写 login/page.tsx**

Edit `src/app/login/page.tsx`，整体替换为：
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { makeLoginSchema } from "@/schemas/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./login.module.css";

export default function LoginPage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [form, setForm] = useState({ username: "", password: "" });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = makeLoginSchema(t).safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? t("error.inputInvalid"));
      return;
    }
    setLoading(true);
    try {
      await login(form.username, form.password);
      router.push("/");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t("error.loginFailed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.container}>
      <div className="starsec" />
      <div className="starthird" />
      <div className="starfourth" />
      <div className="starfifth" />

      <div className={styles.card}>
        <h3 className={styles.title}>{t("login.title")}</h3>
        <p className={styles.subtitle}>{t("login.subtitle")}</p>
        <form onSubmit={onSubmit}>
          <input
            className={styles.input}
            placeholder={t("login.username")}
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
          />
          <input
            className={styles.input}
            type="password"
            placeholder={t("login.password")}
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
          {error && <p className={styles.error}>{error}</p>}
          <button className={styles.loginBtn} type="submit" disabled={loading}>
            {loading ? t("login.submitting") : t("login.submit")}
          </button>
        </form>
        <a className={styles.forgotLink} onClick={() => alert(t("error.comingSoon"))}>
          {t("login.forgotPassword")}
        </a>
        <div className={styles.socialList}>
          {["f", "g", "t", "d"].map((s) => (
            <button
              key={s}
              className={styles.socialBtn}
              onClick={() => alert(t("error.comingSoon"))}
              aria-label="social login"
            >
              {s.toUpperCase()}
            </button>
          ))}
        </div>
        <p className={styles.registerLink}>
          {t("login.register")}
          <Link href="/register">{t("login.registerLink")}</Link>
        </p>
      </div>
      <LanguageSwitcher />
    </div>
  );
}
```

- [ ] **Step 3: typecheck + build**

Run:
```bash
npm run typecheck && npm run build
```
Expected: 无错误，out/login/index.html 生成。

- [ ] **Step 4: dev 验证**

后端在跑（Task 1 启动的）。启动前端 dev（run_in_background）：
```bash
npm run dev
```
等 8 秒，curl：
```bash
curl -s http://localhost:3000/login/ | grep -oE "登录 Shop App|Login Shop App" | head -1
```
Expected: 输出标题（取决于默认语言，首次应为中文"登录 Shop App"）。

---

## Task 7: 注册页重构

**Files:**
- Create: `src/app/register/register.module.css`
- Modify: `src/app/register/page.tsx`（重写）

- [ ] **Step 1: 创建 register.module.css**

Create `src/app/register/register.module.css`（结构与 login 一致，max-width 略大容纳 5 字段）：
```css
.container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #181828;
  padding: 16px;
  position: relative;
  overflow: hidden;
}

.card {
  position: relative;
  z-index: 1;
  background: #141421;
  border: 1px solid #2e2e4c;
  border-radius: 10px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), -3px -3px 10px rgba(255, 255, 255, 0.06),
    inset 14px 14px 26px rgb(0, 0, 0, 0.3), inset -3px -3px 15px rgba(255, 255, 255, 0.05);
  padding: 24px;
  width: 100%;
  max-width: 320px;
}

.title {
  color: #00ffaaed;
  text-align: center;
  font-size: 1.25rem;
  margin: 0 0 16px 0;
}

.input {
  background: rgba(28, 31, 47, 0.16);
  border: 1px solid #2e344d;
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  margin-top: 12px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s;
}

.input:focus {
  outline: 0;
  border: 1px solid #344d2e;
  background: rgb(17, 20, 31);
}

.registerBtn {
  background: #1c1f2f;
  border: 1px solid #2e344d;
  border-radius: 30px;
  width: 100%;
  padding: 11px;
  margin-top: 24px;
  color: #fff;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0px 2px 26px rgb(0, 0, 0, 0.5), 0px 7px 13px rgba(255, 255, 255, 0.03);
}

.registerBtn:hover:not(:disabled) {
  border-radius: 50px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), -3px -3px 10px rgba(255, 255, 255, 0.06),
    inset 14px 14px 26px rgb(0, 0, 0, 0.3), inset -3px -3px 15px rgba(255, 255, 255, 0.05);
}

.registerBtn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error {
  color: #ff5555;
  font-size: 0.8rem;
  margin-top: 12px;
  text-align: center;
}

.loginLink {
  text-align: center;
  color: rgba(255, 255, 255, 0.41);
  font-size: 0.85rem;
  margin-top: 16px;
}

.loginLink a {
  color: #00ffaaed;
  text-decoration: none;
  margin-left: 4px;
}

.loginLink a:hover {
  text-decoration: underline;
}

@media (min-width: 768px) {
  .card {
    max-width: 420px;
    padding: 36px;
  }
  .title {
    font-size: 1.5rem;
  }
}
```

- [ ] **Step 2: 重写 register/page.tsx**

Edit `src/app/register/page.tsx`，整体替换为：
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { makeRegisterSchema } from "@/schemas/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./register.module.css";

export default function RegisterPage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const [form, setForm] = useState({
    username: "",
    password: "",
    nickname: "",
    email: "",
    phone: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = makeRegisterSchema(t).safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? t("error.inputInvalid"));
      return;
    }
    setLoading(true);
    try {
      await register(form);
      router.push("/login");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t("error.registerFailed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.container}>
      <div className="starsec" />
      <div className="starthird" />
      <div className="starfourth" />
      <div className="starfifth" />

      <div className={styles.card}>
        <h3 className={styles.title}>{t("register.title")}</h3>
        <form onSubmit={onSubmit}>
          <input className={styles.input} placeholder={t("register.username")} value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
          <input className={styles.input} type="password" placeholder={t("register.password")} value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
          <input className={styles.input} placeholder={t("register.nickname")} value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} />
          <input className={styles.input} type="email" placeholder={t("register.email")} value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
          <input className={styles.input} placeholder={t("register.phone")} value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
          {error && <p className={styles.error}>{error}</p>}
          <button className={styles.registerBtn} type="submit" disabled={loading}>
            {loading ? t("register.submitting") : t("register.submit")}
          </button>
        </form>
        <p className={styles.loginLink}>
          {t("register.login")}
          <Link href="/login">{t("register.loginLink")}</Link>
        </p>
      </div>
      <LanguageSwitcher />
    </div>
  );
}
```

- [ ] **Step 3: typecheck + build**

Run:
```bash
npm run typecheck && npm run build
```
Expected: 无错误，out/register/index.html 生成。

---

## Task 8: 首页适配

**Files:**
- Create: `src/app/page.module.css`
- Modify: `src/app/page.tsx`（重写）

- [ ] **Step 1: 创建 page.module.css**

Create `src/app/page.module.css`：
```css
.container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #181828;
  padding: 16px;
  position: relative;
  overflow: hidden;
  gap: 16px;
  z-index: 1;
}

.title {
  position: relative;
  z-index: 1;
  color: #00ffaaed;
  font-size: 1.5rem;
  margin: 0;
}

.text {
  position: relative;
  z-index: 1;
  color: rgba(255, 255, 255, 0.41);
  font-size: 0.9rem;
}

.logoutBtn {
  position: relative;
  z-index: 1;
  background: #1c1f2f;
  border: 1px solid #2e344d;
  border-radius: 30px;
  padding: 10px 32px;
  color: #fff;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0px 2px 26px rgb(0, 0, 0, 0.5);
}

.logoutBtn:hover {
  border-radius: 50px;
  box-shadow: 3px 9px 16px rgb(0, 0, 0, 0.4), inset 14px 14px 26px rgb(0, 0, 0, 0.3);
}

@media (min-width: 768px) {
  .title {
    font-size: 2rem;
  }
}
```

- [ ] **Step 2: 重写 page.tsx**

Edit `src/app/page.tsx`，整体替换为：
```tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./page.module.css";

export default function HomePage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const { token, user, logout, isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
    }
  }, [isAuthenticated, router]);

  if (!token) return null;

  return (
    <div className={styles.container}>
      <div className="starsec" />
      <div className="starthird" />
      <div className="starfourth" />
      <div className="starfifth" />

      <h1 className={styles.title}>{t("home.welcome", { name: user?.username ?? "" })}</h1>
      <p className={styles.text}>{t("home.loginSuccess")}</p>
      <button
        className={styles.logoutBtn}
        onClick={() => {
          logout();
          router.push("/login");
        }}
      >
        {t("home.logout")}
      </button>
      <LanguageSwitcher />
    </div>
  );
}
```

- [ ] **Step 3: typecheck + build**

Run:
```bash
npm run typecheck && npm run build
```
Expected: 无错误。

- [ ] **Step 4: dev 端到端验证**

前端 dev 在跑。验证三页：
```bash
curl -s http://localhost:3000/login/ | grep -oE "登录 Shop App" | head -1
curl -s http://localhost:3000/register/ | grep -oE "注册账号" | head -1
curl -s -o /dev/null -w "/ -> %{http_code}\n" http://localhost:3000/
```
Expected: 前两个输出标题，`/` 200。

---

## Task 9: CLAUDE.md 加样式与 i18n 规范

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: 在 CLAUDE.md 追加规范章节**

Read `CLAUDE.md` 找到合适位置（如「接口文档约定」章节后），追加：

```markdown

---

## 前端样式与 i18n 规范

### 样式架构
- **全局入口**：`src/app/globals.css` 是唯一全局 CSS 文件（Tailwind + CSS 变量 + 全局动画如星尘 + 基础重置）
- **组件样式**：统一用 CSS Modules（`xxx.module.css`，与 .tsx 同目录），不写内联样式
- **PC/移动双适配**：每个 module.css 内部用 @media 分两段
  - 默认段写移动样式（移动优先）
  - `@media (min-width: 768px)` 写 PC 覆盖样式
- **class 命名**：语义化（.loginCard/.loginInput），不用 Bootstrap/Tailwind 类名风格
- **特殊视觉页**（登录/注册等）用纯 module.css 自定义组件；普通页可继续用 Tailwind + shadcn/ui

### i18n
- 用 react-i18next，语言文件 `src/locales/{lang}/common.json`
- 默认 zh，支持 en，新增语言只需加目录 + 在 `src/i18n/config.ts` 的 `supportedLanguages` 注册
- 默认语言优先级：localStorage(`shop-app-lang`) → 后端 `/api/config` 的 defaultLanguage → zh
- 所有用户可见文案必须走 `t()`，禁止硬编码中文
- zod 校验文案 i18n 化：用 `makeXxxSchema(t)` 函数式 schema，组件层校验
```

- [ ] **Step 2: 验证 CLAUDE.md 更新**

Run:
```bash
grep -c "前端样式与 i18n 规范" /home/mungdong/workspace/full-stack-devkit/projects/shop-app/CLAUDE.md
```
Expected: 输出 1（或更多）。

---

## Task 10: 全链路验证（dev + docker 重建）

**Files:** 无（验证任务）

- [ ] **Step 1: 重新构建后端二进制（含 /api/config）**

```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
make build BINS=shop-apiserver
```

- [ ] **Step 2: 重新构建前端产物**

```bash
cd frontend && npm run build
```

- [ ] **Step 3: 本地二进制端到端验证**

停旧服务：`pkill -f 'shop-apiserver -c' 2>/dev/null; sleep 1`

run_in_background 启动：
```bash
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml
```
等 5 秒，验证：
```bash
curl -s http://127.0.0.1:5555/api/config
curl -s -o /dev/null -w "/ -> %{http_code}\n" http://127.0.0.1:5555/
curl -s -o /dev/null -w "/login/ -> %{http_code}\n" http://127.0.0.1:5555/login/
curl -s -o /dev/null -w "/api/healthz -> %{http_code}\n" http://127.0.0.1:5555/api/healthz
curl -s -o /dev/null -w "/swagger/index.html -> %{http_code}\n" http://127.0.0.1:5555/swagger/index.html
```
Expected: /api/config 返回含 defaultLanguage；其余 200。

- [ ] **Step 4: 重建 docker 镜像**

```bash
cd /home/mungdong/workspace/full-stack-devkit/projects/shop-app
docker build -f build/docker/shop-apiserver/Dockerfile -t shop-apiserver:dev . 2>&1 | tail -10
```
Expected: Successfully tagged shop-apiserver:dev。

- [ ] **Step 5: 容器端到端验证**

```bash
docker rm -f shop-test 2>/dev/null
docker run -d --name shop-test -p 5555:5555 \
  --link shop-mariadb:shop-mariadb \
  -v /tmp/shop-apiserver-container.yaml:/app/configs/shop-apiserver.yaml \
  shop-apiserver:dev
```
> 注：/tmp/shop-apiserver-container.yaml 是之前 Task 10 创建的容器配置（mysql.addr=shop-mariadb:3306）。若不存在，先从 configs/shop-apiserver.yaml 复制并改 mysql.addr。本次 yaml 新增了 i18n 配置，需重新生成该文件：`cp configs/shop-apiserver.yaml /tmp/shop-apiserver-container.yaml && sed -i 's|addr: 127.0.0.1:3306|addr: shop-mariadb:3306|' /tmp/shop-apiserver-container.yaml`

等 5 秒，验证：
```bash
docker logs shop-test 2>&1 | tail -5
curl -s http://127.0.0.1:5555/api/config
curl -s -o /dev/null -w "/login/ -> %{http_code}\n" http://127.0.0.1:5555/login/
```

- [ ] **Step 6: 验收标准对照**

对照 spec §9 验收标准：
1. ✅ `npm run build` 产出 out/（Step 2）
2. ✅ `make build` 通过（Step 1）
3. ✅ 登录页暗色星尘风（浏览器验证）
4. ✅ 注册页同视觉 5 字段（浏览器验证）
5. ✅ 首页暗色星尘风登出可用（浏览器验证）
6. ✅ zh/en 切换即时生效刷新保持（浏览器验证）
7. ✅ 清 localStorage 后默认语言取自后端（Step 3 的 /api/config）
8. ✅ GET /api/config 返回 defaultLanguage（Step 3/5）
9. ✅ Swagger 显示 /api/config（curl /swagger）
10. ✅ CLAUDE.md 含规范（Task 9）

> 浏览器交互验证（3/4/5/6）需主人确认。浮浮酱完成 curl 层验证后，告知主人浏览器访问地址。

- [ ] **Step 7: 清理**

容器保留或按需删：`docker rm -f shop-test`。本地服务保持运行供主人浏览器验证。

---

## 验收标准对照（spec §9）

1. `npm run build` 产出 out/ —— Task 3/6/7/8/10
2. `make build` 通过 —— Task 1/10
3. 登录页暗色星尘风 PC/移动响应式 —— Task 6（浏览器验证）
4. 注册页同视觉 5 字段 —— Task 7（浏览器验证）
5. 首页暗色星尘风登出可用 —— Task 8（浏览器验证）
6. zh/en 切换即时生效刷新保持 —— Task 4 + 浏览器验证
7. 清 localStorage 后默认语言取自后端 —— Task 1/2 + 浏览器验证
8. GET /api/config 返回 defaultLanguage —— Task 1/10
9. Swagger 显示 /api/config —— Task 1
10. CLAUDE.md 含样式与 i18n 规范 —— Task 9
