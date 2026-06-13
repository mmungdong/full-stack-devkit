# Backend 项目初始化与规范落地实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 使用 osbuilder 初始化 backend Go 项目，确保生成的项目结构与 `backend/CLAUDE.md` 中定义的规范一致。

**Architecture:** 基于 onexstack 技术栈，使用 osbuilder 脚手架生成包含 gin WebServer + mariadb 存储的 Go 项目。项目遵循简洁架构四层分层（Handler→Biz→Store→Model），采用 Wire 依赖注入、zap 日志、errorsx 错误包。

**Tech Stack:** Go 1.25, osbuilder v0.15.0, gin, gRPC, GORM/MariaDB, Wire, zap, cobra/pflag/viper

---

## 前置条件

- ✅ osbuilder 二进制已就绪：`backend/bin/osbuilder`（v0.15.0）
- ✅ Go 1.25.11 已安装
- ✅ `backend/CLAUDE.md` 规范文档已写入
- ⚠️ 当前 `backend/` 目录下有 `bin/` 和 `references/` 目录，需要在临时位置生成项目后迁移

## 文件影响范围

| 操作 | 文件路径 | 说明 |
|------|---------|------|
| 创建 | `backend/onexstack.yaml` | osbuilder 项目配置文件 |
| 创建 | `backend/cmd/` | 程序入口目录 |
| 创建 | `backend/internal/` | 私有业务代码 |
| 创建 | `backend/pkg/` | 公共包 |
| 创建 | `backend/configs/` | 配置文件模板 |
| 创建 | `backend/scripts/` | 构建脚本 |
| 创建 | `backend/api/` | Protobuf 定义 |
| 创建 | `backend/docs/` | 项目文档 |
| 创建 | `backend/Makefile` | 构建管理 |
| 创建 | `backend/go.mod` | Go 模块定义 |
| 创建 | `backend/.golangci.yaml` | 静态检查配置 |
| 创建 | `backend/.air.toml` | 热加载配置 |
| 创建 | `backend/.gitignore` | Git 忽略规则 |
| 保留 | `backend/CLAUDE.md` | 已写入的规范文档 |
| 保留 | `backend/bin/osbuilder` | 脚手架工具 |
| 保留 | `backend/references/` | 参考源码 |

---

### Task 1: 编写 osbuilder 项目配置文件

**Files:**
- Create: `backend/onexstack.yaml`

- [ ] **Step 1: 创建项目配置文件**

在 `backend/` 目录下创建 `onexstack.yaml` 配置文件：

```yaml
scaffold: onex
version: v1
metadata:
  modulePath: github.com/mungdong/devkit
  author: "mungdong"
  email: mungdong@example.com
  deploymentMethod: docker
  makefileMode: structured
  image:
    registryPrefix: docker.io/mungdong
    dockerfileMode: combined
    distrolessMode: auto
webServers:
  - binaryName: dk-apiserver
    webFramework: gin
    storageType: mariadb
    withHealthz: true
    withUser: true
    withOTel: true
```

**配置说明：**
- `modulePath`: `github.com/mungdong/devkit`（项目简写 dk = devkit）
- `binaryName`: `dk-apiserver`（项目前缀 dk）
- `webFramework`: `gin`（CLAUDE.md 规范的 HTTP 框架）
- `storageType`: `mariadb`（生产级存储，osbuilder 会自动将 mysql 替换为 mariadb）
- `withHealthz`: 启用健康检查（前端 Next.js 需要访问 `/healthz`）
- `withUser`: 启用用户模块（含认证/授权，CLAUDE.md 中的安全规范）
- `withOTel`: 启用 OpenTelemetry（可观测性）
- `makefileMode`: `structured`（CLAUDE.md 规范的结构化 Makefile）
- `deploymentMethod`: `docker`（开发阶段用 Docker 部署）

- [ ] **Step 2: 验证配置文件语法**

```bash
cat backend/onexstack.yaml
```

确认 YAML 格式正确，字段完整。

---

### Task 2: 使用 osbuilder 生成项目

**Files:**
- Create: 整个项目结构（约 50+ 文件）

- [ ] **Step 1: 在临时目录生成项目**

因为 `backend/` 目录已存在（含 `bin/` 和 `references/`），osbuilder 的 `create project` 不允许在非空目录执行。先在临时目录生成：

```bash
OSBUILDER=/home/mungdong/workspace/full-stack-devkit/backend/bin/osbuilder
TMPDIR=$(mktemp -d)
$OSBUILDER create project "$TMPDIR" --config /home/mungdong/workspace/full-stack-devkit/backend/onexstack.yaml
```

预期输出：项目文件生成成功，打印 Getting Started 提示。

- [ ] **Step 2: 检查生成的项目结构**

```bash
find "$TMPDIR" -type f ! -path "*/.git/*" | sort
```

确认生成了以下关键目录和文件：
- `cmd/dk-apiserver/main.go`
- `internal/apiserver/` 下的 biz/handler/store/model 目录
- `pkg/version/version.go`
- `configs/dk-apiserver.yaml`
- `Makefile`
- `go.mod`
- `.golangci.yaml`
- `.air.toml`
- `.gitignore`

- [ ] **Step 3: 将生成的文件迁移到 backend/ 目录**

将临时目录中生成的文件（排除已存在的 `bin/`、`references/`、`CLAUDE.md`）复制到 `backend/`：

```bash
cd "$TMPDIR"
for item in *; do
  # 跳过 backend/ 已有的目录
  case "$item" in
    bin|references) continue ;;
  esac
  cp -r "$item" /home/mungdong/workspace/full-stack-devkit/backend/
done
```

- [ ] **Step 4: 清理临时目录**

```bash
rm -rf "$TMPDIR"
```

---

### Task 3: 编译 Protobuf 文件

**Files:**
- Modify: `api/` 下的 proto 生成文件
- Create: 生成的 `.pb.go` 文件

- [ ] **Step 1: 编译所有 protobuf 文件**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
make protoc
```

预期输出：protoc 编译成功，生成 `.pb.go`、`_grpc.pb.go` 等文件。

> **注意**：如果 `make protoc` 失败，可能需要先安装 protoc 工具链。检查命令：
> ```bash
> which protoc && protoc --version
> which protoc-gen-go && protoc-gen-go --version
> ```
> 如未安装，参考 osbuilder 生成的 `scripts/` 下的安装脚本。

- [ ] **Step 2: 验证 proto 编译产物**

```bash
find /home/mungdong/workspace/full-stack-devkit/backend -name "*.pb.go" | head -10
```

确认生成了 `.pb.go` 文件。

---

### Task 4: 整理依赖并构建项目

**Files:**
- Modify: `go.mod`、`go.sum`

- [ ] **Step 1: 解决已知依赖问题**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
go get cloud.google.com/go/compute@latest
go get cloud.google.com/go/compute/metadata@latest
```

这是 osbuilder 已知的 ambiguous import 问题。

- [ ] **Step 2: 整理 Go 模块依赖**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
go mod tidy
```

预期输出：无错误，`go.sum` 更新。

- [ ] **Step 3: 运行 go generate**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
go generate ./...
```

预期输出：Wire 生成 `wire_gen.go`，mockgen 生成 mock 文件等。

- [ ] **Step 4: 构建项目**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
make build
```

预期输出：`_output/dk-apiserver` 二进制文件生成成功。

- [ ] **Step 5: 验证构建产物**

```bash
ls -la /home/mungdong/workspace/full-stack-devkit/backend/_output/
./_output/dk-apiserver --version
```

确认版本信息输出正确。

---

### Task 5: 验证项目结构与 CLAUDE.md 规范一致性

**Files:**
- Read: `backend/CLAUDE.md`
- Read: 生成的各层代码文件

- [ ] **Step 1: 验证四层架构分层**

```bash
# 检查 Handler 层
find /home/mungdong/workspace/full-stack-devkit/backend/internal -path "*/handler/*" -name "*.go" | head -5

# 检查 Biz 层
find /home/mungdong/workspace/full-stack-devkit/backend/internal -path "*/biz/*" -name "*.go" | head -5

# 检查 Store 层
find /home/mungdong/workspace/full-stack-devkit/backend/internal -path "*/store/*" -name "*.go" | head -5

# 检查 Model 层
find /home/mungdong/workspace/full-stack-devkit/backend/internal -path "*/model/*" -name "*.go" | head -5
```

预期：四个层级目录均存在，符合 CLAUDE.md 第 1 节规范。

- [ ] **Step 2: 验证目录结构符合 project-layout**

```bash
ls -d /home/mungdong/workspace/full-stack-devkit/backend/{cmd,internal,pkg,configs,scripts,api,docs} 2>/dev/null
```

预期：所有目录均存在，符合 CLAUDE.md 第 2 节规范。

- [ ] **Step 3: 验证核心依赖**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
grep -E "cobra|pflag|viper|zap|gorm|wire|automaxprocs" go.mod | head -10
```

预期：包含 CLAUDE.md 第 10 节要求的所有核心依赖。

- [ ] **Step 4: 验证 Makefile 结构**

```bash
cat /home/mungdong/workspace/full-stack-devkit/backend/Makefile | head -20
ls /home/mungdong/workspace/full-stack-devkit/backend/scripts/make-rules/ 2>/dev/null
```

预期：结构化 Makefile，包含 `scripts/make-rules/` 下的 `.mk` 文件，符合 CLAUDE.md 第 10.4 节规范。

- [ ] **Step 5: 验证 Wire 依赖注入**

```bash
find /home/mungdong/workspace/full-stack-devkit/backend/internal -name "wire.go" -o -name "wire_gen.go" | head -5
```

预期：`wire.go` 和 `wire_gen.go` 存在，符合 CLAUDE.md 第 10.3 节规范。

- [ ] **Step 6: 验证 golangci-lint 配置**

```bash
ls /home/mungdong/workspace/full-stack-devkit/backend/.golangci.yaml
```

预期：`.golangci.yaml` 存在，符合 CLAUDE.md 第 3.4 节规范。

---

### Task 6: 运行快速验证测试

**Files:**
- Read: `configs/dk-apiserver.yaml`

- [ ] **Step 1: 启动 API 服务器**

```bash
cd /home/mungdong/workspace/full-stack-devkit/backend
./_output/dk-apiserver --config configs/dk-apiserver.yaml &
```

> **注意**：如果使用 mariadb 存储类型，需要先启动 MariaDB。可用 Docker 快速启动：
> ```bash
> docker run -d --name devkit-mariadb \
>   -e MYSQL_ROOT_PASSWORD=devkit123 \
>   -e MYSQL_DATABASE=devkit \
>   -p 3306:3306 \
>   mariadb:11
> ```

- [ ] **Step 2: 验证健康检查端点**

```bash
curl http://127.0.0.1:5555/healthz
```

预期：返回 200 + JSON 响应，与前端 `NEXT_PUBLIC_API_BASE_URL` 对接。

- [ ] **Step 3: 验证版本端点**

```bash
curl http://127.0.0.1:5555/version
```

预期：返回版本信息 JSON。

- [ ] **Step 4: 停止服务器**

```bash
pkill dk-apiserver
```

---

### Task 7: 更新 CLAUDE.md 补充项目特定信息

**Files:**
- Modify: `backend/CLAUDE.md`

- [ ] **Step 1: 在 CLAUDE.md 顶部添加项目特定信息**

在 `backend/CLAUDE.md` 的开头（第 1 行之前）插入项目特定信息：

```markdown
# Backend Go 项目规范

本项目基于 onexstack 技术栈，使用 osbuilder 脚手架生成，遵循简洁架构（Clean Architecture）。

## 项目信息

- **项目名**: devkit
- **模块路径**: github.com/mungdong/devkit
- **二进制前缀**: dk
- **API 服务器**: dk-apiserver
- **Go 版本**: 1.25
- **Web 框架**: gin + gRPC
- **存储**: MariaDB (GORM)
- **部署**: Docker

## 常用命令

```bash
make build          # 构建项目
make run            # 运行 API 服务器
make lint           # 静态代码检查
make format         # 代码格式化
make test           # 运行测试
make cover          # 测试覆盖率
make protoc         # 编译 Protobuf
make tidy           # 整理依赖
make clean          # 清理构建产物
make add-copyright  # 添加版权头
make help           # 查看所有 make 目标
```
```

- [ ] **Step 2: 验证 CLAUDE.md 完整性**

```bash
wc -l /home/mungdong/workspace/full-stack-devkit/backend/CLAUDE.md
head -30 /home/mungdong/workspace/full-stack-devkit/backend/CLAUDE.md
```

确认项目信息和常用命令已正确添加。
