# 踩坑记录

本文件记录 shop-app 项目开发中踩过的坑及解决方案，供后续 Agent 与开发者参考。

---

## 坑 1：osbuilder 生成 go.mod 带本地 replace 指令

**现象**

`osbuilder create project` 生成的 `go.mod` 末尾包含：

```go
replace github.com/onexstack/onexstack => /home/colin/workspace/golang/src/github.com/onexstack/onexstack
```

执行 `go mod tidy` 报错：

```
github.com/onexstack/onexstack@v0.0.0-00010101000000-000000000000: replacement directory /home/colin/workspace/golang/src/github.com/onexstack/onexstack does not exist
```

**根因**

osbuilder 模板在生成 `go.mod` 时，硬编码了 osbuilder 作者（孔令飞）本地的 `onexstack` 仓库绝对路径作为 replace 目标。该路径在其他人机器上不存在，导致依赖无法解析。

**解决**

删除该 replace 指令，改为引用远程模块版本（见坑 2 选定版本）。

```diff
- replace github.com/onexstack/onexstack => /home/colin/workspace/golang/src/github.com/onexstack/onexstack
```

---

## 坑 2：osbuilder 脚手架模板与 onexstack 库版本不匹配

**现象**

删除本地 replace 后，按 `go mod tidy` 默认拉取最新 `onexstack@v0.3.29`，执行 `go generate ./...` 报错：

```
wire: undefined: genericoptions.HTTPOptions
wire: *ginServer does not implement server.Server (wrong type for method RunOrDie)
        have RunOrDie()
        want RunOrDie(context.Context)
wire: cannot use c.TLSOptions ... as *SecureServingOptions value
```

**根因**

osbuilder v0.11.1 的项目模板使用的是 **旧版 onexstack API**：

- `server.Server.RunOrDie()` 无参签名
- `genericoptions.HTTPOptions`、`genericoptions.TLSOptions` 类型

而 onexstack 在后续版本做了 breaking change：

- commit `e384535`（2026-02-01 "refactor: update code"）将 `RunOrDie()` 改为 `RunOrDie(ctx context.Context)`，从 **v0.3.20** 起生效。
- commit `c702c77`（2026-03-03 "refactor: update code"）移除 `HTTPOptions`，改用 `SecureServingOptions`，从 **v0.3.24** 起生效。

最新版 v0.3.29 同时含上述两处变更，与模板完全不兼容。

**定位过程**

浅克隆 onexstack 仓库，用 `git log -S` 搜索 API 符号的变更点：

```bash
git clone --filter=blob:none --no-checkout https://github.com/onexstack/onexstack.git
git log --oneline --all -S "RunOrDie()" -- pkg/server/server.go
git tag --contains <commit>   # 确认变更从哪个 tag 开始
```

**解决**

锁定 onexstack 到旧 API 的最后版本 **v0.3.19**：

```bash
go get github.com/onexstack/onexstack@v0.3.19
go mod tidy
go generate ./...
make build BINS=shop-apiserver
```

验证：v0.3.19 同时具备无参 `RunOrDie()` 与 `HTTPOptions`，与脚手架模板匹配，`go generate` 与 `make build` 均通过。

**⚠️ 维护提醒**

- **不要随意执行 `go get -u` / `go get github.com/onexstack/onexstack@latest`**，否则会升级到 v0.3.20+，触发 `RunOrDie` 签名不兼容；升级到 v0.3.24+ 还会触发 `HTTPOptions` 移除。
- 如需升级 onexstack，必须同步等待 osbuilder 发布使用新 API 的模板版本，并重新生成项目或手动迁移调用点。
- 升级 osbuilder 后重新 `osbuilder create project` 前，建议先用 `osbuilder version` 确认模板版本，并核对新模板对应的 onexstack 兼容版本。

---

## 坑 3：Swagger 集成相关

### 3.1 osbuilder 生成的 openapi swagger.json 是空的

**现象**

`api/openapi/apiserver/v1/*.swagger.json` 文件存在，但 `paths: {}` 为空，无任何接口路径。

**根因**

gin 项目没有 gRPC service，proto 里没有 `google.api.http` 注解，protoc-gen-openapiv2 无法生成路径。osbuilder 默认生成的 openapi 文件对 gin 项目无实际价值。

**解决**

改用 [swaggo/swag](https://github.com/swaggo/swag) 从 Go handler 代码注解生成 Swagger 文档，配合 gin-swagger 提供 UI。详见 [开发规范 - Swagger 规范](./development.md#6-swagger-规范)。

### 3.2 swag CLI 与 swag 库版本错位

**现象**

`swag init` 生成的 `docs/apidocs/docs.go` 编译报错：

```
unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
```

**根因**

swag CLI 装的是 v1.16.4，生成代码用了 `LeftDelim/RightDelim` 字段；但 `go get github.com/swaggo/gin-swagger` 间接拉入的 swag 库是 v1.8.12，无此字段。

**解决**

对齐库版本：

```bash
go get github.com/swaggo/swag@v1.16.4
```

### 3.3 swag init 的 `-g` 与 `-d` 路径重复拼接

**现象**

执行 `swag init -g cmd/shop-apiserver/main.go -d ...` 报错：

```
cannot parse source files .../cmd/shop-apiserver/cmd/shop-apiserver/main.go: no such file or directory
```

**根因**

`-g` 指定的入口文件路径是相对 `-d` 第一个目录的，不能带目录前缀，否则会重复拼接。

**解决**

```bash
swag init -g main.go -d cmd/shop-apiserver,internal/apiserver/handler -o docs/apidocs --parseDependency --parseInternal
```

### 3.4 gin-swagger 的函数名

gin-swagger v1.6.1 导出的是 `ginSwagger.WrapHandler(...)`，不是 `WrapH`。

---

## 坑 4：后台进程被 shell 回收

**现象**

在自动化环境中用 `./shop-apiserver &` 后台启动服务，进程很快消失，curl 连不上。

**根因**

某些非交互 shell 退出时会回收子进程组。

**解决**

使用长驻后台执行机制（如 `run_in_background`），或 `nohup ... & disown` 并确认进程脱离会话。

---

## 版本兼容矩阵（实测）

| 组件 | 版本 | 备注 |
|------|------|------|
| osbuilder | v0.11.1 | 当前生成模板的脚手架版本 |
| onexstack | **v0.3.19** | 与 osbuilder v0.11.1 模板兼容的最后版本（旧 API） |
| go | 1.25.11+ | go.mod 声明；`go mod tidy` 可能自动升级 toolchain 至 1.26.x 以满足 k8s.io/apiserver 依赖 |
| grpc replace | v1.64.0 | go.mod 中 `replace google.golang.org/grpc => v1.64.0`，为兼容 polarismesh/grpc-go-polaris，勿删 |
| swag / gin-swagger | v1.16.4 / v1.6.1 | Swagger 文档生成与 UI |

---

## 其他已知非阻塞问题

### proto namespace conflict 警告

运行 `shop-apiserver --version` 时输出：

```
WARNING: proto: file "auth.proto" is already registered
See https://protobuf.dev/reference/go/faq#namespace-conflict
```

来源：onexstack 库内部 proto 注册冲突，属上游已知问题，不影响服务运行，可忽略。

### swag init 的第三方库常量解析 warning

`swag init --parseDependency` 时会输出若干 `failed to evaluate const` warning（来自 mongo-driver/etcd 等第三方库的大整数常量），不影响文档生成，可忽略。
