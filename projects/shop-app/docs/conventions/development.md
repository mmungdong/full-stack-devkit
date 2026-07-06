# 开发规范

本文档定义 shop-app 项目的代码与工程开发规范，所有贡献者（含 AI Agent）须遵循。

---

## 1. 技术栈与版本约束

| 组件 | 版本 | 约束 |
|------|------|------|
| Go | 1.25.11+（toolchain 可能升至 1.26.x） | 以 `go.mod` 声明为准 |
| Web 框架 | gin | 脚手架生成，勿换 |
| 存储 | MariaDB（GORM） | storageType=mariadb |
| 核心依赖 | `github.com/onexstack/onexstack` @ **v0.3.19** | ⚠️ 版本锁定，见 [踩坑记录](./pitfalls.md) |
| 脚手架 | osbuilder v0.11.1 | 模板版本 |

**关键约束：禁止 `go get -u` / `go get github.com/onexstack/onexstack@latest`**，会触发 API 不兼容。

---

## 2. 项目分层结构

shop-app 遵循 osbuilder 生成的标准分层（OneX 风格），各层职责严格分离：

```
cmd/shop-apiserver/          # 入口：main.go、app/server.go、app/options/
internal/apiserver/
  ├── handler/               # HTTP 路由层：解析请求、调用 biz、返回响应
  ├── biz/v1/<kind>/         # 业务逻辑层：核心业务规则
  ├── store/                 # 数据访问层：GORM 实现 + IStore 接口
  ├── model/                 # GORM 模型（*.gen.go 由 proto 生成）+ Hook
  ├── pkg/validation/        # 请求校验
  ├── pkg/conversion/        # model <-> api 类型转换
  └── server.go / wire.go    # 依赖注入（wire 生成 wire_gen.go）
internal/pkg/
  ├── errno/                 # 错误码定义
  ├── middleware/gin/        # gin 中间件（authn/authz/requestid/header）
  ├── rid/                   # 资源 ID 生成
  └── contextx/ known/       # 上下文与常量
pkg/api/apiserver/v1/        # 对外 API 类型（proto 生成 *.pb.go）
```

### 分层调用规则

- **handler → biz → store**：单向依赖，禁止反向调用。
- handler 层禁止写业务逻辑，只做请求/响应编排。
- store 层禁止包含业务判断，只做 CRUD。
- biz 层是业务唯一入口，跨资源的业务编排在此完成。

---

## 3. 新增 REST 资源规范

新增业务资源（如商品 product、订单 order）**必须**用 osbuilder 生成，禁止手写脚手架文件：

```bash
osbuilder create api --kinds product --binary-name shop-apiserver
```

生成后执行：

```bash
make protoc.apiserver        # 重新编译 proto
go mod tidy
go generate ./...            # 重新生成 wire
make build BINS=shop-apiserver
```

> 详见 osbuilder-iterative-dev skill。osbuilder 会通过 AST 注入自动修改 `store.go`、`biz.go`，勿手动改这些文件的接口签名。

---

## 4. 命名规范

| 类别 | 规则 | 示例 |
|------|------|------|
| Go 文件 | 小写 + 下划线 | `user.go`、`hook_user.go` |
| 资源 kind | snake_case | `cron_job`、`product` |
| proto message | PascalCase + Request/Response 后缀 | `CreateUserRequest` |
| GORM 模型 | PascalCase + `M` 后缀（osbuilder 约定） | `UserM` |
| 错误码 | `Err` 前缀 + PascalCase | `ErrUserNotFound` |
| 路由 | RESTful，复数名词 | `/v1/users`、`/v1/users/:userID` |

---

## 5. 代码注释规范

- **注释语言：与现有代码库一致，使用中文**（osbuilder 生成的代码即为中文注释）。
- 每个导出函数/类型必须有注释，以名称开头。
- handler 函数若需暴露到 Swagger，必须加 swag 注解（见 [Swagger 规范](#6-swagger-规范)）。

```go
// Login 用户登录并返回 JWT Token.
//
// @Summary      用户登录
// @Tags         认证
// @Router       /login [post]
func (h *Handler) Login(c *gin.Context) { ... }
```

---

## 6. Swagger 规范

项目使用 [swaggo/swag](https://github.com/swaggo/swag) + gin-swagger 提供 API 文档。

- **访问地址**：`http://localhost:5555/swagger/index.html`
- **文档生成目录**：`docs/apidocs/`（swag 生成，勿手动编辑）
- **新增/修改 handler 后必须重新生成**：

```bash
swag init -g main.go -d cmd/shop-apiserver,internal/apiserver/handler -o docs/apidocs --parseDependency --parseInternal
```

- 全局 API 信息注解在 `cmd/shop-apiserver/main.go` 顶部维护。
- 每个 handler 的注解须包含：`@Summary`、`@Tags`、`@Router`、`@Param`、`@Success`。

---

## 7. 构建与运行

```bash
make deps                       # 安装工具依赖（首次）
make protoc.apiserver           # 编译 protobuf
go generate ./...               # 生成 wire 依赖注入
make build BINS=shop-apiserver  # 构建
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml   # 运行
```

构建产物路径：`_output/platforms/linux/amd64/shop-apiserver`

---

## 8. 配置管理

- 配置文件：`configs/shop-apiserver.yaml`
- 启动时用 `-c` 指定配置路径（默认路径 `~/.shop-app/` 不存在）。
- **敏感信息（密码、jwt-key）禁止提交**：当前 `jwt-key: xxxxxxxxxxxx` 为占位符，生产环境必须替换并通过环境变量或挂载注入。
- MariaDB 连接信息见 [Docker 本地测试规则](./docker-local-test.md)。

---

## 9. go.mod 维护红线

- ✅ 允许：`go get <具体版本>`、`go mod tidy`
- ❌ 禁止：`go get -u`、`go get ...@latest`（onexstack 相关）
- ❌ 禁止：删除 `replace google.golang.org/grpc => v1.64.0`（polaris 兼容所需）
- ❌ 禁止：手写 `replace ... => /本地路径` 指令

详见 [踩坑记录](./pitfalls.md)。
