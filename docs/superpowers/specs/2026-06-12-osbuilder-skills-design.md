# osbuilder Skill 拆分设计

> 日期: 2026-06-12
> 状态: 待审阅

## 背景

osbuilder 是 OneX 技术栈的 CLI 脚手架工具，功能覆盖项目生成、增量资源添加、语义版本管理、辅助工具等。需要将其知识按使用目的拆分为独立 skill，便于 AI 助手在不同场景下精准调用。

## 设计原则

1. **按使用目的拆分**，而非按命令映射——每个 skill 有明确的"什么时候用"
2. **查阅参考与操作流程分离**——reference 纯查字典，操作 skill 纯照着做
3. **以实际代码为准**——文档与代码冲突时以代码为准
4. **交叉引用而非重复**——操作 skill 需要查阅细节时引用 reference

## 拆分方案：4 Skill

### Skill 1: `osbuilder-reference`

**性质：** 查阅参考（Reference）

**Frontmatter:**
```yaml
---
name: osbuilder-reference
description: Use when looking up osbuilder config types, available options, template functions, architecture, or auxiliary command usage
---
```

**章节结构：**

| # | 章节 | 内容要点 |
|---|------|---------|
| 1 | Overview | osbuilder 是 OneX 技术栈的 CLI 脚手架工具，本文档为查阅型参考 |
| 2 | Architecture Overview | 入口 → cmd → types → helper → file 的调用链 |
| 3 | Project Config Types | Project/Metadata/ImageConfig 字段说明（YAML key → Go field → 说明） |
| 4 | Component Types | WebServer/JobServer/MQServer/CLITool 的字段和默认值 |
| 5 | ProjectGen（派生字段） | WorkDir/ProjectName/ModuleName/APIVersion 等计算逻辑 |
| 6 | Available Options（以代码为准） | 实际可用的框架/存储/部署/Makefile/Dockerfile/服务注册选项表 |
| 7 | REST Naming Convention | REST/RESTGen 的命名变体计算（kind path → SingularName/PluralName/SingularLower 等） |
| 8 | Template System | RenderTemplate 流程、FuncMap 函数列表、Pairs() 机制、statik 嵌入 |
| 9 | AST Injection | file/method.go（Go AST 注入）和 file/proto.go（proto 注入）的机制 |
| 10 | Auxiliary Commands | addlicense/sysload/cleanup-zombies/upgrade/version 快速参考 |

**文档 vs 代码差异记录（重要）：**

| 项目 | 文档声称 | 代码实际 |
|------|---------|---------|
| Web 框架 | gin/grpc/grpc-gateway/kratos/go-zero/kitex/heartz/onex | 仅 gin/grpc 可用（其余注释） |
| 存储后端 | memory/mariadb/redis/sqlite/postgresql/mongo/etcd | 仅 memory/mariadb/mysql/sqlite/postgresql 可用（redis/mongo/etcd 注释） |
| 服务注册 | none/polaris/eureka/consul/nacos | 仅 none/polaris 可用（其余注释） |
| MQ 框架 | 多种 | 仅 kafka 可用 |
| 应用类型 | webserver/watch/cli | 仅 webserver 可用（watch/cli 注释） |

---

### Skill 2: `osbuilder-init-project`

**性质：** 初始化操作指南（Technique）

**Frontmatter:**
```yaml
---
name: osbuilder-init-project
description: Use when creating a brand new Go project from scratch with osbuilder, with no existing project content
---
```

**章节结构：**

| # | 章节 | 内容要点 |
|---|------|---------|
| 1 | Overview | 从零创建新 Go 项目的完整流程 |
| 2 | When to Use | 完全没有项目内容时使用；已有项目时用 osbuilder-iterative-dev |
| 3 | Prerequisites | osbuilder 已安装、理解 YAML 配置 |
| 4 | Step 1: Prepare Config | 编写 project.yaml（最小示例 + 各组件类型示例） |
| 5 | Step 2: Create Project | `osbuilder create project <DIR> --config <yaml>` 的用法和参数 |
| 6 | Step 3: Quickstart (Optional) | `osbuilder create quickstart` 一键生成含多种组件的示例项目 |
| 7 | What Gets Generated | 生成的项目目录结构说明、各组件对应生成的文件 |
| 8 | Post-creation Steps | cd → make deps → make protoc → go mod tidy → go generate → make build |
| 9 | Configuration Defaults | correctProjectConfig() 的自动修正逻辑（空字段默认值、MySQL→MariaDB、非 GRPC 清空 GRPCServiceName 等） |
| 10 | Common Mistakes | 配置了被注释的框架、模块路径格式错误、目录已存在、makefileMode=none 需手动构建 |
| 11 | Internal Flow | Complete → Validate → Run → Generate 的调用链（引用 osbuilder-reference 查详细字段） |

---

### Skill 3: `osbuilder-iterative-dev`

**性质：** 迭代开发操作指南（Technique）

**Frontmatter:**
```yaml
---
name: osbuilder-iterative-dev
description: Use when adding REST API resources, CLI commands, async jobs, or Kafka consumers to an existing osbuilder-generated project
---
```

**章节结构：**

| # | 章节 | 内容要点 |
|---|------|---------|
| 1 | Overview | 向已有项目增量添加资源的完整流程 |
| 2 | When to Use | 已有 osbuilder 生成的项目，需要添加新资源时 |
| 3 | Prerequisites | 项目已创建、PROJECT 文件存在 |
| 4 | Add REST API (`create api`) | 命令用法、--kinds/--binary-name/--job-server/--force 参数、kind path 格式 |
| 5 | Add CLI Command (`create cmd`) | 命令用法和参数 |
| 6 | Add Async Job (`create job`) | 命令用法和参数 |
| 7 | Add MQ Consumer (`create mq`) | 命令用法和参数 |
| 8 | What Gets Generated Per Command | 各命令生成的文件列表 |
| 9 | Side Effects (AST/Proto Injection) | create api 的副作用：修改 store.go（IStore 接口+store 结构体方法）、修改 biz.go（IBiz 接口+biz 结构体方法+import）、修改 proto 文件（import+RPC 方法） |
| 10 | Post-addition Steps | make protoc → go mod tidy → make build |
| 11 | Common Mistakes | kind path 不用 snake_case、重复添加已存在的 kind、AST 注入后格式化错误 |
| 12 | Internal Flow | Load PROJECT → Complete → Validate → RenderTemplate → AddNewMethod/AddNewGRPCMethod（引用 osbuilder-reference 查 AST 机制详情） |

---

### Skill 4: `osbuilder-semver`

**性质：** 版本发布操作指南（Technique）

**Frontmatter:**
```yaml
---
name: osbuilder-semver
description: Use when managing semantic versioning, creating version tags, generating changelogs, or executing release pipelines with osbuilder
---
```

**章节结构：**

| # | 章节 | 内容要点 |
|---|------|---------|
| 1 | Overview | osbuilder 内置的语义版本管理引擎 |
| 2 | When to Use | 需要打版本标签、递增版本号、生成变更日志、执行完整发布 |
| 3 | Command Reference | semver tag/bump/changelog/release/check 的用法 |
| 4 | Release Pipeline | semver release 的完整流水线：gitcheck → fetchtag → nextsemver → bump → gittag → changelog → gitcommit → nextcommit → push |
| 5 | Task System | semver/task/ 下的 Runner 接口和各 task 职责 |
| 6 | Hook System | before/after 钩子：beforebump/afterbump/beforechangelog/afterchangelog/beforetag/aftertag |
| 7 | GPG Signing | GPG 密钥导入和签名机制 |
| 8 | Uplift Config | semver/config/uplift.go 的配置解析 |
| 9 | Version Parsing | semver/semver/parser.go 和 version.go |
| 10 | Common Mistakes | Git 工作区不干净、GPG 密钥未配置、conventional commit 格式不对 |
| 11 | Internal Architecture | semver engine 的代码组织（context/config/gpg/semver/task/version） |

## 交叉引用关系

```
osbuilder-init-project ──→ osbuilder-reference (查配置字段、可用选项、默认值)
osbuilder-iterative-dev ──→ osbuilder-reference (查 REST 命名规则、AST 机制、Pairs 映射)
osbuilder-semver ──→ (自闭环，内部实现文档齐全)
osbuilder-reference ──→ (被引用，纯查阅)
```

## Skill 文件位置

```
~/.claude/skills/
  osbuilder-reference/
    SKILL.md
  osbuilder-init-project/
    SKILL.md
  osbuilder-iterative-dev/
    SKILL.md
  osbuilder-semver/
    SKILL.md
```
