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

## ⚠️ 已知问题
- **`internal/application/attachment` 包当前编译失败**（2026-08-07 时点）：存在未提交的半迁移重构。
  - `models.go` 为 untracked 文件，与 `service.go` 类型重声明冲突；
  - `handler.go` 仍调用已删除方法（`ListByEntity`/`Create`/`Get`）；
  - `WithDetails([]errs.FieldDetail{...})` 参数类型不匹配（应为单个 `FieldDetail`）。
  - 修复方案应是：要么完成重构（handler 改用预签名直传 + 新 service 方法），要么回退重构，使 service 与 handler 恢复匹配。**其余所有包均可正常编译。**
