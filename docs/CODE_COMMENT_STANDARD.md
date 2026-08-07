# Ydsz Plane 代码注释规范（CODE_COMMENT_STANDARD）

> 版本：v1.0 ｜ 适用范围：全部 Go 与 TypeScript/Vue 源码
> 依据：Google 开源项目注释指南、字节跳动《Go 语言代码规范》、腾讯前端注释规范（大厂通用实践）

## 1. 总则

1. **注释必须为中文**。术语（如 JWT、REST、RabbitMQ）可保留英文原文；专有名词首字母缩写保持原样。
2. **注释解释「为什么」，而非「是什么」**。代码本身可读的（如 `i++`）不写注释；注释回答「为什么这么做」「边界条件是什么」「调用方需要注意什么」。
3. **不写噪音注释**：禁止 `// 获取用户` 配 `func GetUser()` 这类与签名重复的注释；禁止整段照抄代码逻辑的注释。
4. **注释与代码同步维护**：改代码必须同步改注释，禁止出现过期注释。
5. **只注释代码本身**：不写作者、修改日期等易腐烂信息（由 Git 记录），如需写归属信息放文件头。

## 2. Go 注释规范

### 2.1 包注释

每个包必须有包注释，置于 `package` 声明上方，说明该包职责、主要入口、典型用法。

```go
// Package auth implements authentication use cases: password login, token
// issue/refresh, and token parsing.
//
// 该包是认证领域层的应用服务入口，依赖基础设施层的持久化与缓存能力。
package auth
```

> 大厂实践：包注释一句话说明职责；如包内存在多个聚合根（如 workspace），可展开列出子域。

### 2.2 导出符号注释（强制）

**所有导出（大写开头）的类型、函数、方法、常量、变量、字段，必须紧邻上方有注释**，缺失将无法通过 golint/golangci-lint 的 `revive` 检查。

- 注释以**符号名开头**（Go 官方 doc 约定）：`// Service 提供认证用例：…`
- 方法接收者无需重复包名。
- 字段注释可写在字段上方，也可写在同行行尾（短字段用行尾注释更紧凑）。

```go
// Service 提供认证用例：登录、注册、token 签发/解析。
type Service struct {
	db               *pgxpool.Pool
	secret           []byte        // JWT 签名密钥（HS256 共享密钥）。
	issuer           string        // JWT iss 声明，用于多租户签发者区分。
}
```

### 2.3 未导出符号注释（推荐）

未导出符号**非强制**，但满足以下条件之一必须补注释：

- 逻辑复杂、分支多（>10 行且含多个 if/switch）；
- 有安全/并发/性能隐含约束（如锁、goroutine、panic 路径）；
- 返回值语义不直观（错误码、状态机迁移）。

### 2.4 函数/方法注释格式

```
// 函数名 <动词短语>（做什么）。
//
// <补充：参数约定、返回值语义、错误条件、并发安全、性能注意、调用方约束>
// 每条补充用 "//" 独立成段，段间空行。
```

示例（大厂标准写法）：

```go
// Reset 重置用户密码。
//
// 参数：
//   - userID：目标用户 ID，须为已存在用户。
//
// 返回值：
//   - 成功返回 nil；用户不存在返回 errs.ErrNotFound。
//
// 注意：调用方必须保证该函数在事务内执行，否则密码散列更新可能丢失。
func (s *Service) Reset(ctx context.Context, userID int64) error {
```

### 2.5 行内注释

- 复杂算法步骤、防呆分支（防御性检查）、性能优化意图、跨函数不变量使用行内注释。
- 行内注释与代码同行时，与代码之间至少隔一个空格；独占一行时顶格对齐所在缩进。
- 禁止用行内注释解释显而易见的语法。

```go
// 幂等检查：同一 outbox 事件重复投递时直接跳过，防止消费端重复处理。
if exists, err := s.exists(ctx, eventID); err == nil && exists {
	return nil
}
```

### 2.6 TODO / FIXME / XXX 标记

```go
// TODO(负责人): 迁移到 RS256 非对称签名（当前 HS256 仅用于 MVP 阶段）。
// FIXME: 该分支在并发下存在竞态，需要加锁。
```

### 2.7 测试注释

- 测试函数可省注释，但表驱动测试的**每个用例必须有 name 说明**，且 name 使用中文描述意图（如 `"密码错误返回 401"`）。
- 复杂测试 Setup / 断言步骤使用行内注释说明业务预期。

## 3. TypeScript / Vue 注释规范

### 3.1 文件头注释（强制）

每个 `.ts` / `.vue` 源文件顶部必须有 JSDoc 文件头，说明文件职责。

```ts
/**
 * auth 认证服务 API 封装。
 * 提供登录 / 注册 / 刷新 / 登出等接口调用，与后端 internal/interfaces/http 对应。
 */
```

`.vue` 文件头建议放在 `<script>` 块内或文件第一行（`<template>` 之前用 HTML 注释可能被构建器剥离，统一放 `<script>` 内）。

### 3.2 导出成员注释（强制）

所有 `export` 的 interface / type / class / function / const 必须紧邻上方有 `/** */` 注释。

```ts
/** 用户简要信息（列表场景使用，不含敏感字段） */
export interface UserBrief {
  id: number;
  /** 登录邮箱，唯一 */
  email: string;
  display_name: string;
}
```

- interface 字段注释写在字段上方；紧凑结构可同行行尾。
- 复杂函数的 `@param` / `@returns` 需写明类型语义与边界。

### 3.3 组件注释

Vue 组件（`.vue`）需在 `<script setup lang="ts">` 内对组件本身写 JSDoc（props/emits 说明用途），对复杂 props 逐项注释。

```vue
<script setup lang="ts">
/**
 * 看板视图组件：以泳道卡片形式展示工作项，支持拖拽流转状态。
 * 依赖：useIssueStore；事件：issue:move。
 */
</script>
```

### 3.4 行内注释

- 与 Go 相同：解释「为什么」与边界，不解释语法。
- 异步时序、状态机流转、错误分支必须注释。

## 4. 禁止事项

| 禁止 | 示例 | 正确做法 |
| --- | --- | --- |
| 注释与签名重复 | `// 获取用户` + `GetUser()` | 说明参数/返回值/错误语义 |
| 注释整段代码 | `// a = b + c 将 b 与 c 相加赋值给 a` | 删除 |
| 过期注释 | 注释描述旧逻辑 | 同步更新 |
| 英文注释 | `// get user info` | 改为中文（术语除外） |
| 写作者/日期 | `// by zhangsan 2024-01-01` | 交给 Git 管理 |

## 5. 达标线

| 指标 | 目标 |
| --- | --- |
| Go 包注释覆盖率 | 100% |
| Go 导出符号注释覆盖率 | 100% |
| Go 未导出复杂符号注释 | ≥90% |
| TS/Vue 文件头注释覆盖率 | 100% |
| TS/Vue 导出成员 JSDoc 覆盖率 | ≥98% |
| 注释语言 | 100% 中文（术语除外） |

## 6. 扫描与验收

- 覆盖率扫描脚本：`docs/scripts/scan_comments.py`
- 全量报告：`python docs/scripts/scan_comments.py`
- 待办明细：`python docs/scripts/scan_comments.py --todo`
- 合并前必须通过 `go build ./...`、`go vet ./...`、前端 `tsc --noEmit` 与 ESLint。
