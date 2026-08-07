# Ydsz Plane 项目长期记忆

## 技术栈与架构
- Go 1.26 + Gin + pgx（PostgreSQL）+ Redis + RabbitMQ + MinIO/S3 + Prometheus，分层架构：
  - `internal/application` 领域应用服务（按域分子包：issue、sprint、version、workspace、auth 等）
  - `internal/infrastructure` 基础设施（cache/events/mail/mq/persistence/storage/telemetry/ws）
  - `internal/interfaces` HTTP handler / middleware / dto
  - `cmd/{api,migrate,worker}` 入口
- 前端：Vue 3 + Pinia + Vue Router + Vite + pnpm workspace，位于 `web/`

## 代码注释约定（对标大厂）
- **Go**：每个包 `// Package xxx ...` 包文档 + 每个导出符号 `//` doc 注释 + 关键逻辑详尽行内注释（中文）。
  - 参考标杆文件：`internal/config/config.go`、`internal/application/version/service.go`。
- **Web 前端**：文件顶部 `/** */` JSDoc 或 `<template>` 顶部 `<!-- 文件级说明 -->`，组件 Props/Events 注释，方法级 JSDoc。
  - 参考标杆：`web/src/api/services/issue.ts`、`web/src/stores/sprint.ts`。
- 注释语言统一为简体中文。

## S7 实施收尾（2026-08-07）
- S7 核心闭环（7.1-7.10）已完成，MVP v0.1 可发布。
- **新增交付物**：
  - `docs/perf/` 性能基线报告目录
  - `scripts/seed-scale/` 大规模造数脚本（支持 100 万工作项并发生成）
  - 工作项 xlsx 导出（纯标准库实现，无第三方依赖）
  - WS 断线重连补偿机制（`ws-client.ts` onReconnect + 后端 `since` 参数）
- **出口检查**：通知系统 6 项检查中，站内信完整闭环（检查项 2）✅，WS 断线补偿已补（检查项 6）✅；邮件/IM/摘要归 S9，规则矩阵完整度归后续迭代。

## ⚠️ 已知问题
- **`internal/application/attachment` 包当前编译失败**（2026-08-07 时点）：存在未提交的半迁移重构。
  - `models.go` 为 untracked 文件，与 `service.go` 类型重声明冲突；
  - `handler.go` 仍调用已删除方法（`ListByEntity`/`Create`/`Get`）；
  - `WithDetails([]errs.FieldDetail{...})` 参数类型不匹配（应为单个 `FieldDetail`）。
  - 修复方案应是：要么完成重构（handler 改用预签名直传 + 新 service 方法），要么回退重构，使 service 与 handler 恢复匹配。**其余所有包均可正常编译。**
- **`internal/application/workspace` 包**：`riskrules.go:83` 引用了 `pgx.CommandTag`，该类型在当前 pgx v5 版本中已变更/移除，导致 `cmd/api` 入口编译失败。
