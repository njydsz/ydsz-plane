# Changelog

本文档记录 Ydsz Plane 每个版本的变更。格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

---

## [0.2.0] - 2026-08-07

### 新增 (Added)

#### 核心工作项增强
- **项目模板系统**：新建项目时可选择 敏捷 / 瀑布 / 通用 模板，自动初始化状态流与工作流
- **甘特图视图**：只读甘特图前端组件，支持水平时间轴条形图 + SVG 箭头依赖展示
- **日历视图**：月/周/日视图，按工作项目标日期分布，点击跳转详情
- **缺陷分析报表**：4 KPI 卡片 + 趋势折线 + 严重度/阶段/模块/根因/龄期分布图 + CSV 导出
- **审计日志报表**：前端审计日志展示页，汇总统计 + 操作分布条形图 + 操作者排行

#### 性能增强
- 新增 7 个覆盖索引（`sql/0019_perf_indexes.up.sql`）：针对 100 万工作项量级的 Index-Only Scan 优化
- 新增 `scripts/perf/benchmark.go` 性能基准采样工具，覆盖 10 个查询场景的 P50/P95/P99 统计
- 新增 1M 工作项量级性能调优报告（`docs/reports/S12-performance-tuning.md`）

#### 安全加固
- 安全加固复查通过等保三级基线审查（详见 `docs/security/S12-security-hardening-review.md`）
- 安全响应头中间件：CSP + X-Frame-Options + X-Content-Type-Options
- Redis 令牌桶全局限流，超限返回 429 + Retry-After
- webhook 出站 HMAC-SHA256 签名验证

#### 运维与文档
- **用户手册**（`docs/user-manual.md`）：面向最终用户的完整操作指南
- **运维手册**（`docs/operations-manual.md`）：部署、监控、备份、升级、故障排查
- **API 参考**（`docs/api-reference.md`）：全部 REST/Webhook/WebSocket 端点清单
- **升级指南**（`docs/upgrade-guide.md`）：v0.1 → v0.2 平滑升级步骤

### 改进 (Changed)
- 项目列表页增加模板选择器 UI（三卡片 radio）与模板归属徽章
- 工作项列表查询支持按创建时间 / 更新时间 / 优先级 / 目标日期排序
- 审计日志写入带 c.ClientIP() 与 actor 上下文
- API Token 创建时仅一次回显明文，后续仅可见末 4 位

### 修复 (Fixed)
- 评论内容前端渲染富文本不一致：统一服务端 bluemonday 白名单净化 + 前端 DOMPurify 二次过滤
- WebSocket origin 校验防止跨站劫持
- 甘特图无法在无后端 /gantt 端点时降级使用 issue list
- 日历视图 API 调用从 issueApi.list() 切换为 issueApi.listIssues() 适配新分页结构

---

## [0.1.0] - 2026-07-15

### 初始版本

#### 核心功能
- 工作空间 CRUD + slug 路由 + 成员 4 级 RBAC（owner/admin/member/guest）
- 项目 CRUD + 项目内工作项 CRUD（需求/任务/缺陷 三类型统一）
- 工作项状态机（backlog/started/completed/cancelled 四态组）
- 看板视图 + 视图偏好持久化
- 迭代（Sprint）管理 + 燃尽图 + 站会模式
- 版本管理 + 发布 + 交付报告
- 仪表盘：优先级/状态分布 + 逾期清单
- 全局搜索（PG FTS tsvector）
- 通知中心 + 通知偏好
- 附件上传（MinIO 私有桶 + 预签名直传）
- Webhook 出站 + 自动化规则引擎
- 收件箱（Intake）公开工单提交
- 效能度量：Velocity / Lead Time / Quality / DORA 指标
- Prometheus RED 指标 + /metrics 端点
- Swagger UI（/swagger/index.html）
- Postman Collection

#### 基础设施
- PostgreSQL RLS 租户隔离
- Redis 分布式锁 + 限流
- RabbitMQ Outbox 事件投递
- WebSocket 实时推送
- API Token 鉴权（ydz_ 前缀 + SHA-256 hash）
- CI/CD：GitHub Actions 集成 + 端到端 + 冒烟/负载/压力 k6 测试
- Docker Compose 一键部署

#### 技术栈
- Go 1.26.5 + Gin 1.12
- Vue 3.5 + TypeScript + Vite 6 + Pinia + Vue Router
- PostgreSQL 18
- Redis 8
- RabbitMQ 4
