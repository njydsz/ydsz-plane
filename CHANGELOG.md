# Changelog

本项目遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 与 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

发布流程与分支策略见 [docs/architecture/03-工程基座与开发规范.md](docs/architecture/03-工程基座与开发规范.md) 与 `docs/architecture/15-发布管理.md`。

## [Unreleased]

### 后续规划（Phase 3+）
- 通知多渠道：IM（企微 / 钉钉 / 飞书）、邮件摘要
- 全局搜索升级为 Elasticsearch（IK 分词 + 类 JQL 语法解析器）
- 甘特图 / 日历 / 电子表格视图
- SSO / SAML 集成、国际化、PWA
- 数据迁移工具（Jira / 云效 / ONES 导入）
- AI 功能（智能分配、重复检测）
- 信创数据库实测（达梦 / 人大金仓 + 国密算法）
- Webhook / Intake / 自动化 / 效能度量的前端管理视图

## [0.2.0] - 2026-08-07（M6 开放与智能 — S8–S11 发布）

> 对应路线图 M6：搜索、工作台、仪表盘、视图偏好、Webhook、Intake、自动化规则引擎、效能度量全部就绪。
> 特性：多对象全文搜索、个人工作台首屏、项目 Widget 仪表盘、Webhook 订阅与投递、Intake 收件箱与公开提交通道、TCA 自动化规则引擎、DORA/速度/前置时间/质量效能度量。

### 新增（已实现）

#### S8 搜索
- **后端**：PostgreSQL FTS（search_tsv tsvector + GIN 索引 + 触发器自动同步）、多对象检索（issues/sprints/versions 通过 search_documents 统一表）、搜索历史（search_history）、搜索书签（search_bookmarks 支持共享）
- **前端**：SearchView 独立搜索页（左过滤器 + 中间结果分组 + 关键词高亮 + 历史/书签侧边栏）

#### S9 工作台 / 仪表盘 / 视图偏好
- **后端**：个人工作台首屏聚合（workbench_summary：任务分桶 / 迭代概览 / 最近访问 / 快捷操作）、工作台配置与模板应用；项目仪表盘 Widget 框架（10 种 Widget 类型）、dashboard_snapshots 快照加速、risk_rules 风险预警（6 种规则类型）、risk_alerts 告警管理、dashboard_templates 预设模板（3 套）；view_preferences 视图偏好持久化
- **前端**：WorkbenchView、DashboardView（CSS Grid 12 列响应式）、Widget 组件注册表、风险告警列表、通知偏好设置页

#### S10 Webhook / Intake
- **后端**：webhook 管理（CRUD + HMAC-SHA256 签名 + events 过滤 + unhealthy 标记）、webhook_logs 投递日志（30 天保留 + 测试投递 / 手动重试）；intake_channels 收件通道（公开表单 + 限流 + 验证码 + 自定义字段 + 自动分配规则）、intake_issues 收件工单（审核 + 转正 + tracking 回执）
- **前端**：暂无独立管理页（API 就绪，待后续迭代补齐前端视图）

#### S11 自动化规则 / 效能度量
- **后端**：automation_rules（TCA DSL JSONB + draft/active/disabled/error 状态机 + 连续失败自动降级）、rule_executions 执行审计（防循环 + dry-run + 触发深度链路）、automation_templates 7 条内置模板；metric_snapshots 每日效能快照（DORA + 速度 + 前置时间 + 质量 + WIP）、deployment_events 部署事件上报（幂等）、metric_adjustments 管理员数据校准
- **前端**：暂无独立管理页（API 就绪，待后续迭代补齐前端视图）

### 增强（已有功能迭代）
- **Issue 域扩展**：batchIssues 批量创建工作项、reorderIssue 看板排序、工时记录完整 CRUD（updateTimeLog / deleteTimeLog）、评论完整 CRUD（updateComment / deleteComment）
- **认证扩展**：API Token 管理（创建 / 吊销 / scopes）、Principal 双通道认证（JWT + API Token 复合解析器）
- **CI 扩展**：CodeQL 安全分析工作流、k6 perf 性能压测工作流
- **seeds 扩展**：seed-scale 大规模造数脚本（100 万工作项、8 并发、断点续传）
- **导出基础设施**：internal/interfaces/http/export 通用导出包（CSV UTF-8 BOM / 最小合法 OOXML xlsx）

### 新增数据库迁移（0008–0017）
- `0008_search` — search_tsv 列 + search_documents/search_history/search_bookmarks 表 + FTS 触发器
- `0009_version_fix` + `0009_workbench` — 版本修复 + 工作台首屏
- `0010_dashboard` + `0010_notifications` — 仪表盘 Widget / 风险告警 / 通知表
- `0011_issue_comments` — 评论主表扩容
- `0012_notification_settings` — 通知偏好设置
- `0013_attachments` + `0013_s6_hardening` — 附件表 + S6 加固
- `0014_view_preferences` — 视图偏好持久化
- `0015_automation` + `0015_webhooks` — 自动化规则 + Webhook 管理
- `0016_intake` + `0016_metrics` — Intake 收件箱 + 效能度量
- `0017_notification_dispatcher` — 通知分发 Worker 支撑（next_retry_at + 部分索引）

## [0.1.0] - 2026-08-07（MVP v0.1 发布）

> 对应路线图 M0–M5：功能底座、核心域 API 与前端视图骨架全部就绪，Sprint 1–7 完成。
> 特性：附件（MinIO）、评论（富文本 + @提及）、通知（站内信 + 偏好）、WebSocket 实时推送（Redis Pub/Sub 多节点扇出）、CSV/xlsx 导出、性能基线脚本与 CI 压测流水线。

### 新增（已实现）
- **工程基座**：Monorepo（Go module + pnpm workspace）、Docker Compose 全栈、Makefile、DB 迁移（0001–0017）
- **鉴权链路**：注册 / 登录（bcrypt + JWT access/refresh）、Cookie 会话、401 单飞刷新重放、忘记 / 重置密码
- **RBAC**：Owner/Admin/Member/Guest 四角色 × 10 项权限中间件
- **工作空间**：CRUD、Slug 校验、成员邀请（token + 7 天有效 + 可撤销）、审计日志
- **项目**：CRUD、Identifier 生成、RLS 租户隔离
- **工作项**：Issue CRUD、状态机流转、WBS 三级、类型差异化字段、工时记录、关联（6 种关系）、依赖（FS/SS/FF/SF）、CSV/xlsx 导出
- **迭代**：Sprint 生命周期、燃尽图、容量规划、速率建议、复盘快照、站会模式
- **版本日**：Version CRUD + 状态机、进度聚合、交付报告、Release Notes、缺陷面板
- **协同增强**：评论（富文本 + @提及 + 嵌套回复）、附件（MinIO 预签名上传/下载/预览）、通知系统（站内信 + 默认规则矩阵 + 铃铛下拉面板 + 偏好设置）、WebSocket 实时推送（Redis Pub/Sub 多节点扇出）、API Token 认证
- **缺陷分析**：`/analytics/defects` 聚合查询（严重程度/发现阶段/模块/根因/缺陷龄/周趋势）+ `/analytics/defects/export` 明细导出（CSV / xlsx，支持按版本过滤）
- **可观测与安全**：Prometheus RED 指标、8 项安全头、Swagger、SMTP 邮件抽象

### S7 收尾（MVP v0.1）
- 工作项导出：新增 xlsx 格式导出（纯标准库实现，无需第三方依赖），前端增加导出格式下拉菜单（CSV / Excel）
- 缺陷分析导出：抽取通用导出基础设施 `internal/interfaces/http/export`（CSV UTF-8 BOM / 最小合法 OOXML xlsx，含 XML 转义单测），缺陷面板新增导出入口（按当前版本过滤）
- 性能基线：创建 `docs/perf/` 目录归档基线报告，新增 `scripts/seed-scale/` 大规模造数脚本（支持 100 万工作项并发写入、断点续传）；k6 三件套脚本（smoke / load / stress）就绪并验证
- 性能流水线：修复 `perf.yml` CI 脚本路径与 Makefile 压测命令（统一指向 `tests/perf/`），支持冒烟 → 负载 → 压力完整流程
- 通知中心：铃铛下拉面板（未读角标、领域图标、已读/未读高亮）、通知列表页（筛选/分页/全部已读/归档）、通知偏好设置页
- MVP 加固：全链路加载态/空态/错误态组件（AppLoadingState / AppEmptyState / AppErrorState / NotFoundView）覆盖通知、评论、看板、工作项列表、仪表盘、燃尽图、交付报告、发布向导、工作空间设置、版本列表、关联面板；修复 SearchView 既有类型错误

### 已知待完善（后续版本）
- 前端测试覆盖率仍偏低，需按测试策略文档持续补充
- CI 需接入真实后端的 E2E 环节
- 国密算法、信创数据库方言层为预留接口，尚未完成实测链路
- 性能基线实测数据待首轮 staging 全量压测后回填（`docs/perf/README.md`）
