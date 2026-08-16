<p align="center">
  <h1 align="center">Ydsz Plane</h1>
  <p align="center">
    面向中国软件团队的开源项目管理平台
  </p>
</p>

<p align="center">
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" alt="License: MIT"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.26"></a>
  <a href="https://vuejs.org"><img src="https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue 3.5"></a>
  <img src="https://img.shields.io/badge/PostgreSQL-18-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL 18">
  <img src="https://img.shields.io/badge/Redis-8-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis 8">
  <img src="https://img.shields.io/badge/ES-8.14-005571?style=flat-square&logo=elasticsearch&logoColor=white" alt="Elasticsearch 8">
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> · <a href="./docs">架构文档</a> · <a href="./docs/swagger/swagger.yaml">API 参考</a> · <a href="./LICENSE">许可证</a>
</p>

---

> **项目定位**：Ydsz Plane 是一款对标 [Jira](https://www.atlassian.com/software/jira)、[Linear](https://linear.app)、[云效](https://www.aliyun.com/product/yunxiao)、[TAPD](https://www.tapd.cn)、[ONES](https://ones.cn) 的开源项目管理平台，面向中国软件团队量身定制。以 **模块化单体 + 异步 Worker** 为核心架构，覆盖 **工作空间 ⟶ 项目 ⟶ 工作项（需求/任务/缺陷）** 主价值链，辅以迭代（Sprint）、版本（Version）、看板、效能度量、全文检索、实时通知、Webhook 与自动化规则等能力。遵循 DDD 分层设计，向后兼容微服务拆分；全栈 MIT 开源，支持私有化部署与信创交付。核心模型：WBS 三层级（Epic→Feature→Story）、PDM 四类依赖、M2M 多分配人、需求/任务/缺陷独立追踪。

---

## 目录

- [核心特性](#核心特性)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [功能模块](#功能模块)
- [目录结构](#目录结构)
- [性能目标](#性能目标)
- [测试与质量](#测试与质量)
- [部署](#部署)
- [API 文档](#api-文档)
- [架构文档](#架构文档)
- [开发规范](#开发规范)
- [对标竞品](#对标竞品)
- [路线图](#路线图)
- [如何贡献](#如何贡献)
- [许可证](#许可证)

---

## 特性

### 🎯 核心域（0 → 1 主价值链）

| 能力 | 说明 |
|------|------|
| **工作空间** | 多租户隔离（RLS + 应用层双保险）、4 级角色（Owner/Admin/Member/Guest）、邮箱邀请、SSO/OIDC（规划中） |
| **项目** | 标识符路由（Identifier 唯一键）、网络类型（公开/私有/内部）、功能模块开关（Intake/Sprint/Version/Estimate） |
| **需求** | 产品需求管理，WBS 三层级（Epic→Feature→Story），M2M 分配，状态机流转 |
| **任务** | 技术任务管理，WBS 三层级（主任务→子任务→子子任务），PDM 四类依赖，支持工时估算与实际工作量 |
| **缺陷** | 测试缺陷管理，WBS 三层级（主缺陷→子缺陷→子子缺陷），严重程度 5 级（致命/严重/一般/提示/建议），发现/修复版本追溯 |
| **模块** | 工作项归档属性（非独立管理对象），按项目维护，可配置必填/选填，支持 Owner 与目标版本关联 |
| **版本** | 产品发版里程碑容器（semver 校验），聚合 1~N 个迭代，自动生成 Release Notes、交付报告与变更日志 |
| **迭代** | Sprint 生命周期（规划→执行→复盘）、容量规划、燃尽/燃起图、速率趋势、迭代快照自动留存 |
| **状态机** | 项目级自定义状态与流转规则，6 状态组（backlog/unstarted/started/completed/cancelled/triage），内置三组行业模板 |

### 🚀 增强域（对标大厂标准）

| 能力 | 说明 |
|------|------|
| **看板** | 拖拽排序、列约束（状态组映射）、组视图（分配人/状态/优先级）、快速创建、WIP 限制 |
| **工作台** | 跨空间聚合（指派、收藏、最近访问）、聚焦模式（当前迭代/我的待办） |
| **全局搜索** | 类 JQL 语法，ES 异步双写，IK 中文分词，级联过滤 |
| **实时通知** | 站内 + 邮件 + IM，订阅矩阵 + 摘要汇总（定时 Digest） |
| **仪表盘** | 可拖拽 Widget，DORA 指标、燃起/燃尽、累积流图（CFD）、速率趋势 |
| **收件箱** | 匿名提报（Intake）+ 转正（转为正式工作项）+ 自定义字段 |
| **自动化** | JSON DSL 规则引擎，事件驱动（创建/更新/删除/状态变更）+ DLQ 保障 |
| **Webhook** | HMAC 签名 + 重放防护 + 投递日志 + 手动重试 |
| **AI 集成** | OpenAI/Claude 接口：摘要、分类、估算建议、智能标签 |
| **知识库** | 层级页面、版本历史、评论、@mention、协同编辑 |

---

## 技术栈

| 层 | 技术选型 | 版本 |
|----|----------|------|
| **后端** | Go · Gin · pgx | Go 1.26.5 / Gin 1.12 |
| **前端** | Vue 3 · TypeScript · Vite 6 · Pinia | Vue 3.5 / Vite 6.0 |
| **UI 框架** | Element Plus · TipTap · Tailwind CSS | — |
| **数据库** | PostgreSQL（信创适配达梦/人大金仓） | PG 18 |
| **缓存** | Redis | Redis 8 |
| **消息队列** | RabbitMQ | RabbitMQ 4 |
| **搜索引擎** | Elasticsearch（IK Analyzer） | ES 8.14 |
| **对象存储** | MinIO / S3 | — |
| **实时通信** | WebSocket（gorilla/websocket） | — |
| **可观测性** | Zap · Prometheus · OTel | — |
| **包管理** | pnpm（前端 monorepo） | pnpm 10 |
| **数据迁移** | golang-migrate | v4.19 |

---

## 快速开始

### 前置条件

- [Go 1.26+](https://go.dev/doc/install)
- [Node.js 22+](https://nodejs.org/) + pnpm 10
- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- Elasticsearch 8（可选，搜索功能需要）

### 一行启动

```shell
# 1. 克隆仓库
git clone https://github.com/njydsz/ydsz-plane.git && cd ydsz-plane

# 2. 启动基础设施（PostgreSQL + Redis + Mailpit）
make up

# 3. 数据库迁移
make migrate

# 4. 初始化种子数据
make seed

# 5. 后端（热重载）
make dev-api

# 6. 前端（另开终端）
make dev-web
```

> 浏览器访问 http://localhost:5173 → 使用默认种子账号 `admin@njydsz.com` / `Admin@1020` 登录

### 环境变量（最小配置）

```shell
cp .env.example .env
# 编辑 .env 填入实际配置（数据库、Redis、JWT 密钥等）
```

详见 [本地开发指南](./docs/Ydsz%20Plane%20本地开发环境.md)

---

## 功能模块

根据 DDD 限界上下文划分，后端采用 20 个应用服务模块：

```
internal/application/
├── ai/            # AI 辅助服务
├── apitoken/      # API 令牌管理
├── attachment/    # 文件附件
├── auth/          # 认证与会话
├── automation/    # 自动化规则引擎
├── dashboard/     # 仪表盘
├── dlq/           # 死信队列处理
├── intake/        # 收件箱
├── issue/         # 工作项核心（含评论/关系/依赖/工时）
├── metrics/       # 效能度量
├── notification/  # 通知编排
├── pages/         # 知识库页面
├── preference/    # 用户偏好
├── search/        # 搜索查询编排
├── sprint/        # 迭代
├── version/       # 版本（Version），版本日为其发布日期属性
├── webhook/       # Webhook 出站
├── workbench/     # 工作台
└── workspace/     # 工作空间
```

前端页面覆盖（Playwright E2E 全流程）：

```
web/src/views/
├── IssueDetailView · IssueListView · KanbanBoardView · SpreadsheetView
├── SprintDetailView · SprintPlanningView · SprintStandupView
├── GanttChartView · CalendarView
├── DashboardView · MetricsView · DefectAnalyticsView · DeliveryReportView
├── SearchView
├── IntakeSettingsView · IntakePublicView
├── NotificationView · NotificationPreferencesView
├── WebhookSettingsView · AutomationView
├── ApiTokensView · AuditReportView
└── WorkbenchView · FocusModeView
```

---

## 目录结构

```
ydsz-plane/
├── cmd/
│   ├── api/        # HTTP API Server
│   ├── worker/     # 异步 Worker（通知/ES 索引/自动化）
│   ├── migrate/    # 迁移执行入口
│   ├── notification-service/  # 通知服务独立部署
│   └── search-service/        # 搜索服务独立部署
├── internal/       # 应用核心（Go package 隔离）
│   ├── application/  # 应用服务层（用例编排）
│   ├── domain/       # 领域层（实体/值对象/领域事件）
│   ├── infrastructure/# 基础设施（PG/Redis/ES/MinIO）
│   ├── interfaces/   # 接口层（HTTP Handler/Middleware）
│   ├── rbac/         # RBAC 权限引擎
│   └── config/       # Viper 配置
├── pkg/            # 公共库（crypto/errs/searchql）
├── web/            # 前端（pnpm workspace）
│   ├── src/          # 主应用源码
│   ├── packages/ui/  # 共享 UI 库
│   ├── .storybook/   # 组件文档
│   └── e2e/          # Playwright 端到端测试
├── sql/            # 数据库迁移脚本
├── scripts/        # 运维脚本（seed/reindex/压测）
├── docs/           # 完整架构文档矩阵（18 篇）
│   ├── architecture/ # 架构设计文档
│   ├── deployments/  # 部署物料（Docker/Nginx）
│   ├── ops/          # 运维手册（Oncall/Runbook）
│   ├── perf/         # 压测报告
│   └── swagger/      # OpenAPI 规格
├── deployments/    # 一键部署（docker-compose / Helm）
├── .github/        # CI/CD（GitHub Actions）
└── Makefile        # 常用开发命令
```

---

## 性能目标

| 指标 | 目标值 | 对标标准 |
|------|--------|----------|
| API P95 延迟 | ≤ 200ms | Google SRE / 字节跳动 |
| 页面加载 P95 | ≤ 2s | Web Vitals INP |
| 单项目工作项 | ≥ 100 万 | 大型团队使用场景 |
| 并发用户 | ≥ 1000 | 中大型团队 |
| API 可用性 | ≥ 99.9%/30d | SLA 承诺 |
| 事件投递延迟 | P95 ≤ 5s | Outbox → Worker |

详见 [01-系统架构设计](./docs/architecture/01-系统架构设计.md) §1.3 与 [17-性能基线](./docs/architecture/17-性能基线.md)

---

## 测试与质量

遵循 Google《Software Engineering at Google》测试金字塔原则：

| 层级 | 占比 | 工具 | 门禁 |
|------|------|------|------|
| 单元测试 | 60% | go test · testify · vitest | `go test -race` |
| 集成测试 | 30% | testcontainers（PG/Redis） | 验收前必须通过 |
| E2E 测试 | 10% | Playwright | 冒烟 CI 门控 |

- **CI 流水线**：lint → unit test (race) → build → e2e-smoke → CodeQL 安全扫描
- **覆盖率**：领域层 ≥ 70%，整体 ≥ 50%（ratchet 只升不降）
- **安全**：OWASP Top 10 用例覆盖、govulncheck 每夜扫描、CodeQL 安全分析
- **性能压测**：k6 脚本（读写比 9:1），基线回归 ±20% 门控

详见 [14-测试策略与质量保障](./docs/architecture/14-测试策略与质量保障.md)

---

## 部署

### 一键 Docker Compose（推荐）

```shell
cd deployments
docker compose up -d
# 或使用完整 profile（含 ES + MinIO）
docker compose --profile full up -d
```

### 生产规格

| 形态 | 规格 | 支撑 |
|------|------|------|
| 最小 | 2C4G × 1（All-in-One） | 50 用户 / 5 万工作项 |
| **推荐** | API 2C4G ×2 + Worker + PG 4C16G + Redis 2C4G + ES 4C8G | 1000 并发 / 100 万工作项 |

### 信创交付

- 双架构镜像（amd64 / arm64）+ openEuler / 麒麟基础镜像
- PostgreSQL 国产替代：达梦 / 人大金仓方言层
- 国密算法：build tag `gmssl` 切换 SM2/3/4

详见 [13-部署运维与可靠性设计](./docs/architecture/13-部署运维与可靠性设计.md)

---

## API 文档

基于 [Microsoft REST API Guidelines](https://github.com/microsoft/api-guidelines/blob/vNext/Guidelines.md) 与 [Google AIP](https://google.aip.dev) 设计：

- 风格：REST 资源导向，路径版本 `/api/v1`
- 协商：JSON Merge Patch（PATCH）、信封式错误（`code` / `message` / `details` / `request_id`）
- 分页：Cursor + Offset 双模式，可组合 filter/sort/expand
- 幂等：Idempotency-Key 保证写操作安全重试
- 限流：滑动窗口 + `Retry-After`

在线浏览 Swagger UI：启动服务后访问 `http://localhost:8080/swagger/index.html`

本地生成：

```shell
make openapi
```

详见 [05-API 设计规范](./docs/architecture/05-API设计规范.md) 与 [docs/swagger/swagger.yaml](./docs/swagger/swagger.yaml)

---

## 架构文档

详见 `docs/architecture/` 矩阵（共 18 篇，覆盖全 S1–S12 开发周期）：

| 文档 | 内容 |
|------|------|
| [01-系统架构设计](./docs/architecture/01-系统架构设计.md) | 总体架构 · ADR · 领域上下文 · 技术选型 |
| [03-工程基座与开发规范](./docs/architecture/03-工程基座与开发规范.md) | Monorepo · 12-Factor · CI/CD · 分支策略 |
| [04-数据模型设计](./docs/architecture/04-数据模型设计.md) | DDL 定稿 · 索引 · RLS · 信创方言 |
| [05-API 设计规范](./docs/architecture/05-API设计规范.md) | REST 约定 · 错误码 · 分页/过滤/幂等 |
| [06-权限与安全设计](./docs/architecture/06-权限与安全设计.md) | RBAC 双级 · 认证 · OWASP 映射 · 等保 |
| [07-工作项与状态机](./docs/architecture/07-工作项与状态机设计.md) | Issue 聚合 · WBS · 状态机模板 |
| [09-通知与实时推送](./docs/architecture/09-通知与实时推送设计.md) | 通知管道 · 摘要 · IM · WS 扇出 |
| [12-开放集成](./docs/architecture/12-开放集成设计.md) | Webhook · OpenAPI · 自动化 DSL |
| [18-开发者实现参考](./docs/architecture/18-开发者实现参考手册.md) | 代码→文档全量可追溯映射 |

> **阅读路径**
> - 新成员上手：01 → 03 → 02
> - 后端开发：04 → 05 → 07 → 对应域文档
> - 前端开发：01 §6 → 05（接口约定）
> - QA：14 → 各域验收清单
> - 运维/SRE：13 → 03 §4

---

## 开发规范

- **Go 编码**：遵循 [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) + `golangci-lint` 0 error 门禁
- **前端编码**：ESLint + Prettier + vue-tsc 强类型检查，无 any 逃逸
- **提交规范**：[Conventional Commits](https://www.conventionalcommits.org/)（CI 用 commitlint 校验）
- **分支策略**：Trunk-Based Development，PR squash merge，核心域 2 人 Approve
- **版本管理**：SemVer，CHANGELOG 由 git-cliff 自动生成

---

## 对标竞品

| 能力域 | 云效 | TAPD | ONES | PingCode | Jira | Ydsz Plane |
|--------|------|------|------|----------|------|------------|
| 需求管理（含子需求） | ● | ● | ● | ● | ● | ● |
| 任务管理（含子任务） | ● | ● | ● | ● | ● | ● |
| 缺陷管理（含子缺陷） | ● | ● | ● | ● | ● | ● |
| 模块（=工作项归档属性） | ● | ● | ● | ● | ● | ● |
| 迭代（Sprint） | ● | ● | ● | ● | ● | ● |
| 版本（=1~N 迭代） | ● | ○ | ● | ● | ● | ● |
| 仪表盘 | ● | ● | ● | ● | ● | ● |
| 工作台 | ● | ● | ● | ● | ● | ● |
| 收件箱 | ● | ● | ● | ● | ● | ● |
| 效能度量 | ● | ○ | ● | ● | ○ | ● |
| 开源/信创 | ✕ | ✕ | △ | △ | ✕ | ● 核心差异 |

> 图例：● 已支持 / △ 部分 / ○ 不支持 / ✕ 不适用

---

## 路线图

- [x] S1 工程基座（Monorepo + CI/CD + 编码规范）
- [x] S2 认证 + 工作空间 + RBAC
- [x] S3 项目 + 工作项核心（WBS/状态机/看板）
- [x] S4 视图（甘特/日历/表格）+ 评论体系
- [x] S5 迭代（Sprint）生命周期
- [x] S6 版本 + Release Notes
- [x] S7 通知管道 + 实时推送
- [x] S8 全局搜索（JQL + ES）
- [x] S9 仪表盘 + 效能度量（DORA）
- [x] S10 Webhook + 开放集成
- [x] S11 自动化规则 + AI + 知识库
- [x] S12 安全/性能/信创交付 + E2E 收尾
- [ ] S13 Phase 2：OIDC/SAML、多语言翻译
- [x] S14 Phase 3（已完成）：微服务拆分
  - [x] Proto 契约先行：NotificationService + SearchService gRPC API
  - [x] 通知服务独立部署（notification-svc）：独立 PG 数据库 + RabbitMQ 消费
  - [x] 搜索服务独立部署（search-svc）：ES 读写 + PG FTS 降级
  - [ ] Phase 4：Webhook/Metrics 服务独立（长期）

---

## 如何贡献

欢迎所有形式的贡献！请确保：

1. **Fork** 本仓库，在 `feature/xxx` 或 `fix/xxx` 分支开发
2. 提交前运行 `make lint && make test`，确保 CI 绿
3. 提交信息遵循 Conventional Commits 规范
4. PR 描述清楚变更动机、影响范围与测试结果
5. 至少 1 人 Approve（核心域需 2 人）后 squash merge 到 main

详见 [CONTRIBUTING.md](./CONTRIBUTING.md)（即将推出）。

---

## 致谢

- [Plane](https://github.com/makeplane/plane) — 视觉层设计语言参考
- [云效](https://www.aliyun.com/product/yunxiao)、[TAPD](https://www.tapd.cn)、[ONES](https://ones.cn) — 产品设计与本土化最佳实践参考
- [Elasticsearch IK Analyzer](https://github.com/medcl/elasticsearch-analysis-ik) — 中文分词
- [golang-migrate](https://github.com/golang-migrate/migrate) — 数据库迁移
- [go-chi/gin](https://github.com/gin-gonic/gin) — HTTP 框架
- [Pinia](https://pinia.vuejs.org/)、[Vue Router](https://router.vuejs.org/)、[Vue I18n](https://vue-i18n.intlify.dev/)
- [TipTap](https://tiptap.dev/) — 富文本编辑器
- [k6](https://k6.io/) — 性能压测

---

## 许可证

本项目基于 [MIT License](./LICENSE) 开源发布。

> Copyright © 2026 ydsz-plane

---

<p align="center">
  ⭐ 如果这个项目对您有帮助，请点 Star 支持 — 让更多中国软件团队看到！
</p>
