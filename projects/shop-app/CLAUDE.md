# CLAUDE.md

本文件是 shop-app 项目的 AI 上下文索引。开发者与 AI Agent 在参与本项目前必须先阅读本文档及索引指向的规范文档。

---

## 项目概述

shop-app 是基于 OneX 技术栈（osbuilder v0.11.1 脚手架）生成的电商后端服务，使用 gin + MariaDB，包含用户登录注册模块。后续规划补 Redis 缓存层。

- **技术栈**：gin / MariaDB(GORM) / docker / onexstack@v0.3.19
- **入口**：`cmd/shop-apiserver/main.go`
- **配置**：`configs/shop-apiserver.yaml`
- **服务端口**：5555

---

## 文档索引

所有开发约定与历史经验记录在 `docs/conventions/` 下，参与开发前必读：

| 文档 | 内容 |
|------|------|
| [开发规范](docs/conventions/development.md) | 技术栈版本约束、分层结构、新增资源流程、命名/注释/Swagger 规范、go.mod 维护红线 |
| [踩坑记录](docs/conventions/pitfalls.md) | onexstack 版本不匹配、本地 replace、Swagger 集成坑、版本兼容矩阵 |
| [Docker 本地测试规则](docs/conventions/docker-local-test.md) | MariaDB 容器启动、连接验证、服务联调、容器管理 |

---

## 接口文档约定（强制）

**本项目每个 HTTP 接口都必须体现在 Swagger 文档中，便于后续测试。** 无论是新增、修改还是删除接口，都要同步 Swagger，保证文档与代码始终一致。

### 硬性要求

1. **每个 handler 函数必须标注 swag 注解**，至少包含：`@Summary`、`@Tags`、`@Router`、`@Param`、`@Success`。无入参/无出参的接口也要标注对应行。
2. **新增接口后必须重新生成文档**，否则 Swagger 不更新：

   ```bash
   swag init -g main.go -d cmd/shop-apiserver,internal/apiserver/handler -o docs/apidocs --parseDependency --parseInternal
   make build BINS=shop-apiserver
   ```

3. **修改接口签名/字段后必须重新生成**，确保 `@Param` 与 `@Success` 引用的类型与代码一致。
4. **删除接口时同步删除其注解**，避免 Swagger 出现失效路径。
5. **访问地址**：`http://localhost:5555/swagger/index.html`，用于接口调试与回归测试。

### 自查清单（提交前）

- [ ] 新增的 handler 是否都加了 swag 注解？
- [ ] 注解里的 `@Router` 路径与实际注册的路由一致？
- [ ] 是否执行了 `swag init` 重新生成 `docs/apidocs/`？
- [ ] Swagger UI 能否看到全部接口且能正常发起请求？

> 详细注解写法与生成命令见 [开发规范 - Swagger 规范](docs/conventions/development.md#6-swagger-规范)；常见坑见 [踩坑记录 - 坑 3](docs/conventions/pitfalls.md#坑-3swagger-集成相关)。

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

---

## 代码提交规范

本项目遵循 **Conventional Commits** 规范（与仓库历史一致）。所有提交信息须用该格式。

### 提交信息格式

```
<type>(<scope>): <subject>

<body>
```

### type 取值

| type | 用途 |
|------|------|
| `feat` | 新功能（如新增 REST 资源、新增接口） |
| `fix` | Bug 修复 |
| `docs` | 文档变更（README、conventions、Swagger 注解等） |
| `style` | 代码格式调整（不影响逻辑） |
| `refactor` | 重构（非新功能、非修 Bug） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建、依赖、配置等杂项 |
| `ci` | CI 配置变更 |

### scope（可选）

受影响的模块，如 `user`、`auth`、`store`、`swagger`、`deps`、`config`。

### subject 规则

- 使用祈使句、现在时：用「添加」而非「添加了」
- 首字母小写，结尾不加句号
- 不超过 50 字符
- 描述「做了什么」，不描述「怎么做」

### 示例

```
feat(user): add change-password endpoint
fix(store): correct user query condition
docs: add swagger integration guide
refactor(auth): extract token validation logic
chore(deps): pin onexstack to v0.3.19
```

### 提交红线

- ❌ 一个提交混合多个无关改动——拆分提交
- ❌ 提交信息含 WIP / 临时 / 无意义内容
- ❌ 提交生成产物（`_output/`、`docs/apidocs/`）——须在 `.gitignore` 排除
- ❌ 提交敏感信息（密码、真实 jwt-key、密钥）
- ✅ 每个提交应能独立构建通过
- ✅ body 说明「为什么」做这个改动（动机、背景）

### 示例（带 body）

```
fix(deps): pin onexstack to v0.3.19

onexstack v0.3.20+ changes RunOrDie signature and v0.3.24+ removes
HTTPOptions, both breaking the osbuilder v0.11.1 template. v0.3.19 is
the last compatible version. See docs/conventions/pitfalls.md.
```

---

## 快速上手

```bash
# 1. 启动 MariaDB（见 docker-local-test.md）
# 2. 构建
make build BINS=shop-apiserver
# 3. 运行
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml
# 4. 访问 Swagger
#    http://localhost:5555/swagger/index.html
```

详细命令见 [开发规范 - 构建与运行](docs/conventions/development.md#7-构建与运行)。
