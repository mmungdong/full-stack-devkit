# osbuilder Skills 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建 4 个 osbuilder skill 文件，按使用目的拆分为查阅参考、初始化项目、迭代开发、语义版本管理

**Architecture:** 每个 skill 独立目录 `~/.claude/skills/<skill-name>/SKILL.md`，reference 纯查阅型，其余三个为操作指南型，通过交叉引用避免内容重复

**Tech Stack:** Markdown skill 文件，YAML frontmatter，遵循 agentskills.io 规范

---

### Task 1: 创建 osbuilder-reference Skill

**Files:**
- Create: `~/.claude/skills/osbuilder-reference/SKILL.md`

- [ ] **Step 1: 创建 skill 目录**

```bash
mkdir -p ~/.claude/skills/osbuilder-reference
```

- [ ] **Step 2: 编写 SKILL.md 完整内容**

核心要点：
- Frontmatter: `name: osbuilder-reference`, `description: Use when looking up osbuilder config types, available options, template functions, architecture, or auxiliary command usage`
- 10 个章节按设计文档顺序：Overview → Architecture → Project Config Types → Component Types → ProjectGen → Available Options → REST Naming → Template System → AST Injection → Auxiliary Commands
- Available Options 必须以代码 `available.go` 为准，标注哪些被注释
- Auxiliary Commands 快速参考覆盖 addlicense/sysload/cleanup-zombies/upgrade/version
- 文档 vs 代码差异记录表

- [ ] **Step 3: 验证 frontmatter 格式**

确认 name 仅含字母/数字/连字符，description 以 "Use when" 开头，总长度 < 1024 字符

- [ ] **Step 4: 验证内容准确性**

对照源码确认：
- `known/available.go` 中的可用选项与 skill 描述一致
- `types/types.go` 中 Project 字段与 skill 描述一致
- `helper/helper.go` 中 FuncMap 函数列表与 skill 描述一致
- `known/known.go` 中常量名与 skill 描述一致

---

### Task 2: 创建 osbuilder-init-project Skill

**Files:**
- Create: `~/.claude/skills/osbuilder-init-project/SKILL.md`

- [ ] **Step 1: 创建 skill 目录**

```bash
mkdir -p ~/.claude/skills/osbuilder-init-project
```

- [ ] **Step 2: 编写 SKILL.md 完整内容**

核心要点：
- Frontmatter: `name: osbuilder-init-project`, `description: Use when creating a brand new Go project from scratch with osbuilder, with no existing project content`
- 11 个章节按设计文档顺序
- When to Use 明确区分：完全没有项目用此 skill，已有项目用 osbuilder-iterative-dev
- Step 1 (Prepare Config) 包含最小 project.yaml 示例和各组件类型示例
- Step 2 (Create Project) 包含命令用法和参数说明
- Step 3 (Quickstart) 包含命令用法和所有 flag
- Configuration Defaults 覆盖 correctProjectConfig() 的所有默认值逻辑
- Internal Flow 引用 osbuilder-reference 查详细字段
- Common Mistakes 包含：配置了被注释的框架、模块路径格式错误、目录已存在、makefileMode=none

- [ ] **Step 3: 验证交叉引用**

确认所有引用 osbuilder-reference 的地方使用格式：`**REQUIRED SUB-SKILL:** Use osbuilder-reference`

- [ ] **Step 4: 验证命令用法准确性**

对照 `create_project.go` 和 `create_quickstart.go` 确认命令参数和用法

---

### Task 3: 创建 osbuilder-iterative-dev Skill

**Files:**
- Create: `~/.claude/skills/osbuilder-iterative-dev/SKILL.md`

- [ ] **Step 1: 创建 skill 目录**

```bash
mkdir -p ~/.claude/skills/osbuilder-iterative-dev
```

- [ ] **Step 2: 编写 SKILL.md 完整内容**

核心要点：
- Frontmatter: `name: osbuilder-iterative-dev`, `description: Use when adding REST API resources, CLI commands, async jobs, or Kafka consumers to an existing osbuilder-generated project`
- 12 个章节按设计文档顺序
- Prerequisites 明确要求 PROJECT 文件存在
- Add REST API 包含 `--kinds`, `--binary-name`, `--job-server`, `--force` 参数，kind path 格式说明
- Add CLI Command 包含 `--kinds`, `--binary-name`, `--force`
- Add Async Job 包含 `--kinds`, `--binary-name`, `--force`
- Add MQ Consumer 包含 `--kinds`, `--binary-name`, `--force`
- Side Effects 详细说明三个修改操作：
  1. store.go: AddNewMethod("store", ...) → IStore 接口添加方法 + store 结构体添加方法
  2. biz.go: AddNewMethod("biz", ...) → IBiz 接口添加方法 + biz 结构体添加方法 + addImport
  3. proto: AddNewGRPCMethod(...) → addImportProto + addRPCsToAPIServer (Create/Update/Delete/DeleteCollection/Get/List)
- Internal Flow 引用 osbuilder-reference 查 AST 机制和 REST 命名规则

- [ ] **Step 3: 验证命令参数准确性**

对照 `create_api.go`, `create_cmd.go`, `create_job.go`, `create_mq.go` 确认命令参数

- [ ] **Step 4: 验证 Side Effects 描述准确性**

对照 `file/method.go` 和 `file/proto.go` 确认 AST 注入和 proto 注入的行为描述

---

### Task 4: 创建 osbuilder-semver Skill

**Files:**
- Create: `~/.claude/skills/osbuilder-semver/SKILL.md`

- [ ] **Step 1: 创建 skill 目录**

```bash
mkdir -p ~/.claude/skills/osbuilder-semver
```

- [ ] **Step 2: 编写 SKILL.md 完整内容**

核心要点：
- Frontmatter: `name: osbuilder-semver`, `description: Use when managing semantic versioning, creating version tags, generating changelogs, or executing release pipelines with osbuilder`
- 11 个章节按设计文档顺序
- Command Reference 覆盖 5 个子命令的用法和关键 flags
- Release Pipeline 完整流水线：gitcheck → before → gpgimport → scm → fetchtag → nextsemver → nextcommit → beforebump → bump → afterbump → beforechangelog → changelog → afterchangelog → gitcommit → beforetag → gittag → aftertag → after
- Task System 说明 Runner 接口（Run + Skip + String）和 Execute 顺序执行逻辑
- Hook System 列出 8 个钩子点：before/after/beforebump/afterbump/beforechangelog/afterchangelog/beforetag/aftertag
- GPG Signing 说明 gpgimport 任务
- Uplift Config 说明配置文件名优先级：`.osbuilder.yml`, `.osbuilder.yaml`, `osbuilder.yml`, `osbuilder.yaml`
- Version Parsing 说明 Increment 类型（NoIncrement/PatchIncrement/MinorIncrement/MajorIncrement/PreReleaseIncrement）和 conventional commit 解析规则
- Internal Architecture 说明代码组织：context/config/gpg/semver/task/version

- [ ] **Step 3: 验证命令参数准确性**

对照 `cmd/semver/` 下的文件确认命令参数和流水线

- [ ] **Step 4: 验证流水线描述准确性**

对照 `cmd/semver/release.go` 和 `cmd/semver/tag.go` 确认任务执行顺序

---

### Task 5: 最终验证与清理

**Files:**
- Review: `~/.claude/skills/osbuilder-reference/SKILL.md`
- Review: `~/.claude/skills/osbuilder-init-project/SKILL.md`
- Review: `~/.claude/skills/osbuilder-iterative-dev/SKILL.md`
- Review: `~/.claude/skills/osbuilder-semver/SKILL.md`

- [ ] **Step 1: 验证所有 skill 文件存在**

```bash
ls -la ~/.claude/skills/osbuilder-*/SKILL.md
```

预期：4 个文件都存在

- [ ] **Step 2: 验证 frontmatter 格式一致性**

每个文件：
- name 仅含字母、数字、连字符
- description 以 "Use when" 开头
- name + description 总长度 < 1024 字符

- [ ] **Step 3: 验证交叉引用正确性**

- `osbuilder-init-project` 引用了 `osbuilder-reference`
- `osbuilder-iterative-dev` 引用了 `osbuilder-reference`
- `osbuilder-semver` 自闭环，无外部引用
- `osbuilder-reference` 不引用其他 skill

- [ ] **Step 4: 验证内容完整性**

对照设计文档的章节结构，确认每个 skill 覆盖了所有规划的章节

- [ ] **Step 5: 验证代码准确性**

关键检查点：
- Available Options 与 `known/available.go` 一致
- REST 命名规则与 `types/webserver.go` 中 `prepareRESTMetadata` 一致
- 配置默认值与 `create_project.go` 中 `correctProjectConfig` 一致
- Side Effects 与 `file/method.go` 和 `file/proto.go` 一致
- Release Pipeline 与 `cmd/semver/release.go` 一致