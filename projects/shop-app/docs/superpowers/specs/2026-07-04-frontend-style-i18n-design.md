# shop-app 前端样式重构与 i18n 设计

- **日期**：2026-07-04
- **范围**：重构登录/注册/首页样式（CSS Modules + 暗色星尘风），全站接入 react-i18next（zh/en，可扩展），后端新增 `/api/config` 接口提供默认语言，约定写入 CLAUDE.md
- **状态**：已确认，待生成实施计划

---

## 1. 背景与目标

当前前端登录/注册页用 shadcn/ui + Tailwind 内联样式，无 i18n。主人要求：
- 样式重构为「全局唯一 CSS 入口 + CSS Modules」架构
- PC/移动双适配（媒体查询分区）
- 用提供的暗色星尘登录页样式重构登录页
- 从开始就支持 i18n，默认中文，支持英文，后续加语言只需加文案文件
- 默认语言从后端配置获取

目标：登录/注册/首页视觉统一为暗色星尘风，全站文案 i18n 化，样式与 i18n 规范固化为 CLAUDE.md 约定。

非目标：
- 不改后端业务接口（仅新增 /api/config）
- 不接入真实 OAuth（社交按钮仅占位）
- 不做找回密码后端（仅占位）
- 不改 auth store 的 API 调用逻辑（仅剥离校验）

---

## 2. 决策矩阵

| 维度 | 决策 | 依据 |
|------|------|------|
| 全局 CSS 入口 | 保留 Tailwind，`globals.css` 作唯一全局入口 | 主人确认，shadcn/ui 仍可用 |
| 组件样式 | CSS Modules（`xxx.module.css`，与 .tsx 同目录） | 主人指定 |
| PC/移动适配 | 单文件 + 内部 `@media (min-width:768px)` 分区，移动优先 | 主人确认（优于双文件） |
| i18n 库 | react-i18next | 主人确认，客户端渲染友好 |
| 默认语言来源 | 后端 `GET /api/config` 返回 `defaultLanguage` | 主人确认 |
| 语言优先级 | localStorage → /api/config → zh | localStorage 记用户选择 |
| 登录页元素 | 保留星尘 + 社交按钮（占位）+ 找回密码（占位）；去掉主题切换器 | 主人多选 |
| 登录页样式实现 | 纯 CSS Modules 自定义，不用 shadcn 组件 | 方案 A，视觉风格不同 |
| 校验位置 | 移到组件层（`makeXxxSchema(t)`），store 不再 parse | i18n 文案动态化 |

---

## 3. 样式架构

### 3.1 CSS 文件分层

```
src/
├── app/
│   └── globals.css          # 唯一全局入口
├── app/login/
│   ├── page.tsx
│   └── login.module.css     # 登录页样式
├── app/register/
│   ├── page.tsx
│   └── register.module.css  # 注册页样式
└── app/
    ├── page.tsx             # 首页
    └── page.module.css      # 首页样式
```

### 3.2 globals.css 职责（唯一全局入口）
- `@import "tailwindcss"`（保留，全局工具/重置）
- CSS 变量定义（暗色调色板：`--bg:#181828`、`--accent:#00ffaaed`、`--card-bg:#141421`、`--border:#2e2e4c` 等）
- **星尘动画**（`@keyframes animStar` + `.starsec/.starthird/.starfourth/.starfifth`）——全局，多页复用
- `body` 基础背景色 `#181828`

### 3.3 module.css 约定（写入 CLAUDE.md）
1. 每个组件/页面一个 `xxx.module.css`，与 `.tsx` 同目录
2. PC/移动双适配：内部 `@media` 分两段——默认段（移动优先）+ `@media (min-width:768px)`（PC 覆盖）
3. class 命名语义化（`.loginCard`、`.loginInput`、`.loginBtn`），不用 Bootstrap/Tailwind 类名风格
4. 登录/注册等特殊视觉页用纯 module.css 自定义；普通页可继续用 Tailwind + shadcn/ui

### 3.4 登录页转译决策

| 原元素 | 处理 |
|--------|------|
| Bootstrap 类（container/row/col） | 丢弃，用 `.container` flex 居中 |
| `top-header` + `main-header` + 1/2/3/4 主题切换器 | 去掉（未保留） |
| `folio-btn` 圆点装饰 | 去掉（装饰无功能） |
| 星尘 `.starsec` 等 | 保留，放 globals.css 全局 |
| FontAwesome 图标（`fab fa-*`） | 换 inline SVG，不引字体库 |
| 外链背景图（toptal/honeycomb） | 去掉（原主题切换用） |
| `.card` 亮绿背景 `#14edaa` | 改用 `.wow-bg` 暗色 `#141421`（原 demo 遗留违和） |
| jQuery 主题切换脚本 | 删除 |

---

## 4. i18n 架构

### 4.1 目录结构

```
src/
├── i18n/
│   ├── config.ts            # i18next 配置（资源、默认、fallback）
│   └── client.ts            # 客户端初始化（initI18n）
├── locales/
│   ├── zh/common.json       # 中文（默认）
│   └── en/common.json       # 英文
└── app/layout.tsx           # 包裹 I18nBootstrap
```

### 4.2 语言文件结构（`common.json`）

zh/en 的 key 完全一致。新增语言只需加 `locales/{lang}/common.json` + 在 config.ts 注册。

```json
{
  "login": { "title", "subtitle", "username", "password", "submit", "submitting", "register", "forgotPassword", "socialLogin" },
  "register": { "title", "username", "password", "nickname", "email", "phone", "submit", "submitting", "login" },
  "validation": { "usernameMin", "passwordMin", "nicknameRequired", "emailInvalid", "phoneRequired", "phoneInvalid" },
  "error": { "loginFailed", "registerFailed", "inputInvalid", "comingSoon" },
  "home": { "welcome", "loginSuccess", "logout" },
  "language": { "zh", "en" }
}
```

### 4.3 默认语言优先级
1. `localStorage["shop-app-lang"]`（用户曾手动切换）
2. 后端 `GET /api/config` 的 `defaultLanguage`
3. 兜底 `"zh"`

### 4.4 初始化流程（client.ts）
```ts
"use client";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import zh from "@/locales/zh/common.json";
import en from "@/locales/en/common.json";

export function initI18n(defaultLng: string) {
  i18n.use(initReactI18next).init({
    resources: { zh: { common: zh }, en: { common: en } },
    lng: defaultLng,
    fallbackLng: "zh",
    defaultNS: "common",
    interpolation: { escapeValue: false },
  });
}
```

### 4.5 语言切换器
- `src/components/LanguageSwitcher.tsx` + `.module.css`（复用，登录/注册页右下角）
- 切换：`i18n.changeLanguage(lng)` + `localStorage.setItem("shop-app-lang", lng)`
- 当前语言 + 下拉（zh/en）

### 4.6 I18nBootstrap 组件（layout.tsx 包裹）
- 挂载时取 localStorage → 无则请求 `/api/config` → `initI18n(lng)`
- 初始化完成前显示 loading，避免文案闪空

---

## 5. 后端 /api/config 接口

### 5.1 yaml 配置项
`configs/shop-apiserver.yaml` 新增：
```yaml
i18n:
  defaultLanguage: zh
```

### 5.2 options
`cmd/shop-apiserver/app/options/options.go` 加 `I18nOptions`（含 `DefaultLanguage string`），从 yaml 读取。

### 5.3 接口
**`GET /api/config`**（无需认证，与 /login 同级）

响应 struct（`internal/apiserver/handler/config.go` 内定义，或加 `pkg/api/apiserver/v1/config.proto`）：
```go
type ConfigResponse struct {
    DefaultLanguage string `json:"defaultLanguage"`
}
```

### 5.4 handler & 路由
- `internal/apiserver/handler/config.go`：`GetConfig` 读 options 返回
- `httpserver.go` `InstallRESTAPI`：`api.GET("/config", hdl.GetConfig)`

### 5.5 Swagger 注解
```go
// @Summary      获取全局配置
// @Tags         系统
// @Produce      json
// @Success      200  {object}  ConfigResponse
// @Router       /api/config [get]
```
重新 `swag init` 同步。

---

## 6. 登录页重构

### 6.1 页面结构（`src/app/login/page.tsx`）
```tsx
<div className={styles.container}>
  <div className="starsec" /> <div className="starthird" />
  <div className="starfourth" /> <div className="starfifth" />

  <div className={styles.card}>
    <div className={styles.cardBody}>
      <h3 className={styles.title}>{t("login.title")}</h3>
      <p className={styles.subtitle}>{t("login.subtitle")}</p>
      <input className={styles.input} placeholder={t("login.username")} ... />
      <input className={styles.input} type="password" placeholder={t("login.password")} ... />
      {error && <p className={styles.error}>{error}</p>}
      <button className={styles.loginBtn} disabled={loading}>
        {loading ? t("login.submitting") : t("login.submit")}
      </button>
      <a className={styles.forgotLink} onClick={() => alert(t("error.comingSoon"))}>
        {t("login.forgotPassword")}
      </a>
      <div className={styles.socialList}>
        {/* 4 个 socialBtn，inline SVG 图标，点击提示开发中 */}
      </div>
      <p className={styles.registerLink}>
        {t("login.register")} <Link href="/register">…</Link>
      </p>
    </div>
  </div>
  <LanguageSwitcher />
</div>
```

### 6.2 `login.module.css` 关键样式
- `.container`：全屏 flex 居中，`background:#181828`，`position:relative; overflow:hidden`（星尘容器）
- `.card`：`background:#141421; border:1px solid #2e2e4c; border-radius:10px; box-shadow:3px 9px 16px rgb(0,0,0,.4)...`
- `.title`：`color:#00ffaaed`
- `.input`：`background:rgba(28,31,47,.16); border:1px solid #2e344d; border-radius:4px; margin-top:15px; color:#fff`
- `.loginBtn`：`background:#1c1f2f; border:1px solid #2e344d; border-radius:30px; box-shadow:0px 2px 26px rgb(0,0,0,.5)`；`:hover` 圆角变 50px + 霓虹阴影
- `.socialBtn`：40×40，`border-radius:10%`，暗色阴影
- `@media (min-width:768px)`：`.card{max-width:420px;padding:36px}` `.title{font-size:1.5rem}`

---

## 7. 注册页 + 首页 + zod i18n 化

### 7.1 注册页
- 与登录页同套视觉（星尘 + 暗色卡片 + 霓虹标题）
- 5 字段（username/password/nickname/email/phone）纵向排列
- `register.module.css` 结构同 login，移动段间距收紧，PC 段 `max-width:420px`
- 文案走 `t("register.*")`

### 7.2 首页（`src/app/page.tsx` + `page.module.css`）
- 背景沿用暗色 `#181828` + 星尘（视觉统一）
- 文案 i18n：`t("home.welcome",{name:user?.username})`、`t("home.loginSuccess")`、`t("home.logout")`
- 登出按钮样式与 `.loginBtn` 一致
- 软守卫逻辑不变

### 7.3 zod schema i18n 化（`src/schemas/auth.ts`）
改为函数式，组件层校验，store 不再 parse：
```ts
export const makeLoginSchema = (t: (k: string) => string) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
  });

export const makeRegisterSchema = (t: (k: string) => string) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
    nickname: z.string().min(1, t("validation.nicknameRequired")),
    email: z.string().email(t("validation.emailInvalid")),
    phone: z.string().min(1, t("validation.phoneRequired"))
      .regex(/^1[3-9]\d{9}$/, t("validation.phoneInvalid")),
  });
```

### 7.4 store 调整（`src/stores/auth.ts`）
- 移除 `loginSchema.parse` / `registerSchema.parse`（校验移到组件层）
- `login`/`register` 只做 API 调用 + set state
- store 不依赖 i18n

---

## 8. CLAUDE.md 约定新增

在 shop-app `CLAUDE.md` 加「前端样式与 i18n 规范」一节：

```markdown
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
- 默认 zh，支持 en，新增语言只需加目录 + 在 `src/i18n/config.ts` 注册
- 默认语言优先级：localStorage(`shop-app-lang`) → 后端 `/api/config` 的 defaultLanguage → zh
- 所有用户可见文案必须走 `t()`，禁止硬编码中文
- zod 校验文案 i18n 化：用 `makeXxxSchema(t)` 函数式 schema，组件层校验
```

---

## 9. 测试策略

### 9.1 前端
- typecheck + build 无错
- dev 模式手工验证：登录/注册/首页渲染，星尘动画，PC/移动响应式（浏览器切屏宽）
- i18n 切换：zh/en 切换即时生效，刷新保持（localStorage）
- 默认语言：清 localStorage 后，前端取后端 /api/config 的值

### 9.2 后端
- `make build` 通过
- curl `GET /api/config` 返回 `{"defaultLanguage":"zh"}`
- Swagger UI 显示 /api/config

### 9.3 验收标准
1. ✅ `npm run build` 产出 out/，无报错
2. ✅ `make build BINS=shop-apiserver` 通过
3. ✅ 登录页暗色星尘风，PC/移动响应式
4. ✅ 注册页同视觉，5 字段
5. ✅ 首页暗色星尘风，登出可用
6. ✅ zh/en 切换即时生效，刷新保持
7. ✅ 清 localStorage 后默认语言取自后端
8. ✅ `GET /api/config` 返回 defaultLanguage
9. ✅ Swagger 显示 /api/config
10. ✅ CLAUDE.md 含样式与 i18n 规范

---

## 10. 实施顺序（粗略，详细计划由 writing-plans 生成）

1. 后端：yaml 加 i18n 配置 + options + config handler + 路由 + swag init
2. 前端：装 react-i18next，搭 i18n/config + client + locales/zh,en + I18nBootstrap + layout 包裹
3. 前端：globals.css 加暗色变量 + 星尘动画
4. 前端：LanguageSwitcher 组件
5. 前端：zod schema 改函数式，store 剥离校验
6. 前端：登录页重构（page.tsx + login.module.css）
7. 前端：注册页重构（page.tsx + register.module.css）
8. 前端：首页适配（page.tsx + page.module.css）
9. CLAUDE.md 加约定
10. 全链路验证（dev + docker）
