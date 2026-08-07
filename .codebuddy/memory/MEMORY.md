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

## M6 开放与智能（S8–S11，2026-08-07）
- **S8 搜索 ✅**：PostgreSQL FTS（search_tsv + GIN + 触发器自动同步 search_documents）、搜索历史/书签、SearchView 前端页
- **S9 工作台/仪表盘/视图偏好 ✅**：WorkbenchView + 首屏聚合 API；DashboardView + Widget 框架（10 种类型）+ 风险告警 + 3 套模板；view_preferences 持久化
- **S10 Webhook/Intake ✅**：webhook + webhook_logs（HMAC-SHA256 + 30 天日志 + 测试/重试）；intake_channels + intake_issues（公开表单 + 审核转正）
- **S11 自动化/效能度量 ✅**：automation_rules（TCA DSL + 7 内置模板 + 执行审计）；metric_snapshots + deployment_events（DORA 四指标）
- **认证扩展**：apitoken 模块（JWT + API Token 双通道 Principal Parser）
- **Issue 扩展**：batch/reorder、工时完整 CRUD、评论完整 CRUD
- **迁移递进至 0017**

## ⚠️ 已知问题
- **`internal/application/attachment` 包存在编译风险**（历史遗留）：`handler.go` 与 `service.go` 可能存在接口不匹配，需回归验证。
- **`internal/application/workspace` 包**：`riskrules.go:83` 引用 `pgx.CommandTag`，该类型在当前 pgx v5 版本中已变更/移除，可能导致 `cmd/api` 编译失败。
- **M6 模块前端视图**：Webhook / Intake / 自动化 / 效能度量的后端 API 已就绪，但前端管理视图尚待后续迭代补齐。
