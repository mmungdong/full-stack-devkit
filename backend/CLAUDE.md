# Backend Go 项目规范

本项目基于 onexstack 技术栈，使用 osbuilder 脚手架生成，遵循简洁架构（Clean Architecture）。

---

## 1. 架构设计规范

### 1.1 核心架构：简洁架构四层分层

采用简洁架构（Clean Architecture），按依赖规则分为四层：

| 层 | 职责 | 对应标准层 |
|---|---|---|
| **Handler 层** | 接收 HTTP/gRPC 请求；参数解析、参数校验、业务逻辑分发、请求返回 | Controller |
| **Biz 层** | 具体业务逻辑实现；按 REST 资源类型分模块 | Usecases |
| **Store 层** | 数据访问层，仅做 CRUD，不封装业务逻辑 | Frameworks & Drivers |
| **Model 层** | 数据结构定义，GORM Model | Entities |

### 1.2 依赖规则

- **代码依赖由上向下，单向单层依赖**：Handler → Biz → Store → Model
- 下层不感知上层的任何对象
- 内层组件提供能力（接口），外层组件调用
- **开发顺序**：先 Store → 再 Biz → 最后 Handler

### 1.3 接口通信

- 各层通过接口通信，支持 Mock 测试
- 每层入口使用抽象工厂模式（如 `IStore`/`IBiz` 接口）
- 使用 `var _ Interface = (*impl)(nil)` 进行编译期接口检查
- 使用 `sync.Once` 确保单例

---

## 2. 目录结构规范

### 2.1 采用 project-layout 目录结构

```
backend/
├── cmd/                        # 程序入口
│   └── <app>-apiserver/        # API 服务器入口
│       ├── main.go             # 精简入口，导入 automaxprocs
│       └── app/                # 初始化配置和框架代码（控制面）
├── internal/                   # 私有代码（不可外部导入）
│   └── <app>/
│       ├── apiserver.go        # 服务器核心结构体
│       ├── biz/                # 业务逻辑层
│       │   ├── biz.go          # IBiz 接口 + 工厂
│       │   └── v1/             # API v1 版本
│       │       └── <resource>/ # 按资源分目录
│       ├── handler/            # 请求处理层
│       │   ├── http/           # HTTP Handler
│       │   └── grpc/           # gRPC Handler
│       ├── store/              # 数据访问层
│       │   ├── store.go        # IStore 接口 + 工厂
│       │   └── <resource>/     # 按资源分目录
│       ├── pkg/                # 应用内共享包
│       │   ├── conversion/     # 结构体类型转换
│       │   ├── validation/     # 请求参数校验
│       │   └── middleware/     # 中间件
│       └── wire.go             # Wire 依赖注入
├── pkg/                        # 可外部导入的公共包
│   └── version/                # 版本信息
├── configs/                    # 配置文件模板
├── scripts/                    # 构建/部署脚本
│   └── make-rules/             # 结构化 Makefile 片段
├── api/                        # Protobuf 定义
├── docs/                       # 项目文档
└── _output/                    # 构建产物（.gitignore）
```

### 2.2 命名规范

- **项目名**：纯小写精短名字，过长用中杠线分割（如 `controller-manager`）
- **组件名**：加项目简写前缀（如 `mb-apiserver`，mb = miniblog）
- **GORM Model**：`<表名大驼峰>M`（如 `UserM`、`PostM`）
- **接口**：`I` 前缀（如 `IStore`、`IBiz`）
- **错误码**：`<平台级>.<资源级>` 两级格式（如 `NotFound.UserNotFound`）
- **转换函数**：`<Source>TypeTo<Target>Type`（如 `PostModelToPostV1`）
- **校验函数**：`Validate<请求参数结构体名>`（如 `ValidateCreateUserRequest`）
- **工厂函数**：`NewXxx()` 创建实例

### 2.3 空目录规范

- 空目录用 `.keep` 文件让 Git 追踪
- 提前创建的标准目录：`cmd/`、`configs/`、`docs/`、`scripts/`

---

## 3. 代码规范

### 3.1 社区规范参考

遵循以下社区规范（按优先级排序）：
1. Uber Go Style Guide
2. CodeReviewComments
3. Effective Go
4. Google Style Guides
5. Kubernetes Code Conventions

### 3.2 编程哲学

- **面向接口编程**：在可能有多处实现的地方使用接口；接口解耦上下游、提高可测性
- **面向"对象"编程**：用结构体实现类/封装，用匿名嵌套实现组合（替代继承），用接口实现多态
- **组合优于继承**：通过嵌入 + 重写方法实现扩展

### 3.3 SOLID 原则

| 原则 | Go 实践 |
|---|---|
| **S** 单一职责 | 函数/方法功能单一；拆分耦合函数 |
| **O** 开闭原则 | 通过组合+重写方法扩展，而非修改现有代码 |
| **L** 里氏替换 | 通过接口实现，子类型可替换父类型 |
| **I** 接口隔离 | 将臃肿接口拆分为更小更具体的接口 |
| **D** 依赖倒置 | 高层模块依赖接口而非具体实现 |

### 3.4 静态代码检查

- 使用 **golangci-lint** 进行静态检查
- 配置文件：`.golangci.yaml`
- 本地和 CI/CD 使用相同命令确保一致性
- Makefile 集成：`make lint`

### 3.5 代码格式化

- 使用 `gofumpt`（比 gofmt 更严格）
- Makefile 集成：`make format`

---

## 4. 接口规范

### 4.1 REST 接口规范

- HTTP 接口采用 REST 规范，**一切皆资源**
- 使用标准 HTTP 方法：GET（查询）/ POST（创建）/ PUT（更新）/ DELETE（删除）
- 路由注册严格遵循 RESTful 规范：

```
POST   /v1/users          - 创建用户
PUT    /v1/users/:userID   - 更新用户
DELETE /v1/users/:userID   - 删除用户
GET    /v1/users/:userID   - 查询用户详情
GET    /v1/users           - 查询用户列表
```

- 路由参数使用 Protobuf 注入 Go 标签：`@gotags: uri:"userID"`

### 4.2 RPC 接口规范

- 方法名使用大写驼峰命名法（如 `UpdateUser`）
- 方法名见名知义
- 方法参数只能有一个，类型为结构体指针（如 `*UpdateUserRequest`）
- 方法返回值只能有一个，类型为结构体指针（如 `*UpdateUserResponse`）
- Mutation 接口需保证**幂等性**

### 4.3 请求处理流程

1. **ReadRequest**：参数绑定 → Default() 设置默认值 → validators 参数校验
2. 调用 Biz 层方法处理业务
3. **WriteResponse**：成功返回 200 + data；失败返回对应 HTTP 状态码 + ErrorResponse

---

## 5. 日志规范

### 5.1 日志包设计

- 底层封装 **zap**，包路径：`internal/pkg/log/`
- 类型定义为不可导出的 `zapLogger`
- 接口只保留核心方法：`Debugw`/`Infow`/`Warnw`/`Errorw`/`Panicw`/`Fatalw`/`Sync`
- 全局变量 `std` 用 `sync.Mutex` 保护并发

### 5.2 日志记录规范

- **使用结构化记录方式**（`Infow`/`Errorw` 等），不使用格式化方式
- 错误日志在**最初发生错误的位置**打印
- 上层如不需要添加信息，直接返回下层 Error，不再重复打印
- 嵌套 Error 只在最初位置打印一次

### 5.3 日志内容规范

- 不输出敏感信息（密码、密钥等）
- 成功日志格式：`<动词> + <一些事>`（如 `Create user successfully`）
- 失败日志格式：`Fail to <动词> + <一些事>`（如 `Fail to create user`）
- 日志内容以大写字母开头
- 最好包含 RequestID、User、Action

### 5.4 日志级别使用

| 级别 | 使用场景 |
|---|---|
| Debug | 开发调试信息，支持动态开关 |
| Info | 关键业务流程记录 |
| Warn | 非预期但可恢复的情况 |
| Error | 错误，需要关注 |
| Panic | 不可恢复的严重错误 |
| Fatal | 致命错误，程序退出 |

### 5.5 日志打印位置

- ✅ 分支语句处打印
- ✅ 错误产生的最原始位置打印
- ✅ 接口请求处打印（需加 Filter 逻辑）
- ❌ 不在循环中打印

### 5.6 日志保存

- 容器化部署优先输出到**标准输出**（stdout）
- 时间格式：`2006-01-02 15:04:05.000`
- MessageKey 改为 `message`，TimeKey 改为 `timestamp`
- `zap.AddCallerSkip(2)` 跳过封装调用栈

---

## 6. 错误规范

### 6.1 错误包设计

- 包名：`errorsx`（避免与标准库 `errors` 冲突，x 表示扩展）
- 核心结构体 `ErrorX`：
  - `Code`：HTTP 状态码
  - `Reason`：业务错误码（字符串，见名知义）
  - `Message`：错误信息
  - `Metadata`：元数据
- 实现 `Error()`、`GRPCStatus()`、`Is()` 方法
- `WithMessage`/`WithMetadata`/`KV` 方法支持**链式调用**
- `FromError` 支持 error → *ErrorX 转换，兼容 gRPC status

### 6.2 错误码设计

- 采用**两级错误码**：`<平台级>.<资源级>`（参考腾讯云 API 3.0）
- 错误码语义化：如 `InvalidArgument.UsernameInvalid`、`NotFound.UserNotFound`
- 错误码采用**字符串**（非整数），见名知义

### 6.3 错误返回格式

```json
{
  "code": 404,
  "reason": "NotFound.UserNotFound",
  "message": "User not found.",
  "metadata": {},
  "Trace-Id": "xxx"
}
```

### 6.4 错误返回规范

- 所有接口返回 `errorsx.ErrorX` 类型
- 失败时返回对应 HTTP/gRPC 状态码 + 业务错误码
- 在错误**最原始位置**用 `errno.ErrXXX` 返回，其他位置直接透传
- 不在日志中打印敏感信息
- HTTP 状态码不建议映射太多，保持简洁
- 返回简洁的错误信息，不建议返回复杂字段（如 HelpUrl）

### 6.5 预定义错误

- 保存在 `internal/pkg/errno/` 下
- `Is()` 方法比较 Code 和 Reason，不比较 Message

---

## 7. 提交与版本规范

### 7.1 Git 提交规范（Conventional Commits）

- 采用 **Angular 提交规范**
- 格式：`<type>[optional scope]: <description>` + Body + Footer
- **type 类型**：
  - `feat`：新功能
  - `fix`：修复 bug
  - `docs`：文档变更
  - `style`：代码格式（不影响逻辑）
  - `refactor`：重构
  - `perf`：性能优化
  - `test`：测试相关
  - `chore`：构建/工具变更
  - `ci`：CI 配置变更
  - `revert`：回滚
- **subject 规则**：祈使句、不全大写开头、末尾不加句号
- **Footer**：`Closes #xxx` / `BREAKING CHANGE`

### 7.2 Git 工作流

- 采用 **GitHub Flow**
- 基于 main 新建功能分支开发，完成后通过 PR 合并
- PR 流程支持 Code Review

### 7.3 语义化版本规范（SemVer）

- 版本号格式：`X.Y.Z`（MAJOR.MINOR.PATCH）
- **MAJOR**：不兼容的 API 修改
- **MINOR**：向下兼容的功能新增
- **PATCH**：向下兼容的问题修正
- 先行版本号：`X.Y.Z-alpha`、`X.Y.Z-beta`
- 首个开发版本：`0.1.0`；首次稳定发布：`1.0.0`
- 版本号递增规则：
  - `fix` 类型 commit → PATCH+1
  - `feat` 类型 commit → MINOR+1
  - `BREAKING CHANGE` → MAJOR+1

### 7.4 版本信息注入

- 通过 `go build -ldflags "-X importpath.name=value"` 编译时注入版本信息
- 版本信息保存在 `pkg/version/version.go`
- Info 结构体：gitVersion、gitCommit、gitTreeState、buildDate、goVersion、compiler、platform
- 支持 `--version` 和 `--version=raw` 两种格式输出

---

## 8. 编程实现规范

### 8.1 Store 层规范

- **仅执行 CRUD**，不封装业务逻辑
- 查询条件通过 `where.Options` 灵活配置，不在 Store 层封装多个查询方法
- 删除操作实现幂等（`gorm.ErrRecordNotFound` 时返回 nil）
- 资源标准接口方法顺序：Create → Update → Delete → Get → List + Expansion
- 抽象工厂模式：`IStore` 接口提供 `User()`/`Post()` 等方法返回对应 Store 实例
- GORM Model 命名：`<表名大驼峰>M`
- 用 `gorm.io/gen` 自动生成 GORM Model
- 唯一标识符格式：`<资源前缀>-<6位随机数>`（如 `user-uvalgf`）

### 8.2 Biz 层规范

- 按资源类型分目录：`internal/<app>/biz/v1/<resource>/`
- 保留 v2 目录，为 API 版本升级预留扩展能力
- 使用 `copier.Copy` 简化结构体赋值
- 使用 `errgroup` 并发处理 + `sync.Map` 保证并发安全
- `eg.SetLimit` 限制并发数量
- 数据类型转换统一在 `internal/<app>/pkg/conversion/` 管理
- 转换函数命名：`<Source>TypeTo<Target>Type`

### 8.3 Handler 层规范

- **不执行任何业务逻辑**，直接转发到 Biz 层
- HTTP Handler 目录：`internal/<app>/handler/http/`
- gRPC Handler 目录：`internal/<app>/handler/grpc/`
- HTTP Handler 使用 `core.HandleJSONRequest`/`HandleQueryRequest` 语法糖函数

### 8.4 请求参数校验

- 校验逻辑集中保存在 `internal/<app>/pkg/validation/` 下
- 校验函数命名：`Validate<请求参数结构体名>(ctx, rq) error`
- Validator 结构体注入需要的依赖（如 `store.IStore`）
- `ValidateAllFields`：校验所有字段
- `ValidateSelectedFields`：校验指定字段

### 8.5 配置规范

- 配置项结构体 `ServerOptions`，字段加 `json` 和 `mapstructure` 标签
- 三种配置方式（优先级由低到高）：默认值 → 命令行选项 → 配置文件
- 命令行选项只添加核心配置（如 `--config`），其他走配置文件
- 配置文件格式首选 **YAML**
- 环境变量前缀设为项目名（如 `MINIBLOG_`）
- 配置文件中每个配置项必须添加详细说明
- 初始化配置 vs 运行配置分离：`cmd/<app>/app/` 控制面 vs `internal/<app>/` 数据面

### 8.6 main.go 规范

- main 文件保持简洁，具体实现放在 `cmd/<app>/app/` 下
- 导入 `_ "go.uber.org/automaxprocs"` 自动设置 GOMAXPROCS
- main 包不包含不可导出的标识符，以支持 `go install`
- 错误退出用 `os.Exit(1)`

---

## 9. 测试规范

### 9.1 编写可测试的代码

- 将依赖（数据库、第三方服务）抽象成**接口**
- 测试时传入 mock/fake 类型解耦依赖
- 减少函数依赖，编写功能单一、职责分明的函数

### 9.2 Mock 工具

| 层级 | Mock 工具 | 用途 |
|---|---|---|
| Store 层 | `sqlmock` | 模拟数据库连接 |
| Store 层 | `httpmock` | 模拟 HTTP 请求 |
| Biz 层 | `golang/mock` | 模拟 Store 层接口 |
| Handler 层 | `golang/mock` | 模拟 Biz 层接口 |

### 9.3 测试覆盖

- 使用 `gotests` 工具自动生成单元测试代码
- 定期检查覆盖率：`go test -race -cover -coverprofile=./coverage.out -timeout=10m -short -v ./...`
- **经常变动的函数，覆盖率要达到 100%**
- 单元测试有效性比覆盖率更重要
- 断言包使用 `github.com/stretchr/testify/assert`

### 9.4 性能要求

- 接口延时建议 **< 500ms**
- 使用 `pprof` 工具进行性能优化
- 进行并发优化和压力测试

---

## 10. 依赖与工具

### 10.1 核心依赖

| 类别 | 包 | 用途 |
|---|---|---|
| 命令行 | `spf13/cobra` | 应用启动框架 |
| 命令行 | `spf13/pflag` | 命令行参数解析 |
| 配置 | `spf13/viper` | 多源配置读取 |
| 日志 | `go.uber.org/zap` | 高性能结构化日志 |
| ORM | `gorm.io/gorm` | 数据库 ORM |
| Model 生成 | `gorm.io/gen` | GORM Model 自动生成 |
| 校验 | `go-playground/validator` | 结构体标签校验 |
| 依赖注入 | `google/wire` | 编译时依赖注入 |
| 容器优化 | `uber-go/automaxprocs` | 自动设置 GOMAXPROCS |

### 10.2 开发工具

| 工具 | 用途 |
|---|---|
| `osbuilder` | 项目脚手架生成 |
| `golangci-lint` | 静态代码检查 |
| `air-verse/air` | 开发热加载 |
| `gofumpt` | 代码格式化 |
| `wire` | 依赖注入代码生成 |
| `mockgen` | Mock 代码生成 |
| `gotests` | 测试代码生成 |
| `protoc-go-inject-tag` | Protobuf 注入 Go 标签 |
| `addlicense` | 自动添加版权头 |

### 10.3 Wire 依赖注入规范

- Injector 保存在 `internal/<app>/wire.go`（加 `//+build wireinject` 标签）
- 运行 `wire .` 生成 `wire_gen.go`
- 依赖更新后执行 `go generate ./...`
- 三个核心概念：Provider（构造函数）、ProviderSet（集合）、Injector（注入入口）

### 10.4 Makefile 规范

- 采用结构化 Makefile：
  ```
  ├── Makefile                    # 聚合入口
  └── scripts/make-rules/
      ├── common.mk               # 通用变量
      ├── golang.mk               # Go 相关
      ├── generate.mk             # 代码生成
      └── tools.mk                # 工具安装
  ```
- `make help` 自动生成帮助信息
- 常用目标：`build`/`format`/`lint`/`test`/`cover`/`tidy`/`clean`/`add-copyright`

### 10.5 新增 REST 资源标准步骤

1. 定义 API 接口（Protobuf 文件）
2. 编译 Protobuf 文件
3. 创建数据库表 + 生成 GORM Model
4. 完善默认值设置方法
5. 实现请求参数校验方法
6. 实现 Store 层代码（嵌入标准 Store）
7. 实现 Model/Proto 转换函数
8. 实现 Biz 层代码
9. 实现 Handler 层代码
